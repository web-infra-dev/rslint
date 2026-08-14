// TestYodaExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
// TestYodaExtrasEditDemand covers autofix/diagnostic-only edit-demand
// invariance for the rule's single deferred fix builder.
//
// Dimension 4 walk (rows that don't apply to yoda, with reasons):
//   - N/A declaration/container forms (class/function shapes): yoda only
//     inspects BinaryExpression comparisons, never functions or classes.
//   - N/A graceful degradation (SpreadAssignment/RestElement, empty bodies):
//     yoda never inspects object/array patterns or statement bodies.
//   - Access/key forms (identifier/string/numeric/computed/private key,
//     bracket forms) are NOT N/A — they are exercised extensively by the
//     upstream exceptRange suite via IsSameReference and are not repeated
//     here.
package yoda

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestYodaExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&YodaRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: single- and multi-level parenthesized receiver on the
			// non-literal side ----
			{Code: `if ((x) === "red") {}`, Options: "never"},
			{Code: `if (((x)) === "red") {}`, Options: "never"},

			// ---- Dimension 4: TS non-null assertion on the non-literal side ----
			{Code: `if (x! === "red") {}`, Options: "never"},

			// ---- Dimension 4: TS `as`/`satisfies` wrappers are not unwrapped, matching
			// ESLint (TSAsExpression/TSSatisfiesExpression are not an ESTree Literal) ----
			{Code: `if (("red" as const) === value) {}`, Options: "never"},
			{Code: `if (("red" satisfies string) === value) {}`, Options: "never"},

			// ---- Dimension 4: optional chain member as the non-literal side ----
			{Code: `if (foo?.bar === "red") {}`, Options: "never"},

			// ---- Dimension 4: PrivateIdentifier access in a range test ----
			{Code: `if (0 <= this.#x && this.#x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: a bare range test wrapped in genuinely redundant parens
			// (not an if/while condition) is still recognized as parenthesized ----
			{Code: `(0 <= x && x < 1);`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: redundant parens around one side of a range test are
			// walked through when finding the logical parent (tsgo ParenthesizedExpression,
			// unlike ESTree, would otherwise sit between the comparison and its `&&`) ----
			{Code: `if ((0 <= x) && x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Real-user: `undefined` is a global Identifier, not an ESTree Literal,
			// so yoda does not enforce ordering against it (a common "why doesn't this
			// get flagged" surprise) ----
			{Code: `if (undefined === x) {}`, Options: "never"},

			// ---- Real-user: a namespaced/enum-style constant on the left is a
			// PropertyAccessExpression, not a literal, so it is never mistaken for a
			// Yoda condition ----
			{Code: `if (Color.Red === value) {}`, Options: "never"},

			// ---- Locks in upstream isEqualityOperator(): `!=`/`!==` are comparison
			// operators but not equality operators, so onlyEquality disables enforcement
			// for them too, not just `<`/`>`/`<=`/`>=` ----
			{Code: `if (x !== 5) {}`, Options: []any{"always", map[string]any{"onlyEquality": true}}},

			// ---- Combination matrix: exceptRange and onlyEquality both true; onlyEquality
			// alone already exempts the non-equality range comparisons ----
			{Code: `if (0 <= x && x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true, "onlyEquality": true}}},

			// ---- Bug fix: a regex literal is a valid Literal bound for upstream's
			// getNormalizedLiteral (it only checks node.type === "Literal", not the
			// value's runtime type). JS's abstract relational comparison coerces a
			// RegExp via ToPrimitive to its source text, so "/a/" <= "/b/" compares
			// lexicographically and the range is ascending ----
			{Code: `if (/a/ <= x && x < /b/) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: optional chaining under "always" mode ----
			{
				Code:    `if (data?.status === 200) {}`,
				Output:  []string{`if (200 === data?.status) {}`},
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected literal to be on the left side of ===.",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 25,
				}},
			},

			// ---- Locks in upstream isRangeTestOperator(): `>`/`>=` never form a range
			// test, even in an `&&`-joined shape that otherwise looks like one ----
			{
				Code:    `if (0 >= x && x > 1) {}`,
				Output:  []string{`if (x <= 0 && x > 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Locks in upstream isBetweenTest(): `||` never satisfies the `&&`-only
			// "between" test, so a superficially between-shaped `||` expression is not
			// exempted (isOutsideTest's own reference check also fails: 0 and 1 are
			// unrelated literals, not the same reference) ----
			{
				Code:    `if (0 <= x || x < 1) {}`,
				Output:  []string{`if (x >= 0 || x < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Locks in upstream isRangeTest(): the sibling of a comparison inside
			// the `&&`/`||` must itself be a BinaryExpression; a plain identifier
			// operand disqualifies the whole shape from being a range test ----
			{
				Code:    `if (flag && 1 > x) {}`,
				Output:  []string{`if (flag && x < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 13}},
			},

			// ---- Dimension 3: multi-line code — whitespace/newlines around the operator
			// are preserved verbatim by the fix, same as inline comments ----
			{
				Code:    "if (0 <=\n  x) {}",
				Output:  []string{"if (x >=\n  0) {}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 2, EndColumn: 4}},
			},

			// ---- Bug fix: mixed BigInt/Number range bounds must compare exactly, not
			// through a lossy float64 conversion. 9007199254740993n rounds to
			// 9007199254740992 in float64 (both exceed 2^53), which would make the
			// bounds look ascending when they are not; JS's actual BigInt-vs-Number
			// abstract relational comparison is exact, so this is not a valid range
			// test and exceptRange must not exempt it ----
			{
				Code:    `if (9007199254740993n <= x && x < 9007199254740992) {}`,
				Output:  []string{`if (x >= 9007199254740993n && x < 9007199254740992) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Bug fix: mixed BigInt/String range bounds must also compare exactly.
			// JS's abstract relational comparison converts the String bound via
			// StringToBigInt for an exact BigInt comparison, never through a lossy
			// float64 conversion ----
			{
				Code:    `if (9007199254740993n <= x && x < "9007199254740992") {}`,
				Output:  []string{`if (x >= 9007199254740993n && x < "9007199254740992") {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if ("9007199254740993" <= x && x < 9007199254740992n) {}`,
				Output:  []string{`if (x >= "9007199254740993" && x < 9007199254740992n) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
		},
	)
}

// TestYodaExtrasEditDemand verifies that diagnostic identity (range, message)
// is independent of edit demand, and that the autofix is only materialized
// when requested — yoda's only edit artifact is a single deferred fix.
func TestYodaExtrasEditDemand(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     tspath.Path("/edit-demand.ts"),
	}, `1 < a`, core.ScriptKindTS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(YodaRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})

		statement := sourceFile.Statements.Nodes[0]
		target := statement.AsExpressionStatement().Expression
		YodaRule.Run(ctx, nil)[target.Kind](target)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	none := run(rule.EditDemandNone)
	autofix := run(rule.EditDemandAutofix)
	suggestion := run(rule.EditDemandSuggestion)
	all := run(rule.EditDemandAll)

	for _, got := range []rule.RuleDiagnostic{none, autofix, suggestion, all} {
		if got.Range != none.Range || !reflect.DeepEqual(got.Message, none.Message) {
			t.Fatalf("diagnostic identity changed with edit demand: %#v vs %#v", got, none)
		}
		if got.Suggestions != nil {
			t.Fatalf("yoda has no suggestions, but got some: %#v", got.Suggestions)
		}
	}

	if none.FixesPtr != nil {
		t.Fatalf("EditDemandNone consumer received fixes: %#v", none.Fixes())
	}
	if suggestion.FixesPtr != nil {
		t.Fatalf("EditDemandSuggestion consumer received fixes: %#v", suggestion.Fixes())
	}
	if len(autofix.Fixes()) != 1 || autofix.Fixes()[0].Text != "a > 1" {
		t.Fatalf("autofix = %#v, want a single fix replacing with %q", autofix.Fixes(), "a > 1")
	}
	if len(all.Fixes()) != 1 || all.Fixes()[0].Text != "a > 1" {
		t.Fatalf("all-edits fix = %#v, want a single fix replacing with %q", all.Fixes(), "a > 1")
	}
}
