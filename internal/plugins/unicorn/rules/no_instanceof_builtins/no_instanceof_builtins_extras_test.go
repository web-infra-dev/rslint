package no_instanceof_builtins_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_instanceof_builtins"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoInstanceofBuiltinsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestNoInstanceofBuiltinsExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: optional/member wrappers on the left are just operands and do not crash ----
		jsValid("foo?.bar instanceof WebWorker"),
		// ---- Dimension 4: element access on the left is just an operand and does not affect matching ----
		jsValid("foo['bar'] instanceof WebWorker"),
		// ---- Dimension 4: dynamic right-hand member access is not an Identifier ----
		jsValid("foo instanceof constructors.Array"),
		// ---- Dimension 4: element right-hand access is not an Identifier ----
		jsValid("foo instanceof constructors['Array']"),
		// ---- Dimension 4: optional right-hand member access is not an Identifier ----
		jsValid("foo instanceof constructors?.Array"),
		// ---- Dimension 4: call right-hand side is not an Identifier ----
		jsValid("foo instanceof getConstructor()"),
		// ---- Dimension 4: array literal right-hand side is ignored without crashing ----
		jsValid("foo instanceof []"),
		// ---- Dimension 4: object literal right-hand side is ignored without crashing ----
		jsValid("foo instanceof {}"),
		// Locks in upstream create() arm 3: exclude returns before every reporting branch.
		jsValidOptions("foo instanceof Array", map[string]interface{}{"exclude": []interface{}{"Array"}}),
		// ---- Dimension 4: exclude still wins when strict strategy would otherwise report ----
		jsValidOptions("foo instanceof Map", []interface{}{map[string]interface{}{"strategy": "strict", "exclude": []interface{}{"Map"}}}),
		// ---- Dimension 4: malformed include option is ignored rather than widening matches ----
		jsValidOptions("foo instanceof WebWorker", []interface{}{map[string]interface{}{"include": "WebWorker"}}),
		// ---- Dimension 4: TS assertion on the right is not transparent to upstream's Identifier gate ----
		tsValid("foo instanceof (Array as any)"),
		// ---- Dimension 4: TS non-null on the right is not transparent to upstream's Identifier gate ----
		tsValid("foo instanceof (Array!)"),
		// ---- Dimension 4: TS satisfies on the right is not transparent to upstream's Identifier gate ----
		tsValid("foo instanceof (Array satisfies unknown)"),
		// N/A: access/key forms on object/class members do not apply; this rule only reads BinaryExpression.Right.
		// N/A: private identifiers are not valid right-hand operands for `instanceof`.
		// N/A: declaration/container forms only surround the binary expression; method and arrow bodies are covered below.
		// N/A: overload, abstract, and declare body-absent forms do not apply to binary expressions.

		// ---- Real-user: #2842 many unrelated binary expressions stay cheap and unreported ----
		jsValid(strings.Repeat("foo + bar;\n", 64) + "foo instanceof WebWorker"),

		// Locks in upstream create() arm 1: non-instanceof binary operators return early.
		jsValid("foo in Array"),
		// Locks in upstream create() arm 2: loose strategy does not report strict-only constructors.
		jsValid("foo instanceof Map"),
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized left and right expressions are transparent ----
		arrayInvalid("((foo)) instanceof ((Array))", "((foo)) instanceof ((Array))", "((foo))"),
		// ---- Dimension 4: TS non-null wrappers on the left are preserved by the Array autofix ----
		tsArrayInvalid("foo! instanceof Array", "foo! instanceof Array", "foo!"),
		// ---- Dimension 4: TS assertion wrappers on the left are preserved by the Function autofix ----
		tsFunctionInvalid("(foo as unknown) instanceof Function", "(foo as unknown) instanceof Function", "(foo as unknown)"),
		// ---- Dimension 4: TS satisfies wrappers on the left are preserved by the primitive suggestion ----
		tsPrimitiveInvalid("(foo satisfies unknown) instanceof String", "(foo satisfies unknown) instanceof String", "(foo satisfies unknown)", "String"),
		// ---- Dimension 4: optional property access on the left is preserved by the Array autofix ----
		arrayInvalid("foo?.bar instanceof Array", "foo?.bar instanceof Array", "foo?.bar"),
		// ---- Dimension 4: optional call on the left is preserved by the Function autofix ----
		functionInvalidOptions("foo?.() instanceof Function", "foo?.() instanceof Function", "foo?.()", nil),
		// ---- Dimension 3: comments around Array checks are preserved like upstream's token-level fix ----
		arrayInvalidWithOutput(
			"foo /*a*/ instanceof /*b*/ Array",
			"foo /*a*/ instanceof /*b*/ Array",
			"Array.isArray(foo) /*a*/ /*b*/",
		),
		// ---- Dimension 3: comments around Error checks are preserved when useErrorIsError fixes to a call ----
		invalidWithOutputOptions(
			"foo /*a*/ instanceof /*b*/ Error",
			"foo /*a*/ instanceof /*b*/ Error",
			"Error.isError(foo) /*a*/ /*b*/",
			map[string]interface{}{"useErrorIsError": true},
		),
		// ---- Dimension 3: comments around Function checks stay in place around the rewritten operator ----
		invalidWithOutputOptions(
			"foo /*a*/ instanceof /*b*/ Function",
			"foo /*a*/ instanceof /*b*/ Function",
			"typeof foo /*a*/ === /*b*/ 'function'",
			nil,
		),
		// ---- Dimension 3: primitive wrapper suggestions preserve comments around the operator ----
		primitiveInvalidWithSuggestionOutput(
			"foo /*a*/ instanceof /*b*/ String",
			"foo /*a*/ instanceof /*b*/ String",
			"typeof foo /*a*/ === /*b*/ 'string'",
		),
		// ---- Dimension 4: multi-line parenthesized left expression is preserved by the Function autofix ----
		functionInvalidOptions(
			lines(
				"const result = (",
				"\tfoo",
				") instanceof Function;",
			),
			lines(
				"(",
				"\tfoo",
				") instanceof Function",
			),
			lines(
				"(",
				"\tfoo",
				")",
			),
			nil,
		),
		// ---- Dimension 4: await expressions on the left stay grouped inside the Array autofix ----
		arrayInvalid("async function test() { return await values instanceof Array; }", "await values instanceof Array", "await values"),
		// ---- Dimension 4: method-body traversal reports nested instance checks ----
		arrayInvalid("class C { method() { return this.value instanceof Array; } }", "this.value instanceof Array", "this.value"),
		// ---- Dimension 4: constructor bodies are traversed ----
		arrayInvalid("class C { constructor() { this.value instanceof Array; } }", "this.value instanceof Array", "this.value"),
		// ---- Dimension 4: getter bodies are traversed ----
		arrayInvalid("class C { get value() { return items instanceof Array; } }", "items instanceof Array", "items"),
		// ---- Dimension 4: setter bodies are traversed ----
		arrayInvalid("class C { set value(items) { items instanceof Array; } }", "items instanceof Array", "items"),
		// ---- Dimension 4: arrow-body traversal reports nested instance checks ----
		arrayInvalid("const check = value => value instanceof Array;", "value instanceof Array", "value"),
		// ---- Dimension 4: function-expression bodies are traversed ----
		arrayInvalid("const check = function (value) { return value instanceof Array; };", "value instanceof Array", "value"),
		// ---- Dimension 4: generator bodies are traversed ----
		arrayInvalid("function* check() { return value instanceof Array; }", "value instanceof Array", "value"),
		// ---- Dimension 4: async generator bodies are traversed ----
		arrayInvalid("async function* check() { return value instanceof Array; }", "value instanceof Array", "value"),
		// ---- Dimension 4: class-field arrow bodies are traversed ----
		arrayInvalid("class C { check = value => value instanceof Array; }", "value instanceof Array", "value"),
		// ---- Dimension 4: class static blocks are traversed ----
		arrayInvalid("class C { static { value instanceof Array; } }", "value instanceof Array", "value"),
		// ---- Dimension 4: class extends clauses are traversed ----
		functionInvalidOptions("class C extends (base instanceof Function ? Base : Object) {}", "base instanceof Function", "base", nil),
		// ---- Dimension 4: computed class keys are traversed ----
		arrayInvalid("class C { [value instanceof Array]() {} }", "value instanceof Array", "value"),
		// ---- Dimension 4: computed object keys are traversed ----
		arrayInvalid("const object = { [value instanceof Array]: true };", "value instanceof Array", "value"),
		// ---- Dimension 4: multiple sibling reports keep source order and apply non-overlapping autofixes together ----
		multiInvalidWithOutput(
			"const result = foo instanceof Array || bar instanceof Function;",
			"const result = Array.isArray(foo) || typeof bar === 'function';",
			"foo instanceof Array",
			"bar instanceof Function",
		),

		// ---- Real-user: #2452 proposal examples report under strict strategy ----
		noFixInvalid("foo instanceof Map", "foo instanceof Map", map[string]interface{}{"strategy": "strict"}),
		noFixInvalid("foo instanceof Promise", "foo instanceof Promise", map[string]interface{}{"strategy": "strict"}),
		// ---- Real-user: #2537 unsafe primitive wrapper conversion remains a suggestion only ----
		primitiveInvalidOptions("new Number instanceof Number", "new Number instanceof Number", "new Number", "Number", nil),
		// ---- Real-user: #2157 Function checks get the safe typeof autofix ----
		functionInvalidOptions("foo instanceof Function", "foo instanceof Function", "foo", nil),

		// Locks in upstream create() arm 7: include reports custom constructors with no fix.
		noFixInvalid("foo instanceof HTMLElement", "foo instanceof HTMLElement", map[string]interface{}{"include": []interface{}{"HTMLElement"}}),
		// Locks in upstream create() arm 4: Array uses an autofix function call.
		arrayInvalid("foo instanceof Array", "foo instanceof Array", "foo"),
		// Locks in upstream create() arm 5: Error + useErrorIsError uses Error.isError even in loose mode.
		errorIsErrorInvalid("foo instanceof Error", "foo instanceof Error", "foo", map[string]interface{}{"useErrorIsError": true}),
		// Locks in upstream create() arm 6: primitive wrappers report with a suggestion, not an autofix.
		primitiveInvalidOptions("foo instanceof String", "foo instanceof String", "foo", "String", nil),
		// Locks in upstream create() arm 8: include also reports custom constructors outside the DOM family.
		noFixInvalid("foo instanceof WebWorker", "foo instanceof WebWorker", map[string]interface{}{"include": []interface{}{"WebWorker"}}),
		// Locks in option parsing: array-wrapped rule options use the same path as JS tests.
		noFixInvalid("foo instanceof WebWorker", "foo instanceof WebWorker", []interface{}{map[string]interface{}{"include": []interface{}{"WebWorker"}}}),
		// Locks in strict strategy: strict-only constructors report without fixes.
		noFixInvalid("new (class {})() instanceof Object", "new (class {})() instanceof Object", map[string]interface{}{"strategy": "strict"}),
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_instanceof_builtins.NoInstanceofBuiltinsRule,
		valid,
		invalid,
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func tsArrayInvalid(code string, target string, left string) rule_tester.InvalidTestCase {
	testCase := arrayInvalid(code, target, left)
	testCase.FileName = "file.ts"
	return testCase
}

func tsFunctionInvalid(code string, target string, left string) rule_tester.InvalidTestCase {
	testCase := functionInvalidOptions(code, target, left, nil)
	testCase.FileName = "file.ts"
	return testCase
}

func tsPrimitiveInvalid(code string, target string, left string, constructorName string) rule_tester.InvalidTestCase {
	testCase := primitiveInvalidOptions(code, target, left, constructorName, nil)
	testCase.FileName = "file.ts"
	return testCase
}

func invalidWithOutputOptions(code string, target string, output string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Output:   []string{output},
		Errors:   []rule_tester.InvalidTestCaseError{expectedError(code, target)},
	}
}

func primitiveInvalidWithSuggestionOutput(code string, target string, output string) rule_tester.InvalidTestCase {
	err := expectedError(code, target)
	err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
		MessageId: messageIDSwitchToTypeOf,
		Output:    output,
	}}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors:   []rule_tester.InvalidTestCaseError{err},
	}
}

func multiInvalidWithOutput(code string, output string, targets ...string) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(targets))
	for _, target := range targets {
		errors = append(errors, expectedError(code, target))
	}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors:   errors,
	}
}
