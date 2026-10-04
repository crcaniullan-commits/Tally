package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  password  `json:"-"`
	Nombre        string    `json:"nombre"`
	Role          UserRole  `json:"role"`
	Rut           RUT       `json:"rut"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PlanExpiresAt time.Time `json:"plan_expires_at"`
}

type password struct {
	text *string
	Hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.Hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	if err := bcrypt.CompareHashAndPassword(p.Hash, []byte(text)); err != nil {
		return err
	}

	return nil
}

func (p *password) GetHash() []byte {
	return p.Hash
}
