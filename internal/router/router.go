package router

import (
	"context"
	"net/http"
	"sync"
	"time"

	"fintech-backend/internal/business"
	"fintech-backend/internal/config"
	"fintech-backend/internal/dto"
	"fintech-backend/internal/middleware"
	"fintech-backend/internal/repository"
	"fintech-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type demoTransaction struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Category  string    `json:"category,omitempty"`
	Date      string    `json:"date,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

var txStore = struct {
	sync.Mutex
	items []demoTransaction
}{items: []demoTransaction{}}

type mobileExpense struct {
	ID          string `json:"id"`
	AmountCents int    `json:"amount_cents"`
	Category    string `json:"category"`
	Mood        string `json:"mood,omitempty"`
	Note        string `json:"note,omitempty"`
	SpentAt     string `json:"spent_at"`
}

type mobilePot struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TargetCents int    `json:"target_cents"`
	SavedCents  int    `json:"saved_cents"`
}

var mobileExpenseStore = struct {
	sync.Mutex
	items []mobileExpense
}{items: []mobileExpense{}}

var mobilePotStore = struct {
	sync.Mutex
	items []mobilePot
}{items: []mobilePot{}}

func New(cfg *config.Config, pool *pgxpool.Pool, middlewares ...fiber.Handler) *fiber.App {
	app := fiber.New()

	for _, mw := range middlewares {
		app.Use(mw)
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Use(middleware.CORS())
	app.Use(middleware.APIKeyAuth(cfg))

	v1 := app.Group("/v1")
	mobile := v1.Group("/mobile")

	// Lightweight transactions endpoints (in-memory demo for UI) with DB-availability guard
	v1.Get("/transactions", func(c *fiber.Ctx) error {
		if pool == nil {
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "database unavailable"})
		}
		txStore.Lock()
		defer txStore.Unlock()
		return c.JSON(txStore.items)
	})

	v1.Post("/transactions", func(c *fiber.Ctx) error {
		if pool == nil {
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "database unavailable"})
		}
		var body struct {
			Title    string  `json:"title"`
			Amount   float64 `json:"amount"`
			Type     string  `json:"type"`
			Category string  `json:"category"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.Title == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
		}
		if body.Amount <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "amount must be positive"})
		}
		if body.Type != "income" && body.Type != "expense" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "type must be income or expense"})
		}
		now := time.Now()
		tx := demoTransaction{
			ID:        uuid.NewString(),
			Title:     body.Title,
			Amount:    body.Amount,
			Type:      body.Type,
			Category:  body.Category,
			Date:      now.Format("2006-01-02"),
			CreatedAt: now,
		}
		txStore.Lock()
		txStore.items = append([]demoTransaction{tx}, txStore.items...)
		txStore.Unlock()
		return c.Status(http.StatusCreated).JSON(tx)
	})

	v1.Delete("/transactions/:id", func(c *fiber.Ctx) error {
		if pool == nil {
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "database unavailable"})
		}
		id := c.Params("id")
		if id == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "id required"})
		}
		txStore.Lock()
		defer txStore.Unlock()
		found := false
		filtered := make([]demoTransaction, 0, len(txStore.items))
		for _, it := range txStore.items {
			if it.ID == id {
				found = true
				continue
			}
			filtered = append(filtered, it)
		}
		txStore.items = filtered
		if !found {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}
		return c.JSON(fiber.Map{"deleted": true})
	})

	// Mobile compatibility endpoints used by the Flutter app.
	mobile.Get("/expenses", func(c *fiber.Ctx) error {
		var fromDate time.Time
		var toDate time.Time
		var err error

		if from := c.Query("from"); from != "" {
			fromDate, err = time.Parse("2006-01-02", from)
			if err != nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "from must be YYYY-MM-DD"})
			}
		}
		if to := c.Query("to"); to != "" {
			toDate, err = time.Parse("2006-01-02", to)
			if err != nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "to must be YYYY-MM-DD"})
			}
			toDate = toDate.Add(24*time.Hour - time.Nanosecond)
		}

		mobileExpenseStore.Lock()
		defer mobileExpenseStore.Unlock()

		filtered := make([]mobileExpense, 0, len(mobileExpenseStore.items))
		for _, item := range mobileExpenseStore.items {
			spentAt, parseErr := time.Parse(time.RFC3339, item.SpentAt)
			if parseErr != nil {
				continue
			}
			if !fromDate.IsZero() && spentAt.Before(fromDate) {
				continue
			}
			if !toDate.IsZero() && spentAt.After(toDate) {
				continue
			}
			filtered = append(filtered, item)
		}

		return c.JSON(filtered)
	})

	mobile.Post("/expenses", func(c *fiber.Ctx) error {
		var body struct {
			AmountCents int    `json:"amount_cents"`
			Category    string `json:"category"`
			Mood        string `json:"mood"`
			Note        string `json:"note"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.AmountCents <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "amount_cents must be positive"})
		}
		if body.Category == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "category is required"})
		}

		item := mobileExpense{
			ID:          uuid.NewString(),
			AmountCents: body.AmountCents,
			Category:    body.Category,
			Mood:        body.Mood,
			Note:        body.Note,
			SpentAt:     time.Now().UTC().Format(time.RFC3339),
		}

		mobileExpenseStore.Lock()
		mobileExpenseStore.items = append([]mobileExpense{item}, mobileExpenseStore.items...)
		mobileExpenseStore.Unlock()

		return c.Status(http.StatusCreated).JSON(item)
	})

	mobile.Get("/pots", func(c *fiber.Ctx) error {
		mobilePotStore.Lock()
		defer mobilePotStore.Unlock()
		return c.JSON(mobilePotStore.items)
	})

	mobile.Post("/pots", func(c *fiber.Ctx) error {
		var body struct {
			Name        string `json:"name"`
			TargetCents int    `json:"target_cents"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.Name == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
		}
		if body.TargetCents <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "target_cents must be positive"})
		}

		item := mobilePot{
			ID:          uuid.NewString(),
			Name:        body.Name,
			TargetCents: body.TargetCents,
			SavedCents:  0,
		}

		mobilePotStore.Lock()
		mobilePotStore.items = append([]mobilePot{item}, mobilePotStore.items...)
		mobilePotStore.Unlock()

		return c.Status(http.StatusCreated).JSON(item)
	})

	mobile.Patch("/pots/:id", func(c *fiber.Ctx) error {
		var body struct {
			AddCents int `json:"add_cents"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.AddCents <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "add_cents must be positive"})
		}

		id := c.Params("id")
		mobilePotStore.Lock()
		defer mobilePotStore.Unlock()
		for i := range mobilePotStore.items {
			if mobilePotStore.items[i].ID != id {
				continue
			}
			mobilePotStore.items[i].SavedCents += body.AddCents
			return c.JSON(mobilePotStore.items[i])
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "pot not found"})
	})

	mobile.Post("/coach/plan", func(c *fiber.Ctx) error {
		var body struct {
			IncomeCents int    `json:"income_cents"`
			RentCents   int    `json:"rent_cents"`
			Goal        string `json:"goal"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}

		disposable := body.IncomeCents - body.RentCents
		healthScore := 55
		if body.IncomeCents > 0 {
			switch {
			case disposable > body.IncomeCents/2:
				healthScore = 88
			case disposable > body.IncomeCents/3:
				healthScore = 74
			case disposable > 0:
				healthScore = 63
			default:
				healthScore = 41
			}
		}

		goal := body.Goal
		if goal == "" {
			goal = "your next savings goal"
		}

		rules := []string{
			"Track every expense daily for cleaner insights.",
			"Move at least 10% of income into a savings pot first.",
		}
		if disposable > 0 {
			rules = append(rules, "Keep rent and fixed costs below 50% of income where possible.")
		} else {
			rules = append(rules, "Reduce fixed costs this month before increasing lifestyle spending.")
		}

		dailyNudge := "Spend intentionally today."
		if disposable > 0 {
			dailyNudge = "Set aside a small amount for " + goal + " before discretionary spending."
		}

		weekStart := time.Now().UTC()
		for weekStart.Weekday() != time.Monday {
			weekStart = weekStart.AddDate(0, 0, -1)
		}

		return c.JSON(fiber.Map{
			"week_start":   weekStart.Format("2006-01-02"),
			"rules":        rules,
			"daily_nudge":  dailyNudge,
			"health_score": healthScore,
		})
	})

	// Simple demo UI at /
	app.Get("/", func(c *fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Fintech Demo - Mini AI CFO</title>
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <style>
    body {
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #050816;
      color: #e5e7eb;
      padding: 16px;
    }
    h1, h2 {
      color: #f9fafb;
    }
    .card {
      background: #111827;
      border-radius: 12px;
      padding: 16px;
      margin-bottom: 16px;
      border: 1px solid #1f2933;
    }
    label {
      font-size: 12px;
      color: #9ca3af;
      display: block;
      margin-bottom: 4px;
    }
    input {
      width: 100%;
      padding: 8px;
      margin-bottom: 8px;
      border-radius: 8px;
      border: 1px solid #374151;
      background: #020617;
      color: #e5e7eb;
    }
    button {
      padding: 8px 14px;
      border-radius: 999px;
      border: none;
      cursor: pointer;
      font-size: 14px;
      background: linear-gradient(to right, #22c55e, #16a34a);
      color: #020617;
      font-weight: 600;
      margin-right: 8px;
      margin-bottom: 8px;
    }
    button.secondary {
      background: #111827;
      color: #e5e7eb;
      border: 1px solid #374151;
    }
    pre {
      background: #020617;
      border-radius: 8px;
      padding: 8px;
      font-size: 12px;
      overflow-x: auto;
      border: 1px solid #111827;
    }
    .row {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 16px;
    }
    .tag {
      display: inline-block;
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 999px;
      border: 1px solid #374151;
      color: #9ca3af;
      margin-left: 8px;
    }
  </style>
</head>
<body>
  <h1>Fintech Backend Demo <span class="tag">Mini AI CFO</span></h1>
  <p style="font-size:13px;color:#9ca3af;margin-bottom:16px;">
    This page talks to your Go + Postgres backend. Create a shop, add products and see JSON responses live.
  </p>

  <div class="row">
    <!-- SHOPS CARD -->
    <div class="card">
      <h2>1. Shops</h2>
      <label>Owner Email (must exist in DB, default: admin@example.com)</label>
      <input id="ownerEmail" placeholder="admin@example.com" value="admin@example.com" />

      <label>Shop Name</label>
      <input id="shopName" placeholder="Bablu Enterprises" />

      <label>Address</label>
      <input id="shopAddress" placeholder="Janakpuri, New Delhi" />

      <label>GST Number</label>
      <input id="shopGST" placeholder="07ABCDE1234F1Z5" />

      <button onclick="createShop()">Create Shop</button>
      <button class="secondary" onclick="listShops()">List My Shops</button>

      <p style="font-size:12px;color:#9ca3af;margin-top:8px;">
        Selected Shop ID: <span id="selectedShopId" style="color:#22c55e;">(none)</span>
      </p>

      <pre id="shopsOutput">// Shops responses will appear here</pre>
    </div>

    <!-- PRODUCTS CARD -->
    <div class="card">
      <h2>2. Products</h2>
      <p style="font-size:12px;color:#9ca3af;">
        Uses the selected Shop ID from above.
      </p>

      <label>Product Name</label>
      <input id="productName" placeholder="Jack F4 Sewing Machine" />

      <label>SKU</label>
      <input id="productSKU" placeholder="JACK-F4" />

      <label>Stock</label>
      <input id="productStock" type="number" value="5" />

      <label>Cost Price</label>
      <input id="productCost" type="number" value="18000" />

      <label>Selling Price</label>
      <input id="productSell" type="number" value="22000" />

      <label>Low Stock Threshold</label>
      <input id="productLow" type="number" value="2" />

      <button onclick="createProduct()">Create Product</button>
      <button class="secondary" onclick="listProducts()">List Products</button>

      <pre id="productsOutput">// Products responses will appear here</pre>
    </div>
  </div>

  <script>
    const API_KEY = "supersecretapikey";
    const BASE_URL = window.location.origin;

    let selectedShopId = null;

    function setSelectedShop(id) {
      selectedShopId = id;
      document.getElementById("selectedShopId").textContent = id || "(none)";
    }

    async function createShop() {
      const ownerEmail = document.getElementById("ownerEmail").value.trim();
      const name = document.getElementById("shopName").value.trim();
      const address = document.getElementById("shopAddress").value.trim();
      const gst = document.getElementById("shopGST").value.trim();

      if (!ownerEmail || !name) {
        alert("Owner email and shop name are required");
        return;
      }

      try {
        const res = await fetch(BASE_URL + "/api/shops", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-API-Key": API_KEY
          },
          body: JSON.stringify({
            owner_email: ownerEmail,
            name,
            address,
            gst_number: gst
          })
        });

        const data = await res.json();
        document.getElementById("shopsOutput").textContent = JSON.stringify(data, null, 2);

        if (data.id) {
          setSelectedShop(data.id);
        }
      } catch (err) {
        document.getElementById("shopsOutput").textContent = "Error: " + err.message;
      }
    }

    async function listShops() {
      try {
        const res = await fetch(BASE_URL + "/api/shops", {
          headers: {
            "X-API-Key": API_KEY
          }
        });
        const data = await res.json();
        document.getElementById("shopsOutput").textContent = JSON.stringify(data, null, 2);

        if (Array.isArray(data) && data.length > 0) {
          setSelectedShop(data[0].id);
        }
      } catch (err) {
        document.getElementById("shopsOutput").textContent = "Error: " + err.message;
      }
    }

    async function createProduct() {
      if (!selectedShopId) {
        alert("Select or create a shop first");
        return;
      }

      const name = document.getElementById("productName").value.trim();
      const sku = document.getElementById("productSKU").value.trim();
      const stock = parseInt(document.getElementById("productStock").value || "0", 10);
      const cost = parseFloat(document.getElementById("productCost").value || "0");
      const sell = parseFloat(document.getElementById("productSell").value || "0");
      const low = parseInt(document.getElementById("productLow").value || "0", 10);

      if (!name) {
        alert("Product name is required");
        return;
      }

      try {
        const res = await fetch(BASE_URL + "/api/products", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-API-Key": API_KEY
          },
          body: JSON.stringify({
            shop_id: selectedShopId,
            name,
            sku,
            stock,
            cost_price: cost,
            selling_price: sell,
            low_stock_threshold: low
          })
        });

        const data = await res.json();
        document.getElementById("productsOutput").textContent = JSON.stringify(data, null, 2);
      } catch (err) {
        document.getElementById("productsOutput").textContent = "Error: " + err.message;
      }
    }

    async function listProducts() {
      if (!selectedShopId) {
        alert("Select or create a shop first");
        return;
      }

      try {
        const res = await fetch(BASE_URL + "/api/shops/" + selectedShopId + "/products", {
          headers: {
            "X-API-Key": API_KEY
          }
        });

        const data = await res.json();
        document.getElementById("productsOutput").textContent = JSON.stringify(data, null, 2);
      } catch (err) {
        document.getElementById("productsOutput").textContent = "Error: " + err.message;
      }
    }
  </script>
</body>
</html>`
		return c.Type("html").SendString(html)
	})

	if pool == nil {
		return app
	}

	repo := repository.New(pool)
	svc := service.New(cfg, repo)
	bizRepo := business.NewRepo(pool)
	bizHandler := business.NewHandler(bizRepo)

	// Billing (public, client_id based)
	app.Get("/v1/billing/status", func(c *fiber.Ctx) error {
		clientID := c.Query("client_id")
		if clientID == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "client_id is required"})
		}
		status, err := svc.GetBillingStatus(c.Context(), clientID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})

	app.Post("/v1/billing/create-link", func(c *fiber.Ctx) error {
		var body struct {
			ClientID string `json:"client_id"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.ClientID == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "client_id is required"})
		}
		shortURL, ref, err := svc.CreatePaymentLink(c.Context(), body.ClientID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"short_url":    shortURL,
			"reference_id": ref,
		})
	})

	app.Post("/v1/billing/webhook", func(c *fiber.Ctx) error {
		signature := c.Get("X-Razorpay-Signature")
		if err := svc.HandleBillingWebhook(c.Context(), c.Body(), signature); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true})
	})

	// Business onboarding (creates business + default cash account + categories)
	app.Post("/v1/businesses", bizHandler.Create)

	// AUTH routes
	app.Post("/auth/signup", func(c *fiber.Ctx) error {
		var req dto.SignupRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		resp, err := svc.Signup(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(resp)
	})

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req dto.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		resp, err := svc.Login(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(resp)
	})

	authGroup := app.Group("/auth", middleware.JWTAuth(cfg))
	authGroup.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":    c.Locals("user_id"),
			"user_email": c.Locals("user_email"),
		})
	})

	// protected routes
	api := app.Group("/api", middleware.JWTAuth(cfg))

	// SHOPS
	api.Post("/shops", func(c *fiber.Ctx) error {
		var req dto.CreateShopRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		ownerIDStr, ok := c.Locals("user_id").(string)
		if !ok || ownerIDStr == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		ownerID, err := uuid.Parse(ownerIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shop, err := svc.CreateShop(c.Context(), ownerID, req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(shop)
	})

	api.Get("/shops", func(c *fiber.Ctx) error {
		ownerIDStr, ok := c.Locals("user_id").(string)
		if !ok || ownerIDStr == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		ownerID, err := uuid.Parse(ownerIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shops, err := svc.ListShopsByUser(c.Context(), ownerID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(shops)
	})

	// PRODUCTS
	api.Post("/products", func(c *fiber.Ctx) error {
		var req dto.CreateProductRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		p, err := svc.CreateProduct(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(p)
	})

	api.Get("/shops/:shopId/products", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		ps, err := svc.ListProducts(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ps)
	})

	// INVOICES
	api.Post("/invoices", func(c *fiber.Ctx) error {
		var req dto.CreateInvoiceRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		inv, err := svc.CreateInvoice(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inv)
	})

	api.Get("/shops/:shopId/invoices", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		invs, err := svc.ListInvoices(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(invs)
	})

	api.Get("/invoices/:invoiceId", func(c *fiber.Ctx) error {
		invoiceID := c.Params("invoiceId")
		details, err := svc.GetInvoiceDetails(context.Background(), invoiceID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(details)
	})

	api.Patch("/invoices/:invoiceId/status", func(c *fiber.Ctx) error {
		invoiceID := c.Params("invoiceId")
		var req dto.UpdateInvoiceStatusRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		inv, err := svc.UpdateInvoiceStatus(context.Background(), invoiceID, req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inv)
	})

	api.Get("/invoices/:invoiceId/pdf", func(c *fiber.Ctx) error {
		invoiceID := c.Params("invoiceId")
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		pdfBytes, err := svc.GetInvoicePDF(c.Context(), userID, invoiceID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=invoice.pdf")
		return c.Send(pdfBytes)
	})

	// EXPENSES
	api.Post("/expenses", func(c *fiber.Ctx) error {
		var req dto.CreateExpenseRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		e, err := svc.CreateExpense(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(e)
	})

	api.Get("/shops/:shopId/expenses", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		es, err := svc.ListExpenses(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(es)
	})

	// POTS
	api.Post("/pots", func(c *fiber.Ctx) error {
		var req dto.CreatePotRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		p, err := svc.CreatePot(c.Context(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(p)
	})

	api.Patch("/pots/:potId/deposit", func(c *fiber.Ctx) error {
		potID := c.Params("potId")
		var req dto.DepositPotRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		p, err := svc.DepositPot(context.Background(), potID, req.Amount)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(p)
	})

	api.Get("/shops/:shopId/pots", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		ps, err := svc.ListPots(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ps)
	})

	// DASHBOARD
	api.Get("/shops/:shopId/dashboard", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		summary, err := svc.GetDashboardSummary(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(summary)
	})

	// COACH
	api.Get("/shops/:shopId/coach", func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user context"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		shopIDParam := c.Params("shopId")
		shopID, err := uuid.Parse(shopIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid shop_id"})
		}
		if err := svc.EnsureOwnership(c.Context(), userID, shopID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		insights, err := svc.GetCoachInsights(c.Context(), shopIDParam)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(insights)
	})

	return app
}
