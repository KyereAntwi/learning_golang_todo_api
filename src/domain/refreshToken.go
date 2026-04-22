package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	UserId      uuid.UUID
	HashedToken string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}
