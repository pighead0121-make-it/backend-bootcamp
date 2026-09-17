package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pighead0121-make-it/backend-bootcamp/projects/book-api/config"
)

func Connect(configDB config.Config) (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		configDB.DBUser,
		configDB.DBPassword,
		configDB.DBHost,
		configDB.DBPort,
		configDB.DBName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
