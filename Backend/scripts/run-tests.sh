#!/bin/sh
# Corre los tests del backend de forma tolerante a la falta de cgo.
#
# "go test -race" exige cgo (CGO_ENABLED=1). Hay entornos donde cgo no está
# disponible:
#
#   - Windows: Go viene con CGO_ENABLED=0 salvo que se instale un compilador C
#     (mingw/MSYS2). Sin él, el comando muere con:
#         go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
#   - Alpine (imagen golang:alpine del stage de tests): no trae gcc.
#
# Este script no falla en esos casos: si no se puede correr el race detector,
# lo saca de los flags y corre los tests sin él, avisando por stderr. Los tests
# siguen siendo válidos; lo que se pierde es la detección de carreras.
#
# Uso: sh scripts/run-tests.sh [flags de go test] [paquetes]
# Ej: sh scripts/run-tests.sh -race -cover ./...

set -e

# HayRaceDetector decide si el race detector se puede usar en este entorno.
# CGO_ENABLED=1 no basta: hace falta además un compilador C, porque el linker
# de cgo lo invoca al construir el runtime de race.
HayRaceDetector() {
	if [ "$(go env CGO_ENABLED)" != "1" ]; then
		return 1
	fi

	if command -v gcc >/dev/null 2>&1 || command -v clang >/dev/null 2>&1; then
		return 0
	fi

	return 1
}

if HayRaceDetector; then
	exec go test "$@"
fi

echo "aviso: este entorno no tiene cgo (CGO_ENABLED=$(go env CGO_ENABLED), sin compilador C);" >&2
echo "aviso: se corren los tests SIN el race detector." >&2

# Se sacan -race/-race=true de los argumentos. El filtrado se hace en el
# scope del script (no en una función) para que el exec de abajo vea la lista
# filtrada: dentro de una función, set -- solo cambia los parámetros de esa
# función. La lista se reconstruye con shift + append para no perder el
# quoting de los paths con espacios.
n=$#
i=0

while [ "$i" -lt "$n" ]; do
	arg="$1"
	shift

	case "$arg" in
	-race | -race=true)
		# descartado
		;;
	*)
		set -- "$@" "$arg"
		;;
	esac

	i=$((i + 1))
done

exec go test "$@"
