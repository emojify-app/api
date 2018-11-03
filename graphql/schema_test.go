package graphql

import (
	"fmt"
	"testing"

	gographql "github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func TestNewSchema(t *testing.T) {
	s, err := NewSchema()

	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestExecuteQueryReturns(t *testing.T) {
	createSchema, _ := NewSchema()
	result := gographql.Do(gographql.Params{
		Schema:        createSchema,
		RequestString: `{create {url: "test"}}`,
	})

	assert.Equal(t, 0, len(result.Errors))
	for _, e := range result.Errors {
		fmt.Println(e.Error())
	}
}
