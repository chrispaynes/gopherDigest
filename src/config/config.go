package config

import (
	"database/sql"
	"errors"
	"fmt"
	"gopherDigest/src/format"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"
	r "gopkg.in/gorethink/gorethink.v4"
)

// Config defines main application configuration
type Config struct {
	Dependencies []Dependency
	Connections  MySQLConn
	Env          []SetterReader
	Required     *format.DelimitedCollection
}

// Dependency defines an application runtime dependency
type Dependency struct {
	Name    string
	ExeName string
	Path    string
	Source  string
}

// MySQLEnv defines the host machine's environment variables
type MySQLEnv struct {
	Context, User, Password, Host, Port string
}

// RethinkDBEnv defines the host machine's environment variables
type RethinkDBEnv struct {
	Context, Address, Database, Username, Password string
}

// SetterReader defines the interface implemented by environment variables
type SetterReader interface {
	set(k, v string) error
	Read(k string) string
}

// Auth defines authorization credentials
type Auth struct {
	Username string
	Password string
}

// NewDependency creates a new runtime dependency
func (c *Config) NewDependency(args ...string) *Dependency {
	c.Dependencies = append(c.Dependencies, Dependency{Name: args[0], ExeName: args[1], Path: args[2], Source: args[3]})

	return &Dependency{Name: args[0], ExeName: args[1], Path: args[2], Source: args[3]}
}

// NewEnv stores a new set of operating system environment variable
func (c *Config) NewEnv(dc format.DelimitedCollection, s SetterReader) *SetterReader {
	env := s

	c.Required = &dc

	for _, v := range dc.Collection {
		val, isSet := os.LookupEnv(v.Join())

		if val != "" && isSet {
			env.set(format.SplitToTitlecase(1, v), val)
		}
	}

	c.Env = append(c.Env, env)

	return &env
}

// locateDependencies locates required runtime dependencies
func (c *Config) locateDependencies() (*[]Dependency, error) {
	color.New(color.Bold).Println("\nLocating Dependencies")
	for _, dep := range c.Dependencies {
		_, err := os.Stat(dep.Path)
		if err != nil {
			color.New(color.FgHiRed).Printf("  [x] Unable to locate %s executable at %s\n", dep.Name, dep.Path)
			return &c.Dependencies, fmt.Errorf("ERROR: Missing Required Dependency %s", err)
		}
		color.New(color.FgHiGreen).Printf("  [\u2713] Using %v executable from %s \n", dep.Name, dep.Path)
	}
	fmt.Println()
	return &c.Dependencies, nil
}

func (m MySQLEnv) set(k, v string) error {
	return setEnvVar(&m, k, v)
}

func (r RethinkDBEnv) set(k, v string) error {
	return setEnvVar(&r, k, v)
}

func (m MySQLEnv) Read(k string) string {
	return readEnvVar(&m, k)
}

func (r RethinkDBEnv) Read(k string) string {
	return readEnvVar(&r, k)
}

func readEnvVar(s SetterReader, k string) string {
	return fmt.Sprint(reflect.ValueOf(s).Elem().FieldByName(k))
}

// set is a helper function to set a field's value within a struct
func setEnvVar(s SetterReader, k, v string) error {
	original := fmt.Sprint(reflect.ValueOf(s).Elem().FieldByName(k))
	reflect.ValueOf(s).Elem().FieldByName(k).SetString(v)

	if original == fmt.Sprint(reflect.ValueOf(s).Elem().FieldByName(k)) {
		return fmt.Errorf("could not set the environment variable '%s'", k)
	}

	return nil
}

// Check verifies the necessary environment variables are defined
// and that dependencies are present
// func (c *Config) Check() (*Env, error) {
func (c *Config) check() (*Config, error) {
	missing := []string{}

	for _, v := range c.Required.Collection {
		val, isSet := os.LookupEnv(v.Join())

		if val == "" || !isSet {
			missing = append(missing, v.Join())
		}
	}

	if len(missing) != 0 {
		return c, errors.New(color.RedString(fmt.Sprintf("Please set missing environment variables: %s", missing)))
	}

	_, err := c.locateDependencies()

	if err != nil {
		return c, err
	}

	return c, nil
}

// New creates a new runtime configuration
func New() (*Config, error) {
	cfg := &Config{}
	cfg.NewDependency("MySQL", "mysql", "/usr/bin/mysql", "")
	cfg.NewDependency("PT Query Digest", "pt-query-digest", "/usr/bin/pt-query-digest", "")
	cfg.NewConnString("MYSQL_USER", "MYSQL_PASSWORD", "mysql", "localhost", os.Getenv("MYSQL_PORT"), "tcp")
	cfg.NewEnv(format.NewDelimitedCollection("MYSQL", "_", []string{"USER", "PASSWORD", "HOST", "PORT"}), &MySQLEnv{})
	cfg.NewEnv(format.NewDelimitedCollection("RDB", "_", []string{"ADDRESS", "DATABASE", "USERNAME", "PASSWORD"}), &RethinkDBEnv{})

	return cfg.check()
}

// status defines the status of a network connection
type status struct {
	conn   string
	port   string
	socket string
	errors []error
}

// checkMySQLConn checks for connectivity to external services with the ability to retry connections
func checkMySQLConn(d *sql.DB, totalRetries, remainingRetries int) (*status, error) {
	stat := status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("Checking MySQL Connection Status")

	defer func() {
		if r := recover(); r != nil {
			next := remainingRetries - 1
			fmt.Printf("could not connect to MySQL on the first attempt, attempting to reconnect:\n%d retry attempt(s) remaining\n\n", next+1)
			time.Sleep(15 * time.Second)
			if next < 0 {
				log.Fatal("could not connect to MySQL database")
			}
			checkMySQLConn(d, totalRetries, next)
		}
	}()

	err := d.Ping()

	if err != nil {
		stat.conn = red.Sprint(" CLOSED")
		stat.port = red.Sprint(" NOT FOUND")
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", d.Stats()))
		panic(fmt.Sprintf("could not connect to database on attempt #%d %v", remainingRetries, err))
	}

	sock, err := os.Stat(os.Getenv("MYSQL_SOCKET"))
	if err != nil {
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", err))
		stat.socket = red.Sprintf(" %s", err)
		return &stat, fmt.Errorf("could not locate the MySQL file\n%s", err)
	}

	stat.conn = green.Sprint(" OPEN")
	stat.port = green.Sprint(" 3306")
	stat.socket = green.Sprintf(" %s", sock.Name())

	return &stat, nil
}

// ConnectMySQL initializes the DB connection
func ConnectMySQL(c MySQLConn) (*sql.DB, error) {
	db, err := sql.Open(c.Driver, fmt.Sprintf("%s:%s@%s(%s:%v)/", c.Auth.Username, c.Auth.Password, c.Protocol, c.Host, c.Port))

	if err != nil {
		return db, fmt.Errorf("could not open database connection\n%s", err)
	}

	conn, err := checkMySQLConn(db, 10, 10)

	if err != nil {
		return db, fmt.Errorf("could not maintain database connection\n%s", err)
	}

	// print connection status
	fmt.Printf("  %+v\n\n", conn)

	return db, nil
}

// Check checks for connectivity to external services
func checkRethinkDBConn(d *r.Session) (*status, error) {
	stat := status{}
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	color.New(color.Bold).Println("RethinkDB Connection Status")

	server, err := d.Server()

	if err != nil || !d.IsConnected() {
		stat.conn = red.Sprint(" CLOSED")
		stat.port = red.Sprint(" NOT FOUND")
		stat.errors = append(stat.errors, fmt.Errorf("could not connect to DB\n%v", err))
		return &stat, err
	}

	stat.conn = green.Sprint(" OPEN")
	stat.port = green.Sprint(" 28015")
	stat.socket = green.Sprintf(" %s", server.Name)

	return &stat, nil
}

// ConnectRethinkDB starts Connect RethinkDB connection
func ConnectRethinkDB(c SetterReader) (*r.Session, error) {

	db, err := r.Connect(r.ConnectOpts{
		Address:  c.Read("Address"),
		Database: c.Read("Database"),
		Username: c.Read("Username"),
		Password: c.Read("Password"),
	})

	conn, err := checkRethinkDBConn(db)

	fmt.Printf("  %+v\n\n", conn)

	return db, err
}
