// Package no_exports_in_scripts ports eslint-plugin-unicorn's
// `no-exports-in-scripts` rule.
//
// A file whose first line starts with a `#!` shebang is treated as a script
// meant to be executed directly, not as a module. Any `export` declaration
// inside such a file mixes module and script boundaries and is reported.
package no_exports_in_scripts

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
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
// `export enum ...`, and `export namespace ...`. `KindExportDeclaration`
// and `KindExportAssignment` are *already* export-only nodes and do not
// need this check.
func isExported(node *ast.Node) bool {
	return node != nil && ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/no-exports-in-scripts.js
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
		//
		// tsgo divergence: tsgo's `SourceFile.Text()` strips a leading BOM
		// (the BOM is exposed via `ctx.HasBOM()`), so a file written as
		// `<BOM>#!...` looks identical to a plain `#!...` file to this
		// check. Upstream ESLint does not strip the BOM, so
		// `lines[0].startsWith('#!')` is false for the BOM-prefixed form
		// and the upstream rule does not fire. We follow the rslint
		// convention (treat the shebang as the file's first content) and
		// fire here; a future lock-in could special-case the BOM if
		// matching upstream becomes important.
		text := ctx.SourceFile.Text()
		if scanner.GetShebang(text) == "" {
			return rule.RuleListeners{}
		}

		reportIfExported := func(node *ast.Node) {
			if isExported(node) {
				ctx.ReportNode(node, message())
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
			ast.KindVariableStatement:    reportIfExported,
			ast.KindFunctionDeclaration:  reportIfExported,
			ast.KindClassDeclaration:     reportIfExported,
			ast.KindInterfaceDeclaration: reportIfExported,
			ast.KindTypeAliasDeclaration: reportIfExported,
			ast.KindEnumDeclaration:      reportIfExported,
			ast.KindModuleDeclaration:    reportIfExported,

			// ---- tsgo divergence: ExportDeclaration keeps ESTree's
			// shape for `export {...}`, `export * from '...'`,
			// `export {foo} from '...'`, `export * as ns from '...'`,
			// and the empty `export {};`. Each of these fires the
			// listener and is reported directly.
			ast.KindExportDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, message())
			},

			// ---- tsgo divergence: ExportAssignment covers `export
			// default ...` (ESTree ExportDefaultDeclaration) and
			// `export = ...` (TypeScript's legacy CommonJS-interop form).
			ast.KindExportAssignment: func(node *ast.Node) {
				ctx.ReportNode(node, message())
			},
		}
	},
}
