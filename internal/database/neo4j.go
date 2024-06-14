package database

import (
	"fmt"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ConnectNeo4j(uri, user, password string) (Neo4jDriver, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func InsertDependencies(session Neo4jSession, dependencies []analyzer.Dependency) error {
	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}

	for _, dep := range dependencies {
		query := `
            MERGE (a:%s {name: $source})
            MERGE (b:%s {name: $target})
            MERGE (a)-[:%s]->(b)
        `
		_, err := tx.Run(
			fmt.Sprintf(query, dep.SourceLabel, dep.TargetLabel, dep.Relationship),
			map[string]interface{}{
				"source": dep.Source,
				"target": dep.Target,
			},
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
