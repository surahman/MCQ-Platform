package graphql_resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/surahman/mcq-platform/pkg/cassandra"
	"go.uber.org/zap"
)

// Healthcheck is the resolver for the healthcheck field.
func (r *queryResolver) Healthcheck(ctx context.Context) (string, error) {

	// Database health.
	if _, err := r.DB.Execute(cassandra.HealthcheckQuery, nil); err != nil {
		msg := "Cassandra healthcheck failed"
		r.Logger.Warn(msg, zap.Error(err))
		return "", errors.New(msg)
	}

	// Cache health.
	if err := r.Cache.Healthcheck(); err != nil {
		msg := "Redis healthcheck failed"
		r.Logger.Warn(msg, zap.Error(err))
		return "", errors.New(msg)
	}

	return "OK", nil
}
