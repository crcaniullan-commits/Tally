package model

import (
	"time"

	"github.com/google/uuid"
)

type AccessKey struct {
	ID         uuid.UUID  `json:"id"`
	Code       string     `json:"code"`
	CreatedBy  uuid.UUID  `json:"created_by"`
	RedeemedBy *uuid.UUID `json:"redeemed_by"`
	RedeemedAt *time.Time `json:"redeemed_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	RevokedBy  *uuid.UUID `json:"revoked_by"`
	CreatedAt  time.Time  `json:"created_at"`
}
