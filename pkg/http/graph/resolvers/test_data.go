package graphql_resolvers

// getUsersQuery is a map of test user queries.
func getUsersQuery() map[string]string {
	return map[string]string{
		"register": `{
		"query": "mutation { registerUser(input: { firstname: \"%s\", lastname:\"%s\", email: \"%s\", userLoginCredentials: { username:\"%s\", password: \"%s\" } }) { token, expires, threshold }}"
	}`,
		"login": `{
		"query": "mutation { loginUser(input: { username:\"%s\", password: \"%s\" }) { token, expires, threshold }}"
	}`,
		"refresh": `{
		"query": "mutation { refreshToken() { token expires threshold }}"
	}`,

		"delete": `{
  "query": "mutation { deleteUser(input: { username: \"%s\" password: \"%s\" confirmation:\"I understand the consequences, delete my user account %s\" })}"
}`,
	}
}

// getQuizzesQuery is a map of test quiz queries.
func getQuizzesQuery() map[string]string {
	return map[string]string{
		"create_empty": `{
    "query": "mutation { createQuiz(input: { Title: \"\" MarkingType: \"\" Questions: [ { Description: \"\" Asset: \"\" Options: [\"Option 1\", \"\", \"\"] Answers: [] } { Description: \"\" Asset: \"\" Options: [\"\", \"\"] Answers: [] } ] } )}"
}`,
		"create_valid": `{
    "query": "mutation { createQuiz(input: { Title: \"Sample quiz title\" MarkingType: \"Negative\" Questions: [ { Description: \"Sample quiz description\" Asset: \"http://url-of-asset.com/asset.txt\" Options: [\"Option 1\", \"Option 2\", \"Option 3\"] Answers: [2] } { Description: \"Another question\" Asset: \"http://url-of-another-asset.com/img.jpg\" Options: [\"Another opt 1\", \"Another opt 2\"] Answers: [1] } ] } )}"
}`,
		"create_invalid": `{
    "query": "mutation { createQuiz(input: { Title: \"Sample quiz title\" MarkingType: \"Negative\" Questions: [ { Description: \"Sample quiz description\" Asset: \"http://url-of-asset.com/asset.txt\" Options: [\"Option 1\", \"Option 2\", \"Option 3\"] Answers: [2] } { Description: \"This question only has one option and is invalid\" Asset: \"http://url-of-another-asset.com/img.jpg\" Options: [\"Another opt 1\"] Answers: [0] } ] } )}"
}`,
		"update_valid": `{
    "query": "mutation { updateQuiz( quizID: \"%s\" quiz: { Title: \"Sample quiz title\" MarkingType: \"Negative\" Questions: [ { Description: \"Sample quiz description\" Asset: \"http://url-of-asset.com/asset.txt\" Options: [\"Option 1\", \"Option 2\", \"Option 3\"] Answers: [2] } { Description: \"Another question\" Asset: \"http://url-of-another-asset.com/img.jpg\" Options: [\"Another opt 1\", \"Another opt 2\"] Answers: [1] } ] } )}"
}`,
		"update_invalid": `{
    "query": "mutation { updateQuiz( quizID: \"%s\" quiz: { Title: \"\" MarkingType: \"\" Questions: [ { Description: \"\" Asset: \"\" Options: [\"\", \"\", \"\"] Answers: [] } { Description: \"\" Asset: \"\" Options: [\"\", \"\"] Answers: [] } ] } )}"
}`,
		"view": `{
  	"query": "query { viewQuiz(quizID: \"%s\"){ Title MarkingType Questions { Description Asset Options Answers } }}"
}`,
		"delete": `{
	"query": "mutation { deleteQuiz(quizID:\"%s\")}"
}`,
		"publish": `{
	"query": "mutation { publishQuiz(quizID:\"%s\")}"
}`,
		"take": `{
    "query": "mutation { takeQuiz( quizID:\"%s\" input: { Responses: %v } ) { Username Author Score QuizResponse QuizID }}"
}`,
	}

}

// getScoresQuery is a map of test scores queries.
func getScoresQuery() map[string]string {
	return map[string]string{
		"score": `{
  	"query": "query { getScore(quizID:\"%s\") { Username Author Score QuizResponse QuizID }}"
}`,
	}
}
