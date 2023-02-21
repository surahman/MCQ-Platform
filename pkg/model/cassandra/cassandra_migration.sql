--liquibase formatted sql

--changeset surahman:1
--preconditions onFail:HALT onError:HALT
--comment: Users table contains the general user information and login credentials.
CREATE TABLE IF NOT EXISTS mcq_platform.users (
    account_id  text,
    username    text,
    password    text,
    first_name  text,
    last_name   text,
    email       text,
    is_deleted  boolean,
    PRIMARY KEY ( (username, account_id) )
);
--rollback DROP TABLE mcq_platform.users;

--changeset surahman:2
--preconditions onFail:HALT onError:HALT
--comment: Question UDT that describes a single question with all it's answer options as well as the answer key for the question.
CREATE TYPE IF NOT EXISTS mcq_platform.question (
    description text,                               // Description that contains the text of the question.
    asset       text,                               // URI of an asset to be displayed with question.
    options     list<text>,                         // Available options for the question.
    answers     list<int>                           // Indices of the options that are correct answers in the question.
    );
--rollback DROP TYPE mcq_platform.question;

--changeset surahman:3
--preconditions onFail:HALT onError:HALT
--comment: Quizzes table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.quizzes (
    quiz_id         uuid,                           // Unique identifier for the quiz.
    author          text,                           // Username of the quiz creator.
    title           text,                           // Description of the quiz.
    marking_type    text,                           // Marking type: none, negative, non-negative, or binary.
    questions       frozen<list<frozen<question>>>, // A list of questions in the quiz.
    is_published    boolean,                        // Status indicating whether the quiz can be viewed or taken by other users.
    is_deleted      boolean,                        // Status indicating whether the quiz has been deleted.
    PRIMARY KEY ( (quiz_id) )
);
--rollback DROP TABLE mcq_platform.quizzes;

--changeset surahman:4
--preconditions onFail:HALT onError:HALT
--comment: Responses table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.responses (
    username text,                                      // Username of the test taker.
    quiz_id uuid,                                       // Taken quiz's id.
    author text,                                        // Quiz author's username.
    score double,                                       // Score for this submission.
    responses frozen<list<list<int>>>,                  // Recorded responses for the submission.
    PRIMARY KEY ( (username, quiz_id) )
);
--rollback DROP TABLE mcq_platform.responses;

--changeset surahman:5
--preconditions onFail:HALT onError:HALT
--comment: Index on the responses table used to collate statistics.
CREATE INDEX responses_statistics_index ON mcq_platform.responses (quiz_id);
--rollback DROP INDEX mcq_platform.responses_statistics_index;
