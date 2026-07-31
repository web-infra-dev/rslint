// Package prefer_node_protocol ports eslint-plugin-unicorn's
// `prefer-node-protocol` rule.
//
// It flags a Node.js builtin-module specifier written without the `node:`
// protocol prefix — in `import`/`export` sources, dynamic `import()`,
// `require()`, `process.getBuiltinModule()`, and TypeScript `import(...)` type
// nodes — and autofixes it by inserting `node:` after the opening quote.
package prefer_node_protocol

import (
	"strings"

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

		checkModuleSpecifier := func(node *ast.Node) {
			check(node.ModuleSpecifier())
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
			if call.Expression.Kind == ast.KindImportKeyword {
				check(firstArgument(call))
				return
			}

			// `require("fs")` — isStaticRequire: bare `require` identifier callee,
			// exactly one string-literal argument, non-optional call.
			if isStaticRequire(node) {
				check(firstArgument(call))
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
				if object != nil && object.Kind == ast.KindIdentifier &&
					object.AsIdentifier().Text == "process" {
					check(firstArgument(call))
				}
			}
		}

		// TypeScript `import("fs")` type node. tsgo nests the specifier as a
		// LiteralType (`ImportTypeNode.Argument` → `LiteralTypeNode.Literal`),
		// unlike ESTree where the string Literal is a direct child of TSImportType.
		checkImportType := func(node *ast.Node) {
			argument := node.AsImportTypeNode().Argument
			if argument == nil || argument.Kind != ast.KindLiteralType {
				return
			}
			check(argument.AsLiteralTypeNode().Literal)
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration:   checkModuleSpecifier,
			ast.KindJSImportDeclaration: checkModuleSpecifier,
			ast.KindExportDeclaration:   checkModuleSpecifier,
			ast.KindCallExpression:      checkCall,
			ast.KindImportType:          checkImportType,
		}
	},
}

// shouldPrefix mirrors upstream's value gate: a string that lacks the `node:`
// prefix and whose bare name is a Node.js builtin module with a valid `node:`
// counterpart. builtinModuleNames already encodes "has a valid node: form", so
// a single lookup suffices — e.g. bare `test` is absent (only `node:test`
// exists) and is therefore left alone.
func shouldPrefix(value string) bool {
	return !strings.HasPrefix(value, nodeProtocol) && builtinModuleNames[value]
}

// isStaticRequire mirrors upstream's `isStaticRequire`: a non-optional call to a
// bare `require` identifier with exactly one string-literal argument.
func isStaticRequire(node *ast.Node) bool {
	if !ast.IsCallExpression(node) || ast.IsOptionalChainRoot(node) {
		return false
	}
	call := node.AsCallExpression()
	if call.Expression.Kind != ast.KindIdentifier ||
		call.Expression.AsIdentifier().Text != "require" {
		return false
	}
	args := call.Arguments
	if args == nil || len(args.Nodes) != 1 {
		return false
	}
	// Parentheses are transparent in ESTree, so `require(("fs"))` still counts
	// as a static string-literal argument.
	return ast.IsStringLiteral(ast.SkipParentheses(args.Nodes[0]))
}

// firstArgument returns the sole/first call argument, or nil when absent.
func firstArgument(call *ast.CallExpression) *ast.Node {
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	return call.Arguments.Nodes[0]
}
