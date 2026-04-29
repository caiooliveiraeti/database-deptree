package oracle

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name      string
		rows      [][]any
		wantEdges int
		wantErr   bool
	}{
		{
			name: "single dependency",
			rows: [][]any{
				{"MY_SCHEMA", "calculate_tax", "PROCEDURE", "MY_SCHEMA", "tax_table", "TABLE"},
			},
			wantEdges: 1,
		},
		{
			name: "multiple dependencies",
			rows: [][]any{
				{"MY_SCHEMA", "proc_a", "PROCEDURE", "MY_SCHEMA", "table_a", "TABLE"},
				{"MY_SCHEMA", "proc_b", "PROCEDURE", "MY_SCHEMA", "proc_a", "PROCEDURE"},
			},
			wantEdges: 2,
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
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
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
	assert.Equal(t, "DEPENDS_ON", e.Relationship)
}
