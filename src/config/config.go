package config

import (
	"database/sql"
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
	Connections  ConnectionStr
	Env          Env
}

// Dependency defines an application runtime dependency
type Dependency struct {
	Name    string
	ExeName string
	Path    string
	Source  string
}

// Env defines the host machine's environment variables
type Env struct {
	Context  string
	Username string
	Password string
	Host     string
	Port     string
}

// Auth defines authorization credentials
type Auth struct {
	Username string
	Password string
}

// ConnectionStr defines data source and the means of connecting to it
type ConnectionStr struct {
	Auth     Auth
	Driver   string
	Host     string
	Port     string
	Status   interface{}
	Protocol string
}

func newDelimitedCollection(prefix, delimiter string, suffixColl []string) format.DelimitedCollection {
	dsc := format.DelimitedCollection{Delimiter: delimiter}

	for _, suffix := range suffixColl {
		dsc.Collection = append(dsc.Collection, format.NewDelimitedString(prefix, delimiter, suffix))
	}

	return dsc
}

// NewConnString creates a connection string
func NewConnString(a Auth, args ...string) ConnectionStr {
	cn := ConnectionStr{}
	cn.Auth = a
	cn.Driver = args[0]
	cn.Host = args[1]
	cn.Port = args[2]
	cn.Protocol = args[3]

	return cn
}

// NewDependency creates a new runtime dependency
func NewDependency(args ...string) Dependency {
	return Dependency{Name: args[0], ExeName: args[1], Path: args[2], Source: args[3]}
}

// LocateDependencies locates required deps
func LocateDependencies(d []Dependency) error {
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

// set is helper function to set a field's value within a struct
func (e Env) set(k, v string) Env {
	reflect.ValueOf(&e).Elem().FieldByName(k).SetString(v)
	return e
}

// Check verifies the necessary environment variables are defined
func Check(c string) (Env, error) {
	env := Env{}
	required := newDelimitedCollection("MYSQL", "_", []string{"USERNAME", "PASSWORD", "HOST", "PORT"})
	missing := []string{}
	var err error

	for _, v := range required.Collection {
		val, isSet := os.LookupEnv(v.Join())

		if val != "" && isSet {
			env = env.set(format.SplitToTitlecase(1, v), val)
		} else {
			missing = append(missing, v.Join())
		}
	}

	if len(missing) != 0 {
		err = errors.New(color.RedString(fmt.Sprintf("Please set missing environment variables: %s", missing)))
	}

	return env, err
}

func verify(db *sql.DB, a, v string) bool {
	var varKey, varValue string

	fmt.Println(fmt.Sprintf("mysql> SHOW VARIABLES LIKE '%s'", v))
	row := db.QueryRow(fmt.Sprintf("SHOW VARIABLES LIKE '%s'", v))

	row.Scan(&varKey, &varValue)

	return varValue == a

}

// AddConfig adds a runtime depedency to the configuration
func (cfg *Config) AddConfig(ctx string, args ...string) {
	switch ctx {
	case "dependency":
		cfg.Dependencies = append(cfg.Dependencies, NewDependency(args...))
	case "connection":
		a := Auth{cfg.Env.Username, cfg.Env.Password}
		cfg.Connections = NewConnString(a, args...)
	default:
		return
	}

}
