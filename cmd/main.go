package main

import (
	"log"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/caiooliveiraeti/database-deptree/internal/analyzer/java"
	"github.com/caiooliveiraeti/database-deptree/internal/analyzer/oracle"
	"github.com/caiooliveiraeti/database-deptree/internal/config"
	"github.com/caiooliveiraeti/database-deptree/internal/database"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	// Carregar configurações
	cfg := config.LoadConfig()

	// Inicializar os analisadores
	javaAnalyzer := java.JavaAnalyzer{RootDir: cfg.JavaRootDir}
	oracleAnalyzer := oracle.OracleAnalyzer{
		User:     cfg.OracleUser,
		Password: cfg.OraclePassword,
		DSN:      cfg.OracleDSN,
	}

	analyzers := []analyzer.Analyzer{javaAnalyzer, oracleAnalyzer}

	// Conectar ao Neo4j
	driver, err := database.ConnectNeo4j(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword)
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	for _, analyzer := range analyzers {
		deps, err := analyzer.Analyze()
		if err != nil {
			log.Fatalf("Error analyzing: %v", err)
		}

		err = database.InsertDependencies(session, deps)
		if err != nil {
			log.Fatalf("Error inserting dependencies: %v", err)
		}
	}
}
