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

type siteBuilder struct {
	config SiteConfig
}

func newSiteBuilder(entries []Entry) (siteBuilder, error) {
	for _, entry := range entries {
		if entry.Name() == "ssg.toml" {
			ssgToml := entry.Content()
			config, err := parseConfig(entries, ssgToml)
			if err != nil {
				return siteBuilder{}, fmt.Errorf(
					"failed to parse config: %w",
					err,
				)
			}
			return siteBuilder{config}, nil
		}
	}
	return siteBuilder{}, fmt.Errorf("no ssg.toml file found in project root")
}

func (sb *siteBuilder) build(entries []Entry) (Site, error) {
	nodes, err := sb.buildNodes(entries)
	if err != nil {
		return Site{}, err
	}

	indexFound := false
	for i, node := range nodes {
		if strings.HasSuffix(node.Name, "index.html") {
			indexFound = true

			// this is really stupid, buildNode formats index.html
			// using generateBlogHTML and we don't want that for index.html files
			// so we reset it here. should think of a better way for doing this
			// maybe only files in content/ get the blog.html template
			doc, err := markdown.ToHTML(entries[i].Content())
			if err != nil {
				return Site{}, fmt.Errorf("markdown.ToHTML failed: %w", err)
			}
			nodes[i].Content = doc.Content

			break
		}
	}

	if !indexFound {
		node, err := generateIndexNode(
			rootPageInfo{
				Title: sb.config.Title,
				Theme: sb.config.Theme,
				Blogs: nodes,
			},
		)
		if err != nil {
			return Site{}, fmt.Errorf("error generating index node: %w", err)
		}

		nodes = append(nodes, node)
	}

	return Site{
		Nodes:  nodes,
		Config: sb.config,
	}, nil
}

func (sb *siteBuilder) buildNodes(entries []Entry) ([]Node, error) {
	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		node, err := sb.buildNode(entry)
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

func (sb *siteBuilder) buildNode(entry Entry) (*Node, error) {
	switch entry.Type() {
	case DirectoryEntry:
		children := entry.Children()
		if len(children) == 0 {
			return nil, nil
		}
		childNodes, err := sb.buildNodes(children)
		if err != nil {
			return nil, err
		}
		return &Node{
			Name:     entry.Name(),
			Type:     DirectoryNode,
			Content:  nil,
			Children: childNodes,
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

			content, err = generateBlogHTML(doc, sb.config.Theme)
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
	Theme         string
	PublishedDate string
	Blog          template.HTML
}

// dd-mm-yyyy
const dateLayout = "02-01-2006"

//go:embed templates/blog.html
var blogRes string
var blogTmpl = template.Must(template.New("blog").Parse(blogRes))

func generateBlogHTML(doc markdown.HTMLDoc, theme string) ([]byte, error) {
	author, _ := doc.Metadata["author"]

	dateString, _ := doc.Metadata["date"]

	title, ok := doc.Metadata["title"]
	if !ok {
		return nil, fmt.Errorf("blog title not found")
	}

	blogInfo := blogTemplate{
		Title:         title,
		Theme:         theme,
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
	var html bytes.Buffer
	err := indexTmpl.Execute(&html, rpi)
	if err != nil {
		return Node{}, err
	}

	return Node{
		Name:     "index.html",
		Type:     HTMLNode,
		Content:  html.Bytes(),
		Children: nil,
	}, nil
}

func BuildFromEntries(entries []Entry) (Site, error) {
	sb, err := newSiteBuilder(entries)
	if err != nil {
		return Site{}, fmt.Errorf("failed to build site: %w", err)
	}
	return sb.build(entries)
}

func parseConfig(entries []Entry, ssgToml []byte) (SiteConfig, error) {
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

	themeName, err := findThemeName(entries, theme)
	if err != nil {
		return SiteConfig{}, err
	}

	return SiteConfig{
		Author: author,
		Title:  title,
		Theme:  "/" + themeName,
	}, nil
}

func findThemeName(entries []Entry, theme string) (string, error) {
	themesDirFound := false
	for _, node := range entries {
		if node.Name() == "themes" {
			themesDirFound = true

			for _, child := range node.Children() {
				if strings.HasSuffix(child.Name(), theme+".css") {
					return child.Name(), nil
				}
			}
		}
	}

	if !themesDirFound {
		return "", fmt.Errorf("no themes/ directory in project root")
	}
	return "", fmt.Errorf("theme %s not found in themes/", theme)
}

type SiteConfig struct {
	Author string
	Title  string
	Theme  string
}

type Site struct {
	Nodes  []Node
	Config SiteConfig
}
