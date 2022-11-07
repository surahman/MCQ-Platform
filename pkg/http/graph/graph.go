package graphql

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/http/graph/resolvers"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/redis"
	"go.uber.org/zap"
)

// Generate the GraphQL Go code for resolvers.
//go:generate go run github.com/99designs/gqlgen generate

// Server is the HTTP GraphQL server.
type Server struct {
	auth    auth.Auth
	cache   redis.Redis
	db      cassandra.Cassandra
	grading grading.Grading
	conf    *config
	logger  *logger.Logger
	router  *gin.Engine
	wg      *sync.WaitGroup
}

// NewServer will create a new REST server instance in a non-running state.
func NewServer(fs *afero.Fs, auth auth.Auth, cassandra cassandra.Cassandra, redis redis.Redis,
	grading grading.Grading, logger *logger.Logger, wg *sync.WaitGroup) (server *Server, err error) {
	// Load configurations.
	conf := newConfig()
	if err = conf.Load(*fs); err != nil {
		return
	}

	// Add to the wait group to stop bootstrap thread from exiting.
	wg.Add(1)

	return &Server{
			conf:    conf,
			auth:    auth,
			cache:   redis,
			db:      cassandra,
			grading: grading,
			logger:  logger,
			wg:      wg,
		},
		err
}

// Run brings the HTTP service up.
func (s *Server) Run() {
	// Configure routes.
	s.initialize()

	// Create server.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.conf.Server.PortNumber),
		Handler: s.router,
	}

	// Start HTTP listener.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Panic(fmt.Sprintf("listening port: %d", s.conf.Server.PortNumber), zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt.
	<-quit
	s.logger.Info("Shutting down server...", zap.Duration("waiting", time.Duration(s.conf.Server.ShutdownDelay)*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.conf.Server.ShutdownDelay)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Panic("Failed to shutdown server", zap.Error(err))
	}

	// 5 second wait to exit.
	<-ctx.Done()

	s.logger.Info("Server exited")

	// Indicate completion to wait group.
	s.wg.Done()
}

// initialize will configure the HTTP server routes.
func (s *Server) initialize() {
	s.router = gin.Default()

	// Endpoint configurations
	api := s.router.Group(s.conf.Server.BasePath)
	// TODO: Add middleware handler to store authentication data in context for access in GraphQL resolvers.

	api.POST(s.conf.Server.QueryPath, graphql_resolvers.QueryHandler(s.auth, s.cache, s.db, s.grading, s.logger))
	api.GET(s.conf.Server.PlaygroundPath, graphql_resolvers.PlaygroundHandler(s.conf.Server.QueryPath))
}