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

// NewDependency creates
func NewDependency(n, e, p, s string) Dependency {
	dep := Dependency{}
	dep.Name = n
	dep.ExeName = e
	dep.Path = p
	dep.Source = s

	return dep
}

var cfg = new(Config)
var db *sql.DB

func main() {
	env, err := checkEnv("mysql")

	if err != nil {
		log.Fatal(err)
	}

	cfg.Dependencies = []Dependency{NewDependency("MySQL", "mysql", "/usr/bin/mysql", ""),
		NewDependency("PT Query Digest", "pt-query-digest", "/usr/bin/vendor_perl/pt-query-digest", "")}
	cfg.Connections = NewConnString(Auth{env.Username, env.Password}, "mysql", env.Host, env.Port, "tcp")
	err = locateDependencies(cfg.Dependencies)

	if err != nil {
		log.Fatal(err)
	}

	initConn()

}

// initConn initializes the DB connection
func initConn() {
	cs := cfg.Connections

	db, err := sql.Open(cs.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", cs.Auth.Username, cs.Auth.Password, cs.Protocol, cs.Host, cs.Port))
	defer db.Close()

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
	required := []string{"MYSQL_USERNAME", "MYSQL_PASSWORD", "MYSQL_HOST", "MYSQL_PORT"}
	missing := []string{}
	var err error

	for _, v := range required {
		val, isSet := os.LookupEnv(v)

		if val != "" && isSet {
			env = env.set(splitToTitleCase(v, "_", 1), val)
		} else {
			missing = append(missing, v)
			fmt.Printf("missing $%s\n", v)
		}
	}

	if len(missing) != 0 {
		err = errors.New(color.RedString("Please set missing environment variables"))
	}

	return env, err
}

// splitToTitleCase splits a string based on the delimiter and titlecases a parts' results
func splitToTitleCase(s, d string, p int) string {
	return strings.Title((strings.ToLower(strings.Split(s, d)[p])))
}
