package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSheetId(t *testing.T) {
	tests := []struct {
		name string
		url  string
		res  string
		ok   bool
	}{
		{
			"correct url",
			"https://docs.google.com/spreadsheets/d/1v0JuSxEI0wDh1_VXgKKBb8tCuS5yjJ7bBfz8xbNN5SU/edit#gid=0",
			"1v0JuSxEI0wDh1_VXgKKBb8tCuS5yjJ7bBfz8xbNN5SU",
			true,
		},
		{
			"incorrect url",
			"https://docs.google.com/spreadshedssets/d/1v0JuSxEI0wDh1_VXgKKBb8tCuS5yjJ7bBfz8xbNN5SU/edit#gid=0",
			"",
			false,
		},
	}

	for _, test := range tests{
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseSheetId(test.url)
			if test.ok {
				assert.Nil(t, err)
				assert.Equal(t, test.res, res)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
