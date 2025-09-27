package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (repo *PostgresUserRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	query := `
	INSERT INTO "User"
		("googleId", "isAdmin")
	VALUES
		(@googleId, @isAdmin)
	RETURNING
		*
	`

	rows, err := repo.db.Query(ctx, query, pgx.NamedArgs{
		"googleId": user.GoogleId,
		"isAdmin":  user.IsAdmin,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &u, nil
}

func (repo *PostgresUserRepository) GetUserById(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT
			*
		FROM
			"User"
		WHERE
			id = @id
	`

	rows, err := repo.db.Query(ctx, query, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (repo *PostgresUserRepository) GetUserByGoogleId(ctx context.Context, googleId string) (*User, error) {
	query := `
		SELECT
			*
		FROM
			"User"
		WHERE
			"googleId" = @googleId
	`

	rows, err := repo.db.Query(ctx, query, pgx.NamedArgs{
		"googleId": googleId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (repo *PostgresUserRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {
	query := `
		UPDATE 
			"User"
		SET
			"googleId" = @googleId,
			"isAdmin" = @isAdmin,
			"githubToken" = @githubToken,
			"githubUsername" = @githubUsername
		WHERE
			id = @id
	`

	ct, err := repo.db.Exec(ctx, query, pgx.NamedArgs{
		"id":             user.Id,
		"googleId":       user.GoogleId,
		"isAdmin":        user.IsAdmin,
		"githubToken":    user.GithubToken,
		"githubUsername": user.GithubUsername,
	})
	if err != nil {
		return user, fmt.Errorf("failed to update user: %w", err)
	}

	if ct.RowsAffected() != 1 {
		return user, fmt.Errorf("expected to update 1 user, updated %d", ct.RowsAffected())
	}

	return user, nil
}

func (repo *PostgresUserRepository) DeleteUserById(ctx context.Context, user *User) (*User, error) {
	query := `
		DELETE FROM
			"User"
		WHERE
			id = @id
	`

	ct, err := repo.db.Exec(ctx, query, pgx.NamedArgs{
		"id": user.Id,
	})
	if err != nil {
		return user, fmt.Errorf("failed to delete user by id: %w", err)
	}

	if ct.RowsAffected() != 1 {
		return user, fmt.Errorf("expected to update 1 user, updated %d", ct.RowsAffected())
	}

	return user, nil
}

// this doesn't do anything useful. it's purpose is to type check the repository against
// the interface
var _ UserRepository = new(PostgresUserRepository)
