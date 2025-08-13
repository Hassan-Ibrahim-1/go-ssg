package server

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

func New(addr string, nodes []site.Node) (*http.Server, error) {
	// indexHTML, err := generateIndexHTML(
	// 	rootPageInfo{"A blog", filterNodes(nodes, site.HTMLNode)},
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate index.html: %w", err)
	// }

	handler := newNodeHandler(nodes)

	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}, nil
}

func newNodeHandler(nodes []site.Node) *http.ServeMux {
	mux := http.NewServeMux()
	newNodeMux(nodes, mux)
	for _, node := range nodes {
		if isIndex(node.Name) {
			mux.HandleFunc(
				"/",
				func(w http.ResponseWriter, r *http.Request) {
					fmt.Println("/ is handling the request for", r.URL.Path)
					w.Header().Set("Content-Type", "text/html")
					w.Write(node.Content)
				},
			)
			break
		}
	}

	return mux
}

func newNodeMux(nodes []site.Node, mux *http.ServeMux) {
	if len(nodes) == 0 || mux == nil {
		return
	}

	for _, node := range nodes {
		mux.HandleFunc(
			"/"+node.Name,
			func(w http.ResponseWriter, r *http.Request) {
				// fmt.Println("/" + node.Name + " is handling the request")
				switch node.Type {
				case site.HTMLNode:
					fmt.Println(
						"/"+node.Name+" is handling the request for",
						r.URL.Path,
					)
					w.Header().Set("Content-Type", "text/html")
					w.Write(node.Content)

				case site.DirectoryNode:
					if index := indexNode(node); index != nil {
						w.Header().Set("Content-Type", "text/html")
						w.Write(index.Content)
						return
					}
					// does this make sense?
					http.NotFound(w, r)

				default:
					http.NotFound(w, r)
				}
			},
		)

		if len(node.Children) != 0 {
			newNodeMux(node.Children, mux)

			if index := indexNode(node); index != nil {
				mux.HandleFunc(
					"/"+node.Name+"/",
					func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "text/html")
						w.Write(index.Content)
					},
				)
			}
		}
	}
}

func indexNode(node site.Node) *site.Node {
	for _, child := range node.Children {
		if isIndex(child.Name) {
			return &child
		}
	}
	return nil
}

func isIndex(nodeName string) bool {
	return strings.HasSuffix(nodeName, "index.html")
}

//go:embed templates/index.html
var indexRes string
var indexTmpl = template.Must(template.New("index").Parse(indexRes))

type rootPageInfo struct {
	Title string
	Blogs []site.Node
}

func generateIndexHTML(rpi rootPageInfo) ([]byte, error) {
	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, rpi)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
