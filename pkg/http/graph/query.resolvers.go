package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// ViewQuiz is the resolver for the viewQuiz field.
func (r *queryResolver) ViewQuiz(ctx context.Context, quizID string) (*model_cassandra.QuizCore, error) {
	panic(fmt.Errorf("not implemented: ViewQuiz - viewQuiz"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
