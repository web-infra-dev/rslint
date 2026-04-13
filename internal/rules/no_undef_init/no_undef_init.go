package no_undef_init

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-undef-init
var NoUndefInitRule = rule.Rule{
	Name: "no-undef-init",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}

				// Check initializer is the identifier `undefined`.
				// Skip parentheses to match ESLint behavior (ESTree strips them).
				init := varDecl.Initializer
				if init == nil {
					return
				}
				unwrapped := ast.SkipParentheses(init)
				if unwrapped.Kind != ast.KindIdentifier || unwrapped.Text() != "undefined" {
					return
				}

				// Must be inside a VariableDeclarationList (not a CatchClause)
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindVariableDeclarationList {
					return
				}

				// Skip const, using, await using — only report var and let
				if ast.IsVarConst(parent) || ast.IsVarUsing(parent) || ast.IsVarAwaitUsing(parent) {
					return
				}

				// Skip if `undefined` is shadowed by a local declaration
				if utils.IsShadowed(unwrapped, "undefined") {
					return
				}

				nameNode := varDecl.Name()
				if nameNode == nil {
					return
				}
				nameText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, nameNode, false)

				msg := rule.RuleMessage{
					Id:          "unnecessaryUndefinedInit",
					Description: "It's not necessary to initialize '" + nameText + "' to undefined.",
				}

				// Only provide autofix for `let` (not `var` — hoisting semantics differ)
				if !ast.IsVarLet(parent) {
					ctx.ReportNode(node, msg)
					return
				}

				// No autofix for destructuring patterns
				if ast.IsBindingPattern(nameNode) {
					ctx.ReportNode(node, msg)
					return
				}

				// Compute fix range: remove " = undefined" while preserving type annotations
				// and the definite assignment token (!).
				// NOTE: Unlike ESLint, which removes from name-end to declarator-end
				// (losing type annotations in TS), we start the removal from the end of the
				// last node before the initializer to preserve TypeScript syntax.
				// VariableDeclaration children in order: Name, ExclamationToken?, Type?, Initializer
				var fixStart int
				if varDecl.Type != nil {
					fixStart = varDecl.Type.End()
				} else if varDecl.ExclamationToken != nil {
					fixStart = varDecl.ExclamationToken.End()
				} else {
					fixStart = nameNode.End()
				}
				nodeEnd := node.End()

				// No autofix if comments exist in the removal range.
				// Safe to use string matching since the span only contains
				// whitespace, `=`, `undefined`, and potential comments — no string literals.
				between := ctx.SourceFile.Text()[fixStart:nodeEnd]
				if strings.Contains(between, "/*") || strings.Contains(between, "//") {
					ctx.ReportNode(node, msg)
					return
				}

				// Autofix: remove " = undefined"
				fix := rule.RuleFixRemoveRange(core.NewTextRange(fixStart, nodeEnd))
				ctx.ReportNodeWithFixes(node, msg, fix)
			},
		}
	},
}
