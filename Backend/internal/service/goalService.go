package service

import (
	"github.com/crcaniullan-commits/Tally/internal/store"
)

type GoalService struct {
	store store.StoreGoal
}

func NewGoalService(store store.StoreGoal) *GoalService {
	return &GoalService{store: store}
}