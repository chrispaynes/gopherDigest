package rethinkdb

import (
	"fmt"
	"gopherDigest/pkg/config"
	"gopherDigest/pkg/format"

	r "gopkg.in/gorethink/gorethink.v4"

	"github.com/fatih/color"
)

// QueryDump represents a MySQL Query Performance Dump
type QueryDump struct {
	Search     string     `gorethink:"Search"`
	QueryTime  r.Term     `gorethink:"QueryTime"`
	SQLExplain SQLExplain `gorethink:"SQLExplain"`
	Timestamp  int64      `gorethink:"Timestamp"`
}

//SQLExplain represents a MySQL Explain Result
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

// RethinkDB defines the host machine's environment variables
type RethinkDB struct {
	address, database, user, password string
}

// New creates a new RethinkDB Database configuration
func New(args ...string) *RethinkDB {
	return &RethinkDB{
		user: args[0], password: args[1], database: args[2], address: args[3],
	}
}

// Init initializes the connection
func Init(rdb RethinkDB) (*r.Session, error) {
	if err := executeAdminDuties(rdb); err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	RDBsession, err := Connect(rdb)

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	return RDBsession, err
}

func executeAdminDuties(rdb RethinkDB) error {
	// connect as admin
	// TODO initialize DB with admin password and add to ENV
	RDBsession, err := r.Connect(r.ConnectOpts{
		Address: rdb.address,
	})

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	var DBrow []string
	res, err := r.DBList().Run(RDBsession)

	if err != nil {
		fmt.Println("could not load the RethinkDB databases")
	}

	res.All(&DBrow)

	// create Database if it doesn't already exist
	if format.IndexOfString("GopherDigest", DBrow) == -1 {
		r.DBCreate("GopherDigest").Exec(RDBsession)
	}

	err = r.DB("rethinkdb").Table("users").Insert(map[string]string{
		"id":       rdb.user,
		"password": rdb.password,
	}).Exec(RDBsession)

	if err != nil {
		return fmt.Errorf("failed to create user %s", rdb.user)
	}

	var tblRow []string
	tables, err := r.DB("GopherDigest").TableList().Run(RDBsession)

	tables.All(&tblRow)

	// create table if it doesn't already exist
	if format.IndexOfString("Queries", tblRow) == -1 {
		if err := r.DB("GopherDigest").TableCreate("Queries").Exec(RDBsession); err != nil {
			return fmt.Errorf("failed to create the 'Queries' table")
		}
	}

	if err != nil {
		fmt.Println("could not load the GopherDigest database tables")
	}

	// then grant that user access to any tables or databases they need access to
	err = r.DB("GopherDigest").Table("Queries").Grant(rdb.user, map[string]bool{
		"read":  true,
		"write": true,
	}).Exec(RDBsession)

	if err != nil {
		return fmt.Errorf("could not grant administrative access to the 'Queries' table for user %s", rdb.user)
	}

	return nil

}

func checkConnection(s *r.Session) (*config.Health, error) {
	stat := config.Health{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("RethinkDB Connection Status")

	server, err := s.Server()

	if err != nil || !s.IsConnected() {
		stat.Conn = red.Sprint(" CLOSED")
		stat.Port = red.Sprint(" NOT FOUND")
		stat.Errors = append(stat.Errors, fmt.Errorf("could not connect to DB\n%v", err))
		return &stat, err
	}

	stat.Conn = green.Sprint(" OPEN")
	stat.Port = green.Sprint(" 28015")
	stat.Socket = green.Sprintf(" %s", server.Name)

	return &stat, nil
}

// Connect creates a connection to a RethinkDB database
func Connect(c RethinkDB) (*r.Session, error) {
	db, err := r.Connect(r.ConnectOpts{
		Address:  c.address,
		Database: c.database,
		Username: c.user,
		Password: c.password,
	})

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	conn, err := checkConnection(db)

	fmt.Printf("  %+v\n\n", conn)

	return db, nil
}
