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
	Name    string
	ExeName string
	Path    string
	Source  string
}

// Health defines the health of network connection
type Health struct {
	conn   string
	port   string
	socket string
	errors []error
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
	c.Dependencies = append(c.Dependencies, Dependency{Name: args[0], ExeName: args[1], Path: args[2], Source: args[3]})

	return &Dependency{Name: args[0], ExeName: args[1], Path: args[2], Source: args[3]}
}

// Verify verifies the necessary required runtime dependencies are present
func (c *Config) verifyDependencies() (*Config, error) {
	color.New(color.Bold).Println("\nLocating Dependencies")

	for _, dep := range c.Dependencies {
		_, err := os.Stat(dep.Path)
		if err != nil {
			color.New(color.FgHiRed).Printf("  [x] Unable to locate %s executable at %s\n", dep.Name, dep.Path)
			return nil, fmt.Errorf("ERROR: Missing Required Dependency %s", err)
		}
		color.New(color.FgHiGreen).Printf("  [\u2713] Using %v executable from %s \n", dep.Name, dep.Path)
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
