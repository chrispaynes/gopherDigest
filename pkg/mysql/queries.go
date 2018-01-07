package mysql

import (
	"database/sql"
	"fmt"
	"gopherDigest/pkg/rethinkdb"
)

//CommandQueryVerifier ...
type CommandQueryVerifier interface {
	Command(db *sql.DB, c string)
	Query(db sql.DB, q CommandQueryVerifier) string
	Verify(db *sql.DB, statement, assertion string) bool
}

// Explain runs a MySQL explain statement on a given query
func Explain(db *sql.DB, s string) (*sql.Rows, error) {
	rows, err := db.Query("EXPLAIN " + s)

	return rows, err
}

// QueryScanRow queries and scans a single row into a destination
func QueryScanRow(db *sql.DB, q string) string {
	var result string
	db.QueryRow(q).Scan(&result)

	return result
}

// QueryScanRows queries and scans multiple rows into a destination
func QueryScanRows(db *sql.DB, s string) (*sql.Rows, error) {
	rows, err := db.Query(s)

	return rows, err
}

// ExplainScan runs a MySQL explain statement on a given query and scans
// the results onto a destination
func ExplainScan(db *sql.DB, q string) string {
	var result string
	db.QueryRow(q).Scan(&result)

	return result
}

// VerifyScan queries the database to determine if an assertion about a
// single value in the database is true
func VerifyScan(db *sql.DB, statement, assertion string) (bool, error) {
	var result string
	db.QueryRow(statement).Scan(&result)

	return result == assertion, nil
}

// ScanRows scans a collections of rows onto a given destination
func ScanRows(r *sql.Rows, se rethinkdb.SQLExplain) error {
	defer r.Close()

	for r.Next() {
		err := r.Scan(&se.ID, &se.SelectType, &se.Table,
			&se.Partitions, &se.Ztype, &se.PossibleKeys, &se.Key,
			&se.KeyLen, &se.Ref, &se.Rows, &se.Filtered, &se.Extra)
		if err != nil {
			fmt.Printf("ERROR %s", err)
		}
	}

	return nil
}
