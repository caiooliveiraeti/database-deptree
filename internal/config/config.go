package config

import (
	"flag"
	"os"
)

type Config struct {
	OracleUser     string
	OraclePassword string
	OracleDSN      string
	Neo4jURI       string
	Neo4jUser      string
	Neo4jPassword  string
	JavaRootDir    string
}

func LoadConfig() *Config {
	// Definir flags para linha de comando
	oracleUserFlag := flag.String("oracleUser", "", "Oracle database user")
	oraclePasswordFlag := flag.String("oraclePassword", "", "Oracle database password")
	oracleDSNFlag := flag.String("oracleDSN", "", "Oracle DSN")
	neo4jURIFlag := flag.String("neo4jURI", "bolt://localhost:7687", "Neo4j URI")
	neo4jUserFlag := flag.String("neo4jUser", "neo4j", "Neo4j user")
	neo4jPasswordFlag := flag.String("neo4jPassword", "test", "Neo4j password")
	javaRootDirFlag := flag.String("javaRootDir", "./java", "Root directory for Java source files")

	// Parse flags
	flag.Parse()

	return &Config{
		OracleUser:     getEnvOrDefault("ORACLE_USER", *oracleUserFlag),
		OraclePassword: getEnvOrDefault("ORACLE_PASSWORD", *oraclePasswordFlag),
		OracleDSN:      getEnvOrDefault("ORACLE_DSN", *oracleDSNFlag),
		Neo4jURI:       getEnvOrDefault("NEO4J_URI", *neo4jURIFlag),
		Neo4jUser:      getEnvOrDefault("NEO4J_USER", *neo4jUserFlag),
		Neo4jPassword:  getEnvOrDefault("NEO4J_PASSWORD", *neo4jPasswordFlag),
		JavaRootDir:    getEnvOrDefault("JAVA_ROOT_DIR", *javaRootDirFlag),
	}
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão se a variável de ambiente não estiver definida
func getEnvOrDefault(envKey, defaultValue string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}
	return defaultValue
}
