package config

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/storage"
)

// SetGlobals sets global MySQL variables
func SetGlobals(db *sql.DB) error {
	var hasMySQLDB string
	fmt.Println("mysql> SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'")
	db.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'").Scan(&hasMySQLDB)

	// enable slow query logs on MySQL
	return storage.PrintExec(db, []string{
		fmt.Sprintf("USE %s", hasMySQLDB),
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
