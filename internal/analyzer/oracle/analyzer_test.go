package oracle

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
)

func TestAnalyzeOracleTree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockOracleDB(ctrl)
	rows := sqlmock.NewRows([]string{"OWNER", "NAME", "TYPE", "REFERENCED_OWNER", "REFERENCED_NAME", "REFERENCED_TYPE"}).
		AddRow("owner", "name", "type", "ref_owner", "ref_name", "ref_type")

	mockDB.EXPECT().Query(gomock.Any(), gomock.Any()).Return(rows, nil).Times(1)

	oracleAnalyzer := OracleAnalyzer{
		User:     "user",
		Password: "password",
		DSN:      "dsn",
		DB:       mockDB,
	}

	deps, err := oracleAnalyzer.Analyze()
	if err != nil {
		t.Fatalf("Failed to analyze Oracle dependencies: %v", err)
	}

	if len(deps) == 0 {
		t.Fatalf("Expected dependencies to be found, got %d", len(deps))
	}
}
