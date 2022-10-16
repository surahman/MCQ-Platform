// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Saad Ur Rahman",
            "url": "https://www.linkedin.com/in/saad-ur-rahman/",
            "email": "saad.ur.rahman@gmail.com"
        },
        "license": {
            "name": "GPL-3.0",
            "url": "https://opensource.org/licenses/GPL-3.0"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/quiz/create/": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "This endpoint will create a quiz with randomly generated Test ID and associate it with the requester.\nThe username will be extracted from the JWT and associated with the Test ID.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "create test quiz"
                ],
                "summary": "Create a quiz.",
                "operationId": "createQuiz",
                "parameters": [
                    {
                        "description": "The Quiz to be created as unpublished",
                        "name": "quiz",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_cassandra.QuizCore"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The message will contain the Quiz ID of the newly generated quiz",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "409": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/quiz/delete/{quiz_id}": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "This endpoint will delete a quiz with the provided Test ID if it was created by the requester.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "delete remove test quiz"
                ],
                "summary": "Delete a quiz.",
                "operationId": "deleteQuiz",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the quiz being deleted.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The message will contain a confirmation of deletion",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "401": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/quiz/publish/{quiz_id}": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "When a quiz is submitted it is not published by default and is thus unavailable to be taken.\nThis endpoint will publish a quiz with the provided Test ID if it was created by the requester.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "publish test quiz create"
                ],
                "summary": "Publish a quiz.",
                "operationId": "publishQuiz",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the quiz being published.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The message will contain a confirmation of publishing",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "401": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/quiz/take/{quiz_id}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Take a quiz by submitting an answer sheet. The username will be extracted from the JWT and associated with the scorecard.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "take test quiz submit answer"
                ],
                "summary": "Take a quiz.",
                "operationId": "takeQuiz",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the answers being submitted.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "The answer card to be submitted.",
                        "name": "answers",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_cassandra.QuizResponse"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Score will be in the payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "401": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/quiz/update/{quiz_id}": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "This endpoint will update a quiz with the provided Test ID if it was created by the requester and is not published.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "update modify test quiz"
                ],
                "summary": "Update a quiz.",
                "operationId": "updateQuiz",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the quiz being updated.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "The Quiz to replace the one already submitted",
                        "name": "quiz",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_cassandra.QuizCore"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The message will contain a confirmation of the update",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "401": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "403": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/quiz/view/{quiz_id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "This endpoint will retrieve a quiz with a provided quiz ID if it is published.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "view test quiz"
                ],
                "summary": "View a quiz.",
                "operationId": "viewQuiz",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The quiz ID for the quiz being requested.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The message will contain the quiz ID and the payload will contain the quiz",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "403": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/score/stats/{quiz_id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets the statistics associated with a specific test if the user created the test.\nExtracts username from the JWT and the Test ID is provided as a path parameter.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "score scores stats statistics"
                ],
                "summary": "Get all statistics associated with a specific test.",
                "operationId": "getStats",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the requested statistics.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Statistics will be in the payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "401": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/score/test/{quiz_id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets a scorecard for a user. Extracts username from the JWT and Test ID is provided as a path parameter.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "score scores"
                ],
                "summary": "Get a user's score.",
                "operationId": "getScore",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The Test ID for the requested scorecard.",
                        "name": "quiz_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Score will be in the payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "404": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "Error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/user/delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Deletes a user stored in the database by marking it as deleted. The user must supply their login credentials as well as complete the following confirmation message: \"I understand the consequences, delete my user account USERNAME HERE\"",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users delete security"
                ],
                "summary": "Deletes a user. The user must supply their credentials as well as a confirmation message.",
                "operationId": "deleteUser",
                "parameters": [
                    {
                        "description": "The request payload for deleting an account",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_rest.DeleteUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "message with a confirmation of a deleted user account",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Success"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "409": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "Logs in a user by validating credentials and returning a JWT.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users login security"
                ],
                "summary": "Login a user.",
                "operationId": "loginUser",
                "parameters": [
                    {
                        "description": "Username and password to login with",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_cassandra.UserLoginCredentials"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "JWT in the api-key",
                        "schema": {
                            "$ref": "#/definitions/model_rest.JWTAuthResponse"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "403": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/user/refresh": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time. JWT must be expiring in under 60 seconds.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users login refresh security"
                ],
                "summary": "Refresh a user's JWT by extending its expiration time.",
                "operationId": "loginRefresh",
                "responses": {
                    "200": {
                        "description": "A new valid JWT",
                        "schema": {
                            "$ref": "#/definitions/model_rest.JWTAuthResponse"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "403": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "510": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        },
        "/user/register": {
            "post": {
                "description": "Creates a user account by inserting credentials into the database. A hashed password is stored.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users register security"
                ],
                "summary": "Register a user.",
                "operationId": "registerUser",
                "parameters": [
                    {
                        "description": "Username, password, first and last name, email address of user",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model_cassandra.UserAccount"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "a valid JWT token for the new account",
                        "schema": {
                            "$ref": "#/definitions/model_rest.JWTAuthResponse"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "409": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/model_rest.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model_cassandra.Question": {
            "type": "object",
            "required": [
                "answers",
                "description",
                "options"
            ],
            "properties": {
                "answers": {
                    "description": "The indices of the options that are correct answers in the question.",
                    "type": "array",
                    "maxItems": 5,
                    "minItems": 1,
                    "items": {
                        "type": "integer"
                    }
                },
                "asset": {
                    "description": "URI of an asset to be displayed with question.",
                    "type": "string"
                },
                "description": {
                    "description": "The description that contains the text of the question.",
                    "type": "string"
                },
                "options": {
                    "description": "The available options for the question.",
                    "type": "array",
                    "maxItems": 5,
                    "minItems": 2,
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "model_cassandra.QuizCore": {
            "type": "object",
            "required": [
                "questions",
                "title"
            ],
            "properties": {
                "marking_type": {
                    "description": "Marking scheme type can be not marked, negative marking, or all or nothing.",
                    "type": "string",
                    "enum": [
                        "None",
                        "none",
                        "Negative",
                        "negative",
                        "Non-negative",
                        "non-negative",
                        "Binary",
                        "binary"
                    ]
                },
                "questions": {
                    "description": "A list of questions in the quiz.",
                    "type": "array",
                    "maxItems": 10,
                    "minItems": 1,
                    "items": {
                        "$ref": "#/definitions/model_cassandra.Question"
                    }
                },
                "title": {
                    "description": "The title description of the quiz.",
                    "type": "string"
                }
            }
        },
        "model_cassandra.QuizResponse": {
            "type": "object",
            "required": [
                "responses"
            ],
            "properties": {
                "responses": {
                    "description": "The answer card to a quiz. The rows indices are the question numbers and the columns indices are the selected option numbers.",
                    "type": "array",
                    "maxItems": 10,
                    "minItems": 0,
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "integer"
                        }
                    }
                }
            }
        },
        "model_cassandra.UserAccount": {
            "type": "object",
            "required": [
                "email",
                "first_name",
                "last_name",
                "password",
                "username"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string"
                },
                "last_name": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "minLength": 8
                }
            }
        },
        "model_cassandra.UserLoginCredentials": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "minLength": 8
                }
            }
        },
        "model_rest.DeleteUserRequest": {
            "type": "object",
            "required": [
                "confirmation",
                "password",
                "username"
            ],
            "properties": {
                "confirmation": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "minLength": 8
                }
            }
        },
        "model_rest.Error": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "payload": {}
            }
        },
        "model_rest.JWTAuthResponse": {
            "type": "object",
            "required": [
                "expires",
                "threshold",
                "token"
            ],
            "properties": {
                "expires": {
                    "description": "Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.",
                    "type": "integer"
                },
                "threshold": {
                    "description": "The window in seconds before expiration during which the token can be refreshed.",
                    "type": "integer"
                },
                "token": {
                    "description": "JWT string sent to and validated by the server.",
                    "type": "string"
                }
            }
        },
        "model_rest.Success": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "payload": {}
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.1",
	Host:             "localhost:44243",
	BasePath:         "/api/rest/v1",
	Schemes:          []string{"http"},
	Title:            "Multiple Choice Question Platform.",
	Description:      "Multiple Choice Question Platform API.\nThis application supports the creation, managing, marking, viewing, retrieving stats, and scores of quizzes.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
