package graphql

import gographql "github.com/graphql-go/graphql"

//NewCreateQuery returns a new graphql query used for emojifying images
func NewCreateQuery(resolver func(p gographql.ResolveParams) (interface{}, error)) *gographql.Object {
	return gographql.NewObject(
		gographql.ObjectConfig{
			Name: "Create",
			Fields: gographql.Fields{
				"url": &gographql.Field{
					Type:    gographql.String,
					Resolve: resolver,
				},
			},
		},
	)
}
