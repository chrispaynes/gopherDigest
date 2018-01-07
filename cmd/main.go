package main

import (
	"fmt"
	"gopherDigest/pkg/config"
	"gopherDigest/pkg/mysql"
	"gopherDigest/pkg/rethinkdb"
	"log"
	"os"
	"time"

	r "gopkg.in/gorethink/gorethink.v4"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var err error

	_, err = config.New()

	if err != nil {
		log.Fatal(err)
	}

	rConfig := rethinkdb.New(
		config.GetSecrets(os.Getenv, "RDB", "_", "USERNAME", "PASSWORD", "DATABASE", "ADDRESS")...)

	RDBsession, err := rethinkdb.Init(*rConfig)
	defer RDBsession.Close()

	if err != nil {
		log.Fatalln(err)
	}

	mConfig := mysql.New("",
		config.GetSecrets(os.Getenv, "MYSQL", "_", "USER", "PASSWORD", "HOST", "PORT")...)

	// TODO: only init if there isn't a connection
	db, err := mysql.Init(mConfig)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	db.Exec("USE employees")

	// generate slow query log data and move data to RethinkDB
	queryString := "SELECT * FROM salaries s LEFT JOIN employees e USING(emp_no) LEFT JOIN dept_emp d USING(emp_no)"

	// TODO: set MySQL max connections as a configurable based on available ram and buffers
	for i := 0; i < 10000; i++ {

		fmt.Println(i)
		go db.Query(queryString)
		defer db.Close()

		explainRows, err := mysql.Explain(db, queryString)
		defer explainRows.Close()

		if err != nil {
			log.Fatalf("query execution failed \n%s", err)
		}

		se := rethinkdb.SQLExplain{}

		if err := mysql.ScanRows(explainRows, se); err != nil {
			log.Fatalf("failed to copy the row columns to the destination \n%s", err)
		}

		rows, err := db.Query(` 
				SELECT essbd.DIGEST_TEXT from performance_schema.events_statements_summary_by_digest essbd
				LEFT JOIN performance_schema.events_statements_history esh
				ON essbd.DIGEST = esh.DIGEST
				WHERE SCHEMA_NAME = "employees" AND essbd.DIGEST_TEXT LIKE "SELECT%"
				ORDER BY essbd.LAST_SEEN DESC LIMIT 1
			`)
		defer rows.Close()

		if err != nil {
			log.Fatal("failed to compute performance query")
		}

		var test []string

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Fatal(err)
			}
			test = append(test, name)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		go r.Table("Queries").Insert(rethinkdb.QueryDump{Search: queryString, Timestamp: time.Now().Unix(), QueryTime: r.Now(), SQLExplain: se}).Run(RDBsession)
	}

}
