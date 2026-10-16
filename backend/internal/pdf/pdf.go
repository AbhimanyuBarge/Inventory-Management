package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/inventory/backend/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// GenerateInvoicePDF creates a sleek, professional invoice PDF
func GenerateInvoicePDF(bill *models.Bill, setting *models.Setting) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Colors
	primaryColor := []int{24, 43, 73}      // Deep Navy (#182B49)
	accentColor := []int{0, 122, 255}      // Vibrant Blue (#007AFF)
	textDark := []int{33, 37, 41}          // Charcoal
	textMuted := []int{108, 117, 125}      // Gray
	tableHeaderBg := []int{240, 244, 248}  // Light Slate

	// Header Banner / Title
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Rect(15, 15, 180, 26, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 18)
	shopName := "Spare Parts Pro"
	if setting != nil && setting.ShopName != "" {
		shopName = setting.ShopName
	}
	pdf.SetXY(20, 20)
	pdf.Cell(100, 8, shopName)

	pdf.SetFont("Arial", "B", 14)
	pdf.SetXY(120, 20)
	pdf.CellFormat(70, 8, "TAX INVOICE", "", 0, "R", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(220, 230, 245)
	pdf.SetXY(20, 29)
	shopSub := "Automotive & Industrial Spare Parts Hub"
	if setting != nil && setting.ShopGSTIN != "" {
		shopSub = fmt.Sprintf("GSTIN: %s | Ph: %s", setting.ShopGSTIN, setting.ShopPhone)
	}
	pdf.Cell(120, 6, shopSub)

	pdf.SetXY(120, 29)
	pdf.CellFormat(70, 6, fmt.Sprintf("Date: %s", bill.CreatedAt.Format("02 Jan 2006, 15:04")), "", 0, "R", false, 0, "")

	pdf.Ln(20)

	// Shop Info & Customer Info Side-by-Side
	yPos := 48.0
	pdf.SetY(yPos)

	// Left: Shop Details
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(15, yPos)
	pdf.Cell(85, 6, "ISSUED BY:")

	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "", 9)
	if setting != nil {
		pdf.SetXY(15, yPos+6)
		pdf.MultiCell(85, 4.5, fmt.Sprintf("%s\n%s\nPhone: %s\nEmail: %s", setting.ShopName, setting.ShopAddress, setting.ShopPhone, setting.ShopEmail), "", "L", false)
	}

	// Right: Customer & Invoice Info Box
	pdf.SetFillColor(tableHeaderBg[0], tableHeaderBg[1], tableHeaderBg[2])
	pdf.SetDrawColor(220, 225, 230)
	pdf.RoundedRect(105, yPos, 90, 32, 2, "1234", "FD")

	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(108, yPos+2)
	pdf.Cell(84, 5, fmt.Sprintf("Invoice #: %s", bill.InvoiceNumber))

	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "", 9)

	custName := "Walk-in Customer"
	custPhone := "-"
	if bill.Customer != nil {
		custName = bill.Customer.Name
		if bill.Customer.Mobile != "" {
			custPhone = bill.Customer.Mobile
		}
	}

	pdf.SetXY(108, yPos+8)
	pdf.Cell(84, 5, fmt.Sprintf("Customer: %s", custName))
	pdf.SetXY(108, yPos+13)
	pdf.Cell(84, 5, fmt.Sprintf("Contact: %s", custPhone))
	pdf.SetXY(108, yPos+18)
	pdf.SetTextColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(84, 5, fmt.Sprintf("Customer Sequence: Bill #%d", bill.CustomerBillSequence))

	pdf.SetXY(108, yPos+24)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.SetFont("Arial", "", 8.5)
	billedBy := "Staff"
	if bill.User != nil {
		billedBy = bill.User.FullName
	}
	pdf.Cell(84, 5, fmt.Sprintf("Payment: %s | Billed by: %s", bill.PaymentMode, billedBy))

	// Items Table
	tableY := 86.0
	pdf.SetY(tableY)

	// Header Row
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(12, 8, "Sr.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(75, 8, "Item Description / SKU", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "Rack Location", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 8, "Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(23, 8, "Total", "1", 1, "R", true, 0, "")

	// Table Rows
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "", 8.5)

	fill := false
	for idx, item := range bill.Items {
		if fill {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		desc := fmt.Sprintf("%s (%s)", item.PartName, item.PartSKU)
		rack := item.RackInfo
		if rack == "" {
			rack = "-"
		}

		pdf.CellFormat(12, 7, fmt.Sprintf("%d", idx+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(75, 7, desc, "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 7, rack, "1", 0, "C", true, 0, "")
		pdf.CellFormat(18, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 7, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(23, 7, fmt.Sprintf("%.2f", item.TotalPrice), "1", 1, "R", true, 0, "")
		fill = !fill
	}

	pdf.Ln(4)

	// Summary Calculations Box (Right aligned)
	calcY := pdf.GetY()
	calcX := 115.0
	calcWidth := 80.0

	pdf.SetXY(calcX, calcY)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])

	pdf.CellFormat(45, 6, "Subtotal:", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 6, fmt.Sprintf("Rs. %.2f", bill.Subtotal), "", 1, "R", false, 0, "")

	if bill.GSTApplied {
		pdf.SetX(calcX)
		pdf.CellFormat(45, 6, fmt.Sprintf("GST (%.1f%%):", bill.GSTRate), "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("+ Rs. %.2f", bill.GSTAmount), "", 1, "R", false, 0, "")
	} else {
		pdf.SetX(calcX)
		pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		pdf.CellFormat(45, 6, "GST:", "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, "Not Applied (0%)", "", 1, "R", false, 0, "")
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	}

	if bill.Discount > 0 {
		pdf.SetX(calcX)
		pdf.CellFormat(45, 6, "Discount:", "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("- Rs. %.2f", bill.Discount), "", 1, "R", false, 0, "")
	}

	// Grand Total Highlight
	pdf.SetDrawColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Line(calcX, pdf.GetY()+1, calcX+calcWidth, pdf.GetY()+1)
	pdf.Ln(2)

	pdf.SetX(calcX)
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 8, " GRAND TOTAL:", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("Rs. %.2f ", bill.TotalAmount), "1", 1, "R", true, 0, "")

	// Notes & Terms Footer
	pdf.SetY(245)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(180, 4, "Terms & Conditions:\n1. Goods once sold can be exchanged within 7 days with original invoice.\n2. Electrical and electronic spare parts carry manufacturer warranty only.\n3. Thank you for choosing us!", "", "L", false)

	// Signature
	pdf.SetY(265)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.CellFormat(90, 5, "Customer Signature", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 5, "Authorized Signatory", "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateRackLabelPDF creates printable rack labels
func GenerateRackLabelPDF(rack *models.Rack, items []models.Item) ([]byte, error) {
	// A6 Landscape size (148 x 105 mm)
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 148, Ht: 105},
	})
	pdf.SetMargins(8, 8, 8)
	pdf.AddPage()

	// Outer border
	pdf.SetDrawColor(24, 43, 73)
	pdf.SetLineWidth(1.0)
	pdf.Rect(6, 6, 136, 93, "D")

	// Header Box
	pdf.SetFillColor(24, 43, 73)
	pdf.Rect(6, 6, 136, 26, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(10, 8)
	pdf.Cell(128, 4, "INVENTORY STORAGE LOCATION")

	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(10, 14)
	pdf.Cell(128, 8, fmt.Sprintf("%s  •  #%s", rack.Name, rack.Number))

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(200, 220, 245)
	pdf.SetXY(10, 24)
	desc := rack.Description
	if desc == "" {
		desc = "General Inventory Storage"
	}
	pdf.Cell(128, 4, desc)

	// Items stored inside this rack
	pdf.SetY(36)
	pdf.SetTextColor(24, 43, 73)
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(132, 5, fmt.Sprintf("ASSIGNED INVENTORY ITEMS (%d Items):", len(items)))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(40, 40, 40)

	maxShow := 5
	for i, item := range items {
		if i >= maxShow {
			pdf.SetFont("Arial", "I", 7.5)
			pdf.SetTextColor(100, 100, 100)
			pdf.Cell(132, 4, fmt.Sprintf("+ %d more items stored in this rack...", len(items)-maxShow))
			pdf.Ln(4)
			break
		}
		pdf.CellFormat(25, 4.5, item.SKU, "B", 0, "L", false, 0, "")
		pdf.CellFormat(85, 4.5, item.Name, "B", 0, "L", false, 0, "")
		pdf.CellFormat(22, 4.5, fmt.Sprintf("Stock: %d", item.Quantity), "B", 1, "R", false, 0, "")
	}

	if len(items) == 0 {
		pdf.SetTextColor(120, 120, 120)
		pdf.SetFont("Arial", "I", 8.5)
		pdf.Cell(132, 6, "No items currently assigned to this rack.")
		pdf.Ln(6)
	}

	// Barcode / Label Footer
	pdf.SetY(88)
	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetTextColor(100, 100, 100)
	timestamp := time.Now().Format("02-01-2006")
	pdf.CellFormat(66, 5, fmt.Sprintf("ID: %04d | TAG: %s-%s", rack.ID, rack.Name, rack.Number), "", 0, "L", false, 0, "")
	pdf.CellFormat(66, 5, fmt.Sprintf("Printed on: %s", timestamp), "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateReturnPDF creates a credit note / return invoice PDF
func GenerateReturnPDF(returnBill *models.ReturnBill, originalBill *models.Bill, setting *models.Setting) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Theme Colors: Distinct Crimson/Rust theme for returns
	primaryColor := []int{160, 32, 40}      // Rust Crimson
	textDark := []int{33, 37, 41}           // Charcoal
	textMuted := []int{108, 117, 125}       // Gray
	tableHeaderBg := []int{253, 242, 242}   // Light Rose

	// Header Banner
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Rect(15, 15, 180, 26, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 18)
	shopName := "Spare Parts Pro"
	if setting != nil && setting.ShopName != "" {
		shopName = setting.ShopName
	}
	pdf.SetXY(20, 20)
	pdf.Cell(100, 8, shopName)

	pdf.SetFont("Arial", "B", 13)
	pdf.SetXY(110, 20)
	pdf.CellFormat(80, 8, "CREDIT NOTE / RETURN", "", 0, "R", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(255, 225, 225)
	pdf.SetXY(20, 29)
	shopSub := "Sales Return & Refund Voucher"
	if setting != nil && setting.ShopGSTIN != "" {
		shopSub = fmt.Sprintf("GSTIN: %s | Ph: %s", setting.ShopGSTIN, setting.ShopPhone)
	}
	pdf.Cell(110, 6, shopSub)

	pdf.SetXY(120, 29)
	pdf.CellFormat(70, 6, fmt.Sprintf("Return Date: %s", returnBill.CreatedAt.Format("02 Jan 2006, 15:04")), "", 0, "R", false, 0, "")

	pdf.Ln(20)

	// Shop & Customer Info
	yPos := 48.0
	pdf.SetY(yPos)

	// Left: Shop
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(15, yPos)
	pdf.Cell(85, 6, "ISSUED BY:")

	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "", 9)
	if setting != nil {
		pdf.SetXY(15, yPos+6)
		pdf.MultiCell(85, 4.5, fmt.Sprintf("%s\n%s\nPhone: %s\nEmail: %s", setting.ShopName, setting.ShopAddress, setting.ShopPhone, setting.ShopEmail), "", "L", false)
	}

	// Right: Return & Original Invoice Info Box
	pdf.SetFillColor(tableHeaderBg[0], tableHeaderBg[1], tableHeaderBg[2])
	pdf.SetDrawColor(240, 210, 210)
	pdf.RoundedRect(105, yPos, 90, 36, 2, "1234", "FD")

	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(108, yPos+2)
	pdf.Cell(84, 5, fmt.Sprintf("Return Voucher #: %s", returnBill.ReturnNumber))

	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "B", 9)
	origInv := "N/A"
	if originalBill != nil {
		origInv = originalBill.InvoiceNumber
	}
	pdf.SetXY(108, yPos+7)
	pdf.Cell(84, 5, fmt.Sprintf("Ref Original Invoice: %s", origInv))

	pdf.SetFont("Arial", "", 8.5)
	custName := "Walk-in Customer"
	custPhone := "-"
	if returnBill.Customer != nil {
		custName = returnBill.Customer.Name
		if returnBill.Customer.Mobile != "" {
			custPhone = returnBill.Customer.Mobile
		}
	}

	pdf.SetXY(108, yPos+12)
	pdf.Cell(84, 5, fmt.Sprintf("Customer: %s (Ph: %s)", custName, custPhone))

	reason := returnBill.Reason
	if reason == "" {
		reason = "Customer Return / Restock"
	}
	pdf.SetXY(108, yPos+17)
	pdf.Cell(84, 5, fmt.Sprintf("Return Reason: %s", reason))

	pdf.SetXY(108, yPos+22)
	refundMode := returnBill.RefundMode
	if refundMode == "" {
		refundMode = "Cash"
	}
	staffName := "Staff"
	if returnBill.User != nil {
		staffName = returnBill.User.FullName
	}
	pdf.Cell(84, 5, fmt.Sprintf("Refund Mode: %s | Processed by: %s", refundMode, staffName))

	// Returned Items Table
	tableY := 90.0
	pdf.SetY(tableY)

	// Table Header
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(12, 8, "Sr.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(75, 8, "Returned Item / SKU", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "Restocked Rack", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 8, "Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(23, 8, "Refund", "1", 1, "R", true, 0, "")

	// Table Rows
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Arial", "", 8.5)

	fill := false
	for idx, item := range returnBill.Items {
		if fill {
			pdf.SetFillColor(254, 248, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		desc := fmt.Sprintf("%s (%s)", item.PartName, item.PartSKU)
		rack := item.RackInfo
		if rack == "" {
			rack = "-"
		}

		pdf.CellFormat(12, 7, fmt.Sprintf("%d", idx+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(75, 7, desc, "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 7, rack, "1", 0, "C", true, 0, "")
		pdf.CellFormat(18, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 7, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(23, 7, fmt.Sprintf("%.2f", item.TotalPrice), "1", 1, "R", true, 0, "")
		fill = !fill
	}

	pdf.Ln(4)

	// Summary Calculations Box
	calcY := pdf.GetY()
	calcX := 115.0
	calcWidth := 80.0

	pdf.SetXY(calcX, calcY)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])

	pdf.CellFormat(45, 6, "Refund Subtotal:", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 6, fmt.Sprintf("Rs. %.2f", returnBill.RefundSubtotal), "", 1, "R", false, 0, "")

	if returnBill.GSTRefunded && returnBill.GSTAmount > 0 {
		pdf.SetX(calcX)
		pdf.CellFormat(45, 6, fmt.Sprintf("GST Refunded (%.1f%%):", returnBill.GSTRate), "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("+ Rs. %.2f", returnBill.GSTAmount), "", 1, "R", false, 0, "")
	}

	// Grand Refund Total
	pdf.SetDrawColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Line(calcX, pdf.GetY()+1, calcX+calcWidth, pdf.GetY()+1)
	pdf.Ln(2)

	pdf.SetX(calcX)
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 8, " TOTAL REFUND:", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("Rs. %.2f ", returnBill.TotalRefundAmount), "1", 1, "R", true, 0, "")

	// Notes & Terms Footer
	pdf.SetY(245)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(180, 4, "Notes:\n1. Returned items have been inspected and restocked into their respective storage racks.\n2. This Credit Note is valid for cash refund or credit against future purchases.\n3. Keep this voucher for your records.", "", "L", false)

	// Signature
	pdf.SetY(265)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.CellFormat(90, 5, "Customer Signature", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 5, "Authorized Signatory", "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
