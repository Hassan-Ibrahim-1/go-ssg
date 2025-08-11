package markdown

import (
	"bytes"
	"fmt"
	"maps"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		input             string
		expectedMetadata  map[string]string
		expectedRemaining string
	}{
		{`
+++
author = "Jane Doe"
title = "Test markdown"
description = "A basic markdown file"
+++

Hello, World
`,
			map[string]string{
				"author":      "Jane Doe",
				"title":       "Test markdown",
				"description": "A basic markdown file",
			},
			"Hello, World",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			metadata, remaining := parseMetadata([]byte(tt.input))
			if !maps.Equal(tt.expectedMetadata, metadata) {
				t.Errorf(
					"bad metadata. expected=%+v. got=%+v",
					tt.expectedMetadata,
					metadata,
				)
			}
			if !bytes.Equal([]byte(tt.expectedRemaining), remaining) {
				t.Errorf(
					"bad remaining markdown. expected=%q. got=%q",
					tt.expectedRemaining,
					string(remaining),
				)
			}
		})
	}
}
