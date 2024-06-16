package database

import (
	"context"
	"testing"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
	"github.com/caiooliveiraeti/database-deptree/test/mocks"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/mock"
)

func TestInsertDependencies(t *testing.T) {
	mockSession := mocks.NewNeo4jSession(t)
	mockTransaction := NewNeo4jExplicitTransaction(t) // Use the custom mock

	ctx := context.Background()

	mockSession.On("BeginTransaction", ctx).Return(mockTransaction, nil)
	mockTransaction.On("Run", ctx, mock.Anything, mock.Anything).Return(nil, nil)
	mockTransaction.On("Commit", ctx).Return(nil)

	deps := []analyzer.Dependency{
		{
			Source:       "TestSource",
			SourceLabel:  "TestLabel",
			Target:       "TestTarget",
			TargetLabel:  "TestLabel",
			Relationship: "TEST_REL",
		},
	}

	InsertDependencies(ctx, mockSession, deps)
}

type MockNeo4jExplicitTransaction struct {
	neo4j.ExplicitTransaction
	mock.Mock
}

func (_m *MockNeo4jExplicitTransaction) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *MockNeo4jExplicitTransaction) Commit(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *MockNeo4jExplicitTransaction) Rollback(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *MockNeo4jExplicitTransaction) Run(ctx context.Context, cypher string, params map[string]interface{}) (neo4j.ResultWithContext, error) {
	ret := _m.Called(ctx, cypher, params)

	if len(ret) == 0 {
		panic("no return value specified for Run")
	}

	var r0 neo4j.ResultWithContext
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]interface{}) (neo4j.ResultWithContext, error)); ok {
		return rf(ctx, cypher, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]interface{}) neo4j.ResultWithContext); ok {
		r0 = rf(ctx, cypher, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(neo4j.ResultWithContext)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, map[string]interface{}) error); ok {
		r1 = rf(ctx, cypher, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewNeo4jExplicitTransaction(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockNeo4jExplicitTransaction {
	mock := &MockNeo4jExplicitTransaction{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
