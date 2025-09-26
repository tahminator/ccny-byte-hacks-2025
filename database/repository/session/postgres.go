package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSessionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSessionRepository(db *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		db: db,
	}
}

func (repo *PostgresSessionRepository) CreateSession(ctx context.Context, session *Session) (*Session, error) {
	query := `
		INSERT INTO "Session"
			("userId", "createdAt", "expiresAt")
		VALUES
			(@userId, @createdAt, @expiresAt)
		RETURNING
			*
	`

	rows, err := repo.db.Query(ctx, query, pgx.NamedArgs{
		"userId":    session.UserId,
		"createdAt": session.CreatedAt,
		"expiresAt": session.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &s, nil
}

func (repo *PostgresSessionRepository) GetSessionById(ctx context.Context, id uuid.UUID) (*Session, error) {
	query := `
		SELECT
			*
		FROM
			"Session"
		WHERE
			id = @id
	`

	rows, err := repo.db.Query(ctx, query, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (repo *PostgresSessionRepository) GetSessionsByUserId(ctx context.Context) (*[]Session, error) {
	query := `
		SELECT
			*
		FROM
			"Session"
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	sessions, err := pgx.CollectRows(rows, pgx.RowToStructByName[Session])
	if err != nil {
		return nil, fmt.Errorf("failed to collect sessions: %w", err)
	}

	return &sessions, nil
}

func (repo *PostgresSessionRepository) UpdateSession(ctx context.Context, session *Session) (*Session, error) {
	query := `
		UPDATE
			"Session"
		SET
			"expiresAt" = @expiresAt
		WHERE
			id = @id
	`

	_, err := repo.db.Exec(ctx, query, pgx.NamedArgs{
		"id":        session.Id,
		"expiresAt": session.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

func (repo *PostgresSessionRepository) DeleteSession(ctx context.Context, session *Session) (*Session, error) {
	query := `
		DELETE FROM
			"Session"
		WHERE
			id = @id
	`

	_, err := repo.db.Exec(ctx, query, pgx.NamedArgs{
		"id": session.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return session, nil
}

// this doesn't do anything useful. it's purpose is to type check the repository against
// the interface
var _ SessionRepository = new(PostgresSessionRepository)
