package no_native

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

var NoNativeRule = rule.Rule{
	Name:   "promise/no-native",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Upstream accepts any authored definition in the global scope, even
		// one in a different TypeScript namespace. Module and CommonJS wrapper
		// bindings instead have to resolve the individual reference.
		sourceType := ctx.LanguageOptions.EffectiveSourceType()
		globalScope := sourceType == "script" ||
			(sourceType == "commonjs" && !ast.IsInJSFile(ctx.SourceFile.AsNode()))
		if globalScope && ctx.Refs.IsNameDefinedInFileWithMeaning(ctx.SourceFile.AsNode(), "Promise",
			ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
			return nil
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() != "Promise" || !scope.IsReferenceIdentifier(node) {
					return
				}
				meaning := scope.ESLintReferenceSpace(node).DeclarationMeaning()
				if ctx.Refs.ResolveInFileWithMeaning(node, meaning) != nil {
					return
				}
				// Configured globals, inline globals and TypeScript's libraries
				// do not supply the file-local definition required by this rule.
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "name",
					Description: `"Promise" is not defined.`,
				})
			},
		}
	},
}
