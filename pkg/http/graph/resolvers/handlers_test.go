package graphql_resolvers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	http_common "github.com/surahman/mcq-platform/pkg/http"
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

	handler := QueryHandler("Authorization", mockAuth, mockRedis, mockCassandra, mockGrader, zapLogger)

	require.NotNil(t, handler, "failed to create graphql endpoint handler")
}

func TestPlaygroundHandler(t *testing.T) {
	handler := PlaygroundHandler("/base-url", "/query-endpoint-url")
	require.NotNil(t, handler, "failed to create playground endpoint handler")
}

func TestGinContextToContextMiddleware(t *testing.T) {
	router := http_common.GetTestRouter()
	router.POST("/middleware-test", GinContextToContextMiddleware())
	req, _ := http.NewRequest("POST", "/middleware-test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify responses
	require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")
}
