# Apache Cassandra

## Table of contents

- [Apache Cassandra](#apache-cassandra)
    - [Case Study and Justification](#case-study-and-justification)
    - [Users Table Schema](#users-table-schema)
        - [User](#user)
        - [User Account](#user-account)
        - [CQL Query](#cql-query)
    - [Quizzes Table Schema](#quizzes-table-schema)
        - [User Defined Type(s)](#user-defined-types)
            - [Question](#question)
        - [Quizzes](#quizzes)
        - [Quiz Core](#quiz-core)
        - [CQL Query](#cql-query)
    - [Responses Table Schema](#responses-table-schema)
        - [Quizzes](#quizzes)
        - [CQL Query](#cql-query)
 
<br/>

## Case Study and Justification

A NoSQL database is ideal for this application for the following reasons:
* Eventual consistency is adequate for everything except the `User` table.
* Reads are more frequent than writes.
* Data access patterns do not require table joins.
* Data de-normalization is minimal or non-existent.
* Foreign key integrity in the tables is maintained by the service logic.
* The `Quiz` table updates are only conducted by a single user. Once a record is marked as published it becomes 
  read-only with updates to mark it as deleted.
* The `Score` table is read and write only with no updates permitted.
* The `User` table is read and write only with the only updates allowed to delete an account or update fields that are
  not the primary key (`username`).

As such, Apache Cassandra fits these requirements perfectly:
* Supports querying a table data through the `CQL` language.
* Consistency is tunable at the query level. A `Local Quorum` based consistency is adequate for most tables whereas the
  `User` table can be set to `Quorum` in the queries that access the table.
* Multi-master architecture that provides high availability with minimal downtime.
* Architecture also provides excellent performance for both read and write operations as it can coordinate multiple
  writes at a time.

<br/>

## Users Table Schema

### User

This `struct` embeds the `UserAccount` `struct` to create a representation of the user account table.

| Name (Struct) | Data Type (Struct) | Column Name | Column Type | Description                                                                                              |
|---------------|--------------------|-------------|-------------|----------------------------------------------------------------------------------------------------------|
| Account_ID    | string             | account_id  | text        | The account id unique identifier is a SHA3-256 hash of the `username`. Partition Key and Clustering Key. |
| Username      | string             | username    | text        | The username unique identifier. Partition Key and Clustering Key.                                        |
| Password      | string             | password    | text        | User's hashed password.                                                                                  |
| FirstName     | string             | first_name  | text        | User's first name.                                                                                       |
| LastName      | string             | last_name   | text        | User's last name.                                                                                        |
| Email         | string             | email       | text        | Email address.                                                                                           |
| IsDeleted     | bool               | is_deleted  | boolean     | Indicator as to whether the account is deleted. Prevents username reassignment.                          |

The Primary/Partition Key (`username`) might result in a hot partition. To potentially alleviate this issue the column
`account_id` has been created and is a SHA3-256 hash of `username`. This will allow the creation of a Compound
Primary/Partition Key (`username`, `account_id`). The `username` will be provided as input to the service logic and will
generate the consistent hash to use in the queries.

### User Account

This struct is created to be exposed for use with the HTTP handlers. This ensures consistency with the `User` `struct`.
It contains the `Username`, `Password`, `FirstName`, `LastName`, and `Email` fields.

### CQL Query
The query to generate the user table can be found [here](user.cql).

<br/>

## Quizzes Table Schema

### User Defined Type(s)

##### Question

Describes a single question with all it's answer options as well as the answer key for the question.

| Name (Struct) | Data Type (Struct) | Column Name | Column Type | Description                                                          |
|---------------|--------------------|-------------|-------------|----------------------------------------------------------------------|
| Description   | string             | description | text        | The description that contains the text of the question.              |
| Options       | [ ] string         | options     | list<text>  | The available options for the question.                              |
| Answers       | [ ] int            | answers     | list<int>   | The indices of the options that are correct answers in the question. |

### Quizzes

This `struct` embeds the `QuizCore` `struct` to create a representation of the quizzes table.

| Name (Struct) | Data Type (Struct) | Column Name  | Column Type                    | Description                                                               |
|---------------|--------------------|--------------|--------------------------------|---------------------------------------------------------------------------|
| Quiz_ID       | gocql.UUID         | quiz_id      | uuid                           | Account id unique identifier. Partition Key.                              |
| Author        | string             | author       | text                           | Username of the quiz creator.                                             |
| Title         | string             | title        | text                           | Description of the quiz.                                                  |
| Questions     | [ ] Question       | questions    | frozen<list<frozen<question>>> | A list of `question` UDTs in the quiz.                                    |
| IsPublished   | bool               | is_published | boolean                        | Status indicating whether the quiz can be viewed or taken by other users. |
| IsDeleted     | bool               | is_deleted   | boolean                        | Status indicating whether the quiz has been deleted.                      |

Since the Primary/Partition Key (`quiz_id`) is a `UUID`, it should help distribute the records evenly across the cluster
nodes. Quizzes are requested by their unique `quiz_id`'s.

### Quiz Core

This struct is created to be exposed for use with the HTTP handlers. This ensures consistency with the `Quiz` `struct`.
It contains the `Title` and `Question` fields and is the actual data used to create as well as what is presented when
viewing a quiz.

### CQL Query
The query to generate the user table can be found [here](quiz.cql).

<br/>

## Responses Table Schema

### Responses

This `struct` creates a representation of the responses table.

| Name (Struct) | Data Type (Struct) | Column Name | Column Type              | Description                                         |
|---------------|--------------------|-------------|--------------------------|-----------------------------------------------------|
| Username      | string             | username    | text                     | Username of the test taker. Compound Partition Key. |
| QuizID        | gocql.UUID         | quiz_id     | uuid                     | Taken quiz's id. Compound Partition Key.            |
| Score         | float64            | score       | double                   | Score for this submission. Clustering Key.          |
| Responses     | QuizResponse       | responses   | frozen<list<list<int>>>, | Recorded responses for the submission.              |

It would not be an arbitrary assumption that some quizzes will be more popular than others, leading to a hot partition. The
Compound Primary/Partition Key (`username`, `quiz_id`) should be unique enough to help distribute the records evenly
across the cluster nodes. The rationale behind the selection of this key is there will be more users taking the
quizzes than users who are also authors and requesting statistics of their quizzes. A Clustering Index of `score` will
sort the records on each node by `score`.

A secondary index will be constructed on just the `quiz_id` column. This is an ideal candidate for a secondary index because
of its perceived high cardinality. When a quiz author requests statistics for their published quiz by its `quiz_id`,
this index will be used to retrieve all the required records.

### CQL Query
The query to generate the user table can be found [here](responses.cql).