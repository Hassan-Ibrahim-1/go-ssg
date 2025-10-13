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

func TestLoadDirectoryEntries(t *testing.T) {
	outerContent := "outer.md"
	innerContent := "inner.md"

	testDir, err := setupTestDirectory(t, innerContent, outerContent)
	if err != nil {
		t.Fatal(err)
	}

	expectedEntries := []Entry{
		&testEntry{
			name: "content",
			typ:  DirectoryEntry,
			children: []Entry{
				&testEntry{
					name:    "content/inner.md",
					typ:     FileEntry,
					content: innerContent,
				},
			},
		},
		&testEntry{
			name:    "index.html",
			typ:     FileEntry,
			content: "index",
		},
		&testEntry{
			name:    "outer.md",
			typ:     FileEntry,
			content: outerContent,
		},
		&testEntry{
			name:    "ssg.toml",
			typ:     FileEntry,
			content: defaultSsgToml(),
		},
		defaultThemeDirEntry(),
	}

	entries, err := loadDirectoryEntries(testDir.name)
	if err != nil {
		t.Fatal("loadDirectoryEntries failed:", err)
	}

	_ = testEntriesEqual(t, entries, expectedEntries)
}

type testDirectory struct {
	name       string
	contentDir string
	innerFile  string
	outerFile  string
}

func TestNewSiteBuilder(t *testing.T) {
	tests := []struct {
		entries        []Entry
		expectedConfig SiteConfig
		expectedErr    error
	}{
		{
			[]Entry{
				defaultSsgTomlEntry(), defaultThemeDirEntry(),
			}, SiteConfig{"test author", "test blog", "/themes/dark.css", false, false}, nil,
		},
		{
			[]Entry{
				&testEntry{
					name:    "index.html",
					content: "index",
					typ:     FileEntry,
				},
			}, SiteConfig{}, fmt.Errorf("no ssg.toml file found in project root"),
		},
		{
			[]Entry{
				defaultSsgTomlEntry(),
			}, SiteConfig{}, fmt.Errorf("failed to parse config: no themes/ directory in project root"),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			sb, err := newSiteBuilder(tt.entries, BuildOptions{})
			if !errEqual(err, tt.expectedErr) {
				t.Fatalf("wrong err. expected=%q. got=%q", tt.expectedErr, err)
			}

			if sb.config != tt.expectedConfig {
				t.Errorf(
					"wrong config. expected=%v. got=%v",
					tt.expectedConfig,
					sb.config,
				)
			}
		})
	}
}

func TestIsDraft(t *testing.T) {
	tests := []struct {
		md          string
		expected    bool
		expectedErr error
	}{
		{`
+++
title=test
draft=true
+++
`,
			true, nil,
		},
		{`
+++
title=test
+++
`,
			false, nil,
		},
		{`
+++
title=test
draft = false
+++
`,
			false, nil,
		},
		{`
+++
title=test
draft = 420
+++
`,
			false, fmt.Errorf("Invalid value for draft 420. expected true or false"),
		},
		{`
+++
title=test
draft = !false
+++
`,
			false, fmt.Errorf("Invalid value for draft !false. expected true or false"),
		},
		{`
+++
title=test
draft = !true
+++
`,
			false, fmt.Errorf("Invalid value for draft !true. expected true or false"),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			doc, err := markdown.ToHTML([]byte(tt.md))
			if err != nil {
				t.Fatalf("markdown.ToHTMl failed: %v", err)
			}
			b, err := isDraft(doc)
			if !errEqual(err, tt.expectedErr) {
				t.Fatalf("wrong err. expected=%v got=%v", tt.expectedErr, err)
			}
			if b != tt.expected {
				t.Errorf(
					"wrong isDraft value. expected=%v got=%v",
					tt.expected,
					b,
				)
			}
		})
	}
}

func mdToHTML(t *testing.T, md string) markdown.HTMLDoc {
	doc, err := markdown.ToHTML([]byte(md))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestBuildDrafts(t *testing.T) {
	indexMarkdown := "+++\ntitle = Index\n+++\nindex"
	indexHTML := mdToHTML(t, indexMarkdown).Content

	draftMarkdown := `
+++
title = "some blog"
draft = true
+++
hello
`
	draftDoc := mdToHTML(t, draftMarkdown)
	draftHTML, err := generateBlogHTML(
		draftDoc,
		blogConfig{siteTitle: "test blog", theme: "/themes/dark.css"},
	)
	if err != nil {
		t.Fatal(err)
	}

	nonDraftMarkdown := `
+++
title = "some blog"
+++
hello
`
	nonDraftDoc := mdToHTML(t, nonDraftMarkdown)
	nonDraftHTML, err := generateBlogHTML(
		nonDraftDoc,
		blogConfig{siteTitle: "test blog", theme: "/themes/dark.css"},
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		buildDrafts bool
		entries     []Entry
		expected    []Node
	}{
		{
			true,
			[]Entry{
				defaultSsgTomlEntry(),
				defaultThemeDirEntry(),
				&testEntry{
					name:    "index.md",
					typ:     FileEntry,
					content: indexMarkdown,
				},
				&testEntry{
					name:    "draft.md",
					typ:     FileEntry,
					content: draftMarkdown,
				},
				&testEntry{
					name:    "nondraft.md",
					typ:     FileEntry,
					content: nonDraftMarkdown,
				},
			},
			[]Node{
				defaultSsgTomlNode(),
				defaultThemeDirNode(),
				{
					Name:    "index.html",
					Type:    HTMLNode,
					Content: indexHTML,
				},
				{
					Name:    "draft.html",
					Type:    HTMLNode,
					Content: draftHTML,
				},
				{
					Name:    "nondraft.html",
					Type:    HTMLNode,
					Content: nonDraftHTML,
				},
			},
		},
		{
			false,
			[]Entry{
				defaultSsgTomlEntry(),
				defaultThemeDirEntry(),
				&testEntry{
					name:    "index.md",
					typ:     FileEntry,
					content: indexMarkdown,
				},
				&testEntry{
					name:    "draft.md",
					typ:     FileEntry,
					content: draftMarkdown,
				},
				&testEntry{
					name:    "nondraft.md",
					typ:     FileEntry,
					content: nonDraftMarkdown,
				},
			},
			[]Node{
				defaultSsgTomlNode(),
				defaultThemeDirNode(),
				{
					Name:    "index.html",
					Type:    HTMLNode,
					Content: indexHTML,
				},
				{
					Name:    "nondraft.html",
					Type:    HTMLNode,
					Content: nonDraftHTML,
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			site, err := BuildFromEntries(
				tt.entries,
				BuildOptions{BuildDrafts: tt.buildDrafts},
			)
			if err != nil {
				t.Fatal("BuildFromEntries failed:", err)
			}
			testNodesEqual(t, site.Nodes, tt.expected)
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		entries        []Entry
		ssgToml        string
		expectedConfig SiteConfig
		expectedErr    error
	}{
		{
			[]Entry{defaultThemeDirEntry()},
			defaultSsgToml(),
			SiteConfig{
				"test author",
				"test blog",
				"/themes/dark.css",
				false,
				false,
			},
			nil,
		},
		{
			[]Entry{defaultThemeDirEntry()},
			`
author = "test author"
theme = "dark"
`,
			SiteConfig{},
			fmt.Errorf("no title provided in ssg.toml"),
		},
		{
			[]Entry{defaultThemeDirEntry()},
			`
title = "test blog"
theme = "dark"
`,
			SiteConfig{},
			fmt.Errorf("no author provided in ssg.toml"),
		},
		{
			[]Entry{defaultThemeDirEntry()},
			`
title = "test blog"
author = "test author"
`,
			SiteConfig{},
			fmt.Errorf("no theme provided in ssg.toml"),
		},
		{
			[]Entry{defaultThemeDirEntry()},
			`
title = "test blog"
author = "test author"
theme = "dne"
`,
			SiteConfig{},
			fmt.Errorf("theme dne not found in themes/"),
		},
		{
			[]Entry{},
			`
title = "test blog"
author = "test author"
theme = "dark"
`,
			SiteConfig{},
			fmt.Errorf("no themes/ directory in project root"),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			config, err := parseConfig(tt.entries, []byte(tt.ssgToml))
			if !errEqual(err, tt.expectedErr) {
				t.Fatalf("wrong err. expected=%q. got=%q", tt.expectedErr, err)
			}

			if config != tt.expectedConfig {
				t.Errorf(
					"wrong config. expected=%v. got=%v",
					tt.expectedConfig,
					config,
				)
			}
		})
	}
}

func TestBuildFromEntries(t *testing.T) {
	indexMarkdown := "+++\ntitle = Index\n+++\nindex"
	indexHTML := mdToHTML(t, indexMarkdown).Content

	innerMarkdown := `
+++
title = "some blog"
+++
hello
`
	innerContentDoc, err := markdown.ToHTML([]byte(innerMarkdown))
	if err != nil {
		t.Fatal(err)
	}

	innerHTML, err := generateBlogHTML(
		innerContentDoc,
		blogConfig{siteTitle: "test blog", theme: "/themes/dark.css"},
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		entries  []Entry
		expected []Node
	}{
		{
			[]Entry{
				defaultSsgTomlEntry(),
				defaultThemeDirEntry(),
				&testEntry{
					name:    "index.md",
					typ:     FileEntry,
					content: indexMarkdown,
				},
			},
			[]Node{
				defaultSsgTomlNode(),
				defaultThemeDirNode(),
				{
					Name:    "index.html",
					Type:    HTMLNode,
					Content: indexHTML,
				},
			},
		},
		{
			[]Entry{
				defaultSsgTomlEntry(),
				defaultThemeDirEntry(),
				&testEntry{
					name:    "index.md",
					typ:     FileEntry,
					content: indexMarkdown,
				},
				&testEntry{
					name:    "something.txt",
					typ:     FileEntry,
					content: "blah blah",
				},
			},
			[]Node{
				defaultSsgTomlNode(),
				defaultThemeDirNode(),
				{
					Name:    "index.html",
					Type:    HTMLNode,
					Content: indexHTML,
				},
				{
					Name:    "something.txt",
					Type:    FileNode,
					Content: []byte("blah blah"),
				},
			},
		},
		{
			[]Entry{
				defaultSsgTomlEntry(),
				defaultThemeDirEntry(),
				&testEntry{
					name:    "index.md",
					typ:     FileEntry,
					content: indexMarkdown,
				},
				&testEntry{
					name: "content",
					typ:  DirectoryEntry,
					children: []Entry{
						&testEntry{
							name:    "content/inner.md",
							typ:     FileEntry,
							content: innerMarkdown,
						},
					},
				},
			},
			[]Node{
				defaultSsgTomlNode(),
				defaultThemeDirNode(),
				{
					Name:    "index.html",
					Type:    HTMLNode,
					Content: indexHTML,
				},
				{
					Name: "content",
					Type: DirectoryNode,
					Children: []Node{
						{
							Name:    "content/inner.html",
							Type:    HTMLNode,
							Content: innerHTML,
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			site, err := BuildFromEntries(tt.entries, BuildOptions{})
			if err != nil {
				t.Fatal("BuildFromEntries failed:", err)
			}
			testNodesEqual(t, site.Nodes, tt.expected)
		})
	}
}

func TestGenerateBlogHTML(t *testing.T) {
	t.Skip()
	// TODO: fix this
	tests := []struct {
		markdown string
		theme    string
		expected string
	}{
		{`
+++
title = some blog
+++
hello
`,
			"/themes/dark.css", `<!doctype html>
<html lang="en">
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="stylesheet" href="/themes/dark.css" />
        <script>
            const ws = new WebSocket(` + "`" + `ws://${location.host}/fsevents` + "`" + `);
            ws.onopen = () => {
                console.log("Connected to server");
                ws.send(window.location.href);
            };

            ws.onmessage = (event) => {
                const parser = new DOMParser();
                // Parse the string into a Document
                const doc = parser.parseFromString(event.data, "text/html");

                const newTheme = document.getElementById("theme");
                const theme = document.getElementById("theme");

                const updatedTheme = theme.cloneNode();
                updatedTheme.href =
                    newTheme.href + "?v=" + new Date().getTime();

                updatedTheme.onload = () => theme.remove();

                theme.parentNode.insertBefore(updatedTheme, theme.nextSibling);

                document.title = doc.title;
                document.body.innerHTML = doc.body.innerHTML;
            };

            ws.onclose = () => {
                console.log("Disconnected from server");
            };
        </script>
        <title>some blog</title>
    </head>
    <body>
        <div id="title">
            <h1>some blog</h1>
        </div>
        <div id="metadata">
            <p id="author-name"></p>
            <p id="data-published"></p>
        </div>
        <article><p>hello</p>
        </article>
    </body>
</html>
`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			doc, err := markdown.ToHTML([]byte(tt.markdown))
			if err != nil {
				t.Fatal(err)
			}

			html, err := generateBlogHTML(doc, blogConfig{theme: tt.theme})
			if err != nil {
				t.Fatalf("generateBlogHTML failed: %v", err)
			}

			htmlLines := bytes.Split(html, []byte{'\n'})
			expectedLines := bytes.Split([]byte(tt.expected), []byte{'\n'})

			if len(htmlLines) != len(expectedLines) {
				t.Errorf(
					"invalid html expected has %d lines while result has %d\nexpected=\n%s\ngot=\n%s",
					len(expectedLines),
					len(htmlLines),
					tt.expected,
					string(html),
				)
			}
			for i, expected := range expectedLines {
				line := bytes.TrimSpace(htmlLines[i])
				expected = bytes.TrimSpace(expected)

				if !bytes.Equal(line, expected) {
					t.Errorf(
						"html not equal at line %d\nexpected=\n%s\ngot=\n%s\n",
						i+1,
						expected,
						line,
					)
				}
			}
		})
	}
}

func errEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

// creates the following structure
// tmp/content/inner.md
// tmp/outer.md
// themes/dark.css
// ssg.toml
// index.html `content: "index"`
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

	err = os.WriteFile(
		filepath.Join(tmpDir, "index.html"),
		[]byte("index"),
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
			"node contents not equal for %s. expected=%s.\n got=%s",
			expected.Name,
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
