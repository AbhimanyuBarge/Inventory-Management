package models

import (
	"time"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleSeller Role = "seller"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FullName     string    `gorm:"size:100;not null" json:"full_name"`
	Role         Role      `gorm:"size:20;not null;default:'seller'" json:"role"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Setting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ShopName       string    `gorm:"size:150;default:'Inventory Pro'" json:"shop_name"`
	ShopAddress    string    `gorm:"type:text" json:"shop_address"`
	ShopPhone      string    `gorm:"size:30" json:"shop_phone"`
	ShopEmail      string    `gorm:"size:100" json:"shop_email"`
	ShopGSTIN      string    `gorm:"size:30" json:"shop_gstin"`
	DefaultGSTRate float64   `gorm:"type:numeric(5,2);default:18.00" json:"default_gst_rate"`
	InvoicePrefix  string    `gorm:"size:10;default:'INV'" json:"invoice_prefix"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Rack struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Number      string    `gorm:"size:50;not null" json:"number"`
	Description string    `gorm:"type:text" json:"description"`
	Items       []Item    `gorm:"foreignKey:RackID" json:"items,omitempty"`
	ItemsCount  int64     `gorm:"-" json:"items_count"`
	PartsCount  int64     `gorm:"-" json:"parts_count"` // backward compatibility
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Vendor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:150;not null" json:"name"`
	ContactNo string    `gorm:"size:30" json:"contact_no"`
	Email     string    `gorm:"size:100" json:"email"`
	Address   string    `gorm:"type:text" json:"address"`
	Items     []Item    `gorm:"foreignKey:VendorID" json:"items,omitempty"`
	ItemsCount int64    `gorm:"-" json:"items_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Item struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SKU           string    `gorm:"uniqueIndex;size:50;not null" json:"sku"`
	Name          string    `gorm:"size:200;not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	Price         float64   `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	CostPrice     float64   `gorm:"type:numeric(12,2);default:0" json:"cost_price"`
	GSTApplicable bool      `gorm:"default:true" json:"gst_applicable"`
	GSTRate       float64   `gorm:"type:numeric(5,2);default:18.00" json:"gst_rate"`
	Quantity      int       `gorm:"not null;default:0" json:"quantity"`
	MinStockAlert int       `gorm:"not null;default:5" json:"min_stock_alert"`
	RackID        *uint     `gorm:"index" json:"rack_id"`
	Rack          *Rack     `gorm:"foreignKey:RackID" json:"rack,omitempty"`
	VendorID      *uint     `gorm:"index" json:"vendor_id"`
	Vendor        *Vendor   `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SparePart is an alias for Item for full backward compatibility
type SparePart = Item

type Customer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:150;not null" json:"name"`
	Mobile     string    `gorm:"size:30;index" json:"mobile"`
	Email      string    `gorm:"size:100" json:"email"`
	Address    string    `gorm:"type:text" json:"address"`
	TotalBills int       `gorm:"default:0" json:"total_bills"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Bill struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	InvoiceNumber        string     `gorm:"uniqueIndex;size:50;not null" json:"invoice_number"`
	CustomerID           *uint      `gorm:"index" json:"customer_id"`
	Customer             *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerBillSequence int        `gorm:"not null;default:1" json:"customer_bill_sequence"`
	UserID               *uint      `gorm:"index" json:"user_id"`
	User                 *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Subtotal             float64    `gorm:"type:numeric(12,2);not null;default:0" json:"subtotal"`
	GSTApplied           bool       `gorm:"default:false" json:"gst_applied"`
	GSTRate              float64    `gorm:"type:numeric(5,2);default:0" json:"gst_rate"`
	GSTAmount            float64    `gorm:"type:numeric(12,2);default:0" json:"gst_amount"`
	Discount             float64    `gorm:"type:numeric(12,2);default:0" json:"discount"`
	TotalAmount          float64    `gorm:"type:numeric(12,2);not null;default:0" json:"total_amount"`
	PaymentMode          string     `gorm:"size:30;default:'cash'" json:"payment_mode"`
	Notes                string     `gorm:"type:text" json:"notes"`
	Items                []BillItem `gorm:"foreignKey:BillID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt            time.Time  `json:"created_at"`
}

type BillItem struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	BillID      uint     `gorm:"index;not null" json:"bill_id"`
	ItemID      *uint    `gorm:"index" json:"item_id"`
	SparePartID *uint    `gorm:"index" json:"spare_part_id,omitempty"`
	Item        *Item    `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
	PartName    string   `gorm:"size:200;not null" json:"part_name"`
	PartSKU     string   `gorm:"size:50;not null" json:"part_sku"`
	RackInfo    string   `gorm:"size:150" json:"rack_info"`
	UnitPrice   float64  `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Quantity    int      `gorm:"not null" json:"quantity"`
	TotalPrice  float64  `gorm:"type:numeric(12,2);not null" json:"total_price"`
}

type ReturnBill struct {
	ID                uint         `gorm:"primaryKey" json:"id"`
	ReturnNumber      string       `gorm:"uniqueIndex;size:50;not null" json:"return_number"`
	OriginalBillID    uint         `gorm:"index;not null" json:"original_bill_id"`
	OriginalBill      *Bill        `gorm:"foreignKey:OriginalBillID" json:"original_bill,omitempty"`
	CustomerID        *uint        `gorm:"index" json:"customer_id"`
	Customer          *Customer    `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	UserID            *uint        `gorm:"index" json:"user_id"`
	User              *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RefundSubtotal    float64      `gorm:"type:numeric(12,2);not null;default:0" json:"refund_subtotal"`
	GSTRefunded       bool         `gorm:"default:false" json:"gst_refunded"`
	GSTRate           float64      `gorm:"type:numeric(5,2);default:0" json:"gst_rate"`
	GSTAmount         float64      `gorm:"type:numeric(12,2);default:0" json:"gst_amount"`
	TotalRefundAmount float64      `gorm:"type:numeric(12,2);not null;default:0" json:"total_refund_amount"`
	RefundMode        string       `gorm:"size:30;default:'cash'" json:"refund_mode"`
	Reason            string       `gorm:"size:255" json:"reason"`
	Notes             string       `gorm:"type:text" json:"notes"`
	Items             []ReturnItem `gorm:"foreignKey:ReturnBillID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt         time.Time    `json:"created_at"`
}

type ReturnItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ReturnBillID uint      `gorm:"index;not null" json:"return_bill_id"`
	BillItemID   *uint     `gorm:"index" json:"bill_item_id"`
	BillItem     *BillItem `gorm:"foreignKey:BillItemID" json:"bill_item,omitempty"`
	ItemID       *uint     `gorm:"index" json:"item_id"`
	SparePartID  *uint     `gorm:"index" json:"spare_part_id,omitempty"`
	Item         *Item     `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
	PartName     string    `gorm:"size:200;not null" json:"part_name"`
	PartSKU      string    `gorm:"size:50;not null" json:"part_sku"`
	RackInfo     string    `gorm:"size:150" json:"rack_info"`
	UnitPrice    float64   `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	TotalPrice   float64   `gorm:"type:numeric(12,2);not null" json:"total_price"`
}
