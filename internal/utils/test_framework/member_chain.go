package test_framework

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// GetMemberEntries extracts identifiers and static property names from a
// call/member chain. It is syntax-only; framework parsers decide which roots
// and member sequences are legal.
func GetMemberEntries(node *ast.Node) []MemberEntry {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return []MemberEntry{{
			Name: node.AsIdentifier().Text,
			Node: node,
		}}
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		left := GetMemberEntries(property.Expression)
		nameNode := property.Name()
		if name := propertyName(nameNode); name != "" {
			return append(left, MemberEntry{
				Name: name,
				Node: nameNode,
			})
		}
		return left
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		left := GetMemberEntries(element.Expression)
		nameNode := ast.SkipParentheses(element.ArgumentExpression)
		if name := elementAccessName(nameNode); name != "" {
			return append(left, MemberEntry{
				Name: name,
				Node: nameNode,
			})
		}
		return nil
	case ast.KindCallExpression:
		entries := GetMemberEntries(node.AsCallExpression().Expression)
		if len(entries) > 0 {
			entries[len(entries)-1].Call = node
		}
		return entries
	case ast.KindTaggedTemplateExpression:
		return GetMemberEntries(node.AsTaggedTemplateExpression().Tag)
	default:
		return nil
	}
}

func propertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text
	default:
		return ""
	}
}

func elementAccessName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	default:
		return ""
	}
}

// IsFunction reports nodes that declare a callable body. Rules use it to find
// the function a test callback or assertion sits in.
func IsFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsFunctionDeclaration(node) ||
		ast.IsFunctionExpressionOrArrowFunction(node) ||
		node.Kind == ast.KindMethodDeclaration ||
		node.Kind == ast.KindConstructor ||
		node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor
}

// CalleeChainName returns a dotted name for a call callee expression,
// mirroring eslint-plugin-jest's getNodeName for CallExpression callees.
//
// It differs from GetMemberEntries in two ways, both deliberate: bracket
// notation contributes a segment only when the index is a supported accessor
// name (identifier, string literal or no-substitution template) and an
// unsupported key breaks the whole chain rather than truncating it; and
// NewExpression is peeled, so new (require('x')).y becomes a chain. Rules that
// match user-configured names against a call site (expect-expect's
// assertFunctionNames, no-standalone-expect's additionalTestBlockFunctions)
// need these semantics; rules that walk a framework call chain want
// GetMemberEntries instead.
func CalleeChainName(expr *ast.Node) string {
	if expr == nil {
		return ""
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil {
		return ""
	}

	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		property := expr.AsPropertyAccessExpression()
		left := CalleeChainName(property.Expression)
		name := propertyName(property.Name())
		if left == "" || name == "" {
			return left
		}
		return left + "." + name
	case ast.KindElementAccessExpression:
		element := expr.AsElementAccessExpression()
		left := CalleeChainName(element.Expression)
		key := elementAccessName(element.ArgumentExpression)
		if left == "" || key == "" {
			return ""
		}
		return left + "." + key
	case ast.KindCallExpression:
		return CalleeChainName(expr.AsCallExpression().Expression)
	case ast.KindNewExpression:
		newExpression := expr.AsNewExpression()
		if newExpression == nil {
			return ""
		}
		return CalleeChainName(newExpression.Expression)
	case ast.KindTaggedTemplateExpression:
		return CalleeChainName(expr.AsTaggedTemplateExpression().Tag)
	default:
		return ""
	}
}

func JoinMemberEntries(entries []MemberEntry) string {
	if len(entries) == 0 {
		return ""
	}

	parts := make([]string, len(entries))
	for i, entry := range entries {
		parts[i] = entry.Name
	}
	return strings.Join(parts, ".")
}

// MemberEntriesRange returns the range spanning the first through the last
// member entry. sourceFile is required because the first entry's Pos() includes
// its leading trivia; see accessor.go.
func MemberEntriesRange(sourceFile *ast.SourceFile, entries []MemberEntry) (core.TextRange, bool) {
	if len(entries) == 0 || entries[len(entries)-1].Node == nil {
		return core.TextRange{}, false
	}
	start, ok := AccessorRange(sourceFile, entries[0].Node)
	if !ok {
		return core.TextRange{}, false
	}
	return core.NewTextRange(start.Pos(), entries[len(entries)-1].Node.End()), true
}

// ResolveFirstIdentifier walks the effective callee of a call/member chain and
// returns its first identifier, if any. A comma expression contributes only
// its right operand because that is the value JavaScript calls.
func ResolveFirstIdentifier(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindCallExpression:
		return ResolveFirstIdentifier(node.AsCallExpression().Expression)
	case ast.KindPropertyAccessExpression:
		return ResolveFirstIdentifier(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return ResolveFirstIdentifier(node.AsElementAccessExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return ResolveFirstIdentifier(node.AsTaggedTemplateExpression().Tag)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			return ResolveFirstIdentifier(binary.Right)
		}
		return nil
	default:
		return nil
	}
}
