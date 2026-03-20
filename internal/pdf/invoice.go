package pdf

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"fintech-backend/internal/repository"
)

// GenerateInvoicePDF renders a printable invoice document.
func GenerateInvoicePDF(shop *repository.Shop, invoice *repository.Invoice, items []repository.InvoiceItem, products map[string]string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(190, 10, "INVOICE")
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 12)

	pdf.Cell(190, 7, fmt.Sprintf("Shop: %s", shop.Name))
	pdf.Ln(6)
	pdf.Cell(190, 7, fmt.Sprintf("Address: %s", shop.Address))
	pdf.Ln(6)
	pdf.Cell(190, 7, fmt.Sprintf("GST: %s", shop.GSTNumber))
	pdf.Ln(10)

	pdf.Cell(190, 7, fmt.Sprintf("Invoice No: %s", invoice.InvoiceNumber))
	pdf.Ln(6)
	pdf.Cell(190, 7, fmt.Sprintf("Date: %s", invoice.CreatedAt.Format("02 Jan 2006")))
	pdf.Ln(6)

	pdf.Cell(190, 7, fmt.Sprintf("Customer: %s", invoice.CustomerName))
	pdf.Ln(6)
	pdf.Cell(190, 7, fmt.Sprintf("Phone: %s", invoice.CustomerPhone))
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(80, 10, "Item")
	pdf.Cell(30, 10, "Qty")
	pdf.Cell(40, 10, "Unit Price")
	pdf.Cell(40, 10, "Total")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 12)
	for _, it := range items {
		name := products[it.ProductID.String()]

		pdf.Cell(80, 8, name)
		pdf.Cell(30, 8, fmt.Sprintf("%d", it.Quantity))
		pdf.Cell(40, 8, fmt.Sprintf("%.2f", it.UnitPrice))
		pdf.Cell(40, 8, fmt.Sprintf("%.2f", float64(it.Quantity)*it.UnitPrice))
		pdf.Ln(8)
	}

	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(190, 10, fmt.Sprintf("Subtotal: %.2f", invoice.Subtotal))
	pdf.Ln(7)
	pdf.Cell(190, 10, fmt.Sprintf("Tax: %.2f", invoice.TaxAmount))
	pdf.Ln(7)
	pdf.Cell(190, 10, fmt.Sprintf("Discount: %.2f", invoice.DiscountAmount))
	pdf.Ln(7)
	pdf.Cell(190, 10, fmt.Sprintf("Grand Total: %.2f", invoice.TotalAmount))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
