package main

import (
	"database/sql"
	"fmt"
	"gopherDigest/src/config"
	"gopherDigest/src/conn"
	"gopherDigest/src/storage"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var err error
	var cfg = new(config.Config)
	var db *sql.DB

	cfg.Env, err = config.Check("mysql")

	if err != nil {
		log.Fatal(err)
	}

	cfg.AddConfig("connection", "mysql", "localhost", cfg.Env.Port, "tcp")
	cfg.AddConfig("dependency", "MySQL", "mysql", "/usr/bin/mysql", "")
	cfg.AddConfig("dependency", "PT Query Digest", "pt-query-digest", "/usr/bin/vendor_perl/pt-query-digest", "")

	err = config.LocateDependencies(cfg.Dependencies)

	if err != nil {
		log.Fatal(err)
	}

	db = conn.Init(cfg)
	defer db.Close()

	var hasMySQLDB string
	fmt.Println("mysql> SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'")
	db.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'").Scan(&hasMySQLDB)

	storage.PrintExec(db, []string{
		fmt.Sprintf("USE %s", hasMySQLDB),
		"SET GLOBAL slow_query_log = 'ON'",
		"SET GLOBAL long_query_time = 0.000000",
		"SET GLOBAL log_slow_verbosity = 'query_plan'",
	})

}
