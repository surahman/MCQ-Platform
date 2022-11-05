package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	model_http "github.com/surahman/mcq-platform/pkg/model/http"
)

// CreateQuiz is the resolver for the createQuiz field.
func (r *mutationResolver) CreateQuiz(ctx context.Context, input model_http.QuizCreate) (string, error) {
	panic(fmt.Errorf("not implemented: CreateQuiz - createQuiz"))
}

// UpdateQuiz is the resolver for the updateQuiz field.
func (r *mutationResolver) UpdateQuiz(ctx context.Context, input model_http.QuizCreate) (string, error) {
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

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
