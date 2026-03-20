package business

import "time"

type Business struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id"`
	Name        string    `json:"name"`
	Industry    *string   `json:"industry,omitempty"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBusinessRequest struct {
	OwnerUserID string  `json:"owner_user_id"`
	Name        string  `json:"name"`
	Industry    *string `json:"industry,omitempty"`
	Currency    *string `json:"currency,omitempty"` // default INR
}

type CreateBusinessResponse struct {
	BusinessID string `json:"business_id"`
	AccountID  string `json:"default_account_id"`
	Message    string `json:"message"`
}
