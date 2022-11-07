package main

import (
	"log"
	"sync"

	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	graphql "github.com/surahman/mcq-platform/pkg/http/graph"
	"github.com/surahman/mcq-platform/pkg/http/rest"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/redis"
	"go.uber.org/zap"
)

func main() {
	var (
		err           error
		serverREST    *rest.Server
		serverGraphQL *graphql.Server
		logging       *logger.Logger
		authorization auth.Auth
		database      cassandra.Cassandra
		cache         redis.Redis
		grader        = grading.NewGrading()
		waitGroup     sync.WaitGroup
	)

	// File system setup.
	fs := afero.NewOsFs()

	// Logger setup.
	logging = logger.NewLogger()
	if err = logging.Init(&fs); err != nil {
		log.Fatalf("failed to initialize logger module: %v", err)
	}

	// Authorization setup.
	if authorization, err = auth.NewAuth(&fs, logging); err != nil {
		logging.Panic("failed to configure authorization module", zap.Error(err))
	}

	// Cassandra setup.
	if database, err = cassandra.NewCassandra(&fs, logging); err != nil {
		logging.Panic("failed to configure Cassandra module", zap.Error(err))
	}
	if err = database.Open(); err != nil {
		logging.Panic("failed open a connection to the Cassandra cluster", zap.Error(err))
	}
	defer func(database cassandra.Cassandra) {
		if err = database.Close(); err != nil {
			logging.Panic("failed close the connection to the Cassandra cluster", zap.Error(err))
		}
	}(database)

	// Cache setup
	if cache, err = redis.NewRedis(&fs, logging); err != nil {
		logging.Panic("failed to configure Redis module", zap.Error(err))
	}
	if err = cache.Open(); err != nil {
		logging.Panic("failed open a connection to the Redis cluster", zap.Error(err))
	}
	defer func(cache redis.Redis) {
		if err = cache.Close(); err != nil {
			logging.Panic("failed close the connection to the Redis cluster", zap.Error(err))
		}
	}(cache)

	// Setup REST server and start it.
	if serverREST, err = rest.NewServer(&fs, authorization, database, cache, grader, logging, &waitGroup); err != nil {
		logging.Panic("failed to create the REST server", zap.Error(err))
	}
	go serverREST.Run()

	// Setup GraphQL server and start it.
	if serverGraphQL, err = graphql.NewServer(&fs, authorization, database, cache, grader, logging, &waitGroup); err != nil {
		logging.Panic("failed to create the GraphQL server", zap.Error(err))
	}
	go serverGraphQL.Run()

	waitGroup.Wait()
}
