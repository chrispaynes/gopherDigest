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

// ExplainScanRow runs a MySQL explain statement on a given query and scans
// the results onto a destination
func ExplainScanRow(db *sql.DB, q string) string {
	var result string
	db.QueryRow(q).Scan(&result)

	return result
}

// ExplainScanRows runs a MySQL explain statement on a given query and scans
// the results onto a destination
func ExplainScanRows(db *sql.DB, query string) (*rethinkdb.SQLExplainRow, error) {
	se := &rethinkdb.SQLExplainRow{}

	rows, err := Explain(db, query)
	// defer rows.Close()

	if err != nil {
		return se, fmt.Errorf("%s", err)
	}

	if err := ScanRows(rows, se); err != nil {
		return se, fmt.Errorf("%s", err)
	}

	return se, nil

}

// VerifyScan queries the database to determine if an assertion about a
// single value in the database is true
func VerifyScan(db *sql.DB, statement, assertion string) (bool, error) {
	var result string
	db.QueryRow(statement).Scan(&result)

	return result == assertion, nil
}

// ScanRows scans a collections of rows onto a given destination
func ScanRows(r *sql.Rows, se *rethinkdb.SQLExplainRow) error {
	for r.Next() {
		err := r.Scan(&se.ID, &se.SelectType, &se.Table,
			&se.Partitions, &se.Ztype, &se.PossibleKeys, &se.Key,
			&se.KeyLen, &se.Ref, &se.Rows, &se.Filtered, &se.Extra)
		if err != nil {
			return fmt.Errorf("failed to copy the row columns to the destination \n%s", err)
		}
	}

	return nil
}
