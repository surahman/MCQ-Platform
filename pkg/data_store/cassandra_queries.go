package data_store

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
	conn := c.(*CassandraImpl)
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
	conn := c.(*CassandraImpl)
	input := params.(*model_cassandra.User)

	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.CreateUser,
		input.Username, input.AccountID, input.Password, input.FirstName, input.LastName, input.Email).ScanCAS(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to create input record",
			zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, err
	}

	if !applied {
		conn.logger.Error("username has already been taken",
			zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, errors.New("failed to create user record it already exists")
	}

	return nil, nil
}

// ReadUserQuery will read a user record from the users table.
// Param: pointer to the user struct containing the query parameters
// Return: address to a user record
func ReadUserQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*CassandraImpl)
	input := params.(*model_cassandra.User)
	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}}

	if err = conn.session.Query(model_cassandra.ReadUser, input.Username, input.AccountID).Scan(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to read user record",
			zap.String("username", input.Username), zap.String("account_id", input.AccountID), zap.Error(err))
	}

	return &resp, err
}

// DeleteUserQuery will mark a user record as deleted in the users table.
// Param: pointer to the user struct containing the query parameters
func DeleteUserQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*CassandraImpl)
	input := params.(*model_cassandra.User)

	resp := model_cassandra.User{UserAccount: &model_cassandra.UserAccount{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.DeleteUser, input.Username, input.AccountID).ScanCAS(
		&resp.Username, &resp.AccountID, &resp.Email, &resp.FirstName, &resp.IsDeleted, &resp.LastName, &resp.Password); err != nil {
		conn.logger.Error("failed to create user record",
			zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, err
	}

	if !applied {
		conn.logger.Error("failed to mark user record as deleted",
			zap.Strings("Account info:", []string{input.Username, input.AccountID}), zap.Error(err))
		return nil, errors.New("failed to create user record it already exists")
	}

	return nil, err
}

// -----   Quizzes Table Queries   -----

// -----   Responses Table Queries   -----

// CreateResponseQuery will insert a response record into the responses table.
// Param: pointer to the user struct containing the response parameters
func CreateResponseQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*CassandraImpl)
	input := params.(*model_cassandra.Response)

	resp := model_cassandra.Response{QuizResponse: &model_cassandra.QuizResponse{}} // Discarded, only used as container for Cassandra response.
	applied := false
	if applied, err = conn.session.Query(model_cassandra.CreateResponse,
		input.Username, input.QuizID, input.Score, input.Responses).ScanCAS(
		&resp.Username, &resp.QuizID, &resp.Responses, &resp.Score); err != nil {
		conn.logger.Error("failed to create input record",
			zap.Strings("Response info:", []string{input.Username, input.QuizID.String()}), zap.Error(err))
		return nil, err
	}

	if !applied {
		conn.logger.Error("user has already taken this quiz",
			zap.Strings("Response info:", []string{input.Username, input.QuizID.String()}), zap.Error(err))
		return nil, errors.New("user has already submitted a response for this quiz")
	}

	return nil, nil
}

// ReadResponseQuery will read a response record from the responses table.
// Param: pointer to the response struct containing the query parameters
// Return: address to a response record
func ReadResponseQuery(c Cassandra, params any) (response any, err error) {
	conn := c.(*CassandraImpl)
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
	conn := c.(*CassandraImpl)
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
