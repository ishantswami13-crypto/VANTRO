package dto

type ExpenseCreate struct {
	AmountCents int    `json:"amount_cents"`
	Category    string `json:"category"`
	Mood        string `json:"mood"`
	Note        string `json:"note"`
}
