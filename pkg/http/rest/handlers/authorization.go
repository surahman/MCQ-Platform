package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
)

// AuthMiddleware is the middleware that checks whether a JWT is valid and can access an endpoint.
func AuthMiddleware(auth auth.Auth) gin.HandlerFunc {
	handler := func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(http.StatusUnauthorized, "request does not contain an access token")
			context.Abort()
			return
		}
		if _, err := auth.ValidateJWT(tokenString); err != nil {
			context.JSON(http.StatusUnauthorized, err.Error())
			context.Abort()
			return
		}
		context.Next()
	}
	return handler
}
