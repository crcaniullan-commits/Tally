package users

import (
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type Users struct {
	ID           uuid.UUID     `json:"id"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-"`
	Nombre       string        `json:"nombre"`
	Role         util.UserRole `json:"role"`
	Rut          util.RUT      `json:"rut"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
