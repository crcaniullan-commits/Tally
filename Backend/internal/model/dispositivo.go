package model

import (
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Token      string    `json:"token"`
	Platform   string    `json:"platform"`
	LastSeenAt time.Time `json:"last_seen_at"`
}
