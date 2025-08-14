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
		if strings.HasSuffix(node.Name, ".toml") {
			return nil
		}
		return os.WriteFile(filepath.Join(dir, node.Name), node.Content, 0644)

	case DirectoryNode:
		err := createIfNotExists(filepath.Join(dir, node.Name))
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
