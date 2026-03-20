package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"fintech-backend/internal/config"
	"fintech-backend/internal/dto"
	"fintech-backend/internal/pdf"
	"fintech-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	cfg  *config.Config
	repo *repository.Repository
}

func New(cfg *config.Config, repo *repository.Repository) *Service {
	return &Service{cfg: cfg, repo: repo}
}

// ===== Auth service =====

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type jwtCustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *Service) generateJWT(userID, email string) (string, error) {
	claims := jwtCustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("email and password are required")
	}

	if existing, _ := s.repo.GetUserByEmail(ctx, req.Email); existing != nil {
		return nil, errors.New("user already exists")
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, req.Name, req.Email, hash)
	if err != nil {
		return nil, err
	}

	token, err := s.generateJWT(user.ID.String(), user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.AuthUserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil || user.PasswordHash == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := checkPassword(*user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateJWT(user.ID.String(), user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.AuthUserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// ========== SHOPS ==========

func (s *Service) CreateShop(ctx context.Context, ownerID uuid.UUID, req dto.CreateShopRequest) (*repository.Shop, error) {
	return s.repo.CreateShop(ctx, ownerID, req.Name, req.Address, req.GSTNumber)
}

func (s *Service) ListShopsByUser(ctx context.Context, ownerID uuid.UUID) ([]repository.Shop, error) {
	return s.repo.ListShopsByUser(ctx, ownerID)
}

func (s *Service) EnsureOwnership(ctx context.Context, userID, shopID uuid.UUID) error {
	owns, err := s.repo.UserOwnsShop(ctx, userID, shopID)
	if err != nil {
		return err
	}
	if !owns {
		return errors.New("you do not have permission to access this shop")
	}
	return nil
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
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("invoice must have at least one item")
	}

	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id")
	}

	if _, err := s.repo.GetShopByID(ctx, shopID); err != nil {
		return nil, fmt.Errorf("shop not found")
	}

	var subtotal float64
	items := make([]repository.InvoiceItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be positive")
		}
		if it.UnitPrice <= 0 {
			return nil, fmt.Errorf("unit_price must be positive")
		}

		pid, err := uuid.Parse(it.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", it.ProductID)
		}

		lineTotal := float64(it.Quantity) * it.UnitPrice
		subtotal += lineTotal

		items = append(items, repository.InvoiceItem{
			ProductID: pid,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		})
	}

	if req.TaxAmount < 0 || req.DiscountAmount < 0 {
		return nil, fmt.Errorf("tax and discount cannot be negative")
	}

	total := subtotal + req.TaxAmount - req.DiscountAmount
	if total < 0 {
		total = 0
	}

	invoiceNumber, err := s.repo.GenerateInvoiceNumber(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("cannot generate invoice number: %w", err)
	}

	var dueTime *time.Time
	if strings.TrimSpace(req.DueDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate))
		if err != nil {
			return nil, fmt.Errorf("invalid due_date, expected YYYY-MM-DD")
		}
		dueTime = &t
	}

	status := "PAID"
	if dueTime != nil && total > 0 && time.Now().Before(*dueTime) {
		status = "DUE"
	}

	var paymentMethod *string
	if pm := strings.TrimSpace(req.PaymentMethod); pm != "" {
		paymentMethod = &pm
	}

	inv := repository.Invoice{
		ShopID:         shopID,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		Subtotal:       subtotal,
		TaxAmount:      req.TaxAmount,
		DiscountAmount: req.DiscountAmount,
		TotalAmount:    total,
		InvoiceNumber:  invoiceNumber,
		Status:         status,
		PaymentMethod:  paymentMethod,
		DueDate:        dueTime,
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

type InvoiceDetails struct {
	Invoice repository.Invoice       `json:"invoice"`
	Items   []repository.InvoiceItem `json:"items"`
}

func (s *Service) GetInvoiceDetails(ctx context.Context, invoiceIDStr string) (*InvoiceDetails, error) {
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice_id")
	}

	inv, items, err := s.repo.GetInvoiceWithItems(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	return &InvoiceDetails{
		Invoice: *inv,
		Items:   items,
	}, nil
}

func (s *Service) UpdateInvoiceStatus(ctx context.Context, invoiceIDStr string, req dto.UpdateInvoiceStatusRequest) (*repository.Invoice, error) {
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice_id")
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	switch status {
	case "PAID", "DUE", "PARTIALLY_PAID":
	default:
		return nil, fmt.Errorf("invalid status")
	}

	paymentMethod := strings.TrimSpace(req.PaymentMethod)

	return s.repo.UpdateInvoiceStatus(ctx, invoiceID, status, paymentMethod)
}

func (s *Service) GetInvoicePDF(ctx context.Context, userID uuid.UUID, invoiceID string) ([]byte, error) {
	id, err := uuid.Parse(invoiceID)
	if err != nil {
		return nil, errors.New("invalid invoice_id")
	}

	inv, items, shop, products, err := s.repo.GetInvoiceFullData(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.EnsureOwnership(ctx, userID, inv.ShopID); err != nil {
		return nil, err
	}

	return pdf.GenerateInvoicePDF(shop, inv, items, products)
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

// ========== BILLING / SUBSCRIPTIONS ==========

type BillingStatus struct {
	Active    bool       `json:"active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (s *Service) GetBillingStatus(ctx context.Context, clientID string) (*BillingStatus, error) {
	sub, err := s.repo.GetSubscriptionByClientID(ctx, clientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &BillingStatus{Active: false, ExpiresAt: nil}, nil
		}
		return nil, err
	}

	active := sub.Status == "active" && sub.ExpiresAt != nil && sub.ExpiresAt.After(time.Now())
	return &BillingStatus{Active: active, ExpiresAt: sub.ExpiresAt}, nil
}

func (s *Service) CreatePaymentLink(ctx context.Context, clientID string) (string, string, error) {
	if strings.TrimSpace(clientID) == "" {
		return "", "", fmt.Errorf("client_id is required")
	}

	referenceID := clientID
	if len(referenceID) > 40 {
		referenceID = referenceID[:40]
	}

	if err := s.repo.UpsertPendingSubscription(ctx, clientID, referenceID); err != nil {
		return "", "", err
	}

	reqBody := map[string]interface{}{
		"amount":       19900, // paise
		"currency":     "INR",
		"description":  "Vantro Premium (Monthly)",
		"reference_id": referenceID,
		"notes": map[string]string{
			"client_id": clientID,
		},
	}

	buf, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/payment_links", bytes.NewReader(buf))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(s.cfg.RazorpayKeyID, s.cfg.RazorpayKeySecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var respBody struct {
		ShortURL    string `json:"short_url"`
		ReferenceID string `json:"reference_id"`
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var raw map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		return "", "", fmt.Errorf("razorpay error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", "", err
	}

	if respBody.ShortURL == "" {
		return "", "", fmt.Errorf("missing short_url from razorpay")
	}

	return respBody.ShortURL, respBody.ReferenceID, nil
}

func (s *Service) HandleBillingWebhook(ctx context.Context, body []byte, signature string) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.RazorpayWebhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		log.Printf("razorpay webhook: signature mismatch expected=%s got=%s", expected, signature)
		return fmt.Errorf("invalid signature")
	}

	var payload struct {
		Event   string `json:"event"`
		Payload struct {
			PaymentLink struct {
				Entity struct {
					ReferenceID string `json:"reference_id"`
				} `json:"entity"`
			} `json:"payment_link"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	if payload.Event != "payment_link.paid" {
		return nil
	}

	clientID := payload.Payload.PaymentLink.Entity.ReferenceID
	if clientID == "" {
		return fmt.Errorf("missing reference_id in webhook")
	}

	expires := time.Now().Add(30 * 24 * time.Hour)
	return s.repo.ActivateSubscription(ctx, clientID, clientID, expires)
}
