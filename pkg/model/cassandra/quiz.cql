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
    PRIMARY KEY ( (quiz_id), author )
);