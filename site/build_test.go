package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSite(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(tmpDir, "remove.html"),
		[]byte("foo"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	site := Site{
		Nodes: []Node{
			{
				Name:    "index.html",
				Content: []byte("index.html"),
				Type:    HTMLNode,
			},
			{
				Name: "content",
				Type: DirectoryNode,
				Children: []Node{
					{
						Name:    "inner.html",
						Content: []byte("inner"),
						Type:    HTMLNode,
					},
				},
			},
			{
				Name: "themes",
				Type: DirectoryNode,
				Children: []Node{
					{
						Name:    "dark.css",
						Content: []byte("dark"),
						Type:    FileNode,
					},
				},
			},
		},
	}

	err = BuildSite(site, tmpDir)
	if err != nil {
		t.Fatalf("failed to build site: %v", err)
	}

	type File struct {
		name    string
		content string
	}

	expected := make(map[File]bool, len(site.Nodes))
	for _, node := range flattenNodes(site.Nodes) {
		file := File{
			name:    node.Name,
			content: string(node.Content),
		}
		expected[file] = false
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != len(expected) {
		t.Errorf("wrong number of entries")
	}

	for _, entry := range entries {
		entryFile := File{name: entry.Name()}
		if entry.Type().IsRegular() {
			b, err := os.ReadFile(
				filepath.Join(tmpDir, entryFile.name),
			)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			entryFile.content = string(b)
		}
	}
}

func flattenNodes(nodes []Node) []Node {
	var flat []Node
	for _, node := range nodes {
		flat = append(flat, node)
		if len(node.Children) > 0 {
			flat = append(flat, flattenNodes(node.Children)...)
		}
	}
	return flat
}
