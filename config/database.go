package config

import "fmt"

type Database struct {
	DbHost     string
	DbPort     string
	DbName     string
	DbUser     string
	DbPassword string
}

// URL generator.
func (db *Database) Url() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		db.DbUser,
		db.DbPassword,
		db.DbHost,
		db.DbPort,
		db.DbName,
	)
}
