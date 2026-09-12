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
	ShopName       string    `gorm:"size:150;default:'Spare Parts Pro'" json:"shop_name"`
	ShopAddress    string    `gorm:"type:text" json:"shop_address"`
	ShopPhone      string    `gorm:"size:30" json:"shop_phone"`
	ShopEmail      string    `gorm:"size:100" json:"shop_email"`
	ShopGSTIN      string    `gorm:"size:30" json:"shop_gstin"`
	DefaultGSTRate float64   `gorm:"type:numeric(5,2);default:18.00" json:"default_gst_rate"`
	InvoicePrefix  string    `gorm:"size:10;default:'INV'" json:"invoice_prefix"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Rack struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `gorm:"size:100;not null" json:"name"`
	Number      string      `gorm:"size:50;not null" json:"number"`
	Description string      `gorm:"type:text" json:"description"`
	SpareParts  []SparePart `gorm:"foreignKey:RackID" json:"spare_parts,omitempty"`
	PartsCount  int64       `gorm:"-" json:"parts_count"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type SparePart struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SKU            string    `gorm:"uniqueIndex;size:50;not null" json:"sku"`
	Name           string    `gorm:"size:200;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	Price          float64   `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	CostPrice      float64   `gorm:"type:numeric(12,2);default:0" json:"cost_price"`
	GSTApplicable  bool      `gorm:"default:true" json:"gst_applicable"`
	GSTRate        float64   `gorm:"type:numeric(5,2);default:18.00" json:"gst_rate"`
	Quantity       int       `gorm:"not null;default:0" json:"quantity"`
	MinStockAlert  int       `gorm:"not null;default:5" json:"min_stock_alert"`
	RackID         *uint     `gorm:"index" json:"rack_id"`
	Rack           *Rack     `gorm:"foreignKey:RackID" json:"rack,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

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
	ID          uint       `gorm:"primaryKey" json:"id"`
	BillID      uint       `gorm:"index;not null" json:"bill_id"`
	SparePartID *uint      `gorm:"index" json:"spare_part_id"`
	SparePart   *SparePart `gorm:"foreignKey:SparePartID" json:"spare_part,omitempty"`
	PartName    string     `gorm:"size:200;not null" json:"part_name"`
	PartSKU     string     `gorm:"size:50;not null" json:"part_sku"`
	RackInfo    string     `gorm:"size:150" json:"rack_info"`
	UnitPrice   float64    `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Quantity    int        `gorm:"not null" json:"quantity"`
	TotalPrice  float64    `gorm:"type:numeric(12,2);not null" json:"total_price"`
}
