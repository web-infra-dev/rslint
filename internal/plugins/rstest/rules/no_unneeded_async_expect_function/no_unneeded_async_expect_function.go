package no_unneeded_async_expect_function

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func buildSuggestions(ctx rule.RuleContext, match shared.Match) []rule.RuleSuggestion {
	fix := shared.UnwrapFix(ctx, match)
	if fix == nil {
		return nil
	}
	return []rule.RuleSuggestion{{
		Message:  removeWrapperSuggestion,
		FixesArr: []rule.RuleFix{*fix},
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
