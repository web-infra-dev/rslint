package no_wrapper_object_types

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// classNames mirrors the upstream `classNames` Set verbatim. Order is
// alphabetical so a `grep` of "Object" or "Symbol" lands on the canonical
// entry rather than a comment.
var classNames = map[string]struct{}{
	"BigInt":  {},
	"Boolean": {},
	"Number":  {},
	"Object":  {},
	"String":  {},
	"Symbol":  {},
}

func buildBannedClassTypeMessage(typeName, preferred string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedClassType",
		Description: "Prefer using the primitive `" + preferred + "` as a type name, rather than the upper-cased `" + typeName + "`.",
		Data: map[string]string{
			"preferred": preferred,
			"typeName":  typeName,
		},
	}
}

var NoWrapperObjectTypesRule = rule.CreateRule(rule.Rule{
	Name:             "no-wrapper-object-types",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// checkBannedTypes mirrors upstream's local `checkBannedTypes` —
		// `node` is the inner Identifier the rule visits (TypeReference's
		// TypeName, or ExpressionWithTypeArguments's Expression). `includeFix`
		// is true only for TSTypeReference, matching upstream's listener wiring.
		checkBannedTypes := func(node *ast.Node, includeFix bool) {
			if node == nil || node.Kind != ast.KindIdentifier {
				return
			}
			typeName := node.AsIdentifier().Text
			if _, banned := classNames[typeName]; !banned {
				return
			}
			if !typescriptutil.IsReferenceToGlobalIdentifier(ctx, node) {
				return
			}
			preferred := strings.ToLower(typeName)
			msg := buildBannedClassTypeMessage(typeName, preferred)
			if includeFix {
				ctx.ReportNodeWithFixes(node, msg, rule.RuleFixReplace(ctx.SourceFile, node, preferred))
				return
			}
			ctx.ReportNode(node, msg)
		}

		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				ref := node.AsTypeReferenceNode()
				if ref == nil {
					return
				}
				checkBannedTypes(ref.TypeName, true)
			},
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				if !typescriptutil.IsClassImplementsOrInterfaceExtends(node) {
					return
				}
				expr := node.AsExpressionWithTypeArguments()
				if expr == nil {
					return
				}
				checkBannedTypes(expr.Expression, false)
			},
		}
	},
})
