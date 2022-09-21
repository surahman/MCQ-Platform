-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- A users answer card to a quiz.
-- The rows indices are the question numbers and the columns indices are the selected option numbers.
CREATE TYPE IF NOT EXISTS mcq_platform.answers (
    response list<frozen<list<int>>>            // [Question Number][List of options indices selected].
);

-- Responses table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.responses (
    username text,                              // Username of the test taker.
    quiz_id uuid,                               // Taken quiz's id.
    score double,                               // Score for this submission.
    responses frozen<answers>,                  // Recorded responses for the submission.
    PRIMARY KEY ( (username, quiz_id), score )
)
WITH CLUSTERING ORDER BY (score DESC);