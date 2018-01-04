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

// MySQL defines data source and the means of connecting to it
type MySQL struct {
	user     string
	password string
	host     string
	port     int
	protocol string
}

func enableSlowQueryLogs(db *sql.DB) error {
	_, err := storage.PrintExec(db, []string{
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

	return err
}

// NewMySQL creates a new MySQL Database configuration
func NewMySQL(u, p1, h, p2 string, p3 int) MySQL {
	return MySQL{
		user:     u,
		password: p1,
		host:     h,
		protocol: p2,
		port:     p3,
	}
}

// CheckMySQLConn checks for connectivity to external services with the ability to retry connections
func CheckMySQLConn(d *sql.DB, totalRetries, remainingRetries int) (*Health, error) {
	stat := Health{}
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
			CheckMySQLConn(d, totalRetries, next)
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
func InitMySQLDB(m MySQL) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%v)/", m.user, m.password, m.protocol, m.host, m.port))

	if err != nil {
		return nil, fmt.Errorf("could not open database connection\n%s", err)
	}

	conn, err := CheckMySQLConn(db, 10, 10)

	if err != nil {
		return nil, fmt.Errorf("could not maintain database connection\n%s", err)
	}

	_, err = storage.PrintExec(db, []string{
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'",
	})

	if err != nil {
		return nil, fmt.Errorf("could not find the 'mysql' database schema\n%s", err)
	}

	// print connection status
	fmt.Printf("  %+v\n\n", conn)

	err = enableSlowQueryLogs(db)

	if err != nil {
		return nil, fmt.Errorf("could not enable slow query logs\n%s", err)
	}

	return db, nil
}
