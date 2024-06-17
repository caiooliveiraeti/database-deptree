package database

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jDriver interface {
	NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext

	Close(ctx context.Context) error
}

type Neo4jSession interface {
	BeginTransaction(ctx context.Context, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ExplicitTransaction, error)
	Close(ctx context.Context) error
}
