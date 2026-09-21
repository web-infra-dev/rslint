package no_unneeded_async_expect_function

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unneeded_async_expect_function"
)

var removeWrapperSuggestion = rule.RuleMessage{
	Id:          "suggestRemoveAsyncWrapper",
	Description: "Pass the awaited call to `expect()` directly",
}

// hasPromiseModifier reports whether the assertion goes through `resolves` or
// `rejects`, the only modifiers that make the wrapper redundant.
//
// `rejects` calls the wrapper itself, so writing the call inside an async
// function adds nothing, and `resolves` rejects a function outright with "You
// must provide a Promise to expect()". Every other chain keeps the function as
// the asserted value: `expect(async () => { await run() }).toBeInstanceOf(
// Function)` and `expect(async () => { await run() }).toThrow()` assert on the
// function object, and unwrapping them would assert on a promise instead.
func hasPromiseModifier(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	return slices.Contains(parsed.Modifiers, "resolves") ||
		slices.Contains(parsed.Modifiers, "rejects")
}

// keepsWrapperBindings reports whether the wrapper's own bindings survive the
// unwrap.
//
// `rejects` calls the wrapper with no arguments and no receiver, so a
// parameter is always `undefined` inside it, but the name the awaited call
// reads disappears with the wrapper and would resolve to something else — or
// to nothing — at the assertion's own scope. Type parameters, the name of a
// named function expression, `this` and `arguments` are bound the same way.
// Arrow functions take `this` and `arguments` from the enclosing scope
// already, so unwrapping leaves them pointing at the same bindings.
func keepsWrapperBindings(fn *ast.Node, awaited *ast.Node) bool {
	if len(fn.Parameters()) != 0 || len(fn.TypeParameters()) != 0 {
		return false
	}
	if fn.Kind != ast.KindFunctionExpression {
		return true
	}
	if fn.Name() != nil {
		return false
	}
	return !referencesCallerBindings(awaited)
}

// referencesCallerBindings reports whether the expression reads `this` or
// `arguments`, which a function expression binds and an expression in the
// assertion's scope does not.
func referencesCallerBindings(node *ast.Node) bool {
	found := false
	var walk func(*ast.Node) bool
	walk = func(child *ast.Node) bool {
		if found || child == nil {
			return true
		}
		switch child.Kind {
		case ast.KindThisKeyword:
			found = true
			return true
		case ast.KindIdentifier:
			if child.Text() == "arguments" {
				found = true
				return true
			}
		}
		return child.ForEachChild(walk)
	}
	walk(node)
	return found
}

// keepsComments reports whether every comment the wrapper carries survives the
// unwrap. Only the awaited call's own text is kept, so a comment written
// anywhere else inside the wrapper would be deleted.
func keepsComments(ctx rule.RuleContext, wrapper *ast.Node, awaited *ast.Node) bool {
	wrapperRange := utils.TrimNodeTextRange(ctx.SourceFile, wrapper)
	awaitedRange := utils.TrimNodeTextRange(ctx.SourceFile, awaited)
	comments := ctx.Comments.All()
	return !utils.HasCommentInSpan(comments, wrapperRange.Pos(), awaitedRange.Pos()) &&
		!utils.HasCommentInSpan(comments, awaitedRange.End(), wrapperRange.End())
}

func buildSuggestions(ctx rule.RuleContext, match shared.Match) []rule.RuleSuggestion {
	// expect<T>() pins the asserted value's type, and the wrapper is what T
	// describes; the unwrapped call has the awaited value's type instead.
	if match.HeadCall.AsCallExpression().TypeArguments != nil {
		return nil
	}
	fn := ast.SkipParentheses(match.Wrapper)
	if !keepsWrapperBindings(fn, match.Awaited) || !keepsComments(ctx, match.Wrapper, match.Awaited) {
		return nil
	}

	return []rule.RuleSuggestion{{
		Message: removeWrapperSuggestion,
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(
			ctx.SourceFile,
			match.Wrapper,
			utils.TrimmedNodeText(ctx.SourceFile, match.Awaited),
		)},
	}}
}

var NoUnneededAsyncExpectFunctionRule = shared.NewRule(shared.Config{
	Name: "rstest/no-unneeded-async-expect-function",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{ParseExpect: func(node *ast.Node) *ast.Node {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				// Only expect() and expect.soft() assert on the value handed to
				// them. expect.poll() takes a callback it retries, and combining
				// it with resolves or rejects throws, while expect.element()
				// asserts on a browser locator.
				(parsed.Entry != rstestUtils.RstestExpectEntryCall &&
					parsed.Entry != rstestUtils.RstestExpectEntrySoft) ||
				!hasPromiseModifier(parsed) {
				return nil
			}
			return parsed.Head
		}}
	},
	// The unwrap is reported as a suggestion rather than a fix: a call that
	// throws synchronously rejects the promise the wrapper returns, but throws
	// out of expect()'s argument list once the wrapper is gone, which turns a
	// passing assertion into a failing test. Whether the awaited call can do
	// that is not decidable from source alone.
	Report: func(ctx rule.RuleContext, match shared.Match) {
		ctx.ReportNodeWithDeferredSuggestions(
			match.Wrapper,
			shared.Message,
			func() []rule.RuleSuggestion { return buildSuggestions(ctx, match) },
		)
	},
})
