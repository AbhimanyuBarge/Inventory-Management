package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type UpdateSettingsRequest struct {
	ShopName       string  `json:"shop_name"`
	ShopAddress    string  `json:"shop_address"`
	ShopPhone      string  `json:"shop_phone"`
	ShopEmail      string  `json:"shop_email"`
	ShopGSTIN      string  `json:"shop_gstin"`
	DefaultGSTRate float64 `json:"default_gst_rate"`
	InvoicePrefix  string  `json:"invoice_prefix"`
}

func GetSettings(c *gin.Context) {
	var setting models.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		c.JSON(http.StatusOK, models.Setting{
			ShopName:       "Spare Parts Pro",
			DefaultGSTRate: 18.00,
			InvoicePrefix:  "INV",
		})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var setting models.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		setting = models.Setting{}
	}

	setting.ShopName = req.ShopName
	setting.ShopAddress = req.ShopAddress
	setting.ShopPhone = req.ShopPhone
	setting.ShopEmail = req.ShopEmail
	setting.ShopGSTIN = req.ShopGSTIN
	setting.DefaultGSTRate = req.DefaultGSTRate
	if req.InvoicePrefix != "" {
		setting.InvoicePrefix = req.InvoicePrefix
	}

	if err := database.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, setting)
}
