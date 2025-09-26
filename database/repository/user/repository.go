package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByGoogleId(ctx context.Context, googleId string) (*User, error)
	DeleteUserById(ctx context.Context, user *User) (*User, error)
}
