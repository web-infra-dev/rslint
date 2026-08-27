package no_restricted_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedPropertiesUpstream migrates the full valid/invalid suite from
// upstream https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-restricted-properties.js
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the no_restricted_properties_extras_test.go file.
func TestNoRestrictedPropertiesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedPropertiesRule,
		[]rule_tester.ValidTestCase{
			// ---- basic object+property pair ----
			{Code: `someObject.someProperty`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}}},
			{Code: `anotherObject.disallowedProperty`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}}},
			{Code: `someObject.someProperty()`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}}},
			{Code: `anotherObject.disallowedProperty()`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}}},
			{Code: `anotherObject.disallowedProperty()`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty", "message": "Please use someObject.allowedProperty instead."}}},
			{Code: `anotherObject['disallowedProperty']()`, Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}}},
			{Code: `obj.toString`, Options: []any{map[string]any{"object": "obj", "property": "__proto__"}}},
			{Code: `toString.toString`, Options: []any{map[string]any{"object": "obj", "property": "foo"}}},
			{Code: `obj.toString`, Options: []any{map[string]any{"object": "obj", "property": "foo"}}},
			{Code: `foo.bar`, Options: []any{map[string]any{"property": "baz"}}},
			{Code: `foo.bar`, Options: []any{map[string]any{"object": "baz"}}},
			{Code: `foo()`, Options: []any{map[string]any{"object": "foo"}}},
			{Code: `foo;`, Options: []any{map[string]any{"object": "foo"}}},
			{Code: `foo[/(?<zero>0)/]`, Options: []any{map[string]any{"property": "null"}}},

			// ---- destructuring: valid shapes (nested / unrelated / non-identifier source) ----
			{Code: `let bar = foo;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let {baz: bar} = foo;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let {unrelated} = foo;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let {baz: {bar: qux}} = foo;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let {bar} = foo.baz;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let {baz: bar} = foo;`, Options: []any{map[string]any{"property": "bar"}}},
			{Code: `let baz; ({baz: bar} = foo)`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let bar;`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},

			// ---- array destructuring is never checked (rule has no ArrayPattern listener) ----
			{Code: `let bar; ([bar = 5] = foo);`, Options: []any{map[string]any{"object": "foo", "property": "1"}}},
			{Code: `function qux({baz: bar} = foo) {}`, Options: []any{map[string]any{"object": "foo", "property": "bar"}}},
			{Code: `let [bar, baz] = foo;`, Options: []any{map[string]any{"object": "foo", "property": "1"}}},
			{Code: `let [, bar] = foo;`, Options: []any{map[string]any{"object": "foo", "property": "0"}}},
			{Code: `let [, bar = 5] = foo;`, Options: []any{map[string]any{"object": "foo", "property": "1"}}},
			{Code: `let bar; ([bar = 5] = foo);`, Options: []any{map[string]any{"object": "foo", "property": "0"}}},
			{Code: `function qux([bar] = foo) {}`, Options: []any{map[string]any{"object": "foo", "property": "0"}}},
			{Code: `function qux([, bar] = foo) {}`, Options: []any{map[string]any{"object": "foo", "property": "0"}}},
			{Code: `function qux([, bar] = foo) {}`, Options: []any{map[string]any{"object": "foo", "property": "1"}}},

			// ---- private class field access is never a static property name ----
			{Code: `class C { #foo; foo() { this.#foo; } }`, Options: []any{map[string]any{"property": "#foo"}}},

			// ---- allowObjects (property-only restriction) ----
			{Code: `someObject.disallowedProperty`, Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"someObject"}}}},
			{Code: `someObject.disallowedProperty; anotherObject.disallowedProperty();`, Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"someObject", "anotherObject"}}}},
			{Code: `someObject.disallowedProperty()`, Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"someObject"}}}},
			{Code: `someObject['disallowedProperty']()`, Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"someObject"}}}},
			{Code: `let {bar} = foo;`, Options: []any{map[string]any{"property": "bar", "allowObjects": []any{"foo"}}}},
			{Code: `let {baz: bar} = foo;`, Options: []any{map[string]any{"property": "baz", "allowObjects": []any{"foo"}}}},

			// ---- allowProperties (object-only restriction) ----
			{Code: `someObject.disallowedProperty`, Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"disallowedProperty"}}}},
			{Code: `someObject.disallowedProperty; someObject.anotherDisallowedProperty();`, Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"disallowedProperty", "anotherDisallowedProperty"}}}},
			{Code: `someObject.disallowedProperty()`, Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"disallowedProperty"}}}},
			{Code: `someObject['disallowedProperty']()`, Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"disallowedProperty"}}}},
			{Code: `let {bar} = foo;`, Options: []any{map[string]any{"object": "foo", "allowProperties": []any{"bar"}}}},
			{Code: `let {baz: bar} = foo;`, Options: []any{map[string]any{"object": "foo", "allowProperties": []any{"baz"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- basic object+property pair ----
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'someObject.disallowedProperty' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"object": "someObject", "property": "disallowedProperty", "message": "Please use someObject.allowedProperty instead."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'someObject.disallowedProperty' is restricted from being used. Please use someObject.allowedProperty instead.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code: `someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()`,
				Options: []any{
					map[string]any{"object": "someObject", "property": "disallowedProperty"},
					map[string]any{"object": "anotherObject", "property": "anotherDisallowedProperty"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedObjectProperty",
						Message:   "'someObject.disallowedProperty' is restricted from being used.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
					},
					{
						MessageId: "restrictedObjectProperty",
						Message:   "'anotherObject.anotherDisallowedProperty' is restricted from being used.",
						Line:      1, Column: 32, EndLine: 1, EndColumn: 71,
					},
				},
			},
			{
				Code:    `foo.__proto__`,
				Options: []any{map[string]any{"property": "__proto__", "message": "Please use Object.getPrototypeOf instead."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'__proto__' is restricted from being used. Please use Object.getPrototypeOf instead.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 14,
				}},
			},
			{
				Code:    `foo['__proto__']`,
				Options: []any{map[string]any{"property": "__proto__", "message": "Please use Object.getPrototypeOf instead."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'__proto__' is restricted from being used. Please use Object.getPrototypeOf instead.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `foo.bar.baz;`,
				Options: []any{map[string]any{"object": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar();`,
				Options: []any{map[string]any{"object": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar.baz();`,
				Options: []any{map[string]any{"object": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar.baz;`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar();`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo.bar.baz();`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:    `foo[/(?<zero>0)/]`,
				Options: []any{map[string]any{"property": "/(?<zero>0)/"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'/(?<zero>0)/' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 18,
				}},
			},
			{
				Code:    `require.call({}, 'foo')`,
				Options: []any{map[string]any{"object": "require", "message": "Please call require() directly."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'require.call' is restricted from being used. Please call require() directly.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 13,
				}},
			},
			{
				Code:    `require['resolve']`,
				Options: []any{map[string]any{"object": "require"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'require.resolve' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 19,
				}},
			},

			// ---- destructuring ----
			{
				Code:    `let {bar} = foo;`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `let {bar: baz} = foo;`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 15,
				}},
			},
			{
				Code:    `let {'bar': baz} = foo;`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `let {bar: {baz: qux}} = foo;`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 22,
				}},
			},
			{
				Code:    `let {bar} = foo;`,
				Options: []any{map[string]any{"object": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `let {bar: baz} = foo;`,
				Options: []any{map[string]any{"object": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 15,
				}},
			},
			{
				Code:    `let {bar} = foo;`,
				Options: []any{map[string]any{"property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bar' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `let bar; ({bar} = foo);`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:    `let bar; ({bar: baz = 1} = foo);`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 25,
				}},
			},
			{
				Code:    `function qux({bar} = foo) {}`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 19,
				}},
			},
			{
				Code:    `function qux({bar: baz} = foo) {}`,
				Options: []any{map[string]any{"object": "foo", "property": "bar"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 24,
				}},
			},
			{
				Code:    `var {['foo']: qux, bar} = baz`,
				Options: []any{map[string]any{"object": "baz", "property": "foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'baz.foo' is restricted from being used.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 24,
				}},
			},
			{
				Code:    `obj['#foo']`,
				Options: []any{map[string]any{"property": "#foo"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'#foo' is restricted from being used.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 12,
				}},
			},
			{
				Code:    `const { bar: { bad } = {} } = foo;`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 21,
				}},
			},
			{
				Code:    `const { bar: { bad } } = foo;`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 21,
				}},
			},
			{
				Code:    `const { bad } = foo();`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 7, EndLine: 1, EndColumn: 14,
				}},
			},
			{
				Code:    `({ bad } = foo());`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 2, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `({ bar: { bad } } = foo);`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:    `({ bar: { bad } = {} } = foo);`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:    `({ bad }) => {};`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 2, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `({ bad } = {}) => {};`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 2, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `({ bad: bar }) => {};`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 2, EndLine: 1, EndColumn: 14,
				}},
			},
			{
				Code:    `({ bar: { bad } = {} }) => {};`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:    `[{ bad }] = foo;`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 2, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				Code:    `const [{ bad }] = foo;`,
				Options: []any{map[string]any{"property": "bad"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'bad' is restricted from being used.",
					Line:      1, Column: 8, EndLine: 1, EndColumn: 15,
				}},
			},

			// ---- allowObjects (property-only restriction) ----
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"anotherObject"}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'disallowedProperty' is restricted from being used. Property 'disallowedProperty' is only allowed on these objects: anotherObject.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"property": "disallowedProperty", "allowObjects": []any{"anotherObject"}, "message": "Please use someObject.allowedProperty instead."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedProperty",
					Message:   "'disallowedProperty' is restricted from being used. Property 'disallowedProperty' is only allowed on these objects: anotherObject. Please use someObject.allowedProperty instead.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code: `someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()`,
				Options: []any{
					map[string]any{"property": "disallowedProperty", "allowObjects": []any{"anotherObject"}},
					map[string]any{"property": "anotherDisallowedProperty", "allowObjects": []any{"someObject"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedProperty",
						Message:   "'disallowedProperty' is restricted from being used. Property 'disallowedProperty' is only allowed on these objects: anotherObject.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
					},
					{
						MessageId: "restrictedProperty",
						Message:   "'anotherDisallowedProperty' is restricted from being used. Property 'anotherDisallowedProperty' is only allowed on these objects: someObject.",
						Line:      1, Column: 32, EndLine: 1, EndColumn: 71,
					},
				},
			},

			// ---- allowProperties (object-only restriction) ----
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"allowedProperty"}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'someObject.disallowedProperty' is restricted from being used. Only these properties are allowed: allowedProperty.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code:    `someObject.disallowedProperty`,
				Options: []any{map[string]any{"object": "someObject", "allowProperties": []any{"allowedProperty"}, "message": "Please use someObject.allowedProperty instead."}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'someObject.disallowedProperty' is restricted from being used. Only these properties are allowed: allowedProperty. Please use someObject.allowedProperty instead.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code: `someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()`,
				Options: []any{
					map[string]any{"object": "someObject", "allowProperties": []any{"anotherDisallowedProperty"}},
					map[string]any{"object": "anotherObject", "allowProperties": []any{"disallowedProperty"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedObjectProperty",
						Message:   "'someObject.disallowedProperty' is restricted from being used. Only these properties are allowed: anotherDisallowedProperty.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 30,
					},
					{
						MessageId: "restrictedObjectProperty",
						Message:   "'anotherObject.anotherDisallowedProperty' is restricted from being used. Only these properties are allowed: disallowedProperty.",
						Line:      1, Column: 32, EndLine: 1, EndColumn: 71,
					},
				},
			},
		},
	)
}
