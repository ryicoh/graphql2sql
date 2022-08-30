package graphql2sql

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		operation string
		variables map[string]any
		sql       string
		args      []any
		err       error
	}{
		{
			`query FindTweet($keyword: String!) {
        users(where: {name: {_contains: $keyword}}) {
          id
          name
        }
      }`,
			map[string]any{"keyword": "hello"},
			"SELECT id, name FROM users WHERE position(users.name in ?) > 0",
			[]any{"hello"},
			nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.operation, func(t *testing.T) {
			sql, args, err := Parse([]byte(tC.operation), tC.variables)
			assertEquals(t, tC.err, err)
			assertEquals(t, tC.sql, sql)
			assertEquals(t, tC.args, args)
		})
	}
}

func assertEquals[T any](t *testing.T, expect, actual T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("expect (%v), but got (%v)", expect, actual)
	}
}
