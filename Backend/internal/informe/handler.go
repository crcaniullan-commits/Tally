package informe

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/google/uuid"
)

// TODO: exponer los endpoints para consultar una vez toda la información de un
// usuario con sus empresas o de una sola empresa. Sin lógica por ahora.
type IServiceInforme interface {
	obtenerTodo(context.Context, uuid.UUID)
	obtenerTodoDeUnaEmpresa(context.Context, uuid.UUID, uuid.UUID)
}

type HandlerInforme struct {
	errors  errores.RespuestaErrores
	service IServiceInforme
}

func NuevoHandlerInforme(err errores.RespuestaErrores, svc IServiceInforme) *HandlerInforme {
	return &HandlerInforme{errors: err, service: svc}
}
