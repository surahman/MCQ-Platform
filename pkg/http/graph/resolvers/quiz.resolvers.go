package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/redis"
	"github.com/surahman/mcq-platform/pkg/validator"
	"go.uber.org/zap"
)

// CreateQuiz is the resolver for the createQuiz field.
func (r *mutationResolver) CreateQuiz(ctx context.Context, input model_cassandra.QuizCore) (string, error) {
	var err error
	var username string

	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return "", err
	}

	if err = validator.ValidateStruct(&input); err != nil {
		return "", err
	}

	// Prepare quiz by adding username and generating quiz id, then insert record.
	quiz := model_cassandra.Quiz{
		QuizCore:    &input,
		QuizID:      gocql.TimeUUID(),
		Author:      username,
		IsPublished: false,
		IsDeleted:   false,
	}
	if _, err = r.DB.Execute(cassandra.CreateQuizQuery, &quiz); err != nil {
		return "", err
	}

	return quiz.QuizID.String(), nil
}

// UpdateQuiz is the resolver for the updateQuiz field.
func (r *mutationResolver) UpdateQuiz(ctx context.Context, quizID string, quiz model_cassandra.QuizCore) (string, error) {
	var err error
	var username string
	var quizUUID gocql.UUID

	if quizUUID, err = gocql.ParseUUID(quizID); err != nil {
		return "", errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return "", err
	}

	if err = validator.ValidateStruct(&quiz); err != nil {
		return "", err
	}

	// Prepare quiz by adding username and provided quiz UUID, then insert record.
	updateRequest := model_cassandra.QuizMutateRequest{
		Username: username,
		QuizID:   quizUUID,
		Quiz: &model_cassandra.Quiz{
			QuizCore: &quiz,
		},
	}
	if _, err = r.DB.Execute(cassandra.UpdateQuizQuery, &updateRequest); err != nil {
		return "", err
	}

	return quizUUID.String(), nil
}

// PublishQuiz is the resolver for the publishQuiz field.
func (r *mutationResolver) PublishQuiz(ctx context.Context, quizID string) (string, error) {
	var err error
	var username string
	var quizId gocql.UUID
	var response any
	var quiz *model_cassandra.Quiz

	if quizId, err = gocql.ParseUUID(quizID); err != nil {
		return "", errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return "", err
	}

	// Publish quiz record in database.
	request := model_cassandra.QuizMutateRequest{
		Username: username,
		QuizID:   quizId,
	}
	if _, err = r.DB.Execute(cassandra.PublishQuizQuery, &request); err != nil {
		return "", err
	}

	// Success message should be set here because publishing succeeded.
	// Any failures below this point are cache related and should be logged but not propagated to the end user.
	returnMsg := fmt.Sprintf("published quiz with id %s", quizId.String())

	// Place quiz in cache.
	// [1] Retrieve the quiz from Cassandra.
	// [2] Place into Redis.

	// Get quiz record from database.
	if response, err = r.DB.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
		r.Logger.Error("error retrieving quiz from database to be placed in cache post publishing", zap.Error(err))
		return returnMsg, nil
	}
	quiz = response.(*model_cassandra.Quiz)

	if err = r.Cache.Set(quizId.String(), quiz); err != nil {
		r.Logger.Error("error placing quiz in cache after publishing", zap.Error(err))
		return returnMsg, nil
	}

	return returnMsg, nil
}

// DeleteQuiz is the resolver for the deleteQuiz field.
func (r *mutationResolver) DeleteQuiz(ctx context.Context, quizID string) (string, error) {
	var err error
	var username string
	var quizId gocql.UUID

	if quizId, err = gocql.ParseUUID(quizID); err != nil {
		return "", errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return "", err
	}

	// Evict from cache, if present.
	// This step must be executed before deletion to ensure the end user is able to reattempt the command in the event of failure.
	// It must not be the case that data marked as deleted remains in the cache till LRU eviction or TTL expiration.
	if err = r.Cache.Del(quizId.String()); err != nil && err.(*redis.Error).Code != redis.ErrorCacheMiss {
		r.Logger.Error("failed to evict data from cache", zap.Error(err))
		return "", errors.New("please retry the command at a later time")
	}

	// Delete quiz record from database.
	request := model_cassandra.QuizMutateRequest{
		Username: username,
		QuizID:   quizId,
	}
	if _, err = r.DB.Execute(cassandra.DeleteQuizQuery, &request); err != nil {
		return "", err
	}

	return fmt.Sprintf("successfully deleted %s", quizId.String()), nil
}

// ViewQuiz is the resolver for the viewQuiz field.
func (r *queryResolver) ViewQuiz(ctx context.Context, quizID string) (*model_cassandra.QuizCore, error) {
	var err error
	var quiz *model_cassandra.Quiz
	var username string
	var quizUUID gocql.UUID

	if quizUUID, err = gocql.ParseUUID(quizID); err != nil {
		return nil, errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return nil, err
	}

	// Get quiz:
	// [1] Cache call.
	// [2] Cache miss: read from the database and store it in the cache.
	if quiz, err = http_common.GetQuiz(quizUUID, r.DB, r.Cache); err != nil {
		return nil, err
	}

	// Check to see if quiz can be set to requester.
	// [1] Requested quiz is NOT published OR IS deleted
	// [2] Requester is not the author
	// FAIL
	if (!quiz.IsPublished || quiz.IsDeleted) && username != quiz.Author {
		return nil, errors.New("quiz is not available")
	}

	// If the requester is not the author remove the answer key.
	if username != quiz.Author {
		for idx := range quiz.Questions {
			quiz.Questions[idx].Answers = nil
		}
	}

	return quiz.QuizCore, nil
}

// Query returns graphql_generated.QueryResolver implementation.
func (r *Resolver) Query() graphql_generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
