package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// TakeQuiz is the resolver for the takeQuiz field.
func (r *mutationResolver) TakeQuiz(ctx context.Context, quizID string, input model_cassandra.QuizResponse) (float64, error) {
	panic(fmt.Errorf("not implemented: TakeQuiz - takeQuiz"))
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
