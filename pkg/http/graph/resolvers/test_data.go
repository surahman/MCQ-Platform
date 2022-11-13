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
    "query": "mutation { createQuiz(input: { title: \"\" markingType: \"\" questions: [ { description: \"\" asset: \"\" options: [\"Option 1\", \"\", \"\"] answers: [] } { description: \"\" asset: \"\" options: [\"\", \"\"] answers: [] } ] } )}"
}`,
		"create_valid": `{
    "query": "mutation { createQuiz(input: { title: \"Sample quiz title\" markingType: \"Negative\" questions: [ { description: \"Sample quiz description\" asset: \"http://url-of-asset.com/asset.txt\" options: [\"Option 1\", \"Option 2\", \"Option 3\"] answers: [2] } { description: \"Another question\" asset: \"http://url-of-another-asset.com/img.jpg\" options: [\"Another opt 1\", \"Another opt 2\"] answers: [1] } ] } )}"
}`,
		"create_invalid": `{
    "query": "mutation { createQuiz(input: { title: \"Sample quiz title\" markingType: \"Negative\" questions: [ { description: \"Sample quiz description\" asset: \"http://url-of-asset.com/asset.txt\" options: [\"Option 1\", \"Option 2\", \"Option 3\"] answers: [2] } { description: \"This question only has one option and is invalid\" asset: \"http://url-of-another-asset.com/img.jpg\" options: [\"Another opt 1\"] answers: [0] } ] } )}"
}`,
		"update_valid": `{
    "query": "mutation { updateQuiz( quizID: \"%s\" quiz: { title: \"Sample quiz title\" markingType: \"Negative\" questions: [ { description: \"Sample quiz description\" asset: \"http://url-of-asset.com/asset.txt\" options: [\"Option 1\", \"Option 2\", \"Option 3\"] answers: [2] } { description: \"Another question\" asset: \"http://url-of-another-asset.com/img.jpg\" options: [\"Another opt 1\", \"Another opt 2\"] answers: [1] } ] } )}"
}`,
		"update_invalid": `{
    "query": "mutation { updateQuiz( quizID: \"%s\" quiz: { title: \"\" markingType: \"\" questions: [ { description: \"\" asset: \"\" options: [\"\", \"\", \"\"] answers: [] } { description: \"\" asset: \"\" options: [\"\", \"\"] answers: [] } ] } )}"
}`,
		"view": `{
  	"query": "query { viewQuiz(quizID: \"%s\"){ title markingType questions { description asset options answers } }}"
}`,
		"delete": `{
	"query": "mutation { deleteQuiz(quizID:\"%s\")}"
}`,
		"publish": `{
	"query": "mutation { publishQuiz(quizID:\"%s\")}"
}`,
		"take": `{
    "query": "mutation { takeQuiz( quizID:\"%s\" input: { responses: %v } ) { username author score quizResponse quizID }}"
}`,
	}

}

// getScoresQuery is a map of test scores queries.
func getScoresQuery() map[string]string {
	return map[string]string{
		"score": `{
  	"query": "query { getScore(quizID:\"%s\") { username author score quizResponse quizID }}"
}`,
		"stats": `{
    "query": "query { getStats(quizID:\"%s\", pageSize: %d, cursor:\"%s\") { records { username author score quizResponse quizID } metadata { quizID numRecords } nextPage { pageSize cursor } }}"
}`,
		"stats_quiz_id": `{
    "query": "query { getStats(quizID:\"%s\") { records { username author score quizResponse quizID } metadata { quizID numRecords } nextPage { pageSize cursor } }}"
}`,
		"stats_page_size": `{
    "query": "query { getStats(quizID:\"%s\", pageSize: %d) { records { username author score quizResponse quizID } metadata { quizID numRecords } nextPage { pageSize cursor } }}"
}`,
	}
}
