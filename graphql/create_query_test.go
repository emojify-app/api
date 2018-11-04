package graphql

import (
	"testing"

	gographql "github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func TestNewCreatesMutationWithNameCreate(t *testing.T) {
	q := NewCreateMutation(func(q gographql.ResolveParams) (interface{}, error) {
		return nil, nil
	})

	assert.Equal(t, "RootMutation", q.Name())
}

func TestNewCreatesMutationWithArgumentURL(t *testing.T) {
	q := NewCreateMutation(func(q gographql.ResolveParams) (interface{}, error) {
		return nil, nil
	})

	assert.Equal(t, imageType, q.Fields()["createImage"].Type)
}
