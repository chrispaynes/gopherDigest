package conn

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/config"
	"os"

	"github.com/fatih/color"
)

// Status definies the status of a network connection
type Status struct {
	connection string
	port       string
	socket     string
	errors     []error
}

// NewConnString creates a connection string
func NewConnString(a config.Auth, d, h, p, ptc string) config.ConnectionStr {
	cn := config.ConnectionStr{}
	cn.Driver = d
	cn.Auth = a
	cn.Host = h
	cn.Port = p
	cn.Protocol = ptc

	return cn
}

func (s Status) printConnStatus() {
	fmt.Printf("  %+v\n", s)
}

// CheckConnectivity does
func CheckConnectivity(d *sql.DB) Status {
	color.New(color.Bold).Println("Checking MySQL Connection")

	sock, sockErr := os.Stat("/var/run/mysqld/mysqld.sock")
	cs := Status{}

	err := d.Ping()

	if err == nil && sockErr == nil {
		green := color.New(color.FgGreen, color.Bold)
		cs.connection = green.Sprint(" OPEN")
		cs.socket = green.Sprintf(" %s", sock.Name())
		cs.port = green.Sprint(" 3306")
	} else {
		red := color.New(color.FgRed, color.Bold)
		cs.connection = red.Sprint(" CLOSED")
		cs.socket = red.Sprintf(" %s", sockErr)
		cs.port = red.Sprint(" NONE")
		cs.errors = append(cs.errors, fmt.Errorf("Could not connect to DB, %v", d.Stats()))
	}

	return cs
}

// Init initializes the DB connection
func Init(cfg *config.Config) *sql.DB {
	cs := cfg.Connections

	var err error

	db, err := sql.Open(cs.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", cs.Auth.Username, cs.Auth.Password, cs.Protocol, cs.Host, cs.Port))

	CheckConnectivity(db).printConnStatus()

	if err != nil {
		fmt.Print(err)
	}

	return db
}
