package site

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hassan-Ibrahim-1/go-ssg/markdown"
	"github.com/Hassan-Ibrahim-1/go-ssg/toml"
)

type NodeType int

const (
	HTMLNode NodeType = iota
	// stuff like css, images, etc
	FileNode
	DirectoryNode
)

type EntryType int

const (
	DirectoryEntry EntryType = iota
	FileEntry
)

type Node struct {
	Name string
	Type NodeType
	// nil if Type is DirectoryNode
	Content  []byte
	Children []Node
}

func (n Node) String() string {
	var children strings.Builder

	for _, child := range n.Children {
		children.WriteString(child.String() + "\n")
	}

	return fmt.Sprintf(
		"{Name: %q, Type: %d, Content: %s, Children: [%s]}\n",
		n.Name,
		n.Type,
		string(n.Content),
		children.String(),
	)
}

type Entry interface {
	Name() string

	// returns nil if Type() is DirectoryEntry
	Content() []byte

	Type() EntryType

	Children() []Entry
}

type directoryEntry struct {
	name     string
	typ      EntryType
	content  []byte
	children []Entry
}

func (de *directoryEntry) Name() string {
	return de.name
}

func (de *directoryEntry) Type() EntryType {
	return de.typ
	// switch {
	// case de.dirEntry.Type().IsDir():
	// 	return DirectoryEntry
	// case de.dirEntry.Type().IsRegular():
	// 	return FileEntry
	// }
	// panic(fmt.Sprint("invalid dirEntry Type:", de.dirEntry.Type()))
}

func (de *directoryEntry) Content() []byte {
	return de.content
}

func (de *directoryEntry) Children() []Entry {
	return de.children
}

func newDirectoryEntry(
	de os.DirEntry,
	parentName string,
) (*directoryEntry, error) {
	deName := filepath.Join(parentName, de.Name())

	if de.Type().IsRegular() {
		content, err := os.ReadFile(deName)
		if err != nil {
			return nil, err
		}
		return &directoryEntry{
			typ:      FileEntry,
			name:     deName,
			children: nil,
			content:  content,
		}, nil
	}

	dirEntries, err := os.ReadDir(deName)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}

	children := make([]Entry, len(dirEntries))
	for i, entry := range dirEntries {
		children[i], err = newDirectoryEntry(entry, deName)
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to create a directory entry: %w",
				err,
			)
		}
	}

	return &directoryEntry{
		name:     deName,
		typ:      DirectoryEntry,
		children: children,
		content:  nil,
	}, nil
}

func loadDirectoryEntries(dir string) ([]Entry, error) {
	dirContents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to open directory: %w", err)
	}

	entries := make([]Entry, len(dirContents))
	for i, entry := range dirContents {
		de, err := newDirectoryEntry(entry, dir)
		if err != nil {
			return nil, err
		}

		stripEntryPrefix(de, dir)

		entries[i] = de

	}
	return entries, nil
}

func stripEntryPrefix(de *directoryEntry, prefix string) {
	de.name = strings.TrimPrefix(de.name, prefix)

	if de.name[0] == '/' {
		de.name = de.name[1:]
	}

	for _, child := range de.children {
		// yuck
		stripEntryPrefix(child.(*directoryEntry), prefix)
	}
}

func Build(dir string) (Site, error) {
	entries, err := loadDirectoryEntries(dir)
	if err != nil {
		return Site{}, err
	}
	return BuildFromEntries(entries)
}

func buildNodes(entries []Entry) ([]Node, error) {
	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		node, err := buildNode(entry)
		if err != nil {
			return nil, err
		}
		if node == nil {
			continue
		}

		nodes = append(nodes, *node)
	}

	return nodes, nil
}

func buildNode(entry Entry) (*Node, error) {
	switch entry.Type() {
	case DirectoryEntry:
		children := entry.Children()
		if len(children) == 0 {
			return nil, nil
		}
		childrenNodes, err := buildNodes(children)
		if err != nil {
			return nil, err
		}
		return &Node{
			Name:     entry.Name(),
			Type:     DirectoryNode,
			Content:  nil,
			Children: childrenNodes,
		}, nil

	case FileEntry:
		const MarkdownExtension = ".md"
		name := entry.Name()
		content := entry.Content()
		nodeType := FileNode

		// convert all markdown files to html
		if strings.HasSuffix(name, MarkdownExtension) {
			extensionIndex := len(name) - len(MarkdownExtension)
			name = name[:extensionIndex] + ".html"
			doc, err := markdown.ToHTML(content)
			if err != nil {
				return nil, fmt.Errorf("markdown.ToHTML failed: %w", err)
			}

			content, err = generateBlogHTML(doc)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to generate blog html for %s: %w",
					name,
					err,
				)
			}
			nodeType = HTMLNode
		}

		return &Node{
			Name:     name,
			Type:     nodeType,
			Children: nil,
			Content:  content,
		}, nil
	}
	// unreachable
	panic(fmt.Sprint("unreachabe. invalid entry type:", entry.Type()))
}

type blogTemplate struct {
	Title         string
	AuthorName    string
	PublishedDate string
	Blog          template.HTML
}

// dd-mm-yyyy
const dateLayout = "02-01-2006"

//go:embed templates/blog.html
var blogRes string
var blogTmpl = template.Must(template.New("blog").Parse(blogRes))

func generateBlogHTML(doc markdown.HTMLDoc) ([]byte, error) {
	author, _ := doc.Metadata["author"]

	dateString, _ := doc.Metadata["date"]

	title, ok := doc.Metadata["title"]
	if !ok {
		return nil, fmt.Errorf("blog title not found")
	}

	blogInfo := blogTemplate{
		Title:         title,
		AuthorName:    author,
		PublishedDate: dateString,
		Blog:          template.HTML(doc.Content),
	}

	var buf bytes.Buffer
	err := blogTmpl.Execute(&buf, blogInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute blog template: %w", err)
	}

	return buf.Bytes(), nil
}

//go:embed templates/index.html
var indexRes string
var indexTmpl = template.Must(template.New("index").Parse(indexRes))

type rootPageInfo struct {
	Title string
	Blogs []Node
	Theme string
}

func generateIndexNode(rpi rootPageInfo) (Node, error) {
	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, rpi)
	if err != nil {
		return Node{}, err
	}

	html := buf.Bytes()
	return Node{
		Name:     "index.html",
		Type:     HTMLNode,
		Content:  html,
		Children: nil,
	}, nil
}

func BuildFromEntries(entries []Entry) (Site, error) {
	nodes, err := buildNodes(entries)
	if err != nil {
		return Site{}, err
	}

	var tomlNode *Node

	indexFound := false
	for _, node := range nodes {
		if strings.HasSuffix(node.Name, "index.html") {
			indexFound = true
			break
		}
		if node.Name == "ssg.toml" {
			tomlNode = &node
		}
	}

	if tomlNode == nil {
		return Site{}, fmt.Errorf("no ssg.toml file found in project root")
	}

	config, err := parseConfig(nodes, tomlNode.Content)
	if err != nil {
		return Site{}, fmt.Errorf("error parsing config: %w", err)
	}

	if !indexFound {
		node, err := generateIndexNode(
			rootPageInfo{
				Title: config.Title,
				Theme: config.Theme.Name,
				Blogs: nodes,
			},
		)
		if err != nil {
			return Site{}, fmt.Errorf("error generating index node: %w", err)
		}

		nodes = append(nodes, node)
	}

	// TODO: delete the ssg.toml node here

	return Site{
		Nodes:  nodes,
		Config: config,
	}, nil
}

func parseConfig(nodes []Node, ssgToml []byte) (SiteConfig, error) {
	config, err := toml.Parse(ssgToml)
	if err != nil {
		return SiteConfig{}, fmt.Errorf(
			"failed to parse ssg.toml file: %w",
			err,
		)
	}

	title, ok := config["title"]
	if !ok {
		return SiteConfig{}, fmt.Errorf("no title provided in ssg.toml")
	}

	author, ok := config["author"]
	if !ok {
		return SiteConfig{}, fmt.Errorf("no author provided in ssg.toml")
	}

	theme, ok := config["theme"]
	if !ok {
		return SiteConfig{}, fmt.Errorf("no theme provided in ssg.toml")
	}

	themeNode, err := findThemeNode(nodes, theme)
	if err != nil {
		return SiteConfig{}, err
	}

	return SiteConfig{
		Author: author,
		Title:  title,
		Theme:  themeNode,
	}, nil
}

func findThemeNode(nodes []Node, theme string) (Node, error) {
	themesDirFound := false
	for _, node := range nodes {
		if node.Name == "themes" {
			themesDirFound = true

			for _, child := range node.Children {
				if strings.HasSuffix(child.Name, theme+".css") {
					return child, nil
				}
			}
		}
	}

	if !themesDirFound {
		return Node{}, fmt.Errorf("no themes/ directory in project root")
	}
	return Node{}, fmt.Errorf("theme %s not found in themes/", theme)
}

type SiteConfig struct {
	Author string
	Title  string
	Theme  Node
}

type Site struct {
	Nodes  []Node
	Config SiteConfig
}
