package session

import (
	"context"

	"github.com/google/uuid"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionById(ctx context.Context, id uuid.UUID) (*Session, error)
	GetSessionsByUserId(ctx context.Context) (*[]Session, error)
	DeleteSession(ctx context.Context, session *Session) (*Session, error)
}
