package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Mock Auth interface stub generation.
//go:generate mockgen -destination=../mocks/mock_auth.go -package=mocks github.com/surahman/mcq-platform/pkg/auth Auth

// Auth is the interface through which the authorization operations can be accessed. Created to support mock testing.
type Auth interface {
	// HashPassword will take a plaintext string and generate a hashed representation of it.
	HashPassword(string) (string, error)

	// CheckPassword will take the plaintext and hashed passwords as input, in that order, and verify if they match.
	CheckPassword(string, string) error

	// GenerateJWT will create a valid JSON Web Token and return it in a JWT Authorization Response structure.
	GenerateJWT(string) (*model_rest.JWTAuthResponse, error)

	// ValidateJWT will take the JSON Web Token and validate it. It will extract and return the username and expiration
	// time (Unix timestamp) or an error if validation fails.
	ValidateJWT(string) (string, int64, error)

	// RefreshJWT will take a valid JSON Web Token, and if valid and expiring soon, issue a fresh valid JWT with the time
	// extended in JWT Authorization Response structure.
	RefreshJWT(string) (*model_rest.JWTAuthResponse, error)

	// RefreshThreshold returns the time before the end of the JSON Web Tokens validity interval that a JWT can be
	// refreshed in.
	RefreshThreshold() int64
}

// Check to ensure the Auth interface has been implemented.
var _ Auth = &authImpl{}

// authImpl implements the Auth interface and contains the logic for authorization functionality.
type authImpl struct {
	conf   *config
	logger *logger.Logger
}

// NewAuth will create a new Authorization configuration by loading it.
func NewAuth(fs *afero.Fs, logger *logger.Logger) (Auth, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system of logger supplied")
	}
	return newAuthImpl(fs, logger)
}

// newAuthImpl will create a new authImpl configuration and load it from disk.
func newAuthImpl(fs *afero.Fs, logger *logger.Logger) (a *authImpl, err error) {
	a = &authImpl{conf: newConfig(), logger: logger}
	if err = a.conf.Load(*fs); err != nil {
		a.logger.Error("failed to load Authorization configurations from disk", zap.Error(err))
		return nil, err
	}
	return
}

// HashPassword hashes a password using the Bcrypt algorithm to avoid plaintext storage.
func (a *authImpl) HashPassword(plaintext string) (hashed string, err error) {
	var bytes []byte
	if bytes, err = bcrypt.GenerateFromPassword([]byte(plaintext), a.conf.General.BcryptCost); err != nil {
		return
	}
	hashed = string(bytes)
	return
}

// CheckPassword checks a hashed password against a plaintext password using the Bcrypt algorithm.
func (a *authImpl) CheckPassword(hashed, plaintext string) (err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)); err != nil {
		return
	}
	return
}

// jwtClaim is used internally by the JWT generation and validation routines.
type jwtClaim struct {
	Username string `json:"username" yaml:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a payload consisting of the JWT with the username as well as expiration time.
func (a *authImpl) GenerateJWT(username string) (*model_rest.JWTAuthResponse, error) {
	claims := &jwtClaim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.conf.JWTConfig.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.conf.JWTConfig.ExpirationDuration) * time.Second).UTC()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.conf.JWTConfig.Key))
	if err != nil {
		return nil, err
	}

	authResponse := &model_rest.JWTAuthResponse{
		Token:     tokenString,
		Expires:   claims.ExpiresAt.Time.Unix(),
		Threshold: a.conf.JWTConfig.RefreshThreshold,
	}

	return authResponse, err
}

// ValidateJWT will validate a signed JWT and extracts the username from it.
func (a *authImpl) ValidateJWT(signedToken string) (string, int64, error) {
	claims := jwt.MapClaims{}
	if _, err := jwt.ParseWithClaims(signedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.conf.JWTConfig.Key), nil
	}); err != nil {
		return "", -1, err
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return "", -1, errors.New("token has expired")
	}
	if !claims.VerifyIssuer(a.conf.JWTConfig.Issuer, true) {
		return "", -1, errors.New("unauthorized issuer")
	}

	username, ok := claims["username"]
	if !ok {
		return "", -1, errors.New("username not found")
	}

	expiresAt, ok := claims["exp"]
	if !ok {
		return "", -1, errors.New("expiration time not found")
	}

	return username.(string), int64(expiresAt.(float64)), nil
}

// RefreshJWT will extend a valid JWT's lease by generating a fresh valid JWT.
func (a *authImpl) RefreshJWT(token string) (authResponse *model_rest.JWTAuthResponse, err error) {
	var username string
	if username, _, err = a.ValidateJWT(token); err != nil {
		return
	}
	if authResponse, err = a.GenerateJWT(username); err != nil {
		return
	}

	return
}

// RefreshThreshold is the seconds before expiration that a JWT can be refreshed in.
func (a *authImpl) RefreshThreshold() int64 {
	return a.conf.JWTConfig.RefreshThreshold
}
