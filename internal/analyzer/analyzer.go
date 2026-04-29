package analyzer

import (
	"context"

	"github.com/caiooliveiraeti/database-deptree/internal/graph"
)

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context) ([]graph.Edge, error)
}
