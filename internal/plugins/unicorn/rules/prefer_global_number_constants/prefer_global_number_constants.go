// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package prefer_global_number_constants

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

var PreferGlobalNumberConstantsRule = rule.Rule{
	Name: "unicorn/prefer-global-number-constants", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			reported := make(map[*ast.Node]bool)
			properties := make(map[string]*referencetracker.Trace, 3)
			for _, constant := range []struct{ property, replacement, global string }{{"NaN", "NaN", "NaN"}, {"POSITIVE_INFINITY", "Infinity", "Infinity"}, {"NEGATIVE_INFINITY", "-Infinity", "Infinity"}} {
				properties[constant.property] = &referencetracker.Trace{Read: func(node *ast.Node) {
					if reported[node] {
						return
					}
					reported[node] = true
					parent := utils.ESTreeParent(node)
					if (parent != nil && parent.Kind == ast.KindDeleteExpression) || utils.IsWriteReference(node) || !ctx.Refs.IsGlobalNameReference(node, constant.global, ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
						return
					}
					ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{Id: "prefer-global-number-constants", Description: "Prefer `" + constant.replacement + "` over `Number." + constant.property + "`.", Data: map[string]string{"replacement": constant.replacement, "property": constant.property}}, func() []rule.RuleFix {
						if !ast.IsAccessExpression(node) || constant.property == "NEGATIVE_INFINITY" || utils.HasCommentInsideNode(ctx.SourceFile, node) {
							return nil
						}
						return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, constant.replacement)}
					})
				}}
			}
			referencetracker.NewForReplacement(ctx).TrackGlobals(map[string]*referencetracker.Trace{"Number": {Properties: properties}})
		}}
	},
}
