// TestDotNotationExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in. See dot_notation_upstream_test.go for the migrated upstream suite.
package dot_notation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDotNotationExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DotNotationRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: access/key forms - element access via a non-literal
		// computed expression (property access, not a literal) must not be
		// treated as a literal key. ----
		{Code: "a[Symbol.iterator];"},

		// ---- Dimension 4: access/key forms - numeric-looking string that
		// isn't a valid identifier (starts with a digit) stays as bracket
		// notation. ----
		{Code: "data['123abc'];"},

		// ---- N/A: Declaration/container forms (class/function shapes) don't
		// apply - this rule only inspects MemberExpression / bracket-access
		// nodes, never function or class declarations. ----

		// ---- N/A: Nesting/traversal boundaries don't apply - the rule has no
		// ancestor walk or scope-boundary logic; each MemberExpression is
		// judged independently of its container. ----

		// ---- N/A: Graceful-degradation shapes (SpreadAssignment inside an
		// object literal, RestElement in a binding pattern, empty class/
		// function bodies) don't apply - the rule never visits object
		// literals or binding patterns directly (the one destructuring
		// interaction, bare `let[`, is covered by Layer 1's `let.if()` /
		// `let?.true` cases). ----
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: receiver wrapper - multi-level parenthesized key,
		// `a[(('b'))]`. SkipParentheses must unwrap repeatedly. ----
		{
			Code:   "a[(('b'))];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
		},
		// ---- Dimension 4: receiver wrapper - multi-level parenthesized
		// object; raw source text (including parens) is preserved verbatim. ----
		{
			Code:   "((foo))['bar'];",
			Output: []string{"((foo)).bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 9}},
		},
		// ---- Dimension 4: receiver wrapper - TS non-null assertion on the
		// object (`X!.y`). tsgo represents `!` as an explicit
		// NonNullExpression node; the rule must read its raw source text
		// (including the `!`) rather than unwrap it like parens. ----
		{
			Code:   "a!['b'];",
			Output: []string{"a!.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 4}},
		},
		// ---- Dimension 4: receiver wrapper - TS `as` type-assertion on the
		// object. ----
		{
			Code:   "(x as any)['b'];",
			Output: []string{"(x as any).b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 12}},
		},
		// ---- Dimension 4: receiver wrapper - TS `satisfies` on the object. ----
		{
			Code:   "(x satisfies object)['b'];",
			Output: []string{"(x satisfies object).b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 22}},
		},
		// ---- Dimension 4: receiver wrapper - optional-chain object (`a?.b`)
		// feeding into a further bracket access; tsgo has no ChainExpression
		// wrapper, just a flag on each access node in the chain. ----
		{
			Code:   "a?.b['c'];",
			Output: []string{"a?.b.c;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 6}},
		},
		// ---- Locks in checkComputedProperty(): `typeof value === "number"`
		// exclusion from isDecimalInteger - a BigInt literal object must NOT
		// get the disambiguating space (only `number`-typed literals do). ----
		{
			Code:   "5n['prop'];",
			Output: []string{"5n.prop;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 4}},
		},
		// ---- Real-user: bracket-accessing an environment-style namespace
		// object (`process.env['NAME']`) is one of the most common
		// dot-notation false-negative shapes reported against upstream. ----
		{
			Code:   "process.env['NODE_ENV'];",
			Output: []string{"process.env.NODE_ENV;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 13}},
		},
		// ---- Real-user: chained optional-access on an API response object
		// (`response?.data?.['items']`) - a common shape once optional
		// chaining shipped, combining two independent optional accesses. ----
		{
			Code:   "response?.data?.['items'];",
			Output: []string{"response?.data?.items;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 18}},
		},
		// ---- Real-user: `obj['hasOwnProperty']`-style prototype-guard idiom.
		// Some codebases intentionally use bracket notation here to visually
		// flag "this is a lookup, not a real method call", but the rule
		// doesn't special-case Object.prototype member names - it still
		// reports/converts, matching upstream exactly. ----
		{
			Code:   "obj['hasOwnProperty'];",
			Output: []string{"obj.hasOwnProperty;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
		},
		// ---- Regression: a comment sitting between the object and the `[`
		// must not be mistaken for the bracket itself just because it
		// contains a literal '[' character. Locating the bracket via the
		// real token stream (skipping trivia) instead of a raw byte scan
		// keeps the comment intact in the fixed output, matching eslint. ----
		{
			Code:   "foo /* [ */ ['bar'];",
			Output: []string{"foo /* [ */ .bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 14}},
		},
		// ---- Regression: same class of bug in the dot->bracket direction -
		// a comment between the object and the `.` containing a literal '?'
		// must not be mistaken for an optional-chain operator. ----
		{
			Code:    "foo /* ? */ .while",
			Output:  []string{`foo /* ? */ ["while"]`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 14}},
		},
		// ---- Regression: a comment between the object and the `.`
		// containing a literal '.' must not be mistaken for the real
		// operator either. ----
		{
			Code:    "foo /* . */ .while",
			Output:  []string{`foo /* . */ ["while"]`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 14}},
		},
		// ---- Regression: a comment nested inside parens around the key
		// must suppress the autofix - it can't survive a `.bar` rewrite. The
		// diagnostic is still reported, but --fix leaves the code unchanged. ----
		{
			Code:   "foo[(/* keep */ 'bar')];",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 17}},
		},
		// ---- Regression: nested bracket accesses. Each fix replaces
		// only its own `['key']` bracket part (never the whole
		// ElementAccessExpression), so the two ranges don't overlap and both
		// convert in a single fix pass, preserving inter-access whitespace. ----
		{
			Code:   "a['b']  ['c'];",
			Output: []string{"a.b  .c;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot", Line: 1, Column: 10},
				{MessageId: "useDot", Line: 1, Column: 3},
			},
		},
		// ---- Regression: same single-pass requirement for a chain of
		// optional bracket accesses - `?.['b']` becomes `?.b` by replacing
		// only the `['b']` part (the untouched `?.` supplies the dot). ----
		{
			Code:   "a?.['b']?.['c'];",
			Output: []string{"a?.b?.c;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot", Line: 1, Column: 12},
				{MessageId: "useDot", Line: 1, Column: 5},
			},
		},
		// ---- Regression: same single-pass requirement in the
		// dot->bracket direction - each fix replaces only its own `.keyword`
		// operator + name part, never the whole PropertyAccessExpression. ----
		{
			Code:    "a.if.while;",
			Output:  []string{`a["if"]["while"];`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useBrackets", Line: 1, Column: 6},
				{MessageId: "useBrackets", Line: 1, Column: 3},
			},
		},
		// ---- Regression: whitespace between an optional-chain `?.` and the
		// following `[` must be preserved verbatim, not collapsed into a
		// second (invalid) copy of the operator. ----
		{
			Code:   "obj?.  ['prop'];",
			Output: []string{"obj?.  prop;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 9}},
		},
		// ---- Regression: same whitespace-preservation requirement across a
		// newline between `?.` and `[`. ----
		{
			Code:   "obj?.\n['prop'];",
			Output: []string{"obj?.\nprop;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 2, Column: 2}},
		},
	})
}
