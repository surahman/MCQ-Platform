package http_handlers

import (
	"fmt"
	"net/http"
	"strconv"

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
// @Failure     403     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /score/stats/{quiz_id} [get]
func GetStats(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var dbRecord any
		var response []*model_cassandra.Response
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

		// Get quiz record from database and check to ensure requester is author.
		if dbRecord, err = db.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error verifying quiz author", Payload: cassandraError.Message})
			return
		}
		if username != dbRecord.(*model_cassandra.Quiz).Author {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "error verifying quiz author"})
			return
		}

		// Get scorecard record from database.
		if dbRecord, err = db.Execute(cassandra.ReadResponseStatisticsQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error retrieving score card", Payload: cassandraError.Message})
			return
		}
		response = dbRecord.([]*model_cassandra.Response)

		context.JSON(http.StatusOK, &model_rest.Success{Message: "score cards", Payload: response})
	}
}

// GetStatsPage will retrieve paginated test statistics with the provided test id and the username from the JWT payload.
// @Summary     Get paginated statistics associated with a specific test.
// @Description Gets the paginated statistics associated with a specific test if the user created the test.
// @Description Extracts username from the JWT and the Test ID is provided as a query parameter.
// @Description A query string to be appended to the next request to retrieve the next page of data will be returned in the response.
// @Tags        score scores stats statistics
// @Id          getStatsPaged
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id    path     string                   true  "The Test ID for the requested statistics."
// @Param       pageCursor query    string                   false "The page cursor into the query results records."
// @Param       pageSize   query    int                      false "The number of records to retrieve on this page."
// @Success     200        {object} model_rest.StatsResponse "A page of statistics data"
// @Failure     400        {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     403        {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     404        {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     500        {object} model_rest.Error         "Error message with any available details in payload"
// @Router      /score/stats/{quiz_id} [get]
func GetStatsPage(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var dbRecord any
		var response []*model_cassandra.Response
		var username string
		var quizId gocql.UUID

		if quizId, err = gocql.ParseUUID(context.Param("quizId")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "unable to verify username"})
			return
		}

		// Get quiz record from database and check to ensure requester is author.
		if dbRecord, err = db.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error verifying quiz author", Payload: cassandraError.Message})
			return
		}
		if username != dbRecord.(*model_cassandra.Quiz).Author {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "error verifying quiz author"})
			return
		}

		// Get scorecard record from database.
		if dbRecord, err = db.Execute(cassandra.ReadResponseStatisticsQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error retrieving score card", Payload: cassandraError.Message})
			return
		}
		response = dbRecord.([]*model_cassandra.Response)

		context.JSON(http.StatusOK, &model_rest.Success{Message: "score cards", Payload: response})
	}
}

// prepareStatsRequest will prepare the paged statistics request for the database query.
func prepareStatsRequest(auth auth.Auth, quizId gocql.UUID, cursor string, size string) (req *model_cassandra.StatsRequest, err error) {
	req = &model_cassandra.StatsRequest{QuizID: quizId}

	if req.PageSize, err = strconv.Atoi(size); err != nil {
		return nil, fmt.Errorf("failed to convert page size: %s", err.Error())
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	// Null cursor must be set if there was no cursor in the URI.
	if len(cursor) != 0 {
		if req.PageCursor, err = auth.DecryptFromString(cursor); err != nil {
			return nil, fmt.Errorf("failed to decrypt page cursor: %s", err.Error())
		}
	}

	return
}
