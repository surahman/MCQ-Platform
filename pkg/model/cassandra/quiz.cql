-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- Question UDT that describes a single question with all it's answer options as well as the answer key for the question.
CREATE TYPE IF NOT EXISTS mcq_platform.question (
    description text,                               // Description that contains the text of the question.
    options     list<text>,                         // Available options for the question.
    answers     list<int>                           // Indices of the options that are correct answers in the question.
);

-- Quizzes table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.quizzes (
    quiz_id         uuid,                           // Unique identifier for the quiz.
    author          text,                           // Username of the quiz creator.
    title           text,                           // Description of the quiz.
    questions       frozen<list<frozen<question>>>, // A list of questions in the quiz.
    is_published    boolean,                        // Status indicating whether the quiz can be viewed or taken by other users.
    is_deleted      boolean,                        // Status indicating whether the quiz has been deleted.
    PRIMARY KEY ( (quiz_id), author )
);


-- Insertion example.
INSERT INTO mcq_platform.quizzes (quiz_id, author, title, questions, is_published)
VALUES (8748e21a-3a00-11ed-a261-0242ac120002,
        'user-1',
        'Quiz Title',
        [{description: 'Question Description', options: ['one', 'two', 'three'], answers: [1]}],
        false)
IF NOT EXISTS ;

-- Retrieval example.
SELECT * FROM mcq_platform.quizzes
WHERE quiz_id = 8748e21a-3a00-11ed-a261-0242ac120002;