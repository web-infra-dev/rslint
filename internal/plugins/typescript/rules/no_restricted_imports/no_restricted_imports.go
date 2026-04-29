// Package no_restricted_imports provides the @typescript-eslint/no-restricted-imports
// rule, which extends the core ESLint no-restricted-imports rule with first-class
// support for TypeScript-only syntax: `import type ...`, inline `import { type X }`,
// `import x = require(...)`, `import type x = require(...)`, and `export type ...`.
//
// Compared with the core (ESLint base) rule, this wrapper applies two upstream-
// specific behaviors:
//
//  1. Source-level short-circuit on whole-type-only declarations:
//     when an import / export / import-equals is whole-type-only AND the source
//     matches any path/pattern with allowTypeImports=true, the declaration is
//     skipped before per-entry checks. This matches typescript-eslint's wrapper
//     and eliminates the conflict-case divergence (e.g. duplicate path entries
//     where one has allowTypeImports=true and another doesn't).
//
//  2. Synthesized default specifier for `import x = require(...)`:
//     upstream's wrapper rewrites import-equals into an ImportDeclaration with
//     a single ImportDefaultSpecifier so that `importNames: ['default']`,
//     `allowImportNames`, and `importNamePattern` apply to the local binding.
//     ESLint base treats import-equals as having no specifiers; we follow
//     upstream typescript-eslint here.
package no_restricted_imports

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_restricted_imports"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoRestrictedImportsRule = rule.CreateRule(rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		engine := core.NewEngine(options)
		if !engine.IsActive() {
			return rule.RuleListeners{}
		}
		isAllowedTypeSource := core.BuildAllowTypeImportSourceFilter(options)

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				d := node.AsImportDeclaration()
				if d.ModuleSpecifier == nil {
					return
				}
				source := strings.TrimSpace(utils.GetStaticStringValue(d.ModuleSpecifier))
				if source == "" {
					return
				}
				if isAllowedTypeSource != nil && core.IsTypeOnlyDeclaration(node) && isAllowedTypeSource(source) {
					return
				}
				engine.Check(&ctx, node, source, core.ExtractImportNames(d))
			},

			ast.KindExportDeclaration: func(node *ast.Node) {
				d := node.AsExportDeclaration()
				if d.ModuleSpecifier == nil {
					return
				}
				source := strings.TrimSpace(utils.GetStaticStringValue(d.ModuleSpecifier))
				if source == "" {
					return
				}
				// Upstream: ExportNamedDeclaration[source] gets the source-level
				// short-circuit; ExportAllDeclaration (incl. namespace export) is
				// passed straight to base, so its type-only handling falls through
				// to per-entry allowTypeImports.
				isNamedExport := d.ExportClause != nil && d.ExportClause.Kind == ast.KindNamedExports
				if isNamedExport && isAllowedTypeSource != nil && core.IsTypeOnlyDeclaration(node) && isAllowedTypeSource(source) {
					return
				}
				engine.Check(&ctx, node, source, core.ExtractExportNames(d))
			},

			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				ie := node.AsImportEqualsDeclaration()
				if ie.ModuleReference == nil || ie.ModuleReference.Kind != ast.KindExternalModuleReference {
					return
				}
				ext := ie.ModuleReference.AsExternalModuleReference()
				if ext.Expression == nil {
					return
				}
				source := strings.TrimSpace(utils.GetStaticStringValue(ext.Expression))
				if source == "" {
					return
				}
				if isAllowedTypeSource != nil && core.IsTypeOnlyDeclaration(node) && isAllowedTypeSource(source) {
					return
				}
				// Synthesize a default specifier so that importNames/allowImportNames/
				// importNamePattern checks apply to the local binding — upstream
				// typescript-eslint does this by rewriting import-equals into an
				// ImportDeclaration with a single ImportDefaultSpecifier.
				names := core.NewOrderedImportNames()
				if id := ie.Name(); id != nil {
					names.Add("default", core.NewSpecifierInfo(id, ie.IsTypeOnly))
				}
				engine.Check(&ctx, node, source, names)
			},
		}
	},
})
