package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"go.uber.org/zap"
)

// GetScore will retrieve a test score with the provided test id and the username from the JWT payload.
// @Summary     Get a user's score.
// @Description Gets a scorecard for a user. Extracts username from the JWT and Test ID is provided as a path parameter.
// @Tags        score scores
// @Id          getScore
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string             true "The Test ID for the requested scorecard."
// @Success     200     {object} model_rest.Success "Score will be in the payload"
// @Failure     400     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /score/test/{quiz_id} [get]
func GetScore(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var dbRecord any
		var response *model_cassandra.Response
		var username string
		var quizId gocql.UUID

		if quizId, err = gocql.ParseUUID(context.Param("quiz_id")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "unable to verify username"})
			return
		}

		// Get scorecard record from database.
		scoreRequest := &model_cassandra.QuizMutateRequest{
			Username: username,
			QuizID:   quizId,
		}
		if dbRecord, err = db.Execute(cassandra.ReadResponseQuery, scoreRequest); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error retrieving score card", Payload: cassandraError.Message})
			return
		}
		response = dbRecord.(*model_cassandra.Response)

		context.JSON(http.StatusOK, &model_rest.Success{Message: "score card", Payload: response})
	}
}

// GetStats will retrieve test statistics with the provided test id and the username from the JWT payload.
// @Summary     Get all statistics associated with a specific test.
// @Description Gets the statistics associated with a specific test if the user created the test.
// @Description Extracts username from the JWT and the Test ID is provided as a path parameter.
// @Tags        score scores stats statistics
// @Id          getStats
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string             true "The Test ID for the requested statistics."
// @Success     200     {object} model_rest.Success "Statistics will be in the payload"
// @Failure     400     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     401     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /score/stats/{quiz_id} [get]
func GetStats(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}
