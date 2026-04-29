package oracle

import "github.com/caiooliveiraeti/database-deptree/internal/analyzer"

func init() {
	analyzer.Register(analyzer.Factory{
		Group: "database",
		Name:  "oracle",
		Short: "Analyze Oracle DB objects and dependencies (DBA_DEPENDENCIES)",
		Flags: []analyzer.Flag{
			{Name: "user", Default: "", Usage: "Oracle DB user"},
			{Name: "password", Default: "", Usage: "Oracle DB password"},
			{Name: "dsn", Default: "", Usage: "Oracle DSN (e.g. localhost:1521/orcl)", Required: true},
			{Name: "schema", Default: "", Usage: "Oracle schema/owner to analyze (e.g. MY_SCHEMA)", Required: true},
		},
		Build: func(cfg analyzer.Config) (analyzer.Analyzer, error) {
			return New(cfg.Get("user"), cfg.Get("password"), cfg.Get("dsn"), cfg.Get("schema")), nil
		},
	})
}
