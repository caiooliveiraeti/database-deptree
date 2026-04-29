package main

import (
	"log/slog"
	"os"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"

	// Blank imports trigger init() registration for each analyzer.
	// To add a new analyzer: create its package and add a blank import here.
	_ "github.com/caiooliveiraeti/database-deptree/internal/analyzer/java"
	_ "github.com/caiooliveiraeti/database-deptree/internal/analyzer/oracle"

	"github.com/spf13/cobra"
)

var configFile string
var systemName string

func main() {
	setupLogging()

	root := &cobra.Command{
		Use:   "deptree",
		Short: "Map database/code dependency graphs into Neo4j",
		Long: `deptree analyzes dependencies from various sources and stores them as a graph in Neo4j.
Run it multiple times safely — it uses MERGE semantics, so nodes and edges are never duplicated.`,
	}

	root.PersistentFlags().StringVar(&configFile, "config", "deptree.yaml", "path to config file (Neo4j settings)")
	root.PersistentFlags().StringVar(&systemName, "system", "", "name of the system being analyzed (e.g. billing, petclinic)")
	root.PersistentFlags().Bool("dry-run", false, "print edges without writing to Neo4j")

	// Build group commands from the registry.
	// Each registered analyzer Factory contributes one leaf command under its group.
	groups := map[string]*cobra.Command{}
	for _, f := range analyzer.All() {
		if _, ok := groups[f.Group]; !ok {
			g := &cobra.Command{
				Use:   f.Group,
				Short: "Analyze " + f.Group + " sources",
			}
			groups[f.Group] = g
			root.AddCommand(g)
		}
		groups[f.Group].AddCommand(buildCommand(f))
	}

	root.AddCommand(newAllCmd())
	root.AddCommand(newServeCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildCommand turns a Factory into a cobra leaf command.
// Adding a new analyzer never requires touching this function.
func buildCommand(f analyzer.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   f.Name,
		Short: f.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := make(analyzer.Config)
			for _, flag := range f.Flags {
				v, _ := cmd.Flags().GetString(flag.Name)
				if flag.Required && v == "" {
					return &missingFlagError{flag: flag.Name}
				}
				cfg[flag.Name] = v
			}
			a, err := f.Build(cfg)
			if err != nil {
				return err
			}
			return runAnalyzer(cmd, a, f)
		},
	}
	for _, flag := range f.Flags {
		cmd.Flags().String(flag.Name, flag.Default, flag.Usage)
	}
	return cmd
}

func setupLogging() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}
