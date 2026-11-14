package entorno

import (
	"os"
	"strconv"
)

func ObtenerCadena(key, fallback string) string {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return val
}

func ObtenerEntero(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valsInt
}

func ObtenerBooleano(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)

	if err != nil {
		return fallback
	}

	return boolVal
}
