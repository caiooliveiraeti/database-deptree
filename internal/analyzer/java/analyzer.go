package java

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
)

type JavaAnalyzer struct {
	RootDir string
}

type Entity struct {
	File  string
	Class string
	Table string
}

type Repository struct {
	File        string
	EntityClass string
	Queries     []string
}

func (ja JavaAnalyzer) Analyze() ([]analyzer.Dependency, error) {
	var dependencies []analyzer.Dependency
	entities, repositories := analyzeJavaFiles(ja.RootDir)

	entityTableMap := make(map[string]string)
	for _, entity := range entities {
		entityTableMap[entity.Class] = entity.Table

		dependencies = append(dependencies, analyzer.Dependency{
			Source:       entity.Class,
			SourceLabel:  "Entity",
			Target:       entity.Table,
			TargetLabel:  "Table",
			Relationship: "STORED_IN",
		})
	}

	for _, repo := range repositories {
		dependencies = append(dependencies, analyzer.Dependency{
			Source:       repo.File,
			SourceLabel:  "Repository",
			Target:       repo.EntityClass,
			TargetLabel:  "Entity",
			Relationship: "MANAGES",
		})

		for _, query := range repo.Queries {
			dependencies = append(dependencies, analyzer.Dependency{
				Source:       repo.File,
				SourceLabel:  "Repository",
				Target:       query,
				TargetLabel:  "Query",
				Relationship: "QUERIES",
			})

			tables := extractTablesFromQuery(query)
			for _, table := range tables {
				dependencies = append(dependencies, analyzer.Dependency{
					Source:       query,
					SourceLabel:  "Query",
					Target:       table,
					TargetLabel:  "Table",
					Relationship: "USES_TABLE",
				})
			}
		}
	}
	return dependencies, nil
}

func analyzeJavaFiles(rootDir string) ([]Entity, []Repository) {
	var entities []Entity
	var repositories []Repository

	entityPattern := regexp.MustCompile(`@Entity\s*(?:\(.*\))?`)
	tablePattern := regexp.MustCompile(`@Table\s*\(\s*name\s*=\s*\"([^\"]*)\"\s*\)`)
	classPattern := regexp.MustCompile(`public\s+class\s+([a-zA-Z_][a-zA-Z_0-9]*)`)
	repositoryPattern := regexp.MustCompile(`extends\s+(JpaRepository|CrudRepository)<([a-zA-Z_][a-zA-Z_0-9]*),`)
	queryPattern := regexp.MustCompile(`@Query\("([^"]*)"\)`)
	namedQueryPattern := regexp.MustCompile(`@NamedQuery\s*\(\s*name\s*=\s*"([^"]*)"\s*,\s*query\s*=\s*"([^"]*)"\s*\)`)
	namedQueriesPattern := regexp.MustCompile(`@NamedQueries\s*\(\s*{\s*(?:@NamedQuery\s*\(\s*name\s*=\s*"([^"]*)"\s*,\s*query\s*=\s*"([^"]*)"\s*\)\s*,?\s*)+\s*}\s*\)`)

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".java" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var content string
			for scanner.Scan() {
				content += scanner.Text() + "\n"
			}

			if entityPattern.MatchString(content) {
				classMatch := classPattern.FindStringSubmatch(content)
				if classMatch != nil {
					className := classMatch[1]
					tableMatch := tablePattern.FindStringSubmatch(content)
					tableName := className
					if tableMatch != nil {
						tableName = tableMatch[1]
					}
					entities = append(entities, Entity{File: path, Class: className, Table: tableName})
				}
			}

			if repositoryPattern.MatchString(content) {
				repoMatch := repositoryPattern.FindStringSubmatch(content)
				if repoMatch != nil {
					entityClass := repoMatch[2]
					queries := queryPattern.FindAllStringSubmatch(content, -1)
					var queryList []string
					for _, query := range queries {
						queryList = append(queryList, query[1])
					}

					namedQueryMatches := namedQueryPattern.FindAllStringSubmatch(content, -1)
					for _, match := range namedQueryMatches {
						queryList = append(queryList, match[2])
					}

					namedQueriesMatches := namedQueriesPattern.FindAllStringSubmatch(content, -1)
					for _, match := range namedQueriesMatches {
						for i := 1; i < len(match); i += 2 {
							if match[i] != "" && match[i+1] != "" {
								queryList = append(queryList, match[i+1])
							}
						}
					}

					repositories = append(repositories, Repository{File: path, EntityClass: entityClass, Queries: queryList})
				}
			}
		}
		return nil
	})

	return entities, repositories
}

func extractTablesFromQuery(query string) []string {
	tablePattern := regexp.MustCompile(`FROM\s+([a-zA-Z_][a-zA-Z_0-9]*)`)
	matches := tablePattern.FindAllStringSubmatch(query, -1)
	var tables []string
	for _, match := range matches {
		tables = append(tables, match[1])
	}
	return tables
}
