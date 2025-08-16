package site

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSite(t *testing.T) {
	outDir := t.TempDir()

	err := os.Mkdir(filepath.Join(outDir, ".git"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(outDir, ".git/gitfile.txt"),
		[]byte("git"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(outDir, ".git/gitfile2.txt"),
		[]byte("git"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(outDir, "removeDir"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(outDir, "removeDir/removeFile.txt"),
		[]byte("removeFile.txt"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(outDir, "content"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(outDir, "content/test.html"),
		[]byte("test.html"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	siteDir := t.TempDir()
	err = os.Mkdir(filepath.Join(siteDir, ".git"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, ".git/gitfile.txt"),
		[]byte("site git"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(siteDir, "content"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "content/index.md"),
		[]byte(markdownHeader()+"content/index.md"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "content/inner.md"),
		[]byte(markdownHeader()+"content/inner.md"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "index.md"),
		[]byte(markdownHeader()+"index.md"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(siteDir, "themes"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "themes/dark.css"),
		[]byte("themes/dark.css"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "themes/light.css"),
		[]byte("themes/light.css"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(siteDir, "ssg.toml"),
		[]byte(defaultSsgToml()),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	site, err := Build(siteDir, BuildOptions{})
	if err != nil {
		t.Fatalf("failed to build site: %v", err)
	}

	err = BuildSite(site, outDir)
	if err != nil {
		t.Fatalf("failed to build site: %v", err)
	}

	files := []string{
		".git",
		".git/gitfile.txt",
		".git/gitfile2.txt",
		"index.html",
		"content",
		"content/inner.html",
		"content/index.html",
		"themes",
		"themes/dark.css",
		"themes/light.css",
	}

	expected := make(map[string]bool, len(files))
	for _, file := range files {
		expected[file] = false
	}

	err = filepath.WalkDir(
		outDir,
		func(path string, d fs.DirEntry, err error) error {
			if path == outDir {
				return nil
			}
			path = strings.TrimPrefix(path, outDir)[1:]
			if err != nil {
				t.Fatalf("failed to walk dir %s: %v", path, err)
			}

			_, ok := expected[path]
			if !ok {
				t.Errorf("unexpected file %s", path)
				return nil
			}
			expected[path] = true

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	for file, found := range expected {
		if !found {
			t.Errorf("%s was not found", file)
		}
	}
}

func markdownHeader() string {
	return "+++\ntitle=blog\n+++\n"
}

func TestBuildSiteManual(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(tmpDir, "remove.html"),
		[]byte("foo"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(tmpDir, "testdir"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(
		filepath.Join(tmpDir, "testdir/remove.txt"),
		[]byte("bar"),
		0644,
	)

	err = os.Mkdir(filepath.Join(tmpDir, "content"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(
		filepath.Join(tmpDir, "content/remove.css"),
		[]byte("bar"),
		0644,
	)

	err = os.Mkdir(filepath.Join(tmpDir, ".git"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(
		filepath.Join(tmpDir, ".git/somegitfile.txt"),
		[]byte("bar"),
		0644,
	)

	os.WriteFile(
		filepath.Join(tmpDir, ".config"),
		[]byte("config"),
		0644,
	)

	err = os.Mkdir(filepath.Join(tmpDir, "content/baz"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(
		filepath.Join(tmpDir, "content/baz/baz.txt"),
		[]byte("bar"),
		0644,
	)

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
						Name:    "content/inner.html",
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
						Name:    "themes/dark.css",
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

	files := []File{
		{
			name: ".git",
		},
		{
			name:    ".git/somegitfile.txt",
			content: "bar",
		},
		{
			name:    ".config",
			content: "config",
		},
	}

	for _, file := range files {
		expected[file] = false
	}

	err = filepath.WalkDir(
		tmpDir,
		func(path string, d fs.DirEntry, err error) error {
			if path == tmpDir {
				return nil
			}
			path = strings.TrimPrefix(path, tmpDir)[1:]
			if err != nil {
				t.Fatalf("failed to walk dir %s: %v", path, err)
			}

			entryFile := File{name: path}
			if d.Type().IsRegular() {
				b, err := os.ReadFile(
					filepath.Join(tmpDir, entryFile.name),
				)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				entryFile.content = string(b)
			}

			_, ok := expected[entryFile]
			if !ok {
				t.Errorf("unexpected file %s", entryFile.name)
				return nil
			}
			expected[entryFile] = true

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	for file, found := range expected {
		if !found {
			t.Errorf("%s was not found", file.name)
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
