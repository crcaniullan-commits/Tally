package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type IncomeService struct {
	store store.StoreIncome
}

func NewIncomeService(store store.StoreIncome) *IncomeService {
	return &IncomeService{store: store}
}
