// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model_http

type JWTResponse struct {
	Token     string `json:"Token"`
	Expires   int    `json:"Expires"`
	Threshold int    `json:"Threshold"`
}

type QuestionCreate struct {
	Description string   `json:"Description"`
	Asset       string   `json:"Asset"`
	Options     []string `json:"Options"`
	Answers     []int    `json:"Answers"`
}

type QuizCreate struct {
	Title       string            `json:"Title"`
	MarkingType string            `json:"MarkingType"`
	Questions   []*QuestionCreate `json:"Questions"`
}

type QuizResponseInput struct {
	Responses [][]int `json:"Responses"`
}

type UserDeletion struct {
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	Confirmation string `json:"Confirmation"`
}

type UserLogin struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type UserRegistration struct {
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Email     string `json:"Email"`
	Username  string `json:"Username"`
	Password  string `json:"Password"`
}
