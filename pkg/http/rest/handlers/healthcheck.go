package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"go.uber.org/zap"
)

// Healthcheck checks if the service is healthy.
// @Summary     Healthcheck for service liveness.
// @Description This endpoint is exposed to allow load balancers etc. to check the health of the service.
// @Tags        health healthcheck liveness
// @Id          healthcheck
// @Accept      json
// @Produce     json
// @Success     200 {object} model_rest.Success "message: healthy"
// @Failure     503 {object} model_rest.Error   "error message with any available details"
// @Router      /health [get]
func Healthcheck(logger *logger.Logger, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error

		if _, err = db.Execute(cassandra.HealthcheckQuery, nil); err != nil {
			msg := "Cassandra healthcheck failed"
			logger.Warn(msg, zap.Error(err))
			context.AbortWithStatusJSON(http.StatusServiceUnavailable, &model_rest.Error{Message: msg})
			return
		}

		context.JSON(http.StatusOK, &model_rest.Success{Message: "healthy"})
	}
}
