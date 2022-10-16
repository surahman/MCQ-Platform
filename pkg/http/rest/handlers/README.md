# HTTP REST API Endpoints

The REST API schema can be tested and reviewed through the Swagger UI that is exposed when the server is started.

<br/>

## Table of contents


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
- A marking type of `None`, `Binary`, `Negative`, or `Non-negative` is accepted.
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

