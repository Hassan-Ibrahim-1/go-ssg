package server

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
	"github.com/fsnotify/fsnotify"
)

type Server struct {
	site    site.Site
	mux     *http.ServeMux
	watcher *fsnotify.Watcher
	server  *http.Server

	watcherMu sync.Mutex
	fsCond    *sync.Cond
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	if err := s.watcher.Close(); err != nil {
		log.Println("failed to close watcher", err)
	}
	return s.Close()
}

func New(addr, dir string) (*Server, error) {
	st, err := site.Build(dir)
	if err != nil {
		log.Fatalln("failed to build site:", err)
	}

	handler, err := newNodeHandler(st.Nodes)
	if err != nil {
		return nil, err
	}

	w, err := newWatcher(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	server := &Server{
		site:    st,
		mux:     handler,
		watcher: w,
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
		},
	}
	server.fsCond = sync.NewCond(&server.watcherMu)

	server.mux.Handle("/events", server.eventHandler())

	go listenToEvents(server)

	return server, nil
}

func listenToEvents(s *Server) {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			log.Println("event", event)
			s.fsCond.Broadcast()
			time.Sleep(10 * time.Millisecond)

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (s *Server) eventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(
				w,
				"Streaming unsupported",
				http.StatusInternalServerError,
			)
			return
		}

		log.Println("registering an event handler")

		for {
			s.fsCond.L.Lock()
			s.fsCond.Wait()
			s.fsCond.L.Unlock()

			log.Println("sending a reload signal")

			w.Write([]byte("data: reload\n\n"))
			flusher.Flush()
		}
	}
}

func newWatcher(dir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// TODO: consider filepath.WalkDir
	err = addDirToWatcher(watcher, dir)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func addDirToWatcher(w *fsnotify.Watcher, dir string) error {
	w.Add(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, de := range entries {
		if de.Type().IsDir() {
			err = addDirToWatcher(w, filepath.Join(dir, de.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func newNodeHandler(nodes []site.Node) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	addNodesToMux(nodes, mux)
	for _, node := range nodes {

		// this check is done here because all the reserved names can only
		// conflict with directories at the root level.
		if isReserved(node.Name) {
			return nil, fmt.Errorf(
				"/%s is reserved for the server, choose another directory name",
				node.Name,
			)
		}

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

	return mux, nil
}

func addNodesToMux(nodes []site.Node, mux *http.ServeMux) {
	if len(nodes) == 0 || mux == nil {
		return
	}

	for _, node := range nodes {
		mux.Handle("/"+node.Name, nodeHandler(node))

		if len(node.Children) != 0 {
			addNodesToMux(node.Children, mux)

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

func nodeHandler(node site.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("/" + node.Name + " is handling the request")
		switch node.Type {
		case site.HTMLNode:
			w.Header().Set("Content-Type", "text/html")
			w.Write(node.Content)

		case site.FileNode:
			ext := filepath.Ext(node.Name)
			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
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
	})
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

func isReserved(name string) bool {
	switch name {
	case "events":
		return true
	}
	return false
}
