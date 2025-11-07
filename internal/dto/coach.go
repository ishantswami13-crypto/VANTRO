package dto

type CoachInput struct {
	IncomeCents int    `json:"income_cents"`
	RentCents   int    `json:"rent_cents"`
	Goal        string `json:"goal"`
}
