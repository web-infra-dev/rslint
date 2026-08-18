package no_object_constructor

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func preferLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferLiteral",
		Description: "The object literal notation {} is preferable.",
	}
}

func useLiteralMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: fmt.Sprintf("Replace with '%s'.", replacement),
		Data:        map[string]string{"replacement": replacement},
	}
}

func useLiteralAfterSemicolonMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteralAfterSemicolon",
		Description: fmt.Sprintf("Replace with '%s', add preceding semicolon.", replacement),
		Data:        map[string]string{"replacement": replacement},
	}
}

// calleeAndArguments extracts the callee and argument list shared by
// CallExpression and NewExpression, the two forms this rule inspects.
func calleeAndArguments(node *ast.Node) (*ast.Node, *ast.NodeList) {
	switch node.Kind {
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		return callExpr.Expression, callExpr.Arguments
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		return newExpr.Expression, newExpr.Arguments
	}
	return nil, nil
}

// needsWrappingParens mirrors upstream's needsParentheses: a bare `{}` would
// be misparsed (as a block, or as the arrow function's own block body) in
// either of two positions — the start of an ExpressionStatement, or directly
// after an `=>`, where it opens the arrow function's concise body regardless
// of whether the call is the whole body (`() => Object()`) or only starts it
// (`() => Object().x`).
func needsWrappingParens(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if utils.IsStartOfExpressionStatement(sourceFile, node) {
		return true
	}
	prevToken, ok := utils.TokenBeforePosition(sourceFile, utils.TrimNodeTextRange(sourceFile, node).Pos())
	return ok && prevToken.Kind == ast.KindEqualsGreaterThanToken
}

func buildSuggestion(ctx rule.RuleContext, node *ast.Node) rule.RuleSuggestion {
	sourceFile := ctx.SourceFile

	if !needsWrappingParens(sourceFile, node) {
		return rule.RuleSuggestion{
			Message:  useLiteralMessage("{}"),
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, "{}")},
		}
	}

	// NOTE: Unlike ESLint, utils.NeedsPrecedingSemicolon doesn't model
	// TypeScript-only node kinds (type positions, ambient/overload function
	// declarations, import-equals declarations), so it falls back to the
	// conservative "needs a semicolon" answer there — e.g. after
	// `type T = Foo` or `import Foo = Bar`. The extra `;` never changes
	// behavior, only adds a redundant character; see the rule doc's
	// "Differences from ESLint" section.
	if utils.NeedsPrecedingSemicolon(sourceFile, node) {
		return rule.RuleSuggestion{
			Message:  useLiteralAfterSemicolonMessage("({})"),
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, ";({})")},
		}
	}
	return rule.RuleSuggestion{
		Message:  useLiteralMessage("({})"),
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, "({})")},
	}
}

// https://eslint.org/docs/latest/rules/no-object-constructor
var NoObjectConstructorRule = rule.Rule{
	Name:   "no-object-constructor",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			callee, args := calleeAndArguments(node)
			if callee == nil {
				return
			}
			// ESTree flattens grouping parentheses around the callee, so
			// `(Object)()` still has an Identifier callee; TS-only wrappers
			// (non-null assertion, `as`/`satisfies`) are deliberately left
			// alone since ESTree never collapses those back to a bare
			// Identifier either.
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "Object" {
				return
			}
			if args != nil && len(args.Nodes) > 0 {
				return
			}

			// A local declaration shadows the global constructor.
			if utils.IsShadowed(callee, "Object") {
				return
			}
			// A config `/* global Object: off */` / `languageOptions.globals`
			// entry un-declares the builtin, so `Object` no longer resolves to
			// a known global — ESLint's `getVariableByName` would return
			// undefined and the rule stays silent.
			if !ctx.Globals.Access("Object").IsDeclared() {
				return
			}

			ctx.ReportNodeWithDeferredSuggestions(node, preferLiteralMessage(), func() []rule.RuleSuggestion {
				return []rule.RuleSuggestion{buildSuggestion(ctx, node)}
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
}
