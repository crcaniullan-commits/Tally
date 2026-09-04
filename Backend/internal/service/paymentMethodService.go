package service

type PaymentMethodService struct {
	db string
}

func NewPaymentMethodService(db string) *PaymentMethodService {
	return &PaymentMethodService{db: db}
}
