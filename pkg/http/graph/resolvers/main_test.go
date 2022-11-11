package graphql_resolvers

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"go.uber.org/zap"
)

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// testQuizData is the test quiz data.
var testAuthHeaderKey = "Authorization"

// testUserData is the test user account data.
var testUserData = cassandra.GetTestUsers()

// testQuizData is the test quiz data.
var testUserQuery = getUsersQuery()

// testQuizData is the test quiz data.
var testQuizQuery = getQuizzesQuery()

func TestMain(m *testing.M) {
	var err error
	// Configure logger.
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		log.Printf("Test suite logger setup failed: %v\n", err)
		os.Exit(1)
	}

	// Setup test space.
	if err = setup(); err != nil {
		zapLogger.Error("Test suite setup failure", zap.Error(err))
		os.Exit(1)
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err = tearDown(); err != nil {
		zapLogger.Error("Test suite teardown failure:", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

// setup will configure the auth test object.
func setup() (err error) {
	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	return
}

// getRouter creates a gin router testing instance.
func getRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

// mockAuthData is the parameter data for Auth mocking that is used in the test grid.
type mockAuthData struct {
	inputParam1  string
	inputParam2  string
	outputParam1 any
	outputParam2 int64
	outputErr    error
	times        int
}

// mockCassandraData is the parameter data for Cassandra mocking that is used in the test grid.
type mockCassandraData struct {
	inputFunc   func(cassandra.Cassandra, any) (any, error)
	inputParam  any
	outputParam any
	outputErr   error
	times       int
}

// mockGraderData is the parameter data for Grader mocking that is used in the test grid.
type mockGraderData struct {
	inputQuizResp *model_cassandra.QuizResponse
	inputQuiz     *model_cassandra.Quiz
	outputParam   float64
	outputErr     error
	times         int
}

// mockRedisData is the parameter data for Redis mocking that is used in the test grid.
type mockRedisData struct {
	param1 any
	param2 any
	err    error
	times  int
}

// verifyErrorReturned will check an HTTP response to ensure an error was returned.
func verifyErrorReturned(t *testing.T, response map[string]any) {
	value, ok := response["data"]
	require.True(t, ok, "data key expected but not set")
	require.Nil(t, value, "data value should be set to nil")

	value, ok = response["errors"]
	require.True(t, ok, "error key expected but not set")
	require.NotNil(t, value, "error value should not be nil")
}

// verifyJWTReturned will check an HTTP response from a resolver to ensure a correct JWT was returned.
func verifyJWTReturned(t *testing.T, response map[string]any, functionName string, expectedJWT *model_http.JWTAuthResponse) {
	data, ok := response["data"]
	require.True(t, ok, "data key expected but not set")

	authToken := model_http.JWTAuthResponse{}
	jsonStr, err := json.Marshal(data.(map[string]any)[functionName])
	require.NoError(t, err, "failed to generate JSON string")
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &authToken), "failed to unmarshall to JWT Auth Response")
	require.True(t, reflect.DeepEqual(*expectedJWT, authToken), "auth tokens did not match")
}
