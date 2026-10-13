package empresas

import (
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
)

type ServiceEmpresas interface {
}

type EmpresasHandler struct {
	service ServiceEmpresas
	errors  errorhandler.ErrorsResponse
}

func NewEmpresasHandler(s ServiceEmpresas, e errorhandler.ErrorsResponse) *EmpresasHandler {
	return &EmpresasHandler{s, e}
}
