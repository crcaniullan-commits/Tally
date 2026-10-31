package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizarNombre(t *testing.T) {
	casos := []struct {
		nombre   string
		esperado string
	}{
		{"Comida", "comida"},
		{"comida", "comida"},
		{"COMIDA", "comida"},
		{"  Comida  ", "comida"},
		{"Áqui", "aqui"},
		{"áQui", "aqui"},
		{"Café", "cafe"},
		{"Niño", "nino"},
		{"Über", "uber"},
		{"", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			assert.Equal(t, caso.esperado, NormalizarNombre(caso.nombre))
		})
	}
}
