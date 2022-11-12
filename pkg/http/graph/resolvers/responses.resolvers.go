package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/validator"
)

// TakeQuiz is the resolver for the takeQuiz field.
func (r *mutationResolver) TakeQuiz(ctx context.Context, quizID string, input model_cassandra.QuizResponse) (*model_cassandra.Response, error) {
	var err error
	var username string
	var quiz *model_cassandra.Quiz
	var quizId gocql.UUID
	var score float64

	if quizId, err = gocql.ParseUUID(quizID); err != nil {
		return nil, errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return nil, err
	}

	if err = validator.ValidateStruct(&input); err != nil {
		return nil, err
	}

	// Get quiz:
	// [1] Cache call.
	// [2] Cache miss: read from the database and store it in the cache.
	if quiz, err = http_common.GetQuiz(quizId, r.DB, r.Cache); err != nil {
		return nil, err
	}

	// Check to see if the quiz is deleted or unpublished.
	if !quiz.IsPublished || quiz.IsDeleted {
		return nil, err
	}

	// Grade the quizResponse.
	if score, err = r.Grading.Grade(&input, quiz.QuizCore); err != nil {
		return nil, err
	}

	// Insert updated record.
	response := model_cassandra.Response{
		Username:     username,
		Score:        score,
		QuizResponse: &input,
		QuizID:       quizId,
	}
	if _, err = r.DB.Execute(cassandra.CreateResponseQuery, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// QuizResponse is the resolver for the QuizResponse field.
func (r *responseResolver) QuizResponse(ctx context.Context, obj *model_cassandra.Response) ([][]int32, error) {
	if obj.QuizResponse == nil {
		return nil, nil
	}
	return obj.QuizResponse.Responses, nil
}

// QuizID is the resolver for the QuizID field.
func (r *responseResolver) QuizID(ctx context.Context, obj *model_cassandra.Response) (string, error) {
	return obj.QuizID.String(), nil
}

// Response returns graphql_generated.ResponseResolver implementation.
func (r *Resolver) Response() graphql_generated.ResponseResolver { return &responseResolver{r} }

type responseResolver struct{ *Resolver }
