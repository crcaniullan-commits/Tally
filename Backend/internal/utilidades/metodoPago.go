package utilidades

import "github.com/crcaniullan-commits/Tally/internal/model"

// MetodoPago vive en model; se re-exporta acá para no cambiar los call sites.
type MetodoPago = model.PaymentMethod

const (
	MetodoPagoDebito        = model.PaymentMethodDebito
	MetodoPagoCredito       = model.PaymentMethodCredito
	MetodoPagoTransferencia = model.PaymentMethodTransferencia
	MetodoPagoEfectivo      = model.PaymentMethodEfectivo
)
