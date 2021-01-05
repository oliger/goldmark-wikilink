package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// KindWikiLink is a NodeKind of the wiki link node.
var KindWikiLink = gast.NewNodeKind("WikiLink")

// A WikiLink represents a wiki link.
type WikiLink struct {
	gast.BaseInline
	Alias          []byte
	RawDestination []byte
	Destination    string
	Exists         bool
}

// Dump implements Node.Dump.
func (n *WikiLink) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements Node.Kind.
func (n *WikiLink) Kind() gast.NodeKind {
	return KindWikiLink
}
