package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_EnvVars(t *testing.T) {
	t.Setenv("NEO4J_URI", "bolt://testhost:7687")
	t.Setenv("NEO4J_USER", "testuser")
	t.Setenv("NEO4J_PASSWORD", "testpass")

	cfg, err := Load("nonexistent.yaml")
	require.NoError(t, err)

	assert.Equal(t, "bolt://testhost:7687", cfg.Neo4j.URI)
	assert.Equal(t, "testuser", cfg.Neo4j.User)
	assert.Equal(t, "testpass", cfg.Neo4j.Password)
}

func TestLoad_YAMLFile(t *testing.T) {
	content := `
neo4j:
  uri: bolt://yamlhost:7687
  user: yamluser
  password: yamlpass
`
	cfgFile := filepath.Join(t.TempDir(), "deptree.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := Load(cfgFile)
	require.NoError(t, err)

	assert.Equal(t, "bolt://yamlhost:7687", cfg.Neo4j.URI)
	assert.Equal(t, "yamluser", cfg.Neo4j.User)
	assert.Equal(t, "yamlpass", cfg.Neo4j.Password)
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	content := `
neo4j:
  uri: bolt://yamlhost:7687
  user: yamluser
  password: yamlpass
`
	cfgFile := filepath.Join(t.TempDir(), "deptree.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	t.Setenv("NEO4J_URI", "bolt://envhost:7687")

	cfg, err := Load(cfgFile)
	require.NoError(t, err)

	assert.Equal(t, "bolt://envhost:7687", cfg.Neo4j.URI, "env var should override YAML")
	assert.Equal(t, "yamluser", cfg.Neo4j.User, "YAML value should be kept when no env var")
}

func TestLoad_RunSection(t *testing.T) {
	content := `
neo4j:
  uri: bolt://localhost:7687
run:
  - group: files
    name: java
    config:
      root-dir: ./src
  - group: database
    name: oracle
    config:
      dsn: localhost:1521/orcl
      schema: MY_SCHEMA
`
	cfgFile := filepath.Join(t.TempDir(), "deptree.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := Load(cfgFile)
	require.NoError(t, err)

	require.Len(t, cfg.Run, 2)
	assert.Equal(t, "files", cfg.Run[0].Group)
	assert.Equal(t, "java", cfg.Run[0].Name)
	assert.Equal(t, "./src", cfg.Run[0].Config.Get("root-dir"))
	assert.Equal(t, "MY_SCHEMA", cfg.Run[1].Config.Get("schema"))
}
