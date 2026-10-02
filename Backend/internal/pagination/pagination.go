package pagination

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
)

const (
	// DefaultLimit es el tamaño de página que se usa si el cliente no manda
	// ?limit. Also es el tope: el tag lte de Limit tiene que coincidir con este
	// valor.
	DefaultLimit = 30
	// MaxLimit es el límite duro de ?limit. Un page_size más grande se
	// responde con 400 en vez de recortar en silencio.
	MaxLimit = 30
)

// IncomePaginationQuery son los query params de GET /app/incomes.
//
// Since y Until van como "AAAA-MM-DD" porque incomes.fecha es un DATE (ver
// migrations/000001). Si vienen vacíos el store los trata como "sin filtro".
//
// Ojo con los tags de validate: sin espacios después de la coma. Un tag mal
// escrito no da 400, hace panear el Validate.Struct y el Recoverer de chi
// convierte el endpoint entero en un 500 sin cuerpo.
type IncomePaginationQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=30"`
	Offset int    `json:"offset" validate:"gte=0"`
	Since  string `json:"since"`
	Until  string `json:"until"`
}

// NewIncomePaginationQuery devuelve la paginación por defecto (primera página
// del tamaño tope) para partirla de ahí en el Parse.
func NewIncomePaginationQuery() IncomePaginationQuery {
	return IncomePaginationQuery{Limit: DefaultLimit, Offset: 0}
}

// Parse arma la paginación desde la query string. Ante un parámetro inválido
// devuelve error en vez de ignorarlo en silencio: si ?limit=abc se comiera,
// el cliente creería que pidió una página y recibiría otra.
func (fq IncomePaginationQuery) Parse(r *http.Request) (IncomePaginationQuery, error) {
	qs := r.URL.Query()

	if v := qs.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return fq, fmt.Errorf("limit inválido: %q no es un entero", v)
		}
		fq.Limit = limit
	}

	if v := qs.Get("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil {
			return fq, fmt.Errorf("offset inválido: %q no es un entero", v)
		}
		fq.Offset = offset
	}

	if v := qs.Get("since"); v != "" {
		date, err := parseDate(v)
		if err != nil {
			return fq, fmt.Errorf("since inválido: %q no es una fecha AAAA-MM-DD", v)
		}
		fq.Since = date
	}

	if v := qs.Get("until"); v != "" {
		date, err := parseDate(v)
		if err != nil {
			return fq, fmt.Errorf("until inválido: %q no es una fecha AAAA-MM-DD", v)
		}
		fq.Until = date
	}

	if err := util.Validate.Struct(fq); err != nil {
		return fq, err
	}

	return fq, nil
}

// parseDate normaliza a "AAAA-MM-DD". incomes.fecha es DATE, así que no
// alcanza con time.DateTime ("2006-01-02 15:04:05"): con ese layout un
// ?since=2026-09-01 no parsea y el filtro de fecha se pierde en silencio.
func parseDate(s string) (string, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return "", err
	}

	return t.Format(time.DateOnly), nil
}
