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
	html := bluemonday.UGCPolicy().SanitizeBytes(unsanitized)
	return html, nil
}
