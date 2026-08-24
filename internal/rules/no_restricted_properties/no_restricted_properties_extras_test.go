package no_restricted_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedPropertiesExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 1 (AST node types): covered inline by the Dimension 4 rows below
// (element-access literal kinds, optional chaining) plus the upstream suite's
// own dot/bracket/regex/computed-key coverage.
//
// Dimension 2 (scoping & nesting) and Dimension 3 (autofix boundaries):
//
//	N/A — the rule does no scope analysis (no ctx.Refs, no function/class-body
//	tracking) and has no autofix or suggestions. Its only "nesting" concern is
//	how deep a destructuring pattern can be, covered by the D4 nesting row.
func TestNoRestrictedPropertiesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedPropertiesRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null / type-expression wrappers on the receiver ----
			// `foo!.bar` / `(foo as Foo).bar` / `(foo satisfies Foo).bar` are never
			// unwrapped back to a bare Identifier (ESTree/typescript-eslint don't
			// collapse these either), so an `object` restriction never matches.
			{Code: `foo!.bar`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `(foo as Foo).bar`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `(foo satisfies Foo).bar`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},

			// ---- Dimension 4: computed pattern key that isn't statically known ----
			// `[x]` names whatever `x` currently holds, not the literal text "x";
			// getStaticPropertyName can't resolve it, so the property is skipped.
			{Code: `let {[x]: y} = foo;`, Options: []any{map[string]any{"property": "x"}}},

			// ---- Real-user: eslint/eslint#19809 (declined) — dotted `object` names never match ----
			// The rule only ever compares a plain Identifier receiver; a compound
			// path in the `object` option can never match any AST shape.
			{Code: `config.apiKey.other`, Options: []any{map[string]any{"object": "config.apiKey", "property": "other"}}},

			// ---- Locks in upstream checkPropertyAccess() arm: matchedObject present
			// but propertyName absent from it does NOT fall back to a whole-object
			// restriction on the same object name ----
			{
				Code: `foo.baz`,
				Options: []any{
					map[string]any{"object": "foo", "property": "bar"},
					map[string]any{"object": "foo"},
				},
			},

			// ---- Locks in getStaticPropertyName: RestElement/SpreadElement is never
			// a static property name (must not crash, must not match) ----
			{Code: `let {...rest} = foo;`, Options: []any{map[string]any{"property": "rest"}}},
			{Code: `({...rest} = foo);`, Options: []any{map[string]any{"property": "rest"}}},

			// ---- Dimension 4: numeric keys stringify with JavaScript's
			// Number::toString, which leaves fixed notation outside
			// [1e-6, 1e21) — so the fixed spelling of such a key never matches ----
			{Code: `foo[1e21]`, Options: []any{map[string]any{"property": "1000000000000000000000"}}},
			{Code: `foo[1e-7]`, Options: []any{map[string]any{"property": "0.0000001"}}},
			{Code: `let {[1e21]: x} = foo;`, Options: []any{map[string]any{"property": "1000000000000000000000"}}},

			// ---- Review regression: Espree parses large radix literals from the
			// raw token before applying JavaScript Number-to-string semantics. The
			// differently rounded value produced from tsgo's literal text must not
			// match an element-access restriction. ----
			{Code: `foo[0x1000000000000281]`, Options: []any{map[string]any{"property": "1152921504606847700"}}},
			{Code: `foo[0x1000000000000281n]`, Options: []any{map[string]any{"property": "1152921504606847500"}}},

			// ---- Graceful degradation: empty object patterns ----
			{Code: `let {} = foo;`, Options: []any{map[string]any{"object": "foo"}}},
			{Code: `({} = foo);`, Options: []any{map[string]any{"object": "foo"}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver, single and multi-level ----
			// tsgo keeps ParenthesizedExpression as an explicit node; ESTree has none,
			// so `(foo).bar` reads `foo` as object.name directly. Unwrapping via
			// ast.SkipParentheses is required to match.
			{
				Code:    `(foo).bar`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `((foo)).bar`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 12,
				}},
			},

			// ---- Dimension 4: property-only restriction is unaffected by the
			// receiver's shape — only the object-side match cares about it ----
			{
				Code:    `foo!.bar`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `(foo as Foo).bar`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 17,
				}},
			},

			// ---- Dimension 4: optional chaining does not change matching ----
			{
				Code:    `foo?.bar`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `foo?.['bar']`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 13,
				}},
			},

			// ---- Dimension 4: element-access key forms normalize like ESLint's
			// String(node.value) — numeric-literal and string-literal keys are the
			// same equivalence class ----
			{
				Code:    `foo[0]`,
				Options: []any{map[string]any{"property": "0"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'0' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 7,
				}},
			},
			{
				Code:    `foo['0']`,
				Options: []any{map[string]any{"property": "0"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'0' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    "foo[`bar`]",
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 11,
				}},
			},

			// ---- Dimension 4: a numeric key at or past 1e21 (and below 1e-6)
			// is named by its exponential spelling, both in element access and
			// as a computed destructuring key ----
			{
				Code:    `foo[1e21]`,
				Options: []any{map[string]any{"property": "1e+21"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1e+21' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `foo[1e-7]`,
				Options: []any{map[string]any{"property": "1e-7"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1e-7' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `foo[1e21]`,
				Options: []any{map[string]any{"object": "foo", "property": "1e+21"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.1e+21' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `let {[1e21]: x} = foo;`,
				Options: []any{map[string]any{"property": "1e+21"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1e+21' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:    `foo[0x1000000000000281]`,
				Options: []any{map[string]any{"property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847500' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 24,
				}},
			},
			{
				Code:    `foo[0b1000000000000000000000000000000000000000000000000001010000001]`,
				Options: []any{map[string]any{"property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847500' is restricted from being used.",
				}},
			},
			{
				Code:    `foo[0o100000000000000001201]`,
				Options: []any{map[string]any{"property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847500' is restricted from being used.",
				}},
			},
			{
				Code:    `foo[0X1_0000_0000_0000_281]`,
				Options: []any{map[string]any{"property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847500' is restricted from being used.",
				}},
			},
			{
				Code:    `foo[(0x1000000000000281)]`,
				Options: []any{map[string]any{"object": "foo", "property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.1152921504606847500' is restricted from being used.",
				}},
			},
			{
				Code:    `foo?.[0x1000000000000281]`,
				Options: []any{map[string]any{"property": "1152921504606847500"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847500' is restricted from being used.",
				}},
			},
			{
				Code:    `foo[0x1000000000000281n]`,
				Options: []any{map[string]any{"property": "1152921504606847617"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'1152921504606847617' is restricted from being used.",
				}},
			},

			// ---- Dimension 4: nesting/traversal — a restricted property several
			// levels deep, each level carrying its own `= {}` default, is still
			// found; the walk doesn't stop at (or bleed past) intermediate levels ----
			{
				Code:    `let {a: {b: {c} = {}} = {}} = foo;`,
				Options: []any{map[string]any{"property": "c"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'c' is restricted from being used.",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 16,
				}},
			},

			// ---- Real-user: eslint/eslint#16412 — destructuring an aliased
			// property straight off a function call result ----
			{
				Code:    `const { bad: alias } = foo();`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 7, EndLine: 1, EndColumn: 21,
				}},
			},

			// ---- Locks in upstream checkPropertyAccess() arm: the pair restriction
			// itself still fires normally alongside the whole-object entry from the
			// same option set (companion to the `foo.baz` valid case above) ----
			{
				Code: `foo.bar`,
				Options: []any{
					map[string]any{"object": "foo", "property": "bar"},
					map[string]any{"object": "foo"},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- Locks in upstream checkPropertyAccess() `else if`: an object-side
			// match that turns out to be allowed still falls through to the
			// whole-property check, instead of short-circuiting to "no report" ----
			{
				Code: `someObj.x`,
				Options: []any{
					map[string]any{"object": "someObj", "allowProperties": []any{"x"}},
					map[string]any{"property": "x"},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'x' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},

			// ---- Locks in upstream isAllowed(): an unresolvable object name
			// (non-Identifier receiver) never matches allowObjects, no matter what
			// the list contains ----
			{
				Code:    `foo().bar`,
				Options: []any{map[string]any{"property": "bar", "allowObjects": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used. Property 'bar' is only allowed on these objects: foo.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
				}},
			},

			// ---- Locks in message formatting: an explicitly empty allow-list is
			// still "present" (JS truthy array), so the message clause renders with
			// nothing joined in — distinct from allow-list absent entirely ----
			{
				Code:    `foo.bad`,
				Options: []any{map[string]any{"property": "bad", "allowObjects": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used. Property 'bad' is only allowed on these objects: .",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar`,
				Options: []any{map[string]any{"object": "foo", "allowProperties": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used. Only these properties are allowed: .",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- Locks in message formatting: an explicitly empty `message` is
			// falsy upstream, so no suffix (and no trailing space) is appended ----
			{
				Code:    `foo.bar`,
				Options: []any{map[string]any{"object": "foo", "property": "bar", "message": ""}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar`,
				Options: []any{map[string]any{"object": "foo", "message": ""}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar`,
				Options: []any{map[string]any{"property": "bar", "message": ""}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- Locks in getStaticPropertyName: a rest sibling never masks the
			// restricted-property check on the other pattern elements ----
			{
				Code:    `let {bad, ...rest} = foo;`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 19,
				}},
			},
		},
	)
}
