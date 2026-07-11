package no_unsafe_function_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildBannedFunctionTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedFunctionType",
		Description: "The `Function` type accepts any function-like value.\nPrefer explicitly defining any function parameters and return type.",
	}
}

var NoUnsafeFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-function-type",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkBannedType := func(ident *ast.Node) {
			if ident == nil || ident.Kind != ast.KindIdentifier {
				return
			}
			if ident.AsIdentifier().Text != "Function" {
				return
			}
			if !typescriptutil.IsReferenceToGlobalIdentifier(ctx, ident) {
				return
			}
			ctx.ReportNode(ident, buildBannedFunctionTypeMessage())
		}

		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				ref := node.AsTypeReferenceNode()
				if ref == nil {
					return
				}
				checkBannedType(ref.TypeName)
			},
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				if !typescriptutil.IsClassImplementsOrInterfaceExtends(node) {
					return
				}
				expr := node.AsExpressionWithTypeArguments()
				if expr == nil {
					return
				}
				checkBannedType(expr.Expression)
			},
		}
	},
})
