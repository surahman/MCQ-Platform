package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/validator"
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
	panic(fmt.Errorf("not implemented: PublishQuiz - publishQuiz"))
}

// DeleteQuiz is the resolver for the deleteQuiz field.
func (r *mutationResolver) DeleteQuiz(ctx context.Context, quizID string) (string, error) {
	panic(fmt.Errorf("not implemented: DeleteQuiz - deleteQuiz"))
}

// ViewQuiz is the resolver for the viewQuiz field.
func (r *queryResolver) ViewQuiz(ctx context.Context, quizID string) (*model_cassandra.QuizCore, error) {
	panic(fmt.Errorf("not implemented: ViewQuiz - viewQuiz"))
}

// Query returns graphql_generated.QueryResolver implementation.
func (r *Resolver) Query() graphql_generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
