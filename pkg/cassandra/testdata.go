package cassandra

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

var quizzesUUIDMapping = GetQuizzedUUIDMapping()

// configTestData will return a map of test data containing valid and invalid Cassandra configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"valid-ci": `
authentication:
  username: cassandra
  password: cassandra
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [localhost]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"password_empty": `
authentication:
  username: admin
  password:
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"username_empty": `
authentication:
  username:
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"keyspace_empty": `
authentication:
  username: admin
  password: root
keyspace:
  name:
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"consistency_missing": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"ip_empty": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: []
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5`,

		"timeout_zero": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 0
  max_connection_attempts: 5`,

		"invalid_min_max_conn_attempts": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 0`,

		"test_suite": `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 2`,
	}
}

// GetQuizzedUUIDMapping returns a map containing a Quiz title to UUID mapping.
func GetQuizzedUUIDMapping() map[string]gocql.UUID {
	data := make(map[string]gocql.UUID)
	var err error

	if data["providedPubQuiz"], err = gocql.ParseUUID("729c6e51-b8ae-4f23-9419-9401a51968c1"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["providedNoPubQuiz"], err = gocql.ParseUUID("f37a7ba3-1541-4d8a-95af-03689a91d39c"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["myPubQuiz"], err = gocql.ParseUUID("d1aec072-201e-4c91-89c8-d2cc402e2cd9"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["myNoPubQuiz"], err = gocql.ParseUUID("320b5b74-ebcd-40c1-91bd-da28b8b3293a"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["invalidOptionsNoPubQuiz"], err = gocql.ParseUUID("45ba24bf-8bd9-4b20-af7a-4dddbf2428b9"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["myPubQuizDeleted"], err = gocql.ParseUUID("17166e35-1a70-47da-8779-21460e0d52f5"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}
	if data["myNoPubQuizDeleted"], err = gocql.ParseUUID("166bdda4-9d9d-406f-a3a7-9eb67b5311ad"); err != nil {
		log.Fatalln("Failed to parse UUID")
	}

	return data
}

// GetTestUsers will generate a number of dummy users for testing.
func GetTestUsers() map[string]*model_cassandra.User {
	users := make(map[string]*model_cassandra.User)
	username := "username%d"
	password := "user-password-%d"
	firstname := "firstname-%d"
	lastname := "lastname-%d"
	email := "user%d@email-address.com"

	for idx := 1; idx < 5; idx++ {
		uname := fmt.Sprintf(username, idx)
		users[uname] = &model_cassandra.User{
			UserAccount: &model_cassandra.UserAccount{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: uname,
					Password: fmt.Sprintf(password, idx),
				},
				FirstName: fmt.Sprintf(firstname, idx),
				LastName:  fmt.Sprintf(lastname, idx),
				Email:     fmt.Sprintf(email, idx),
			},
			AccountID: blake2b256(uname),
			IsDeleted: false,
		}
	}

	return users
}

// GetTestQuizzes is a map of test quiz data.
func GetTestQuizzes() map[string]*model_cassandra.Quiz {
	data := make(map[string]*model_cassandra.Quiz)
	temperatureQuestion := model_cassandra.Question{Description: "Temperature can be measured in",
		Options: []string{"Kelvin", "Fahrenheit", "Gram", "Celsius", "Liters"},
		Answers: []int32{0, 1, 3}}
	moonQuestion := model_cassandra.Question{Description: "Moon is a star",
		Options: []string{"Yes", "No"},
		Answers: []int32{1}}
	weightQuestion := model_cassandra.Question{Description: "Weight can be measured in",
		Options: []string{"Kelvin", "Fahrenheit", "Gram", "Celsius", "Liters"},
		Answers: []int32{2}}
	booleanQuestion := model_cassandra.Question{Description: "Boolean can be true or false",
		Options: []string{"Yes", "No"},
		Answers: []int32{0}}
	invalidQuestion := model_cassandra.Question{Description: "This question only has one option and is invalid",
		Options: []string{"Yes"},
		Answers: []int32{0}}

	data["providedPubQuiz"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["providedPubQuiz"], Author: "user-1", IsPublished: true,
		QuizCore: &model_cassandra.QuizCore{
			Title:       "Sample quiz published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
		}}
	data["providedNoPubQuiz"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["providedNoPubQuiz"], Author: "user-1",
		QuizCore: &model_cassandra.QuizCore{
			Title:       "Sample quiz not published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
		}}
	data["myPubQuiz"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["myPubQuiz"], Author: "user-2", IsPublished: true,
		QuizCore: &model_cassandra.QuizCore{
			Title:       "My sample quiz published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&weightQuestion, &booleanQuestion},
		}}
	data["myNoPubQuiz"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["myNoPubQuiz"], Author: "user-3",
		QuizCore: &model_cassandra.QuizCore{
			Title:       "My sample quiz not published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&weightQuestion, &booleanQuestion},
		}}
	data["invalidOptionsNoPubQuiz"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["invalidOptionsNoPubQuiz"], Author: "user-4", QuizCore: &model_cassandra.QuizCore{
		Title:       "Invalid Options (1 options) unpublished",
		MarkingType: "Negative",
		Questions:   []*model_cassandra.Question{&invalidQuestion},
	}}
	data["myPubQuizDeleted"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["myPubQuizDeleted"], Author: "user-2", IsPublished: true, IsDeleted: true,
		QuizCore: &model_cassandra.QuizCore{
			Title:       "My sample quiz published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&weightQuestion, &booleanQuestion},
		}}
	data["myNoPubQuizDeleted"] = &model_cassandra.Quiz{QuizID: quizzesUUIDMapping["myNoPubQuizDeleted"], Author: "user-3", IsDeleted: true,
		QuizCore: &model_cassandra.QuizCore{
			Title:       "My sample quiz not published",
			MarkingType: "Negative",
			Questions:   []*model_cassandra.Question{&weightQuestion, &booleanQuestion},
		}}

	return data
}

// GetTestResponses is a map of test response data.
func GetTestResponses() map[string]*model_cassandra.Response {
	return map[string]*model_cassandra.Response{
		"user-1_myPubQuiz": {
			Username:     "user-1",
			Author:       "user-2",
			Score:        1.0,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{3}, {0}}},
			QuizID:       quizzesUUIDMapping["myPubQuiz"],
		},
		"user-2_myPubQuiz": {
			Username:     "user-2",
			Author:       "user-2",
			Score:        0.5,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{3}, {1}}},
			QuizID:       quizzesUUIDMapping["myPubQuiz"],
		},
		"user-3_providedPubQuiz": {
			Username:     "user-3",
			Author:       "user-1",
			Score:        0.5,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 2}, {1}}},
			QuizID:       quizzesUUIDMapping["providedPubQuiz"],
		},
		"user-2_providedPubQuiz": {
			Username:     "user-2",
			Author:       "user-1",
			Score:        2,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 2}, {0}}},
			QuizID:       quizzesUUIDMapping["providedPubQuiz"],
		},
		"user-4_myPubQuiz": {
			Username:     "user-4",
			Author:       "user-2",
			Score:        4.4,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{3, 4}, {0, 4}}},
			QuizID:       quizzesUUIDMapping["myPubQuiz"],
		},
		"user-5_myPubQuiz": {
			Username:     "user-5",
			Author:       "user-2",
			Score:        5.5,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{4, 5}, {1, 5}}},
			QuizID:       quizzesUUIDMapping["myPubQuiz"],
		},
		"user-6_myPubQuiz": {
			Username:     "user-6",
			Author:       "user-2",
			Score:        6.6,
			QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{5, 6}, {2, 6}}},
			QuizID:       quizzesUUIDMapping["myPubQuiz"],
		},
	}
}
