package graphql_resolvers

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/redis"
)

// QueryHandler is the endpoint through which GraphQL can be accessed.
func QueryHandler(authHeaderKey string, auth auth.Auth, cache redis.Redis, db cassandra.Cassandra,
	grading grading.Grading, logger *logger.Logger) gin.HandlerFunc {
	h := handler.NewDefaultServer(graphql_generated.NewExecutableSchema(
		graphql_generated.Config{
			Resolvers: &Resolver{
				AuthHeaderKey: authHeaderKey,
				Auth:          auth,
				Cache:         cache,
				DB:            db,
				Grading:       grading,
				Logger:        logger,
			},
		},
	))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler is the endpoint through which the GraphQL playground can be accessed.
func PlaygroundHandler(baseURL, queryURL string) gin.HandlerFunc {
	h := playground.Handler("GraphQL", fmt.Sprintf("/%s%s", baseURL, queryURL))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GinContextToContextMiddleware is middleware that will place the Gin context into a context for the GraphQL resolvers.
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "GinContextKey", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
