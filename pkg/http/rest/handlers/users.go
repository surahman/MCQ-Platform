package http_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterUser will handle an HTTP request to create a user.
// @Summary     Register a user.
// @Description Creates a user account by inserting credentials into the database. A hashed password is stored.
// @Tags        user users register security
// @Id          registerUser
// @Accept      json
// @Produce     json
// @Param       user body     models.User     true "Username and password to register user"
// @Success     201  {object} models.Response "message with the registered username."
// @Failure     400  {object} models.Response "error message with any available details in payload"
// @Failure     409  {object} models.Response "error message with any available details in payload"
// @Failure     500  {object} models.Response "error message with any available details in payload"
// @Router      /user/register [post]
func RegisterUser(context *gin.Context) {
	context.JSON(http.StatusNotImplemented, nil)
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
