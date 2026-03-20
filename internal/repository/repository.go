package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	ID           uuid.UUID
	Name         string
	Email        string
	APIKey       string
	PasswordHash *string
}

type Shop struct {
	ID        uuid.UUID
	Name      string
	Address   string
	GSTNumber string
	OwnerID   uuid.UUID
	CreatedAt time.Time
}

// ========== SUBSCRIPTIONS ==========

type Subscription struct {
	ID          int64
	ClientID    string
	Status      string
	ExpiresAt   *time.Time
	ReferenceID *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, api_key, password_hash
		FROM users
		WHERE email = $1
	`, email)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, api_key, password_hash
		FROM users
		WHERE api_key = $1
	`, apiKey)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, name, email, passwordHash string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	apiKey := uuid.NewString()

	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (name, email, api_key, password_hash)
		VALUES ($1,$2,$3,$4)
		RETURNING id, name, email, api_key, password_hash
	`, name, email, apiKey, passwordHash).Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.PasswordHash)
	if err != nil {
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
		VALUES ($1,$2,$3,$4)
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
		FROM shops
		WHERE id = $1
	`, id)

	var s Shop
	if err := row.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Address, &s.GSTNumber, &s.CreatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UserOwnsShop(ctx context.Context, userID uuid.UUID, shopID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM shops WHERE id = $1 AND owner_id = $2
		)
	`, shopID, userID).Scan(&exists)
	return exists, err
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
	`, p.ShopID, p.Name, p.SKU, p.Stock, p.CostPrice, p.SellingPrice, p.LowStockThreshold).
		Scan(&p.ID)
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
		FROM products
		WHERE shop_id = $1
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

func getProductForUpdate(ctx context.Context, tx pgx.Tx, productID uuid.UUID) (*Product, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, shop_id, name, sku, stock, cost_price, selling_price, low_stock_threshold
		FROM products
		WHERE id = $1
		FOR UPDATE
	`, productID)

	var p Product
	if err := row.Scan(&p.ID, &p.ShopID, &p.Name, &p.SKU, &p.Stock, &p.CostPrice, &p.SellingPrice, &p.LowStockThreshold); err != nil {
		return nil, err
	}
	return &p, nil
}

// ========== INVOICES ==========

type Invoice struct {
	ID             uuid.UUID
	ShopID         uuid.UUID
	CustomerName   string
	CustomerPhone  string
	Subtotal       float64
	TaxAmount      float64
	DiscountAmount float64
	TotalAmount    float64
	InvoiceNumber  string
	Status         string
	PaymentMethod  *string
	DueDate        *time.Time
	CreatedAt      time.Time
}

type InvoiceItem struct {
	ID        uuid.UUID
	InvoiceID uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}

func (r *Repository) CreateInvoiceWithItems(ctx context.Context, inv Invoice, items []InvoiceItem) (*Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for i := range items {
		item := &items[i]

		p, err := getProductForUpdate(ctx, tx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if item.Quantity <= 0 {
			return nil, errors.New("quantity must be positive")
		}
		if p.Stock < item.Quantity {
			return nil, errors.New("not enough stock for product: " + p.Name)
		}
		if p.ShopID != inv.ShopID {
			return nil, errors.New("product does not belong to shop: " + p.Name)
		}

		newStock := p.Stock - item.Quantity
		_, err = tx.Exec(ctx, `
			UPDATE products
			SET stock = $1
			WHERE id = $2
		`, newStock, p.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO invoices (shop_id, customer_name, customer_phone, subtotal, tax_amount, discount_amount, total_amount, payment_method, invoice_number, payment_status, due_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at
	`, inv.ShopID, inv.CustomerName, inv.CustomerPhone, inv.Subtotal, inv.TaxAmount, inv.DiscountAmount, inv.TotalAmount, inv.PaymentMethod, inv.InvoiceNumber, inv.Status, inv.DueDate).
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
		`, item.InvoiceID, item.ProductID, item.Quantity, item.UnitPrice).
			Scan(&item.ID)
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
		SELECT id, shop_id, customer_name, customer_phone, subtotal, tax_amount, discount_amount, total_amount,
		       COALESCE(invoice_number, '') AS invoice_number,
		       payment_status, payment_method, due_date, created_at
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
		if err := rows.Scan(
			&iv.ID,
			&iv.ShopID,
			&iv.CustomerName,
			&iv.CustomerPhone,
			&iv.Subtotal,
			&iv.TaxAmount,
			&iv.DiscountAmount,
			&iv.TotalAmount,
			&iv.InvoiceNumber,
			&iv.Status,
			&iv.PaymentMethod,
			&iv.DueDate,
			&iv.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, iv)
	}
	return result, nil
}

func (r *Repository) GenerateInvoiceNumber(ctx context.Context, shopID uuid.UUID) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var last int
	err := r.pool.QueryRow(ctx, `
        INSERT INTO invoice_counters (shop_id, last_number)
        VALUES ($1, 1)
        ON CONFLICT (shop_id)
        DO UPDATE SET last_number = invoice_counters.last_number + 1
        RETURNING last_number
    `, shopID).Scan(&last)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("INV-%05d", last), nil
}

func (r *Repository) GetInvoiceWithItems(ctx context.Context, invoiceID uuid.UUID) (*Invoice, []InvoiceItem, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var inv Invoice
	err := r.pool.QueryRow(ctx, `
		SELECT id, shop_id, customer_name, customer_phone, subtotal, tax_amount, discount_amount, total_amount,
		       COALESCE(invoice_number, '') AS invoice_number,
		       payment_status, payment_method, due_date, created_at
		FROM invoices
		WHERE id = $1
	`, invoiceID).Scan(
		&inv.ID,
		&inv.ShopID,
		&inv.CustomerName,
		&inv.CustomerPhone,
		&inv.Subtotal,
		&inv.TaxAmount,
		&inv.DiscountAmount,
		&inv.TotalAmount,
		&inv.InvoiceNumber,
		&inv.Status,
		&inv.PaymentMethod,
		&inv.DueDate,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, invoice_id, product_id, quantity, unit_price
		FROM invoice_items
		WHERE invoice_id = $1
	`, invoiceID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []InvoiceItem
	for rows.Next() {
		var it InvoiceItem
		if err := rows.Scan(&it.ID, &it.InvoiceID, &it.ProductID, &it.Quantity, &it.UnitPrice); err != nil {
			return nil, nil, err
		}
		items = append(items, it)
	}

	return &inv, items, nil
}

func (r *Repository) UpdateInvoiceStatus(ctx context.Context, invoiceID uuid.UUID, status, paymentMethod string) (*Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var inv Invoice
	err := r.pool.QueryRow(ctx, `
		UPDATE invoices
		SET payment_status = $1,
		    payment_method = COALESCE(NULLIF($2, ''), payment_method)
		WHERE id = $3
		RETURNING id, shop_id, customer_name, customer_phone, subtotal, tax_amount, discount_amount, total_amount,
		          COALESCE(invoice_number, '') AS invoice_number,
		          payment_status, payment_method, due_date, created_at
	`, status, paymentMethod, invoiceID).Scan(
		&inv.ID,
		&inv.ShopID,
		&inv.CustomerName,
		&inv.CustomerPhone,
		&inv.Subtotal,
		&inv.TaxAmount,
		&inv.DiscountAmount,
		&inv.TotalAmount,
		&inv.InvoiceNumber,
		&inv.Status,
		&inv.PaymentMethod,
		&inv.DueDate,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func (r *Repository) GetInvoiceFullData(ctx context.Context, invoiceID uuid.UUID) (*Invoice, []InvoiceItem, *Shop, map[string]string, error) {
	inv, items, err := r.GetInvoiceWithItems(ctx, invoiceID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	shop, err := r.GetShopByID(ctx, inv.ShopID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	products := make(map[string]string)
	for _, it := range items {
		var name string
		if err := r.pool.QueryRow(ctx, `SELECT name FROM products WHERE id = $1`, it.ProductID).Scan(&name); err == nil {
			products[it.ProductID.String()] = name
		}
	}

	return inv, items, shop, products, nil
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
	`, e.ShopID, e.Category, e.Amount, e.Note).
		Scan(&e.ID, &e.SpentAt)
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
	`, p.ShopID, p.Name, p.TargetAmount).
		Scan(&p.ID, &p.CurrentAmount, &p.CreatedAt)
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
	`, amount, potID).
		Scan(&p.ID, &p.ShopID, &p.Name, &p.TargetAmount, &p.CurrentAmount, &p.CreatedAt)
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
		WHERE shop_id = $1
		  AND created_at >= now() - ($2 || ' days')::interval
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
		WHERE shop_id = $1
		  AND spent_at >= now() - ($2 || ' days')::interval
	`, shopID, days)

	var total float64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ========== BILLING / SUBSCRIPTIONS ==========

func (r *Repository) GetSubscriptionByClientID(ctx context.Context, clientID string) (*Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, client_id, status, expires_at, reference_id, created_at, updated_at
		FROM subscriptions
		WHERE client_id = $1
	`, clientID)

	var sub Subscription
	if err := row.Scan(&sub.ID, &sub.ClientID, &sub.Status, &sub.ExpiresAt, &sub.ReferenceID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *Repository) UpsertPendingSubscription(ctx context.Context, clientID, referenceID string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions (client_id, status, expires_at, reference_id)
		VALUES ($1, 'inactive', NULL, $2)
		ON CONFLICT (client_id) DO UPDATE
		SET status = 'inactive',
			expires_at = NULL,
			reference_id = EXCLUDED.reference_id,
			updated_at = now()
	`, clientID, referenceID)
	return err
}

func (r *Repository) ActivateSubscription(ctx context.Context, clientID, referenceID string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions (client_id, status, expires_at, reference_id)
		VALUES ($1, 'active', $2, $3)
		ON CONFLICT (client_id) DO UPDATE
		SET status = 'active',
			expires_at = $2,
			reference_id = $3,
			updated_at = now()
	`, clientID, expiresAt, referenceID)
	return err
}
