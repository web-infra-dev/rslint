package accessorutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
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
// private names, and dynamic computed names compare by their token streams.
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
		return Key{Kind: KeyPrivate, Text: nameNode.AsPrivateIdentifier().Text}
	}
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return Key{Kind: KeyStatic, Text: name}
	}
	expr := nameNode
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr = nameNode.AsComputedPropertyName().Expression
	}
	return Key{Kind: KeyDynamic, Expr: ast.SkipParentheses(expr)}
}

func KeysEqual(sourceFile *ast.SourceFile, left Key, right Key) bool {
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case KeyStatic, KeyPrivate:
		return left.Text == right.Text
	case KeyDynamic:
		return computedKeysEqual(sourceFile, left.Expr, right.Expr)
	default:
		return false
	}
}

// computedKeysEqual mirrors ESLint's token-list comparison while accounting
// for parentheses that ESTree omits from child expressions. Raw numeric and
// bigint spelling remains significant because ESLint compares token values.
func computedKeysEqual(sourceFile *ast.SourceFile, left *ast.Node, right *ast.Node) bool {
	if left == nil || right == nil {
		return left == right
	}
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil {
		return left == right
	}
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindIdentifier:
		return left.AsIdentifier().Text == right.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return left.AsPrivateIdentifier().Text == right.AsPrivateIdentifier().Text
	case ast.KindStringLiteral:
		return left.AsStringLiteral().Text == right.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return left.AsNoSubstitutionTemplateLiteral().Text == right.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail,
		ast.KindRegularExpressionLiteral:
		return left.Text() == right.Text()
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, left, false) ==
			scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, right, false)
	}
	var leftChildren, rightChildren []*ast.Node
	left.ForEachChild(func(child *ast.Node) bool {
		leftChildren = append(leftChildren, child)
		return false
	})
	right.ForEachChild(func(child *ast.Node) bool {
		rightChildren = append(rightChildren, child)
		return false
	})
	if len(leftChildren) != len(rightChildren) {
		return false
	}
	for index := range leftChildren {
		if !computedKeysEqual(sourceFile, leftChildren[index], rightChildren[index]) {
			return false
		}
	}
	return true
}
