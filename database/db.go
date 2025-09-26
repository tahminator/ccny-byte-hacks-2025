package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tahminator/go-react-template/config"
)

var pool *pgxpool.Pool

func Connect() error {
	dbConfig := config.Database{
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
		DbName:     os.Getenv("DB_NAME"),
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
	}
	dbUrl := dbConfig.Url()

	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return err
	}

	pool = dbPool
	return nil
}

func GetPool() (*pgxpool.Pool, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool not initialized. Call database.Connect() first")
	}
	return pool, nil
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}
