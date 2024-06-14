package oracle

import "database/sql"

//go:generate mockery --name=OracleDB --outpkg=mocks
type OracleDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Close() error
}
