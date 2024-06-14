package database

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

//go:generate mockery --name=Neo4jDriver --outpkg=mocks
type Neo4jDriver interface {
	NewSession(config neo4j.SessionConfig) neo4j.Session
	Close() error
}

//go:generate mockery --name=Neo4jSession --outpkg=mocks
type Neo4jSession interface {
	BeginTransaction(configs ...func(*neo4j.TransactionConfig)) (neo4j.Transaction, error)
	Close() error
}

//go:generate mockery --name=Neo4jTransaction --outpkg=mocks
type Neo4jTransaction interface {
	Run(query string, params map[string]interface{}) (neo4j.Result, error)
	Commit() error
	Rollback() error
	Close() error
}
