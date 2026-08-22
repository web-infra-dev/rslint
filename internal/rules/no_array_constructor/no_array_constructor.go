package no_array_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func preferLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

func useLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "Replace with an array literal.",
	}
}

func useLiteralAfterSemicolonMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteralAfterSemicolon",
		Description: "Replace with an array literal, add preceding semicolon.",
	}
}

// calleeArgsAndTypeArgs extracts the callee, argument list, and type-argument
// list shared by CallExpression and NewExpression, the two forms this rule
// inspects.
func calleeArgsAndTypeArgs(node *ast.Node) (callee *ast.Node, args *ast.NodeList, typeArgs *ast.NodeList) {
	switch node.Kind {
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		return callExpr.Expression, callExpr.Arguments, callExpr.TypeArguments
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		return newExpr.Expression, newExpr.Arguments, newExpr.TypeArguments
	}
	return nil, nil, nil
}

// findOpenParen scans forward from calleeEnd (the position right after the
// possibly-parenthesized callee) for the call's own opening paren, stopping
// at nodeEnd. Only `?.` punctuation and trivia can appear before it, since
// type arguments were already excluded by the caller. found is false when the
// node has no parentheses at all, e.g. a bare `new Array`.
func findOpenParen(sourceFile *ast.SourceFile, calleeEnd int, nodeEnd int) (pos int, found bool) {
	text := sourceFile.Text()
	p := calleeEnd
	for p < nodeEnd {
		p = scanner.SkipTrivia(text, p)
		if p >= nodeEnd {
			break
		}
		if text[p] == '(' {
			return p, true
		}
		p++
	}
	return 0, false
}

func countNonSpread(args *ast.NodeList) int {
	if args == nil {
		return 0
	}
	count := 0
	for _, arg := range args.Nodes {
		if arg.Kind != ast.KindSpreadElement {
			count++
		}
	}
	return count
}

// https://eslint.org/docs/latest/rules/no-array-constructor
var NoArrayConstructorRule = rule.Rule{
	Name:   "no-array-constructor",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			callee, args, typeArgs := calleeArgsAndTypeArgs(node)
			if callee == nil {
				return
			}
			// ESTree flattens grouping parentheses around the callee, so
			// `(Array)()` still has an Identifier callee; TS-only wrappers
			// (non-null assertion, `as`/`satisfies`) are deliberately left
			// alone since ESTree never collapses those back to a bare
			// Identifier either.
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "Array" {
				return
			}
			// `Array<Foo>()` — explicit type arguments mean this isn't the
			// bare-constructor call the rule targets.
			if typeArgs != nil && len(typeArgs.Nodes) > 0 {
				return
			}
			// A single non-spread argument (`Array(9)`, `Array(x)`) may be
			// intentionally creating a sparse array of that length — leave
			// it alone, since the argument's runtime type can't be known
			// statically.
			if args != nil && len(args.Nodes) == 1 && args.Nodes[0].Kind != ast.KindSpreadElement {
				return
			}

			// A local declaration shadows the global constructor.
			if utils.IsShadowed(callee, "Array") {
				return
			}
			// scope-manager keeps class type parameters visible inside static
			// members, where TypeScript's resolver hides them, and keeps a
			// function's parameter initializers in the same scope as its body
			// declarations, where both TypeScript's resolver and IsShadowed
			// follow runtime lexical semantics instead.
			if utils.HasEnclosingTypeParameter(callee, "Array") ||
				utils.IsShadowedFromParameterInitializer(callee, "Array") {
				return
			}
			// scope-manager also creates an `Array` variable for TypeScript
			// type-space declarations — type aliases, interfaces, type
			// parameters, type-only imports — which utils.IsShadowed doesn't
			// model. Resolving the name in every declaration space at the call
			// site covers them.
			if ctx.Refs != nil && ctx.Refs.IsNameDefinedInFileWithMeaning(
				callee,
				"Array",
				ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
			) {
				return
			}
			// A config `/* global Array: off */` / `languageOptions.globals`
			// entry un-declares the builtin, so `Array` no longer resolves to
			// a known global — ESLint's `getVariableByName` would return
			// undefined and the rule stays silent.
			if !ctx.Globals.Access("Array").IsDeclared() {
				return
			}

			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			openParen, hasParens := findOpenParen(ctx.SourceFile, callee.End(), nodeRange.End())
			argsText := ""
			commentBound := nodeRange.End()
			if hasParens {
				argsText = ctx.SourceFile.Text()[openParen+1 : nodeRange.End()-1]
				commentBound = openParen
			}
			fixText := "[" + argsText + "]"

			// NOTE: Unlike ESLint, utils.NeedsPrecedingSemicolon doesn't model
			// TypeScript-only node kinds (type positions, ambient/overload
			// function declarations, import-equals declarations), so it falls
			// back to the conservative "needs a semicolon" answer there —
			// e.g. after `type T = Foo` or `import Foo = Bar`. The extra `;`
			// never changes behavior, only adds a redundant character; see
			// the rule doc's "Differences from ESLint" section.
			addSemicolon := utils.IsStartOfExpressionStatement(ctx.SourceFile, node) &&
				utils.NeedsPrecedingSemicolon(ctx.SourceFile, node)
			if addSemicolon {
				fixText = ";" + fixText
			}

			// Replacing the call with an array literal directly (rather than
			// as an opt-in suggestion) is only safe when doing so can't
			// change runtime behavior or silently drop source text:
			//   - `Array?.(...)` short-circuits to `undefined` when `Array`
			//     is nullish; `[...]` always evaluates.
			//   - A single spread plus at most one more argument
			//     (`Array(5, ...args)`) is equivalent to the sparse-array
			//     single-argument form when the spread happens to be empty
			//     at runtime — an outcome an autofix must not risk.
			//   - A comment between the node's start and the call's own
			//     opening paren (`Array/*a*/()`) would be silently discarded
			//     by the fix. Comments inside the argument list itself don't
			//     count: the fix copies that text through verbatim.
			shouldSuggest := ast.IsOptionalChainRoot(node) ||
				(args != nil && len(args.Nodes) > 0 && countNonSpread(args) < 2) ||
				utils.HasCommentInSpan(ctx.Comments.All(), nodeRange.Pos(), commentBound)

			buildFixes := func() []rule.RuleFix {
				if shouldSuggest {
					return nil
				}
				return []rule.RuleFix{rule.RuleFixReplaceRange(nodeRange, fixText)}
			}
			buildSuggestions := func() []rule.RuleSuggestion {
				if !shouldSuggest {
					return nil
				}
				msg := useLiteralMessage()
				if addSemicolon {
					msg = useLiteralAfterSemicolonMessage()
				}
				return []rule.RuleSuggestion{{
					Message:  msg,
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(nodeRange, fixText)},
				}}
			}

			ctx.ReportNodeWithDeferredFixesAndSuggestions(node, preferLiteralMessage(), buildFixes, buildSuggestions)
		}

		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
}
