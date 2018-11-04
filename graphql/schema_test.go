package graphql

import (
	"fmt"
	"testing"

	"github.com/emojify-app/api/logging"
	gographql "github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func createSchema(t *testing.T) gographql.Schema {
	l, err := logging.NewLogger("localhost:2231", "test", "test")
	assert.Nil(t, err)

	s, err := NewSchema(l)

	assert.Nil(t, err)

	return s
}

func TestNewSchema(t *testing.T) {
	s := createSchema(t)

	assert.NotNil(t, s)
}

func TestExecuteMutationReturns(t *testing.T) {
	createSchema := createSchema(t)

	result := gographql.Do(gographql.Params{
		Schema:        createSchema,
		RequestString: `mutation CreateNewImage {newImage: createImage(url: "test"){id url}}`,
	})

	assert.Equal(t, 0, len(result.Errors))
	for _, e := range result.Errors {
		fmt.Println(e.Error())
	}
}

func TestExecuteQueryReturns(t *testing.T) {
	createSchema := createSchema(t)

	result := gographql.Do(gographql.Params{
		Schema:        createSchema,
		RequestString: `query GetImageWithURL {newImage: image(url: "test"){id url}}`,
	})

	assert.Equal(t, 0, len(result.Errors))
	for _, e := range result.Errors {
		fmt.Println(e.Error())
	}
}
