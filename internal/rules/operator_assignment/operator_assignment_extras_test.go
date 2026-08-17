// TestOperatorAssignmentExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package operator_assignment

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOperatorAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&OperatorAssignmentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: template interpolation is out of scope (matches
			// ESLint's actual v10.8.1 behavior; a rule-change request to catch
			// this — eslint/eslint#15840 — was closed without landing) ----
			{Code: "foo = `${foo} baz ${qux}`"},

			// ---- Real-user: eslint/eslint#15840 (rule-change request, closed,
			// not implemented) — template-literal RHS must not be treated as a
			// shorthand-able BinaryExpression ----
			{Code: "foo = `${foo}`"},

			// ---- Locks in upstream verify() branch: node.right.type !== "BinaryExpression" (CallExpression) ----
			{Code: `x = foo()`},

			// ---- Locks in upstream verify() branch: comma operator RHS is not in either shorthand operator list ----
			{Code: `x = (a, b)`},

			// ---- Locks in upstream branch: commutative-vs-non-commutative gating (an operator in neither list must not report even when isSameReference matches) ----
			{Code: `x = x == y`},

			// ---- Locks in upstream branch: reversed-operand requires `commutative` specifically (non-commutative reversed must not report at all, not even without a fix) ----
			{Code: `x = y % x`},

			// ---- Locks in prohibit() isLogicalAssignmentOperator exclusion: mixed logical-assignment RHS must still be excluded ----
			{Code: `x &&= y + 1`, Options: []any{"never"}},
			{Code: `x ||= y * 2`, Options: []any{"never"}},
			{Code: `x ??= y - 1`, Options: []any{"never"}},

			// ---- Dimension 4: private identifier vs string-literal key must not
			// be treated as the same reference (different Kind entirely) ----
			{Code: `class C { #x = 0; m() { this.#x = this["#x"] + 1; } }`},

			// ---- Dimension 4: no-substitution-template-literal computed key
			// must NOT be treated as the same reference. ESTree gives `` `y` ``
			// its own "TemplateLiteral" type, distinct from "Literal", and
			// ESLint's isSameReference has no case for it (falls to
			// `default: return false`). Verified against real ESLint 10.8.1:
			// `x[`a`] = x[`a`] + y` reports nothing, unlike the string-literal
			// equivalent `x["a"] = x["a"] + y` ----
			{Code: "x[`y`] = x[`y`] + z"},

			// ---- Dimension 4: graceful degradation — destructuring assignment
			// target (ArrayLiteralExpression) on the left must not crash or
			// misfire; isSameReference has no case for it ----
			{Code: `[a] = a + y`},
			{Code: `({ a } = a + y)`},

			// N/A: SpreadAssignment inside an object literal / RestElement inside
			// a binding pattern — this rule only inspects the top-level shape of
			// BinaryExpression.Left/Right (Identifier / access-expression chains
			// / literal keys); it never walks into object or array literal
			// contents, so spread/rest members inside a literal cannot reach any
			// branch of this rule.
			// N/A: empty class body / empty function body / empty destructuring
			// pattern / empty arguments list — this rule's only listener is
			// KindBinaryExpression; it never visits class, function, or
			// parameter-list nodes.
			// N/A: overload signatures / abstract / declare members — same
			// reason, unrelated node kinds this rule never visits.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single parenthesized receiver is transparent (matches ESLint's paren-eliding ESTree model) ----
			{
				Code:   `(x).y = (x).y + z`,
				Output: []string{`(x).y += z`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			// ---- Dimension 4: multi-level parenthesized receiver ----
			{
				Code:   `((x)).y = ((x)).y + z`,
				Output: []string{`((x)).y += z`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: TS non-null assertion receiver (`X!.y`) is
			// transparent for both same-reference comparison and fixability —
			// `!` has no runtime effect, so the raw-text fix preserves it ----
			{
				Code:    `x!.y = x!.y + z`,
				Output:  []string{`x!.y += z`},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			// ---- Dimension 4: TS type-expression wrapper receiver (`(x as any).y`) — same as above, transparent and fixable ----
			{
				Code:   `(x as any).y = (x as any).y + z`,
				Output: []string{`(x as any).y += z`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: the assertions on the two occurrences must match
			// for a fix to be offered. The fix deletes the right-hand
			// occurrence, so an `as` / `!` that only the deleted text carries
			// would be lost — `x *= 2` would type-check `x` against its
			// declared type again (TS2362 / TS18048). Report only ----
			{
				Code:   `declare let x: string | number; x = (x as number) * 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 33}},
			},
			{
				Code:   `declare let x: number | undefined; x = x! + 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 36}},
			},
			// ---- ... including when the mismatch is nested in the receiver of
			// a member access rather than at the top level ----
			{
				Code:   `declare const obj: any; obj.x = (obj as any).x * 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 25}},
			},
			{
				Code:   `declare const obj: any; (obj as any).x = obj.x * 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 25}},
			},

			// ---- Dimension 4: optional chain propagated through a non-optional
			// continuation (`a?.b.c`) — the `.c` access is not itself optional but
			// IsOptionalChain still reports true for it, so canBeFixed rejects it
			// even though isSameReference matches (optionality is transparent) ----
			{
				Code:   `a.b.c = a?.b.c + d`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: private identifier receiver (this.#x) IS fixable — only the object receiver is checked, not the property name ----
			{
				Code:   `class C { #x = 0; m() { this.#x = this.#x + 1; } }`,
				Output: []string{`class C { #x = 0; m() { this.#x += 1; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 25}},
			},

			// ---- Dimension 4: numeric-literal computed key ----
			{
				Code:   `x[0] = x[0] + y`,
				Output: []string{`x[0] += y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			// ---- Dimension 4: bigint-literal computed key ----
			{
				Code:   `x[0n] = x[0n] + y`,
				Output: []string{`x[0n] += y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			// ---- Dimension 4: regular-expression-literal computed key ----
			{
				Code:   `x[/re/] = x[/re/] + y`,
				Output: []string{`x[/re/] += y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			// ---- Dimension 4: boolean/null keyword-literal computed keys ----
			{
				Code:   `x[true] = x[true] + y`,
				Output: []string{`x[true] += y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},
			{
				Code:   `x[null] = x[null] + y`,
				Output: []string{`x[null] += y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: non-static computed key (Symbol.iterator) — reported (isSameReference recurses structurally) but NOT fixed (argument is not a literal) ----
			{
				Code:   `x[Symbol.iterator] = x[Symbol.iterator] + y`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: `super` is a same-reference base case upstream
			// (isSameReference returns true for two Super nodes), so the
			// assignment is reported; canBeFixed still rejects it because the
			// receiver is neither an Identifier nor `this` ----
			{
				Code:   `class B { x: any } class C extends B { m(y: any) { super.x = super.x + y; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 52}},
			},
			{
				Code:   `class B { x: any } class C extends B { m(y: any) { super["x"] = super["x"] + y; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 52}},
			},
			// ---- Dimension 4: non-static computed key on `super` — the
			// structural fallback compares the key nodes recursively ----
			{
				Code:   `class B { x: any } class C extends B { m(y: any) { super[y] = super[y] + y; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 52}},
			},

			// ---- Locks in upstream branch: a different non-commutative operator with no upstream case (modulo) ----
			// (kept invalid to prove the rule DOES report the forward form for
			// contrast with the reversed-operand valid case above)
			{
				Code:   `x = x % y`,
				Output: []string{`x %= y`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaced", Line: 1, Column: 1}},
			},

			// ---- Locks in "never" fixer precedence branch: higher-precedence right side (CallExpression) needs no parens ----
			{
				Code:    `foo -= bar()`,
				Output:  []string{`foo = foo - bar()`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- Locks in "never" fixer precedence branch: higher-precedence right side (member access) needs no parens ----
			{
				Code:    `foo -= bar.baz`,
				Output:  []string{`foo = foo - bar.baz`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- Locks in "never" fixer precedence branch: an optional-chain
			// right side is a ChainExpression upstream (precedence 18), so it
			// needs no parens ----
			{
				Code:    `foo -= bar?.baz`,
				Output:  []string{`foo = foo - bar?.baz`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- Locks in "never" fixer precedence branch: TS-only right sides
			// are unknown to ESLint's precedence table, which assigns them the
			// lowest precedence so they get parenthesized. Verified against
			// ESLint 10.8.1 + @typescript-eslint/parser ----
			{
				Code:    `foo -= bar!`,
				Output:  []string{`foo = foo - (bar!)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			{
				Code:    `foo -= bar as number`,
				Output:  []string{`foo = foo - (bar as number)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			{
				Code:    `foo -= bar satisfies number`,
				Output:  []string{`foo = foo - (bar satisfies number)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- ... but an already-parenthesized TS-only right side keeps its
			// own parentheses instead of gaining a second pair ----
			{
				Code:    `foo -= (bar!)`,
				Output:  []string{`foo = foo - (bar!)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- Locks in "never" fixer: a right side whose leftmost operand is
			// a bare `as` / `satisfies` is parenthesized as a whole. The root
			// operator (`*`) binds tighter than the one being written (`+`), but
			// `x = x + a as number * b` re-parses as `((x + a) as number) * b`,
			// which computes something else entirely ----
			{
				Code:    `x += a as number * b`,
				Output:  []string{`x = x + (a as number * b)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			{
				Code:    `x += a satisfies number * b`,
				Output:  []string{`x = x + (a satisfies number * b)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},
			// ---- ... but parentheses already in the source stop the walk, so
			// the equivalent explicit grouping needs no second pair ----
			{
				Code:    `x += (a as number) * b`,
				Output:  []string{`x = x + (a as number) * b`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},

			// ---- Locks in "never" fixer: `<<` immediately followed by `(` is
			// re-scanned as the start of a type argument list, so parenthesizing
			// a right side that contains another `<` would emit unparsable
			// source (`x = x << (foo<T>)` is TS1005). Report only ----
			{
				Code:    `type T = number; declare const foo: any; x <<= foo<T>;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 42}},
			},
			{
				Code:    `type T = number; declare const foo: any; x <<= foo<T>()!;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 42}},
			},
			{
				Code:    `type T = number; declare const y: any; x <<= y as Array<T>;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 40}},
			},
			// ---- ... while a `<<` right side with no `<` in it is still fixed ----
			{
				Code:    `foo <<= bar | 1`,
				Output:  []string{`foo = foo << (bar | 1)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},

			// ---- Locks in "never" fixer precedence branch: right-associative `**` still gets conservative parens at equal precedence (matches ESLint's uniform `<=` comparison, which doesn't special-case associativity) ----
			{
				Code:    `foo **= bar ** baz`,
				Output:  []string{`foo = foo ** (bar ** baz)`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 1}},
			},

			// ---- Real-user / regression lock-in: eslint/eslint#15759 fixed a
			// double-`=` bug in the message text (e.g. "(+==)" instead of
			// "(+=)"); assert the exact current message text so this can't
			// silently regress ----
			{
				Code:    `a += b`,
				Output:  []string{`a = a + b`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected",
					Message:   "Unexpected operator assignment (+=) shorthand.",
					Line:      1, Column: 1,
				}},
			},
		},
	)
}

// TestOperatorAssignmentEditDemand locks in that fixes are only materialized
// under a matching edit demand, and that diagnostic identity (message,
// range) stays invariant across demands.
func TestOperatorAssignmentEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`x = x + y;`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      program,
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     OperatorAssignmentRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return OperatorAssignmentRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		if diagnostics[0].Message.Id != "replaced" {
			t.Fatalf("demand %d: message id = %q, want replaced", demand, diagnostics[0].Message.Id)
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)[0]
	autofixOnly := run(rule.EditDemandAutofix)[0]
	suggestionOnly := run(rule.EditDemandSuggestion)[0]
	allEdits := run(rule.EditDemandAll)[0]

	withoutEdits := func(d rule.RuleDiagnostic) rule.RuleDiagnostic {
		d.FixesPtr = nil
		d.Suggestions = nil
		return d
	}
	for demand, d := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       diagnosticsOnly,
		rule.EditDemandAutofix:    autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got, want := withoutEdits(d), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}

	if diagnosticsOnly.FixesPtr != nil {
		t.Error("EditDemandNone unexpectedly materialized a fix")
	}
	if suggestionOnly.FixesPtr != nil {
		t.Error("EditDemandSuggestion unexpectedly materialized a fix (rule has no suggestions)")
	}
	if diagnosticsOnly.Suggestions != nil || autofixOnly.Suggestions != nil || allEdits.Suggestions != nil {
		t.Error("rule unexpectedly materialized suggestions (it has none)")
	}

	if autofixOnly.FixesPtr == nil || len(*autofixOnly.FixesPtr) != 1 {
		t.Fatalf("EditDemandAutofix: fixes = %#v, want exactly one", autofixOnly.FixesPtr)
	}
	if (*autofixOnly.FixesPtr)[0].Text != "x += y" {
		t.Errorf("fix text = %q, want %q", (*autofixOnly.FixesPtr)[0].Text, "x += y")
	}
	if allEdits.FixesPtr == nil || !reflect.DeepEqual(*autofixOnly.FixesPtr, *allEdits.FixesPtr) {
		t.Errorf("autofix-only and all-edits demands produced different fixes")
	}
}
