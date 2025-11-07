package dto

type PotCreate struct {
	Name        string `json:"name"`
	TargetCents int    `json:"target_cents"`
}

type PotUpdate struct {
	AddCents *int `json:"add_cents"`
}
