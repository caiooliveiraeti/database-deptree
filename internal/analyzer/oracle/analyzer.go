package oracle

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/caiooliveiraeti/database-deptree/internal/graph"
	_ "github.com/sijms/go-ora/v2" // registers the "oracle" driver
)

const depsQuery = `
	SELECT OWNER, NAME, TYPE, REFERENCED_OWNER, REFERENCED_NAME, REFERENCED_TYPE
	FROM DBA_DEPENDENCIES
	WHERE OWNER = :schema
`

type Analyzer struct {
	User     string
	Password string
	DSN      string
	Schema   string
	DB       OracleDB
}

func New(user, password, dsn, schema string) *Analyzer {
	return &Analyzer{
		User:     user,
		Password: password,
		DSN:      dsn,
		Schema:   schema,
	}
}

func (a *Analyzer) Name() string { return "oracle" }

func (a *Analyzer) Analyze(ctx context.Context) ([]graph.Edge, error) {
	db := a.DB
	if db == nil {
		conn, err := sql.Open("oracle", fmt.Sprintf("oracle://%s:%s@%s", a.User, a.Password, a.DSN))
		if err != nil {
			return nil, fmt.Errorf("opening oracle connection: %w", err)
		}
		db = conn
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, depsQuery, sql.Named("schema", a.Schema))
	if err != nil {
		return nil, fmt.Errorf("querying DBA_DEPENDENCIES: %w", err)
	}
	defer rows.Close()

	var edges []graph.Edge
	for rows.Next() {
		if ctx.Err() != nil {
			return edges, ctx.Err()
		}

		var owner, name, typ, refOwner, refName, refType string
		if err := rows.Scan(&owner, &name, &typ, &refOwner, &refName, &refType); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		src := graph.NewNode(typ, owner+"."+name, map[string]any{"owner": owner, "shortName": name})
		dst := graph.NewNode(refType, refOwner+"."+refName, map[string]any{"owner": refOwner, "shortName": refName})
		edges = append(edges, graph.NewEdge(src, dst, "DEPENDS_ON", nil))
	}

	return edges, rows.Err()
}
