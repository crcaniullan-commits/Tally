package util

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	rutRegex = regexp.MustCompile(`^(?i)^([0-9]{1,2})\.?([0-9]{3})\.?([0-9]{3})\-?([0-9kK])$`)

	ErrFormatoInvalido = errors.New("el formato del RUT no es válido")
	ErrDVInvalido      = errors.New("el dígito verificador del RUT no es matemáticamente correcto")
)

// RUT representa el Value Object con sus dos componentes para la base de datos
type RUT struct {
	Cuerpo int    // Ej: 19234567
	DV     string // Ej: "K" o "4"
}

// Parse convierte un string a un struct RUT validando SOLO el formato básico.
// Ideal para parámetros de ruta en búsquedas (ej: /usuarios/19234567K o /usuarios/19.234.567-K)
func ParseRUT(rutRaw string) (RUT, error) {
	if !rutRegex.MatchString(rutRaw) {
		return RUT{}, ErrFormatoInvalido
	}

	// Limpiar caracteres
	limpio := strings.ReplaceAll(rutRaw, ".", "")
	limpio = strings.ReplaceAll(limpio, "-", "")
	limpio = strings.ToUpper(limpio)

	// Separar cuerpo y dígito verificador
	dv := string(limpio[len(limpio)-1])
	cuerpoStr := limpio[:len(limpio)-1]
	cuerpo, _ := strconv.Atoi(cuerpoStr) // El regex ya asegura que son solo números

	return RUT{Cuerpo: cuerpo, DV: dv}, nil
}

// Validar verifica matemáticamente si el RUT es real usando el Módulo 11.
// Ideal para el caso de CREAR un usuario.
func (r RUT) Validar() error {
	sum := 0
	mul := 2

	// Algoritmo Módulo 11 recorriendo el cuerpo de derecha a izquierda
	cuerpoStr := strconv.Itoa(r.Cuerpo)
	for i := len(cuerpoStr) - 1; i >= 0; i-- {
		num, _ := strconv.Atoi(string(cuerpoStr[i]))
		sum += num * mul
		mul++
		if mul > 7 {
			mul = 2
		}
	}

	dvEsperado := 11 - (sum % 11)
	var dvString string
	if dvEsperado == 11 {
		dvString = "0"
	} else if dvEsperado != 10 {
		dvString = strconv.Itoa(dvEsperado)
	} else {
		dvString = "K"
	}

	if r.DV != dvString {
		return ErrDVInvalido
	}
	return nil
}

// String devuelve el RUT formateado de manera estándar (ej: 19.234.567-K)
func (r RUT) String() string {
	cuerpoStr := strconv.Itoa(r.Cuerpo)

	// Agregar puntos de manera dinámica según el largo
	var resultado string
	l := len(cuerpoStr)
	if l > 6 {
		resultado = cuerpoStr[:l-6] + "." + cuerpoStr[l-6:l-3] + "." + cuerpoStr[l-3:]
	} else if l > 3 {
		resultado = cuerpoStr[:l-3] + "." + cuerpoStr[l-3:]
	} else {
		resultado = cuerpoStr
	}
	return resultado + "-" + r.DV
}
