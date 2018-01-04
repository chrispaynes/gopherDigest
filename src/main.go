package main

import (
	"fmt"
	"gopherDigest/src/config"
	"gopherDigest/src/storage"
	"log"

	_ "github.com/go-sql-driver/mysql"
	r "gopkg.in/gorethink/gorethink.v4"
)

func main() {
	var err error

	cfg, err := config.New()

	if err != nil {
		log.Fatal(err)
	}

	RDBsession, err := config.ConnectRethinkDB(cfg.Env[1])
	defer RDBsession.Close()

	if err != nil {
		log.Fatalln(err)
	}

	// WIP - RETHINKDB move to config module
	// To login with a username and password you should first create a user,
	// this can be done by writing to the users system table and then grant
	// that user access to any tables or databases they need access to.
	r.DB("rethinkdb").Table("users").Insert(map[string]string{
		"id":       "fakeUser123",
		"password": "fakePassword123",
	}).Exec(RDBsession)

	// WIP - RETHINKDB move to config module
	// then grant that user access to any tables or databases they need access to
	r.Table("Queries").Grant("fakeUser123", map[string]bool{
		"read":  true,
		"write": true,
	}).Exec(RDBsession)

	var row []string
	res, err := r.DBList().Run(RDBsession)
	if err != nil {
		fmt.Println("could not load the RethinkDB databases")
	}

	res.All(&row)

	// check if DB already exists
	// todo add similar check for DB table
	if indexOfString("GopherDigest", row) == -1 {
		r.DBCreate("GopherDigest").Exec(RDBsession)
		r.DB("GopherDigest").TableCreate("Queries").Exec(RDBsession)
	}

	db, err := config.ConnectMySQL(cfg.Connections)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	config.SetMySQLGlobals(db)

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

	r.Table("Queries").Insert(storage.QueryDump{queryString, 0.34343, r.Now(), se}).Run(RDBsession)

}

// indexOfStrings returns the index of a string within a slice or -1 if it does not exist
func indexOfString(dbname string, collection []string) int {
	for i := range collection {
		if collection[i] == dbname {
			return i
		}
	}
	return -1
}
