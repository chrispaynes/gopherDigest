package config

import (
	"fmt"
	"gopherDigest/src/format"

	r "gopkg.in/gorethink/gorethink.v4"

	"github.com/fatih/color"
)

// InitRethinkDB initializes the connection
func InitRethinkDB(c *Config) (*r.Session, error) {

	RDBsession, err := connectRethinkDB(c.Env[1])

	// WIP - RETHINKDB move to config module
	// To login with a username and password you should first create a user,
	// this can be done by writing to the users system table and then grant
	// that user access to any tables or databases they need access to.
	r.DB("rethinkdb").Table("users").Insert(map[string]string{
		"id":       "fakeUser123",
		"password": "fakePassword123",
	}).Exec(RDBsession)

	// WIP - RETHINKDB move to config module
	// then grant that user access to any tables or databases they need access to
	r.Table("Queries").Grant("fakeUser123", map[string]bool{
		"read":  true,
		"write": true,
	}).Exec(RDBsession)

	var row []string
	res, err := r.DBList().Run(RDBsession)
	if err != nil {
		fmt.Println("could not load the RethinkDB databases")
	}

	res.All(&row)

	// check if DB already exists
	// todo add similar check for DB table
	if format.IndexOfString("GopherDigest", row) == -1 {
		r.DBCreate("GopherDigest").Exec(RDBsession)
		r.DB("GopherDigest").TableCreate("Queries").Exec(RDBsession)
	}

	return RDBsession, err
}

// Check checks for connectivity to external services
func checkRethinkDBConn(d *r.Session) (*status, error) {
	stat := status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("RethinkDB Connection Status")

	server, err := d.Server()

	if err != nil || !d.IsConnected() {
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

func connectRethinkDB(c SetterReader) (*r.Session, error) {

	db, err := r.Connect(r.ConnectOpts{
		Address:  c.Read("Address"),
		Database: c.Read("Database"),
		Username: c.Read("Username"),
		Password: c.Read("Password"),
	})

	conn, err := checkRethinkDBConn(db)

	fmt.Printf("  %+v\n\n", conn)

	return db, err
}
