package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type CreateVendorRequest struct {
	Name      string `json:"name" binding:"required"`
	ContactNo string `json:"contact_no"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

type UpdateVendorRequest struct {
	Name      string `json:"name"`
	ContactNo string `json:"contact_no"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

// ListVendors returns all vendors with their linked item counts and search filter
func ListVendors(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	var vendors []models.Vendor
	query := database.DB.Model(&models.Vendor{})

	if search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR contact_no LIKE ? OR LOWER(email) LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Order("name asc").Find(&vendors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vendors"})
		return
	}

	// Calculate item counts for each vendor
	for i := range vendors {
		var count int64
		database.DB.Model(&models.Item{}).Where("vendor_id = ?", vendors[i].ID).Count(&count)
		vendors[i].ItemsCount = count
	}

	c.JSON(http.StatusOK, vendors)
}

// SearchVendors is a lightweight endpoint for autocomplete during item creation
func SearchVendors(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	var vendors []models.Vendor

	dbQuery := database.DB.Model(&models.Vendor{})
	if query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR contact_no LIKE ?", pattern, pattern)
	}

	if err := dbQuery.Limit(10).Order("name asc").Find(&vendors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search vendors"})
		return
	}

	c.JSON(http.StatusOK, vendors)
}

// GetVendor returns a single vendor with all their linked items
func GetVendor(c *gin.Context) {
	id := c.Param("id")
	var vendor models.Vendor
	if err := database.DB.Preload("Items.Rack").First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	var count int64
	database.DB.Model(&models.Item{}).Where("vendor_id = ?", vendor.ID).Count(&count)
	vendor.ItemsCount = count

	c.JSON(http.StatusOK, vendor)
}

// CreateVendor adds a new vendor
func CreateVendor(c *gin.Context) {
	var req CreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nameTrimmed := strings.TrimSpace(req.Name)
	if nameTrimmed == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vendor name is required"})
		return
	}

	var existing models.Vendor
	if err := database.DB.Where("LOWER(name) = ?", strings.ToLower(nameTrimmed)).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vendor with this name already exists"})
		return
	}

	vendor := models.Vendor{
		Name:      nameTrimmed,
		ContactNo: strings.TrimSpace(req.ContactNo),
		Email:     strings.TrimSpace(req.Email),
		Address:   strings.TrimSpace(req.Address),
	}

	if err := database.DB.Create(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vendor"})
		return
	}

	c.JSON(http.StatusCreated, vendor)
}

// UpdateVendor updates existing vendor details
func UpdateVendor(c *gin.Context) {
	id := c.Param("id")
	var vendor models.Vendor
	if err := database.DB.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	var req UpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		nameTrimmed := strings.TrimSpace(req.Name)
		var existing models.Vendor
		if err := database.DB.Where("LOWER(name) = ? AND id != ?", strings.ToLower(nameTrimmed), vendor.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Another vendor with this name already exists"})
			return
		}
		vendor.Name = nameTrimmed
	}

	vendor.ContactNo = strings.TrimSpace(req.ContactNo)
	vendor.Email = strings.TrimSpace(req.Email)
	vendor.Address = strings.TrimSpace(req.Address)

	if err := database.DB.Save(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor"})
		return
	}

	c.JSON(http.StatusOK, vendor)
}

// DeleteVendor deletes a vendor ONLY if no items are linked to them
func DeleteVendor(c *gin.Context) {
	id := c.Param("id")
	var vendor models.Vendor
	if err := database.DB.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	var itemCount int64
	database.DB.Model(&models.Item{}).Where("vendor_id = ?", vendor.ID).Count(&itemCount)
	if itemCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete vendor: this vendor has items linked to them. Please reassign or delete the items first.",
		})
		return
	}

	if err := database.DB.Delete(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vendor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vendor deleted successfully"})
}
