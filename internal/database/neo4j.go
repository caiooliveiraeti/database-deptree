package database

import (
	"context"
	"errors"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ConnectNeo4j(uri, user, password string) (Neo4jDriver, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func InsertDependencies(ctx context.Context, session Neo4jSession, dependencies []analyzer.Dependency) error {
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	query := `
		MERGE (a:%s {name: $source})
		MERGE (b:%s {name: $target})
		MERGE (a)-[:%s]->(b)
	`
	for _, dep := range dependencies {
		_, err := tx.Run(
			ctx,
			query,
			map[string]interface{}{
				"source": dep.Source,
				"target": dep.Target,
			},
		)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return errors.Join(err, errRollback)
			}
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
