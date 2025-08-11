package site

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hassan-Ibrahim-1/go-ssg/markdown"
)

type NodeType int

const (
	HTMLNode NodeType = iota
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

type Entry interface {
	Name() string

	// []byte should be nil if Type() is DirectoryEntry
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
			return nil, err
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
		return nil, err
	}

	entries := make([]Entry, len(dirContents))
	for i, entry := range dirContents {
		entries[i], err = newDirectoryEntry(entry, dir)
		if err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func Build(dir string) ([]Node, error) {
	entries, err := loadDirectoryEntries(dir)
	if err != nil {
		return nil, err
	}
	return BuildFromEntries(entries), nil
}

func buildNodes(entries []Entry) []Node {
	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		node := buildNode(entry)
		if node == nil {
			continue
		}
		nodes = append(nodes, *node)
	}
	return nodes
}

func buildNode(entry Entry) *Node {
	switch entry.Type() {
	case DirectoryEntry:
		children := entry.Children()
		if len(children) == 0 {
			return nil
		}
		return &Node{
			Name:     entry.Name(),
			Type:     DirectoryNode,
			Content:  nil,
			Children: buildNodes(children),
		}

	case FileEntry:
		content := entry.Content()
		content = markdown.ToHTML(content)
		return &Node{
			Name:     entry.Name(),
			Type:     HTMLNode,
			Children: nil,
			Content:  content,
		}
	}
	// unreachable
	panic(fmt.Sprint("unreachabe. invalid entry type:", entry.Type()))
}

func BuildFromEntries(entries []Entry) []Node {
	return buildNodes(entries)
}
