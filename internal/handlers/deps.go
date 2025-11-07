package handlers

import "github.com/ishantswami13-crypto/vantro/internal/services"

type Deps struct {
	Expenses *services.ExpensesService
	Pots     *services.PotsService
	Coach    *services.CoachService
}
