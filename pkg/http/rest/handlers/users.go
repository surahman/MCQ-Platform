package http_handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/constants"
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
// @Param       user body     model_cassandra.UserAccount true "Username, password, first and last name, email address of user"
// @Success     200  {object} model_rest.JWTAuthResponse  "a valid JWT token for the new account"
// @Failure     400  {object} model_rest.Error            "error message with any available details in payload"
// @Failure     409  {object} model_rest.Error            "error message with any available details in payload"
// @Failure     500  {object} model_rest.Error            "error message with any available details in payload"
// @Router      /user/register [post]
func RegisterUser(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var authToken *model_rest.JWTAuthResponse
		var user model_cassandra.UserAccount

		if err = context.ShouldBindJSON(&user); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&user); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		if user.Password, err = auth.HashPassword(user.Password); err != nil {
			logger.Error("failure hashing password", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
			return
		}

		if _, err = db.Execute(cassandra.CreateUserQuery, &model_cassandra.User{UserAccount: &user}); err != nil {
			context.AbortWithStatusJSON(err.(*cassandra.Error).Status, &model_rest.Error{Message: err.Error()})
			return
		}

		if authToken, err = auth.GenerateJWT(user.Username); err != nil {
			logger.Error("failure generating JWT during account creation", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
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
// @Param       credentials body     model_cassandra.UserLoginCredentials true "Username and password to login with"
// @Success     200         {object} model_rest.JWTAuthResponse           "JWT in the api-key"
// @Failure     400         {object} model_rest.Error                     "error message with any available details in payload"
// @Failure     403         {object} model_rest.Error                     "error message with any available details in payload"
// @Failure     500         {object} model_rest.Error                     "error message with any available details in payload"
// @Router      /user/login [post]
func LoginUser(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var authToken *model_rest.JWTAuthResponse
		var loginRequest model_cassandra.UserLoginCredentials
		var dbResponse any

		if err = context.ShouldBindJSON(&loginRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&loginRequest); err != nil {
			context.JSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		if dbResponse, err = db.Execute(cassandra.ReadUserQuery, loginRequest.Username); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "invalid username or password"})
			return
		}

		truth := dbResponse.(*model_cassandra.User)
		if err = auth.CheckPassword(truth.Password, loginRequest.Password); err != nil || truth.IsDeleted {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "invalid username or password"})
			return
		}

		if authToken, err = auth.GenerateJWT(loginRequest.Username); err != nil {
			logger.Error("failure generating JWT during login", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
			return
		}

		context.JSON(http.StatusOK, authToken)
	}
}

// LoginRefresh validates a JWT token and issues a fresh token.
// @Summary     Refresh a user's JWT by extending its expiration time.
// @Description Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time. JWT must be expiring in under 60 seconds.
// @Tags        user users login refresh security
// @Id          loginRefresh
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} model_rest.JWTAuthResponse "A new valid JWT"
// @Failure     400 {object} model_rest.Error           "error message with any available details in payload"
// @Failure     403 {object} model_rest.Error           "error message with any available details in payload"
// @Failure     500 {object} model_rest.Error           "error message with any available details in payload"
// @Failure     510 {object} model_rest.Error           "error message with any available details in payload"
// @Router      /user/refresh [post]
func LoginRefresh(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra, authHeaderKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var freshToken *model_rest.JWTAuthResponse
		var username string
		var dbResponse any
		var expiresAt int64
		originalToken := context.GetHeader(authHeaderKey)

		if username, expiresAt, err = auth.ValidateJWT(originalToken); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: err.Error()})
			return
		}

		if dbResponse, err = db.Execute(cassandra.ReadUserQuery, username); err != nil {
			logger.Warn("failed to read user record for a valid JWT", zap.String("username", username), zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "please retry your request later"})
			return
		}

		if dbResponse.(*model_cassandra.User).IsDeleted {
			logger.Warn("attempt to refresh a JWT for a deleted user", zap.String("username", username))
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "invalid token"})
			return
		}

		// Do not refresh tokens that have more than a minute left to expire.
		if math.Abs(float64(time.Now().Unix()-expiresAt)) > float64(auth.RefreshThreshold()) {
			context.AbortWithStatusJSON(http.StatusNotExtended, &model_rest.Error{Message: "JWT is still valid for more than 60 seconds"})
			return
		}

		if freshToken, err = auth.GenerateJWT(username); err != nil {
			logger.Error("failure generating JWT during token refresh", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: err.Error()})
			return
		}

		context.JSON(http.StatusOK, freshToken)
	}
}

// DeleteUser will mark a user as deleted in the database.
// @Summary     Deletes a user. The user must supply their credentials as well as a confirmation message.
// @Description Deletes a user stored in the database by marking it as deleted. The user must supply their login credentials as well as complete the following confirmation message: "I understand the consequences, delete my user account USERNAME HERE"
// @Tags        user users delete security
// @Id          deleteUser
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       request body     model_rest.DeleteUserRequest true "The request payload for deleting an account"
// @Success     200     {object} model_rest.Success           "message with a confirmation of a deleted user account"
// @Failure     400     {object} model_rest.Error             "error message with any available details in payload"
// @Failure     409     {object} model_rest.Error             "error message with any available details in payload"
// @Failure     500     {object} model_rest.Error             "error message with any available details in payload"
// @Router      /user/delete [delete]
func DeleteUser(logger *logger.Logger, auth auth.Auth, db cassandra.Cassandra, authHeaderKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		var err error
		var deleteRequest model_rest.DeleteUserRequest
		var username string
		var dbResponse any
		jwt := context.GetHeader(authHeaderKey)

		// Get the delete request from the message body and validate it.
		if err = context.ShouldBindJSON(&deleteRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: err.Error()})
			return
		}

		if err = validator.ValidateStruct(&deleteRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "validation", Payload: err})
			return
		}

		// Validate the JWT and extract the username, compare the username against the deletion request login credentials.
		if username, _, err = auth.ValidateJWT(jwt); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: err.Error()})
			return
		}

		if username != deleteRequest.Username {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "invalid deletion request"})
			return
		}

		// Check confirmation message.
		if fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), username) != deleteRequest.Confirmation {
			context.AbortWithStatusJSON(http.StatusBadRequest, &model_rest.Error{Message: "incorrect or incomplete deletion request confirmation"})
			return
		}

		// Get user record.
		if dbResponse, err = db.Execute(cassandra.ReadUserQuery, username); err != nil {
			logger.Warn("failed to read user record during an account deletion request", zap.String("username", username), zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "please retry your request later"})
			return
		}
		truth := dbResponse.(*model_cassandra.User)

		// Check to make sure the account is not already deleted.
		if truth.IsDeleted {
			logger.Warn("attempt to delete an already deleted user account", zap.String("username", username))
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "user account is already deleted"})
			return
		}

		if err = auth.CheckPassword(truth.Password, deleteRequest.Password); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &model_rest.Error{Message: "invalid username or password"})
			return
		}

		// Mark account as deleted.
		if _, err = db.Execute(cassandra.DeleteUserQuery, username); err != nil {
			logger.Warn("failed to mark a user record as deleted", zap.String("username", username), zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &model_rest.Error{Message: "please retry your request later"})
			return
		}

		context.JSON(http.StatusOK, model_rest.Success{Message: "account successfully deleted"})
	}
}
