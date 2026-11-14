package auth

import "errors"

var (
	// ErrorCodigoRecuperacionInvalido: el código no corresponde al emitido.
	ErrorCodigoRecuperacionInvalido = errors.New("el código de recuperación no es válido")
	// ErrorCodigoRecuperacionVencido: el código pasó su ventana de validez.
	ErrorCodigoRecuperacionVencido = errors.New("el código de recuperación venció")
	// ErrorCodigoRecuperacionUsado: el código ya se consumió.
	ErrorCodigoRecuperacionUsado = errors.New("el código de recuperación ya fue usado")
	// ErrorCodigoRecuperacionNoVerificado: hay que verificar el código antes
	// de confirmar el cambio de contraseña.
	ErrorCodigoRecuperacionNoVerificado = errors.New("el código de recuperación todavía no fue verificado")
)
