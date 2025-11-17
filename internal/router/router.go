package router

import (
	"context"
	"net/http"

	"fintech-backend/internal/config"
	"fintech-backend/internal/dto"
	"fintech-backend/internal/middleware"
	"fintech-backend/internal/repository"
	"fintech-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg *config.Config, pool *pgxpool.Pool) *fiber.App {
	app := fiber.New()

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

	// health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// protected routes
	api := app.Group("/api", middleware.APIKeyAuth(cfg))

	repo := repository.New(pool)
	svc := service.New(repo)

	// SHOPS
	api.Post("/shops", func(c *fiber.Ctx) error {
		var req dto.CreateShopRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		shop, err := svc.CreateShop(context.Background(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(shop)
	})

	api.Get("/shops", func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		shops, err := svc.ListShops(context.Background(), apiKey)
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
		p, err := svc.CreateProduct(context.Background(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(p)
	})

	api.Get("/shops/:shopId/products", func(c *fiber.Ctx) error {
		shopID := c.Params("shopId")
		ps, err := svc.ListProducts(context.Background(), shopID)
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
		inv, err := svc.CreateInvoice(context.Background(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inv)
	})

	api.Get("/shops/:shopId/invoices", func(c *fiber.Ctx) error {
		shopID := c.Params("shopId")
		invs, err := svc.ListInvoices(context.Background(), shopID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(invs)
	})

	// EXPENSES
	api.Post("/expenses", func(c *fiber.Ctx) error {
		var req dto.CreateExpenseRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		e, err := svc.CreateExpense(context.Background(), req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(e)
	})

	api.Get("/shops/:shopId/expenses", func(c *fiber.Ctx) error {
		shopID := c.Params("shopId")
		es, err := svc.ListExpenses(context.Background(), shopID)
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
		p, err := svc.CreatePot(context.Background(), req)
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
		shopID := c.Params("shopId")
		ps, err := svc.ListPots(context.Background(), shopID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ps)
	})

	// DASHBOARD
	api.Get("/shops/:shopId/dashboard", func(c *fiber.Ctx) error {
		shopID := c.Params("shopId")
		summary, err := svc.GetDashboardSummary(context.Background(), shopID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(summary)
	})

	// COACH
	api.Get("/shops/:shopId/coach", func(c *fiber.Ctx) error {
		shopID := c.Params("shopId")
		insights, err := svc.GetCoachInsights(context.Background(), shopID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(insights)
	})

	return app
}
