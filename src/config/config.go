package config

import (
	"errors"
	"fmt"
	"gopherDigest/src/format"
	"os"
	"reflect"

	"github.com/fatih/color"
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
