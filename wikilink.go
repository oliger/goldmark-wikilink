package wikilink

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/oliger/goldmark-wikilink/ast"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	delimiterCount             = 2
	openDelimiter         byte = '['
	closeDelimiter        byte = ']'
	defaultAliasDelimiter byte = ':'
)

// ResolveDestinationFunc maps a raw destination to a file path and determines
// if the file exists.
type ResolveDestinationFunc func(rawDest []byte) (string, bool)

func defaultResolveDestinationFunc(rawDest []byte) (string, bool) {
	dest := strings.TrimSpace(string(rawDest))
	dest = strings.ToLower(dest)
	dest = strings.ReplaceAll(dest, " ", "-")
	dest = url.PathEscape(dest)

	return dest, true
}

type wikiLinkParser struct {
	aliasDelimiter         byte
	resolveDestinationFunc ResolveDestinationFunc
}

func (p *wikiLinkParser) Trigger() []byte {
	return []byte{openDelimiter}
}

func (p *wikiLinkParser) Parse(parent gast.Node, block text.Reader, ctx parser.Context) gast.Node {
	line, _ := block.PeekLine()

	// Skip lines containing less than 4 characters as they cannot contains
	// wiki links.
	if len(line) <= 4 || line[1] != openDelimiter {
		return nil
	}

	// Skip open delimiters.
	openPos := delimiterCount
	closePos := -1
	aliasPos := -1
	for i := openPos; i < len(line)-1; i++ {
		if line[i] == closeDelimiter && line[i+1] == closeDelimiter {
			closePos = i
			break
		}

		if line[i] == p.aliasDelimiter {
			aliasPos = i
		}
	}

	// Ignore empty wiki links.
	if openPos == closePos || closePos == -1 {
		return nil
	}

	wl := &ast.WikiLink{}
	if aliasPos == -1 || openPos == aliasPos || aliasPos+1 == closePos {
		wl.RawDestination = line[openPos:closePos]
		wl.Alias = wl.RawDestination
	} else {
		wl.RawDestination = line[openPos:aliasPos]
		wl.Alias = line[aliasPos+1 : closePos]
	}
	dest, exists := p.resolveDestinationFunc(wl.RawDestination)
	wl.Destination = dest
	wl.Exists = exists

	block.Advance(closePos + delimiterCount)

	return wl
}

// RenderFunc renders a wiki link.
type RenderFunc func(wl *ast.WikiLink) string

func defaultRenderFunc(wl *ast.WikiLink) string {
	return fmt.Sprintf(`<a href="%s">%s</a>`, wl.Destination, wl.Alias)
}

type wikilinkRenderer struct {
	renderFunc RenderFunc
}

func (r *wikilinkRenderer) RegisterFuncs(register renderer.NodeRendererFuncRegisterer) {
	register.Register(ast.KindWikiLink, r.render)
}

func (r *wikilinkRenderer) render(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}

	wl := n.(*ast.WikiLink)
	w.Write([]byte(r.renderFunc(wl)))

	return gast.WalkContinue, nil
}

type wikiLinkExtension struct {
	aliasDelimiter         byte
	resolveDestinationFunc ResolveDestinationFunc
	renderFunc             RenderFunc
}

func newWithDefaultOptions() *wikiLinkExtension {
	return &wikiLinkExtension{
		resolveDestinationFunc: defaultResolveDestinationFunc,
		aliasDelimiter:         defaultAliasDelimiter,
		renderFunc:             defaultRenderFunc,
	}
}

// WikiLink is a goldmark extension configured with default options.
var WikiLink = newWithDefaultOptions()

// New returns a goldmark extension that can be configured with options.
func New(opts ...Option) goldmark.Extender {
	ext := newWithDefaultOptions()
	for _, opt := range opts {
		opt(ext)
	}
	return ext
}

func (ext *wikiLinkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(
				&wikiLinkParser{
					aliasDelimiter:         ext.aliasDelimiter,
					resolveDestinationFunc: ext.resolveDestinationFunc,
				},
				150,
			),
		),
	)

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(
				&wikilinkRenderer{
					renderFunc: ext.renderFunc,
				},
				500,
			),
		),
	)
}

// Option is used to configure the extension.
type Option func(*wikiLinkExtension)

// WithAliasDelimiter sets alias delimiter.
func WithAliasDelimiter(aliasDelimiter byte) Option {
	return func(ext *wikiLinkExtension) {
		ext.aliasDelimiter = aliasDelimiter
	}
}

// WithResolveDestinationFunc sets the function used to resolve destination.
func WithResolveDestinationFunc(resolveDestinationFunc ResolveDestinationFunc) Option {
	return func(ext *wikiLinkExtension) {
		ext.resolveDestinationFunc = resolveDestinationFunc
	}
}

// WithRenderFunc sets the function used to render wikilinks.
func WithRenderFunc(renderFunc RenderFunc) Option {
	return func(ext *wikiLinkExtension) {
		ext.renderFunc = renderFunc
	}
}
