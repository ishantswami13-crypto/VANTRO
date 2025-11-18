package dto

// ====== SHOPS ======

type CreateShopRequest struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	GSTNumber  string `json:"gst_number"`
	OwnerEmail string `json:"owner_email"`
}

// ====== PRODUCTS ======

type CreateProductRequest struct {
	ShopID            string  `json:"shop_id"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku"`
	Stock             int     `json:"stock"`
	CostPrice         float64 `json:"cost_price"`
	SellingPrice      float64 `json:"selling_price"`
	LowStockThreshold int     `json:"low_stock_threshold"`
}

// ====== INVOICES ======

type InvoiceItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type CreateInvoiceRequest struct {
	ShopID        string               `json:"shop_id"`
	CustomerName  string               `json:"customer_name"`
	CustomerPhone string               `json:"customer_phone"`
	TaxAmount     float64              `json:"tax_amount"`
	Items         []InvoiceItemRequest `json:"items"`
}

// ====== EXPENSES ======

type CreateExpenseRequest struct {
	ShopID   string  `json:"shop_id"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Note     string  `json:"note"`
}

// ====== POTS ======

type CreatePotRequest struct {
	ShopID       string  `json:"shop_id"`
	Name         string  `json:"name"`
	TargetAmount float64 `json:"target_amount"`
}

type DepositPotRequest struct {
	Amount float64 `json:"amount"`
}
