package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type DebtorService struct {
	store store.StoreDebtor
}

func NewDebtorService(store store.StoreDebtor) *DebtorService {
	return &DebtorService{store: store}
}