// cspell:ignore sctx
package no_implied_eval

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var evalLikeFunctions = []string{"setTimeout", "setInterval", "execScript"}
var globalCandidates = []string{"window", "global", "globalThis", "self"}

// calleeOuterKinds only skips parentheses when inspecting the call's callee or
// receiver. ESLint's `isSpecificId` / `isSpecificMemberAccess` reject
// TSNonNullExpression / TSAsExpression / TSTypeAssertion / TSSatisfiesExpression
// (they are not `Identifier` / `MemberExpression`), so `setTimeout!('code')` and
// `(window as any).setTimeout('code')` are not flagged. Match that strictly.
const calleeOuterKinds = ast.OEKParentheses

// argOuterKinds skips parens + TS outer expressions when evaluating the first
// argument. ESLint's `getStaticValue` transparently unwraps TS assertions, so
// `setTimeout('code' as any)` still folds to the string "code" and is flagged.
const argOuterKinds = ast.OEKParentheses | ast.OEKAssertions

func buildImpliedEvalMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "impliedEval",
		Description: "Implied eval. Consider passing a function instead of a string.",
	}
}

func buildExecScriptMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "execScript",
		Description: "Implied eval. Do not use execScript().",
	}
}

// https://eslint.org/docs/latest/rules/no-implied-eval
var NoImpliedEvalRule = rule.Rule{
	Name: "no-implied-eval",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sctx := newStrCtx(ctx)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}

				callee := ast.SkipOuterExpressions(call.Expression, calleeOuterKinds)
				if callee == nil {
					return
				}

				var calleeName string

				switch callee.Kind {
				case ast.KindIdentifier:
					name := callee.AsIdentifier().Text
					if !slices.Contains(evalLikeFunctions, name) {
						return
					}
					if utils.IsShadowed(callee, name) {
						return
					}
					calleeName = name

				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
					name, ok := utils.AccessExpressionStaticName(callee)
					if !ok || !slices.Contains(evalLikeFunctions, name) {
						return
					}
					if !isGlobalCandidateChain(utils.AccessExpressionObject(callee)) {
						return
					}
					calleeName = name

				default:
					return
				}

				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				firstArg := call.Arguments.Nodes[0]

				if !sctx.isString(firstArg) {
					return
				}

				if calleeName == "execScript" {
					ctx.ReportNode(node, buildExecScriptMessage())
				} else {
					ctx.ReportNode(node, buildImpliedEvalMessage())
				}
			},
		}
	},
}

// isGlobalCandidateChain reports whether `node` is a reference to one of the
// known global objects (`window`, `global`, `globalThis`, `self`), possibly
// reached through a chain of **same-named** member accesses (e.g.
// `window.window`, `global['global']['global']`). Cross-candidate chains such
// as `window.global.setTimeout` or `self.window.setTimeout` are rejected, to
// match ESLint's per-candidate scope-manager walk. Shadowed root identifiers
// are also rejected.
func isGlobalCandidateChain(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, calleeOuterKinds)
	if node == nil {
		return false
	}

	// Descend through property / element access to find the root identifier.
	root := node
	for root.Kind == ast.KindPropertyAccessExpression || root.Kind == ast.KindElementAccessExpression {
		obj := utils.AccessExpressionObject(root)
		if obj == nil {
			return false
		}
		root = ast.SkipOuterExpressions(obj, calleeOuterKinds)
		if root == nil {
			return false
		}
	}
	if root.Kind != ast.KindIdentifier {
		return false
	}
	rootName := root.AsIdentifier().Text
	if !slices.Contains(globalCandidates, rootName) {
		return false
	}
	if utils.IsShadowed(root, rootName) {
		return false
	}

	// Ascend from the outermost access, verifying every property name equals rootName.
	current := node
	for current.Kind == ast.KindPropertyAccessExpression || current.Kind == ast.KindElementAccessExpression {
		name, ok := utils.AccessExpressionStaticName(current)
		if !ok || name != rootName {
			return false
		}
		obj := utils.AccessExpressionObject(current)
		if obj == nil {
			return false
		}
		current = ast.SkipOuterExpressions(obj, calleeOuterKinds)
		if current == nil {
			return false
		}
	}
	return true
}
