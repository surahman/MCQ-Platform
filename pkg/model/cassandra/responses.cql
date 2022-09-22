-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- Responses table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.responses (
    username text,                                      // Username of the test taker.
    quiz_id uuid,                                       // Taken quiz's id.
    score double,                                       // Score for this submission.
    responses frozen<list<list<int>>>,                  // Recorded responses for the submission.
    PRIMARY KEY ( (username, quiz_id))
);
CREATE INDEX responses_statistics_index ON mcq_platform.responses (quiz_id);


-- Insertion example.
INSERT INTO mcq_platform.responses (username, quiz_id, score, responses)
VALUES ('user-1', 8748e21a-3a00-11ed-a261-0242ac120002, 1.0, [[1,2,3],[4,5,6]])
IF NOT EXISTS;

-- Retrieval example for a user score.
SELECT * FROM mcq_platform.responses
WHERE username = 'user-1' AND quiz_id = 8748e21a-3a00-11ed-a261-0242ac120002;

-- Retrieval example statistics using index.
SELECT * FROM mcq_platform.responses
WHERE quiz_id = 8748e21a-3a00-11ed-a261-0242ac120002;
