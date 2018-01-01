package storage

import r "gopkg.in/gorethink/gorethink.v4"

// QueryDump represents a MySQL Query Performance Dump
type QueryDump struct {
	Search        string     `gorethink:"Search"`
	ExecutionTime float64    `gorethink:"ExecutionTime"`
	QueryTime     r.Term     `gorethink:"QueryTime"`
	SQLExplain    SQLExplain `gorethink:"SQLExplain"`
}

// SQLExplain represents a MySQL Explain Result
type SQLExplain struct {
	ID           int     `gorethink:"ZID"`
	SelectType   *string `gorethink:"SelectType"`
	Table        *string `gorethink:"Table"`
	Partitions   *string `gorethink:"Partitions"`
	Ztype        *string `gorethink:"Ztype"`
	PossibleKeys *string `gorethink:"PossibleKeys"`
	Key          *string `gorethink:"Key"`
	KeyLen       *string `gorethink:"KeyLen"`
	Ref          *string `gorethink:"Ref"`
	Rows         int     `gorethink:"Rows"`
	Filtered     []byte  `gorethink:"Filtered"`
	Extra        *string `gorethink:"Extra"`
}
