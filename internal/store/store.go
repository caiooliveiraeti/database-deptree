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
	QueryTraversal(ctx context.Context, q TraversalQuery) (GraphData, error)
	SearchNodes(ctx context.Context, term string) ([]NodeData, error)
	GetNode(ctx context.Context, id string) (NodeDetail, error)
	QueryInsights(ctx context.Context) (InsightsData, error)
	QueryImpactedSystems(ctx context.Context, nodeID string) ([]NodeData, error)
	Close() error
}

// TraversalQuery defines a depth-limited graph traversal from a starting node.
type TraversalQuery struct {
	StartNodeID string
	Depth       int    // 1..5
	Direction   string // "outgoing" | "incoming" | "both"
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

type NodeDetail struct {
	ID         string         `json:"id"`
	Label      string         `json:"label"`
	Name       string         `json:"name"`
	System     string         `json:"system,omitempty"`
	Properties map[string]any `json:"properties"`
}

type EdgeData struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Rel    string `json:"rel"`
}

// InsightNode is a single result row returned by an insight query.
type InsightNode struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	Name  string   `json:"name"`
	Count int      `json:"count,omitempty"` // in-degree, procedure count, call depth, etc.
	Tags  []string `json:"tags,omitempty"`  // related names (systems, entities, procedures)
}

// InsightsData groups all pre-computed analytical queries for the web UI panel.
type InsightsData struct {
	// Entendimento do sistema
	OrphanEntities []InsightNode `json:"orphan_entities"`
	NativeQueries  []InsightNode `json:"native_queries"`
	// Impacto de mudança
	DualAccess  []InsightNode `json:"dual_access"`
	MappedViews []InsightNode `json:"mapped_views"`
	// Refatoração
	SingleEntityTables []InsightNode `json:"single_entity_tables"`
	SingleCallerProcs  []InsightNode `json:"single_caller_procs"`
	UnusedProcs        []InsightNode `json:"unused_procs"`
	// Modernização / migração
	Hotspots       []InsightNode `json:"hotspots"`
	SharedTables   []InsightNode `json:"shared_tables"`
	HeavyPackages  []InsightNode `json:"heavy_packages"`
	DeepCallChains []InsightNode `json:"deep_call_chains"`
	UncoveredProcs []InsightNode `json:"uncovered_procs"`
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

func (s *Neo4jStore) QueryTraversal(ctx context.Context, q TraversalQuery) (GraphData, error) {
	if q.Depth < 1 {
		q.Depth = 1
	}
	if q.Depth > 5 {
		q.Depth = 5
	}

	var matchPattern string
	switch q.Direction {
	case "incoming":
		matchPattern = "<-[*1..%d]-"
	case "both":
		matchPattern = "-[*1..%d]-"
	default: // outgoing
		matchPattern = "-[*1..%d]->"
	}
	traverseRel := fmt.Sprintf(matchPattern, q.Depth)

	cypher := fmt.Sprintf(`
		MATCH (start {id: $startId})
		OPTIONAL MATCH (start)%s(n)
		WITH start, collect(DISTINCT n) AS reached
		WITH reached + [start] AS allNodes
		UNWIND allNodes AS node
		WITH collect(DISTINCT {
		  id: node.id, label: labels(node)[0],
		  name: coalesce(node.name, node.id),
		  system: coalesce(node.system, '')
		}) AS nodes, allNodes
		UNWIND allNodes AS src
		OPTIONAL MATCH (src)-[r]->(dst) WHERE dst IN allNodes
		RETURN nodes, collect(DISTINCT {source: src.id, target: dst.id, rel: type(r)}) AS edges
	`, traverseRel)

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, cypher, map[string]any{"startId": q.StartNodeID})
	if err != nil {
		return GraphData{}, fmt.Errorf("traversal query: %w", err)
	}

	if !result.Next(ctx) {
		return GraphData{Nodes: []NodeData{}, Edges: []EdgeData{}}, result.Err()
	}

	rec := result.Record()
	rawNodes, _ := rec.Get("nodes")
	rawEdges, _ := rec.Get("edges")

	return parseTraversalResult(rawNodes, rawEdges), nil
}

func parseTraversalResult(rawNodes, rawEdges any) GraphData {
	var nodes []NodeData
	if ns, ok := rawNodes.([]any); ok {
		for _, n := range ns {
			if m, ok := n.(map[string]any); ok {
				nodes = append(nodes, NodeData{
					ID:     str(m["id"]),
					Label:  str(m["label"]),
					Name:   str(m["name"]),
					System: str(m["system"]),
				})
			}
		}
	}

	var edges []EdgeData
	if es, ok := rawEdges.([]any); ok {
		for _, e := range es {
			if m, ok := e.(map[string]any); ok {
				src, tgt, rel := str(m["source"]), str(m["target"]), str(m["rel"])
				if src == "" || tgt == "" || rel == "" {
					continue // OPTIONAL MATCH returned null relationship
				}
				edges = append(edges, EdgeData{Source: src, Target: tgt, Rel: rel})
			}
		}
	}

	if nodes == nil {
		nodes = []NodeData{}
	}
	if edges == nil {
		edges = []EdgeData{}
	}
	return GraphData{Nodes: nodes, Edges: edges}
}

func (s *Neo4jStore) SearchNodes(ctx context.Context, term string) ([]NodeData, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	cypher := `
		MATCH (n)
		WHERE toLower(n.name) CONTAINS toLower($term)
		   OR toLower(n.id)   CONTAINS toLower($term)
		RETURN n.id AS id, labels(n)[0] AS label,
		       coalesce(n.name, n.id) AS name,
		       coalesce(n.system, '') AS system
		ORDER BY n.name
		LIMIT 20
	`

	result, err := session.Run(ctx, cypher, map[string]any{"term": term})
	if err != nil {
		return nil, fmt.Errorf("searching nodes: %w", err)
	}

	var nodes []NodeData
	for result.Next(ctx) {
		rec := result.Record()
		id, _ := rec.Get("id")
		label, _ := rec.Get("label")
		name, _ := rec.Get("name")
		system, _ := rec.Get("system")
		nodes = append(nodes, NodeData{
			ID:     str(id),
			Label:  str(label),
			Name:   str(name),
			System: str(system),
		})
	}
	return nodes, result.Err()
}

func (s *Neo4jStore) GetNode(ctx context.Context, id string) (NodeDetail, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (n {id: $id}) RETURN labels(n)[0] AS label, properties(n) AS props LIMIT 1`,
		map[string]any{"id": id},
	)
	if err != nil {
		return NodeDetail{}, fmt.Errorf("get node query: %w", err)
	}
	if !result.Next(ctx) {
		if e := result.Err(); e != nil {
			return NodeDetail{}, fmt.Errorf("get node result: %w", e)
		}
		return NodeDetail{}, fmt.Errorf("node not found: %s", id)
	}

	rec := result.Record()
	label, _ := rec.Get("label")
	rawProps, _ := rec.Get("props")

	props, _ := rawProps.(map[string]any)
	if props == nil {
		props = map[string]any{}
	}

	name := str(props["name"])
	if name == "" {
		name = id
	}

	skip := map[string]bool{"id": true, "name": true, "system": true}
	filtered := make(map[string]any, len(props))
	for k, v := range props {
		if !skip[k] {
			filtered[k] = v
		}
	}

	return NodeDetail{
		ID:         id,
		Label:      str(label),
		Name:       name,
		System:     str(props["system"]),
		Properties: filtered,
	}, result.Err()
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

func (s *Neo4jStore) QueryImpactedSystems(ctx context.Context, nodeID string) ([]NodeData, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (app:APPLICATION)-[*1..10]->(n {id: $id})
		RETURN DISTINCT app.id AS id, 'APPLICATION' AS label,
		       coalesce(app.name, app.id) AS name,
		       coalesce(app.system, '') AS system
		ORDER BY name
	`, map[string]any{"id": nodeID})
	if err != nil {
		return nil, fmt.Errorf("impacted systems query: %w", err)
	}

	var nodes []NodeData
	for result.Next(ctx) {
		rec := result.Record()
		id, _ := rec.Get("id")
		label, _ := rec.Get("label")
		name, _ := rec.Get("name")
		system, _ := rec.Get("system")
		nodes = append(nodes, NodeData{
			ID:     str(id),
			Label:  str(label),
			Name:   str(name),
			System: str(system),
		})
	}
	if nodes == nil {
		nodes = []NodeData{}
	}
	return nodes, result.Err()
}

func (s *Neo4jStore) QueryInsights(ctx context.Context) (InsightsData, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	run := func(cypher string, params map[string]any) ([]map[string]any, error) {
		res, err := session.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		var rows []map[string]any
		for res.Next(ctx) {
			rec := res.Record()
			row := make(map[string]any, len(rec.Keys))
			for _, k := range rec.Keys {
				v, _ := rec.Get(k)
				row[k] = v
			}
			rows = append(rows, row)
		}
		return rows, res.Err()
	}

	toNode := func(row map[string]any) InsightNode {
		return InsightNode{
			ID:    str(row["id"]),
			Label: str(row["label"]),
			Name:  str(row["name"]),
			Count: int(toInt(row["count"])),
			Tags:  toStrSlice(row["tags"]),
		}
	}

	mapRows := func(rows []map[string]any) []InsightNode {
		out := make([]InsightNode, 0, len(rows))
		for _, r := range rows {
			out = append(out, toNode(r))
		}
		return out
	}

	var data InsightsData
	var err error
	var rows []map[string]any

	// ── Entendimento ──────────────────────────────────────────────────────────

	rows, err = run(`
		MATCH (e:ENTITY) WHERE NOT (e)<-[:MANAGES]-()
		RETURN e.id AS id, 'ENTITY' AS label, coalesce(e.name, e.id) AS name`, nil)
	if err != nil {
		return data, fmt.Errorf("orphan_entities: %w", err)
	}
	data.OrphanEntities = mapRows(rows)

	rows, err = run(`
		MATCH (r:REPOSITORY)-[:QUERIES]->(q:QUERY)-[:USES_TABLE]->(t)
		WHERE q.native = true
		RETURN DISTINCT q.id AS id, 'QUERY' AS label, coalesce(q.name, q.id) AS name,
		       collect(DISTINCT coalesce(t.name, t.id)) AS tags`, nil)
	if err != nil {
		return data, fmt.Errorf("native_queries: %w", err)
	}
	data.NativeQueries = mapRows(rows)

	// ── Impacto ───────────────────────────────────────────────────────────────

	rows, err = run(`
		MATCH (e:ENTITY)-[:STORED_IN]->(t:TABLE)<-[:USES_TABLE]-(p:PROCEDURE)
		RETURN DISTINCT t.id AS id, 'TABLE' AS label, coalesce(t.name, t.id) AS name,
		       collect(DISTINCT coalesce(e.name, e.id)) + collect(DISTINCT coalesce(p.name, p.id)) AS tags`, nil)
	if err != nil {
		return data, fmt.Errorf("dual_access: %w", err)
	}
	data.DualAccess = mapRows(rows)

	rows, err = run(`
		MATCH (t:TABLE)-[:MAPS_TO]->(v:VIEW)
		RETURN v.id AS id, 'VIEW' AS label, coalesce(v.name, v.id) AS name,
		       [coalesce(t.name, t.id)] AS tags`, nil)
	if err != nil {
		return data, fmt.Errorf("mapped_views: %w", err)
	}
	data.MappedViews = mapRows(rows)

	// ── Refatoração ───────────────────────────────────────────────────────────

	rows, err = run(`
		MATCH (e:ENTITY)-[:STORED_IN]->(t:TABLE)
		WITH t, collect(DISTINCT coalesce(e.name, e.id)) AS ents
		WHERE size(ents) = 1
		RETURN t.id AS id, 'TABLE' AS label, coalesce(t.name, t.id) AS name, ents AS tags`, nil)
	if err != nil {
		return data, fmt.Errorf("single_entity_tables: %w", err)
	}
	data.SingleEntityTables = mapRows(rows)

	rows, err = run(`
		MATCH (caller)-[:CALLS]->(p:PROCEDURE)
		WITH p, count(caller) AS c
		WHERE c = 1
		RETURN p.id AS id, 'PROCEDURE' AS label, coalesce(p.name, p.id) AS name, c AS count`, nil)
	if err != nil {
		return data, fmt.Errorf("single_caller_procs: %w", err)
	}
	data.SingleCallerProcs = mapRows(rows)

	rows, err = run(`
		MATCH (p:PROCEDURE) WHERE NOT ()-[:CALLS]->(p)
		RETURN p.id AS id, 'PROCEDURE' AS label, coalesce(p.name, p.id) AS name`, nil)
	if err != nil {
		return data, fmt.Errorf("unused_procs: %w", err)
	}
	data.UnusedProcs = mapRows(rows)

	// ── Modernização ──────────────────────────────────────────────────────────

	rows, err = run(`
		MATCH (n)<-[r]-()
		WITH n, count(r) AS d
		WHERE d > 0
		RETURN n.id AS id, labels(n)[0] AS label, coalesce(n.name, n.id) AS name, d AS count
		ORDER BY d DESC LIMIT 10`, nil)
	if err != nil {
		return data, fmt.Errorf("hotspots: %w", err)
	}
	data.Hotspots = mapRows(rows)

	rows, err = run(`
		MATCH (app:APPLICATION)-[*1..5]->(t:TABLE)
		WITH t, collect(DISTINCT coalesce(app.name, app.id)) AS apps
		WHERE size(apps) > 1
		RETURN t.id AS id, 'TABLE' AS label, coalesce(t.name, t.id) AS name,
		       size(apps) AS count, apps AS tags
		ORDER BY size(apps) DESC`, nil)
	if err != nil {
		return data, fmt.Errorf("shared_tables: %w", err)
	}
	data.SharedTables = mapRows(rows)

	rows, err = run(`
		MATCH (pkg:PACKAGE)-[:CONTAINS]->(p)
		WITH pkg, count(p) AS n
		RETURN pkg.id AS id, 'PACKAGE' AS label, coalesce(pkg.name, pkg.id) AS name, n AS count
		ORDER BY n DESC LIMIT 10`, nil)
	if err != nil {
		return data, fmt.Errorf("heavy_packages: %w", err)
	}
	data.HeavyPackages = mapRows(rows)

	rows, err = run(`
		MATCH path=(a:PROCEDURE)-[:CALLS*]->(b:PROCEDURE)
		WITH a, max(length(path)) AS depth
		RETURN a.id AS id, 'PROCEDURE' AS label, coalesce(a.name, a.id) AS name, depth AS count
		ORDER BY depth DESC LIMIT 10`, nil)
	if err != nil {
		return data, fmt.Errorf("deep_call_chains: %w", err)
	}
	data.DeepCallChains = mapRows(rows)

	rows, err = run(`
		MATCH (p:PROCEDURE)
		WHERE NOT ()-[:CALLS]->(p) AND NOT (p)<-[:CONTAINS]-()
		RETURN p.id AS id, 'PROCEDURE' AS label, coalesce(p.name, p.id) AS name`, nil)
	if err != nil {
		return data, fmt.Errorf("uncovered_procs: %w", err)
	}
	data.UncoveredProcs = mapRows(rows)

	return data, nil
}

func toInt(v any) int64 {
	if v == nil {
		return 0
	}
	n, _ := v.(int64)
	return n
}

func toStrSlice(v any) []string {
	if v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func str(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
