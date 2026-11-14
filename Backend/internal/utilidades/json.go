package utilidades

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validador *validator.Validate

func init() {
	Validador = validator.New(validator.WithRequiredStructEnabled())
}

func EscribirJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func LeerJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// LeerJSONYValidar lee el body JSON en dst y valida sus tags `validate`.
// Devuelve el error de decodificación (incluido un campo desconocido, por
// DisallowUnknownFields) o el de validación; el handler lo traduce a 400.
func LeerJSONYValidar(w http.ResponseWriter, r *http.Request, dst any) error {
	if err := LeerJSON(w, r, dst); err != nil {
		return err
	}

	return Validador.Struct(dst)
}

func EscribirErrorJSON(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}
	return EscribirJSON(w, status, &envelope{Error: message})
}

func RespuestaJSON(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Datos any `json:"datos"`
	}

	return EscribirJSON(w, status, &envelope{Datos: data})
}
