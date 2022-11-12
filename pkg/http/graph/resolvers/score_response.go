package graphql_resolvers

import (
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/auth"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	model_http "github.com/surahman/mcq-platform/pkg/model/http"
)

// prepareStatsResponse will prepare the GraphQL response struct.
func prepareStatsResponse(auth auth.Auth, dbResponse *model_cassandra.StatsResponse, quizId gocql.UUID) (response *model_http.StatsResponseGraphQL, err error) {
	response = &model_http.StatsResponseGraphQL{Records: dbResponse.Records}
	response.Metadata.QuizID = quizId
	response.Metadata.NumRecords = len(dbResponse.Records)

	// Encrypt page cursor link.
	if len(dbResponse.PageCursor) != 0 {
		if response.NextPage.Cursor, err = auth.EncryptToString(dbResponse.PageCursor); err != nil {
			return nil, err
		}
	}

	// Construct page size link segment.
	if len(dbResponse.PageCursor) != 0 && dbResponse.PageSize > 0 {
		response.NextPage.PageSize = dbResponse.PageSize
	}

	return
}
