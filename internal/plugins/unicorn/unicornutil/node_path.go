package unicornutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// NodeMatchesPath mirrors unicorn's isNodeMatchesNameOrPath helper. It accepts
// dotted, non-computed, non-optional paths rooted at an identifier, this,
// super, or a meta property. ast.IsDottedName is intentionally not used
// because it also accepts property accesses containing optional-chain links.
func NodeMatchesPath(node *ast.Node, path string) bool {
	parts := strings.Split(strings.TrimSpace(path), ".")
	return nodeMatchesPathParts(ast.SkipParentheses(node), parts)
}

func nodeMatchesPathParts(node *ast.Node, parts []string) bool {
	if node == nil || len(parts) == 0 {
		return false
	}

	if len(parts) == 1 {
		switch node.Kind {
		case ast.KindIdentifier:
			return node.AsIdentifier().Text == parts[0]
		case ast.KindThisKeyword:
			return parts[0] == "this"
		case ast.KindSuperKeyword:
			return parts[0] == "super"
		default:
			return false
		}
	}

	if len(parts) == 2 && node.Kind == ast.KindMetaProperty {
		meta := node.AsMetaProperty()
		return meta != nil &&
			scanner.TokenToString(meta.KeywordToken) == parts[0] &&
			meta.Name() != nil &&
			meta.Name().Text() == parts[1]
	}

	if !ast.IsPropertyAccessExpression(node) {
		return false
	}

	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || ast.IsOptionalChainRoot(node) {
		return false
	}
	name := propertyAccess.Name()
	if name == nil || !ast.IsIdentifier(name) ||
		name.AsIdentifier().Text != parts[len(parts)-1] {
		return false
	}

	return nodeMatchesPathParts(
		ast.SkipParentheses(propertyAccess.Expression),
		parts[:len(parts)-1],
	)
}
