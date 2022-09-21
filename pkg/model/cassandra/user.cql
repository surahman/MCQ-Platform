-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- User table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.users (
    account_id  text,
    username    text,
    password    text,
    first_name  text,
    last_name   text,
    email       text,
    is_deleted  boolean,
    PRIMARY KEY ( (username, account_id), email )
);

-- Insertion example.
CONSISTENCY ALL;
INSERT INTO mcq_platform.users (username, account_id, password, first_name, last_name, email, is_deleted)
VALUES ('username-1', 'user-1-account_id', 'password-1', 'first-name-user-1', 'last-name-user-1', 'user_1@email.com', false)
IF NOT EXISTS;
CONSISTENCY QUORUM;

-- Retrieval example.
CONSISTENCY ALL;
SELECT * FROM mcq_platform.users
WHERE username = 'username-1' AND account_id = 'user-1-account_id';
CONSISTENCY QUORUM;
