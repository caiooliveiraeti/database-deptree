package database

import (
	"testing"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/golang/mock/gomock"
)

func TestInsertDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockNeo4jSession(ctrl)
	mockTransaction := mocks.NewMockNeo4jTransaction(ctrl)

	mockSession.EXPECT().BeginTransaction().Return(mockTransaction, nil).Times(1)
	mockTransaction.EXPECT().Run(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	mockTransaction.EXPECT().Commit().Return(nil).Times(1)

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
