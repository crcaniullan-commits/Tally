package accesskeys

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type AccessKeyStorage struct {
	ID         uuid.UUID
	Code       string
	CreatedBy  uuid.UUID
	RedeemedBy *uuid.UUID
	RedeemedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

type StoreAccessKey struct {
	db *sql.DB
}