package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
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
	var err error
	var dbRecord any
	var statRequest *model_cassandra.StatsRequest
	var username string
	var quizId gocql.UUID

	if quizId, err = gocql.ParseUUID(quizID); err != nil {
		return nil, errors.New("invalid quiz id supplied, must be a valid UUID")
	}

	// Get username from JWT.
	if username, _, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return nil, err
	}

	// Prepare stats page request for database.
	if statRequest, err = http_common.PrepareStatsRequest(r.Auth, quizId, *cursor, strconv.Itoa(*pageSize)); err != nil {
		return nil, fmt.Errorf("malformed query request %v", err)
	}

	// Get scorecard record page from database.
	if dbRecord, err = r.DB.Execute(cassandra.ReadResponseStatisticsPageQuery, statRequest); err != nil {
		return nil, err
	}
	statsResponse := dbRecord.(*model_cassandra.StatsResponse)

	// Verify authorization.
	if len(statsResponse.Records) == 0 {
		return nil, errors.New("could not locate results")
	}
	if username != statsResponse.Records[0].Author {
		return nil, errors.New("error verifying quiz author")
	}

	// Prepare GraphQL response.
	return prepareStatsResponse(r.Auth, statsResponse, quizId)
}

// Metadata returns graphql_generated.MetadataResolver implementation.
func (r *Resolver) Metadata() graphql_generated.MetadataResolver { return &metadataResolver{r} }

type metadataResolver struct{ *Resolver }
