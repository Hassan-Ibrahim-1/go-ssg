package markdown

import (
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

type HTMLDoc struct {
	Metadata map[string]string
	Content  []byte
}

func ToHTML(md []byte) (HTMLDoc, error) {
	unsanitized := markdown.ToHTML(md, nil, nil)
	_ = bluemonday.UGCPolicy().SanitizeBytes(unsanitized)
	return HTMLDoc{}, nil
}
