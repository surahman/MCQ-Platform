package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"github.com/surahman/mcq-platform/pkg/validator"
	"go.uber.org/zap"
)

// RegisterUser will handle an HTTP request to create a user.
// @Summary     Register a user.
// @Description Creates a user account by inserting credentials into the database. A hashed password is stored.
// @Tags        user users register security
// @Id          registerUser
// @Accept      json
// @Produce     json
// @Param       user body     model_cassandra.UserAccount     true "Username, password, first and last name, email address of user"
// @Success     200  {object} model_rest.JWTAuthResponse "a valid JWT token for the new account"
// @Failure     400  {object} model_rest.Error "error message with any available details in payload"
// @Failure     409  {object} model_rest.Error "error message with any available details in payload"
// @Failure     500  {object} model_rest.Error "error message with any available details in payload"
// @Router      /user/register [post]
func RegisterUser(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var authToken *model_rest.JWTAuthResponse
		var user model_cassandra.UserAccount

		if err = context.ShouldBindJSON(&user); err != nil {
			context.JSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			context.Abort()
			return
		}

		if err = validator.ValidateStruct(&user); err != nil {
			context.JSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err.Error()})
			return
		}

		if user.Password, err = auth.HashPassword(user.Password); err != nil {
			logger.Error("failure hashing password", zap.Error(err))
			context.JSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
			context.Abort()
			return
		}

		if _, err = db.Execute(cassandra.CreateUserQuery, &model_cassandra.User{UserAccount: &user}); err != nil {
			context.JSON(err.(*cassandra.Error).Status, &model_rest.Error{Message: err.Error()})
			context.Abort()
			return
		}

		if authToken, err = auth.GenerateJWT(user.Username); err != nil {
			logger.Error("failure generating JWT after account creation", zap.Error(err))
			context.JSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
			context.Abort()
			return
		}

		context.JSON(http.StatusOK, authToken)
	}
}

// LoginUser validates login credentials and generates a JWT.
// @Summary     Login a user.
// @Description Logs in a user by validating credentials and returning a JWT.
// @Tags        user users login security
// @Id          loginUser
// @Accept      json
// @Produce     json
// @Param       user body     models.User     true "Username and password to register user with"
// @Success     200  {object} models.Response "JWT in the api-key"
// @Failure     400  {object} models.Response "error message with any available details in payload"
// @Failure     401  {object} models.Response "error message with any available details in payload"
// @Failure     404  {object} models.Response "error message with any available details in payload"
// @Failure     500  {object} models.Response "error message with any available details in payload"
// @Router      /user/login [post]
func LoginUser(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// LoginRefresh validates a JWT token and issues a fresh token.
// @Summary     Refresh a user's JWT by extending its expiration time.
// @Description Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time.
// @Tags        user users login refresh security
// @Id          loginRefresh
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       user body     models.User     true "Username and password to register user with"
// @Success     200  {object} models.Response "JWT in the api-key"
// @Failure     400  {object} models.Response "error message with any available details in payload"
// @Failure     401  {object} models.Response "error message with any available details in payload"
// @Failure     404  {object} models.Response "error message with any available details in payload"
// @Failure     500  {object} models.Response "error message with any available details in payload"
// @Router      /user/refresh [post]
func LoginRefresh(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}

// DeleteUser will mark a user as deleted in the database.
// @Summary     Delete a user.
// @Description Deletes a user stored in the database by marking it as deleted.
// @Tags        user users delete security
// @Id          deleteUser
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       user body     models.User     true "Username and password to register user"
// @Success     201  {object} models.Response "message with the registered username."
// @Failure     400  {object} models.Response "error message with any available details in payload"
// @Failure     409  {object} models.Response "error message with any available details in payload"
// @Failure     500  {object} models.Response "error message with any available details in payload"
// @Router      /user/delete [delete]
func DeleteUser(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
}
