package oracle

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/caiooliveiraeti/database-deptree/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name      string
		rows      [][]any
		wantEdges int
		wantRels  []string
		wantErr   bool
	}{
		{
			name: "procedure uses table",
			rows: [][]any{
				{"MY_SCHEMA", "calculate_tax", "PROCEDURE", "MY_SCHEMA", "tax_table", "TABLE"},
			},
			wantEdges: 1,
			wantRels:  []string{"USES_TABLE"},
		},
		{
			name: "procedure calls procedure",
			rows: [][]any{
				{"MY_SCHEMA", "proc_b", "PROCEDURE", "MY_SCHEMA", "proc_a", "PROCEDURE"},
			},
			wantEdges: 1,
			wantRels:  []string{"CALLS"},
		},
		{
			name: "package contains procedure",
			rows: [][]any{
				{"MY_SCHEMA", "tax_pkg", "PACKAGE", "MY_SCHEMA", "calculate_tax", "PROCEDURE"},
			},
			wantEdges: 1,
			wantRels:  []string{"CONTAINS"},
		},
		{
			name: "view reads table",
			rows: [][]any{
				{"MY_SCHEMA", "tax_view", "VIEW", "MY_SCHEMA", "tax_table", "TABLE"},
			},
			wantEdges: 1,
			wantRels:  []string{"READS"},
		},
		{
			name: "trigger uses table",
			rows: [][]any{
				{"MY_SCHEMA", "trg_audit", "TRIGGER", "MY_SCHEMA", "tax_table", "TABLE"},
			},
			wantEdges: 1,
			wantRels:  []string{"USES_TABLE"},
		},
		{
			name: "fallback depends_on for structural types",
			rows: [][]any{
				{"MY_SCHEMA", "proc_a", "PROCEDURE", "MY_SCHEMA", "tax_type", "TYPE"},
			},
			wantEdges: 1,
			wantRels:  []string{"DEPENDS_ON"},
		},
		{
			name: "multiple dependencies mixed",
			rows: [][]any{
				{"MY_SCHEMA", "proc_a", "PROCEDURE", "MY_SCHEMA", "table_a", "TABLE"},
				{"MY_SCHEMA", "proc_b", "PROCEDURE", "MY_SCHEMA", "proc_a", "PROCEDURE"},
			},
			wantEdges: 2,
			wantRels:  []string{"USES_TABLE", "CALLS"},
		},
		{
			name:      "no rows",
			rows:      [][]any{},
			wantEdges: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			cols := []string{"OWNER", "NAME", "TYPE", "REFERENCED_OWNER", "REFERENCED_NAME", "REFERENCED_TYPE"}
			rows := mock.NewRows(cols)
			for _, r := range tt.rows {
				rows.AddRow(r[0], r[1], r[2], r[3], r[4], r[5])
			}
			mock.ExpectQuery("SELECT").WillReturnRows(rows)

			a := &Analyzer{Schema: "MY_SCHEMA", DB: db}
			edges, err := a.Analyze(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, edges, tt.wantEdges)
			for i, rel := range tt.wantRels {
				assert.Equal(t, rel, edges[i].Relationship)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAnalyze_MapsToView(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cols := []string{"OWNER", "NAME", "TYPE", "REFERENCED_OWNER", "REFERENCED_NAME", "REFERENCED_TYPE"}
	mock.ExpectQuery("SELECT").WillReturnRows(
		mock.NewRows(cols).AddRow("MY_SCHEMA", "proc_a", "PROCEDURE", "MY_SCHEMA", "owners", "VIEW"),
	)

	a := &Analyzer{Schema: "MY_SCHEMA", DB: db}
	edges, err := a.Analyze(context.Background())
	require.NoError(t, err)

	var mapsTo []graph.Edge
	for _, e := range edges {
		if e.Relationship == "MAPS_TO" {
			mapsTo = append(mapsTo, e)
		}
	}

	require.Len(t, mapsTo, 2, "should emit two MAPS_TO edges (unqualified and qualified TABLE)")

	targets := map[string]bool{}
	sources := map[string]bool{}
	for _, e := range mapsTo {
		assert.Equal(t, "VIEW:my_schema.owners", e.Target.ID)
		targets[e.Target.ID] = true
		sources[e.Source.ID] = true
	}
	assert.True(t, sources["TABLE:owners"], "unqualified TABLE bridge expected")
	assert.True(t, sources["TABLE:my_schema.owners"], "schema-qualified TABLE bridge expected")
}

func TestAnalyzeNodeIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cols := []string{"OWNER", "NAME", "TYPE", "REFERENCED_OWNER", "REFERENCED_NAME", "REFERENCED_TYPE"}
	mock.ExpectQuery("SELECT").WillReturnRows(
		mock.NewRows(cols).AddRow("MY_SCHEMA", "calculate_tax", "PROCEDURE", "MY_SCHEMA", "tax_table", "TABLE"),
	)

	a := &Analyzer{Schema: "MY_SCHEMA", DB: db}
	edges, err := a.Analyze(context.Background())
	require.NoError(t, err)
	require.Len(t, edges, 1)

	e := edges[0]
	// NodeID includes owner so objects with same name in different schemas don't collide
	assert.Equal(t, "PROCEDURE:my_schema.calculate_tax", e.Source.ID)
	assert.Equal(t, "TABLE:my_schema.tax_table", e.Target.ID)
	assert.Equal(t, "USES_TABLE", e.Relationship)
}
