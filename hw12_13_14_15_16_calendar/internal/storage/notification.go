package storage

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Start  time.Time `json:"start"`
	UserID uuid.UUID `json:"userId"`
}
