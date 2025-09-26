package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `db:"id" json:"id"`
	GoogleId  string    `db:"googleId" json:"googleId"`
	IsAdmin   bool      `db:"isAdmin" json:"isAdmin"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}
