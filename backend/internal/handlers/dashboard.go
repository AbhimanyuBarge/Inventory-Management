package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type DashboardStatsResponse struct {
	TotalItemsCount     int64         `json:"total_items_count"`
	TotalPartsCount     int64         `json:"total_parts_count"` // alias
	TotalStockQuantity  int64         `json:"total_stock_quantity"`
	LowStockCount       int64         `json:"low_stock_count"`
	TotalRacksCount     int64         `json:"total_racks_count"`
	TotalCustomersCount int64         `json:"total_customers_count"`
	TodaySales          float64       `json:"today_sales"`
	TodayBillsCount     int64         `json:"today_bills_count"`
	WeekSales           float64       `json:"week_sales"`
	MonthSales          float64       `json:"month_sales"`
	TotalRevenue        float64       `json:"total_revenue"`
	TotalReturnsCount   int64         `json:"total_returns_count"`
	TotalRefundAmount   float64       `json:"total_refund_amount"`
	NetRevenue          float64       `json:"net_revenue"`
	LowStockItems       []models.Item `json:"low_stock_items"`
	RecentBills         []models.Bill `json:"recent_bills"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStatsResponse

	// Total unique items
	database.DB.Model(&models.Item{}).Count(&stats.TotalItemsCount)
	stats.TotalPartsCount = stats.TotalItemsCount

	// Total quantity of all items
	type SumResult struct {
		TotalQty float64
	}
	var sumRes SumResult
	database.DB.Model(&models.Item{}).Select("COALESCE(SUM(quantity), 0) as total_qty").Scan(&sumRes)
	stats.TotalStockQuantity = int64(sumRes.TotalQty)

	// Low stock count & list
	database.DB.Model(&models.Item{}).Where("quantity <= min_stock_alert").Count(&stats.LowStockCount)
	database.DB.Preload("Rack").Where("quantity <= min_stock_alert").Order("quantity asc").Limit(8).Find(&stats.LowStockItems)

	// Total racks
	database.DB.Model(&models.Rack{}).Count(&stats.TotalRacksCount)

	// Total customers
	database.DB.Model(&models.Customer{}).Count(&stats.TotalCustomersCount)

	// Sales Calculations
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfToday.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type SalesSum struct {
		Sum float64
	}
	var salesRes SalesSum

	// Today's Sales & Count
	database.DB.Model(&models.Bill{}).Where("created_at >= ?", startOfToday).Select("COALESCE(SUM(total_amount), 0) as sum").Scan(&salesRes)
	stats.TodaySales = salesRes.Sum
	database.DB.Model(&models.Bill{}).Where("created_at >= ?", startOfToday).Count(&stats.TodayBillsCount)

	// Week's Sales
	database.DB.Model(&models.Bill{}).Where("created_at >= ?", startOfWeek).Select("COALESCE(SUM(total_amount), 0) as sum").Scan(&salesRes)
	stats.WeekSales = salesRes.Sum

	// Month's Sales
	database.DB.Model(&models.Bill{}).Where("created_at >= ?", startOfMonth).Select("COALESCE(SUM(total_amount), 0) as sum").Scan(&salesRes)
	stats.MonthSales = salesRes.Sum

	// Lifetime Total Revenue
	database.DB.Model(&models.Bill{}).Select("COALESCE(SUM(total_amount), 0) as sum").Scan(&salesRes)
	stats.TotalRevenue = salesRes.Sum

	// Return Statistics
	database.DB.Model(&models.ReturnBill{}).Count(&stats.TotalReturnsCount)
	var refundRes SalesSum
	database.DB.Model(&models.ReturnBill{}).Select("COALESCE(SUM(total_refund_amount), 0) as sum").Scan(&refundRes)
	stats.TotalRefundAmount = refundRes.Sum

	stats.NetRevenue = stats.TotalRevenue - stats.TotalRefundAmount

	// Recent Bills with full associations
	database.DB.Preload("Customer").Preload("User").Preload("Items").Order("created_at desc").Limit(6).Find(&stats.RecentBills)

	c.JSON(http.StatusOK, stats)
}
