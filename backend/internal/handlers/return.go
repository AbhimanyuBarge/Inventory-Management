package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
	"github.com/inventory/backend/internal/pdf"
	"gorm.io/gorm"
)

type CreateReturnItemRequest struct {
	BillItemID uint `json:"bill_item_id" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,gt=0"`
}

type CreateReturnRequest struct {
	OriginalBillID uint                      `json:"original_bill_id" binding:"required"`
	RefundMode     string                    `json:"refund_mode"`
	Reason         string                    `json:"reason"`
	Notes          string                    `json:"notes"`
	Items          []CreateReturnItemRequest `json:"items" binding:"required,min=1"`
}

type ReturnableItemResponse struct {
	BillItemID         uint    `json:"bill_item_id"`
	ItemID             *uint   `json:"item_id"`
	SparePartID        *uint   `json:"spare_part_id,omitempty"`
	PartName           string  `json:"part_name"`
	PartSKU            string  `json:"part_sku"`
	RackInfo           string  `json:"rack_info"`
	UnitPrice          float64 `json:"unit_price"`
	OriginalQuantity   int     `json:"original_quantity"`
	ReturnedQuantity   int     `json:"returned_quantity"`
	ReturnableQuantity int     `json:"returnable_quantity"`
}

func GetBillReturnable(c *gin.Context) {
	billID := c.Param("id")
	var bill models.Bill
	if err := database.DB.Preload("Customer").Preload("User").Preload("Items").First(&bill, billID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bill not found"})
		return
	}

	var returnableItems []ReturnableItemResponse
	for _, item := range bill.Items {
		// Sum already returned quantity for this bill item
		var returnedQtySum struct {
			Total float64
		}
		database.DB.Model(&models.ReturnItem{}).
			Joins("JOIN return_bills ON return_bills.id = return_items.return_bill_id").
			Where("return_bills.original_bill_id = ? AND return_items.bill_item_id = ?", bill.ID, item.ID).
			Select("COALESCE(SUM(return_items.quantity), 0) as total").
			Scan(&returnedQtySum)

		alreadyReturned := int(returnedQtySum.Total)
		returnable := item.Quantity - alreadyReturned
		if returnable < 0 {
			returnable = 0
		}

		targetID := item.ItemID
		if targetID == nil {
			targetID = item.SparePartID
		}

		returnableItems = append(returnableItems, ReturnableItemResponse{
			BillItemID:         item.ID,
			ItemID:             targetID,
			SparePartID:        targetID,
			PartName:           item.PartName,
			PartSKU:            item.PartSKU,
			RackInfo:           item.RackInfo,
			UnitPrice:          item.UnitPrice,
			OriginalQuantity:   item.Quantity,
			ReturnedQuantity:   alreadyReturned,
			ReturnableQuantity: returnable,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"bill":  bill,
		"items": returnableItems,
	})
}

func CreateReturn(c *gin.Context) {
	var req CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, exists := c.Get("user_id")
	var userID *uint
	if exists {
		uid := userIDVal.(uint)
		userID = &uid
	}

	var createdReturn models.ReturnBill

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch Original Bill
		var originalBill models.Bill
		if err := tx.Preload("Customer").Preload("Items").First(&originalBill, req.OriginalBillID).Error; err != nil {
			return fmt.Errorf("original invoice #%d not found: %w", req.OriginalBillID, err)
		}

		// 2. Generate Sequential Return Number (RET-YYYY-XXXX)
		yearStr := time.Now().Format("2006")
		var totalReturnsCount int64
		tx.Model(&models.ReturnBill{}).Count(&totalReturnsCount)
		returnNumber := fmt.Sprintf("RET-%s-%04d", yearStr, totalReturnsCount+1)

		// 3. Validate & Process Return Items and Stock Re-addition
		var refundSubtotal float64 = 0
		var returnItems []models.ReturnItem

		for _, retItemReq := range req.Items {
			// Find original bill item
			var originalItem *models.BillItem
			for i := range originalBill.Items {
				if originalBill.Items[i].ID == retItemReq.BillItemID {
					originalItem = &originalBill.Items[i]
					break
				}
			}

			if originalItem == nil {
				return fmt.Errorf("item ID %d does not belong to invoice %s", retItemReq.BillItemID, originalBill.InvoiceNumber)
			}

			// Check past returns for this bill item
			var returnedQtySum struct {
				Total float64
			}
			tx.Model(&models.ReturnItem{}).
				Joins("JOIN return_bills ON return_bills.id = return_items.return_bill_id").
				Where("return_bills.original_bill_id = ? AND return_items.bill_item_id = ?", originalBill.ID, originalItem.ID).
				Select("COALESCE(SUM(return_items.quantity), 0) as total").
				Scan(&returnedQtySum)

			alreadyReturned := int(returnedQtySum.Total)
			remainingReturnable := originalItem.Quantity - alreadyReturned

			if retItemReq.Quantity > remainingReturnable {
				return fmt.Errorf("cannot return %d of '%s'. Purchased: %d, Already returned: %d, Remaining returnable: %d",
					retItemReq.Quantity, originalItem.PartName, originalItem.Quantity, alreadyReturned, remainingReturnable)
			}

			// Increase inventory stock
			targetItemID := originalItem.ItemID
			if targetItemID == nil {
				targetItemID = originalItem.SparePartID
			}
			if targetItemID != nil {
				var item models.Item
				if err := tx.First(&item, *targetItemID).Error; err == nil {
					item.Quantity += retItemReq.Quantity
					if err := tx.Save(&item).Error; err != nil {
						return fmt.Errorf("failed to restock item '%s': %w", item.Name, err)
					}
				}
			}

			itemRefundTotal := math.Round((originalItem.UnitPrice*float64(retItemReq.Quantity))*100) / 100
			refundSubtotal += itemRefundTotal

			bItemID := originalItem.ID
			returnItems = append(returnItems, models.ReturnItem{
				BillItemID:  &bItemID,
				ItemID:      targetItemID,
				SparePartID: targetItemID,
				PartName:    originalItem.PartName,
				PartSKU:     originalItem.PartSKU,
				RackInfo:    originalItem.RackInfo,
				UnitPrice:   originalItem.UnitPrice,
				Quantity:    retItemReq.Quantity,
				TotalPrice:  itemRefundTotal,
			})
		}

		// 4. Calculate GST Refund
		gstRefunded := originalBill.GSTApplied
		gstRate := 0.0
		gstAmount := 0.0
		if gstRefunded && originalBill.GSTRate > 0 {
			gstRate = originalBill.GSTRate
			gstAmount = math.Round((refundSubtotal*(gstRate/100.0))*100) / 100
		}

		totalRefundAmount := math.Round((refundSubtotal+gstAmount)*100) / 100

		refundMode := req.RefundMode
		if refundMode == "" {
			refundMode = "cash"
		}

		createdReturn = models.ReturnBill{
			ReturnNumber:      returnNumber,
			OriginalBillID:    originalBill.ID,
			CustomerID:        originalBill.CustomerID,
			UserID:            userID,
			RefundSubtotal:    refundSubtotal,
			GSTRefunded:       gstRefunded,
			GSTRate:           gstRate,
			GSTAmount:         gstAmount,
			TotalRefundAmount: totalRefundAmount,
			RefundMode:        refundMode,
			Reason:            req.Reason,
			Notes:             req.Notes,
			Items:             returnItems,
			CreatedAt:         time.Now(),
		}

		if err := tx.Create(&createdReturn).Error; err != nil {
			return fmt.Errorf("failed to save return bill: %w", err)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("OriginalBill").Preload("Customer").Preload("User").Preload("Items").First(&createdReturn, createdReturn.ID)
	c.JSON(http.StatusCreated, createdReturn)
}

func ListReturns(c *gin.Context) {
	query := database.DB.Preload("OriginalBill").Preload("Customer").Preload("User").Preload("Items")

	customerID := c.Query("customer_id")
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	search := strings.TrimSpace(c.Query("q"))
	if search != "" {
		pat := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN customers ON customers.id = return_bills.customer_id").
			Joins("LEFT JOIN bills ON bills.id = return_bills.original_bill_id").
			Where("LOWER(return_bills.return_number) LIKE ? OR LOWER(bills.invoice_number) LIKE ? OR LOWER(customers.name) LIKE ? OR customers.mobile LIKE ?",
				pat, pat, pat, pat)
	}

	var returns []models.ReturnBill
	if err := query.Order("return_bills.created_at desc").Find(&returns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch return history"})
		return
	}

	c.JSON(http.StatusOK, returns)
}

func GetReturn(c *gin.Context) {
	id := c.Param("id")
	var returnBill models.ReturnBill
	if err := database.DB.Preload("OriginalBill").Preload("Customer").Preload("User").Preload("Items").First(&returnBill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Return record not found"})
		return
	}
	c.JSON(http.StatusOK, returnBill)
}

func GetReturnPDF(c *gin.Context) {
	id := c.Param("id")
	var returnBill models.ReturnBill
	if err := database.DB.Preload("OriginalBill").Preload("Customer").Preload("User").Preload("Items").First(&returnBill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Return record not found"})
		return
	}

	var originalBill models.Bill
	if returnBill.OriginalBillID > 0 {
		database.DB.First(&originalBill, returnBill.OriginalBillID)
	}

	var setting models.Setting
	database.DB.First(&setting)

	pdfBytes, err := pdf.GenerateReturnPDF(&returnBill, &originalBill, &setting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate return voucher PDF: %v", err)})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=CreditNote-%s.pdf", returnBill.ReturnNumber))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
