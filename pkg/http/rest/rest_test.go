package rest

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/mocks"
)

func TestNewRESTServer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockAuth auth.Auth = mocks.NewMockAuth(mockCtrl)
	var mockCassandra cassandra.Cassandra = mocks.NewMockCassandra(mockCtrl)
	var mockGrading grading.Grading = mocks.NewMockGrading(mockCtrl)

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetHTTPRESTFileName(),
		[]byte(restConfigTestData["valid"]), 0644), "Failed to write in memory file")

	server, err := NewRESTServer(&fs, &mockAuth, &mockCassandra, zapLogger, &mockGrading)
	require.NoError(t, err, "error whilst creating mock server")
	require.NotNil(t, server, "failed to create mock server")

}
