package markdown

import (
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

func ToHTML(md []byte) []byte {
	unsanitized := markdown.ToHTML(md, nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsanitized)
	return html
}
