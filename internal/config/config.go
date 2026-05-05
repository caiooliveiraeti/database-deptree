package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"gopkg.in/yaml.v3"
)

type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// AnalyzerRun is one entry in the "run" section of deptree.yaml.
// Group and Name select the registered analyzer; the remaining keys are its flags.
type AnalyzerRun struct {
	Group  string          `yaml:"group"`
	Name   string          `yaml:"name"`
	Config analyzer.Config `yaml:"config"`
}

type fileConfig struct {
	Neo4j Neo4jConfig   `yaml:"neo4j"`
	Run   []AnalyzerRun `yaml:"run"`
}

// File is the fully loaded deptree.yaml.
type File struct {
	Neo4j Neo4jConfig
	Run   []AnalyzerRun
}

// Load reads deptree.yaml and applies env var overrides for Neo4j settings.
func Load(configFile string) (File, error) {
	fc := fileConfig{
		Neo4j: Neo4jConfig{
			URI:  "bolt://localhost:7687",
			User: "neo4j",
		},
	}

	data, err := os.ReadFile(configFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return File{}, fmt.Errorf("reading config file %s: %w", configFile, err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return File{}, fmt.Errorf("parsing config file %s: %w", configFile, err)
		}
	}

	// Env vars override file values
	if v := os.Getenv("NEO4J_URI"); v != "" {
		fc.Neo4j.URI = v
	}
	if v := os.Getenv("NEO4J_USER"); v != "" {
		fc.Neo4j.User = v
	}
	if v := os.Getenv("NEO4J_PASSWORD"); v != "" {
		fc.Neo4j.Password = v
	}

	if fc.Neo4j.URI == "" {
		return File{}, errors.New("neo4j URI is required (set NEO4J_URI or neo4j.uri in config file)")
	}

	return File{Neo4j: fc.Neo4j, Run: fc.Run}, nil
}
