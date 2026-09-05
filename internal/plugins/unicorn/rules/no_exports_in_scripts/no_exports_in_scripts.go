// Package no_exports_in_scripts ports eslint-plugin-unicorn's
// `no-exports-in-scripts` rule.
//
// A file whose first line starts with a `#!` shebang is treated as a script
// meant to be executed directly, not as a module. Any `export` declaration
// inside such a file mixes module and script boundaries and is reported.
package no_exports_in_scripts

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-exports-in-scripts"

func message() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Do not use exports in scripts.",
	}
}

// isExported reports whether the node is preceded by an `export` keyword in
// the source. The check is a syntactic modifier lookup on the node itself, so
// it correctly handles `export const x = 1;`, `export function foo() {}`,
// `export class Foo {}`, `export interface Foo {}`, `export type ...`,
// `export enum ...`, `export namespace ...`, and TypeScript
// `export import ...` declarations. `KindExportDeclaration` is already an
// export-only node and does not need this check.
func exportedDeclarationRange(node *ast.Node, sourceFile *ast.SourceFile) (core.TextRange, bool) {
	if node == nil || !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return core.TextRange{}, false
	}
	modifiers := node.Modifiers()
	if modifiers == nil {
		return core.TextRange{}, false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			return core.NewTextRange(utils.TrimNodeTextRange(sourceFile, modifier).Pos(), node.End()), true
		}
	}
	return core.TextRange{}, false
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/rules/no-exports-in-scripts.js
var NoExportsInScriptsRule = rule.Rule{
	Name:   "unicorn/no-exports-in-scripts",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Mirror upstream's early-out: only run when the first line of the
		// file starts with `#!`. `scanner.GetShebang` returns the shebang
		// substring (consumed up to and including the newline) or empty when
		// no shebang is present; this matches ESLint's `sourceCode.lines[0]
		// .startsWith('#!')` exactly — a shebang in a comment or string
		// never matches because the file does not literally start with `#!`.
		text := ctx.SourceFile.Text()
		if scanner.GetShebang(text) == "" {
			return rule.RuleListeners{}
		}

		reportIfExported := func(node *ast.Node) {
			if reportRange, ok := exportedDeclarationRange(node, ctx.SourceFile); ok {
				ctx.ReportRange(reportRange, message())
			}
		}

		return rule.RuleListeners{
			// ---- tsgo divergence: in ESTree, `export const foo = 1;`,
			// `export function foo() {}`, `export class Foo {}`,
			// `export interface Foo {}`, `export type ...`, `export enum
			// Foo {}`, and `export namespace ...` all share the
			// ExportNamedDeclaration node type. tsgo collapses the
			// `export` keyword into a syntactic modifier on the
			// underlying declaration, so each surface form is a distinct
			// node kind with `ModifierFlagsExport` set. A single helper
			// checks the modifier for all of them.
			ast.KindVariableStatement:       reportIfExported,
			ast.KindFunctionDeclaration:     reportIfExported,
			ast.KindClassDeclaration:        reportIfExported,
			ast.KindInterfaceDeclaration:    reportIfExported,
			ast.KindTypeAliasDeclaration:    reportIfExported,
			ast.KindEnumDeclaration:         reportIfExported,
			ast.KindModuleDeclaration:       reportIfExported,
			ast.KindImportEqualsDeclaration: reportIfExported,

			// ---- tsgo divergence: ExportDeclaration keeps ESTree's
			// shape for `export {...}`, `export * from '...'`,
			// `export {foo} from '...'`, `export * as ns from '...'`,
			// and the empty `export {};`. Each of these fires the
			// listener and is reported directly.
			ast.KindExportDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, message())
			},

			// ---- tsgo divergence: ExportAssignment covers both `export
			// default ...` (ESTree ExportDefaultDeclaration) and `export =
			// ...` (ESTree TSExportAssignment). Upstream reports only the
			// former.
			ast.KindExportAssignment: func(node *ast.Node) {
				if !node.AsExportAssignment().IsExportEquals {
					ctx.ReportNode(node, message())
				}
			},
		}
	},
}
