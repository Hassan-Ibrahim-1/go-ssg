package markdown

import "bytes"

// map can be nil if there is no metadata
// returns remaining contents of the markdown file not including the metadata
func parseMetadata(md []byte) (map[string]string, []byte) {
}

func getMetadataSlice(md []byte) (start, end int) {
	lines := bytes.Split(md, []byte{'\n'})

	for i := range lines {

	}
}
