package main

import (
	"log/slog"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/caiooliveiraeti/database-deptree/internal/config"
	"github.com/spf13/cobra"
)

func newAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run all analyzers configured in the config file",
		Long: `Reads the "run" section of deptree.yaml and executes each listed analyzer.

Example deptree.yaml:
  neo4j:
    uri: bolt://localhost:7687
    user: neo4j
    password: secret

  run:
    - group: files
      name: java
      config:
        root-dir: ./src/main/java

    - group: database
      name: oracle
      config:
        dsn: localhost:1521/orcl
        schema: MY_SCHEMA
        user: sys
        password: oracle`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}
			if len(cfg.Run) == 0 {
				slog.Warn("no analyzers configured under 'run' in config file", "file", configFile)
				return nil
			}

			var lastErr error
			for _, run := range cfg.Run {
				f, ok := analyzer.Lookup(run.Group, run.Name)
				if !ok {
					slog.Error("unknown analyzer", "group", run.Group, "name", run.Name)
					continue
				}
				a, err := f.Build(run.Config)
				if err != nil {
					slog.Error("failed to build analyzer", "group", run.Group, "name", run.Name, "err", err)
					lastErr = err
					continue
				}
				if err := runAnalyzer(cmd, a, f); err != nil {
					slog.Error("analyzer failed", "group", run.Group, "name", run.Name, "err", err)
					lastErr = err
				}
			}
			return lastErr
		},
	}
}
