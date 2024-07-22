package functions

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/mermaid"
)

type MarkdownFuncs struct {
	md goldmark.Markdown
}

func NewMarkdownFuncs() MarkdownFuncs {
	return MarkdownFuncs{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				&mermaid.Extender{},
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
	}
}

func (mf MarkdownFuncs) MarkdownToHTML(text string) string {
	var buf bytes.Buffer
	if err := mf.md.Convert([]byte(text), &buf); err != nil {
		panic(err)
	}
	return buf.String()
}
