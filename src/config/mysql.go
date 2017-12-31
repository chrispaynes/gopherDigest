package config

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/storage"
)

// SetMySQLGlobals sets global MySQL variables
func SetMySQLGlobals(db *sql.DB) ([]string, error) {
	mySQLDb, err := storage.PrintExec(db, []string{
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'",
	})

	if err != nil {
		return []string{}, fmt.Errorf("could not find the 'mysql' database schema %s", err)
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
