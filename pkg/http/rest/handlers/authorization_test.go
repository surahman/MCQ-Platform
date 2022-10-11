package http_handlers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/mocks"
)

func TestAuthMiddleware(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockAuth auth.Auth = mocks.NewMockAuth(mockCtrl)

	handler := AuthMiddleware(mockAuth)
	require.NotNil(t, handler)
}
