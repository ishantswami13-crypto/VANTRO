package products

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	SKU        *string        `json:"sku,omitempty"`
	PriceCents int            `json:"price_cents"`
	Currency   string         `json:"currency"`
	Stock      int            `json:"stock"`
	Active     bool           `json:"active"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type CreateInput struct {
	Name       string         `json:"name"`
	SKU        *string        `json:"sku,omitempty"`
	PriceCents int            `json:"price_cents"`
	Currency   *string        `json:"currency,omitempty"` // default INR
	Stock      *int           `json:"stock,omitempty"`    // default 0
	Active     *bool          `json:"active,omitempty"`   // default true
	Metadata   map[string]any `json:"metadata,omitempty"` // default {}
}

type UpdateInput struct {
	Name       *string         `json:"name,omitempty"`
	SKU        *string         `json:"sku,omitempty"`
	PriceCents *int            `json:"price_cents,omitempty"`
	Currency   *string         `json:"currency,omitempty"`
	Stock      *int            `json:"stock,omitempty"`
	Active     *bool           `json:"active,omitempty"`
	Metadata   *map[string]any `json:"metadata,omitempty"`
}
