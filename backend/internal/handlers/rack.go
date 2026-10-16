package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
	"github.com/inventory/backend/internal/pdf"
)

type CreateRackRequest struct {
	Name        string `json:"name" binding:"required"`
	Number      string `json:"number" binding:"required"`
	Description string `json:"description"`
}

type UpdateRackRequest struct {
	Name        string `json:"name"`
	Number      string `json:"number"`
	Description string `json:"description"`
}

func ListRacks(c *gin.Context) {
	var racks []models.Rack
	if err := database.DB.Order("name asc, number asc").Find(&racks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch racks"})
		return
	}

	// Attach count of items in each rack
	for i := range racks {
		var count int64
		database.DB.Model(&models.Item{}).Where("rack_id = ?", racks[i].ID).Count(&count)
		racks[i].ItemsCount = count
		racks[i].PartsCount = count
	}

	c.JSON(http.StatusOK, racks)
}

func GetRack(c *gin.Context) {
	id := c.Param("id")
	var rack models.Rack
	if err := database.DB.Preload("Items").First(&rack, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rack not found"})
		return
	}
	rack.ItemsCount = int64(len(rack.Items))
	rack.PartsCount = rack.ItemsCount
	c.JSON(http.StatusOK, rack)
}

func CreateRack(c *gin.Context) {
	var req CreateRackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Rack
	if err := database.DB.Where("name = ? AND number = ?", req.Name, req.Number).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A rack with this Name and Number already exists"})
		return
	}

	rack := models.Rack{
		Name:        req.Name,
		Number:      req.Number,
		Description: req.Description,
	}

	if err := database.DB.Create(&rack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rack"})
		return
	}

	c.JSON(http.StatusCreated, rack)
}

func UpdateRack(c *gin.Context) {
	id := c.Param("id")
	var rack models.Rack
	if err := database.DB.First(&rack, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rack not found"})
		return
	}

	var req UpdateRackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		rack.Name = req.Name
	}
	if req.Number != "" {
		rack.Number = req.Number
	}
	if req.Description != "" {
		rack.Description = req.Description
	}

	if err := database.DB.Save(&rack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rack"})
		return
	}

	c.JSON(http.StatusOK, rack)
}

func DeleteRack(c *gin.Context) {
	id := c.Param("id")
	var rack models.Rack
	if err := database.DB.First(&rack, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rack not found"})
		return
	}

	// Unlink items from this rack
	database.DB.Model(&models.Item{}).Where("rack_id = ?", rack.ID).Update("rack_id", nil)

	if err := database.DB.Delete(&rack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rack"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rack deleted successfully"})
}

func GetRackLabelPDF(c *gin.Context) {
	id := c.Param("id")
	var rack models.Rack
	if err := database.DB.First(&rack, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rack not found"})
		return
	}

	var items []models.Item
	database.DB.Where("rack_id = ?", rack.ID).Order("sku asc").Find(&items)

	pdfBytes, err := pdf.GenerateRackLabelPDF(&rack, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate label: %v", err)})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=rack-label-%s-%s.pdf", rack.Name, rack.Number))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
