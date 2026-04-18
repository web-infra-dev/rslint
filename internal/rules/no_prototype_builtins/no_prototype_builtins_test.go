// cspell:ignore Enumerabl

package no_prototype_builtins

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoPrototypeBuiltinsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPrototypeBuiltinsRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `Object.prototype.hasOwnProperty.call(foo, 'bar')`},
			{Code: `Object.prototype.isPrototypeOf.call(foo, 'bar')`},
			{Code: `Object.prototype.propertyIsEnumerable.call(foo, 'bar')`},
			{Code: `Object.prototype.hasOwnProperty.apply(foo, ['bar'])`},
			{Code: `Object.prototype.isPrototypeOf.apply(foo, ['bar'])`},
			{Code: `Object.prototype.propertyIsEnumerable.apply(foo, ['bar'])`},
			{Code: `foo.hasOwnProperty`},
			{Code: `foo.hasOwnProperty.bar()`},
			{Code: `foo(hasOwnProperty)`},
			{Code: `hasOwnProperty(foo, 'bar')`},
			{Code: `isPrototypeOf(foo, 'bar')`},
			{Code: `propertyIsEnumerable(foo, 'bar')`},
			{Code: `({}.hasOwnProperty.call(foo, 'bar'))`},
			{Code: `({}.isPrototypeOf.call(foo, 'bar'))`},
			{Code: `({}.propertyIsEnumerable.call(foo, 'bar'))`},
			{Code: `({}.hasOwnProperty.apply(foo, ['bar']))`},
			{Code: `({}.isPrototypeOf.apply(foo, ['bar']))`},
			{Code: `({}.propertyIsEnumerable.apply(foo, ['bar']))`},
			// Computed access with non-string/identifier key: not a disallowed name.
			{Code: `foo[hasOwnProperty]('bar')`},
			// Different casing: not one of the forbidden names.
			{Code: `foo['HasOwnProperty']('bar')`},
			// Template with extra char is not the forbidden name.
			{Code: "foo[`isPrototypeOff`]('bar')"},
			// Optional chain + not-quite-the-forbidden-name.
			{Code: `foo?.['propertyIsEnumerabl']('bar')`},
			// Numeric / null keys aren't the forbidden names.
			{Code: `foo[1]('bar')`},
			{Code: `foo[null]('bar')`},
			// Private name is not a Object.prototype method name.
			{Code: `class C { #hasOwnProperty; foo() { obj.#hasOwnProperty('bar'); } }`},

			// Out of scope — dynamic keys that happen to spell the forbidden name
			// at runtime still aren't statically resolvable.
			{Code: `foo['hasOwn' + 'Property']('bar')`},
			{Code: "foo[`hasOwnProperty${''}`]('bar')"},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `foo.hasOwnProperty('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Message:   "Do not access Object.prototype method 'hasOwnProperty' from target object.",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.hasOwnProperty.call(foo, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: `foo.isPrototypeOf('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.isPrototypeOf.call(foo, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: `foo.propertyIsEnumerable('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.propertyIsEnumerable.call(foo, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: `foo.bar.hasOwnProperty('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.hasOwnProperty.call(foo.bar, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: `foo.bar.baz.isPrototypeOf('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.isPrototypeOf.call(foo.bar.baz, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: `foo['hasOwnProperty']('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.hasOwnProperty.call(foo, 'bar')`,
							},
						},
					},
				},
			},
			{
				Code: "foo[`isPrototypeOf`]('bar').baz",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.isPrototypeOf.call(foo, 'bar').baz`,
							},
						},
					},
				},
			},
			{
				Code: `foo.bar["propertyIsEnumerable"]('baz')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.propertyIsEnumerable.call(foo.bar, 'baz')`,
							},
						},
					},
				},
			},
			// Can't suggest Object.prototype when Object is shadowed.
			{
				Code: `(function(Object) {return foo.hasOwnProperty('bar');})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},

			// Optional chaining
			{
				Code: `foo?.hasOwnProperty('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			{
				Code: `foo?.bar.hasOwnProperty('baz')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			{
				Code: `foo.hasOwnProperty?.('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			// hasOwnProperty is inside a chain whose optional link is before it:
			// the call may short-circuit, so we cannot suggest a rewrite.
			{
				Code: `foo?.hasOwnProperty('bar').baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			// Optional link is AFTER the call — the fix is safe.
			{
				Code: `foo.hasOwnProperty('bar')?.baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.hasOwnProperty.call(foo, 'bar')?.baz`,
							},
						},
					},
				},
			},
			// SequenceExpression as object: ESLint wraps it in parens; on tsgo
			// the parens are already in the AST, so the same text lands as-is.
			{
				Code: `(a,b).hasOwnProperty('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prototypeBuildIn",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "callObjectPrototype",
								Output:    `Object.prototype.hasOwnProperty.call((a,b), 'bar')`,
							},
						},
					},
				},
			},
			// Parens around an optional chain: may still short-circuit.
			{
				Code: `(foo?.hasOwnProperty)('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			{
				Code: `(foo?.hasOwnProperty)?.('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			{
				Code: `foo?.['hasOwnProperty']('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
			{
				Code: "(foo?.[`hasOwnProperty`])('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn"},
				},
			},
		},
	)
}
