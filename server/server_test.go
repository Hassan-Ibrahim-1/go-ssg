package server

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

func TestGenerateIndexHTML(t *testing.T) {
	tests := []struct {
		title string
		blogs []string
	}{
		{"A test blog", []string{"blog 1", "blog 2", "blog 3"}},
		{"", []string{}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			nodes := make([]site.Node, len(tt.blogs))
			for i := range tt.blogs {
				nodes[i] = site.Node{Name: tt.blogs[i]}
			}
			rpi := rootPageInfo{tt.title, nodes}
			html, err := generateIndexHTML(rpi)
			if err != nil {
				t.Fatalf("generteIndexHTML failed: %v", err)
			}

			var buf bytes.Buffer
			err = indexTmpl.Execute(&buf, rpi)
			if err != nil {
				t.Fatalf("indexTmpl.Execute failed: %v", err)
			}
			expected := buf.Bytes()

			if !bytes.Equal(html, expected) {
				t.Errorf(
					"bad html template. expected=%s. got=%s",
					expected,
					html,
				)
			}
		})
	}
}
