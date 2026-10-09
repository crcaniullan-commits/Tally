package util

import "github.com/crcaniullan-commits/Tally/internal/model"

// UserRole vive en model; se re-exporta acá para no cambiar los call sites.
type UserRole = model.UserRole

const (
	UserRoleUsuario   = model.UserRoleUsuario
	UserRoleMunicipal = model.UserRoleMunicipal
)
