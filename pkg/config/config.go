package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Config defines main application configuration
type Config struct {
	Dependencies []Dependency
}

// Dependency defines an application runtime dependency
type Dependency struct {
	name    string
	exeName string
	path    string
	source  string
}

// Health defines the health of network connection
type Health struct {
	Conn   string
	Port   string
	Socket string
	Errors []error
}

// New creates a new runtime configuration
func New() (*Config, error) {
	cfg := &Config{}

	cfg.NewDependency("MySQL", "mysql", "/usr/bin/mysql", "")
	cfg.NewDependency("PT Query Digest", "pt-query-digest", "/usr/bin/pt-query-digest", "")

	return cfg.verifyDependencies()
}

// NewDependency creates a new runtime dependency
func (c *Config) NewDependency(args ...string) *Dependency {
	c.Dependencies = append(c.Dependencies, Dependency{name: args[0], exeName: args[1], path: args[2], source: args[3]})

	return &Dependency{name: args[0], exeName: args[1], path: args[2], source: args[3]}
}

// Verify verifies the necessary required runtime dependencies are present
func (c *Config) verifyDependencies() (*Config, error) {
	color.New(color.Bold).Println("\nLocating Dependencies")

	for _, dep := range c.Dependencies {
		_, err := os.Stat(dep.path)
		if err != nil {
			color.New(color.FgHiRed).Printf("  [x] Unable to locate %s executable at %s\n", dep.name, dep.path)
			return nil, fmt.Errorf("ERROR: Missing Required Dependency %s", err)
		}
		color.New(color.FgHiGreen).Printf("  [\u2713] Using %v executable from %s \n", dep.name, dep.path)
	}

	fmt.Println()

	return c, nil
}

// GetSecrets retrieves secret values given a retrieval function
func GetSecrets(retrieve func(string) string, prefix string, delimeter string, args ...string) []string {
	envs := []string{}

	for _, key := range args {
		envs = append(envs, retrieve(prefix+delimeter+key))
	}

	return envs
}
