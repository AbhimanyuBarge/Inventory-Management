package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inventory/backend/internal/database"
	"github.com/inventory/backend/internal/models"
)

type DashboardStatsResponse struct {
	TotalPartsCount     int64              `json:"total_parts_count"`
	TotalStockQuantity  int64              `json:"total_stock_quantity"`
	LowStockCount       int64              `json:"low_stock_count"`
	TotalRacksCount     int64              `json:"total_racks_count"`
	TotalCustomersCount int64              `json:"total_customers_count"`
	TodaySales          float64            `json:"today_sales"`
	TodayBillsCount     int64              `json:"today_bills_count"`
	WeekSales           float64            `json:"week_sales"`
	MonthSales          float64            `json:"month_sales"`
	TotalRevenue        float64            `json:"total_revenue"`
	LowStockItems       []models.SparePart `json:"low_stock_items"`
	RecentBills         []models.Bill      `json:"recent_bills"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStatsResponse

	// Total unique spare parts
	database.DB.Model(&models.SparePart{}).Count(&stats.TotalPartsCount)

	// Total quantity of all parts
	type SumResult struct {
		TotalQty float64
	}
	var sumRes SumResult
	database.DB.Model(&models.SparePart{}).Select("COALESCE(SUM(quantity), 0) as total_qty").Scan(&sumRes)
	stats.TotalStockQuantity = int64(sumRes.TotalQty)

	// Low stock count & list
	database.DB.Model(&models.SparePart{}).Where("quantity <= min_stock_alert").Count(&stats.LowStockCount)
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

	// Recent Bills
	database.DB.Preload("Customer").Preload("User").Order("created_at desc").Limit(6).Find(&stats.RecentBills)

	c.JSON(http.StatusOK, stats)
}
