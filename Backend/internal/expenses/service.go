package expenses

type ExpenseService struct {
	store StoreExpense
}

func NewExpenseService(store StoreExpense) *ExpenseService {
	return &ExpenseService{store: store}
}
