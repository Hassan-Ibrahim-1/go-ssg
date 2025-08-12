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
	nodes     []site.Node
	indexHTML []byte
}

func (n NodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html")
		w.Write(n.indexHTML)
		return
	}

	node := n.resolveURLPath(r.URL.Path)
	if node == nil {
		http.NotFound(w, r)
		return
	}

	if node.Type == site.DirectoryNode {
		for _, child := range node.Children {
			if strings.HasSuffix(child.Name, "/index.html") {
				node = &child
				break
			}
		}
	}

	if node.Content == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(node.Content)
}

func (n NodeHandler) resolveURLPath(path string) *site.Node {
	fmt.Println("resolving", path)
	node := matchNodeName(path, n.nodes)
	if node == nil {
		for _, base := range n.nodes {
			if strings.HasSuffix(base.Name, "/index.html") {
				node = &base
				break
			}
		}
	}
	if node != nil {
		fmt.Println("resolved to", node.Name)
	}
	return node
}

func matchNodeName(name string, nodes []site.Node) *site.Node {
	for _, node := range nodes {
		if n := matchNodeName(name, node.Children); n != nil {
			return n
		}

		nodePath := removeUntil(node.Name, "/")
		fmt.Println("node path", nodePath)
		if nodePath != name {
			continue
		}

		return &node
	}
	return nil
}

func removeUntil(s, delim string) string {
	i := strings.Index(s, delim)
	if i != -1 {
		return s[i:]
	}
	return s
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
