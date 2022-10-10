package rest

import (
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/grading"
	"github.com/surahman/mcq-platform/pkg/logger"
)

// HttpRest is the HTTP REST server.
type HttpRest struct {
	conf      *config
	auth      *auth.Auth
	cassandra *cassandra.Cassandra
	logger    *logger.Logger
	grading   *grading.Grading
}

func NewRESTServer(fs *afero.Fs, auth *auth.Auth, cassandra *cassandra.Cassandra,
	logger *logger.Logger, grading *grading.Grading) (server *HttpRest, err error) {
	// Load configurations.
	conf := newConfig()
	if err = conf.Load(*fs); err != nil {
		return
	}

	return &HttpRest{
			conf:      conf,
			auth:      auth,
			cassandra: cassandra,
			logger:    logger,
			grading:   grading,
		},
		err
}
