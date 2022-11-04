package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// QuizID is the resolver for the QuizID field.
func (r *responseResolver) QuizID(ctx context.Context, obj *model_cassandra.Response) (string, error) {
	panic(fmt.Errorf("not implemented: QuizID - QuizID"))
}

// Response returns generated.ResponseResolver implementation.
func (r *Resolver) Response() generated.ResponseResolver { return &responseResolver{r} }

type responseResolver struct{ *Resolver }
