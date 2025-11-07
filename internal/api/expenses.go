package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Expense struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	AmountCents int       `json:"amount_cents"`
	Category    string    `json:"category"`
	Mood        string    `json:"mood"`
	Note        string    `json:"note"`
	SpentAt     time.Time `json:"spent_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type ExpenseRequest struct {
	AmountCents int    `json:"amount_cents"`
	Category    string `json:"category"`
	Mood        string `json:"mood"`
	Note        string `json:"note"`
}

func RegisterExpenseRoutes(e *echo.Group, db *pgxpool.Pool) {
	e.POST("/expenses", func(c echo.Context) error {
		var req ExpenseRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
		}

		// TODO: Replace with real user_id from auth
		userID := uuid.New()

		id := uuid.New()
		_, err := db.Exec(c.Request().Context(),
			`INSERT INTO expenses (id, user_id, amount_cents, category, mood, note) 
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			id, userID, req.AmountCents, req.Category, req.Mood, req.Note,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, echo.Map{"id": id})
	})

	e.GET("/expenses", func(c echo.Context) error {
		from := c.QueryParam("from")
		to := c.QueryParam("to")

		// TODO: Replace with real user id from auth
		userID := uuid.New()

		rows, err := db.Query(c.Request().Context(),
			`SELECT id, user_id, amount_cents, category, mood, note, spent_at, created_at
			 FROM expenses 
			 WHERE user_id = $1 AND spent_at BETWEEN $2 AND $3
			 ORDER BY spent_at DESC`,
			userID, from, to,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		defer rows.Close()

		var expenses []Expense
		for rows.Next() {
			var exp Expense
			if err := rows.Scan(&exp.ID, &exp.UserID, &exp.AmountCents, &exp.Category, &exp.Mood, &exp.Note, &exp.SpentAt, &exp.CreatedAt); err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
			expenses = append(expenses, exp)
		}

		return c.JSON(http.StatusOK, expenses)
	})
}
