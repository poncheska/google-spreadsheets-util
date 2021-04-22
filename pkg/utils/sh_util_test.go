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

	for _, test := range tests {
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

func TestParseRange(t *testing.T) {
	tests := []struct {
		name string
		in   []ParsedFieldData
		res  struct {
			rngStr string
			rng    Range
		}
		ok bool
	}{
		{
			"correct data",
			[]ParsedFieldData{
				{Row: 2, Col: 1, Content: "kek1"},
				{Row: 6, Col: 4, Content: "kek2"},
				{Row: 7, Col: 3, Content: "kek3"},
			},
			struct {
				rngStr string
				rng    Range
			}{rngStr: "C1:H4", rng: Range{2,1, 7, 4} },
			true,
		},
		{
			"incorrect data: too large number",
			[]ParsedFieldData{
				{Row: 1, Col: 1, Content: "kek1"},
				{Row: 100, Col: 1, Content: "kek2"},
			},
			struct {
				rngStr string
				rng    Range
			}{},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rngStr, rng, err := ParseRange(test.in)
			if test.ok {
				assert.Nil(t, err)
				assert.Equal(t, test.res.rng, rng)
				assert.Equal(t, test.res.rngStr, rngStr)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
