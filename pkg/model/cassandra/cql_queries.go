package model_cassandra

const (
	// CreateKeyspace sets up the keyspace in the cluster.
	// String Param 1: keyspace name
	// String Param 2: replication strategy
	// String Param 4: replication factor
	CreateKeyspace = `CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class' : '%s', 'replication_factor' : %d};`

	// CreateUsersTable creates the Users table.
	CreateUsersTable = `CREATE TABLE IF NOT EXISTS users (
    account_id  text,
    username    text,
    password    text,
    first_name  text,
    last_name   text,
    email       text,
    is_deleted  boolean,
    PRIMARY KEY ( (username, account_id) )
);`

	// CreateQuestionUDT creates the Question UDT that is used by the Questions table. CreateQuestionUDT must be called after this statement.
	CreateQuestionUDT = `CREATE TYPE IF NOT EXISTS question (
    description text,
    asset       text,
    options     list<text>,
    answers     list<int>
);`

	// CreateQuizzesTable creates the Quizzes table. CreateQuestionUDT must be called before this statement.
	CreateQuizzesTable = `CREATE TABLE IF NOT EXISTS quizzes (
    quiz_id         uuid,
    author          text,
    title           text,
    marking_type    text,
    questions       frozen<list<frozen<question>>>,
    is_published    boolean,
    is_deleted      boolean,
    PRIMARY KEY ( (quiz_id) )
);`

	// CreateResponsesTable creates the Responses table. CreateResponsesIndex must be called after this statement.
	CreateResponsesTable = `CREATE TABLE IF NOT EXISTS responses (
    username text,
    quiz_id uuid,
    score double,
    responses frozen<list<list<int>>>,
    PRIMARY KEY ( (username, quiz_id) )
);`

	// CreateResponsesIndex creates an index on the Responses table to support looking up statistics based on the quiz_id.
	// CreateResponsesTable must be called before this statement.
	CreateResponsesIndex = `CREATE INDEX responses_statistics_index ON responses (quiz_id);`

	// -----   Users Table Queries   -----

	// CreateUser inserts a new user record into the Users table if it does not already exist.
	// Query Params: username, account_id, password, first_name, last_name, email
	CreateUser = `INSERT INTO users (username, account_id, password, first_name, last_name, email, is_deleted)
VALUES (?, ?, ?, ?, ?, ?, false)
IF NOT EXISTS;`

	// ReadUser retrieves a user account record from the Users table.
	// Query Params: username, account_id
	ReadUser = `SELECT * FROM users WHERE username = ? AND account_id = ?;`

	// DeleteUser will mark a users account as deleted.
	// Query Params: username, account_id
	DeleteUser = `UPDATE users
SET is_deleted = true
WHERE username = ? AND account_id = ? IF EXISTS;`

	// -----   Quizzes Table Queries   -----

	// CreateQuiz inserts a new Quiz record into the Quizzes table if it does not already exist.
	// Query Params: quiz_id, author, title, questions, marking_type, is_published, is_deleted
	CreateQuiz = `INSERT INTO quizzes (quiz_id, author, title, questions, marking_type, is_published, is_deleted)
VALUES (?, ?, ?, ?, ?, false, false)
IF NOT EXISTS ;`

	// ReadQuiz retrieves a Quiz record from the Quizzes table.
	// Query Params: quiz_id,
	ReadQuiz = `SELECT * FROM quizzes WHERE quiz_id = ?;`

	// UpdateQuiz updates a Quiz record in the Quizzes table if it is not published.
	// Query Params: title, questions, marking_type, quiz_id
	UpdateQuiz = `UPDATE quizzes
SET title = ?, questions = ?, marking_type = ?
WHERE quiz_id = ? IF is_published = false;`

	// DeleteQuiz marks a Quiz record as deleted in the Quizzes table.
	// Query Params: quiz_id
	DeleteQuiz = `UPDATE quizzes
SET is_deleted = true
WHERE quiz_id = ? IF EXISTS;`

	// PublishQuiz marks a Quiz record as published in the Quizzes table.
	// Query Params: quiz_id
	PublishQuiz = `UPDATE quizzes
SET is_published = true
WHERE quiz_id = ? IF EXISTS;`

	// -----   Responses Table Queries   -----

	// CreateResponse inserts a new Response record into the Quizzes table if it does not already exist.
	// Query Params: username, quiz_id, score, responses
	CreateResponse = `INSERT INTO responses (username, quiz_id, score, responses)
VALUES (?, ?, ?, ?)
IF NOT EXISTS;`

	// ReadResponse retrieves a Response record from the Responses table.
	// Query Params: username, quiz_id
	ReadResponse = `SELECT * FROM responses WHERE username = ? AND quiz_id = ?;`

	// ReadResponseStatistics retrieves all Response statistics for a given Quiz from the Responses table.
	// Query Params: quiz_id
	ReadResponseStatistics = `SELECT * FROM responses WHERE quiz_id = ?;`
)
