package site

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hassan-Ibrahim-1/go-ssg/markdown"
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

func TestBuildFromEntries(t *testing.T) {
	testContentMarkdown := "Hello World"
	testContentHTML := markdown.ToHTML([]byte(testContentMarkdown))
	tests := []struct {
		entries  []Entry
		expected []Node
	}{
		{
			[]Entry{
				&testEntry{"index.md", FileEntry, testContentMarkdown, nil},
			},
			[]Node{
				{"index.md", HTMLNode, testContentHTML, nil},
			},
		},
		{
			[]Entry{
				&testEntry{"content/", DirectoryEntry, "", []Entry{
					&testEntry{
						"content/index.md",
						FileEntry,
						testContentMarkdown,
						nil,
					},
				}},
			},
			[]Node{
				{"content/", DirectoryNode, nil, []Node{
					{"content/index.md", HTMLNode, testContentHTML, nil},
				}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			nodes := BuildFromEntries(tt.entries)
			testNodesEqual(t, nodes, tt.expected)
		})
	}
}

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()

	outerContent := "foo bar baz"
	innerContent := "Hello World"

	// innerHTML := markdown.ToHTML([]byte(innerContent))
	// outerHTML := markdown.ToHTML([]byte(outerContent))

	err := os.Mkdir(filepath.Join(tmpDir, "content"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(tmpDir, "content", "inner.md"),
		[]byte(innerContent),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(tmpDir, "outer.md"),
		[]byte(outerContent),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedEntries := []Entry{
		&testEntry{
			filepath.Join(tmpDir, "content"),
			DirectoryEntry,
			"",
			[]Entry{
				&testEntry{
					filepath.Join(tmpDir, "content", "inner.md"),
					FileEntry,
					innerContent,
					nil,
				},
			},
		},
		&testEntry{
			filepath.Join(tmpDir, "outer.md"),
			FileEntry,
			outerContent,
			nil,
		},
	}

	entries, err := loadDirectoryEntries(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if !testEntriesEqual(t, entries, expectedEntries) {
		return
	}
}

func testEntriesEqual(t *testing.T, entries, expected []Entry) bool {
	if len(entries) != len(expected) {
		t.Errorf(
			"expected length=%d. got=%d. expected=%+v. got=%+v",
			len(expected),
			len(entries),
			expected,
			entries,
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
			expected,
			nodes,
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
			"node types not equal. expected=%d. got=%d",
			expected.Type,
			node.Type,
		)
	}

	if !bytes.Equal(node.Content, expected.Content) {
		t.Errorf(
			"node contents not equal. expected=%q. got=%q",
			string(expected.Content),
			string(node.Content),
		)
	}

	testNodesEqual(t, node.Children, expected.Children)
}
