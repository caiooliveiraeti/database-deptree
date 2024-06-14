package database

import (
	"testing"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/caiooliveiraeti/database-deptree/test/mocks"
	"github.com/stretchr/testify/mock"
)

func TestInsertDependencies(t *testing.T) {
	mockSession := mocks.NewNeo4jSession(t)
	mockTransaction := mocks.NewNeo4jTransaction(t)

	mockSession.On("BeginTransaction").Return(mockTransaction, nil)
	mockTransaction.On("Run", mock.Anything, mock.Anything).Return(nil, nil)
	mockTransaction.On("Commit").Return(nil)

	deps := []analyzer.Dependency{
		{
			Source:       "TestSource",
			SourceLabel:  "TestLabel",
			Target:       "TestTarget",
			TargetLabel:  "TestLabel",
			Relationship: "TEST_REL",
		},
	}

	err := InsertDependencies(mockSession, deps)
	if err != nil {
		t.Fatalf("Failed to insert dependencies: %v", err)
	}
}
