package http_handlers

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
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
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_http.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_http.Error{Message: "unable to verify username"})
			return
		}

		// Get scorecard record from database.
		scoreRequest := &model_cassandra.QuizMutateRequest{
			Username: username,
			QuizID:   quizId,
		}
		if dbRecord, err = db.Execute(cassandra.ReadResponseQuery, scoreRequest); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_http.Error{Message: "error retrieving score card", Payload: cassandraError.Message})
			return
		}
		response = dbRecord.(*model_cassandra.Response)

		context.JSON(http.StatusOK, &model_http.Success{Message: "score card", Payload: response})
	}
}

// GetStats will retrieve test statistics with the provided test id and the username from the JWT payload.
// @Summary     Get all statistics associated with a specific test.
// @Description Gets the statistics associated with a specific test if the user created the test.
// @Description Extracts username from the JWT and the Test ID is provided as a path parameter.
// @Tags        score scores stats statistics
// @Id          getStats
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
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_http.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_http.Error{Message: "unable to verify username"})
			return
		}

		// Get scorecard record from database.
		if dbRecord, err = db.Execute(cassandra.ReadResponseStatisticsQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_http.Error{Message: "error retrieving score card", Payload: cassandraError.Message})
			return
		}
		response = dbRecord.([]*model_cassandra.Response)

		// Verify authorization.
		if len(response) == 0 {
			context.AbortWithStatusJSON(http.StatusNotFound, &model_http.Error{Message: "could not locate results"})
			return
		}
		if username != response[0].Author {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_http.Error{Message: "error verifying quiz author"})
			return
		}

		context.JSON(http.StatusOK, &model_http.Success{Message: "score card", Payload: response})
	}
}

// GetStatsPage will retrieve paginated test statistics with the provided test id and the username from the JWT payload.
// @Summary     Get paginated statistics associated with a specific test.
// @Description Gets the paginated statistics associated with a specific test if the user created the test.
// @Description Extracts username from the JWT and the Test ID is provided as a query parameter.
// @Description A query string to be appended to the next request to retrieve the next page of data will be returned in the response.
// @Tags        score scores stats statistics
// @Id          getStatsPaged
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
// @Router      /score/stats-paged/{quiz_id} [get]
func GetStatsPage(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var dbRecord any
		var statRequest *model_cassandra.StatsRequest
		var restResponse *model_http.StatsResponse
		var username string
		var quizId gocql.UUID

		if quizId, err = gocql.ParseUUID(context.Param("quiz_id")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_http.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_http.Error{Message: "unable to verify username"})
			return
		}

		// Prepare stats page request for database.
		if statRequest, err = http_common.PrepareStatsRequest(auth, quizId, context.Query("pageCursor"), context.Query("pageSize")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_http.Error{Message: "malformed query request", Payload: err.Error()})
			return
		}

		// Get scorecard record page from database.
		if dbRecord, err = db.Execute(cassandra.ReadResponseStatisticsPageQuery, statRequest); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_http.Error{Message: "error retrieving scorecard page", Payload: cassandraError.Message})
			return
		}
		statsResponse := dbRecord.(*model_cassandra.StatsResponse)

		// Verify authorization.
		if len(statsResponse.Records) == 0 {
			context.AbortWithStatusJSON(http.StatusNotFound, &model_http.Error{Message: "could not locate results"})
			return
		}
		if username != statsResponse.Records[0].Author {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_http.Error{Message: "error verifying quiz author"})
			return
		}

		// Prepare REST response.
		if restResponse, err = prepareStatsResponse(auth, statsResponse, quizId); err != nil {
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_http.Error{Message: "error preparing stats page", Payload: err.Error()})
		}

		context.JSON(http.StatusOK, restResponse)
	}
}

// prepareStatsResponse will prepare a http response from a database stats response. It will generate a link to the next
// page of data as appropriate.
func prepareStatsResponse(auth auth.Auth, dbResponse *model_cassandra.StatsResponse, quizId gocql.UUID) (response *model_http.StatsResponse, err error) {
	response = &model_http.StatsResponse{Records: dbResponse.Records}
	response.Metadata.QuizID = quizId
	response.Metadata.NumRecords = len(dbResponse.Records)

	var nextPageLink bytes.Buffer

	// Construct page cursor link segment.
	if len(dbResponse.PageCursor) != 0 {
		cursor, err := auth.EncryptToString(dbResponse.PageCursor)
		if err != nil {
			return nil, err
		}
		if _, err := nextPageLink.WriteString(fmt.Sprintf("?pageCursor=%s", cursor)); err != nil {
			return nil, fmt.Errorf("failed to generate next page cursor link segment: %s", err.Error())
		}
	}

	// Construct page size link segment.
	if len(dbResponse.PageCursor) != 0 && dbResponse.PageSize > 0 {
		if _, err := nextPageLink.WriteString(fmt.Sprintf("&pageSize=%d", dbResponse.PageSize)); err != nil {
			return nil, fmt.Errorf("failed to generate next page size link segment: %s", err.Error())
		}
	}

	response.Links.NextPage = nextPageLink.String()

	return
}
