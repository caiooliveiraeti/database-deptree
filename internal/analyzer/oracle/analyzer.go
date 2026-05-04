package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func oracleRel(srcType, dstType string) string {
	switch srcType {
	case "PACKAGE":
		switch dstType {
		case "PROCEDURE", "FUNCTION":
			return "CONTAINS"
		}
	case "PROCEDURE", "FUNCTION":
		switch dstType {
		case "PROCEDURE", "FUNCTION":
			return "CALLS"
		case "TABLE", "VIEW", "SYNONYM":
			return "USES_TABLE"
		}
	case "TRIGGER":
		switch dstType {
		case "TABLE", "VIEW", "SYNONYM":
			return "USES_TABLE"
		}
	case "VIEW":
		switch dstType {
		case "TABLE", "VIEW", "SYNONYM":
			return "READS"
		}
	}
	return "DEPENDS_ON"
}

// normalizeOracleType maps Oracle multi-word type names to valid Neo4j labels.
// "PACKAGE BODY" → "PACKAGE" (body and spec are the same logical object).
// Any remaining spaces are replaced with underscores as a safety net.
func normalizeOracleType(t string) string {
	if t == "PACKAGE BODY" {
		return "PACKAGE"
	}
	return strings.ReplaceAll(t, " ", "_")
}

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

	type viewEntry struct{ owner, name string }
	referencedViews := map[string]viewEntry{} // key = "owner.name" lowercased

	var edges []graph.Edge
	for rows.Next() {
		if ctx.Err() != nil {
			return edges, ctx.Err()
		}

		var owner, name, typ, refOwner, refName, refType string
		if err := rows.Scan(&owner, &name, &typ, &refOwner, &refName, &refType); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		typ    = normalizeOracleType(typ)
		refType = normalizeOracleType(refType)

		src := graph.NewNode(typ, owner+"."+name, map[string]any{"owner": owner, "shortName": name})
		dst := graph.NewNode(refType, refOwner+"."+refName, map[string]any{"owner": refOwner, "shortName": refName})
		edges = append(edges, graph.NewEdge(src, dst, oracleRel(typ, refType), nil))

		// Collect VIEWs that are actively referenced (as targets) within the analyzed schema.
		// These need a MAPS_TO bridge so Java TABLE nodes can reach the actual Oracle VIEW.
		if refType == "VIEW" && strings.EqualFold(refOwner, a.Schema) {
			key := strings.ToLower(refOwner + "." + refName)
			referencedViews[key] = viewEntry{refOwner, refName}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Emit MAPS_TO edges bridging the Java TABLE node(s) to the actual Oracle VIEW node.
	for _, v := range referencedViews {
		view := graph.NewNode("VIEW", v.owner+"."+v.name, map[string]any{"owner": v.owner, "shortName": v.name})
		// Unqualified: Java without --db-schema
		edges = append(edges, graph.NewEdge(graph.NewNode("TABLE", v.name, nil), view, "MAPS_TO", nil))
		// Schema-qualified: Java with --db-schema
		edges = append(edges, graph.NewEdge(graph.NewNode("TABLE", v.owner+"."+v.name, nil), view, "MAPS_TO", nil))
	}

	return edges, nil
}
