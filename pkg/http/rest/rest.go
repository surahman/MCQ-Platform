package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/http/rest/handlers"
	"github.com/surahman/mcq-platform/pkg/logger"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// HttpRest is the HTTP REST server.
type HttpRest struct {
	auth      auth.Auth
	cassandra cassandra.Cassandra
	grading   grading.Grading
	conf      *config
	logger    *logger.Logger
	router    *gin.Engine
}

// NewRESTServer will create a new REST server instance in a non-running state.
func NewRESTServer(fs *afero.Fs, auth auth.Auth, cassandra cassandra.Cassandra, grading grading.Grading,
	logger *logger.Logger) (server *HttpRest, err error) {
	// Load configurations.
	conf := newConfig()
	if err = conf.Load(*fs); err != nil {
		return
	}

	return &HttpRest{
			conf:      conf,
			auth:      auth,
			cassandra: cassandra,
			grading:   grading,
			logger:    logger,
		},
		err
}

// Run brings the HTTP service up.
func (s *HttpRest) Run() {
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
}

// initialize will configure the HTTP server routes.
func (s *HttpRest) initialize() {
	s.router = gin.Default()

	// @title                      Multiple Choice Question Platform.
	// @version                    1.0.1
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
	api := s.router.Group(s.conf.Server.BasePath)

	userGroup := api.Group("/user")
	userGroup.POST("/register", http_handlers.RegisterUser(s.logger, s.auth))
	userGroup.POST("/login", http_handlers.LoginUser)
	userGroup.POST("/refresh", http_handlers.LoginRefresh).Use(http_handlers.AuthMiddleware(s.auth))
	userGroup.DELETE("/delete", http_handlers.DeleteUser).Use(http_handlers.AuthMiddleware(s.auth))

	scoreGroup := api.Group("/score").Use(http_handlers.AuthMiddleware(s.auth))
	scoreGroup.GET("/test/:test_id", http_handlers.GetScore)
	scoreGroup.GET("/stats/:test_id", http_handlers.GetStats)

	quizGroup := api.Group("/quiz").Use(http_handlers.AuthMiddleware(s.auth))
	quizGroup.GET("/view/:test_id", http_handlers.ViewQuiz)
	quizGroup.POST("/create", http_handlers.CreateQuiz)
	quizGroup.PUT("/update/:test_id", http_handlers.UpdateQuiz)
	quizGroup.DELETE("/delete/:test_id", http_handlers.DeleteQuiz)
	quizGroup.PUT("/publish/:test_id", http_handlers.PublishQuiz)
	quizGroup.POST("/take/:test_id", http_handlers.TakeQuiz)
}
