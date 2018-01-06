package mysql

import "database/sql"

type command struct {
	// insert, update, set, drop, create
}

// Command does
func Command(db *sql.DB, command string) {
	db.Exec(command)
}
