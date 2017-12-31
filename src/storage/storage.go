package storage

import (
	"database/sql"
	"fmt"
)

//CommandQueryVerifier does
type CommandQueryVerifier interface {
	Command(db *sql.DB, c string)
	Query(db sql.DB, q CommandQueryVerifier) string
	Verify(db *sql.DB, statement, assertion string) bool
}

// Command does
func Command(db *sql.DB, command string) {
	db.Exec(command)
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

// VerifyExecs does
func VerifyExecs(db *sql.DB, cqv CommandQueryVerifier) bool {
	Command(db, "select ...")
	Query(db, "select")

	return Verify(db, "select", "assert")
}

type command struct {
	// insert, update, set, drop, create
}

type query struct {
	// select
}

// PrintExec prints SQL statements to standard output
func PrintExec(db *sql.DB, stmnts []string) ([]string, error) {
	results := []string{}

	// execute statements and return result rows to array
	for _, s := range stmnts {
		fmt.Printf("mysql> %s;\n", s)
		rows, err := db.Query(s)

		if err != nil {
			return []string{}, fmt.Errorf("could not execute the SQL statement %s", err)
		}

		var row string

		for rows.Next() {
			rows.Scan(&row)
		}

		results = append(results, row)
	}

	return results, nil
}
