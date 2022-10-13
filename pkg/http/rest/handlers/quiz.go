package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ViewQuiz will retrieve a test using a variable in the URL.
// @Summary     View a quiz.
// @Description This endpoint will retrieve a quiz with a provided Test ID if it is published.
// @Tags        view test quiz
// @Id          viewQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       test_id path     string             true "The Test ID for the quiz being requested."
// @Success     200     {object} model_rest.Success "The message will contain the Test ID and the payload will contain the quiz"
// @Failure     400     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     401     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/view/{test_id} [get]
func ViewQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// CreateQuiz will submit a quiz and write back the GetScore ID.
// @Summary     Create a quiz.
// @Description This endpoint will create a quiz with randomly generated Test ID and associate it with the requester.
// @Description The username will be extracted from the JWT and associated with the Test ID.
// @Tags        create test quiz
// @Id          createQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       answers body     model_cassandra.QuizCore true "The Quiz to be created as unpublished"
// @Success     200     {object} model_rest.Success       "The message will contain the Test ID of the newly generated quiz"
// @Failure     400     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error         "Error message with any available details in payload"
// @Router      /quiz/create/ [post]
func CreateQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// UpdateQuiz will update a quiz.
// @Summary     Update a quiz.
// @Description This endpoint will update a quiz with the provided Test ID if it was created by the requester and is not published.
// @Tags        update modify test quiz
// @Id          updateQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       test_id path     string                   true "The Test ID for the quiz being updated."
// @Param       answers body     model_cassandra.QuizCore true "The Quiz to replace the one already submitted"
// @Success     200     {object} model_rest.Success       "The message will contain a confirmation of the update"
// @Failure     400     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     401     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     403     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error         "Error message with any available details in payload"
// @Router      /quiz/update/{test_id} [put]
func UpdateQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// DeleteQuiz will delete a quiz using a variable in the URL.
// @Summary     Delete a quiz.
// @Description This endpoint will delete a quiz with the provided Test ID if it was created by the requester.
// @Tags        delete remove test quiz
// @Id          deleteQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       test_id path     string             true "The Test ID for the quiz being deleted."
// @Success     200     {object} model_rest.Success "The message will contain a confirmation of deletion"
// @Failure     401     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/delete/{test_id} [delete]
func DeleteQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// PublishQuiz will publish a quiz using a variable in the URL.
// @Summary     Publish a quiz.
// @Description When a quiz is submitted it is not published by default and is thus unavailable to be taken.
// @Description This endpoint will publish a quiz with the provided Test ID if it was created by the requester.
// @Tags        publish test quiz create
// @Id          publishQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       test_id path     string             true "The Test ID for the quiz being published."
// @Success     200     {object} model_rest.Success "The message will contain a confirmation of publishing"
// @Failure     401     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/publish/{test_id} [put]
func PublishQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// TakeQuiz will submit the answers to a quiz using a variable in the URL.
// @Summary     Take a quiz.
// @Description Take a quiz by submitting an answer sheet. The username will be extracted from the JWT and associated with the scorecard.
// @Tags        take test quiz submit answer
// @Id          takeQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       test_id path     string                       true "The Test ID for the answers being submitted."
// @Param       answers body     model_cassandra.QuizResponse true "The answer card to be submitted."
// @Success     200     {object} model_rest.Success           "Score will be in the payload"
// @Failure     400     {object} model_rest.Error             "Error message with any available details in payload"
// @Failure     401     {object} model_rest.Error             "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error             "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error             "Error message with any available details in payload"
// @Router      /quiz/take/{test_id} [post]
func TakeQuiz(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}
