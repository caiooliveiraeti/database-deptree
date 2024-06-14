package oracle

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/caiooliveiraeti/database-deptree/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeOracleTree(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlMock.NewRows([]string{"OWNER", "NAME", "TYPE", "REFERENCED_OWNER", "REFERENCED_NAME", "REFERENCED_TYPE"}).
		AddRow("owner", "name", "type", "ref_owner", "ref_name", "ref_type")

	sqlMock.ExpectQuery("SELECT OWNER, NAME, TYPE, REFERENCED_OWNER, REFERENCED_NAME, REFERENCED_TYPE FROM DBA_DEPENDENCIES WHERE OWNER = 'YOUR_SCHEMA'").WillReturnRows(rows)

	// Mock do OracleDB para retornar a conexão sqlmock
	mockDB := mocks.NewOracleDB(t)
	mockDB.On("Query", mock.Anything, mock.Anything).Return(db.Query("SELECT OWNER, NAME, TYPE, REFERENCED_OWNER, REFERENCED_NAME, REFERENCED_TYPE FROM DBA_DEPENDENCIES WHERE OWNER = 'YOUR_SCHEMA'"))
	mockDB.On("Close").Return(nil)

	oracleAnalyzer := OracleAnalyzer{
		User:     "user",
		Password: "password",
		DSN:      "dsn",
		DB:       mockDB,
	}

	deps, err := oracleAnalyzer.Analyze()
	require.NoError(t, err)
	require.NotEmpty(t, deps)

	sqlMock.ExpectationsWereMet()
}
