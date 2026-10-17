package empresas

type StoreEmpresas interface {
}

type EmpresaService struct {
	store StoreEmpresas
}

func NewEmpresaService(store StoreEmpresas) *EmpresaService {
	return &EmpresaService{store}
}
