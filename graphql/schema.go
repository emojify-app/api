package graphql

import gographql "github.com/graphql-go/graphql"

// NewSchema returns a new
func NewSchema() (gographql.Schema, error) {
	return gographql.NewSchema(
		gographql.SchemaConfig{
			Query:    NewGetQuery(getResolver),
			Mutation: NewCreateMutation(createResolver),
		},
	)
}

func createResolver(p gographql.ResolveParams) (interface{}, error) {
	return imageType, nil
}

func getResolver(p gographql.ResolveParams) (interface{}, error) {
	return nil, nil
}
