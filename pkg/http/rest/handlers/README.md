# HTTP REST API Endpoints

The REST API schema can be tested and reviewed through the Swagger UI that is exposed when the server is started.

<br/>

## Table of contents

- [Authorization Response](#authorization-response)
- [Error Response](#error-response)
- [Success Response](#success-response)
- [User Endpoints `/user/`](#user-endpoints-user)
  - [Register](#register)
  - [Login](#login)
  - [Refresh](#refresh)
  - [Delete](#delete)
- [Quiz Endpoints `/quiz/`](#quiz-endpoints-quiz)
  - [Create](#create)
  - [View](#view)
  - [Update](#update)
  - [Delete](#delete)
  - [Publish](#publish)
  - [Take](#take)
- [Score Endpoints `/score/`](#score-endpoints-score)
  - [Test](#test)
  - [Stats](#stats)
  - [Stats - _Paginated_](#stats---paginated)
- [Healthcheck Endpoint `/health`](#healthcheck-endpoint-health)

<br/>

### Authorization Response

Authorization is implemented using JSON Web Tokens. An expiration deadline for the JWT is returned in response. It is 
the client's responsibility to refresh the token before, but no sooner than 60 seconds, before the deadline.

The returned token schema is below.

```json
{
  "expires": "expiration time integer in seconds, Unix time stamp",
  "token": "token string",
  "threshold": "threshold in integer seconds before expiration when the token can be refreshed"
}
```

<br/>

### Error Response

There is a generic error response with a message and optional payload. If there is a validation error of some sort the
details of the failures will be enclosed within the payload section of the response.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

<br/>

### Success Response

A successful request _may_ result in a response object when appropriate. In such an event, a message and an optional
payload will be returned.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

<br/>

### User Endpoints `/user/`

#### Register

_Request:_ All fields are required.
_Response:_ A valid JWT will be returned as an authorization response. 

```json
{
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "password": "string",
  "username": "string"
}
```

#### Login

_Request:_ All fields are required. 
_Response:_ A valid JWT will be returned as an authorization response. 

```json
{
  "password": "string",
  "username": "string"
}
```

#### Refresh

_Request:_ A valid JWT must be provided in the request header and will be validated with a fresh token issued against it. 
_Response:_ A valid JWT will be returned as an authorization response. 

```json
{
  "expires": "expiration time string",
  "token": "token string"
}
```

#### Delete

_Request:_ All fields are required and a valid JWT must be provided in the header. The user must supply their login
           credentials as well as complete the confirmation message `I understand the consequences, delete my user
           account **USERNAME HERE**`
_Response:_ A confirmation message will be returned as a success response. 

```json
{
  "confirmation": "I understand the consequences, delete my user account <USERNAME HERE>",
  "password": "password string",
  "username": "username string"
}
```

<br/>

### Quiz Endpoints `/quiz/`

#### Create

_Request:_ All fields except `asset` are required.
- A marking type of `None`, `Binary`, `Negative`, or `Non-negative` is accepted. Details on marking are available in the [`grading`](../../../grading) package.
- 1 to 10 `question`s are permitted per quiz.
- 1 to 5 options are permitted per `question`.
- Answer must be fewer than the number of options. Each number in the answer is an index to an option and must be in the range [0, 4].
- Every question has an optional asset that is a URL Encoded URI.

_Response:_ A success response containing the `quiz id` in the payload.

```json
{
  "marking_type": "One of: None, Binary, Negative, or Non-negative",
  "questions": [
    {
      "asset": "URL encoded URI of asset",
      "description": "actual question here",
      "options": ["option 1", "option 2", "option 3", "option 4", "option 5"],
      "answers": [0,1,2,3,4]
    },{
      "asset": "URL encoded URI of asset",
      "description": "actual question here",
      "options": ["option 1", "option 2", "option 3", "option 4", "option 5"],
      "answers": [0,1,2,3,4]
    }
  ],
  "title": "The title of the quiz"
}
```

#### View

Only quizzes neither published nor deleted may be viewed by non-authors. Authors of quizzes can view both deleted and
unpublished quizzes. Answer keys will only be returned to requesters who are the quiz's authors. The username of the
requester is extracted from their JWT.

_Request:_ The Quiz ID must be supplied in the request the URL.

_Response:_ A success response containing the `quiz id` in the message and the quiz in the payload.

#### Update

Updates to a quiz can only be made by the author of the quiz. The username of the requester is extracted from their JWT.
The items that can be updated in a quiz are the `marking type`, `title`, and the entirety of the `questions`. The entire
contents of the quiz core must be supplied for the update.

The intended workflow for an update is a client will make a read/view request for the quiz. The quiz will then be rendered
on the client in an editor where changes can be made. The client will subsequently submit the entire quiz, with the
updates, to the API.

_Request:_ The Quiz ID must be supplied in the request the URL along with the contents of the `QuizCore` in the request
body.

_Response:_ A success response containing a confirmation message and the `quiz id` in the payload.

#### Delete

Only the authors of a quiz may mark it as deleted. Once deleted, a quiz will be set to unpublished and will no longer
be eligible for publishing and editing. The quiz will remain in the database and can only be viewed by the author.

_Request:_ The Quiz ID must be supplied in the request URL.

_Response:_ A success response containing a confirmation message and the `quiz id` in the payload.

#### Publish

Only the authors of a quiz may mark it as published. Once published, a quiz will be generally available to all users and
will no longer be eligible for editing. The quiz can be made unavailable by deleting it.

_Request:_ The Quiz ID must be supplied in the request URL.

_Response:_ A success response containing a confirmation message and the `quiz id` in the payload.

#### Take

Any registered user is allowed to take or submit answers to a quiz that is published and has not been deleted yet. A
user may only take a quiz once.

_Request:_ The Quiz ID must be supplied in the request URL. The responses are provided in a two-dimensional array of
integers in the request body. The questions and answers are zero-indexed. The answers for each question must be supplied
in the row number corresponding to the question number. To select options for a question, the user must specify the
indices of the options in the questions row array.

_Response:_ A success response containing a confirmation message with the `quiz id` as well as the score, if applicable,
in the payload. Please see the [`grading`](../../../grading) package for details on marking.

```json
{
  "responses": [
    [0, 1, 2, 3, 4 ],
    [1, 3],
    [1, 2, 4]
  ]
}
```

<br/>

### Score Endpoints `/score/`

#### Test

A user may request the score for a quiz they have already taken, whether the quiz is deleted or not.

_Request:_ The Quiz ID must be supplied in the request URL.

_Response:_ A success response containing a message with the scorecard in the payload. An example response is below.

```json
{
  "message": "score card",
  "payload": {
    "username": "username1",
    "score": 0.6666666666666666,
    "responses": [
      [0, 1, 2],
      [1, 3]
    ],
    "quiz_id": "74522665-4d8a-11ed-b4cb-305a3a460e3e"
  }
}
```

#### Stats

An author of a quiz may request all the scorecards for their quiz. The scorecards will contain the `username`s, `score`s,
as well as `answers` that all users have submitted for the quiz.

_Request:_ The Quiz ID must be supplied in the request URL.

_Response:_ A success response containing a message with the scorecards in the payload. An example response is below.

```json
{
  "message": "score cards",
  "payload": [
    {
      "username": "username4",
      "score": 0.4444,
      "responses": [ [0, 1, 2], [1, 3] ],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    },
    {
      "username": "username3",
      "score": 0.1111,
      "responses": [ [0, 1, 2], [1, 3] ],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    },
    {
      "username": "username2",
      "score": 0.2222,
      "responses": [ [0, 1, 2], [1, 3] ],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    },
    {
      "username": "username1",
      "score": 0.6666666666666666,
      "responses": [ [0, 1, 2], [1, 3] ],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    }
  ]
}
```


#### Stats - _Paginated_

An author of a quiz may request all the scorecards for their quiz. The scorecards will contain the `username`s, `score`s,
as well as `answers` that all users have submitted for the quiz.

The records from this request can potentially number in the thousands. As such, this endpoint will return paged responses
for the request with a link to the subsequent page of data provided in the response. The page cursor is a pointer into
the records in the query where the read will resume from. The data contained in the page cursor can reveal sensitive
information contained in the database and thus must be encrypted before being returned from the HTTP handler.

It is critical to consider that adding or removing rows in other queries may affect the records being retrieved by a
paginated query into the same table. Since paged queries rely on a cursor to a read position in the database table
returned results from a paged query may return duplicated or missed records between pages.

_Request:_ The `Quiz ID` must be supplied in the request URL. The encrypted and Base64 URL encoded page cursor is provided
with the query parameter `pageCursor`. The maximum number of records to retrieve in a request is set via the `pageSize`
query parameter. The first request sent to this endpoint should only include the `pageSize`. The subsequent responses will
have the query string link generated automatically.

_Response:_ A paged statistics response will be returned on a successful request. The following page of data can be
accessed by appending the query string in the `Links.NextPage` to the base request URI that consists of the endpoint URI
with the `Quiz ID` in the path.

```json
{
  "records": [
    {
      "username": "username4",
      "score": 0.4444,
      "responses": [[0, 1, 2], [1, 3]],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    },
    {
      "username": "username6",
      "score": 0.7777,
      "responses": [[0, 1, 2], [1, 3]],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    },
    {
      "username": "username5",
      "score": 0.5555,
      "responses": [[0, 1, 2], [1, 3]],
      "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e"
    }
  ],
  "metadata": {
    "quiz_id": "0a704c4b-4ea2-11ed-bd5a-305a3a460e3e",
    "num_records": 3
  },
  "links": {
    "next_page": "?pageCursor=GnkrBKWMAmEVbB0okIP4Mr0lzU_TAX3yifbc6Fa8lIBOPbF30YOoTKyHOjVosFSnnYF8_3LQ8hQwqa6f6sJpvFbd9A==&pageSize=3"
  }
}
```

To access the next page of data using the example above we would follow the URI below:

localhost:44243/api/rest/v1/0a704c4b-4ea2-11ed-bd5a-305a3a460e3e?pageCursor=GnkrBKWMAmEVbB0okIP4Mr0lzU_TAX3yifbc6Fa8lIBOPbF30YOoTKyHOjVosFSnnYF8_3LQ8hQwqa6f6sJpvFbd9A==&pageSize=3

<br/>

### Healthcheck Endpoint `/health`

The health check endpoint is exposed to facilitate liveness checks on the service. The check will verify whether the
service is connected to all the ancillary services and responds appropriately.

This check is essential for load balancers and container orchestrators to determine whether to route traffic or restart
the container.

_Healthy Response:_ HTTP 200 OK

_Unhealthy Response:_ HTTP 503 Service Unavailable
