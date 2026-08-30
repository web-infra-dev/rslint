package accessorutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type KeyKind uint8

const (
	KeyStatic KeyKind = iota
	KeyPrivate
	KeyDynamic
)

// Key preserves ESLint's three accessor-key equivalence classes. Static names
// compare by their normalized property name, private names only compare with
// private names, and private/dynamic names compare by their token streams.
type Key struct {
	Kind KeyKind
	Text string
	Expr *ast.Node
}

func MakeKey(node *ast.Node) Key {
	nameNode := node.Name()
	if nameNode == nil {
		return Key{Kind: KeyDynamic}
	}
	if nameNode.Kind == ast.KindPrivateIdentifier {
		return Key{Kind: KeyPrivate, Expr: nameNode}
	}
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return Key{Kind: KeyStatic, Text: name}
	}
	expr := nameNode
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr = nameNode.AsComputedPropertyName().Expression
	}
	return Key{Kind: KeyDynamic, Expr: expr}
}

func KeysEqual(sourceFile *ast.SourceFile, left Key, right Key) bool {
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case KeyStatic:
		return left.Text == right.Text
	case KeyPrivate, KeyDynamic:
		return tokenKeysEqual(sourceFile, left.Expr, right.Expr)
	default:
		return false
	}
}

// tokenKeysEqual mirrors ESLint's token-list comparison. HasSameTokens
// removes only transparent outer parentheses, preserves nested parentheses,
// and compares punctuation and operators that tsgo does not expose as child
// nodes.
func tokenKeysEqual(sourceFile *ast.SourceFile, left *ast.Node, right *ast.Node) bool {
	return utils.HasSameTokens(sourceFile, left, right)
}
