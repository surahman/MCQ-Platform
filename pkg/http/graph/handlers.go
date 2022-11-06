package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/http/graph/generated"
	"github.com/surahman/mcq-platform/pkg/http/graph/resolvers"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/redis"
)

// graphQLHandler is the endpoint through which GraphQL can be accessed.
func graphQLHandler(auth auth.Auth, cache redis.Redis, db cassandra.Cassandra,
	grading grading.Grading, logger *logger.Logger) gin.HandlerFunc {
	h := handler.NewDefaultServer(graphql_generated.NewExecutableSchema(
		graphql_generated.Config{
			Resolvers: &graphql_resolvers.Resolver{
				Auth:    auth,
				Cache:   cache,
				DB:      db,
				Grading: grading,
				Logger:  logger,
			},
		},
	))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// playgroundHandler is the endpoint through which the GraphQL playground can be accessed.
func playgroundHandler(endpointURL string) gin.HandlerFunc {
	h := playground.Handler("GraphQL", endpointURL)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
