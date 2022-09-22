package model_cassandra

const (
	// CreateKeyspace sets up the keyspace in the cluster.
	// String Param 1: keyspace name
	// String Param 2: replication strategy
	// String Param 4: replication factor
	CreateKeyspace = `CREATE KEYSPACE IF NOT EXISTS $s WITH replication = {'class' : '$s', 'replication_factor' : $d};`

	// CreateUsersTable creates the Users table.
	// String Param 1: keyspace name
	CreateUsersTable = `CREATE TABLE IF NOT EXISTS %s.users (
    account_id  text,
    username    text,
    password    text,
    first_name  text,
    last_name   text,
    email       text,
    is_deleted  boolean,
    PRIMARY KEY ( (username, account_id), email )
);`

	// CreateQuestionUDT creates the Question UDT that is used by the Questions table. CreateQuestionUDT must be called after this statement.
	// String Param 1: keyspace name
	CreateQuestionUDT = `CREATE TYPE IF NOT EXISTS $s.question (
    description text,
    options     list<text>,
    answers     list<int>
);`

	// CreateQuizzesTable creates the Quizzes table. CreateQuestionUDT must be called before this statement.
	// String Param 1: keyspace name
	CreateQuizzesTable = `CREATE TABLE IF NOT EXISTS %s.quizzes (
    quiz_id         uuid,                           // Unique identifier for the quiz.
    author          text,                           // Username of the quiz creator.
    title           text,                           // Description of the quiz.
    questions       frozen<list<frozen<question>>>, // A list of questions in the quiz.
    is_published    boolean,                        // Status indicating whether the quiz can be viewed or taken by other users.
    is_deleted      boolean,                        // Status indicating whether the quiz has been deleted.
    PRIMARY KEY ( (quiz_id), author )
);`

	// CreateResponsesTable creates the Responses table. CreateResponsesIndex must be called after this statement.
	// String Param 1: keyspace name
	CreateResponsesTable = `CREATE TABLE IF NOT EXISTS %s.responses (
    username text,
    quiz_id uuid,
    score double,
    responses frozen<list<list<int>>>,
    PRIMARY KEY ( (username, quiz_id))
);`

	// CreateResponsesIndex creates an index on the Responses table to support looking up statistics based on the quiz_id.
	// CreateResponsesTable must be called before this statement.
	// Param 1: keyspace name
	CreateResponsesIndex = `CREATE INDEX responses_statistics_index ON %s.responses (quiz_id);`

	// -----   Users Table Queries   -----

	// CreateUser inserts a new user row into the Users table if it does not already exist.
	// String Param 1: keyspace name
	// Query Params: username, account_id, password, first_name, last_name, email, is_deleted
	CreateUser = `INSERT INTO %s.users (username, account_id, password, first_name, last_name, email, is_deleted)
VALUES (?, ?, ?, ?, ?, ?, ?)
IF NOT EXISTS;`

	// ReadUser retrieves a user account row from the Users table.
	// String Param 1: keyspace name
	// Query Params: username, account_id
	ReadUser = `SELECT * FROM %s.users WHERE username = ? AND account_id = ?;`

	// DeleteUser will mark a users account as deleted.
	// String Param 1: keyspace name
	// Query Params: username, account_id
	DeleteUser = `UPDATE %s.users
SET is_deleted = true
WHERE username = ? AND account_id = ? IF EXISTS;`

	// -----   Quizzes Table Queries   -----

	// -----   Responses Table Queries   -----

)
