package business

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{DB: db} }

func (r *Repo) CreateWithDefaults(ctx context.Context, req CreateBusinessRequest) (*CreateBusinessResponse, error) {
	if req.OwnerUserID == "" {
		return nil, fmt.Errorf("owner_user_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	currency := "INR"
	if req.Currency != nil && *req.Currency != "" {
		currency = *req.Currency
	}

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1) create business
	var businessID string
	qBusiness := `
INSERT INTO businesses (owner_user_id, name, industry, currency)
VALUES ($1,$2,$3,$4)
RETURNING id
`
	if err := tx.QueryRow(ctx, qBusiness, req.OwnerUserID, req.Name, req.Industry, currency).Scan(&businessID); err != nil {
		return nil, err
	}

	// 2) create default cash account
	var accountID string
	qAccount := `
INSERT INTO accounts (business_id, name, type, currency, opening_balance)
VALUES ($1,'Cash','cash',$2,0)
RETURNING id
`
	if err := tx.QueryRow(ctx, qAccount, businessID, currency).Scan(&accountID); err != nil {
		return nil, err
	}

	// 3) seed default categories (idempotent-ish by unique constraint)
	defaultIncome := []string{"Sales", "Other Income"}
	defaultExpense := []string{"Rent", "Supplies", "Marketing", "Travel", "Food", "Utilities", "Salary", "Other Expense"}

	qCat := `
INSERT INTO categories (business_id, name, kind)
VALUES ($1,$2,$3)
ON CONFLICT (business_id, kind, name) DO NOTHING
`
	for _, name := range defaultIncome {
		if _, err := tx.Exec(ctx, qCat, businessID, name, "income"); err != nil {
			return nil, err
		}
	}
	for _, name := range defaultExpense {
		if _, err := tx.Exec(ctx, qCat, businessID, name, "expense"); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateBusinessResponse{
		BusinessID: businessID,
		AccountID:  accountID,
		Message:    "Business created with default cash account + categories",
	}, nil
}
