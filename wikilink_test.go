package wikilink

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/oliger/goldmark-wikilink/ast"
	"github.com/yuin/goldmark"
)

func TestWikiLink(t *testing.T) {
	source := "[[Wiki Link:With Some Alias]] [[Regular Wiki Link]] [[]] [[:]] [[:a]] [[a:]]"
	expected := "<p><a href=\"wiki-link\">With Some Alias</a> <a href=\"regular-wiki-link\">Regular Wiki Link</a> [[]] <a href=\":\">:</a> <a href=\":a\">:a</a> <a href=\"a:\">a:</a></p>\n"

	markdown := goldmark.New(goldmark.WithExtensions(WikiLink))
	var buf bytes.Buffer
	err := markdown.Convert([]byte(source), &buf)
	if err != nil {
		panic(err)
	}
	if buf.String() != expected {
		t.Error("invalid output")
	}
}

func TestWikiLinkWithOptions(t *testing.T) {
	ext := New(
		WithAliasDelimiter('|'),
		WithResolveDestinationFunc(func(rawDest []byte) (string, bool) {
			return "path", string(rawDest) == "Exists"
		}),
		WithRenderFunc(func(wl *ast.WikiLink) string {
			if !wl.Exists {
				return string(wl.Alias)
			}

			return fmt.Sprintf(`<a href="%s">%s</a>`, wl.Destination, wl.Alias)
		}),
	)
	markdown := goldmark.New(goldmark.WithExtensions(ext))
	source := "[[Exists|Alias]] [[Does not exist]]"
	expected := "<p><a href=\"path\">Alias</a> Does not exist</p>\n"

	var buf bytes.Buffer
	err := markdown.Convert([]byte(source), &buf)
	if err != nil {
		panic(err)
	}
	if buf.String() != expected {
		t.Error("invalid output")
	}
}
