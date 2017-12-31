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

func (s Status) printStatus() {
	fmt.Printf("  %+v\n", s)
}

// CheckConnectivity checks for connectivity to external services
func CheckConnectivity(d *sql.DB) (*Status, error) {
	cs := Status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("Checking MySQL Connection")

	err := d.Ping()
	if err != nil {
		cs.connection = red.Sprint(" CLOSED")
		cs.port = red.Sprint(" NOT FOUND")
		cs.errors = append(cs.errors, fmt.Errorf("could not connect to DB, %v", d.Stats()))
		return &cs, err
	}

	socketPath := os.Getenv("MYSQL_SOCKET")
	socket, err := os.Stat(socketPath)
	if err != nil {
		cs.socket = red.Sprintf(" %s", err)
		return &cs, fmt.Errorf("could not locate the MySQL Socket file, %s", err)
	}

	cs.connection = green.Sprint(" OPEN")
	cs.port = green.Sprint(" 3306")
	cs.socket = green.Sprintf(" %s", socket.Name())

	return &cs, nil
}

// Init initializes the DB connection
func Init(cfg *config.Config) (*sql.DB, error) {
	cs := cfg.Connections
	db, err := sql.Open(cs.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", cs.Auth.Username, cs.Auth.Password, cs.Protocol, cs.Host, cs.Port))

	if err != nil {
		log.Fatal(err)
		return db, fmt.Errorf("could not open database connection %s", err)
	}

	conn, err := CheckConnectivity(db)

	if err != nil {
		return db, fmt.Errorf("could not maintain database connection %s", err)
	}

	conn.printStatus()
	fmt.Println()

	return db, nil
}
