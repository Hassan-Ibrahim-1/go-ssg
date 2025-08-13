package toml

import (
	"fmt"
	"maps"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
		err      error
	}{
		{"", nil, nil},
		{`theme = "rose-pine"`, map[string]string{"theme": "rose-pine"}, nil},
		{`theme = 'rose-pine'`, map[string]string{"theme": "rose-pine"}, nil},
		{`
title = "A Blog"
theme = "rose-pine"
deploy-to = "github"
author = "Hassan"
`,
			map[string]string{
				"title":     "A Blog",
				"theme":     "rose-pine",
				"deploy-to": "github",
				"author":    "Hassan",
			}, nil,
		},
		{
			`t'heme = 'rose-pine'`,
			nil,
			fmt.Errorf(
				"error on line 1: failed to sanitize key t'heme: ' is not a valid character. a key must only contain letters or a '-'",
			),
		},
		{
			`theme = ~rose-pine'`,
			nil,
			fmt.Errorf(`error on line 1: expected value to start with ' or "`),
		},
		{
			`theme = rose-pine`,
			nil,
			fmt.Errorf(`error on line 1: expected value to start with ' or "`),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			parsed, err := Parse([]byte(tt.input))
			if !errEqual(tt.err, err) {
				t.Fatalf(
					"unexpected err value. expected=%v. got=%v",
					tt.err,
					err,
				)
			}

			if !maps.Equal(tt.expected, parsed) {
				t.Errorf(
					"unexpected parsed value. expected=%+v. got=%+v",
					tt.expected,
					parsed,
				)
			}
		})
	}
}

func errEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
