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
	// Cleanup any stale legacy constraints from previous schema iterations
	db.Exec("ALTER TABLE bill_items DROP CONSTRAINT IF EXISTS fk_bill_items_spare_part;")
	db.Exec("ALTER TABLE return_items DROP CONSTRAINT IF EXISTS fk_return_items_spare_part;")
	db.Exec("DROP TABLE IF EXISTS spare_parts CASCADE;")

	err := db.AutoMigrate(
		&models.User{},
		&models.Setting{},
		&models.Vendor{},
		&models.Rack{},
		&models.Item{},
		&models.Customer{},
		&models.Bill{},
		&models.BillItem{},
		&models.ReturnBill{},
		&models.ReturnItem{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Ensure cleanup after AutoMigrate as well
	db.Exec("ALTER TABLE bill_items DROP CONSTRAINT IF EXISTS fk_bill_items_spare_part;")
	db.Exec("ALTER TABLE return_items DROP CONSTRAINT IF EXISTS fk_return_items_spare_part;")

	log.Println("Database migrations completed successfully")
	return nil
}

func ResetAndSeedMockData(db *gorm.DB) error {
	log.Println("Clearing database tables and seeding fresh mock data...")

	// 1. Clear existing data in reverse dependency order with identity restart
	db.Exec("TRUNCATE TABLE return_items, return_bills, bill_items, bills, items, vendors, racks, customers, users, settings RESTART IDENTITY CASCADE;")

	// 2. Seed System Settings
	defaultSetting := models.Setting{
		ShopName:       "Apex Auto Spares & Parts",
		ShopAddress:    "Plot 42, Industrial Area, Ring Road, Sector 8",
		ShopPhone:      "+91 98765 43210",
		ShopEmail:      "support@apexautospares.com",
		ShopGSTIN:      "27AAACG0000A1Z5",
		DefaultGSTRate: 18.00,
		InvoicePrefix:  "INV",
	}
	db.Create(&defaultSetting)
	log.Println("✓ Settings seeded.")

	// 3. Seed Users
	hashedPassAdmin, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := models.User{
		Username:     "admin",
		PasswordHash: string(hashedPassAdmin),
		FullName:     "System Administrator",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	db.Create(&adminUser)

	hashedPassSeller, _ := bcrypt.GenerateFromPassword([]byte("seller123"), bcrypt.DefaultCost)
	sellerUser := models.User{
		Username:     "seller",
		PasswordHash: string(hashedPassSeller),
		FullName:     "Main Store Seller",
		Role:         models.RoleSeller,
		IsActive:     true,
	}
	db.Create(&sellerUser)
	log.Println("✓ Users seeded (admin/admin123, seller/seller123).")

	// 4. Seed Vendors
	vendors := []models.Vendor{
		{
			Name:      "Bosch Automotive India",
			ContactNo: "+91 98201 12345",
			Email:     "orders@bosch-india.com",
			Address:   "Plant 3, MIDC Industrial Area, Pune",
		},
		{
			Name:      "Brembo Brakes Private Ltd",
			ContactNo: "+91 98450 67890",
			Email:     "sales@brembo-in.com",
			Address:   "Plot 12, Auto Cluster, Chennai",
		},
		{
			Name:      "NGK Spark Plugs Co",
			ContactNo: "+91 97110 54321",
			Email:     "supply@ngk-spark.in",
			Address:   "Sector 18, Udyog Vihar, Gurgaon",
		},
		{
			Name:      "Castrol India Lubricants",
			ContactNo: "+91 98920 98765",
			Email:     "dist@castrol-india.com",
			Address:   "Logistics Hub, BKC, Mumbai",
		},
		{
			Name:      "Valeo Climate & Electricals",
			ContactNo: "+91 98300 45678",
			Email:     "contact@valeo-auto.in",
			Address:   "Sanand Industrial Estate, Gujarat",
		},
	}
	db.Create(&vendors)
	log.Println("✓ Vendors seeded.")

	// 5. Seed Racks
	racks := []models.Rack{
		{Name: "Rack A", Number: "A1", Description: "Engine & Transmission Components"},
		{Name: "Rack A", Number: "A2", Description: "Braking & Hydraulic Systems"},
		{Name: "Rack B", Number: "B1", Description: "Ignition, Batteries & Electricals"},
		{Name: "Rack B", Number: "B2", Description: "Fluids, Oil Filters & Consumables"},
		{Name: "Rack C", Number: "C1", Description: "Suspension, Steering & Struts"},
		{Name: "Rack C", Number: "C2", Description: "Headlights, Wiper Blades & Accessories"},
	}
	db.Create(&racks)
	log.Println("✓ Storage Racks seeded.")

	// Retrieve IDs for relations
	var vBosch, vBrembo, vNGK, vCastrol, vValeo models.Vendor
	db.Where("name = ?", "Bosch Automotive India").First(&vBosch)
	db.Where("name = ?", "Brembo Brakes Private Ltd").First(&vBrembo)
	db.Where("name = ?", "NGK Spark Plugs Co").First(&vNGK)
	db.Where("name = ?", "Castrol India Lubricants").First(&vCastrol)
	db.Where("name = ?", "Valeo Climate & Electricals").First(&vValeo)

	var rA1, rA2, rB1, rB2, rC1, rC2 models.Rack
	db.Where("name = ? AND number = ?", "Rack A", "A1").First(&rA1)
	db.Where("name = ? AND number = ?", "Rack A", "A2").First(&rA2)
	db.Where("name = ? AND number = ?", "Rack B", "B1").First(&rB1)
	db.Where("name = ? AND number = ?", "Rack B", "B2").First(&rB2)
	db.Where("name = ? AND number = ?", "Rack C", "C1").First(&rC1)
	db.Where("name = ? AND number = ?", "Rack C", "C2").First(&rC2)

	// 6. Seed Items
	items := []models.Item{
		{
			SKU:           "ITEM-BRK-001",
			Name:          "Front Ceramic Brake Pad Set",
			Description:   "Premium ceramic formula for low dust and noise-free braking",
			Price:         1650.00,
			CostPrice:     1100.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      32,
			MinStockAlert: 5,
			RackID:        &rA2.ID,
			VendorID:      &vBrembo.ID,
		},
		{
			SKU:           "ITEM-BRK-002",
			Name:          "Rear Brake Disc Rotor (Pair)",
			Description:   "Vented high-carbon cast iron brake disc rotors",
			Price:         3400.00,
			CostPrice:     2400.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      12,
			MinStockAlert: 3,
			RackID:        &rA2.ID,
			VendorID:      &vBrembo.ID,
		},
		{
			SKU:           "ITEM-FLT-001",
			Name:          "Synthetic High-Flow Engine Oil Filter",
			Description:   "Heavy-duty spin-on canister with synthetic blend filter media",
			Price:         380.00,
			CostPrice:     220.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      55,
			MinStockAlert: 10,
			RackID:        &rB2.ID,
			VendorID:      &vBosch.ID,
		},
		{
			SKU:           "ITEM-FLT-002",
			Name:          "Cabin Air Filter with Activated Carbon",
			Description:   "Traps allergens, micro-dust, and eliminates cabin odors",
			Price:         550.00,
			CostPrice:     320.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      24,
			MinStockAlert: 5,
			RackID:        &rB2.ID,
			VendorID:      &vBosch.ID,
		},
		{
			SKU:           "ITEM-OIL-001",
			Name:          "Castrol EDGE 5W-40 Fully Synthetic Oil (4L)",
			Description:   "Fluid Titanium technology engineered for maximum engine performance",
			Price:         2850.00,
			CostPrice:     2100.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      18,
			MinStockAlert: 4,
			RackID:        &rB2.ID,
			VendorID:      &vCastrol.ID,
		},
		{
			SKU:           "ITEM-SPK-001",
			Name:          "Laser Iridium Spark Plugs (Set of 4)",
			Description:   "Ultra-fine 0.6mm iridium center electrode for optimum throttle response",
			Price:         1950.00,
			CostPrice:     1350.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      20,
			MinStockAlert: 5,
			RackID:        &rB1.ID,
			VendorID:      &vNGK.ID,
		},
		{
			SKU:           "ITEM-LGT-001",
			Name:          "H7 Ultinon LED Headlight Conversion Kit (6500K)",
			Description:   "100W Cool White High/Low Beam with integrated cooling fan",
			Price:         2400.00,
			CostPrice:     1650.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      8,
			MinStockAlert: 3,
			RackID:        &rC2.ID,
			VendorID:      &vValeo.ID,
		},
		{
			SKU:           "ITEM-BAT-001",
			Name:          "12V 45Ah Maintenance-Free Car Battery",
			Description:   "High cranking power with silver alloy calcium technology",
			Price:         4800.00,
			CostPrice:     3700.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      3, // Low Stock alert demo
			MinStockAlert: 5,
			RackID:        &rB1.ID,
			VendorID:      &vValeo.ID,
		},
		{
			SKU:           "ITEM-SUS-001",
			Name:          "Front Shock Absorber Strut Assembly",
			Description:   "Gas-pressurized twin-tube damper for enhanced stability and ride comfort",
			Price:         3200.00,
			CostPrice:     2300.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      2, // Low Stock alert demo
			MinStockAlert: 4,
			RackID:        &rC1.ID,
			VendorID:      &vBrembo.ID,
		},
		{
			SKU:           "ITEM-BLT-001",
			Name:          "Heavy Duty Multi-Rib Serpentine Drive Belt",
			Description:   "EPDM rubber construction resistant to heat, oil, and ozone degradation",
			Price:         890.00,
			CostPrice:     550.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      15,
			MinStockAlert: 4,
			RackID:        &rA1.ID,
			VendorID:      &vBosch.ID,
		},
		{
			SKU:           "ITEM-CLT-001",
			Name:          "Complete Clutch Plate & Release Bearing Kit",
			Description:   "OE spec clutch pressure plate, friction disc, and bearing",
			Price:         5600.00,
			CostPrice:     4100.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      6,
			MinStockAlert: 2,
			RackID:        &rA1.ID,
			VendorID:      &vValeo.ID,
		},
		{
			SKU:           "ITEM-WPR-001",
			Name:          "Aerodynamic Silicone Wiper Blade Pair (24\"+18\")",
			Description:   "Streak-free all-weather silicone wiper blades with universal hook adapter",
			Price:         720.00,
			CostPrice:     420.00,
			GSTApplicable: true,
			GSTRate:       18.00,
			Quantity:      28,
			MinStockAlert: 6,
			RackID:        &rC2.ID,
			VendorID:      &vBosch.ID,
		},
	}
	db.Create(&items)
	log.Println("✓ Inventory Items seeded.")

	// 7. Seed Customers
	customers := []models.Customer{
		{Name: "Rajesh Sharma", Mobile: "9823012345", Email: "rajesh.sharma@gmail.com", Address: "Shop #14, Auto Line, City", TotalBills: 2},
		{Name: "Amit Patel", Mobile: "9876543210", Email: "amit.patel@hotmail.com", Address: "Flat 402, Green Heights, Ring Road", TotalBills: 1},
		{Name: "Priya Sundaram", Mobile: "9123456789", Email: "priya.s@yahoo.com", Address: "Plot 89, Sector 5, Phase 2", TotalBills: 1},
		{Name: "Sunil Verma", Mobile: "9988776655", Email: "sunil.verma@outlook.com", Address: "32/4 Cross Street, Old Market", TotalBills: 0},
	}
	db.Create(&customers)
	log.Println("✓ Customers seeded.")

	// Retrieve items and customers for bill seeding
	var cRajesh, cAmit, cPriya models.Customer
	db.Where("mobile = ?", "9823012345").First(&cRajesh)
	db.Where("mobile = ?", "9876543210").First(&cAmit)
	db.Where("mobile = ?", "9123456789").First(&cPriya)

	var itmBrakePads, itmOilFilter, itmCastrolOil, itmSparkPlugs, itmWipers, itmHeadlight, itmDriveBelt models.Item
	db.Where("sku = ?", "ITEM-BRK-001").First(&itmBrakePads)
	db.Where("sku = ?", "ITEM-FLT-001").First(&itmOilFilter)
	db.Where("sku = ?", "ITEM-OIL-001").First(&itmCastrolOil)
	db.Where("sku = ?", "ITEM-SPK-001").First(&itmSparkPlugs)
	db.Where("sku = ?", "ITEM-WPR-001").First(&itmWipers)
	db.Where("sku = ?", "ITEM-LGT-001").First(&itmHeadlight)
	db.Where("sku = ?", "ITEM-BLT-001").First(&itmDriveBelt)

	// 8. Seed Invoices / Bills
	now := time.Now()

	// Bill 1: Rajesh Sharma (Brake pads + Oil Filter + Castrol Oil)
	bill1Subtotal := (1650.00 * 2) + (380.00 * 1) + (2850.00 * 1) // 3300 + 380 + 2850 = 6530.00
	bill1Gst := bill1Subtotal * 0.18                               // 1175.40
	bill1Total := bill1Subtotal + bill1Gst                         // 7705.40
	bill1 := models.Bill{
		InvoiceNumber:        "INV-2026-0001",
		CustomerID:           &cRajesh.ID,
		CustomerBillSequence: 1,
		UserID:               &sellerUser.ID,
		Subtotal:             bill1Subtotal,
		GSTApplied:           true,
		GSTRate:              18.00,
		GSTAmount:            bill1Gst,
		Discount:             0,
		TotalAmount:          bill1Total,
		PaymentMode:          "upi",
		CreatedAt:            now.AddDate(0, 0, -4),
		Items: []models.BillItem{
			{
				ItemID:     &itmBrakePads.ID,
				Quantity:   2,
				UnitPrice:  1650.00,
				TotalPrice: 3300.00,
				PartName:   itmBrakePads.Name,
				PartSKU:    itmBrakePads.SKU,
				RackInfo:   "Rack A (A2)",
			},
			{
				ItemID:     &itmOilFilter.ID,
				Quantity:   1,
				UnitPrice:  380.00,
				TotalPrice: 380.00,
				PartName:   itmOilFilter.Name,
				PartSKU:    itmOilFilter.SKU,
				RackInfo:   "Rack B (B2)",
			},
			{
				ItemID:     &itmCastrolOil.ID,
				Quantity:   1,
				UnitPrice:  2850.00,
				TotalPrice: 2850.00,
				PartName:   itmCastrolOil.Name,
				PartSKU:    itmCastrolOil.SKU,
				RackInfo:   "Rack B (B2)",
			},
		},
	}
	if err := db.Create(&bill1).Error; err != nil {
		log.Printf("Error seeding bill1: %v", err)
	}

	// Bill 2: Amit Patel (Spark Plugs + Wiper Blades)
	bill2Subtotal := (1950.00 * 1) + (720.00 * 1) // 2670.00
	bill2Gst := bill2Subtotal * 0.18              // 480.60
	bill2Total := bill2Subtotal + bill2Gst        // 3150.60
	bill2 := models.Bill{
		InvoiceNumber:        "INV-2026-0002",
		CustomerID:           &cAmit.ID,
		CustomerBillSequence: 1,
		UserID:               &sellerUser.ID,
		Subtotal:             bill2Subtotal,
		GSTApplied:           true,
		GSTRate:              18.00,
		GSTAmount:            bill2Gst,
		Discount:             0,
		TotalAmount:          bill2Total,
		PaymentMode:          "cash",
		CreatedAt:            now.AddDate(0, 0, -2),
		Items: []models.BillItem{
			{
				ItemID:     &itmSparkPlugs.ID,
				Quantity:   1,
				UnitPrice:  1950.00,
				TotalPrice: 1950.00,
				PartName:   itmSparkPlugs.Name,
				PartSKU:    itmSparkPlugs.SKU,
				RackInfo:   "Rack B (B1)",
			},
			{
				ItemID:     &itmWipers.ID,
				Quantity:   1,
				UnitPrice:  720.00,
				TotalPrice: 720.00,
				PartName:   itmWipers.Name,
				PartSKU:    itmWipers.SKU,
				RackInfo:   "Rack C (C2)",
			},
		},
	}
	if err := db.Create(&bill2).Error; err != nil {
		log.Printf("Error seeding bill2: %v", err)
	}

	// Bill 3: Priya Sundaram (Headlights + Serpentine Belt)
	bill3Subtotal := (2400.00 * 1) + (890.00 * 1) // 3290.00
	bill3Gst := bill3Subtotal * 0.18              // 592.20
	bill3Total := bill3Subtotal + bill3Gst        // 3882.20
	bill3 := models.Bill{
		InvoiceNumber:        "INV-2026-0003",
		CustomerID:           &cPriya.ID,
		CustomerBillSequence: 1,
		UserID:               &adminUser.ID,
		Subtotal:             bill3Subtotal,
		GSTApplied:           true,
		GSTRate:              18.00,
		GSTAmount:            bill3Gst,
		Discount:             0,
		TotalAmount:          bill3Total,
		PaymentMode:          "card",
		CreatedAt:            now.AddDate(0, 0, -1),
		Items: []models.BillItem{
			{
				ItemID:     &itmHeadlight.ID,
				Quantity:   1,
				UnitPrice:  2400.00,
				TotalPrice: 2400.00,
				PartName:   itmHeadlight.Name,
				PartSKU:    itmHeadlight.SKU,
				RackInfo:   "Rack C (C2)",
			},
			{
				ItemID:     &itmDriveBelt.ID,
				Quantity:   1,
				UnitPrice:  890.00,
				TotalPrice: 890.00,
				PartName:   itmDriveBelt.Name,
				PartSKU:    itmDriveBelt.SKU,
				RackInfo:   "Rack A (A1)",
			},
		},
	}
	if err := db.Create(&bill3).Error; err != nil {
		log.Printf("Error seeding bill3: %v", err)
	}
	log.Println("✓ Sample Invoices seeded.")

	// 9. Seed Return Bill (Credit Note) on Bill 1
	var firstBill models.Bill
	db.Preload("Items").Where("invoice_number = ?", "INV-2026-0001").First(&firstBill)

	var targetBillItemID uint
	if len(firstBill.Items) > 0 {
		targetBillItemID = firstBill.Items[0].ID
	}

	retRefundAmount := 1650.00 * 1.18 // 1947.00
	returnBill := models.ReturnBill{
		ReturnNumber:      "RET-2026-0001",
		OriginalBillID:    firstBill.ID,
		CustomerID:        &cRajesh.ID,
		UserID:            &sellerUser.ID,
		Reason:            "Wrong Item Fitment / Model",
		RefundSubtotal:    1650.00,
		GSTRefunded:       true,
		GSTRate:           18.00,
		GSTAmount:         1650.00 * 0.18,
		TotalRefundAmount: retRefundAmount,
		RefundMode:        "upi",
		CreatedAt:         now.AddDate(0, 0, -2),
		Items: []models.ReturnItem{
			{
				BillItemID: &targetBillItemID,
				ItemID:     &itmBrakePads.ID,
				Quantity:   1,
				UnitPrice:  1650.00,
				TotalPrice: 1650.00,
				PartName:   itmBrakePads.Name,
				PartSKU:    itmBrakePads.SKU,
				RackInfo:   "Rack A (A2)",
			},
		},
	}
	if err := db.Create(&returnBill).Error; err != nil {
		log.Printf("Error seeding returnBill: %v", err)
	}
	log.Println("✓ Sample Return Bill & Credit Note seeded.")

	log.Println("🎉 Database reset and rich mock dataset seeded successfully!")
	return nil
}

func SeedDefaultData(db *gorm.DB) error {
	var itemCount int64
	db.Model(&models.Item{}).Count(&itemCount)
	if itemCount == 0 {
		return ResetAndSeedMockData(db)
	}
	return nil
}
