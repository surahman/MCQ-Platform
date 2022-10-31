-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- Question UDT that describes a single question with all it's answer options as well as the answer key for the question.
CREATE TYPE IF NOT EXISTS mcq_platform.question (
    description text,                               // Description that contains the text of the question.
    asset       text,                               // URI of an asset to be displayed with question.
    options     list<text>,                         // Available options for the question.
    answers     list<int>                           // Indices of the options that are correct answers in the question.
);

-- Quizzes table creation.
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
