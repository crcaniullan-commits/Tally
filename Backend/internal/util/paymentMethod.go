package util

import "github.com/crcaniullan-commits/Tally/internal/model"

// PaymentMethod vive en model; se re-exporta acá para no cambiar los call sites.
type PaymentMethod = model.PaymentMethod

const (
	PaymentMethodDebito        = model.PaymentMethodDebito
	PaymentMethodCredito       = model.PaymentMethodCredito
	PaymentMethodTransferencia = model.PaymentMethodTransferencia
	PaymentMethodEfectivo      = model.PaymentMethodEfectivo
)
