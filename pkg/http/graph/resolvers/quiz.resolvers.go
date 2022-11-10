package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// CreateQuiz is the resolver for the createQuiz field.
func (r *mutationResolver) CreateQuiz(ctx context.Context, input model_cassandra.QuizCore) (string, error) {
	panic(fmt.Errorf("not implemented: CreateQuiz - createQuiz"))
}

// UpdateQuiz is the resolver for the updateQuiz field.
func (r *mutationResolver) UpdateQuiz(ctx context.Context, input model_cassandra.QuizCore) (string, error) {
	panic(fmt.Errorf("not implemented: UpdateQuiz - updateQuiz"))
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
