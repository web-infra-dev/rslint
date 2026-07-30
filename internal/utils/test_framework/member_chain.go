package test_framework

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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

func MemberEntriesRange(entries []MemberEntry) (core.TextRange, bool) {
	if len(entries) == 0 || entries[0].Node == nil || entries[len(entries)-1].Node == nil {
		return core.TextRange{}, false
	}
	return core.NewTextRange(entries[0].Node.Pos(), entries[len(entries)-1].Node.End()), true
}

// ResolveFirstIdentifier walks the left side of a call/member chain and
// returns its first identifier, if any.
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
	default:
		return nil
	}
}
