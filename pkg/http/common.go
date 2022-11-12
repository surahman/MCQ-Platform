package http

import (
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/redis"
)

// GetQuiz will make a cache call for the quiz. Upon a cache miss it will call the database for the quiz and then load it into
// the cache.
func GetQuiz(quizId gocql.UUID, db cassandra.Cassandra, cache redis.Redis) (*model_cassandra.Quiz, error) {
	var err error
	var quiz model_cassandra.Quiz
	var response any

	// Cache call.
	err = cache.Get(quizId.String(), &quiz)

	// Cache miss:
	// [1] Get quiz record from database.
	// [2] Place quiz in cache. Log but do not propagate errors to user on cache set failures.
	if err != nil {
		// Get quiz record from database.
		if response, err = db.Execute(cassandra.ReadQuizQuery, quizId); err != nil {
			return nil, err
		}
		quiz = *response.(*model_cassandra.Quiz)

		// Only place quiz in cache if it is published and not deleted. Set method will log errors.
		if quiz.IsPublished && !quiz.IsDeleted {
			_ = cache.Set(quizId.String(), &quiz)
		}
	}

	return &quiz, nil
}
