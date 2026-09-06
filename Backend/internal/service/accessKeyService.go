package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type AccessKeyService struct {
	store store.StoreAccessKey
}

func NewAccessKeyService(store store.StoreAccessKey) *AccessKeyService {
	return &AccessKeyService{store: store}
}