// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package prefer_identifier_import_export_specifiers

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferIdentifierImportExportSpecifiersRule = rule.Rule{
	Name: "unicorn/prefer-identifier-import-export-specifiers", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindStringLiteral: func(node *ast.Node) {
			if !isModuleName(node) || !scanner.IsIdentifierText(node.Text(), core.LanguageVariantStandard) {
				return
			}
			identifier := node.Text()
			literal := utils.TrimmedNodeText(ctx.SourceFile, node)
			ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{Id: "prefer-identifier-import-export-specifiers", Description: "Prefer identifier `" + identifier + "` over string literal `" + literal + "`.", Data: map[string]string{"identifier": identifier, "literal": literal}}, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, utils.SafeReplacementText(ctx.SourceFile, node, identifier))}
			})
		}}
	},
}

// Visiting each literal once avoids duplicate reports for unaliased exports,
// while covering both names in a re-export.
func isModuleName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindImportSpecifier:
		return parent.AsImportSpecifier().PropertyName == node
	case ast.KindExportSpecifier:
		if parent.Name() == node {
			return true
		}
		declaration := parent.Parent.Parent
		return declaration.Kind == ast.KindExportDeclaration && declaration.AsExportDeclaration().ModuleSpecifier != nil && parent.AsExportSpecifier().PropertyName == node
	case ast.KindNamespaceExport, ast.KindImportAttribute:
		return parent.Name() == node
	}
	return false
}
