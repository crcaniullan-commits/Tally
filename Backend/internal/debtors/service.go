package debtors

type DebtorService struct {
	store StoreDebtor
}

func NewDebtorService(store StoreDebtor) *DebtorService {
	return &DebtorService{store: store}
}