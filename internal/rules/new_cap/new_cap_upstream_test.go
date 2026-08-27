// TestNewCapUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/new-cap.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the new_cap_extras_test.go file.
package new_cap

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNewCapUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NewCapRule, []rule_tester.ValidTestCase{
		{Code: "var x = new Constructor();"},
		{Code: "var x = new a.b.Constructor();"},
		{Code: "var x = new a.b['Constructor']();"},
		{Code: "var x = new a.b[Constructor]();"},
		{Code: "var x = new a.b[constructor]();"},
		{Code: "var x = new function(){};"},
		{Code: "var x = new _;"},
		{Code: "var x = new $;"},
		{Code: "var x = new Σ;"},
		{Code: "var x = new _x;"},
		{Code: "var x = new $x;"},
		{Code: "var x = new this;"},
		{Code: "var x = Array(42)"},
		{Code: "var x = Boolean(42)"},
		{Code: "var x = Date(42)"},
		{Code: "var x = Date.UTC(2000, 0)"},
		{Code: "var x = Error('error')"},
		{Code: "var x = Function('return 0')"},
		{Code: "var x = Number(42)"},
		{Code: "var x = Object(null)"},
		{Code: "var x = RegExp(42)"},
		{Code: "var x = String(42)"},
		{Code: "var x = Symbol('symbol')"},
		{Code: "var x = BigInt('1n')"},
		{Code: "var x = _();"},
		{Code: "var x = $();"},
		{Code: "var x = Foo(42)", Options: map[string]any{"capIsNew": false}},
		{Code: "var x = bar.Foo(42)", Options: map[string]any{"capIsNew": false}},
		{Code: "var x = Foo.bar(42)", Options: map[string]any{"capIsNew": false}},
		{Code: "var x = bar[Foo](42)"},
		{Code: "var x = bar['Foo'](42)", Options: map[string]any{"capIsNew": false}},
		{Code: "var x = Foo.bar(42)"},
		{Code: "var x = new foo(42)", Options: map[string]any{"newIsCap": false}},
		{Code: "var o = { 1: function() {} }; o[1]();"},
		{Code: "var o = { 1: function() {} }; new o[1]();"},
		{Code: "var x = Foo(42);", Options: map[string]any{"capIsNew": true, "capIsNewExceptions": []any{"Foo"}}},
		{Code: "var x = Foo(42);", Options: map[string]any{"capIsNewExceptionPattern": "^Foo"}},
		{Code: "var x = new foo(42);", Options: map[string]any{"newIsCap": true, "newIsCapExceptions": []any{"foo"}}},
		{Code: "var x = new foo(42);", Options: map[string]any{"newIsCapExceptionPattern": "^foo"}},
		{Code: "var x = Object(42);", Options: map[string]any{"capIsNewExceptions": []any{"Foo"}}},
		{Code: "var x = Foo.Bar(42);", Options: map[string]any{"capIsNewExceptions": []any{"Bar"}}},
		{Code: "var x = Foo.Bar(42);", Options: map[string]any{"capIsNewExceptions": []any{"Foo.Bar"}}},
		{Code: "var x = Foo.Bar(42);", Options: map[string]any{"capIsNewExceptionPattern": "^Foo\\.."}},
		{Code: "var x = new foo.bar(42);", Options: map[string]any{"newIsCapExceptions": []any{"bar"}}},
		{Code: "var x = new foo.bar(42);", Options: map[string]any{"newIsCapExceptions": []any{"foo.bar"}}},
		{Code: "var x = new foo.bar(42);", Options: map[string]any{"newIsCapExceptionPattern": "^foo\\.."}},
		{Code: "var x = new foo.bar(42);", Options: map[string]any{"properties": false}},
		{Code: "var x = Foo.bar(42);", Options: map[string]any{"properties": false}},
		{Code: "var x = foo.Bar(42);", Options: map[string]any{"capIsNew": false, "properties": false}},

		// Optional chaining
		{Code: "foo?.bar();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "(foo?.bar)();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "new (foo?.Bar)();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "(foo?.Bar)();", Options: map[string]any{"properties": false}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "new (foo?.bar)();", Options: map[string]any{"properties": false}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "Date?.UTC();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
		{Code: "(Date?.UTC)();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "var x = new c();",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "lower", Message: "A constructor name should not start with a lowercase letter.",
				Line: 1, Column: 13, EndLine: 1, EndColumn: 14,
			}},
		},
		{Code: "var x = new φ;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 13, EndLine: 1, EndColumn: 14}}},
		{Code: "var x = new a.b.c;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 17, EndLine: 1, EndColumn: 18}}},
		{Code: "var x = new a.b['c'];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}}},
		{
			Code: "var b = Foo();",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "upper", Message: "A function with a name starting with an uppercase letter should only be used as a constructor.",
				Line: 1, Column: 9, EndLine: 1, EndColumn: 12,
			}},
		},
		{Code: "var b = a.Foo();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 11, EndLine: 1, EndColumn: 14}}},
		{Code: "var b = a['Foo']();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 11, EndLine: 1, EndColumn: 16}}},
		{Code: "var b = a.Date.UTC();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 16, EndLine: 1, EndColumn: 19}}},
		{Code: "var b = UTC();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		{Code: "var a = B.C();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 11, EndLine: 1, EndColumn: 12}}},
		{Code: "var a = B\n.C();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 2, Column: 2, EndLine: 2, EndColumn: 3}}},
		{Code: "var a = new B.c();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 15, EndLine: 1, EndColumn: 16}}},
		{Code: "var a = new B.\nc();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 2, Column: 1, EndLine: 2, EndColumn: 2}}},
		{Code: "var a = new c();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 13, EndLine: 1, EndColumn: 14}}},
		{Code: "var a = new b[ ( 'foo' ) ]();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 18, EndLine: 1, EndColumn: 23}}},
		{Code: "var a = new b[`foo`];", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 15, EndLine: 1, EndColumn: 20}}},
		{Code: "var a = b[`\\\nFoo`]();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 11, EndLine: 2, EndColumn: 5}}},
		{Code: "var x = Foo.Bar(42);", Options: map[string]any{"capIsNewExceptions": []any{"Foo"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 13, EndLine: 1, EndColumn: 16}}},
		{Code: "var x = Bar.Foo(42);", Options: map[string]any{"capIsNewExceptionPattern": "^Foo\\.."}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 13, EndLine: 1, EndColumn: 16}}},
		{Code: "var x = new foo.bar(42);", Options: map[string]any{"newIsCapExceptions": []any{"foo"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}}},
		{Code: "var x = new bar.foo(42);", Options: map[string]any{"newIsCapExceptionPattern": "^foo\\.."}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}}},

		// Optional chaining
		{Code: "new (foo?.bar)();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 11, EndLine: 1, EndColumn: 14}}},
		{Code: "foo?.Bar();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 6, EndLine: 1, EndColumn: 9}}},
		{Code: "(foo?.Bar)();", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 7, EndLine: 1, EndColumn: 10}}},
	})
}
