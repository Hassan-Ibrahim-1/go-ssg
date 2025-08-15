package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildSite(s Site, dir string) error {
	err := createIfNotExists(dir)
	if err != nil {
		return err
	}

	err = clearDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to clear files from %s: %w", dir, err)
	}

	for _, node := range s.Nodes {
		err = writeNode(dir, node)
		if err != nil {
			return fmt.Errorf("failed to write node: %s: %w", node.Name, err)
		}
	}

	return nil
}

func writeNode(dir string, node Node) error {
	switch node.Type {

	case HTMLNode:
		return os.WriteFile(filepath.Join(dir, node.Name), node.Content, 0644)
	case FileNode:
		if !strings.HasSuffix(node.Name, ".css") {
			return nil
		}
		return os.WriteFile(filepath.Join(dir, node.Name), node.Content, 0644)

	case DirectoryNode:
		path := filepath.Join(dir, node.Name)
		err := createIfNotExists(path)
		if err != nil {
			return err
		}

		for _, child := range node.Children {
			err = writeNode(dir, child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func clearDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("os.ReadDir failed: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fmt.Printf("entry name: %s\n", entry.Name())

		err = os.RemoveAll(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("os.Remove failed: %w", err)
		}
	}

	return nil
}

func createIfNotExists(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				return fmt.Errorf("Failed to create %s: %w", dir, err)
			}
		} else {
			return fmt.Errorf("failed to open %s: %w", dir, err)
		}
	}
	return nil
}
