package api

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tahminator/go-react-template/database"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
)

type AppContext struct {
	Db                *pgxpool.Pool
	UserRepository    *user.PostgresUserRepository
	SessionRepository *session.PostgresSessionRepository
}

func (c *AppContext) databaseBuilder() {
	if c.Db == nil {
		err := database.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		db, err := database.GetPool()
		if err != nil {
			log.Fatalf("Failed to get database pool: %v", err)
		}
		c.Db = db
	}
}

func (c *AppContext) repositoryBuilder() {
	if c.UserRepository == nil {
		c.UserRepository = user.NewPostgresUserRepository(c.Db)
	}
	if c.SessionRepository == nil {
		c.SessionRepository = session.NewPostgresSessionRepository(c.Db)
	}
}
