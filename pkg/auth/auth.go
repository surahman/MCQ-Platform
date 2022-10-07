package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Mock Auth interface stub generation.
//go:generate mockgen -destination=../mocks/mock_auth.go -package=mocks github.com/surahman/mcq-platform/pkg/auth Auth

// Auth is the interface through which the cluster can be accessed. Created to support mock testing.
type Auth interface {
	HashPassword(string) (string, error)
	CheckPassword(string, string) error
}

// Check to ensure the Cassandra interface has been implemented.
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

// GenerateJWT creates a JWT with a payload consisting of the username.
func (a *authImpl) GenerateJWT(username string) (tokenString string, err error) {
	claims := &jwtClaim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.conf.JWTConfig.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Unix(a.conf.JWTConfig.ExpirationDuration, 0)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(a.conf.JWTConfig.Key)

	return
}

// ValidateJWT will validate a signed JWT.
func (a *authImpl) ValidateJWT(signedToken string) error {
	token, err := jwt.ParseWithClaims(signedToken, &jwtClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.conf.JWTConfig.Key), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*jwtClaim)
	if !ok {
		return errors.New("could not parse claims")
	}

	if claims.VerifyExpiresAt(time.Now(), true) {
		return errors.New("token has expired")
	}
	if claims.VerifyIssuer(a.conf.JWTConfig.Issuer, true) {
		return errors.New("unauthorized issuer")
	}

	return err
}
