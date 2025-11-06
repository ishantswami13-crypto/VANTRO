package mock

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Provider struct{}

func New() *Provider { return &Provider{} }

// CreatePayout simulates an async payout request to a provider.
// Returns providerRef and initial "processing" status.
func (p *Provider) CreatePayout(ctx context.Context, amountCents int64, method string, dest map[string]string) (string, string, error) {
	ref := fmt.Sprintf("mock_%d", time.Now().UnixNano())
	return ref, "processing", nil
}

// Resolve picks a final status (92% success) and generates a fake UTR.
func (p *Provider) Resolve(ctx context.Context, providerRef string) (finalStatus, utr string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	if rand.Intn(100) < 92 {
		return "success", fmt.Sprintf("%d", 1000000000+rand.Intn(899999999))
	}
	return "failed", ""
}
