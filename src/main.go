package main

import (
	"fmt"
	"gopherDigest/src/config"
	"gopherDigest/src/storage"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	r "gopkg.in/gorethink/gorethink.v4"
)

func main() {
	var err error

	_, err = config.New()

	if err != nil {
		log.Fatal(err)
	}

	rConfig := config.NewRethinkDB(os.Getenv("RDB_USERNAME"),
		os.Getenv("RDB_PASSWORD"), os.Getenv("RDB_DATABASE"), os.Getenv("RDB_ADDRESS"))

	RDBsession, err := config.InitRethinkDB(*rConfig)
	defer RDBsession.Close()

	if err != nil {
		log.Fatalln(err)
	}

	mConfig := config.NewMySQL(os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"), "localhost", "tcp", 3306)

	db, err := config.InitMySQLDB(mConfig)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	// TODO MOVE TO the data storage package
	// generate slow query log data and move data to RethinkDB
	var test []string
	queryString := "SELECT * FROM employees.employees LEFT JOIN employees.dept_emp USING(emp_no)"
	db.Exec(queryString)
	se := storage.SQLExplain{}
	explainRows, _ := db.Query("EXPLAIN " + queryString)
	defer explainRows.Close()

	for explainRows.Next() {
		err := explainRows.Scan(&se.ID, &se.SelectType, &se.Table,
			&se.Partitions, &se.Ztype, &se.PossibleKeys, &se.Key,
			&se.KeyLen, &se.Ref, &se.Rows, &se.Filtered, &se.Extra)
		if err != nil {
			fmt.Printf("ERROR %s", err)
		}
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

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		rows.Scan(&name)
		test = append(test, name)

		fmt.Printf("%s\n", name)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	r.Table("Queries").Insert(storage.QueryDump{Search: queryString, ExecutionTime: 0.34343, QueryTime: r.Now(), SQLExplain: se}).Run(RDBsession)

}
