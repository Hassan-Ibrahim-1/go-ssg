package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

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
		{
			Name:    "static",
			Type:    site.DirectoryNode,
			Content: nil,
			Children: []site.Node{
				{
					Name:    "static/images",
					Type:    site.DirectoryNode,
					Content: nil,
					Children: []site.Node{
						{
							Name:     "static/images/image.png",
							Type:     site.HTMLNode,
							Content:  []byte("image"),
							Children: nil,
						},
					},
				},
			},
		}}
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
		{defaultTestSite(), "/content/foo.html", "content index"},
		{defaultTestSite(), "/inner.html", "index"},
		{defaultTestSite(), "/content.html", "index"},
		{defaultTestSite(), "/content/", "content index"},
		{defaultTestSite(), "/", "index"},
		{defaultTestSite(), "/index.html", "index"},
		{defaultTestSite(), "/static/images/", "index"},
		{defaultTestSite(), "/static/images/image.png", "image"},
		{defaultTestSite(), "/static/images/sh.png", "index"},
		{defaultTestSite(), "/staticsh.png", "index"},
		{defaultTestSite(), "/static/sh.png", "index"},
		// TODO: {defaultTestSite(), "/static", "index"},
		{defaultTestSite(), "/dne", "index"},
		{defaultTestSite(), "/dne/", "index"},
		{defaultTestSite(), "/content/", "content index"},
		{defaultTestSite(), "/content/res/dne", "content index"},
		{defaultTestSite(), "/content/res/dne.html", "content index"},
		{defaultTestSite(), "/dne.html", "index"},
		{defaultTestSite(), "/contents/res/dne.html", "index"},
		{
			[]site.Node{
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
					},
				},
			}, "/content/foo.html", "index",
		},
	}

	for i, tt := range tests {
		t.Run(
			fmt.Sprintf("test_%d_req_path_%s", i, tt.requestPath),
			func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
				rc := httptest.NewRecorder()

				n, err := newNodeHandler(tt.nodes)
				if err != nil {
					t.Fatal(err)
				}
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
			},
		)
	}
}
