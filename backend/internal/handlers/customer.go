package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func SearchCustomers(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, []models.Customer{})
		return
	}

	pattern := "%" + strings.ToLower(q) + "%"
	var customers []models.Customer
	database.DB.Where("LOWER(name) LIKE ? OR mobile LIKE ?", pattern, pattern).
		Order("name asc").
		Limit(10).
		Find(&customers)

	c.JSON(http.StatusOK, customers)
}

func ListCustomers(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))
	query := database.DB

	if search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR mobile LIKE ?", pattern, pattern)
	}

	var customers []models.Customer
	if err := query.Order("updated_at desc, name asc").Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	c.JSON(http.StatusOK, customers)
}

func GetCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var bills []models.Bill
	database.DB.Where("customer_id = ?", customer.ID).Order("created_at desc").Find(&bills)

	c.JSON(http.StatusOK, gin.H{
		"customer": customer,
		"bills":    bills,
	})
}

func CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := models.Customer{
		Name:    req.Name,
		Mobile:  req.Mobile,
		Email:   req.Email,
		Address: req.Address,
	}

	if err := database.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer.Name = req.Name
	customer.Mobile = req.Mobile
	customer.Email = req.Email
	customer.Address = req.Address

	if err := database.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}
