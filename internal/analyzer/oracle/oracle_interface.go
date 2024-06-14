package oracle

import "database/sql"

type OracleDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Close() error
}
