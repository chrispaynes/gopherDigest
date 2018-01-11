package main

import (
	"gopherDigest/pkg/config"
	"gopherDigest/pkg/mysql"
	"gopherDigest/pkg/rethinkdb"
	"log"
	"os"
	"sync"

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

		explainCh := make(chan *rethinkdb.SQLExplainRow)

		go mysql.FetchEventSummary(db2, queryString, &explainCh)

		explain := <-explainCh

		rethinkdb.InsertSQLExplain(RDBsession, []rethinkdb.SQLExplainRow{*explain}, queryString)
	}
}
