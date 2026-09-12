package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/inventory/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() (*gorm.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if dbType == "sqlite" {
		dbPath := os.Getenv("SQLITE_PATH")
		if dbPath == "" {
			dbPath = "inventory.db"
		}
		log.Printf("Connecting to SQLite database at %s", dbPath)
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
	} else {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "postgrespassword"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "inventory_db"
		}
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			host, user, password, dbname, port, sslmode)

		log.Printf("Connecting to PostgreSQL at %s:%s dbname=%s...", host, port, dbname)

		// Retry loop for database readiness (useful when spinning up via docker-compose)
		for i := 1; i <= 10; i++ {
			db, err = gorm.Open(postgres.Open(dsn), gormConfig)
			if err == nil {
				sqlDB, errSql := db.DB()
				if errSql == nil {
					if errPing := sqlDB.Ping(); errPing == nil {
						break
					}
				}
			}
			log.Printf("Attempt %d/10: Database not ready yet (%v). Waiting 2s...", i, err)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	log.Println("Database connection established successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")
	err := db.AutoMigrate(
		&models.User{},
		&models.Setting{},
		&models.Rack{},
		&models.SparePart{},
		&models.Customer{},
		&models.Bill{},
		&models.BillItem{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Database migrations completed successfully")
	return nil
}

func SeedDefaultData(db *gorm.DB) error {
	// 1. Seed Settings
	var settingCount int64
	db.Model(&models.Setting{}).Count(&settingCount)
	if settingCount == 0 {
		defaultSetting := models.Setting{
			ShopName:       "Spare Parts Pro",
			ShopAddress:    "123 Industrial Hub, Motor Market, City",
			ShopPhone:      "+91 98765 43210",
			ShopEmail:      "support@sparepartspro.local",
			ShopGSTIN:      "27AAACG0000A1Z5",
			DefaultGSTRate: 18.00,
			InvoicePrefix:  "INV",
		}
		if err := db.Create(&defaultSetting).Error; err != nil {
			log.Printf("Failed to seed default settings: %v", err)
		} else {
			log.Println("Default system settings seeded.")
		}
	}

	// 2. Seed Default Admin User
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
	if adminCount == 0 {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser := models.User{
			Username:     "admin",
			PasswordHash: string(hashedPass),
			FullName:     "System Administrator",
			Role:         models.RoleAdmin,
			IsActive:     true,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("Failed to seed admin user: %v", err)
		} else {
			log.Println("Default admin user created: admin / admin123")
		}
	}

	// 3. Seed Default Seller User
	var sellerCount int64
	db.Model(&models.User{}).Where("role = ?", models.RoleSeller).Count(&sellerCount)
	if sellerCount == 0 {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("seller123"), bcrypt.DefaultCost)
		sellerUser := models.User{
			Username:     "seller",
			PasswordHash: string(hashedPass),
			FullName:     "Main Store Seller",
			Role:         models.RoleSeller,
			IsActive:     true,
		}
		if err := db.Create(&sellerUser).Error; err != nil {
			log.Printf("Failed to seed seller user: %v", err)
		} else {
			log.Println("Default seller user created: seller / seller123")
		}
	}

	// 4. Seed Demo Racks if empty
	var rackCount int64
	db.Model(&models.Rack{}).Count(&rackCount)
	if rackCount == 0 {
		racks := []models.Rack{
			{Name: "Rack A", Number: "A1", Description: "Engine & Transmission Components"},
			{Name: "Rack A", Number: "A2", Description: "Braking & Suspension Parts"},
			{Name: "Rack B", Number: "B1", Description: "Electrical & Lighting Systems"},
			{Name: "Rack B", Number: "B2", Description: "Filters, Plugs & Consumables"},
			{Name: "Rack C", Number: "C1", Description: "Exhaust & Cooling Parts"},
		}
		db.Create(&racks)
		log.Println("Initial sample racks created.")

		// 5. Seed Demo Spare Parts
		var rackA1, rackA2, rackB1, rackB2 models.Rack
		db.Where("name = ? AND number = ?", "Rack A", "A1").First(&rackA1)
		db.Where("name = ? AND number = ?", "Rack A", "A2").First(&rackA2)
		db.Where("name = ? AND number = ?", "Rack B", "B1").First(&rackB1)
		db.Where("name = ? AND number = ?", "Rack B", "B2").First(&rackB2)

		parts := []models.SparePart{
			{
				SKU:           "SP-BRK-001",
				Name:          "Ceramic Brake Pads (Front Set)",
				Description:   "High durability ceramic front brake pads for sedans & SUVs",
				Price:         1450.00,
				CostPrice:     980.00,
				GSTApplicable: true,
				GSTRate:       18.00,
				Quantity:      28,
				MinStockAlert: 5,
				RackID:        &rackA2.ID,
			},
			{
				SKU:           "SP-FLT-002",
				Name:          "Synthetic Engine Oil Filter - Premium",
				Description:   "Spin-on oil filter with high efficiency synthetic media",
				Price:         350.00,
				CostPrice:     210.00,
				GSTApplicable: true,
				GSTRate:       18.00,
				Quantity:      45,
				MinStockAlert: 10,
				RackID:        &rackB2.ID,
			},
			{
				SKU:           "SP-SPK-003",
				Name:          "Iridium Spark Plug (Pack of 4)",
				Description:   "Laser iridium ultra long life spark plugs",
				Price:         1800.00,
				CostPrice:     1250.00,
				GSTApplicable: true,
				GSTRate:       18.00,
				Quantity:      15,
				MinStockAlert: 4,
				RackID:        &rackB2.ID,
			},
			{
				SKU:           "SP-LGT-004",
				Name:          "H7 LED Headlight Bulb Set (6000K)",
				Description:   "120W Super Bright White High/Low Beam",
				Price:         2200.00,
				CostPrice:     1400.00,
				GSTApplicable: true,
				GSTRate:       18.00,
				Quantity:      8,
				MinStockAlert: 3,
				RackID:        &rackB1.ID,
			},
			{
				SKU:           "SP-ENG-005",
				Name:          "Serpentine Drive Belt",
				Description:   "Heavy duty EPDM ribbed accessory drive belt",
				Price:         850.00,
				CostPrice:     520.00,
				GSTApplicable: true,
				GSTRate:       18.00,
				Quantity:      3, // Low stock demo!
				MinStockAlert: 5,
				RackID:        &rackA1.ID,
			},
		}
		db.Create(&parts)
		log.Println("Initial sample spare parts created.")
	}

	// 6. Seed Sample Customers
	var customerCount int64
	db.Model(&models.Customer{}).Count(&customerCount)
	if customerCount == 0 {
		customers := []models.Customer{
			{Name: "Rajesh Sharma", Mobile: "9823012345", Email: "rajesh@gmail.com", Address: "Shop #14, Auto Line, City"},
			{Name: "Amit Patel", Mobile: "9876543210", Email: "amit.patel@hotmail.com", Address: "Flat 402, Green Heights"},
			{Name: "Priya Sundaram", Mobile: "9123456789", Email: "priya.s@yahoo.com", Address: "Plot 89, Sector 5"},
		}
		db.Create(&customers)
		log.Println("Initial sample customers created.")
	}

	return nil
}
