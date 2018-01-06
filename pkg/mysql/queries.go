package mysql

import (
	"database/sql"
)

//CommandQueryVerifier does
type CommandQueryVerifier interface {
	Command(db *sql.DB, c string)
	Query(db sql.DB, q CommandQueryVerifier) string
	Verify(db *sql.DB, statement, assertion string) bool
}

type query struct {
	// select
}

// Query does
func Query(db *sql.DB, q string) string {
	var result string
	db.QueryRow(q).Scan(&result)

	return result
}

// Verify does
func Verify(db *sql.DB, statement, assertion string) bool {
	var result string
	db.QueryRow(statement).Scan(&result)

	return result == assertion
}
