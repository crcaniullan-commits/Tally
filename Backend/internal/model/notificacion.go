package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationLog struct {
	UserID  uuid.UUID `json:"user_id"`
	Kind    string    `json:"kind"`
	RefDate time.Time `json:"ref_date"`
	SentAt  time.Time `json:"sent_at"`
}

type Notification struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
}
