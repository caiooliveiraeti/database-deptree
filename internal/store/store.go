package store

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/caiooliveiraeti/database-deptree/internal/graph"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Store is the persistence interface for dependency graphs.
// Implementations must be idempotent (MERGE semantics — safe to call multiple times).
type Store interface {
	Save(ctx context.Context, edges []graph.Edge) error
	QueryGraph(ctx context.Context, filter GraphFilter) (GraphData, error)
	Close() error
}

// GraphFilter narrows which nodes and edges are returned by QueryGraph.
// Empty slices mean "include all".
type GraphFilter struct {
	Labels  []string
	Rels    []string
	Systems []string
}

// GraphData is the response payload for the web UI.
type GraphData struct {
	Nodes []NodeData `json:"nodes"`
	Edges []EdgeData `json:"edges"`
}

type NodeData struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Name   string `json:"name"`
	System string `json:"system,omitempty"`
}

type EdgeData struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Rel    string `json:"rel"`
}

var reValidIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`)

// Neo4jStore implements Store using Neo4j.
type Neo4jStore struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jStore(uri, user, password string) (*Neo4jStore, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("creating neo4j driver: %w", err)
	}
	return &Neo4jStore{driver: driver}, nil
}

func (s *Neo4jStore) Close() error {
	return s.driver.Close(context.Background())
}

func (s *Neo4jStore) Save(ctx context.Context, edges []graph.Edge) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, e := range edges {
			if err := mergeEdge(ctx, tx, e); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (s *Neo4jStore) QueryGraph(ctx context.Context, filter GraphFilter) (GraphData, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	cypher := `
		MATCH (n)-[r]->(m)
		WHERE (size($labels)  = 0 OR labels(n)[0] IN $labels OR labels(m)[0] IN $labels)
		  AND (size($rels)    = 0 OR type(r) IN $rels)
		  AND (size($systems) = 0 OR n.system IN $systems OR m.system IN $systems)
		RETURN n.id AS srcId, labels(n)[0] AS srcLabel, coalesce(n.name, n.id) AS srcName,
		       coalesce(n.system, '') AS srcSystem,
		       type(r) AS rel,
		       m.id AS dstId, labels(m)[0] AS dstLabel, coalesce(m.name, m.id) AS dstName,
		       coalesce(m.system, '') AS dstSystem
	`

	nilToEmpty := func(ss []string) []string {
		if ss == nil {
			return []string{}
		}
		return ss
	}
	params := map[string]any{
		"labels":  nilToEmpty(filter.Labels),
		"rels":    nilToEmpty(filter.Rels),
		"systems": nilToEmpty(filter.Systems),
	}

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return GraphData{}, fmt.Errorf("querying graph: %w", err)
	}

	nodes := map[string]NodeData{}
	var edges []EdgeData

	for result.Next(ctx) {
		rec := result.Record()
		srcID, _ := rec.Get("srcId")
		srcLabel, _ := rec.Get("srcLabel")
		srcName, _ := rec.Get("srcName")
		srcSystem, _ := rec.Get("srcSystem")
		dstID, _ := rec.Get("dstId")
		dstLabel, _ := rec.Get("dstLabel")
		dstName, _ := rec.Get("dstName")
		dstSystem, _ := rec.Get("dstSystem")
		rel, _ := rec.Get("rel")

		sid := str(srcID)
		did := str(dstID)

		if _, ok := nodes[sid]; !ok {
			nodes[sid] = NodeData{ID: sid, Label: str(srcLabel), Name: str(srcName), System: str(srcSystem)}
		}
		if _, ok := nodes[did]; !ok {
			nodes[did] = NodeData{ID: did, Label: str(dstLabel), Name: str(dstName), System: str(dstSystem)}
		}
		edges = append(edges, EdgeData{Source: sid, Target: did, Rel: str(rel)})
	}

	if err := result.Err(); err != nil {
		return GraphData{}, fmt.Errorf("iterating graph results: %w", err)
	}

	nodeSlice := make([]NodeData, 0, len(nodes))
	for _, n := range nodes {
		nodeSlice = append(nodeSlice, n)
	}
	return GraphData{Nodes: nodeSlice, Edges: edges}, nil
}

func mergeEdge(ctx context.Context, tx neo4j.ManagedTransaction, e graph.Edge) error {
	if err := validateIdentifier(e.Source.Label); err != nil {
		return fmt.Errorf("source label: %w", err)
	}
	if err := validateIdentifier(e.Target.Label); err != nil {
		return fmt.Errorf("target label: %w", err)
	}
	if err := validateIdentifier(e.Relationship); err != nil {
		return fmt.Errorf("relationship: %w", err)
	}

	cypher := fmt.Sprintf(`
		MERGE (a:%s {id: $srcID})
		SET a.name = $srcName, a += $srcProps
		MERGE (b:%s {id: $dstID})
		SET b.name = $dstName, b += $dstProps
		MERGE (a)-[:%s]->(b)
	`, e.Source.Label, e.Target.Label, e.Relationship)

	srcName, _ := e.Source.Properties["name"]
	dstName, _ := e.Target.Properties["name"]

	params := map[string]any{
		"srcID":    e.Source.ID,
		"srcName":  srcName,
		"srcProps": e.Source.Properties,
		"dstID":    e.Target.ID,
		"dstName":  dstName,
		"dstProps": e.Target.Properties,
	}

	_, err := tx.Run(ctx, cypher, params)
	if err != nil {
		slog.Error("failed to merge edge",
			"src", e.Source.ID, "dst", e.Target.ID, "rel", e.Relationship, "err", err)
		return fmt.Errorf("merging %s -[%s]-> %s: %w", e.Source.ID, e.Relationship, e.Target.ID, err)
	}
	return nil
}

func validateIdentifier(s string) error {
	if !reValidIdentifier.MatchString(s) {
		return fmt.Errorf("%q is not a valid Neo4j identifier", s)
	}
	return nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
