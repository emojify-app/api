package graphql

import (
	"github.com/emojify-app/api/logging"
	gographql "github.com/graphql-go/graphql"
)

// getResolver handles the logic for fetching an image from the given graphql parameters
func getResolver(l logging.Logger) func(p gographql.ResolveParams) (interface{}, error) {

	return func(p gographql.ResolveParams) (interface{}, error) {
		return imageType, nil
	}
}
