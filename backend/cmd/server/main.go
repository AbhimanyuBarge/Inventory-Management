package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/handlers"
	"github.com/inventory/backend/internal/middleware"
	"github.com/joho/godotenv"
)

// CORSMiddleware handles cross-origin requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH, HEAD")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	// Load .env if present
	_ = godotenv.Load()

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Auto migrate schema
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Seed default admin, seller, settings and sample data
	if err := database.SeedDefaultData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	router := gin.Default()

	// CORS configuration
	router.Use(CORSMiddleware())

	// Health Check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "spare-parts-inventory-api",
		})
	})

	// Public Routes
	api := router.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)

		// Direct PDF download without header (supports token in query if needed, or open)
		api.GET("/bills/:id/pdf", handlers.GetBillPDF)
		api.GET("/returns/:id/pdf", handlers.GetReturnPDF)
		api.GET("/racks/:id/label-pdf", handlers.GetRackLabelPDF)
	}

	// Protected Routes (Authenticated Admin & Seller)
	auth := api.Group("")
	auth.Use(middleware.RequireAuth())
	{
		// Profile
		auth.GET("/auth/me", handlers.GetProfile)
		auth.PUT("/auth/password", handlers.UpdatePassword)

		// Dashboard Statistics
		auth.GET("/dashboard/stats", handlers.GetDashboardStats)

		// Inventory Items
		auth.GET("/items", handlers.ListItems)
		auth.GET("/items/:id", handlers.GetItem)
		auth.POST("/items", handlers.CreateItem)
		auth.PUT("/items/:id", handlers.UpdateItem)
		auth.POST("/items/:id/adjust", handlers.AdjustStock)
		auth.DELETE("/items/:id", handlers.DeleteItem)

		// Legacy Parts aliases
		auth.GET("/parts", handlers.ListItems)
		auth.GET("/parts/:id", handlers.GetItem)
		auth.POST("/parts", handlers.CreateItem)
		auth.PUT("/parts/:id", handlers.UpdateItem)
		auth.POST("/parts/:id/adjust", handlers.AdjustStock)
		auth.DELETE("/parts/:id", handlers.DeleteItem)

		// Racks
		auth.GET("/racks", handlers.ListRacks)
		auth.GET("/racks/:id", handlers.GetRack)
		auth.POST("/racks", handlers.CreateRack)
		auth.PUT("/racks/:id", handlers.UpdateRack)
		auth.DELETE("/racks/:id", handlers.DeleteRack)

		// Vendors
		auth.GET("/vendors", handlers.ListVendors)
		auth.GET("/vendors/search", handlers.SearchVendors)
		auth.GET("/vendors/:id", handlers.GetVendor)
		auth.POST("/vendors", handlers.CreateVendor)
		auth.PUT("/vendors/:id", handlers.UpdateVendor)
		auth.DELETE("/vendors/:id", handlers.DeleteVendor)

		// Customers
		auth.GET("/customers", handlers.ListCustomers)
		auth.GET("/customers/search", handlers.SearchCustomers)
		auth.GET("/customers/:id", handlers.GetCustomer)
		auth.POST("/customers", handlers.CreateCustomer)
		auth.PUT("/customers/:id", handlers.UpdateCustomer)

		// Bills / Point-of-Sale
		auth.GET("/bills", handlers.ListBills)
		auth.GET("/bills/:id", handlers.GetBill)
		auth.GET("/bills/:id/returnable", handlers.GetBillReturnable)
		auth.POST("/bills", handlers.CreateBill)

		// Returns / Refunds
		auth.GET("/returns", handlers.ListReturns)
		auth.GET("/returns/:id", handlers.GetReturn)
		auth.POST("/returns", handlers.CreateReturn)

		// System Settings (Read-only for Seller, Admin can write)
		auth.GET("/settings", handlers.GetSettings)

		// Admin-Only Routes
		admin := auth.Group("")
		admin.Use(middleware.RequireAdmin())
		{
			admin.PUT("/settings", handlers.UpdateSettings)
			admin.POST("/settings/reset-mock-data", handlers.ResetMockData)
			admin.GET("/users", handlers.ListUsers)
			admin.POST("/users", handlers.CreateUser)
			admin.PUT("/users/:id", handlers.UpdateUser)
			admin.DELETE("/users/:id", handlers.DeleteUser)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
