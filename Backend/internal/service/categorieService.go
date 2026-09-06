package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type CategorieService struct {
	store store.StoreCategorie
}

func NewCategorieService(store store.StoreCategorie) *CategorieService {
	return &CategorieService{store: store}
}