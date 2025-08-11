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

	unsanitized := markdown.ToHTML(content, nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsanitized)

	return HTMLDoc{
		Metadata: metadata,
		Content:  html,
	}, nil
}
