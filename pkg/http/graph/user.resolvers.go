package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	model_http "github.com/surahman/mcq-platform/pkg/model/http"
)

// RegisterUser is the resolver for the registerUser field.
func (r *mutationResolver) RegisterUser(ctx context.Context, input *model_http.UserRegistration) (*model_http.JWTAuthResponse, error) {
	panic(fmt.Errorf("not implemented: RegisterUser - registerUser"))
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, input model_http.UserDeletion) (string, error) {
	panic(fmt.Errorf("not implemented: DeleteUser - deleteUser"))
}

// LoginUser is the resolver for the loginUser field.
func (r *mutationResolver) LoginUser(ctx context.Context, input model_http.UserLogin) (*model_http.JWTAuthResponse, error) {
	panic(fmt.Errorf("not implemented: LoginUser - loginUser"))
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, token string) (*model_http.JWTAuthResponse, error) {
	panic(fmt.Errorf("not implemented: RefreshToken - refreshToken"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
