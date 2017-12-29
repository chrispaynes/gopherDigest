package conn

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/config"
	"log"
	"os"

	"github.com/fatih/color"
)

// Status defines the status of a network connection
type Status struct {
	connection string
	port       string
	socket     string
	errors     []error
}

func (s Status) printConnStatus() {
	fmt.Printf("  %+v\n", s)
}

// CheckConnectivity does
func CheckConnectivity(d *sql.DB) Status {
	color.New(color.Bold).Println("Checking MySQL Connection")
	sock, sockErr := os.Stat(os.Getenv("MYSQL_SOCKET"))
	cs := Status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	err := d.Ping()

	if err == nil {
		cs.connection = green.Sprint(" OPEN")
		cs.port = green.Sprint(" 3306")
	} else {
		cs.connection = red.Sprint(" CLOSED")
		cs.port = red.Sprint(" NOT FOUND")
		cs.errors = append(cs.errors, fmt.Errorf("Could not connect to DB, %v", d.Stats()))
		log.Fatal()
	}

	if sockErr == nil {
		cs.socket = green.Sprintf(" %s", sock.Name())
	} else {
		cs.socket = red.Sprintf(" %s", sockErr)
	}

	return cs
}

// Init initializes the DB connection
func Init(cfg *config.Config) *sql.DB {
	var err error
	cs := cfg.Connections
	db, err := sql.Open(cs.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", cs.Auth.Username, cs.Auth.Password, cs.Protocol, cs.Host, cs.Port))

	if err != nil {
		log.Fatal(err)
	}

	CheckConnectivity(db).printConnStatus()

	return db
}
