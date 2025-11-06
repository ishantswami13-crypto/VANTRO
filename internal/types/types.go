package types

type CreatePayoutRequest struct {
	Amount    float64            `json:"amount"`   // INR, rupees
	Currency  string             `json:"currency"` // "INR"
	Method    string             `json:"method"`   // "upi" | "bank"
	UPI       *UPIDest           `json:"upi,omitempty"`
	Bank      *BankDest          `json:"bank,omitempty"`
	Reference string             `json:"reference_id,omitempty"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

type UPIDest struct {
	VPA  string `json:"vpa"`
	Name string `json:"name,omitempty"`
}

type BankDest struct {
	Account string `json:"account"`
	IFSC    string `json:"ifsc"`
	Name    string `json:"name,omitempty"`
}

type CreatePayoutResponse struct {
	PayoutID           string `json:"payout_id"`
	Status             string `json:"status"`
	Reference          string `json:"reference_id,omitempty"`
	ExpectedSettlement string `json:"expected_settlement,omitempty"`
}

type Payout struct {
	ID          string  `json:"payout_id"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"`
	UTR         string  `json:"utr,omitempty"`
	ProcessedAt string  `json:"processed_at,omitempty"`
}
