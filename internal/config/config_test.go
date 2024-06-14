package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables
	os.Setenv("ORACLE_USER", "test_oracle_user")
	os.Setenv("ORACLE_PASSWORD", "test_oracle_password")
	os.Setenv("ORACLE_DSN", "test_oracle_dsn")
	os.Setenv("NEO4J_URI", "bolt://localhost:7687")
	os.Setenv("NEO4J_USER", "neo4j")
	os.Setenv("NEO4J_PASSWORD", "test")
	os.Setenv("JAVA_ROOT_DIR", "./java")

	cfg := LoadConfig()

	if cfg.OracleUser != "test_oracle_user" {
		t.Errorf("expected OracleUser to be 'test_oracle_user', got %s", cfg.OracleUser)
	}
	if cfg.OraclePassword != "test_oracle_password" {
		t.Errorf("expected OraclePassword to be 'test_oracle_password', got %s", cfg.OraclePassword)
	}
	if cfg.OracleDSN != "test_oracle_dsn" {
		t.Errorf("expected OracleDSN to be 'test_oracle_dsn', got %s", cfg.OracleDSN)
	}
	if cfg.Neo4jURI != "bolt://localhost:7687" {
		t.Errorf("expected Neo4jURI to be 'bolt://localhost:7687', got %s", cfg.Neo4jURI)
	}
	if cfg.Neo4jUser != "neo4j" {
		t.Errorf("expected Neo4jUser to be 'neo4j', got %s", cfg.Neo4jUser)
	}
	if cfg.Neo4jPassword != "test" {
		t.Errorf("expected Neo4jPassword to be 'test', got %s", cfg.Neo4jPassword)
	}
	if cfg.JavaRootDir != "./java" {
		t.Errorf("expected JavaRootDir to be './java', got %s", cfg.JavaRootDir)
	}
}
