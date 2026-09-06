package store

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resources already exists")
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	StorageUsers      *StoreUser
	StorageIncomes    *StoreIncome
	StorageExpenses   *StoreExpense
	StorageDebtors    *StoreDebtor
	StorageCategories *StoreCategorie
	StorageGoals      *StoreGoal
	StorageAccessKey  *StoreAccessKey
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		StorageUsers:      &StoreUser{db},
		StorageIncomes:    &StoreIncome{db},
		StorageExpenses:   &StoreExpense{db},
		StorageDebtors:    &StoreDebtor{db},
		StorageCategories: &StoreCategorie{db},
		StorageGoals:      &StoreGoal{db},
		StorageAccessKey:  &StoreAccessKey{db},
	}
}
