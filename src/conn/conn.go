package conn

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/config"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	r "gopkg.in/gorethink/gorethink.v4"
)

// status defines the status of a network connection
type status struct {
	conn   string
	port   string
	socket string
	errors []error
}

// checkMySQLConn checks for connectivity to external services with the ability to retry connections
func checkMySQLConn(d *sql.DB, totalRetries, remainingRetries int) (*status, error) {
	stat := status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("Checking MySQL Connection Status")

	defer func() {
		if r := recover(); r != nil {
			next := remainingRetries - 1
			fmt.Printf("could not connect to MySQL on the first attempt, attempting to reconnect:\n%d retry attempt(s) remaining\n\n", next+1)
			time.Sleep(15 * time.Second)
			if next < 0 {
				log.Fatal("could not connect to MySQL database")
			}
			checkMySQLConn(d, totalRetries, next)
		}
	}()

	err := d.Ping()

	if err != nil {
		stat.conn = red.Sprint(" CLOSED")
		stat.port = red.Sprint(" NOT FOUND")
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", d.Stats()))
		panic(fmt.Sprintf("could not connect to database on attempt #%d %v", remainingRetries, err))
	}

	sock, err := os.Stat(os.Getenv("MYSQL_SOCKET"))
	if err != nil {
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", err))
		stat.socket = red.Sprintf(" %s", err)
		return &stat, fmt.Errorf("could not locate the MySQL file\n%s", err)
	}

	stat.conn = green.Sprint(" OPEN")
	stat.port = green.Sprint(" 3306")
	stat.socket = green.Sprintf(" %s", sock.Name())

	return &stat, nil
}

// ConnectMySQL initializes the DB connection
func ConnectMySQL(c config.MySQLConn) (*sql.DB, error) {
	db, err := sql.Open(c.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", c.Auth.Username, c.Auth.Password, c.Protocol, c.Host, c.Port))

	if err != nil {
		return db, fmt.Errorf("could not open database connection\n%s", err)
	}

	conn, err := checkMySQLConn(db, 10, 10)

	if err != nil {
		return db, fmt.Errorf("could not maintain database connection\n%s", err)
	}

	// print connection status
	fmt.Printf("  %+v\n\n", conn)

	return db, nil
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

// ConnectRethinkDB starts Connect RethinkDB connection
func ConnectRethinkDB(c config.SetterReader) (*r.Session, error) {

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
