package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// StartPostgresqlConnection initiates a connection to the Cockroach database service
// and returns a reference to the sql.DB database object.
// Return (*sql.DB): Reference to the sql.DB database object
// Return (error): Error
func StartPostgresqlConnection(userName string, host string, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s@%s:26257/%s?sslmode=disable", userName, host, dataSourceName))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
