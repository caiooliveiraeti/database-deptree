package java

import "github.com/caiooliveiraeti/database-deptree/internal/analyzer"

func init() {
	analyzer.Register(analyzer.Factory{
		Group:    "files",
		Name:     "java",
		Short:    "Analyze Java/Spring source code (JPA entities, repositories, @Query, @Procedure)",
		IsApp:    true,
		AppLinks: []string{"ENTITY"},
		Flags: []analyzer.Flag{
			{Name: "root-dir", Default: ".", Usage: "root directory containing Java source files"},
			{Name: "db-schema", Default: "", Usage: "default DB schema to qualify table names (overridden by @Table(schema=...))"},
		},
		Build: func(cfg analyzer.Config) (analyzer.Analyzer, error) {
			return New(cfg.Get("root-dir"), cfg.Get("db-schema")), nil
		},
	})
}
