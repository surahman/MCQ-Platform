package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"github.com/surahman/mcq-platform/pkg/validator"
	"go.uber.org/zap"
)

// ViewQuiz will retrieve a test using a variable in the URL.
// @Summary     View a quiz.
// @Description This endpoint will retrieve a quiz with a provided quiz ID if it is published.
// @Tags        view test quiz
// @Id          viewQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string             true "The quiz ID for the quiz being requested."
// @Success     200     {object} model_rest.Success "The message will contain the quiz ID and the payload will contain the quiz"
// @Failure     403     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     404     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/view/{quiz_id} [get]
func ViewQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var response any
		var quiz *model_cassandra.Quiz
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

		// Get quiz record from database.
		if response, err = db.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error retrieving quiz", Payload: cassandraError.Message})
			return
		}
		quiz = response.(*model_cassandra.Quiz)

		// Check to see if quiz can be set to requester.
		// [1] Requested quiz is NOT published OR IS deleted
		// [2] Requester is not the author
		// FAIL
		if (!quiz.IsPublished || quiz.IsDeleted) && username != quiz.Author {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "quiz is not available"})
			return
		}

		// If the requester is not the author remove the answer key.
		if username != quiz.Author {
			for idx := range quiz.Questions {
				quiz.Questions[idx].Answers = nil
			}
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: quiz.QuizID.String(), Payload: &quiz.QuizCore})
	}
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
// @Param       quiz body     model_cassandra.QuizCore true "The Quiz to be created as unpublished"
// @Success     200  {object} model_rest.Success       "The message will contain the Quiz ID of the newly generated quiz"
// @Failure     400  {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     409  {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     500  {object} model_rest.Error         "Error message with any available details in payload"
// @Router      /quiz/create/ [post]
func CreateQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var username string
		var request model_cassandra.QuizCore

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in create quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "unable to verify username"})
			return
		}

		// Get quiz core from request and validate.
		if err = context.ShouldBindJSON(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		// Prepare quiz by adding username and generating quiz id, then insert record.
		quiz := model_cassandra.Quiz{
			QuizCore:    &request,
			QuizID:      gocql.TimeUUID(),
			Author:      username,
			IsPublished: false,
			IsDeleted:   false,
		}
		if _, err = db.Execute(cassandra.CreateQuizQuery, &quiz); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error creating quiz", Payload: cassandraError.Message})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "created quiz with id", Payload: quiz.QuizID.String()})
	}
}

// UpdateQuiz will update a quiz.
// @Summary     Update a quiz.
// @Description This endpoint will update a quiz with the provided Test ID if it was created by the requester and is not published.
// @Tags        update modify test quiz
// @Id          updateQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string                   true "The Test ID for the quiz being updated."
// @Param       quiz    body     model_cassandra.QuizCore true "The Quiz to replace the one already submitted"
// @Success     200     {object} model_rest.Success       "The message will contain a confirmation of the update"
// @Failure     400     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     403     {object} model_rest.Error         "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error         "Error message with any available details in payload"
// @Router      /quiz/update/{quiz_id} [patch]
func UpdateQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var username string
		var request model_cassandra.QuizCore
		var quizId gocql.UUID

		if quizId, err = gocql.ParseUUID(context.Param("quiz_id")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "invalid quiz id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in update quiz handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "unable to verify username"})
			return
		}

		// Get quiz core from request and validate.
		if err = context.ShouldBindJSON(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		// Prepare quiz by adding username and generating quiz id, then insert record.
		updateRequest := model_cassandra.QuizMutateRequest{
			Username: username,
			QuizID:   quizId,
			Quiz: &model_cassandra.Quiz{
				QuizCore: &request,
			},
		}
		if _, err = db.Execute(cassandra.UpdateQuizQuery, &updateRequest); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error updating quiz", Payload: cassandraError.Message})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "updated quiz with id", Payload: quizId.String()})
	}
}

// DeleteQuiz will delete a quiz using a variable in the URL.
// @Summary     Delete a quiz.
// @Description This endpoint will mark a quiz as delete if it was created by the requester. The provided Test ID is provided is a path parameter.
// @Tags        delete remove test quiz
// @Id          deleteQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string             true "The Test ID for the quiz being deleted."
// @Success     200     {object} model_rest.Success "The message will contain a confirmation of deletion"
// @Failure     403     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/delete/{quiz_id} [delete]
func DeleteQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
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

		// Delete quiz record from database.
		request := model_cassandra.QuizMutateRequest{
			Username: username,
			QuizID:   quizId,
		}
		if _, err = db.Execute(cassandra.DeleteQuizQuery, &request); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error deleting quiz", Payload: cassandraError.Message})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "deleted quiz with id", Payload: quizId.String()})
	}
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
// @Param       quiz_id path     string             true "The Test ID for the quiz being published."
// @Success     200     {object} model_rest.Success "The message will contain a confirmation of publishing"
// @Failure     403     {object} model_rest.Error   "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error   "Error message with any available details in payload"
// @Router      /quiz/publish/{quiz_id} [patch]
func PublishQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
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

		// Publish quiz record in database.
		request := model_cassandra.QuizMutateRequest{
			Username: username,
			QuizID:   quizId,
		}
		if _, err = db.Execute(cassandra.PublishQuizQuery, &request); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error publishing quiz", Payload: cassandraError.Message})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "published quiz with id", Payload: quizId.String()})
	}
}

// TakeQuiz will submit the answers to a quiz using a variable in the URL.
// @Summary     Take a quiz.
// @Description Take a quiz by submitting an answer sheet. The username will be extracted from the JWT and associated with the scorecard.
// @Tags        take test quiz submit answer
// @Id          takeQuiz
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       quiz_id path     string                       true "The Test ID for the answers being submitted."
// @Param       answers body     model_cassandra.QuizResponse true "The answer card to be submitted."
// @Success     200     {object} model_rest.Success           "Score will be in the payload"
// @Failure     400     {object} model_rest.Error             "Error message with any available details in payload"
// @Failure     403     {object} model_rest.Error             "Error message with any available details in payload"
// @Failure     500     {object} model_rest.Error             "Error message with any available details in payload"
// @Router      /quiz/take/{quiz_id} [post]
func TakeQuiz(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra, grader grading.Grading) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var username string
		var quizResponse model_cassandra.QuizResponse
		var quiz *model_cassandra.Quiz
		var quizId gocql.UUID
		var rawQuiz any
		var score float64

		if quizId, err = gocql.ParseUUID(context.Param("quiz_id")); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "invalid quizResponse id supplied, must be a valid UUID"})
			return
		}

		// Get username from JWT.
		if username, _, err = auth.ValidateJWT(context.GetHeader("Authorization")); err != nil {
			logger.Error("failed to validate JWT in update quizResponse handler", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "unable to verify username"})
			return
		}

		// Get quizResponse response from quizResponse and validate.
		if err = context.ShouldBindJSON(&quizResponse); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&quizResponse); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		// Get quizResponse record from database.
		if rawQuiz, err = db.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error retrieving quizResponse", Payload: cassandraError.Message})
			return
		}
		quiz = rawQuiz.(*model_cassandra.Quiz)

		// Check to see if the quiz is deleted or unpublished.
		if !quiz.IsPublished || quiz.IsDeleted {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "quiz is unavailable"})
			return
		}

		// Grade the quizResponse.
		if score, err = grader.Grade(&quizResponse, quiz.QuizCore); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "error marking response", Payload: err.Error()})
			return
		}

		// Insert updated record.
		response := model_cassandra.Response{
			Username:     username,
			Score:        score,
			QuizResponse: &quizResponse,
			QuizID:       quizId,
		}
		if _, err = db.Execute(cassandra.CreateResponseQuery, &response); err != nil {
			cassandraError := err.(*cassandra.Error)
			context.AbortWithStatusJSON(cassandraError.Status, &model_rest.Error{Message: "error submitting response", Payload: cassandraError.Message})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "submitted quiz response", Payload: &response})
	}
}
