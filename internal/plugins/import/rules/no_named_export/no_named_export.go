package no_named_export

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-export.js
// Babel default re-export proposals are outside the parser's supported syntax.
var NoNamedExportRule = rule.Rule{
	Name:   "import/no-named-export",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		if ctx.LanguageOptions.EffectiveSourceType() != "module" {
			return nil
		}

		// Upstream uses a literal message without a message ID.
		message := rule.RuleMessage{Description: "Named exports are not allowed."}
		checkDeclaration := func(node *ast.Node) {
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) {
				return
			}
			// tsgo stores declaration exports as modifiers. Starting at export
			// excludes any preceding decorators, matching the ESTree wrapper.
			for _, modifier := range node.ModifierNodes() {
				// Dotted namespaces have synthetic, zero-width export modifiers.
				if modifier.Kind != ast.KindExportKeyword || ast.NodeIsMissing(modifier) {
					continue
				}
				start := scanner.GetTokenPosOfNode(modifier, ctx.SourceFile, false)
				ctx.ReportRange(core.NewTextRange(start, node.End()), message)
				return
			}
		}

		return rule.RuleListeners{
			ast.KindExportDeclaration: func(node *ast.Node) {
				clause := node.AsExportDeclaration().ExportClause
				if clause != nil && clause.Kind == ast.KindNamedExports {
					specifiers := clause.AsNamedExports().Elements.Nodes
					onlyDefault := len(specifiers) > 0
					for _, specifier := range specifiers {
						if !ast.ModuleExportNameIsDefault(specifier.Name()) {
							onlyDefault = false
							break
						}
					}
					if onlyDefault {
						return
					}
				}
				// Empty exports and all star exports (even * as default) report.
				start := scanner.GetTokenPosOfNode(node, ctx.SourceFile, false)
				ctx.ReportRange(core.NewTextRange(start, node.End()), message)
			},
			ast.KindVariableStatement:       checkDeclaration,
			ast.KindFunctionDeclaration:     checkDeclaration,
			ast.KindClassDeclaration:        checkDeclaration,
			ast.KindInterfaceDeclaration:    checkDeclaration,
			ast.KindTypeAliasDeclaration:    checkDeclaration,
			ast.KindEnumDeclaration:         checkDeclaration,
			ast.KindModuleDeclaration:       checkDeclaration,
			ast.KindImportEqualsDeclaration: checkDeclaration,
		}
	},
}
