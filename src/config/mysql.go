package config

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/storage"
	"os"
)

// ConnectionStr defines data source and the means of connecting to it
type ConnectionStr struct {
	Auth     Auth
	Driver   string
	Host     string
	Port     string
	Status   interface{}
	Protocol string
}

// NewConnString creates a connection string
func (c *Config) NewConnString(args ...string) *ConnectionStr {
	cn := ConnectionStr{
		Auth:     Auth{Username: os.Getenv(args[0]), Password: os.Getenv(args[1])},
		Driver:   args[2],
		Host:     args[3],
		Port:     args[4],
		Protocol: args[5],
	}
	c.Connections = cn

	return &cn
}

// SetMySQLGlobals sets global MySQL variables
func SetMySQLGlobals(db *sql.DB) ([]string, error) {
	mySQLDb, err := storage.PrintExec(db, []string{
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'",
	})

	if err != nil {
		return []string{}, fmt.Errorf("could not find the 'mysql' database schema\n%s", err)
	}

	// enable slow query logs on MySQL
	return storage.PrintExec(db, []string{
		fmt.Sprintf("USE %s", mySQLDb[0]),
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
