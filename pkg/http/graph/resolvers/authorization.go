package graphql_resolvers

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
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
