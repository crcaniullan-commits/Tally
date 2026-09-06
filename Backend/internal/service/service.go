package service

import "github.com/crcaniullan-commits/Tally/internal/store"

type Service struct {
	ServiceUsers      *UserService
	ServiceIncomes    *IncomeService
	ServiceExpenses   *ExpenseService
	ServiceDebtors    *DebtorService
	ServiceCategories *CategorieService
	ServiceGoals      *GoalService
	ServiceAccessKey  *AccessKeyService
}

func NewService(store store.Storage) Service {
	return Service{
		ServiceUsers:      &UserService{*store.StorageUsers},
		ServiceIncomes:    &IncomeService{*store.StorageIncomes},
		ServiceExpenses:   &ExpenseService{*store.StorageExpenses},
		ServiceDebtors:    &DebtorService{*store.StorageDebtors},
		ServiceCategories: &CategorieService{*store.StorageCategories},
		ServiceGoals:      &GoalService{*store.StorageGoals},
		ServiceAccessKey:  &AccessKeyService{*store.StorageAccessKey},
	}
}
