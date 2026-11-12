package utilidades

import "github.com/crcaniullan-commits/Tally/internal/model"

// RolUsuarioTipo vive en model; se re-exporta acá para no cambiar los call sites.
type RolUsuarioTipo = model.UserRole

const (
	RolUsuario   = model.UserRoleUsuario
	RolMunicipal = model.UserRoleMunicipal
)
