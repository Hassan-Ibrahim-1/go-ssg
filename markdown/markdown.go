package markdown

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

type HTMLDoc struct {
	Metadata map[string]string
	Content  []byte
}

func ToHTML(md []byte) (HTMLDoc, error) {
	metadata, content, err := parseMetadata(md)
	if err != nil {
		return HTMLDoc{}, fmt.Errorf("Failed to parse metadata: %w", err)
	}

	html := convertMdSanitized(content)

	return HTMLDoc{
		Metadata: metadata,
		Content:  html,
	}, nil
}

func convertMdSanitized(md []byte) []byte {
	unsanitized := markdown.ToHTML(md, nil, nil)
	return bluemonday.UGCPolicy().SanitizeBytes(unsanitized)
}
