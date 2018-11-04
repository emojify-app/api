package graphql

import gographql "github.com/graphql-go/graphql"

var imageType = gographql.NewObject(gographql.ObjectConfig{
	Name: "Image",
	Fields: gographql.Fields{
		"id": &gographql.Field{
			Type: gographql.String,
		},
		"url": &gographql.Field{
			Type: gographql.String,
		},
	},
})

//NewCreateMutation returns a new graphql mutation used for emojifying images
func NewCreateMutation(resolver func(p gographql.ResolveParams) (interface{}, error)) *gographql.Object {
	return gographql.NewObject(
		gographql.ObjectConfig{
			Name: "RootMutation",
			Fields: gographql.Fields{
				"createImage": &gographql.Field{
					Type: imageType,
					Args: gographql.FieldConfigArgument{
						"url": &gographql.ArgumentConfig{
							Type: gographql.NewNonNull(gographql.String),
						},
					},
					Resolve: resolver,
				},
			},
		},
	)
}

// NewGetQuery returns a new graphql query for querying existing images
func NewGetQuery(resolver func(p gographql.ResolveParams) (interface{}, error)) *gographql.Object {
	return gographql.NewObject(
		gographql.ObjectConfig{
			Name: "RootQuery",
			Fields: gographql.Fields{
				"image": &gographql.Field{
					Type: imageType,
					Args: gographql.FieldConfigArgument{
						"url": &gographql.ArgumentConfig{
							Type: gographql.NewNonNull(gographql.String),
						},
					},
					Resolve: resolver,
				},
			},
		},
	)
}
