package server

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{}

type Server struct {
	site    site.Site
	mux     *http.ServeMux
	watcher *fsnotify.Watcher
	server  *http.Server
	dir     string

	clients   []chan struct{}
	clientsMu sync.Mutex
}

func (s *Server) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.ServeHTTP(w, r)
	}
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
		dir:     dir,
	}

	server.server = &http.Server{
		Addr:              addr,
		Handler:           server.handler(),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	server.mux.Handle("/fsevents", server.eventHandler())

	go listenToEvents(server)

	return server, nil
}

func (s *Server) rebuild() {
	newSite, err := site.Build(s.dir)
	if err != nil {
		log.Println("failed to build site:", err)
		return
	}

	newHandler, err := newNodeHandler(newSite.Nodes)
	if err != nil {
		log.Println("failed to build site:", err)
		return
	}

	s.site = newSite
	s.mux = newHandler
	s.mux.Handle("/fsevents", s.eventHandler())
}

func sprintNodeNames(nodes []site.Node) string {
	names := make([]string, len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
		if len(node.Children) > 0 {
			names = append(names, sprintNodeNames(node.Children))
		}
	}
	return fmt.Sprintf("%+v", names)
}

func listenToEvents(s *Server) {
	timer := time.NewTimer(time.Hour * 24 * 365)
	checkForEvents := true

	for {
		select {
		case _ = <-timer.C:
			checkForEvents = true
			timer.Reset(time.Hour * 24 * 365)

		case _, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if !checkForEvents {
				continue
			}

			// sleep to let the fs rebuild
			time.Sleep(100 * time.Millisecond)

			s.rebuild()
			s.pingClients()

			timer.Reset(200 * time.Millisecond)
			checkForEvents = false

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (s *Server) pingClients() {
	for _, c := range s.clients {
		c <- struct{}{}
	}
}

func (s *Server) addClient() (chan struct{}, int) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	i := len(s.clients)
	s.clients = append(s.clients, make(chan struct{}))
	return s.clients[i], i
}

func (s *Server) removeClient(i int) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	close(s.clients[i])
	s.clients = slices.Delete(s.clients, i, i+1)
	log.Println("removed a client", i)
}

func (s *Server) eventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade err:", err)
			return
		}
		defer conn.Close()

		c, i := s.addClient()
		defer s.removeClient(i)

		ctx := r.Context()

		// read url
		mt, rawUrl, err := conn.ReadMessage()
		if mt != websocket.TextMessage {
			log.Println("expected a TextMessage. got", mt)
			return
		}

		url, err := url.Parse(string(rawUrl))
		if err != nil {
			http.Error(w, "Bad url", http.StatusBadRequest)
			return
		}
		path := trimSlash(url.Path)

	refreshLoop:
		for {
			select {
			case _ = <-ctx.Done():
				log.Printf("Client %s closed the connection\n", r.RemoteAddr)
				return
			case _, ok := <-c:
				if !ok {
					break refreshLoop
				}

				node := matchNode(s.site.Nodes, path)
				if node == nil {
					log.Println("could not find node", path)
					return
				}

				err := conn.WriteMessage(websocket.TextMessage, node.Content)
				if err != nil {
					log.Println("error when writing to client:", err)
					break refreshLoop
				}
			}
		}
		log.Println("closing connection to", r.RemoteAddr)
	}
}

func matchNode(nodes []site.Node, name string) *site.Node {
	for _, node := range nodes {
		if node.Name == name {
			return &node
		}
		if n := matchNode(node.Children, name); n != nil {
			return n
		}
	}
	return nil
}

func trimSlash(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "/"), "/")
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
	case "fsevents":
		return true
	}
	return false
}
