package site

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type testEntry struct {
	name     string
	typ      EntryType
	content  string
	children []Entry
}

func (te *testEntry) Name() string {
	return te.name
}

func (te *testEntry) Type() EntryType {
	return te.typ
}

func (te *testEntry) Content() []byte {
	return []byte(te.content)
}

func (te *testEntry) Children() []Entry {
	return te.children
}

func defaultSsgTomlEntry() Entry {
	return &testEntry{
		name:     "ssg.toml",
		typ:      FileEntry,
		content:  defaultSsgToml(),
		children: nil,
	}
}

func defaultSsgToml() string {
	return `
title = "test blog"
author = "test author"
theme = "dark"
`
}

func defaultSsgTomlNode() Node {
	return Node{
		Name:     "ssg.toml",
		Type:     FileNode,
		Content:  []byte(defaultSsgToml()),
		Children: nil,
	}
}

func defaultDarkTheme() string {
	return "p {color: red;}"
}

func defaultThemeDirEntry() Entry {
	return &testEntry{
		name:    "themes",
		typ:     DirectoryEntry,
		content: "",
		children: []Entry{
			&testEntry{
				name:     "themes/dark.css",
				typ:      FileEntry,
				content:  defaultDarkTheme(),
				children: nil,
			},
		},
	}
}

func defaultThemeDirNode() Node {
	return Node{
		Name:    "themes",
		Type:    DirectoryNode,
		Content: nil,
		Children: []Node{
			{
				Name:     "themes/dark.css",
				Type:     FileNode,
				Content:  []byte(defaultDarkTheme()),
				Children: nil,
			},
		},
	}
}

func TestStripEntryPrefix(t *testing.T) {
	tests := []struct {
		prefix   string
		name     string
		children []string

		expectedName     string
		expectedChildren []string
	}{
		{
			"website",
			"website/content",
			[]string{"website/content/index.html"},
			"content",
			[]string{"content/index.html"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			de := &directoryEntry{
				name:     tt.name,
				children: make([]Entry, len(tt.children)),
			}
			for i, childName := range tt.children {
				de.children[i] = &directoryEntry{name: childName}
			}

			stripEntryPrefix(de, tt.prefix)

			if de.name != tt.expectedName {
				t.Errorf(
					"directoryEntry's name not properly stripped. expected=%q. got=%q",
					tt.expectedName,
					de.name,
				)
			}

			for j, child := range de.children {
				if child.Name() != tt.expectedChildren[i] {
					t.Errorf(
						"child %d's name not properly stripped. expected=%q. got=%q",
						j,
						tt.expectedChildren[i],
						child.Name(),
					)
				}
			}
		})
	}
}

type testDirectory struct {
	name       string
	contentDir string
	innerFile  string
	outerFile  string
}

// creates the following structure
// tmp/content/inner.md
// tmp/outer.md
// themes/dark.css
// ssg.toml
func setupTestDirectory(
	t *testing.T,
	innerContent string,
	outerContent string,
) (testDirectory, error) {
	tmpDir := t.TempDir()

	err := os.Mkdir(filepath.Join(tmpDir, "content"), 0755)
	if err != nil {
		return testDirectory{}, err
	}

	err = os.WriteFile(
		filepath.Join(tmpDir, "content", "inner.md"),
		[]byte(innerContent),
		0644,
	)
	if err != nil {
		return testDirectory{}, err
	}

	err = os.WriteFile(
		filepath.Join(tmpDir, "outer.md"),
		[]byte(outerContent),
		0644,
	)
	if err != nil {
		return testDirectory{}, err
	}

	err = os.WriteFile(
		filepath.Join(tmpDir, "ssg.toml"),
		[]byte(defaultSsgToml()),
		0644,
	)
	if err != nil {
		return testDirectory{}, err
	}

	err = os.Mkdir(filepath.Join(tmpDir, "themes"), 0755)
	if err != nil {
		return testDirectory{}, err
	}
	err = os.WriteFile(
		filepath.Join(tmpDir, "themes/dark.css"),
		[]byte(defaultDarkTheme()),
		0644,
	)
	if err != nil {
		return testDirectory{}, err
	}

	return testDirectory{
		name:       tmpDir,
		contentDir: filepath.Join(tmpDir, "content"),
		innerFile:  filepath.Join(tmpDir, "content", "inner.md"),
		outerFile:  filepath.Join(tmpDir, "outer.md"),
	}, nil
}

func testEntriesEqual(t *testing.T, entries, expected []Entry) bool {
	if len(entries) != len(expected) {
		t.Errorf(
			"expected length=%d. got=%d. expected=%+v. got=%+v",
			len(expected),
			len(entries),
			sprintEntryNames(expected),
			sprintEntryNames(entries),
		)
		return false
	}

	passed := true

	for i := range entries {
		entry := entries[i]
		expected := expected[i]

		if !testEntryEqual(t, entry, expected) {
			passed = false
		}
	}

	return passed
}

func testEntryEqual(t *testing.T, entry, expected Entry) bool {
	passed := true

	if entry.Name() != expected.Name() {
		t.Errorf(
			"wrong name. expected=%q. got=%q",
			expected.Name(),
			entry.Name(),
		)
		passed = false
	}

	if !bytes.Equal(entry.Content(), expected.Content()) {
		t.Errorf(
			"wrong content. expected=%q. got=%q",
			string(expected.Content()),
			string(entry.Content()),
		)
		passed = false
	}

	if entry.Type() != expected.Type() {
		t.Errorf(
			"wrong type. expected=%q. got=%q",
			expected.Type(),
			entry.Type(),
		)
		passed = false
	}

	if !testEntriesEqual(t, entry.Children(), expected.Children()) {
		passed = false
	}

	return passed
}

func testNodesEqual(t *testing.T, nodes, expected []Node) {
	if len(nodes) != len(expected) {
		t.Errorf(
			"expected length=%d. got=%d. expected=%+v. got=%+v",
			len(expected),
			len(nodes),
			sprintNodeNames(expected),
			sprintNodeNames(nodes),
		)
		return
	}

	for i := range nodes {
		node := nodes[i]
		expected := expected[i]

		testNodeEqual(t, node, expected)
	}
}

func testNodeEqual(t *testing.T, node, expected Node) {
	if node.Name != expected.Name {
		t.Errorf(
			"node names not equal. expected=%q. got=%q",
			expected.Name,
			node.Name,
		)
	}

	if node.Type != expected.Type {
		t.Errorf(
			"node types not equal for node %s. expected=%d. got=%d",
			expected.Name,
			expected.Type,
			node.Type,
		)
	}

	if !bytes.Equal(node.Content, expected.Content) {
		t.Errorf(
			"node contents not equal. expected=%s.\n got=%s",
			string(expected.Content),
			string(node.Content),
		)
	}

	testNodesEqual(t, node.Children, expected.Children)
}

func sprintNodeNames(nodes []Node) string {
	names := make([]string, len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
	}
	return fmt.Sprintf("%+v", names)
}

func sprintEntryNames(entries []Entry) string {
	names := make([]string, len(entries))
	for i, node := range entries {
		names[i] = node.Name()
	}
	return fmt.Sprintf("%+v", names)
}
