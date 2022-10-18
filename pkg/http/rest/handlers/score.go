package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
func GetScore(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
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
