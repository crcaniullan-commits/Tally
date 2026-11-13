package testutil

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
)

// NewRUT devuelve un RUT válido (cuerpo + DV módulo 11) sin puntos y con guion.
// El largo queda en 9 caracteres, dentro del max=10 del registro de usuario y
// del min=9/max=12 del registro municipal.
func NewRUT() string {
	var b [4]byte
	_, _ = rand.Read(b[:])

	body := 1_000_000 + int(binary.BigEndian.Uint32(b[:])%9_000_000)

	return fmt.Sprintf("%d-%s", body, digitoVerificador(body))
}

func digitoVerificador(body int) string {
	suma := 0
	mul := 2

	cuerpo := strconv.Itoa(body)
	for i := len(cuerpo) - 1; i >= 0; i-- {
		digito, _ := strconv.Atoi(string(cuerpo[i]))
		suma += digito * mul
		mul++
		if mul > 7 {
			mul = 2
		}
	}

	dv := 11 - (suma % 11)

	switch dv {
	case 11:
		return "0"
	case 10:
		return "K"
	default:
		return strconv.Itoa(dv)
	}
}
