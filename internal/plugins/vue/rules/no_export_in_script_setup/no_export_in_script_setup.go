// Package no_export_in_script_setup ports eslint-plugin-vue's
// `no-export-in-script-setup` rule.
package no_export_in_script_setup

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "forbidden"

func forbiddenMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "`<script setup>` cannot contain ES module exports.",
	}
}

// NoExportInScriptSetupRule disallows an ES module export inside
// `<script setup>`.
//
// A `<script setup>` block is compiled into a component's setup function, so an
// export in it has nothing to attach to and the Vue compiler rejects it. A
// plain `<script>` block beside it is an ordinary module whose exports are
// fine. That is the whole difficulty here: both blocks are projected into one
// parsed text, so the rule has to ask which block a node came from.
//
// # Shape of an export in this AST
//
// Upstream reads ESTree, where every export is an Export*Declaration node.
// TypeScript's AST splits them: `export {}` and `export * from` are export
// declarations, `export default` and `export =` are export assignments, and
// `export const` / `export function` / `export class` are ordinary
// declarations carrying an export modifier. All three shapes are reported.
//
// A type-only export is skipped, as upstream skips `exportKind === 'type'`:
// nothing of it survives compilation, so the Vue compiler never sees it. This
// covers `export type X`, `export type { A }` and `export { type A }`. An
// exported `interface` is skipped for the same reason.
//
// https://github.com/vuejs/eslint-plugin-vue/blob/master/lib/rules/no-export-in-script-setup.js
var NoExportInScriptSetupRule = rule.Rule{
	Name:   "vue/no-export-in-script-setup",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Only a component declaring a `<script setup>` block can violate
		// this, so every other file registers no listeners at all.
		if _, ok := ctx.Component.ScriptSetupRange(); !ok {
			return nil
		}

		report := func(node *ast.Node) {
			// A node's Pos() includes its leading trivia, which for the first
			// statement of a block starts before the block does. Ask where the
			// node's first token is instead, which is also the position the
			// diagnostic is anchored to.
			if !ctx.Component.InScriptSetup(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()) {
				return
			}
			ctx.ReportNode(node, forbiddenMessage())
		}

		reportModifierExport := func(node *ast.Node) {
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
				return
			}
			report(node)
		}

		listeners := rule.RuleListeners{
			// `export default value` and `export = value`.
			ast.KindExportAssignment: report,
			// `export { a }`, `export { a } from "m"`, `export * from "m"`.
			ast.KindExportDeclaration: func(node *ast.Node) {
				if isTypeOnlyExportDeclaration(node) {
					return
				}
				report(node)
			},
		}

		// A declaration carrying an export modifier. An interface and a type
		// alias are absent from this list on purpose: they emit nothing.
		for _, kind := range []ast.Kind{
			ast.KindVariableStatement,
			ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindEnumDeclaration,
			ast.KindModuleDeclaration,
		} {
			listeners[kind] = reportModifierExport
		}
		return listeners
	},
}

// isTypeOnlyExportDeclaration reports whether an export declaration carries no
// runtime value: either the whole clause is `export type`, or every specifier
// in it is individually type-only.
func isTypeOnlyExportDeclaration(node *ast.Node) bool {
	declaration := node.AsExportDeclaration()
	if declaration == nil {
		return false
	}
	if declaration.IsTypeOnly {
		return true
	}
	if declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
		return false
	}
	named := declaration.ExportClause.AsNamedExports()
	if named == nil || named.Elements == nil || len(named.Elements.Nodes) == 0 {
		return false
	}
	for _, element := range named.Elements.Nodes {
		specifier := element.AsExportSpecifier()
		if specifier == nil || !specifier.IsTypeOnly {
			return false
		}
	}
	return true
}
