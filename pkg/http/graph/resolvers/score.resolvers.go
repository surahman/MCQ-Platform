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
	model_http "github.com/surahman/mcq-platform/pkg/model/http"
)

// QuizID is the resolver for the QuizID field.
func (r *metadataResolver) QuizID(ctx context.Context, obj *model_http.Metadata) (string, error) {
	if obj == nil {
		return "", errors.New("invalid quiz id supplied")
	}
	return obj.QuizID.String(), nil
}

// GetScore is the resolver for the getScore field.
func (r *queryResolver) GetScore(ctx context.Context, quizID string) (*model_cassandra.Response, error) {
	var err error
	var dbRecord any
	var response *model_cassandra.Response
	var username string
	var quizId gocql.UUID

	if quizId, err = gocql.ParseUUID(quizID); err != nil {
		return nil, errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return nil, err
	}

	// Get scorecard record from database.
	scoreRequest := &model_cassandra.QuizMutateRequest{
		Username: username,
		QuizID:   quizId,
	}
	if dbRecord, err = r.DB.Execute(cassandra.ReadResponseQuery, scoreRequest); err != nil {
		return nil, err
	}
	response = dbRecord.(*model_cassandra.Response)

	return response, nil
}

// GetStats is the resolver for the getStats field.
func (r *queryResolver) GetStats(ctx context.Context, quizID string, pageSize *int, cursor *string) (*model_http.StatsResponseGraphQL, error) {
	panic(fmt.Errorf("not implemented: GetStats - getStats"))
}

// Metadata returns graphql_generated.MetadataResolver implementation.
func (r *Resolver) Metadata() graphql_generated.MetadataResolver { return &metadataResolver{r} }

type metadataResolver struct{ *Resolver }
