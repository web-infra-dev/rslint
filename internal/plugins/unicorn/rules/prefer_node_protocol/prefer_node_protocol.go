// Package prefer_node_protocol ports eslint-plugin-unicorn's
// `prefer-node-protocol` rule.
//
// It flags a Node.js builtin-module specifier written without the `node:`
// protocol prefix — in imports, named re-exports, dynamic `import()`,
// `require()`, `process.getBuiltinModule()`, and TypeScript `import(...)`
// type nodes — and autofixes it by inserting `node:` after the opening quote.
package prefer_node_protocol

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-node-protocol"

// nodeProtocol is the prefix upstream inserts and tests against.
const nodeProtocol = "node:"

func message(moduleName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Prefer `" + nodeProtocol + moduleName + "` over `" + moduleName + "`.",
		Data:        map[string]string{"moduleName": moduleName},
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-node-protocol.md
var PreferNodeProtocolRule = rule.Rule{
	Name:   "unicorn/prefer-node-protocol",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// check inspects a string-literal module specifier and reports + fixes
		// it when it names a Node.js builtin missing the `node:` prefix.
		//
		// tsgo divergence: ESTree flattens parentheses, so a parenthesized call
		// argument (`require(("fs"))`, `import(("fs"))`) presents its inner string
		// Literal directly to upstream's listener. tsgo keeps the
		// ParenthesizedExpression, so skip it to match the ESTree view.
		check := func(specifier *ast.Node) {
			if specifier == nil {
				return
			}
			specifier = ast.SkipParentheses(specifier)
			if specifier == nil || !ast.IsStringLiteral(specifier) {
				return
			}
			value := specifier.AsStringLiteral().Text
			if !shouldPrefix(value) {
				return
			}

			// Mirror upstream's `sourceCode.getRange(node)[0] + 1` (after the
			// opening quote). TrimNodeTextRange skips leading trivia so a
			// preceding comment (e.g. `import(/* c */"fs")`) doesn't shift the
			// insert position onto the comment.
			insertPosition := utils.TrimNodeTextRange(ctx.SourceFile, specifier).Pos() + 1
			ctx.ReportNodeWithDeferredFixes(specifier, message(value), func() []rule.RuleFix {
				return []rule.RuleFix{
					rule.RuleFixReplaceRange(core.NewTextRange(insertPosition, insertPosition), nodeProtocol),
				}
			})
		}

		checkExternalModuleSpecifier := func(node *ast.Node) {
			// tsgo parses JSDoc type imports again into ImportType nodes. Those nodes
			// live in comments and have no ESTree counterpart, so upstream's
			// Literal listener never observes them.
			if node.Flags&ast.NodeFlagsReparsed != 0 {
				return
			}

			// tsgo represents both ESTree ExportNamedDeclaration and
			// ExportAllDeclaration as ExportDeclaration. Upstream only listens to
			// module specifiers owned by ExportNamedDeclaration.
			if ast.IsExportDeclaration(node) {
				exportClause := node.AsExportDeclaration().ExportClause
				if exportClause == nil || !ast.IsNamedExports(exportClause) {
					return
				}
			}

			check(ast.GetExternalModuleName(node))
		}

		// tsgo divergence: ESTree fires the rule's single `Literal` listener and
		// branches on `node.parent.type` (ImportDeclaration / ImportExpression /
		// require call / getBuiltinModule call / TSImportType). tsgo has no
		// `Literal` kind and represents each of these positions with a distinct
		// node kind, so the ESTree branch table becomes one listener per kind.
		// The CallExpression listener folds together dynamic `import()`,
		// `require()`, and `process.getBuiltinModule()`.
		checkCall := func(node *ast.Node) {
			call := node.AsCallExpression()

			// Dynamic `import("fs")` — tsgo models the ESTree ImportExpression as
			// a CallExpression whose callee is the `import` keyword.
			if ast.IsImportCall(node) {
				check(ast.GetExternalModuleName(node))
				return
			}

			callee := ast.SkipParentheses(call.Expression)
			if callee == nil {
				return
			}

			// `require("fs")` — isStaticRequire: bare `require` identifier callee,
			// exactly one string-literal argument, non-optional call. Gate on the
			// identifier name first so ordinary function calls avoid the complete
			// require-shape check.
			if ast.IsIdentifier(callee) {
				if callee.AsIdentifier().Text == "require" && isStaticRequire(node) {
					check(ast.GetExternalModuleName(node))
				}
				return
			}

			// Only a dot-property callee can match `process.getBuiltinModule`.
			// Avoid the full shared method-call matcher for element-access and
			// other call shapes.
			if !ast.IsPropertyAccessExpression(callee) {
				return
			}

			// `process.getBuiltinModule("fs")` — non-optional dot call on the
			// `process` identifier, exactly one argument. Parentheses around the
			// object are transparent in ESTree, so skip them before the name check.
			argumentsLength := 1
			if match, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Method:          "getBuiltinModule",
				ArgumentsLength: &argumentsLength,
			}); ok {
				object := ast.SkipParentheses(match.Object)
				if object != nil && ast.IsIdentifier(object) &&
					object.AsIdentifier().Text == "process" {
					check(ast.GetExternalModuleName(node))
				}
			}
		}

		// Do not register KindJSImportDeclaration: tsgo reserves it for the
		// synthetic declarations reparsed from JSDoc `@import` tags. Real import
		// declarations in JavaScript files still use KindImportDeclaration.
		return rule.RuleListeners{
			ast.KindImportDeclaration: checkExternalModuleSpecifier,
			ast.KindExportDeclaration: checkExternalModuleSpecifier,
			ast.KindCallExpression:    checkCall,
			ast.KindImportType:        checkExternalModuleSpecifier,
		}
	},
}

// shouldPrefix mirrors upstream's value gate: a string that lacks the `node:`
// prefix and whose bare name is a Node.js builtin module with a valid `node:`
// counterpart. builtinModuleNames already encodes "has a valid node: form", so
// a single lookup suffices — e.g. bare `test` is absent (only `node:test`
// exists) and is therefore left alone.
func shouldPrefix(value string) bool {
	return builtinModuleNames[value]
}

// isStaticRequire mirrors upstream's `isStaticRequire`: a non-optional call to a
// bare `require` identifier with exactly one string-literal argument.
func isStaticRequire(node *ast.Node) bool {
	// ast.IsRequireCall is deliberately not used here: it neither unwraps a
	// parenthesized callee nor rejects optional calls, and its string-literal-like
	// mode also accepts template literals. All three differ from upstream.
	if !ast.IsCallExpression(node) || ast.IsOptionalChainRoot(node) {
		return false
	}
	call := node.AsCallExpression()
	// Parentheses are transparent in ESTree, so `(require)("fs")` presents a
	// bare `require` identifier callee to upstream. tsgo keeps the
	// ParenthesizedExpression, so skip it before the name check.
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || !ast.IsIdentifier(callee) ||
		callee.AsIdentifier().Text != "require" {
		return false
	}
	args := node.Arguments()
	if len(args) != 1 {
		return false
	}
	// Parentheses are transparent in ESTree, so `require(("fs"))` still counts
	// as a static string-literal argument.
	return ast.IsStringLiteral(ast.SkipParentheses(args[0]))
}
