package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const timeout = 5 * time.Second

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ========== USERS / SHOPS ==========

type User struct {
	ID     uuid.UUID
	Name   string
	Email  string
	APIKey string
}

type Shop struct {
	ID        uuid.UUID
	Name      string
	Address   string
	GSTNumber string
	OwnerID   uuid.UUID
	CreatedAt time.Time
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, api_key
		FROM users WHERE email = $1
	`, email)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.APIKey); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, api_key
		FROM users WHERE api_key = $1
	`, apiKey)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.APIKey); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateShop(ctx context.Context, ownerID uuid.UUID, name, address, gst string) (*Shop, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var s Shop
	err := r.pool.QueryRow(ctx, `
		INSERT INTO shops (owner_id, name, address, gst_number)
		VALUES ($1, $2, $3, $4)
		RETURNING id, owner_id, name, address, gst_number, created_at
	`, ownerID, name, address, gst).Scan(
		&s.ID, &s.OwnerID, &s.Name, &s.Address, &s.GSTNumber, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListShopsByUser(ctx context.Context, ownerID uuid.UUID) ([]Shop, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, name, address, gst_number, created_at
		FROM shops
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Shop
	for rows.Next() {
		var s Shop
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Address, &s.GSTNumber, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *Repository) GetShopByID(ctx context.Context, id uuid.UUID) (*Shop, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, address, gst_number, created_at
		FROM shops WHERE id = $1
	`, id)

	var s Shop
	if err := row.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Address, &s.GSTNumber, &s.CreatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

// ========== PRODUCTS ==========

type Product struct {
	ID                uuid.UUID
	ShopID            uuid.UUID
	Name              string
	SKU               *string
	Stock             int
	CostPrice         float64
	SellingPrice      float64
	LowStockThreshold int
}

func (r *Repository) CreateProduct(ctx context.Context, p Product) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO products (shop_id, name, sku, stock, cost_price, selling_price, low_stock_threshold)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, p.ShopID, p.Name, p.SKU, p.Stock, p.CostPrice, p.SellingPrice, p.LowStockThreshold).Scan(&p.ID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListProductsByShop(ctx context.Context, shopID uuid.UUID) ([]Product, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, shop_id, name, sku, stock, cost_price, selling_price, low_stock_threshold
		FROM products WHERE shop_id = $1
		ORDER BY name
	`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.ShopID, &p.Name, &p.SKU, &p.Stock, &p.CostPrice, &p.SellingPrice, &p.LowStockThreshold); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

// ========== INVOICES ==========

type Invoice struct {
	ID            uuid.UUID
	ShopID        uuid.UUID
	CustomerName  string
	CustomerPhone string
	TotalAmount   float64
	TaxAmount     float64
	Status        string
	CreatedAt     time.Time
}

type InvoiceItem struct {
	ID        uuid.UUID
	InvoiceID uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}

func (r *Repository) CreateInvoiceWithItems(ctx context.Context, inv Invoice, items []InvoiceItem) (*Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO invoices (shop_id, customer_name, customer_phone, total_amount, tax_amount, status)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at
	`, inv.ShopID, inv.CustomerName, inv.CustomerPhone, inv.TotalAmount, inv.TaxAmount, inv.Status).
		Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i := range items {
		item := &items[i]
		item.InvoiceID = inv.ID
		err = tx.QueryRow(ctx, `
			INSERT INTO invoice_items (invoice_id, product_id, quantity, unit_price)
			VALUES ($1,$2,$3,$4)
			RETURNING id
		`, item.InvoiceID, item.ProductID, item.Quantity, item.UnitPrice).Scan(&item.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *Repository) ListInvoicesByShop(ctx context.Context, shopID uuid.UUID) ([]Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, shop_id, customer_name, customer_phone, total_amount, tax_amount, status, created_at
		FROM invoices
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Invoice
	for rows.Next() {
		var iv Invoice
		if err := rows.Scan(&iv.ID, &iv.ShopID, &iv.CustomerName, &iv.CustomerPhone, &iv.TotalAmount, &iv.TaxAmount, &iv.Status, &iv.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, iv)
	}
	return result, nil
}

// ========== EXPENSES ==========

type Expense struct {
	ID       uuid.UUID
	ShopID   uuid.UUID
	Category string
	Amount   float64
	Note     *string
	SpentAt  time.Time
}

func (r *Repository) CreateExpense(ctx context.Context, e Expense) (*Expense, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO expenses (shop_id, category, amount, note)
		VALUES ($1,$2,$3,$4)
		RETURNING id, spent_at
	`, e.ShopID, e.Category, e.Amount, e.Note).Scan(&e.ID, &e.SpentAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ListExpensesByShop(ctx context.Context, shopID uuid.UUID) ([]Expense, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, shop_id, category, amount, note, spent_at
		FROM expenses
		WHERE shop_id = $1
		ORDER BY spent_at DESC
	`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.ShopID, &e.Category, &e.Amount, &e.Note, &e.SpentAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// ========== POTS ==========

type Pot struct {
	ID            uuid.UUID
	ShopID        uuid.UUID
	Name          string
	TargetAmount  float64
	CurrentAmount float64
	CreatedAt     time.Time
}

func (r *Repository) CreatePot(ctx context.Context, p Pot) (*Pot, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO pots (shop_id, name, target_amount)
		VALUES ($1,$2,$3)
		RETURNING id, current_amount, created_at
	`, p.ShopID, p.Name, p.TargetAmount).Scan(&p.ID, &p.CurrentAmount, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) DepositToPot(ctx context.Context, potID uuid.UUID, amount float64) (*Pot, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var p Pot
	err := r.pool.QueryRow(ctx, `
		UPDATE pots
		SET current_amount = current_amount + $1
		WHERE id = $2
		RETURNING id, shop_id, name, target_amount, current_amount, created_at
	`, amount, potID).Scan(&p.ID, &p.ShopID, &p.Name, &p.TargetAmount, &p.CurrentAmount, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListPotsByShop(ctx context.Context, shopID uuid.UUID) ([]Pot, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, shop_id, name, target_amount, current_amount, created_at
		FROM pots
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Pot
	for rows.Next() {
		var p Pot
		if err := rows.Scan(&p.ID, &p.ShopID, &p.Name, &p.TargetAmount, &p.CurrentAmount, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

// ========== DASHBOARD / COACH HELPERS ==========

func (r *Repository) SumRevenueLastDays(ctx context.Context, shopID uuid.UUID, days int) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_amount),0)
		FROM invoices
		WHERE shop_id = $1 AND created_at >= now() - ($2 || ' days')::interval
	`, shopID, days)

	var total float64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) SumExpensesLastDays(ctx context.Context, shopID uuid.UUID, days int) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0)
		FROM expenses
		WHERE shop_id = $1 AND spent_at >= now() - ($2 || ' days')::interval
	`, shopID, days)

	var total float64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
