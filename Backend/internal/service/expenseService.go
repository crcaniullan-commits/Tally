package service

type ExpenseService struct {
	db string
}

func NewExpenseService(db string) *ExpenseService {
	return &ExpenseService{db: db}
}
