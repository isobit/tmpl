package functions

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type MarkdownFuncs struct {
	md goldmark.Markdown
}

func NewMarkdownFuncs() MarkdownFuncs {
	return MarkdownFuncs{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
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
