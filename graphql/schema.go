package graphql

import (
	"github.com/emojify-app/api/logging"
	gographql "github.com/graphql-go/graphql"
)

// NewSchema returns a new
func NewSchema(l logging.Logger) (gographql.Schema, error) {
	return gographql.NewSchema(
		gographql.SchemaConfig{
			Query:    NewGetQuery(getResolver(l)),
			Mutation: NewCreateMutation(createResolver(l)),
		},
	)
}
