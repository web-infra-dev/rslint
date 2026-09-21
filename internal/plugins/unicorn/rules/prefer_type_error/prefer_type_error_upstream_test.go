// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-type-error.js
package prefer_type_error_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_type_error"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const preferTypeErrorMessage = "`new Error()` is too unspecific for a type check. Use `new TypeError()` instead."

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}}
}

func invalid(code string) rule_tester.InvalidTestCase {
	needle := "new Error"
	start := strings.Index(code, needle)
	if start < 0 {
		panic("prefer-type-error invalid fixture is missing new Error")
	}
	constructorStart := start + len("new ")
	prefix := code[:constructorStart]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := constructorStart + 1
	if lastNewline >= 0 {
		column = constructorStart - lastNewline
	}
	return rule_tester.InvalidTestCase{
		Code:            code,
		FileName:        "case.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output:          []string{strings.Replace(code, needle, "new TypeError", 1)},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "prefer-type-error",
			Message:   preferTypeErrorMessage,
			Line:      line, Column: column, EndLine: line, EndColumn: column + len("Error"),
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

func TestPreferTypeErrorUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid("if (MrFuManchu.name !== 'Fu Manchu' || MrFuManchu.isMale === false) {\n\tthrow new Error('How cant Fu Manchu be Fu Manchu?');\n}\n"),
		valid("if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (wrapper.g.ary.isArray(foo) || wrapper.f.g.ary.isView(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (wrapper.g.ary(foo) || wrapper.f.g.ary.isPiew(foo)) {\n\tthrow new Error();\n}\n"),
		valid("if (Array.isArray()) {\n\tthrow new Error('Woohoo - isArray is broken!');\n}\n"),
		valid("if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow new CustomError();\n}\n"),
		valid("if (Array.isArray(foo)) {\n\tthrow new Error.foo();\n}\n"),
		valid("if (Array.isArray(foo)) {\n\tthrow new Error.foo;\n}\n"),
		valid("if (Array.isArray(foo)) {\n\tthrow new foo.Error;\n}\n"),
		valid("if (Array.isArray(foo)) {\n\tthrow new foo.Error('My name is Foo Manchu');\n}\n"),
		valid("if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow Error('This is fo FooBar', foo);\n}\n"),
		valid("if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tnew Error('This is fo FooBar', foo);\n}\n"),
		valid("function test(foo) {\n\tif (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\t\treturn new Error('This is fo FooBar', foo);\n\t}\n\treturn foo;\n}\n"),
		valid("if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tlastError = new Error('This is fo FooBar', foo);\n}\n"),
		valid("if (!isFinite(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (isNaN(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (isArray(foo)) {\n\tthrow new Error();\n}\n"),
		valid("if (foo instanceof boo) {\n\tthrow new TypeError();\n}\n"),
		valid("if (typeof boo === 'Boo') {\n\tthrow new TypeError();\n}\n"),
		valid("if (typeof boo === 'Boo') {\n\tsome.thing.else.happens.before();\n\tthrow new Error();\n}\n"),
		valid("if (Number.isNaN(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (Number.isFinite(foo) && Number.isSafeInteger(foo) && Number.isInteger(foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (Array.isArray(foo) || (Blob.isBlob(foo) || Blip.isBlip(foo))) {\n\tthrow new TypeError();\n}\n"),
		valid("if (typeof foo === 'object' || (Object.isFrozen(foo) || 'String' === typeof foo)) {\n\tthrow new TypeError();\n}\n"),
		valid("if (isNaN) {\n\tthrow new Error();\n}\n"),
		valid("if (isObjectLike) {\n\tthrow new Error();\n}\n"),
		valid("if (isNaN.foo()) {\n\tthrow new Error();\n}\n"),
		valid("if (typeof foo !== 'object' || foo.bar() === false) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n"),
		valid("if (foo instanceof Foo) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n"),
		valid("if (!foo instanceof Foo) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n"),
		valid("if (foo instanceof Foo === false) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n"),
		valid("throw new Error('💣')"),
		valid("if (!Number.isNaN(foo) && foo === 10) {\n\tthrow new Error('foo is not 10!');\n}\n"),
		valid("function foo(foo) {\n\tif (!Number.isNaN(foo) && foo === 10) {\n\t\ttimesFooWas10 += 1;\n\t\tif (calculateAnswerToLife() !== 42) {\n\t\t\topenIssue('Your program is buggy!');\n\t\t} else {\n\t\t\treturn printAwesomeAnswer(42);\n\t\t}\n\t\tthrow new Error('foo is 10');\n\t}\n}\n"),
		valid("function foo(foo) {\n\tif (!Number.isNaN(foo)) {\n\t\ttimesFooWas10 += 1;\n\t\tif (calculateAnswerToLife({with: foo}) !== 42) {\n\t\t\topenIssue('Your program is buggy!');\n\t\t} else {\n\t\t\treturn printAwesomeAnswer(42);\n\t\t}\n\t\tthrow new Error('foo is 10');\n\t}\n}\n"),
		valid("if (!x.isFudge()) {\n\tthrow new Error('x is no fudge!');\n}\n"),
		valid("if (!_.isFudge(x)) {\n\tthrow new Error('x is no fudge!');\n}\n"),
		valid("switch (something) {\n\tcase 1:\n\t\tbreak;\n\tdefault:\n\t\tthrow new Error('Unknown');\n}\n"),
		valid("if (foo instanceof Error) throw new Error(\"message\")"),
		valid("if (foo instanceof CustomError) throw new Error(\"message\")"),
		valid("if (foo instanceof lib.Error) throw new Error(\"message\")"),
		valid("if (foo instanceof lib.CustomError) throw new Error(\"message\")"),
		valid("if (typeof window !== 'undefined') {\n\tthrow new Error('This package requires a browser environment.');\n}\n"),
		valid("if (typeof process === 'undefined') {\n\tthrow new Error('This package requires Node.js.');\n}\n"),
		valid("if ('undefined' === typeof self) {\n\tthrow new Error('No global available.');\n}\n"),
		valid("if (typeof window !== 'undefined' && typeof foo !== 'string') {\n\tthrow new Error('Mixed environment and type check.');\n}\n"),
		valid("if (typeof window != 'undefined') {\n\tthrow new Error('This package requires a browser environment.');\n}\n"),
		valid("if (typeof x !== 'undefined' || Array.isArray(x)) {\n\tthrow new Error('message');\n}\n"),
		valid("if (Array.isArray(foo) === false) {\n\tthrow new TypeError('Array expected');\n}"),
		valid("if (Number.isNaN(foo) === false && Number.isInteger(foo) === false) {\n\tthrow new TypeError('Integer expected');\n}"),
		valid("if (isNaN(foo) === false) {\n\tthrow new TypeError('Number expected');\n}"),
		valid("if (typeof foo !== 'function' &&\n\tfoo instanceof CookieMonster === false &&\n\tfoo instanceof Unicorn === false) {\n\tthrow new TypeError('Magic expected');\n}"),
	}
	invalidCases := []rule_tester.InvalidTestCase{
		invalid("if (!isFinite(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (Array.isArray(foo) || _.isString(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (isNaN(foo) === false) {\n\tthrow new Error();\n}\n"),
		invalid("if (Array.isArray(foo)) {\n\tthrow new Error('foo is an Array');\n}\n"),
		invalid("if (foo instanceof bar) {\n\tthrow new Error(foobar);\n}\n"),
		invalid("if (_.isElement(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (_.isElement(foo)) {\n\tthrow new Error;\n}\n"),
		invalid("if (wrapper._.isElement(foo)) {\n\tthrow new Error;\n}\n"),
		invalid("if (typeof foo == 'Foo' || 'Foo' === typeof foo) {\n\tthrow new Error();\n}\n"),
		invalid("if (typeof foo === 'string') {\n\tthrow new Error();\n}\n"),
		invalid("if (Number.isFinite(foo) && Number.isSafeInteger(foo) && Number.isInteger(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (wrapper.n.isFinite(foo) && wrapper.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (wrapper.f.g.n.isFinite(foo) && wrapper.g.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new Error();\n}\n"),
		invalid("if (SomeThing.isArguments(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isArray(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isArrayBuffer(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isArrayLike(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isArrayLikeObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isBigInt(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isBoolean(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isBuffer(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isDate(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isElement(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isError(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isFinite(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isFunction(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isInteger(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isLength(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isMap(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isNaN(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isNative(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isNil(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isNull(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isNumber(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isObjectLike(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isPlainObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isPrototypeOf(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isRegExp(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isSafeInteger(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isSet(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isString(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isSymbol(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isTypedArray(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isUndefined(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isView(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isWeakMap(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isWeakSet(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isWindow(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (SomeThing.isXMLDoc(foo) === bar) {\n\tthrow new Error('foo is bar');\n}"),
		invalid("if (Array.isArray(foo) === false) {\n\tthrow new Error('Array expected');\n}"),
		invalid("if (Number.isNaN(foo) === false && Number.isInteger(foo) === false) {\n\tthrow new Error('Integer expected');\n}"),
		invalid("if (isNaN(foo) === false) {\n\tthrow new Error('Number expected');\n}"),
		invalid("if (typeof foo !== 'function' &&\n\tfoo instanceof CookieMonster === false &&\n\tfoo instanceof Unicorn === false) {\n\tthrow new Error('Magic expected');\n}"),
	}
	if len(validCases) != 52 || len(invalidCases) != 54 {
		t.Fatalf("coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_type_error.PreferTypeErrorRule, validCases, invalidCases)
}

// Upstream also snapshots the first invalid source. It is already asserted above
// with exact ID, message, range, and fixed output; AGENTS.md asks us not to
// duplicate identical cases solely to satisfy a count.
