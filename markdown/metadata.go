package markdown

import (
	"bytes"
	"fmt"
)

// map can be nil if there is no metadata
// returns remaining contents of the markdown file not including the metadata
func parseMetadata(md []byte) (map[string]string, []byte, error) {
	lines := bytes.Split(md, []byte{'\n'})

	start, end := getMetadataSlice(lines)
	// no metadata
	if start == -1 || end == -1 {
		return nil, nil, nil
	}

	metadata := lines[start:end]

	keyValues := make(map[string]string)
	for _, line := range metadata {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		kv := bytes.Split(trimmed, []byte{'='})
		if len(kv) != 2 {
			return nil, nil, fmt.Errorf(
				"Not a valid key value pair %q, expected key = value",
				line,
			)
		}

		trimmedKey := string(bytes.TrimSpace(kv[0]))
		trimmedValue := string(bytes.TrimSpace(kv[1]))
		keyValues[trimmedKey] = trimmedValue
	}

	// remove the '+++'
	remaining := bytes.Join(lines[end+2:], []byte{'\n'})
	return keyValues, remaining, nil
}

// returns -1, -1 if no metadata is found
func getMetadataSlice(lines [][]byte) (start, end int) {
	start = findByteSlice(lines, []byte("+++"))
	if start == -1 {
		return -1, -1
	}

	if !isLinesWhitespace(lines[:start]) {
		return -1, -1
	}

	end = findByteSlice(lines[start+1:], []byte("+++"))
	if end == -1 {
		return -1, -1
	}

	// +1 to remove the '+++' delimiters
	return start + 1, end + start + 1
}

// -1 if not found
// skips whitespace in lines to find expected
func findByteSlice(lines [][]byte, target []byte) int {
	for i, line := range lines {
		s := bytes.TrimSpace(line)
		if bytes.Equal(s, []byte(target)) {
			return i
		}
	}
	return -1
}

func isWhitespace(line []byte) bool {
	return len(bytes.TrimSpace(line)) == 0
}

func isLinesWhitespace(lines [][]byte) bool {
	for _, line := range lines {
		if !isWhitespace(line) {
			return false
		}
	}
	return true
}
