package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/caiooliveiraeti/database-deptree/internal/config"
	"github.com/caiooliveiraeti/database-deptree/internal/graph"
	"github.com/caiooliveiraeti/database-deptree/internal/store"
	"github.com/spf13/cobra"
)

func runAnalyzer(cmd *cobra.Command, a analyzer.Analyzer, f analyzer.Factory) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		return runDry(cmd.Context(), a, f)
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	s, err := store.NewNeo4jStore(cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		return fmt.Errorf("connecting to neo4j: %w", err)
	}
	defer s.Close()

	return analyze(cmd.Context(), a, f, s)
}

func analyze(ctx context.Context, a analyzer.Analyzer, f analyzer.Factory, s store.Store) error {
	slog.Info("starting analyzer", "analyzer", a.Name(), "system", systemName)

	edges, err := a.Analyze(ctx)
	if err != nil {
		return fmt.Errorf("analyzing with %s: %w", a.Name(), err)
	}

	if systemName != "" {
		edges = tagSystem(edges, systemName, f.IsApp, f.AppLinks)
	}

	slog.Info("analysis complete", "analyzer", a.Name(), "edges", len(edges))

	if err := s.Save(ctx, edges); err != nil {
		return fmt.Errorf("saving %s results: %w", a.Name(), err)
	}

	slog.Info("saved to neo4j", "analyzer", a.Name(), "edges", len(edges))
	return nil
}

// tagSystem stamps every node with the system name.
// When isApp is true, also creates APPLICATION:system -[CONTAINS]-> nodes for each appLink label.
func tagSystem(edges []graph.Edge, system string, isApp bool, appLinks []string) []graph.Edge {
	for i := range edges {
		edges[i].Source.Properties["system"] = system
		edges[i].Target.Properties["system"] = system
	}

	if !isApp || len(appLinks) == 0 || system == "" {
		return edges
	}

	linkSet := make(map[string]bool, len(appLinks))
	for _, l := range appLinks {
		linkSet[l] = true
	}

	app := graph.NewNode("APPLICATION", system, nil)
	seen := map[string]bool{}
	var appEdges []graph.Edge

	for _, e := range edges {
		if linkSet[e.Source.Label] && !seen[e.Source.ID] {
			seen[e.Source.ID] = true
			appEdges = append(appEdges, graph.NewEdge(app, e.Source, "CONTAINS", nil))
		}
	}

	return append(edges, appEdges...)
}

func runDry(ctx context.Context, a analyzer.Analyzer, f analyzer.Factory) error {
	slog.Info("dry-run: starting analyzer", "analyzer", a.Name(), "system", systemName)

	edges, err := a.Analyze(ctx)
	if err != nil {
		return fmt.Errorf("analyzing with %s: %w", a.Name(), err)
	}

	if systemName != "" {
		edges = tagSystem(edges, systemName, f.IsApp, f.AppLinks)
	}

	for _, e := range edges {
		slog.Info("edge",
			"src_id", e.Source.ID,
			"rel", e.Relationship,
			"dst_id", e.Target.ID,
		)
	}
	slog.Info("dry-run complete", "analyzer", a.Name(), "edges", len(edges))
	return nil
}
