package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func defaultTestSite() []site.Node {
	return []site.Node{
		{
			Name:     "index.html",
			Type:     site.HTMLNode,
			Content:  []byte("index"),
			Children: nil,
		},
		{
			Name:    "content",
			Type:    site.DirectoryNode,
			Content: nil,
			Children: []site.Node{
				{
					Name:     "content/inner.html",
					Type:     site.HTMLNode,
					Content:  []byte("content inner"),
					Children: nil,
				},
				{
					Name:     "content/index.html",
					Type:     site.HTMLNode,
					Content:  []byte("content index"),
					Children: nil,
				},
			},
		},
	}
}

func TestNodeHandler(t *testing.T) {
	tests := []struct {
		nodes       []site.Node
		requestPath string
		expected    string
	}{
		{defaultTestSite(), "/content", "content index"},
		{defaultTestSite(), "/content/index.html", "content index"},
		{defaultTestSite(), "/content/inner.html", "content inner"},
		{defaultTestSite(), "/", "index"},
		{defaultTestSite(), "/index.html", "index"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rc := httptest.NewRecorder()

			n := NodeHandler{tt.nodes, []byte("index")}
			n.ServeHTTP(rc, req)

			if rc.Code != http.StatusOK {
				t.Errorf("http response not OK. got %d", rc.Code)
			}

			if body := rc.Body.String(); body != tt.expected {
				t.Errorf(
					"unexpected body. expected=%s\n got=%s",
					tt.expected,
					body,
				)
			}
		})
	}
}
