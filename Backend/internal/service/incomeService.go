package service

type IncomeService struct {
	db string
}

func NewIncomeService(db string) *IncomeService {
	return &IncomeService{db: db}
}
