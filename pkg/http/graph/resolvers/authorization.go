package graphql_resolvers

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/logger"
)

// GinContextFromContext will extract the Gin context from the context passed in.
func GinContextFromContext(ctx context.Context, logger *logger.Logger) (*gin.Context, error) {
	ctxValue := ctx.Value("GinContextKey")
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
func AuthorizationCheck(auth auth.Auth, logger *logger.Logger, authHeaderKey string, ctx context.Context) error {
	ginContext, err := GinContextFromContext(ctx, logger)
	if err != nil {
		return err
	}

	tokenString := ginContext.GetHeader(authHeaderKey)
	if tokenString == "" {
		return errors.New("request does not contain an access token")
	}
	if _, _, err := auth.ValidateJWT(tokenString); err != nil {
		return err
	}
	return nil
}
