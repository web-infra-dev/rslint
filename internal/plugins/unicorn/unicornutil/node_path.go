package unicornutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// NodeMatchesPath mirrors unicorn's isNodeMatchesNameOrPath helper. It accepts
// dotted, non-computed, non-optional paths rooted at an identifier, this,
// super, or a meta property. Parentheses and JavaScript JSDoc casts are
// transparent at every segment; authored TypeScript assertions remain
// visible. ast.IsDottedName is intentionally not used because it also accepts
// property accesses containing optional-chain links.
func NodeMatchesPath(node *ast.Node, path string) bool {
	path = ecmascript.StringTrim(path)
	// Member expressions expose the last segment first. Match it before
	// visiting the receiver, without allocating a slice for every candidate.
	for node = utils.ESTreeRuntimeExpression(node); node != nil; node = utils.ESTreeRuntimeExpression(node.Expression()) {
		switch node.Kind {
		case ast.KindIdentifier:
			return node.AsIdentifier().Text == path && !strings.Contains(path, ".")
		case ast.KindThisKeyword:
			return path == "this"
		case ast.KindSuperKeyword:
			return path == "super"
		case ast.KindMetaProperty:
			root, property, dotted := strings.Cut(path, ".")
			meta := node.AsMetaProperty()
			return dotted && meta != nil && meta.Name() != nil && meta.Name().Text() == property &&
				scanner.TokenToString(meta.KeywordToken) == root
		case ast.KindPropertyAccessExpression:
			dot := strings.LastIndexByte(path, '.')
			name := node.Name()
			if dot < 0 || ast.IsOptionalChainRoot(node) || name == nil ||
				!ast.IsIdentifier(name) || name.Text() != path[dot+1:] {
				return false
			}
			path = path[:dot]
		default:
			return false
		}
	}
	return false
}
