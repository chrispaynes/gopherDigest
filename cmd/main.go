package main

import (
	"gopherDigest/pkg/config"
	"gopherDigest/pkg/mysql"
	"gopherDigest/pkg/rethinkdb"
	"log"
	"os"
	"sync"
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

	mysqlRootConfig := mysql.New("",
		config.GetSecrets(os.Getenv, "MYSQL", "_", "USER", "PASSWORD", "HOST", "PORT", "MAX_CONNECTIONS")...)

	mysqlUserConfig := mysql.New("employees",
		config.GetSecrets(os.Getenv, "MYSQL", "_", "USER", "PASSWORD", "HOST", "PORT", "MAX_CONNECTIONS")...)

	// TODO: only init if the database isn't initialized
	db, err := mysql.Init(mysqlRootConfig)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	queryString := "SELECT * FROM salaries s LEFT JOIN employees e USING(emp_no) LEFT JOIN dept_emp d USING(emp_no)"
	db2, _ := mysql.Connect(mysqlUserConfig)
	defer db2.Close()

	// TODO: set MySQL max connections as a configurable based on available ram and buffers
	for i := 0; i < mysqlUserConfig.GetMaxConns(); i++ {
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			go db2.Query(queryString)
		}()

		wg.Wait()

		explain := make(chan *rethinkdb.SQLExplainRow)

		go func() {
			rows, err := db2.Query(`
					SELECT essbd.DIGEST_TEXT from performance_schema.events_statements_summary_by_digest essbd
					LEFT JOIN performance_schema.events_statements_history esh
					ON essbd.DIGEST = esh.DIGEST
					WHERE SCHEMA_NAME = "employees" AND essbd.DIGEST_TEXT LIKE "SELECT%"
					ORDER BY essbd.LAST_SEEN DESC LIMIT 1
				`)

			if err != nil {
				return
			}

			defer rows.Close()

			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					log.Fatal(err)
				}
			}

			if err = rows.Err(); err != nil {
				return
			}

			se, _ := mysql.ExplainScanRows(db2, queryString)

			explain <- se
		}()

		se := <-explain
		seq := []rethinkdb.SQLExplainRow{*se}

		r.Table("Queries").Insert(rethinkdb.QueryDump{Search: queryString, Timestamp: time.Now().Unix(), QueryTime: r.Now(), SQLExplainRows: seq}).Run(RDBsession)
	}

}
