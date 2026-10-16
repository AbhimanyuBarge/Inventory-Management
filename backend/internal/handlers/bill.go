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

type CreateBillItemRequest struct {
	ItemID      uint    `json:"item_id"`
	SparePartID uint    `json:"spare_part_id"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gte=0"`
}

type CreateBillRequest struct {
	// Customer Details (can be existing ID or fresh inputs)
	CustomerID  *uint   `json:"customer_id"`
	Name        string  `json:"name"` // customer name if new/manual
	Mobile      string  `json:"mobile"`
	Email       string  `json:"email"`
	Address     string  `json:"address"`
	GSTApplied  bool    `json:"gst_applied"`
	GSTRate     float64 `json:"gst_rate"` // if custom; if 0 & GSTApplied, uses system default
	Discount    float64 `json:"discount"`
	PaymentMode string  `json:"payment_mode"`
	Notes       string  `json:"notes"`

	Items []CreateBillItemRequest `json:"items" binding:"required,min=1"`
}

func CreateBill(c *gin.Context) {
	var req CreateBillRequest
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

	var createdBill models.Bill

	// Perform all billing actions in an atomic database transaction
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Resolve Customer
		var customer models.Customer
		if req.CustomerID != nil && *req.CustomerID > 0 {
			if err := tx.First(&customer, *req.CustomerID).Error; err != nil {
				return fmt.Errorf("selected customer not found: %w", err)
			}
		} else if strings.TrimSpace(req.Name) != "" {
			// Find existing by mobile if provided, else create new
			cleanMobile := strings.TrimSpace(req.Mobile)
			if cleanMobile != "" {
				tx.Where("mobile = ?", cleanMobile).First(&customer)
			}

			if customer.ID == 0 {
				customer = models.Customer{
					Name:    strings.TrimSpace(req.Name),
					Mobile:  cleanMobile,
					Email:   strings.TrimSpace(req.Email),
					Address: strings.TrimSpace(req.Address),
				}
				if err := tx.Create(&customer).Error; err != nil {
					return fmt.Errorf("failed to save customer record: %w", err)
				}
			}
		}

		// 2. Compute Customer Sequence (Bill #X for this customer)
		var customerSeq int = 1
		if customer.ID > 0 {
			var prevCount int64
			tx.Model(&models.Bill{}).Where("customer_id = ?", customer.ID).Count(&prevCount)
			customerSeq = int(prevCount) + 1
			customer.TotalBills = customerSeq
			tx.Save(&customer)
		}

		// 3. Generate Global Invoice Number
		var setting models.Setting
		tx.First(&setting)
		prefix := "INV"
		if setting.InvoicePrefix != "" {
			prefix = setting.InvoicePrefix
		}
		yearStr := time.Now().Format("2006")
		var totalBillsCount int64
		tx.Model(&models.Bill{}).Count(&totalBillsCount)
		invoiceNumber := fmt.Sprintf("%s-%s-%04d", prefix, yearStr, totalBillsCount+1)

		// 4. Validate & Process Items and Stock Deductions
		var subtotal float64 = 0
		var billItems []models.BillItem

		for _, itemReq := range req.Items {
			targetID := itemReq.ItemID
			if targetID == 0 {
				targetID = itemReq.SparePartID
			}
			if targetID == 0 {
				return fmt.Errorf("item_id is required for all bill items")
			}

			var item models.Item
			if err := tx.Preload("Rack").First(&item, targetID).Error; err != nil {
				return fmt.Errorf("item ID %d not found", targetID)
			}

			if item.Quantity < itemReq.Quantity {
				return fmt.Errorf("insufficient stock for '%s' (SKU: %s). Available: %d, Requested: %d",
					item.Name, item.SKU, item.Quantity, itemReq.Quantity)
			}

			// Deduct stock
			item.Quantity -= itemReq.Quantity
			if err := tx.Save(&item).Error; err != nil {
				return fmt.Errorf("failed to update stock for '%s': %w", item.Name, err)
			}

			rackInfo := ""
			if item.Rack != nil {
				rackInfo = fmt.Sprintf("%s (%s)", item.Rack.Name, item.Rack.Number)
			}

			itemTotal := math.Round((itemReq.UnitPrice*float64(itemReq.Quantity))*100) / 100
			subtotal += itemTotal

			itemID := item.ID
			billItems = append(billItems, models.BillItem{
				ItemID:      &itemID,
				SparePartID: &itemID,
				PartName:    item.Name,
				PartSKU:     item.SKU,
				RackInfo:    rackInfo,
				UnitPrice:   itemReq.UnitPrice,
				Quantity:    itemReq.Quantity,
				TotalPrice:  itemTotal,
			})
		}

		// 5. Calculate GST & Totals
		gstRate := 0.0
		gstAmount := 0.0
		if req.GSTApplied {
			if req.GSTRate > 0 {
				gstRate = req.GSTRate
			} else if setting.DefaultGSTRate > 0 {
				gstRate = setting.DefaultGSTRate
			} else {
				gstRate = 18.00
			}
			gstAmount = math.Round((subtotal*(gstRate/100.0))*100) / 100
		}

		discount := req.Discount
		if discount < 0 {
			discount = 0
		}
		totalAmount := math.Round((subtotal+gstAmount-discount)*100) / 100
		if totalAmount < 0 {
			totalAmount = 0
		}

		var custIDPtr *uint
		if customer.ID > 0 {
			custIDPtr = &customer.ID
		}

		paymentMode := req.PaymentMode
		if paymentMode == "" {
			paymentMode = "cash"
		}

		createdBill = models.Bill{
			InvoiceNumber:        invoiceNumber,
			CustomerID:           custIDPtr,
			CustomerBillSequence: customerSeq,
			UserID:               userID,
			Subtotal:             subtotal,
			GSTApplied:           req.GSTApplied,
			GSTRate:              gstRate,
			GSTAmount:            gstAmount,
			Discount:             discount,
			TotalAmount:          totalAmount,
			PaymentMode:          paymentMode,
			Notes:                req.Notes,
			Items:                billItems,
			CreatedAt:            time.Now(),
		}

		if err := tx.Create(&createdBill).Error; err != nil {
			return fmt.Errorf("failed to save invoice: %w", err)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preload associations for returning
	database.DB.Preload("Customer").Preload("User").Preload("Items").First(&createdBill, createdBill.ID)
	c.JSON(http.StatusCreated, createdBill)
}

func ListBills(c *gin.Context) {
	query := database.DB.Preload("Customer").Preload("User").Preload("Items")

	customerID := c.Query("customer_id")
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	search := strings.TrimSpace(c.Query("q"))
	if search != "" {
		pat := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN customers ON customers.id = bills.customer_id").
			Where("LOWER(bills.invoice_number) LIKE ? OR LOWER(customers.name) LIKE ? OR customers.mobile LIKE ?",
				pat, pat, pat)
	}

	var bills []models.Bill
	if err := query.Order("bills.created_at desc").Find(&bills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	c.JSON(http.StatusOK, bills)
}

func GetBill(c *gin.Context) {
	id := c.Param("id")
	var bill models.Bill
	if err := database.DB.Preload("Customer").Preload("User").Preload("Items").First(&bill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bill not found"})
		return
	}
	c.JSON(http.StatusOK, bill)
}

func GetBillPDF(c *gin.Context) {
	id := c.Param("id")
	var bill models.Bill
	if err := database.DB.Preload("Customer").Preload("User").Preload("Items").First(&bill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bill not found"})
		return
	}

	var setting models.Setting
	database.DB.First(&setting)

	pdfBytes, err := pdf.GenerateInvoicePDF(&bill, &setting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate invoice PDF: %v", err)})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=Invoice-%s.pdf", bill.InvoiceNumber))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
