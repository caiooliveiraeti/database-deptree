package java

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/caiooliveiraeti/database-deptree/internal/graph"
)

var (
	reEntity      = regexp.MustCompile(`@Entity\s*(?:\(.*\))?`)
	reTableName   = regexp.MustCompile(`@Table\s*\([^)]*name\s*=\s*"([^"]*)"`)
	reTableSchema = regexp.MustCompile(`@Table\s*\([^)]*schema\s*=\s*"([^"]*)"`)
	reClass       = regexp.MustCompile(`public\s+(?:class|interface)\s+([a-zA-Z_][a-zA-Z_0-9]*)`)
	reRepository  = regexp.MustCompile(`extends\s+(?:JpaRepository|CrudRepository)<([a-zA-Z_][a-zA-Z_0-9]*),`)
	reQuery       = regexp.MustCompile(`@Query\("([^"]*)"\)`)
	reNamedQuery  = regexp.MustCompile(`@NamedQuery\s*\(\s*name\s*=\s*"[^"]*"\s*,\s*query\s*=\s*"([^"]*)"`)
	reProcedure   = regexp.MustCompile(`@Procedure\s*\(\s*(?:name\s*=\s*)?"([^"]*)"`)
	reFromTable   = regexp.MustCompile(`(?i)FROM\s+([a-zA-Z_][a-zA-Z_0-9]*)`)
	reCallProc    = regexp.MustCompile(`(?i)(?:CALL|EXECUTE)\s+([a-zA-Z_][a-zA-Z_0-9]*)`)

	// JPA join annotations followed (possibly with @JoinColumn etc.) by the field type.
	// Captures: (1) relation type  (2) collection generic e.g. List<Visit>  (3) plain type e.g. Owner
	reJoin = regexp.MustCompile(
		`@(ManyToOne|OneToMany|ManyToMany|OneToOne)` +
			`(?:\([^)]*\))?` +
			`(?:\s*@\w+(?:\([^)]*\))?)*` +
			`\s+(?:(?:private|protected|public|final)\s+)*` +
			`(?:(?:List|Set|Collection|Iterable)<([A-Z][a-zA-Z0-9]*)>|([A-Z][a-zA-Z0-9]*))`,
	)
)

// entityInfo holds what we know about a JPA entity after the first scan pass.
type entityInfo struct {
	className     string
	qualifiedTable string // schema-qualified table name, e.g. "my_schema.owners" or just "owners"
	content       string  // kept for join extraction in pass 2
}

type Analyzer struct {
	RootDir  string
	DBSchema string // default schema; overridden per-entity by @Table(schema="...")
}

func New(rootDir, dbSchema string) *Analyzer {
	return &Analyzer{RootDir: rootDir, DBSchema: dbSchema}
}

func (a *Analyzer) Name() string { return "java" }

func (a *Analyzer) Analyze(ctx context.Context) ([]graph.Edge, error) {
	// Pass 1 — read every .java file, collect raw content.
	type javaFile struct {
		path    string
		content string
	}
	var files []javaFile

	err := filepath.Walk(a.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".java" {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		content, err := readFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		files = append(files, javaFile{path, content})
		return nil
	})
	if err != nil {
		return nil, err
	}

	var edges []graph.Edge
	var entities []entityInfo

	// Pass 1 — entity/table edges + repositories.
	for _, f := range files {
		if reEntity.MatchString(f.content) {
			info := extractEntityInfo(f.content, a.DBSchema)
			entities = append(entities, info)
			src := graph.NewNode("ENTITY", info.className, nil)
			dst := graph.NewNode("TABLE", info.qualifiedTable, nil)
			edges = append(edges, graph.NewEdge(src, dst, "STORED_IN", nil))
		}
		if reRepository.MatchString(f.content) {
			edges = append(edges, extractRepositoryEdges(f.path, f.content)...)
		}
	}

	// Build className → qualifiedTable map for join resolution.
	classToTable := make(map[string]string, len(entities))
	for _, e := range entities {
		classToTable[e.className] = e.qualifiedTable
	}

	// Pass 2 — JPA join edges (TABLE → TABLE).
	for _, e := range entities {
		edges = append(edges, extractJoinEdges(e, classToTable)...)
	}

	return edges, nil
}

// extractEntityInfo parses a single entity file.
// defaultSchema is used when @Table has no schema attribute.
// @Table(schema="...") takes priority (more specific).
func extractEntityInfo(content, defaultSchema string) entityInfo {
	className := ""
	if m := reClass.FindStringSubmatch(content); m != nil {
		className = m[1]
	}

	tableName := className // default: class name == table name
	if m := reTableName.FindStringSubmatch(content); m != nil {
		tableName = m[1]
	}

	// Schema resolution: annotation wins over parameter
	schema := defaultSchema
	if m := reTableSchema.FindStringSubmatch(content); m != nil {
		schema = m[1]
	}

	qualifiedTable := tableName
	if schema != "" {
		qualifiedTable = schema + "." + tableName
	}

	return entityInfo{className: className, qualifiedTable: qualifiedTable, content: content}
}

// extractJoinEdges creates TABLE→TABLE edges from @ManyToOne/@OneToMany/@ManyToMany/@OneToOne.
// The relation type becomes the edge relationship (e.g. MANY_TO_ONE).
func extractJoinEdges(e entityInfo, classToTable map[string]string) []graph.Edge {
	srcTable, ok := classToTable[e.className]
	if !ok {
		srcTable = e.className
	}
	src := graph.NewNode("TABLE", srcTable, nil)

	var edges []graph.Edge
	for _, m := range reJoin.FindAllStringSubmatch(e.content, -1) {
		relType := toSnake(m[1])

		// m[2] = collection generic (List<Visit> → Visit), m[3] = plain type (Owner)
		targetClass := m[2]
		if targetClass == "" {
			targetClass = m[3]
		}
		if targetClass == "" {
			continue
		}

		targetTable, ok := classToTable[targetClass]
		if !ok {
			targetTable = targetClass // unresolved — store as-is
		}

		dst := graph.NewNode("TABLE", targetTable, nil)
		edges = append(edges, graph.NewEdge(src, dst, relType, map[string]any{
			"fromClass": e.className,
			"toClass":   targetClass,
		}))
	}
	return edges
}

// toSnake converts CamelCase to UPPER_SNAKE: ManyToOne → MANY_TO_ONE
func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func extractRepositoryEdges(path, content string) []graph.Edge {
	repoMatch := reRepository.FindStringSubmatch(content)
	if repoMatch == nil {
		return nil
	}
	entityClass := repoMatch[1]

	repoName := filepath.Base(path)
	repo := graph.NewNode("REPOSITORY", repoName, map[string]any{"file": path})
	entity := graph.NewNode("ENTITY", entityClass, nil)

	edges := []graph.Edge{graph.NewEdge(repo, entity, "MANAGES", nil)}

	for _, m := range reQuery.FindAllStringSubmatch(content, -1) {
		edges = append(edges, queryEdges(repo, m[1])...)
	}
	for _, m := range reNamedQuery.FindAllStringSubmatch(content, -1) {
		edges = append(edges, queryEdges(repo, m[1])...)
	}
	for _, m := range reProcedure.FindAllStringSubmatch(content, -1) {
		proc := graph.NewNode("PROCEDURE", m[1], nil)
		edges = append(edges, graph.NewEdge(repo, proc, "CALLS", nil))
	}

	return edges
}

func queryEdges(repo graph.Node, sql string) []graph.Edge {
	query := graph.NewNode("QUERY", sql, nil)
	edges := []graph.Edge{graph.NewEdge(repo, query, "QUERIES", nil)}

	for _, m := range reFromTable.FindAllStringSubmatch(sql, -1) {
		table := graph.NewNode("TABLE", m[1], nil)
		edges = append(edges, graph.NewEdge(query, table, "USES_TABLE", nil))
	}
	for _, m := range reCallProc.FindAllStringSubmatch(sql, -1) {
		proc := graph.NewNode("PROCEDURE", m[1], nil)
		edges = append(edges, graph.NewEdge(query, proc, "CALLS", nil))
	}

	return edges
}

func readFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		sb.WriteString(sc.Text())
		sb.WriteByte('\n')
	}
	return sb.String(), sc.Err()
}
