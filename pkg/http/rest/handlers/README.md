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
  "expires": "expiration time as Unix time stamp",
  "token": "token string"
}
```

### Error Response

There is a generic error response with a message and optional payload. If there is a validation error of some sort the
details of the failures will be enclosed within the payload section of the response.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

### Success Response

A successful request _may_ result in a response object when appropriate. In such an event, a message and an optional
payload will be returned.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

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

_Request:_ All fields are required and a valid JWT must be provided in the header. 
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
