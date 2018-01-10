package rethinkdb

import (
	"time"

	r "gopkg.in/gorethink/gorethink.v4"
)

// InsertSQLExplain inserts a SQLExplain struct into the rethinkDB database
func InsertSQLExplain(rdb *r.Session, seq []SQLExplainRow, queryString string) {
	r.Table("Queries").Insert(queryDump{Search: queryString, Timestamp: time.Now().Unix(), QueryTime: r.Now(), SQLExplainRows: seq}).Run(rdb)
}
