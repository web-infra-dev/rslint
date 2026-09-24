package no_named_default

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-default.js
var NoNamedDefaultRule = rule.Rule{
	Name:   "import/no-named-default",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause == nil {
					return
				}
				clause := declaration.ImportClause.AsImportClause()
				if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
					return
				}
				for _, node := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					specifier := node.AsImportSpecifier()
					// Upstream exempts inline type specifiers, not import type declarations.
					if specifier.IsTypeOnly || node.PropertyNameOrName().Text() != "default" {
						continue
					}
					name := specifier.Name()
					// The local binding is an identifier; skip its trivia without rescanning the token.
					start := scanner.GetTokenPosOfNode(name, ctx.SourceFile, false)
					ctx.ReportRange(core.NewTextRange(start, name.End()), rule.RuleMessage{
						Description: "Use default import syntax to import '" + name.Text() + "'.",
					})
				}
			},
		}
	},
}
