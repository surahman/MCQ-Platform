package graphql_resolvers

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/logger"
)

type GinContextKey struct{}

// GinContextFromContext will extract the Gin context from the context passed in.
func GinContextFromContext(ctx context.Context, logger *logger.Logger) (*gin.Context, error) {
	ctxValue := ctx.Value(GinContextKey{})
	if ctxValue == nil {
		logger.Error("could not retrieve gin.Context")
		return nil, errors.New("malformed request: authorization information not found")
	}

	ginContext, ok := ctxValue.(*gin.Context)
	if !ok {
		logger.Error("gin.Context has wrong type")
		return nil, errors.New("malformed request: authorization information malformed")
	}
	return ginContext, nil
}

// AuthorizationCheck will validate the JWT payload for valid authorization information.
func AuthorizationCheck(auth auth.Auth, logger *logger.Logger, authHeaderKey string, ctx context.Context) (string, int64, error) {
	ginContext, err := GinContextFromContext(ctx, logger)
	if err != nil {
		return "", -1, err
	}

	tokenString := ginContext.GetHeader(authHeaderKey)
	if tokenString == "" {
		return "", -1, errors.New("request does not contain an access token")
	}

	var username string
	var expiresAt int64
	if username, expiresAt, err = auth.ValidateJWT(tokenString); err != nil {
		return username, expiresAt, err
	}
	return username, expiresAt, nil
}
