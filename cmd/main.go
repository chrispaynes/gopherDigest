package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

// Config defines main application configuration
type Config struct {
	Dependencies []Dependency
	Connections  ConnectionStr
}

// Dependency defines an application runtime dependency
type Dependency struct {
	Name    string
	ExeName string
	Path    string
	Source  string
}

// ConnStatus definies the status of a network connection
type ConnStatus struct {
	connection string
	port       string
	socket     string
	errors     []error
}

// ConnectionStr defines data source and the means of connecting to it
type ConnectionStr struct {
	Auth     Auth
	Driver   string
	Host     string
	Port     string
	Status   ConnStatus
	Protocol string
}

// NewConnString creates a connection string
func NewConnString(a Auth, d, h, p, ptc string) ConnectionStr {
	conn := ConnectionStr{}
	conn.Driver = d
	conn.Auth = a
	conn.Host = h
	conn.Port = p
	conn.Protocol = ptc

	return conn
}

// Auth defines authorization credentials
type Auth struct {
	Username string
	Password string
}

// Env defines the host machine's environment variables
type Env struct {
	Context  string
	Username string
	Password string
	Host     string
	Port     string
}

type delimitedString struct {
	Prefix    string
	Delimiter string
	Suffix    string
}

type delimitedCollection struct {
	Collection []delimitedString
	Delimiter  string
}

func newDelimitedString(p, d, s string) delimitedString {
	return delimitedString{Prefix: p, Suffix: s, Delimiter: d}
}

// TitlecaseJoiner is the interface implemented by delimited string values
type TitlecaseJoiner interface {
	Titlecase() string
	Join() string
}

func newDelimitedCollection(prefix, delimiter string, suffixColl []string) delimitedCollection {
	dsc := delimitedCollection{Delimiter: delimiter}

	for _, suffix := range suffixColl {
		dsc.Collection = append(dsc.Collection, newDelimitedString(prefix, delimiter, suffix))
	}

	return dsc
}

func newDependency(n, e, p, s string) Dependency {
	return Dependency{Name: n, ExeName: e, Path: p, Source: s}
}

var cfg = new(Config)
var db *sql.DB

func main() {
	env, err := checkEnv("mysql")

	if err != nil {
		log.Fatal(err)
	}

	cfg.Dependencies = []Dependency{newDependency("MySQL", "mysql", "/usr/bin/mysql", ""),
		newDependency("PT Query Digest", "pt-query-digest", "/usr/bin/vendor_perl/pt-query-digest", "")}
	cfg.Connections = NewConnString(Auth{env.Username, env.Password}, "mysql", env.Host, env.Port, "tcp")
	err = locateDependencies(cfg.Dependencies)

	if err != nil {
		log.Fatal(err)
	}

	initConn()
	defer db.Close()

	var hasMySQLDB string
	fmt.Println("mysql> SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'")
	db.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'mysql'").Scan(&hasMySQLDB)

	printSQLStatement(db, []string{
		fmt.Sprintf("USE %s", hasMySQLDB),
		"SET GLOBAL slow_query_log = 'ON'",
		"SET GLOBAL long_query_time = 0.000000",
		"SET GLOBAL log_slow_verbosity = 'query_plan'",
	})

}

// initConn initializes the DB connection
func initConn() {
	cs := cfg.Connections

	var err error

	db, err = sql.Open(cs.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", cs.Auth.Username, cs.Auth.Password, cs.Protocol, cs.Host, cs.Port))

	conn := checkConnectivity(db)
	conn.printConnStatus()

	if err != nil {
		fmt.Print(err)
	}

}

// locateDependencies locates required deps
func locateDependencies(d []Dependency) error {
	var depErr error

	color.New(color.Bold).Println("\nLocating Dependencies")

	for _, dep := range d {
		_, err := os.Stat(dep.Path)
		if err == nil {
			color.New(color.FgHiGreen).Printf("  [\u2713] Using %v executable from %s \n", dep.Name, dep.Path)
		}
		if err != nil {
			color.New(color.FgHiRed).Printf("  [x] Unable to locate %s executable at %s\n", dep.Name, dep.Path)
			depErr = fmt.Errorf("ERROR: Missing Required Dependencies")
		}
	}

	fmt.Println()

	return depErr

}

func (cs ConnStatus) printConnStatus() {
	fmt.Printf("  %+v\n", cs)
}

func checkConnectivity(d *sql.DB) ConnStatus {
	color.New(color.Bold).Println("Checking MySQL Connection")

	sock, sockErr := os.Stat("/var/run/mysqld/mysqld.sock")
	cs := ConnStatus{}

	err := d.Ping()

	if err == nil && sockErr == nil {
		green := color.New(color.FgGreen, color.Bold)
		cs.connection = green.Sprint(" OPEN")
		cs.socket = green.Sprintf(" %s", sock.Name())
		cs.port = green.Sprint(" 3306")
	} else {
		red := color.New(color.FgRed, color.Bold)
		cs.connection = red.Sprint(" CLOSED")
		cs.socket = red.Sprintf(" %s", sockErr)
		cs.port = red.Sprint(" NONE")
		cs.errors = append(cs.errors, fmt.Errorf("Could not connect to DB, %v", d.Stats()))
	}

	return cs
}

// set is helper function to set a field's value within a struct
func (e Env) set(k, v string) Env {
	reflect.ValueOf(&e).Elem().FieldByName(k).SetString(v)
	return e
}

// checkEnv verifies the necessary environment variables are defined
func checkEnv(c string) (Env, error) {
	env := Env{}
	required := newDelimitedCollection("MYSQL", "_", []string{"USERNAME", "PASSWORD", "HOST", "PORT"})
	missing := []string{}
	var err error

	for _, v := range required.Collection {
		val, isSet := os.LookupEnv(v.Join())

		if val != "" && isSet {
			env = env.set(splitToTitlecase(1, v), val)
		} else {
			missing = append(missing, v.Join())
		}
	}

	if len(missing) != 0 {
		err = errors.New(color.RedString(fmt.Sprintf("Please set missing environment variables: %s", missing)))
	}

	return env, err
}

func splitToTitlecase(p int, tj TitlecaseJoiner) string {
	var str string

	if _, ok := tj.(delimitedString); ok {
		str = tj.Titlecase()
	}

	return str
}

// Titlecase titlecases a delimitedString field
func (ds delimitedString) Titlecase() string {
	return strings.Title(strings.ToLower(ds.Suffix))
}

// Join concatenates a delimitedString
func (ds delimitedString) Join() string {
	return strings.Join([]string{ds.Prefix, ds.Suffix}, ds.Delimiter)
}

func verify(db *sql.DB, a, v string) bool {
	var varKey, varValue string

	fmt.Println(fmt.Sprintf("mysql> SHOW VARIABLES LIKE '%s'", v))
	row := db.QueryRow(fmt.Sprintf("SHOW VARIABLES LIKE '%s'", v))

	row.Scan(&varKey, &varValue)

	return varValue == a

}

func printSQLStatement(db *sql.DB, stmnts []string) {
	for _, s := range stmnts {
		fmt.Printf("mysql> %s;\n", s)
		db.Exec(s)
	}
}
