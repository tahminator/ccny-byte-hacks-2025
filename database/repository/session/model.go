package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id        uuid.UUID `db:"id" json:"id"`
	UserId    uuid.UUID `db:"userId" json:"userId"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	ExpiresAt time.Time `db:"expiresAt" json:"expiresAt"`
}
