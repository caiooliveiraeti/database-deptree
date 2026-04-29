package oracle

import (
	"context"
	"database/sql"
)

//go:generate mockery --name=OracleDB --outpkg=mocks
type OracleDB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Close() error
}
