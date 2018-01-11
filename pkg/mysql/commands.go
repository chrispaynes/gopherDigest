package mysql

import (
	"database/sql"
	"fmt"
	"gopherDigest/pkg/rethinkdb"
	"log"
)

type command struct {
	// insert, update, set, drop, create
}

// FetchEventSummary fetches SQL SELECT statements from the events_statements_summary_by_digest table
func FetchEventSummary(db *sql.DB, query string, ch *chan *rethinkdb.SQLExplainRow) {
	rows, err := db.Query(`
		SELECT esh.DIGEST_TEXT from performance_schema.events_statements_summary_by_digest essbd
			INNER JOIN performance_schema.events_statements_history esh
			ON essbd.DIGEST = esh.DIGEST
			WHERE SCHEMA_NAME = "employees" AND esh.EVENT_NAME = "statement/sql/select"
			ORDER BY essbd.LAST_SEEN DESC LIMIT 1
	`)

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
	}

	if err = rows.Err(); err != nil {
		return
	}

	row, err := ExplainScanRows(db, query)

	if err != nil {
		fmt.Printf("%s", err)
	}

	*ch <- row
}
