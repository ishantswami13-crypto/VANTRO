package service

import (
	"context"
	"fmt"

	"fintech-backend/internal/dto"
	"fintech-backend/internal/repository"

	"github.com/google/uuid"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// ========== SHOPS ==========

func (s *Service) CreateShop(ctx context.Context, req dto.CreateShopRequest) (*repository.Shop, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("owner not found: %w", err)
	}
	return s.repo.CreateShop(ctx, user.ID, req.Name, req.Address, req.GSTNumber)
}

func (s *Service) ListShops(ctx context.Context, apiKey string) ([]repository.Shop, error) {
	user, err := s.repo.GetUserByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return s.repo.ListShopsByUser(ctx, user.ID)
}

// ========== PRODUCTS ==========

func (s *Service) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*repository.Product, error) {
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	if _, err := s.repo.GetShopByID(ctx, shopID); err != nil {
		return nil, err
	}
	var sku *string
	if req.SKU != "" {
		sku = &req.SKU
	}
	p := repository.Product{
		ShopID:            shopID,
		Name:              req.Name,
		SKU:               sku,
		Stock:             req.Stock,
		CostPrice:         req.CostPrice,
		SellingPrice:      req.SellingPrice,
		LowStockThreshold: req.LowStockThreshold,
	}
	return s.repo.CreateProduct(ctx, p)
}

func (s *Service) ListProducts(ctx context.Context, shopIDStr string) ([]repository.Product, error) {
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	return s.repo.ListProductsByShop(ctx, shopID)
}

// ========== INVOICES ==========

func (s *Service) CreateInvoice(ctx context.Context, req dto.CreateInvoiceRequest) (*repository.Invoice, error) {
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("invoice must have at least one item")
	}

	var total float64
	items := make([]repository.InvoiceItem, 0, len(req.Items))
	for _, it := range req.Items {
		pID, err := uuid.Parse(it.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", it.ProductID)
		}
		lineTotal := it.UnitPrice * float64(it.Quantity)
		total += lineTotal
		items = append(items, repository.InvoiceItem{
			ProductID: pID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		})
	}

	inv := repository.Invoice{
		ShopID:        shopID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		TotalAmount:   total + req.TaxAmount,
		TaxAmount:     req.TaxAmount,
		Status:        "PAID",
	}
	return s.repo.CreateInvoiceWithItems(ctx, inv, items)
}

func (s *Service) ListInvoices(ctx context.Context, shopIDStr string) ([]repository.Invoice, error) {
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	return s.repo.ListInvoicesByShop(ctx, shopID)
}

// ========== EXPENSES ==========

func (s *Service) CreateExpense(ctx context.Context, req dto.CreateExpenseRequest) (*repository.Expense, error) {
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	e := repository.Expense{
		ShopID:   shopID,
		Category: req.Category,
		Amount:   req.Amount,
	}
	if req.Note != "" {
		note := req.Note
		e.Note = &note
	}
	return s.repo.CreateExpense(ctx, e)
}

func (s *Service) ListExpenses(ctx context.Context, shopIDStr string) ([]repository.Expense, error) {
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	return s.repo.ListExpensesByShop(ctx, shopID)
}

// ========== POTS ==========

func (s *Service) CreatePot(ctx context.Context, req dto.CreatePotRequest) (*repository.Pot, error) {
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	p := repository.Pot{
		ShopID:       shopID,
		Name:         req.Name,
		TargetAmount: req.TargetAmount,
	}
	return s.repo.CreatePot(ctx, p)
}

func (s *Service) DepositPot(ctx context.Context, potIDStr string, amount float64) (*repository.Pot, error) {
	potID, err := uuid.Parse(potIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pot_id")
	}
	return s.repo.DepositToPot(ctx, potID, amount)
}

func (s *Service) ListPots(ctx context.Context, shopIDStr string) ([]repository.Pot, error) {
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}
	return s.repo.ListPotsByShop(ctx, shopID)
}

// ========== DASHBOARD / COACH ==========

type DashboardSummary struct {
	Last7DaysRevenue   float64 `json:"last_7_days_revenue"`
	Last7DaysExpenses  float64 `json:"last_7_days_expenses"`
	Last30DaysRevenue  float64 `json:"last_30_days_revenue"`
	Last30DaysExpenses float64 `json:"last_30_days_expenses"`
	NetLast30Days      float64 `json:"net_last_30_days"`
}

func (s *Service) GetDashboardSummary(ctx context.Context, shopIDStr string) (*DashboardSummary, error) {
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}

	r7, err := s.repo.SumRevenueLastDays(ctx, shopID, 7)
	if err != nil {
		return nil, err
	}
	e7, err := s.repo.SumExpensesLastDays(ctx, shopID, 7)
	if err != nil {
		return nil, err
	}
	r30, err := s.repo.SumRevenueLastDays(ctx, shopID, 30)
	if err != nil {
		return nil, err
	}
	e30, err := s.repo.SumExpensesLastDays(ctx, shopID, 30)
	if err != nil {
		return nil, err
	}
	return &DashboardSummary{
		Last7DaysRevenue:   r7,
		Last7DaysExpenses:  e7,
		Last30DaysRevenue:  r30,
		Last30DaysExpenses: e30,
		NetLast30Days:      r30 - e30,
	}, nil
}

type CoachInsight struct {
	Message string `json:"message"`
}

func (s *Service) GetCoachInsights(ctx context.Context, shopIDStr string) ([]CoachInsight, error) {
	summary, err := s.GetDashboardSummary(ctx, shopIDStr)
	if err != nil {
		return nil, err
	}

	insights := []CoachInsight{}

	if summary.Last30DaysRevenue == 0 {
		insights = append(insights, CoachInsight{
			Message: "No revenue in the last 30 days. Try recording invoices regularly.",
		})
	}

	if summary.Last30DaysExpenses > 0 && summary.Last30DaysRevenue > 0 {
		expenseRatio := summary.Last30DaysExpenses / summary.Last30DaysRevenue
		if expenseRatio > 0.7 {
			insights = append(insights, CoachInsight{
				Message: "Expenses are more than 70% of revenue in the last 30 days. Time to cut some costs.",
			})
		} else if expenseRatio < 0.3 {
			insights = append(insights, CoachInsight{
				Message: "Expenses are under 30% of revenue. Good profitability, consider reinvesting into growth.",
			})
		}
	}

	if summary.NetLast30Days > 0 {
		insights = append(insights, CoachInsight{
			Message: "You are net positive this month. Allocate part of your profits into a savings pot.",
		})
	} else if summary.NetLast30Days < 0 {
		insights = append(insights, CoachInsight{
			Message: "You are net negative this month. Review high-cost categories and low-margin products.",
		})
	}

	if len(insights) == 0 {
		insights = append(insights, CoachInsight{
			Message: "Data is limited. Add more invoices and expenses to unlock better insights.",
		})
	}

	return insights, nil
}
