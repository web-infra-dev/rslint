package getter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These cases lock in ESLint v10.8.1's getter-return behavior, including the
// SourceCode global-reference checks used for property descriptors.
func TestGetterReturnUpstreamAlignment(t *testing.T) {
	allowImplicit := map[string]any{"allowImplicit": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&GetterReturnRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo = { get: function () {} };`},
			{Code: `Object.defineProperty({ get() {} }, 'foo', { value: 1 });`},
			{Code: `Reflect.defineProperty({ get() {} }, 'foo', { value: 1 });`},
			{Code: `Object.defineProperties({ foo: { get() {} } }, { bar: { value: 1 } });`},
			{Code: `Object.create({ foo: { get() {} } }, { bar: { value: 1 } });`},
			{Code: `let Object; Object.defineProperty(foo, 'bar', { get() {} });`},
			{Code: `function f() { Reflect.defineProperty(foo, 'bar', { get() {} }); var Reflect; }`},
			{Code: `function f(Object) { Object.defineProperties(foo, { bar: { get() {} } }); }`},
			{Code: `if (x) { const Object = getObject(); Object.create(foo, { bar: { get() {} } }); }`},
			{Code: `/* globals Object:off */ Object.defineProperty(foo, 'bar', { get() {} });`},
			{
				Code:            `Reflect.defineProperty(foo, 'bar', { get() {} });`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
			},
			{
				Code:    `Reflect.defineProperty(foo, 'bar', { get() {} });`,
				Globals: map[string]any{"Reflect": "off"},
			},
			{Code: `Object.defineProperty(foo, 'bar', { get: () => value });`},
			{Code: `Object.create(foo, { bar: { get: () => value } });`},
			{Code: `class C { get value() { throw new Error(); } }`},
			{Code: `class C { get value() { while (true) {} } }`},
			{Code: `class C { get value() { try { return 1; } catch (error) {} } }`},
			{Code: `class C { get value() { for (;;) { try { return 1; } finally { continue; } } } }`},
			{
				Code:    `class C { get value() { if (condition) return; return; } }`,
				Options: allowImplicit,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = { get bar() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'bar'.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code: "var foo = { get\n bar () {} };",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'bar'.",
					Line:      1,
					Column:    13,
					EndLine:   2,
					EndColumn: 6,
				}},
			},
			{
				Code: `class C { static get value() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in static getter 'value'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code: `var foo = { get bar() { return; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'bar'.",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 32,
				}},
			},
			{
				Code:    `var foo = { get bar() {} };`,
				Options: allowImplicit,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'bar'.",
					Line:      1,
					Column:    13,
				}},
			},
			{
				Code:    `var foo = { get bar() { if (condition) return; } };`,
				Options: allowImplicit,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAlways",
					Message:   "Expected getter 'bar' to always return a value.",
					Line:      1,
					Column:    13,
				}},
			},
			{
				Code: `var foo = { get bar() { if (condition) return; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Message:   "Expected to return a value in getter 'bar'.",
						Line:      1,
						Column:    40,
					},
					{
						MessageId: "expectedAlways",
						Message:   "Expected getter 'bar' to always return a value.",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { get: function (){} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    37,
					EndLine:   1,
					EndColumn: 51,
				}},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { get(){} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    37,
					EndLine:   1,
					EndColumn: 40,
				}},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { ...descriptor, get() {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    52,
					EndLine:   1,
					EndColumn: 55,
				}},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { get: () => {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    37,
					EndLine:   1,
					EndColumn: 42,
				}},
			},
			{
				Code: `Object.defineProperties(foo, { bar: { get: function () {} } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    39,
				}},
			},
			{
				Code: `Reflect.defineProperty(foo, 'bar', { get: function (){} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    38,
					EndLine:   1,
					EndColumn: 52,
				}},
			},
			{
				Code: `Object.create(foo, { bar: { get: function() {} } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 42,
				}},
			},
			{
				Code: `Object?.defineProperty(foo, 'bar', { ['get']: function (){} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
					Line:      1,
					Column:    38,
				}},
			},
			{
				Code: `(Object)[('defineProperty')](foo, 'bar', ({ get: (function () {}) }));`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in method 'get'.",
				}},
			},
			{
				Code: `class C { get value() { while (true) { if (condition) break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'value'.",
					Line:      1,
					Column:    11,
				}},
			},
			{
				Code: `class C { get value() { try { return 1; } finally { if (condition) return; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'value'.",
					Line:      1,
					Column:    68,
				}},
			},
			{
				Code: `class C { get value() { function nested() { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'value'.",
					Line:      1,
					Column:    11,
				}},
			},
			{
				Code: `class C { get value() { throw error; return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected to return a value in getter 'value'.",
					Line:      1,
					Column:    38,
				}},
			},
			{
				Code: `class C { get value() { if (false) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAlways",
					Message:   "Expected getter 'value' to always return a value.",
					Line:      1,
					Column:    11,
				}},
			},
			{
				Code: `class C { get value() { if (true) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAlways",
					Message:   "Expected getter 'value' to always return a value.",
					Line:      1,
					Column:    11,
				}},
			},
			{
				Code: `class C { get value() { try { return value; } catch (error) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAlways",
					Message:   "Expected getter 'value' to always return a value.",
					Line:      1,
					Column:    11,
				}},
			},
			{
				Code: `class C { get value() { do { try { return 1; } finally { continue; } } while (false); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAlways",
					Message:   "Expected getter 'value' to always return a value.",
					Line:      1,
					Column:    11,
				}},
			},
		},
	)
}
