package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type CreateItemRequest struct {
	SKU             string  `json:"sku" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	Description     string  `json:"description"`
	Price           float64 `json:"price" binding:"required,gte=0"`
	CostPrice       float64 `json:"cost_price" binding:"gte=0"`
	GSTApplicable   bool    `json:"gst_applicable"`
	GSTRate         float64 `json:"gst_rate" binding:"gte=0"`
	Quantity        int     `json:"quantity" binding:"gte=0"`
	MinStockAlert   int     `json:"min_stock_alert" binding:"gte=0"`
	RackID          *uint   `json:"rack_id"`
	VendorID        *uint   `json:"vendor_id"`
	VendorName      string  `json:"vendor_name"`
	VendorContactNo string  `json:"vendor_contact_no"`
}

type UpdateItemRequest struct {
	SKU             string  `json:"sku"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Price           float64 `json:"price" binding:"omitempty,gte=0"`
	CostPrice       float64 `json:"cost_price" binding:"omitempty,gte=0"`
	GSTApplicable   *bool   `json:"gst_applicable"`
	GSTRate         float64 `json:"gst_rate" binding:"omitempty,gte=0"`
	Quantity        *int    `json:"quantity" binding:"omitempty,gte=0"`
	MinStockAlert   *int    `json:"min_stock_alert" binding:"omitempty,gte=0"`
	RackID          *uint   `json:"rack_id"`
	VendorID        *uint   `json:"vendor_id"`
	VendorName      string  `json:"vendor_name"`
	VendorContactNo string  `json:"vendor_contact_no"`
}

func resolveVendor(vendorID *uint, vendorName, vendorContactNo string) *uint {
	if vendorID != nil && *vendorID > 0 {
		return vendorID
	}
	nameTrimmed := strings.TrimSpace(vendorName)
	if nameTrimmed == "" {
		return nil
	}
	var vendor models.Vendor
	if err := database.DB.Where("LOWER(name) = ?", strings.ToLower(nameTrimmed)).First(&vendor).Error; err == nil {
		if strings.TrimSpace(vendorContactNo) != "" && vendor.ContactNo == "" {
			vendor.ContactNo = strings.TrimSpace(vendorContactNo)
			database.DB.Save(&vendor)
		}
		return &vendor.ID
	}

	newVendor := models.Vendor{
		Name:      nameTrimmed,
		ContactNo: strings.TrimSpace(vendorContactNo),
	}
	if err := database.DB.Create(&newVendor).Error; err == nil {
		return &newVendor.ID
	}
	return nil
}

type StockAdjustmentRequest struct {
	Type   string `json:"type" binding:"required,oneof=add subtract set"`
	Amount int    `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason"`
}

// Aliases for compatibility
type CreatePartRequest = CreateItemRequest
type UpdatePartRequest = UpdateItemRequest

func ListItems(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))
	rackID := c.Query("rack_id")
	vendorID := c.Query("vendor_id")
	lowStockOnly := c.Query("low_stock") == "true"

	query := database.DB.Preload("Rack").Preload("Vendor")

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(items.name) LIKE ? OR LOWER(items.sku) LIKE ? OR LOWER(items.description) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if rackID != "" && rackID != "0" {
		query = query.Where("rack_id = ?", rackID)
	}

	if vendorID != "" && vendorID != "0" {
		query = query.Where("vendor_id = ?", vendorID)
	}

	if lowStockOnly {
		query = query.Where("quantity <= min_stock_alert")
	}

	var items []models.Item
	if err := query.Order("name asc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item
	if err := database.DB.Preload("Rack").Preload("Vendor").First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Item
	if err := database.DB.Where("sku = ?", req.SKU).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists"})
		return
	}

	if req.MinStockAlert == 0 {
		req.MinStockAlert = 5
	}
	if req.GSTRate == 0 && req.GSTApplicable {
		var setting models.Setting
		if err := database.DB.First(&setting).Error; err == nil && setting.DefaultGSTRate > 0 {
			req.GSTRate = setting.DefaultGSTRate
		} else {
			req.GSTRate = 18.00
		}
	}

	resolvedVendorID := resolveVendor(req.VendorID, req.VendorName, req.VendorContactNo)

	item := models.Item{
		SKU:           req.SKU,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		CostPrice:     req.CostPrice,
		GSTApplicable: req.GSTApplicable,
		GSTRate:       req.GSTRate,
		Quantity:      req.Quantity,
		MinStockAlert: req.MinStockAlert,
		RackID:        req.RackID,
		VendorID:      resolvedVendorID,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	database.DB.Preload("Rack").Preload("Vendor").First(&item, item.ID)
	c.JSON(http.StatusCreated, item)
}

func UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SKU != "" && req.SKU != item.SKU {
		var existing models.Item
		if err := database.DB.Where("sku = ? AND id <> ?", req.SKU, item.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists on another item"})
			return
		}
		item.SKU = req.SKU
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Price > 0 {
		item.Price = req.Price
	}
	if req.CostPrice >= 0 {
		item.CostPrice = req.CostPrice
	}
	if req.GSTApplicable != nil {
		item.GSTApplicable = *req.GSTApplicable
	}
	if req.GSTRate >= 0 {
		item.GSTRate = req.GSTRate
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.MinStockAlert != nil {
		item.MinStockAlert = *req.MinStockAlert
	}
	item.RackID = req.RackID

	if req.VendorID != nil || req.VendorName != "" {
		item.VendorID = resolveVendor(req.VendorID, req.VendorName, req.VendorContactNo)
	}

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	database.DB.Preload("Rack").Preload("Vendor").First(&item, item.ID)
	c.JSON(http.StatusOK, item)
}

func AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var item models.Item
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var req StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Type {
	case "add":
		item.Quantity += req.Amount
	case "subtract":
		if item.Quantity < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot subtract %d from stock of %d", req.Amount, item.Quantity)})
			return
		}
		item.Quantity -= req.Amount
	case "set":
		item.Quantity = req.Amount
	}

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to adjust stock"})
		return
	}

	database.DB.Preload("Rack").First(&item, item.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock adjusted successfully",
		"item":    item,
		"part":    item, // alias
	})
}

func DeleteItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// Backward-compatible function wrappers
func ListParts(c *gin.Context)   { ListItems(c) }
func GetPart(c *gin.Context)     { GetItem(c) }
func CreatePart(c *gin.Context)  { CreateItem(c) }
func UpdatePart(c *gin.Context)  { UpdateItem(c) }
func DeletePart(c *gin.Context)  { DeleteItem(c) }
