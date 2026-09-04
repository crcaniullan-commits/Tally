package service

type Service struct {
	ServiceUsers          *UserService
	ServiceIncomes        *IncomeService
	ServiceExpenses       *ExpenseService
	ServiceDebtors        *DebtorService
	ServiceCategories     *CategorieService
	ServiceGoals          *GoalService
	ServicePaymentMethods *PaymentMethodService
	ServiceAccessKey      *AccessKeyService
}

func newService(db string) Service {
	return Service{
		ServiceUsers:          &UserService{db},
		ServiceIncomes:        &IncomeService{db},
		ServiceExpenses:       &ExpenseService{db},
		ServiceDebtors:        &DebtorService{db},
		ServiceCategories:     &CategorieService{db},
		ServiceGoals:          &GoalService{db},
		ServicePaymentMethods: &PaymentMethodService{db},
		ServiceAccessKey:      &AccessKeyService{db},
	}
}
