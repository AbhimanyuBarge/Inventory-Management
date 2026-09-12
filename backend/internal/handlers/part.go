package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type CreatePartRequest struct {
	SKU           string  `json:"sku" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gte=0"`
	CostPrice     float64 `json:"cost_price" binding:"gte=0"`
	GSTApplicable bool    `json:"gst_applicable"`
	GSTRate       float64 `json:"gst_rate" binding:"gte=0"`
	Quantity      int     `json:"quantity" binding:"gte=0"`
	MinStockAlert int     `json:"min_stock_alert" binding:"gte=0"`
	RackID        *uint   `json:"rack_id"`
}

type UpdatePartRequest struct {
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"omitempty,gte=0"`
	CostPrice     float64 `json:"cost_price" binding:"omitempty,gte=0"`
	GSTApplicable *bool   `json:"gst_applicable"`
	GSTRate       float64 `json:"gst_rate" binding:"omitempty,gte=0"`
	Quantity      *int    `json:"quantity" binding:"omitempty,gte=0"`
	MinStockAlert *int    `json:"min_stock_alert" binding:"omitempty,gte=0"`
	RackID        *uint   `json:"rack_id"`
}

type StockAdjustmentRequest struct {
	Type   string `json:"type" binding:"required,oneof=add subtract set"`
	Amount int    `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason"`
}

func ListParts(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))
	rackID := c.Query("rack_id")
	lowStockOnly := c.Query("low_stock") == "true"

	query := database.DB.Preload("Rack")

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(description) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if rackID != "" && rackID != "0" {
		query = query.Where("rack_id = ?", rackID)
	}

	if lowStockOnly {
		query = query.Where("quantity <= min_stock_alert")
	}

	var parts []models.SparePart
	if err := query.Order("name asc").Find(&parts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spare parts"})
		return
	}

	c.JSON(http.StatusOK, parts)
}

func GetPart(c *gin.Context) {
	id := c.Param("id")
	var part models.SparePart
	if err := database.DB.Preload("Rack").First(&part, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spare part not found"})
		return
	}
	c.JSON(http.StatusOK, part)
}

func CreatePart(c *gin.Context) {
	var req CreatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.SparePart
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

	part := models.SparePart{
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
	}

	if err := database.DB.Create(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create spare part"})
		return
	}

	database.DB.Preload("Rack").First(&part, part.ID)
	c.JSON(http.StatusCreated, part)
}

func UpdatePart(c *gin.Context) {
	id := c.Param("id")
	var part models.SparePart
	if err := database.DB.First(&part, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spare part not found"})
		return
	}

	var req UpdatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SKU != "" && req.SKU != part.SKU {
		var existing models.SparePart
		if err := database.DB.Where("sku = ? AND id <> ?", req.SKU, part.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists on another part"})
			return
		}
		part.SKU = req.SKU
	}

	if req.Name != "" {
		part.Name = req.Name
	}
	if req.Description != "" {
		part.Description = req.Description
	}
	if req.Price > 0 {
		part.Price = req.Price
	}
	if req.CostPrice >= 0 {
		part.CostPrice = req.CostPrice
	}
	if req.GSTApplicable != nil {
		part.GSTApplicable = *req.GSTApplicable
	}
	if req.GSTRate >= 0 {
		part.GSTRate = req.GSTRate
	}
	if req.Quantity != nil {
		part.Quantity = *req.Quantity
	}
	if req.MinStockAlert != nil {
		part.MinStockAlert = *req.MinStockAlert
	}
	part.RackID = req.RackID

	if err := database.DB.Save(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update spare part"})
		return
	}

	database.DB.Preload("Rack").First(&part, part.ID)
	c.JSON(http.StatusOK, part)
}

func AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var part models.SparePart
	if err := database.DB.First(&part, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spare part not found"})
		return
	}

	var req StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Type {
	case "add":
		part.Quantity += req.Amount
	case "subtract":
		if part.Quantity < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot subtract %d from stock of %d", req.Amount, part.Quantity)})
			return
		}
		part.Quantity -= req.Amount
	case "set":
		part.Quantity = req.Amount
	}

	if err := database.DB.Save(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to adjust stock"})
		return
	}

	database.DB.Preload("Rack").First(&part, part.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock adjusted successfully",
		"part":    part,
	})
}

func DeletePart(c *gin.Context) {
	id := c.Param("id")
	var part models.SparePart
	if err := database.DB.First(&part, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spare part not found"})
		return
	}

	if err := database.DB.Delete(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete spare part"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spare part deleted successfully"})
}
