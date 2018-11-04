package graphql

import (
	"github.com/emojify-app/api/logging"
	gographql "github.com/graphql-go/graphql"
)

// createResolver handles the logic for creating an image from the given graphql parameters
func createResolver(l logging.Logger) func(p gographql.ResolveParams) (interface{}, error) {

	return func(p gographql.ResolveParams) (interface{}, error) {
		defer l.CreateCalled("abc", p.Args["url"].(string)).Finished()

		return imageType, nil
	}
}
