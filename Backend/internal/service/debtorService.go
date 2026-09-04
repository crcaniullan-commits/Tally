package service

type DebtorService struct {
	db string
}

func NewDebtorService(db string) *DebtorService {
	return &DebtorService{db: db}
}
