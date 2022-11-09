package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/surahman/mcq-platform/pkg/cassandra"
	graphql_generated "github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	model_http "github.com/surahman/mcq-platform/pkg/model/http"
	"github.com/surahman/mcq-platform/pkg/validator"
	"go.uber.org/zap"
)

// RegisterUser is the resolver for the registerUser field.
func (r *mutationResolver) RegisterUser(ctx context.Context, input *model_cassandra.UserAccount) (*model_http.JWTAuthResponse, error) {
	var err error
	var authToken *model_http.JWTAuthResponse

	if err = validator.ValidateStruct(input); err != nil {
		return authToken, err
	}

	if input.Password, err = r.Auth.HashPassword(input.Password); err != nil {
		r.Logger.Error("failure hashing password", zap.Error(err))
		return authToken, err
	}

	if _, err = r.DB.Execute(cassandra.CreateUserQuery, &model_cassandra.User{UserAccount: input}); err != nil {
		return authToken, err
	}

	if authToken, err = r.Auth.GenerateJWT(input.Username); err != nil {
		r.Logger.Error("failure generating JWT during account creation", zap.Error(err))
		return authToken, err
	}

	return authToken, err
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, input model_http.DeleteUserRequest) (string, error) {
	panic(fmt.Errorf("not implemented: DeleteUser - deleteUser"))
}

// LoginUser is the resolver for the loginUser field.
func (r *mutationResolver) LoginUser(ctx context.Context, input model_cassandra.UserLoginCredentials) (*model_http.JWTAuthResponse, error) {
	var err error
	var authToken *model_http.JWTAuthResponse
	var dbResponse any

	if err = validator.ValidateStruct(&input); err != nil {
		return nil, err
	}

	if dbResponse, err = r.DB.Execute(cassandra.ReadUserQuery, input.Username); err != nil {
		return nil, err
	}

	truth := dbResponse.(*model_cassandra.User)
	if err = r.Auth.CheckPassword(truth.Password, input.Password); err != nil || truth.IsDeleted {
		return nil, err
	}

	if authToken, err = r.Auth.GenerateJWT(input.Username); err != nil {
		r.Logger.Error("failure generating JWT during login", zap.Error(err))
		return nil, err
	}

	return authToken, err
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context) (*model_http.JWTAuthResponse, error) {
	var err error
	var freshToken *model_http.JWTAuthResponse
	var username string
	var dbResponse any
	var expiresAt int64

	if username, expiresAt, err = AuthorizationCheck(r.Auth, r.Logger, r.AuthHeaderKey, ctx); err != nil {
		return nil, err
	}

	if dbResponse, err = r.DB.Execute(cassandra.ReadUserQuery, username); err != nil {
		r.Logger.Warn("failed to read user record for a valid JWT", zap.String("username", username), zap.Error(err))
		return nil, errors.New("please retry your request later")
	}

	if dbResponse.(*model_cassandra.User).IsDeleted {
		r.Logger.Warn("attempt to refresh a JWT for a deleted user", zap.String("username", username))
		return nil, errors.New("invalid token")
	}

	// Do not refresh tokens that have more than a minute left to expire.
	if math.Abs(float64(time.Now().Unix()-expiresAt)) > float64(r.Auth.RefreshThreshold()) {
		return nil, errors.New("JWT is still valid for more than 60 seconds")
	}

	if freshToken, err = r.Auth.GenerateJWT(username); err != nil {
		r.Logger.Error("failure generating JWT during token refresh", zap.Error(err))
		return nil, errors.New(err.Error())
	}
	return freshToken, nil
}

// Mutation returns graphql_generated.MutationResolver implementation.
func (r *Resolver) Mutation() graphql_generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
