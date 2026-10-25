package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

type memberReceiverNamesKey struct{}

// MemberReceiverNames returns a read-only superset of identifier receivers used
// by the namespace and default-member listeners. Shadowed names and unrelated
// initializers stay in the set; absence alone allows skipping module queries.
func MemberReceiverNames(ctx rule.RuleContext) map[string]bool {
	return rule.CachedByFile(ctx, memberReceiverNamesKey{}, func() map[string]bool {
		names := make(map[string]bool)
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			var object *ast.Node
			switch node.Kind {
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindQualifiedName:
				object, _ = rslint_utils.MemberExpressionParts(node)
			case ast.KindVariableDeclaration:
				object = node.AsVariableDeclaration().Initializer
			}
			if object != nil {
				for _, candidate := range []*ast.Node{ast.SkipParentheses(object), rslint_utils.ESTreeRuntimeExpression(object)} {
					if candidate != nil && candidate.Kind == ast.KindIdentifier {
						names[candidate.Text()] = true
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(ctx.SourceFile.AsNode())
		return names
	})
}
