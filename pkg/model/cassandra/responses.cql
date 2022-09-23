-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- Responses table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.responses (
    username text,                                      // Username of the test taker.
    quiz_id uuid,                                       // Taken quiz's id.
    score double,                                       // Score for this submission.
    responses frozen<list<list<int>>>,                  // Recorded responses for the submission.
    PRIMARY KEY ( (username, quiz_id) )
);
CREATE INDEX responses_statistics_index ON mcq_platform.responses (quiz_id);
