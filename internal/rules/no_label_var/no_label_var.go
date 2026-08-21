package no_label_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-label-var
//
// ESLint's `getVariableByName` walks the scope chain all the way to the global
// scope. The same-file binding checks and framework globals view cover those
// layers without allowing the active TypeScript program to change the result:
//
//  1. utils.IsShadowed — fast, works without type info; covers every binding
//     declared inside the current source file (var/let/const, function, class,
//     enum, namespace, import, parameter, catch, for-init, function-expression
//     name, hoisted vars).
//  2. ctx.Refs — resolves binder-owned names in every declaration space at the
//     label's location, plus value bindings supplied by a file wrapper.
//  3. The shared default TypeScript type-global set mirrors scope-manager for
//     TypeScript files without consulting tsconfig libraries.
//  4. ctx.Globals — catches the selected ECMAScript edition, resolved language
//     globals, config and inline globals, including explicit `off` overrides.
var NoLabelVarRule = rule.Rule{
	Name:   "no-label-var",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "identifierClashWithLabel",
				Description: "Found identifier with same name as label.",
			})
		}

		return rule.RuleListeners{
			ast.KindLabeledStatement: func(node *ast.Node) {
				ls := node.AsLabeledStatement()
				if ls == nil || ls.Label == nil {
					return
				}
				name := ls.Label.Text()

				if utils.IsShadowed(node, name) {
					report(node)
					return
				}
				// scope-manager leaves class type parameters in the lexical scope
				// chain of static members, while TypeScript's resolver deliberately
				// hides them there. Preserve the scope-manager answer.
				if utils.HasEnclosingTypeParameter(node, name) {
					report(node)
					return
				}
				if ctx.Refs != nil && ctx.Refs.IsNameDefinedInFileWithMeaning(
					node,
					name,
					ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
				) {
					report(node)
					return
				}

				if !ast.IsInJSFile(node) && rule.IsDefaultTypeScriptTypeGlobal(name) {
					report(node)
					return
				}

				if ctx.Globals.Access(name).IsDeclared() {
					report(node)
				}
			},
		}
	},
}
