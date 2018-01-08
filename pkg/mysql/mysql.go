package mysql

import (
	"database/sql"
	"fmt"
	"gopherDigest/pkg/config"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
)

// MySQL defines data source and the means of connecting to it
type MySQL struct {
	user           string
	password       string
	host           string
	port           int
	database       string
	maxConnections int
}

func enableSlowQueryLogs(db *sql.DB) error {
	_, err := printExec(db, []string{
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

// New creates a new MySQL Database configuration
func New(d string, args ...string) MySQL {
	port, _ := strconv.Atoi(args[3])
	maxConn, _ := strconv.Atoi(args[4])

	return MySQL{
		user: args[0], password: args[1], host: args[2], port: port, database: "employees", maxConnections: maxConn,
	}
}

// CheckConnection checks for connectivity to external services with the ability to retry connections
func CheckConnection(d *sql.DB, totalRetries, remainingRetries int) (*config.Health, error) {
	stat := config.Health{}
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
			CheckConnection(d, totalRetries, next)
		}
	}()

	err := d.Ping()

	if err != nil {
		stat.Conn = red.Sprint(" CLOSED")
		stat.Port = red.Sprint(" NOT FOUND")
		stat.Errors = append(stat.Errors, fmt.Errorf("could not connect to DB\n%v", d.Stats()))
		panic(fmt.Sprintf("could not connect to database on attempt #%d %v", remainingRetries, err))
	}

	sock, err := os.Stat(os.Getenv("MYSQL_SOCKET"))
	if err != nil {
		stat.Errors = append(stat.Errors, fmt.Errorf("could not connect to DB\n%v", err))
		stat.Socket = red.Sprintf(" %s", err)
		return &stat, fmt.Errorf("could not locate the MySQL file\n%s", err)
	}

	stat.Conn = green.Sprint(" OPEN")
	stat.Port = green.Sprint(" 3306")
	stat.Socket = green.Sprintf(" %s", sock.Name())

	return &stat, nil
}

// Init initializes the MySQL Database connection
func Init(m MySQL) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%v)/employees", m.user, m.password, m.host, m.port))
	defer db.Close()

	db.SetMaxIdleConns(2)

	if err != nil {
		return nil, fmt.Errorf("could not open database connection\n%s", err)
	}

	conn, err := CheckConnection(db, 10, 10)

	if err != nil {
		return nil, fmt.Errorf("could not maintain database connection\n%s", err)
	}

	_, err = printExec(db, []string{
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'",
	})

	if err != nil {
		return nil, fmt.Errorf("could not find the 'mysql' database schema\n%s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET global max_connections = %d", m.maxConnections))

	if err != nil {
		fmt.Printf("could not set max_connections for SQL database\n%s\n", err)
	}

	// print connection status
	fmt.Printf("  %+v\n\n", conn)

	err = enableSlowQueryLogs(db)

	if err != nil {
		return nil, fmt.Errorf("could not enable slow query logs\n%s", err)
	}

	return db, nil
}

// printExec prints SQL statements to standard output
func printExec(db *sql.DB, stmnts []string) ([]string, error) {
	results := []string{}

	// execute statements and return result rows to array
	for _, s := range stmnts {
		fmt.Printf("mysql> %s;\n", s)
		rows, err := db.Query(s)
		defer rows.Close()

		if err != nil {
			return []string{}, fmt.Errorf("could not execute the SQL statement %s", err)
		}

		var row string

		for rows.Next() {
			rows.Scan(&row)
		}

		results = append(results, row)
	}

	return results, nil
}

// Connect initializes the MySQL Database connection
func Connect(m MySQL) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%v)/employees", m.user, m.password, m.host, m.port))

	if err != nil {
		return nil, fmt.Errorf("could not open database connection\n%s", err)
	}

	// _, err = CheckConnection(db, 10, 10)

	// if err != nil {
	// 	return nil, fmt.Errorf("could not maintain database connection\n%s", err)
	// }

	return db, nil
}

// GetMaxConns gets the maximum number of open connections allowed for the database configuration.
func (m MySQL) GetMaxConns() int {
	return m.maxConnections
}
