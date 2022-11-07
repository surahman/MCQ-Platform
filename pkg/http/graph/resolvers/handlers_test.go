package graphql_resolvers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/mocks"
)

func TestQueryHandler(t *testing.T) {
	// Mock configurations.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuth := mocks.NewMockAuth(mockCtrl)
	mockCassandra := mocks.NewMockCassandra(mockCtrl)
	mockRedis := mocks.NewMockRedis(mockCtrl)
	mockGrader := mocks.NewMockGrading(mockCtrl)

	handler := QueryHandler(mockAuth, mockRedis, mockCassandra, mockGrader, zapLogger)

	require.NotNil(t, handler, "failed to create graphql endpoint handler")
}

func TestPlaygroundHandler(t *testing.T) {
	handler := PlaygroundHandler("/query-endpoint-url")
	require.NotNil(t, handler, "failed to create playground endpoint handler")
}