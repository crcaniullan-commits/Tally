package paginacion

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

const (
	// LimitePorDefecto es el tamaño de página que se usa si el cliente no manda
	// ?limite. Además es el tope: el tag lte de Limite tiene que coincidir con
	// este valor.
	LimitePorDefecto = 30
	// LimiteMaximo es el límite duro de ?limite. Un page_size más grande se
	// responde con 400 en vez de recortar en silencio.
	LimiteMaximo = 30
)

// ConsultaPaginacionIngresos son los query params de GET /aplicacion/empresas/{empresaID}/ingresos.
//
// Desde y Hasta van como "AAAA-MM-DD" porque incomes.fecha es un DATE (ver
// migrations/000001). CategoryID es el UUID de incomes.category_id
// (migrations/000006), no el nombre: filtrar por nombre exigiría un JOIN con
// categories y no podría aprovechar el índice de la FK.
//
// Los tres filtros vacíos significan "sin filtro" y llegan al store como "".
//
// Ojo con los tags de validate: sin espacios después de la coma. Un tag mal
// escrito no da 400, hace panear el Validador.Struct y el Recoverer de chi
// convierte el endpoint entero en un 500 sin cuerpo.
type ConsultaPaginacionIngresos struct {
	Limite         int    `json:"limite" validate:"gte=1,lte=30"`
	Desplazamiento int    `json:"desplazamiento" validate:"gte=0"`
	Desde          string `json:"desde"`
	Hasta          string `json:"hasta"`
	CategoryID     string `json:"categoria_id"`
}

// NuevaConsultaPaginacionIngresos devuelve la paginación por defecto (primera página
// del tamaño tope) para partirla de ahí en el Parse.
func NuevaConsultaPaginacionIngresos() ConsultaPaginacionIngresos {
	return ConsultaPaginacionIngresos{Limite: LimitePorDefecto, Desplazamiento: 0}
}

// Parse arma la paginación desde la query string. Ante un parámetro inválido
// devuelve error en vez de ignorarlo en silencio: si ?limite=abc se comiera,
// el cliente creería que pidió una página y recibiría otra.
func (fq ConsultaPaginacionIngresos) Parse(r *http.Request) (ConsultaPaginacionIngresos, error) {
	qs := r.URL.Query()

	if v := qs.Get("limite"); v != "" {
		limite, err := strconv.Atoi(v)
		if err != nil {
			return fq, fmt.Errorf("limite inválido: %q no es un entero", v)
		}
		fq.Limite = limite
	}

	if v := qs.Get("desplazamiento"); v != "" {
		desplazamiento, err := strconv.Atoi(v)
		if err != nil {
			return fq, fmt.Errorf("desplazamiento inválido: %q no es un entero", v)
		}
		fq.Desplazamiento = desplazamiento
	}

	if v := qs.Get("desde"); v != "" {
		date, err := parseDate(v)
		if err != nil {
			return fq, fmt.Errorf("desde inválido: %q no es una fecha AAAA-MM-DD", v)
		}
		fq.Desde = date
	}

	if v := qs.Get("hasta"); v != "" {
		date, err := parseDate(v)
		if err != nil {
			return fq, fmt.Errorf("hasta inválido: %q no es una fecha AAAA-MM-DD", v)
		}
		fq.Hasta = date
	}

	if v := qs.Get("categoria_id"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return fq, fmt.Errorf("categoria_id inválido: %q no es un UUID", v)
		}
		fq.CategoryID = v
	}

	if err := utilidades.Validador.Struct(fq); err != nil {
		return fq, err
	}

	return fq, nil
}

// parseDate normaliza a "AAAA-MM-DD". incomes.fecha es DATE, así que no
// alcanza con time.DateTime ("2006-01-02 15:04:05"): con ese layout un
// ?desde=2026-09-01 no parsea y el filtro de fecha se pierde en silencio.
func parseDate(s string) (string, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return "", err
	}

	return t.Format(time.DateOnly), nil
}
