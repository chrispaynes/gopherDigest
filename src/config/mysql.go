package config

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/storage"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

// MySQLConn defines data source and the means of connecting to it
type MySQLConn struct {
	Auth     Auth
	Driver   string
	Host     string
	Port     string
	Status   interface{}
	Protocol string
}

// NewConnString creates a connection string
func (c *Config) NewConnString(args ...string) *MySQLConn {
	cn := MySQLConn{
		Auth:     Auth{Username: os.Getenv(args[0]), Password: os.Getenv(args[1])},
		Driver:   args[2],
		Host:     args[3],
		Port:     args[4],
		Protocol: args[5],
	}
	c.Connections = cn

	return &cn
}

func enableSlowQueryLogs(db *sql.DB) ([]string, error) {
	return storage.PrintExec(db, []string{
		"USE mysql",
		"SET @@GLOBAL.slow_query_log = 'ON'",
		"SET long_query_time = 0",
		"SET @@GLOBAL.long_query_time = 0",
		"SET @@GLOBAL.log_slow_admin_statements = 'ON'",
		"SET @@GLOBAL.log_slow_slave_statements = 'ON'",
		"SET sql_log_off = 'ON'",
		"SET @@GLOBAL.sql_log_off = 'ON'",
		"SET @@GLOBAL.log_queries_not_using_indexes = 'ON'",
	})
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

// InitMySQLDB initializes the MySQL Database connection
func InitMySQLDB(c MySQLConn) (*sql.DB, error) {
	db, err := sql.Open(c.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", c.Auth.Username, c.Auth.Password, c.Protocol, c.Host, c.Port))

	if err != nil {
		return db, fmt.Errorf("could not open database connection\n%s", err)
	}

	conn, err := checkMySQLConn(db, 10, 10)

	if err != nil {
		return db, fmt.Errorf("could not maintain database connection\n%s", err)
	}

	_, err = storage.PrintExec(db, []string{
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'",
	})

	if err != nil {
		return nil, fmt.Errorf("could not find the 'mysql' database schema\n%s", err)
	}

	// print connection status
	fmt.Printf("  %+v\n\n", conn)

	enableSlowQueryLogs(db)

	return db, nil
}
