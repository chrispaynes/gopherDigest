package config

import (
	"fmt"
	"gopherDigest/src/format"

	r "gopkg.in/gorethink/gorethink.v4"

	"github.com/fatih/color"
)

// RethinkDB defines the host machine's environment variables
type RethinkDB struct {
	address, database, user, password string
}

// NewRethinkDB creates a new RethinkDB Database configuration
func NewRethinkDB(args ...string) *RethinkDB {
	return &RethinkDB{
		user: args[0], password: args[1], database: args[2], address: args[3],
	}
}

// InitRethinkDB initializes the connection
func InitRethinkDB(rdb RethinkDB) (*r.Session, error) {
	if err := executeAdminDuties(rdb); err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	RDBsession, err := connectRethinkDB(rdb)

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

// Check checks for connectivity to external services
func checkRethinkDBConn(s *r.Session) (*Health, error) {
	stat := Health{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("RethinkDB Connection Status")

	server, err := s.Server()

	if err != nil || !s.IsConnected() {
		stat.conn = red.Sprint(" CLOSED")
		stat.port = red.Sprint(" NOT FOUND")
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", err))
		return &stat, err
	}

	stat.conn = green.Sprint(" OPEN")
	stat.port = green.Sprint(" 28015")
	stat.socket = green.Sprintf(" %s", server.Name)

	return &stat, nil
}

func connectRethinkDB(c RethinkDB) (*r.Session, error) {
	db, err := r.Connect(r.ConnectOpts{
		Address:  c.address,
		Database: c.database,
		Username: c.user,
		Password: c.password,
	})

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	conn, err := checkRethinkDBConn(db)

	fmt.Printf("  %+v\n\n", conn)

	return db, nil
}
