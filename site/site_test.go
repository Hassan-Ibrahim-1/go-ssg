package site

import (
	"bytes"
	"fmt"
	"html/template"
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
	testContentMarkdown := "+++\ntitle= Test Author\n+++\nHello World"

	doc, err := markdown.ToHTML([]byte(testContentMarkdown))
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	testContentHTML, err := generateBlogHTML(doc)
	if err != nil {
		t.Fatalf("failed to generate blog html: %v", err)
	}

	tests := []struct {
		entries  []Entry
		expected []Node
	}{
		{
			[]Entry{
				&testEntry{"index.md", FileEntry, testContentMarkdown, nil},
			},
			[]Node{
				{"index.html", HTMLNode, testContentHTML, nil},
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
					{"content/index.html", HTMLNode, testContentHTML, nil},
				}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			nodes, err := BuildFromEntries(tt.entries)
			if err != nil {
				t.Fatalf("Failed to build nodes: %v", err)
			}

			testNodesEqual(t, nodes, tt.expected)
		})
	}
}

func TestLoadDirectoryEntries(t *testing.T) {
	outerContent := "+++\ntitle= Test Blog\n+++\nfoo bar baz"
	innerContent := "+++\ntitle= Test Blog\n+++\nHello World"

	innerDoc, err := markdown.ToHTML([]byte(innerContent))
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	outerDoc, err := markdown.ToHTML([]byte(outerContent))
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	innerHTML, err := generateBlogHTML(innerDoc)
	if err != nil {
		t.Errorf("Failed to generate blog html: %v", err)
	}
	outerHTML, err := generateBlogHTML(outerDoc)
	if err != nil {
		t.Errorf("Failed to generate blog html: %v", err)
	}

	tmpDir, err := setupTestDirectory(t, innerContent, outerContent)
	if err != nil {
		t.Fatal("failed to setup test directory:", err)
	}

	expectedEntries := []Entry{
		&testEntry{
			"content",
			DirectoryEntry,
			"",
			[]Entry{
				&testEntry{
					"content/inner.md",
					FileEntry,
					innerContent,
					nil,
				},
			},
		},
		&testEntry{
			"outer.md",
			FileEntry,
			outerContent,
			nil,
		},
	}

	entries, err := loadDirectoryEntries(tmpDir.name)
	if err != nil {
		t.Fatal(err)
	}

	if !testEntriesEqual(t, entries, expectedEntries) {
		return
	}

	expectedNodes := []Node{
		{
			"content",
			DirectoryNode,
			nil,
			[]Node{
				{
					"content/inner.html",
					HTMLNode,
					innerHTML,
					nil,
				},
			},
		},
		{
			"outer.html",
			HTMLNode,
			outerHTML,
			nil,
		},
	}

	nodes, err := BuildFromEntries(entries)
	if err != nil {
		t.Fatalf("BuildFromEntries failed: %v", err)
	}

	testNodesEqual(t, nodes, expectedNodes)
}

func changeFileExtension(file, from, to string) string {
	extensionIndex := len(file) - len(from)
	return file[:extensionIndex] + to
}

func TestBuild(t *testing.T) {
	innerContent := "+++\ntitle= Test Blog\n+++\nHello World"
	outerContent := "+++\ntitle= Test Blog\n+++\nfoo bar baz"

	innerDoc, err := markdown.ToHTML([]byte(innerContent))
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	outerDoc, err := markdown.ToHTML([]byte(outerContent))
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	innerHTML, err := generateBlogHTML(innerDoc)
	if err != nil {
		t.Errorf("Failed to generate blog html: %v", err)
	}

	outerHTML, err := generateBlogHTML(outerDoc)
	if err != nil {
		t.Errorf("Failed to generate blog html: %v", err)
	}

	tmpDir, err := setupTestDirectory(t, innerContent, outerContent)
	if err != nil {
		t.Fatal("failed to setup test directory:", err)
	}

	expectedNodes := []Node{
		{
			"content",
			DirectoryNode,
			nil,
			[]Node{
				{
					"content/inner.html",
					HTMLNode,
					innerHTML,
					nil,
				},
			},
		},
		{
			"outer.html",
			HTMLNode,
			outerHTML,
			nil,
		},
	}

	nodes, err := Build(tmpDir.name)
	if err != nil {
		t.Fatal("failed to build tmpDir:", err)
	}

	testNodesEqual(t, nodes, expectedNodes)
}

func TestGenerateBlogHTML(t *testing.T) {
	tests := []struct {
		md string
	}{
		{`
+++
title = A Test Blog
author = Test Author
date = 01-06-2025
+++
The brown fox jumped over the lazy dog.
`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			doc, err := markdown.ToHTML([]byte(tt.md))
			if err != nil {
				t.Fatalf("ToHTML failed: %v", err)
			}

			html, err := generateBlogHTML(doc)
			if err != nil {
				t.Fatalf("generateBlogHTML failed: %v", err)
			}

			blogInfo := blogTemplate{
				Title:         doc.Metadata["title"],
				AuthorName:    doc.Metadata["author"],
				PublishedDate: doc.Metadata["date"],
				Blog:          template.HTML(doc.Content),
			}

			var expected bytes.Buffer
			err = blogTmpl.Execute(&expected, blogInfo)
			if err != nil {
				t.Fatalf("Failed to execute blog template: %v", err)
			}

			if !bytes.Equal(html, expected.Bytes()) {
				t.Errorf(
					"bad html template. expected=%s\ngot=%s.",
					expected.String(),
					string(html),
				)
			}
		})
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
