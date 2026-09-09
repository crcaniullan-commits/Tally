package incomes

type IncomeService struct {
	store StoreIncome
}

func NewIncomeService(store StoreIncome) *IncomeService {
	return &IncomeService{store: store}
}
