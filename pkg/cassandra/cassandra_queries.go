package cassandra

import (
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"go.uber.org/zap"
)

// truncateTableQuery will empty a table.
// param: name of the table
func truncateTableQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	tableName := params.(string)
	if err = conn.session.Query(fmt.Sprintf("TRUNCATE TABLE %s;", tableName)).Exec(); err != nil {
		conn.logger.Error("failed to truncate table", zap.String("table name", tableName), zap.Error(err))
	}

	return
}

// -----   Users Table Queries   -----

// CreateUserQuery will insert a user record into the users table.
// Param: pointer to the user struct containing the query parameters
func CreateUserQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.User)

	// Create hash of username using Blake2b 256 hashing algorithm.
	input.AccountID = blake2b256(input.Username)

	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.CreateUser,
		input.Username, input.AccountID, input.Password, input.FirstName, input.LastName, input.Email).ScanCAS(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to create input record",
			zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, NewError(err.Error()).internalError()
	}

	if !applied {
		msg := "failed to create user record it already exists"
		conn.logger.Error(msg, zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, NewError(msg).conflictError()
	}

	return nil, nil
}

// ReadUserQuery will read a user record from the users table.
// Param: username string to lookup
// Return: address to a user record
func ReadUserQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}}
	username := params.(string)

	// Create hash of username using Blake2b 256 hashing algorithm.
	accountID := blake2b256(username)

	if err = conn.session.Query(model_cassandra.ReadUser, username, accountID).Scan(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to read user record",
			zap.String("username", username), zap.String("account_id", accountID), zap.Error(err))
		return nil, NewError("user not found").notFoundError()
	}

	return &resp, err
}

// DeleteUserQuery will mark a user record as deleted in the users table.
// Param: pointer to the user struct containing the query parameters
func DeleteUserQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	username := params.(string)

	// Create hash of username using Blake2b 256 hashing algorithm.
	accountID := blake2b256(username)

	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.DeleteUser, username, accountID).ScanCAS(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to create user record",
			zap.Strings("Account info:", []string{username, accountID}), zap.Error(err))
		return nil, NewError(err.Error()).internalError()
	}

	if !applied {
		msg := "failed to mark user record as deleted"
		conn.logger.Error(msg, zap.Strings("Account info:", []string{username, accountID}), zap.Error(err))
		return nil, NewError("user not found").notFoundError()
	}

	return nil, err
}

// -----   Quizzes Table Queries   -----

// CreateQuizQuery will create a quiz record in the quizzes table.
// Param: pointer to the quiz struct containing the query parameters
func CreateQuizQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.Quiz)

	resp := model_cassandra.Quiz{QuizCore: &model_cassandra.QuizCore{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.CreateQuiz,
		input.QuizID, input.Author, input.Title, input.Questions, input.MarkingType, input.IsPublished, input.IsDeleted).ScanCAS(
		&resp.QuizID, &resp.Author, &resp.IsDeleted, &resp.IsPublished, &resp.MarkingType, &resp.Questions, &resp.Title); err != nil {
		conn.logger.Error("failed to create quiz record",
			zap.Strings("Quiz info:", []string{input.QuizID.String(), input.Author}), zap.Error(err))
		return nil, err
	}

	if !applied {
		msg := "failed to create quiz with id, it already exists"
		conn.logger.Error(msg, zap.Strings("Quiz info:", []string{input.QuizID.String(), input.Author}), zap.Error(err))
		return nil, errors.New(msg)
	}

	return nil, nil
}

// ReadQuizQuery will read a quiz record from the quizzes table.
// Param: quiz id
// Return: address to a quiz record
func ReadQuizQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(gocql.UUID)
	resp := model_cassandra.Quiz{QuizCore: &model_cassandra.QuizCore{}}

	if err = conn.session.Query(model_cassandra.ReadQuiz, input).Scan(
		&resp.QuizID, &resp.Author, &resp.IsDeleted, &resp.IsPublished, &resp.MarkingType, &resp.Questions, &resp.Title); err != nil {
		conn.logger.Error("failed to read quiz record", zap.String("Quiz info:", input.String()), zap.Error(err))
		return nil, NewError("quiz not found").notFoundError()
	}

	return &resp, nil
}

// UpdateQuizQuery will update a quiz record in the quizzes table.
// Param: pointer to the quiz struct containing the query parameters
func UpdateQuizQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.Quiz)
	resp := struct {
		isPublished bool
	}{}

	applied := false
	if applied, err = conn.session.Query(model_cassandra.UpdateQuiz, input.Title, input.Questions, input.MarkingType, input.QuizID).ScanCAS(
		&resp.isPublished); err != nil {
		conn.logger.Error("failed to update quiz record", zap.Strings("Quiz info:", []string{input.QuizID.String(), input.Author}), zap.Error(err))
		return nil, err
	}

	if !applied {
		msg := "failed to update quiz, either it does not exist or is already published"
		conn.logger.Error(msg, zap.Strings("Quiz info:", []string{input.Author, input.QuizID.String()}), zap.Error(err))
		return nil, errors.New(msg)
	}

	return nil, err
}

// DeleteQuizQuery will mark a quiz record as deleted in the quizzes table.
// Param: quiz id
func DeleteQuizQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.QuizDelPubRequest)
	resp := struct {
		author string
	}{}

	applied := false
	if applied, err = conn.session.Query(model_cassandra.DeleteQuiz, input.QuizID, input.Username).ScanCAS(&resp.author); err != nil {
		conn.logger.Error("failed to delete quiz record", zap.Strings("Quiz info:", []string{input.QuizID.String(), input.Username}), zap.Error(err))
		return nil, NewError(err.Error()).internalError()
	}

	if !applied {
		msg := "failed to delete quiz record"
		conn.logger.Error(msg, zap.Strings("Quiz info:", []string{input.QuizID.String(), input.Username}), zap.Error(err))
		return nil, NewError(msg).forbiddenError()
	}

	return nil, err
}

// PublishQuizQuery will mark a quiz record as published in the quizzes table.
// Param: quiz id
func PublishQuizQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(gocql.UUID)
	resp := struct {
		isDeleted bool
	}{}

	applied := false
	if applied, err = conn.session.Query(model_cassandra.PublishQuiz, input).ScanCAS(&resp.isDeleted); err != nil {
		conn.logger.Error("failed to publish quiz record", zap.String("Quiz info:", input.String()), zap.Error(err))
		return nil, NewError("unable to publish quiz").internalError()
	}

	if !applied {
		msg := "failed to find or publish quiz record"
		conn.logger.Error(msg, zap.String("Quiz info:", input.String()), zap.Error(err))
		return nil, NewError(msg).notFoundError()
	}

	return nil, err
}

// -----   Responses Table Queries   -----

// CreateResponseQuery will insert a response record into the responses table.
// Param: pointer to the user struct containing the response parameters
func CreateResponseQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.Response)

	resp := model_cassandra.Response{QuizResponse: &model_cassandra.QuizResponse{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.CreateResponse,
		input.Username, input.QuizID, input.Score, input.Responses).ScanCAS(
		&resp.Username, &resp.QuizID, &resp.Responses, &resp.Score); err != nil {
		conn.logger.Error("failed to create response record",
			zap.Strings("Response info:", []string{input.Username, input.QuizID.String()}), zap.Error(err))
		return nil, err
	}

	if !applied {
		msg := "failed to record response, user has already taken this quiz"
		conn.logger.Error(msg, zap.Strings("Response info:", []string{input.Username, input.QuizID.String()}), zap.Error(err))
		return nil, errors.New(msg)
	}

	return nil, nil
}

// ReadResponseQuery will read a response record from the responses table.
// Param: pointer to the response struct containing the query parameters
// Return: address to a response record
func ReadResponseQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(*model_cassandra.Response)
	resp := model_cassandra.Response{QuizResponse: &model_cassandra.QuizResponse{}}

	if err = conn.session.Query(model_cassandra.ReadResponse, input.Username, input.QuizID).Scan(
		&resp.Username, &resp.QuizID, &resp.Responses, &resp.Score); err != nil {
		conn.logger.Error("failed to read response record",
			zap.Strings("Response info:", []string{input.Username, input.QuizID.String()}), zap.Error(err))
	}

	return &resp, err
}

// ReadResponseStatisticsQuery will read all response records from the responses table corresponding to a Quiz ID.
// Param: QuizID gocql UUID
// Return: address to slice of responses
func ReadResponseStatisticsQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*cassandraImpl)
	input := params.(gocql.UUID)
	var results []*model_cassandra.Response

	scanner := conn.session.Query(model_cassandra.ReadResponseStatistics, input).Iter().Scanner()
	defer func(scanner gocql.Scanner) {
		err := scanner.Err()
		if err != nil {
			conn.logger.Error("failed to close scanner whilst reading response statistics",
				zap.String("quiz_id", input.String()), zap.Error(err))
		}
	}(scanner)

	for scanner.Next() {
		row := model_cassandra.Response{QuizResponse: &model_cassandra.QuizResponse{}}
		err = scanner.Scan(&row.Username, &row.QuizID, &row.Responses, &row.Score)
		if err != nil {
			conn.logger.Error("failed to read row in response statistics",
				zap.String("quiz_id", input.String()), zap.Error(err))
		}
		results = append(results, &row)
	}

	return results, err
}
