package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizarNombre deja un nombre comparable: sin espacios en los extremos, en
// minúsculas y sin acentos. "Comida", "comida" y "COMIDA" normalizan igual, y
// también "Áqui", "aqui" y "áqui".
//
// Es la misma regla que usa CategorieService para validar nombres duplicados.
func NormalizarNombre(nombre string) string {
	descompuesto := norm.NFD.String(strings.TrimSpace(nombre))

	var normalizado strings.Builder
	normalizado.Grow(len(descompuesto))

	for _, r := range descompuesto {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		normalizado.WriteRune(r)
	}

	return strings.ToLower(normalizado.String())
}
