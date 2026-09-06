package util

type PaymentMethod string

const (
	PaymentMethodDebito         PaymentMethod = "debito"
	PaymentMethodCredito        PaymentMethod = "credito"
	PaymentMethodTransferencia  PaymentMethod = "transferencia"
	PaymentMethodEfectivo       PaymentMethod = "efectivo"
)