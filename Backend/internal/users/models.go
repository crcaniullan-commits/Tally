package users

import (
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	ID           uuid.UUID     `json:"id"`
	Email        string        `json:"email"`
	PasswordHash password      `json:"-"`
	Nombre       string        `json:"nombre"`
	Role         util.UserRole `json:"role"`
	Rut          util.RUT      `json:"rut"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	if err := bcrypt.CompareHashAndPassword(p.hash, []byte(text)); err != nil {
		return err
	}

	return nil
}
