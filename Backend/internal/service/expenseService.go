package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type ExpenseService struct {
	store store.StoreExpense
}

func NewExpenseService(store store.StoreExpense) *ExpenseService {
	return &ExpenseService{store: store}
}