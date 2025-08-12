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
	indexHTML, err := generateIndexHTML(
		rootPageInfo{"A blog", filterNodes(nodes, site.HTMLNode)},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate index.html: %w", err)
	}

	return &http.Server{
		Addr:              addr,
		Handler:           NodeHandler{nodes, indexHTML},
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}, nil
}

type NodeHandler struct {
	nodes []site.Node
	// FIXME:
	indexHTML []byte
}

func (n NodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	node := n.resolveURLPath(r.URL.Path)
	if node == nil {
		n.homePage(w, r)
		return
	}

	if node.Type == site.DirectoryNode {
		foundNodeIndex := false
		for _, child := range node.Children {
			if strings.HasSuffix(child.Name, "/index.html") {
				node = &child
				foundNodeIndex = true
				break
			}
		}
		if !foundNodeIndex {
			n.homePage(w, r)
			return
		}
	}

	// ???
	if node.Content == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(node.Content)
}

func (n NodeHandler) homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(n.indexHTML)
}

func (n NodeHandler) resolveURLPath(path string) *site.Node {
	fmt.Println("resolving", path)

	for _, node := range n.nodes {
		if node.Name == trimSlash(path) {
			fmt.Println("resolved to", node.Name)
			return &node
		}
	}

	return nil
}

func trimSlash(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "/"), "/")
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

func filterNodes(nodes []site.Node, typ site.NodeType) []site.Node {
	filtered := make([]site.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Type == typ {
			filtered = append(filtered, node)
		}
		filtered = append(filtered, filterNodes(node.Children, typ)...)
	}
	return filtered
}
