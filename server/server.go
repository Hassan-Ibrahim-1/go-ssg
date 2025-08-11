package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

func New(addr string, nodes []site.Node) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           NodeHandler{nodes},
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
}

type NodeHandler struct {
	nodes []site.Node
}

func (n NodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
