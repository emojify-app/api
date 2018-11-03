package graphql

import (
	"testing"

	gographql "github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func TestNewCreatesQueryWithNameCreate(t *testing.T) {
	q := NewCreateQuery(func(q gographql.ResolveParams) (interface{}, error) {
		return nil, nil
	})

	assert.Equal(t, "Create", q.Name())
}

func TestNewCreatesQueryWithFieldURL(t *testing.T) {
	q := NewCreateQuery(func(q gographql.ResolveParams) (interface{}, error) {
		return nil, nil
	})

	assert.Equal(t, gographql.String, q.Fields()["url"].Type)
}
