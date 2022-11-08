package rest

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
	_ "github.com/surahman/mcq-platform/docs" // Swaggo generated Swagger documentation
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/http/rest/handlers"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/redis"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Format and generate Swagger UI files using makefile.
//go:generate make -C ../../../ swagger

// Server is the HTTP REST server.
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
	// Indicate to bootstrapping thread to wait for completion.
	defer s.wg.Done()

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
	s.logger.Info("Shutting down REST server...", zap.Duration("waiting", time.Duration(s.conf.Server.ShutdownDelay)*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.conf.Server.ShutdownDelay)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Panic("Failed to shutdown REST server", zap.Error(err))
	}

	// 5 second wait to exit.
	<-ctx.Done()

	s.logger.Info("REST server exited")
}

// initialize will configure the HTTP server routes.
func (s *Server) initialize() {
	s.router = gin.Default()

	// @title                      Multiple Choice Question Platform.
	// @version                    1.1.0
	// @description                Multiple Choice Question Platform API.
	// @description                This application supports the creation, managing, marking, viewing, retrieving stats, and scores of quizzes.
	//
	// @schemes                    http
	// @host                       localhost:44243
	// @BasePath                   /api/rest/v1
	//
	// @accept                     json
	// @produce                    json
	//
	// @contact.name               Saad Ur Rahman
	// @contact.url                https://www.linkedin.com/in/saad-ur-rahman/
	// @contact.email              saad.ur.rahman@gmail.com
	//
	// @license.name               GPL-3.0
	// @license.url                https://opensource.org/licenses/GPL-3.0
	//
	// @securityDefinitions.apikey ApiKeyAuth
	// @in                         header
	// @name                       Authorization

	s.router.GET(s.conf.Server.SwaggerPath, ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Endpoint configurations
	authMiddleware := http_handlers.AuthMiddleware(s.auth, s.conf.Authorization.HeaderKey)
	api := s.router.Group(s.conf.Server.BasePath)

	api.GET("/health", http_handlers.Healthcheck(s.logger, s.db, s.cache))

	userGroup := api.Group("/user")
	userGroup.POST("/register", http_handlers.RegisterUser(s.logger, s.auth, s.db))
	userGroup.POST("/login", http_handlers.LoginUser(s.logger, s.auth, s.db))
	userGroup.Use(authMiddleware).POST("/refresh", http_handlers.LoginRefresh(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))
	userGroup.Use(authMiddleware).DELETE("/delete", http_handlers.DeleteUser(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))

	scoreGroup := api.Group("/score").Use(authMiddleware)
	scoreGroup.GET("/test/:quiz_id", http_handlers.GetScore(s.logger, s.auth, s.db))
	scoreGroup.GET("/stats/:quiz_id", http_handlers.GetStats(s.logger, s.auth, s.db))
	scoreGroup.GET("/stats-paged/:quiz_id", http_handlers.GetStatsPage(s.logger, s.auth, s.db))

	quizGroup := api.Group("/quiz").Use(authMiddleware)
	quizGroup.GET("/view/:quiz_id", http_handlers.ViewQuiz(s.logger, s.auth, s.db, s.cache))
	quizGroup.POST("/create", http_handlers.CreateQuiz(s.logger, s.auth, s.db))
	quizGroup.PATCH("/update/:quiz_id", http_handlers.UpdateQuiz(s.logger, s.auth, s.db))
	quizGroup.DELETE("/delete/:quiz_id", http_handlers.DeleteQuiz(s.logger, s.auth, s.db, s.cache))
	quizGroup.PATCH("/publish/:quiz_id", http_handlers.PublishQuiz(s.logger, s.auth, s.db, s.cache))
	quizGroup.POST("/take/:quiz_id", http_handlers.TakeQuiz(s.logger, s.auth, s.db, s.cache, s.grading))
}
