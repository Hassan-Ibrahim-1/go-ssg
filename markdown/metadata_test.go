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
		expectedErr       error
	}{
		{`
+++
author = Jane Doe
title = Test markdown
description = A basic markdown file
+++

Hello, World
`,
			map[string]string{
				"author":      "Jane Doe",
				"title":       "Test markdown",
				"description": "A basic markdown file",
			},
			"Hello, World\n", nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			metadata, remaining, err := parseMetadata([]byte(tt.input))
			if err != tt.expectedErr {
				t.Fatalf(
					"Unexpected err expected=%v. got=%v",
					tt.expectedErr,
					err,
				)
			}

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

func TestGetMetadataSlice(t *testing.T) {
	tests := []struct {
		input string
		start int
		end   int
	}{
		{`
+++
author = "Jane Doe"
title = "Test markdown"
description = "A basic markdown file"
+++

Hello, World
`,
			2, 5},
		{"hey there's no metadata here", -1, -1},
		{"hey \n+++ there's no metadata here", -1, -1},
		{"hey \n+++\n there's\n ++\n no metadata here", -1, -1},
		{"hey \n++++\n there's\n +++\n no metadata here", -1, -1},
		{"hey \n+++\n there's\n+++\n no metadata here", -1, -1},
		{"\n+++\n there's\n+++\n no metadata here", 2, 3},
		{"hmm\n+++\n there's\n+++\n no metadata here", -1, -1},
		{"\n    \t\n+++\n there's\n+++\n no metadata here", 3, 4},
		{"\n    \t\n+++\n there's\n\n no metadata\n+++here", -1, -1},
		{"\n    \t\n+++\n there's\n\n no metadata\n+++\nhere", 3, 6},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			lines := bytes.Split([]byte(tt.input), []byte{'\n'})
			start, end := getMetadataSlice(lines)

			if start != tt.start {
				t.Errorf("wrong start. expected=%d. got=%d", tt.start, start)
			}

			if end != tt.end {
				t.Errorf("wrong end. expected=%d. got=%d", tt.end, end)
			}
		})
	}
}

func TestFindByteSlice(t *testing.T) {
	tests := []struct {
		input    string
		target   string
		expected int
	}{
		{"Hey\n\n", "Hey", 0},
		{"\nHey\n\n", "Hey", 1},
		{"\n   \n\twhoops\n", "Hey", -1},
		{"jaskdjkas\n   \n\twhoops\n", "Hey", -1},
		{"jaskdjkas\n   \n\twhoops\n", "jas", -1},
		{"+++\n   \n\twhoops\n", "+++", 0},
		{"      +++\n   \n\twhoops\n", "+++", 0},
		{"  \n   \t +++      \t\n   \n\twhoops\n", "+++", 1},
		{
			" whasjdkajs \n asdjksad \n   \t +++      \t\n   \n\twhoops\n",
			"+++",
			2,
		},
		{
			`
+++
author = "Jane Doe"
title = "Test markdown"
description = "A basic markdown file"
+++

Hello, World
`,
			"+++", 1,
		},
		{
			`
author = "Jane Doe"
title = "Test markdown"
description = "A basic markdown file"
+++

Hello, World
`,
			"+++", 4,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			lines := bytes.Split([]byte(tt.input), []byte{'\n'})
			index := findByteSlice(lines, []byte(tt.target))
			if index != tt.expected {
				t.Errorf(
					"wrong index. expected=%d, got=%d.",
					tt.expected,
					index,
				)
			}
		})
	}
}
