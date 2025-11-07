package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type SavingPot struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	TargetCents int       `json:"target_cents"`
	SavedCents  int       `json:"saved_cents"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePotRequest struct {
	Name        string `json:"name"`
	TargetCents int    `json:"target_cents"`
}

type UpdatePotRequest struct {
	AddCents *int `json:"add_cents"`
}

func RegisterPotRoutes(e *echo.Group, db *pgxpool.Pool) {
	e.POST("/pots", func(c echo.Context) error {
		var req CreatePotRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
		}

		if req.Name == "" || req.TargetCents <= 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid values"})
		}

		userID := uuid.New() // TODO: Replace with actual auth user id
		id := uuid.New()

		_, err := db.Exec(c.Request().Context(),
			`INSERT INTO saving_pots (id, user_id, name, target_cents, saved_cents) 
			 VALUES ($1, $2, $3, $4, 0)`,
			id, userID, req.Name, req.TargetCents,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, echo.Map{"id": id})
	})

	e.PATCH("/pots/:id", func(c echo.Context) error {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid pot id"})
		}

		var req UpdatePotRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
		}

		if req.AddCents != nil {
			_, err := db.Exec(c.Request().Context(),
				`UPDATE saving_pots SET saved_cents = saved_cents + $1 WHERE id = $2`,
				*req.AddCents, id,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
		}

		return c.JSON(http.StatusOK, echo.Map{"status": "updated"})
	})

	e.GET("/pots", func(c echo.Context) error {
		userID := uuid.New() // TODO: Replace with actual user id

		rows, err := db.Query(c.Request().Context(),
			`SELECT id, user_id, name, target_cents, saved_cents, created_at, updated_at
			 FROM saving_pots WHERE user_id = $1 ORDER BY created_at DESC`,
			userID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		defer rows.Close()

		var pots []SavingPot
		for rows.Next() {
			var pot SavingPot
			if err := rows.Scan(&pot.ID, &pot.UserID, &pot.Name, &pot.TargetCents, &pot.SavedCents, &pot.CreatedAt, &pot.UpdatedAt); err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
			pots = append(pots, pot)
		}

		return c.JSON(http.StatusOK, pots)
	})
}
