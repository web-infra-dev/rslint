package no_instanceof_builtins_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_instanceof_builtins"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDNoInstanceofBuiltins = "no-instanceof-builtins"
	messageIDSwitchToTypeOf       = "switch-to-type-of"
	messageNoInstanceofBuiltins   = "Avoid using `instanceof` for type checking as it can lead to unreliable results."
)

var looseStrategyInvalid = []string{
	// Primitive types
	"foo instanceof String",
	"foo instanceof Number",
	"foo instanceof Boolean",
	"foo instanceof BigInt",
	"foo instanceof Symbol",
	"foo instanceof Function",
	"foo instanceof Array",
}

var strictStrategyInvalid = []string{
	// Error types
	"foo instanceof Error",
	"foo instanceof EvalError",
	"foo instanceof RangeError",
	"foo instanceof ReferenceError",
	"foo instanceof SyntaxError",
	"foo instanceof TypeError",
	"foo instanceof URIError",
	"foo instanceof AggregateError",
	"foo instanceof SuppressedError",

	// Collection types
	"foo instanceof Map",
	"foo instanceof Set",
	"foo instanceof WeakMap",
	"foo instanceof WeakRef",
	"foo instanceof WeakSet",

	// Arrays and Typed Arrays
	"foo instanceof ArrayBuffer",
	"foo instanceof Int8Array",
	"foo instanceof Uint8Array",
	"foo instanceof Uint8ClampedArray",
	"foo instanceof Int16Array",
	"foo instanceof Uint16Array",
	"foo instanceof Int32Array",
	"foo instanceof Uint32Array",
	"foo instanceof Float16Array",
	"foo instanceof Float32Array",
	"foo instanceof Float64Array",
	"foo instanceof BigInt64Array",
	"foo instanceof BigUint64Array",

	// Data types
	"foo instanceof Object",

	// Regular Expressions
	"foo instanceof RegExp",

	// Async and functions
	"foo instanceof Promise",
	"foo instanceof Proxy",

	// Other
	"foo instanceof DataView",
	"foo instanceof Date",
	"foo instanceof SharedArrayBuffer",
	"foo instanceof FinalizationRegistry",
}

// TestNoInstanceofBuiltinsUpstream migrates the full valid/invalid suite from
// upstream test/no-instanceof-builtins.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_instanceof_builtins_extras_test.go file.
func TestNoInstanceofBuiltinsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Loose strategy valid ----
		jsValid("fooLoose instanceof WebWorker"),
	}
	for _, code := range strictStrategyInvalid {
		valid = append(valid, jsValid(strings.Replace(code, "foo", "fooLoose", 1)))
	}

	valid = append(valid,
		// ---- Exclude the specified constructors ----
		jsValidOptions("fooExclude instanceof Function", map[string]interface{}{"exclude": []interface{}{"Function"}}),
		jsValidOptions("fooExclude instanceof Array", map[string]interface{}{"exclude": []interface{}{"Array"}}),
		jsValidOptions("fooExclude instanceof String", map[string]interface{}{"exclude": []interface{}{"String"}}),

		// ---- Port from no-instanceof-array valid ----
		jsValid("Array.isArray(arr)"),
		jsValid("arr instanceof array"),
		jsValid("a instanceof 'array'"),
		jsValid("a instanceof ArrayA"),
		jsValid("a.x[2] instanceof foo()"),
		jsValid("Array.isArray([1,2,3]) === true"),
		jsValid(`"arr instanceof Array"`),
	)

	invalid := []rule_tester.InvalidTestCase{
		// ---- UseErrorIsError option with loose strategy ----
		errorIsErrorInvalid("fooErr instanceof Error", "fooErr instanceof Error", "fooErr", map[string]interface{}{"useErrorIsError": true, "strategy": "loose"}),
		errorIsErrorInvalid("(fooErr) instanceof (Error)", "(fooErr) instanceof (Error)", "(fooErr)", map[string]interface{}{"useErrorIsError": true, "strategy": "loose"}),
	}

	for _, code := range looseStrategyInvalid {
		invalid = append(invalid, upstreamInvalid(code, nil))
	}

	for _, code := range append([]string{}, looseStrategyInvalid...) {
		code = strings.Replace(code, "foo", "fooStrict", 1)
		invalid = append(invalid, upstreamInvalid(code, map[string]interface{}{"strategy": "strict"}))
	}
	for _, code := range strictStrategyInvalid {
		code = strings.Replace(code, "foo", "fooStrict", 1)
		invalid = append(invalid, upstreamInvalid(code, map[string]interface{}{"strategy": "strict"}))
	}

	for _, code := range []string{
		"err instanceof Error",
		"err instanceof EvalError",
		"err instanceof RangeError",
		"err instanceof ReferenceError",
		"err instanceof SyntaxError",
		"err instanceof TypeError",
		"err instanceof URIError",
		"err instanceof AggregateError",
		"err instanceof SuppressedError",
	} {
		invalid = append(invalid, upstreamInvalid(code, map[string]interface{}{"useErrorIsError": true, "strategy": "strict"}))
	}

	invalid = append(invalid,
		// ---- Include the specified constructors ----
		noFixInvalid("fooInclude instanceof WebWorker", "fooInclude instanceof WebWorker", map[string]interface{}{"include": []interface{}{"WebWorker"}}),
		noFixInvalid("fooInclude instanceof HTMLElement", "fooInclude instanceof HTMLElement", map[string]interface{}{"include": []interface{}{"HTMLElement"}}),

		// ---- Port from no-instanceof-array invalid ----
		arrayInvalid("arr instanceof Array", "arr instanceof Array", "arr"),
		arrayInvalid("[] instanceof Array", "[] instanceof Array", "[]"),
		arrayInvalid("[1,2,3] instanceof Array === true", "[1,2,3] instanceof Array", "[1,2,3]"),
		arrayInvalid("fun.call(1, 2, 3) instanceof Array", "fun.call(1, 2, 3) instanceof Array", "fun.call(1, 2, 3)"),
		arrayInvalid("obj.arr instanceof Array", "obj.arr instanceof Array", "obj.arr"),
		arrayInvalid("foo.bar[2] instanceof Array", "foo.bar[2] instanceof Array", "foo.bar[2]"),
		arrayInvalid("(0, array) instanceof Array", "(0, array) instanceof Array", "(0, array)"),
		arrayInvalidWithOutput(
			"function foo(){return[]instanceof Array}",
			"[]instanceof Array",
			"function foo(){return Array.isArray([])}",
		),
		arrayInvalidWithOutput(complexArrayCase(), complexArrayTarget(), complexArrayOutput()),

		// SKIP: rslint does not support vue-eslint-parser / Vue template expressions.
		rule_tester.InvalidTestCase{Code: `<template><div v-if="array instanceof Array" v-for="element of array"></div></template>`, FileName: "file.vue", Skip: true},
		// SKIP: rslint does not support vue-eslint-parser / Vue template expressions.
		rule_tester.InvalidTestCase{Code: `<template><div v-if="(( (( array )) instanceof (( Array )) ))" v-for="element of array"></div></template>`, FileName: "file.vue", Skip: true},
		// SKIP: rslint does not support vue-eslint-parser / Vue template expressions.
		rule_tester.InvalidTestCase{Code: `<template><div>{{(( (( array )) instanceof (( Array )) )) ? array.join(" | ") : array}}</div></template>`, FileName: "file.vue", Skip: true},
		// SKIP: rslint does not support vue-eslint-parser / Vue SFC script blocks.
		rule_tester.InvalidTestCase{Code: `<script>const foo = array instanceof Array</script>`, FileName: "file.vue", Skip: true},
		// SKIP: rslint does not support vue-eslint-parser / Vue SFC script blocks.
		rule_tester.InvalidTestCase{Code: `<script>const foo = (( (( array )) instanceof (( Array )) ))</script>`, FileName: "file.vue", Skip: true},
		// SKIP: rslint does not support vue-eslint-parser / Vue SFC script blocks.
		rule_tester.InvalidTestCase{Code: `<script>foo instanceof Function</script>`, FileName: "file.vue", Skip: true},
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_instanceof_builtins.NoInstanceofBuiltinsRule,
		valid,
		invalid,
	)
}

func upstreamInvalid(code string, options any) rule_tester.InvalidTestCase {
	target := code
	constructorName := constructorNameFromTarget(target)
	switch constructorName {
	case "Array":
		return arrayInvalidOptions(code, target, leftFromTarget(target), options)
	case "Function":
		return functionInvalidOptions(code, target, leftFromTarget(target), options)
	case "String", "Number", "Boolean", "BigInt", "Symbol":
		return primitiveInvalidOptions(code, target, leftFromTarget(target), constructorName, options)
	case "Error":
		if optsMap, ok := options.(map[string]interface{}); ok {
			if useErrorIsError, _ := optsMap["useErrorIsError"].(bool); useErrorIsError {
				return errorIsErrorInvalid(code, target, leftFromTarget(target), options)
			}
		}
		return noFixInvalid(code, target, options)
	default:
		return noFixInvalid(code, target, options)
	}
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func jsValidOptions(code string, options any) rule_tester.ValidTestCase {
	testCase := jsValid(code)
	testCase.Options = options
	return testCase
}

func noFixInvalid(code string, target string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Errors:   []rule_tester.InvalidTestCaseError{expectedError(code, target)},
	}
}

func arrayInvalid(code string, target string, left string) rule_tester.InvalidTestCase {
	return arrayInvalidOptions(code, target, left, nil)
}

func arrayInvalidOptions(code string, target string, left string, options any) rule_tester.InvalidTestCase {
	output := replaceFirst(code, target, "Array.isArray("+left+")")
	return arrayInvalidWithOutputOptions(code, target, output, options)
}

func arrayInvalidWithOutput(code string, target string, output string) rule_tester.InvalidTestCase {
	return arrayInvalidWithOutputOptions(code, target, output, nil)
}

func arrayInvalidWithOutputOptions(code string, target string, output string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Output:   []string{output},
		Errors:   []rule_tester.InvalidTestCaseError{expectedError(code, target)},
	}
}

func functionInvalidOptions(code string, target string, left string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Output:   []string{replaceFirst(code, target, "typeof "+left+" === 'function'")},
		Errors:   []rule_tester.InvalidTestCaseError{expectedError(code, target)},
	}
}

func errorIsErrorInvalid(code string, target string, left string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Output:   []string{replaceFirst(code, target, "Error.isError("+left+")")},
		Errors:   []rule_tester.InvalidTestCaseError{expectedError(code, target)},
	}
}

func primitiveInvalidOptions(code string, target string, left string, constructorName string, options any) rule_tester.InvalidTestCase {
	typeName := strings.ToLower(constructorName)
	err := expectedError(code, target)
	err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
		MessageId: messageIDSwitchToTypeOf,
		Output:    replaceFirst(code, target, "typeof "+left+" === '"+typeName+"'"),
	}}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Errors:   []rule_tester.InvalidTestCaseError{err},
	}
}

func expectedError(code string, target string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, target)
	if start < 0 {
		panic(fmt.Sprintf("target %q not found in %q", target, code))
	}
	end := start + len(target)
	line, column := lineColumnForOffset(code, start)
	endLine, endColumn := lineColumnForOffset(code, end)
	return rule_tester.InvalidTestCaseError{
		MessageId: messageIDNoInstanceofBuiltins,
		Message:   messageNoInstanceofBuiltins,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := 0; i < offset; i++ {
		switch code[i] {
		case '\n':
			line++
			column = 1
		case '\r':
			line++
			column = 1
			if i+1 < offset && code[i+1] == '\n' {
				i++
			}
		default:
			column++
		}
	}
	return line, column
}

func replaceFirst(code string, target string, replacement string) string {
	index := strings.Index(code, target)
	if index < 0 {
		panic(fmt.Sprintf("target %q not found in %q", target, code))
	}
	return code[:index] + replacement + code[index+len(target):]
}

func constructorNameFromTarget(target string) string {
	index := strings.LastIndex(target, "instanceof")
	if index < 0 {
		panic(fmt.Sprintf("instanceof not found in %q", target))
	}
	right := strings.TrimSpace(target[index+len("instanceof"):])
	right = strings.Trim(right, "()")
	return strings.TrimSpace(right)
}

func leftFromTarget(target string) string {
	index := strings.LastIndex(target, "instanceof")
	if index < 0 {
		panic(fmt.Sprintf("instanceof not found in %q", target))
	}
	return strings.TrimSpace(target[:index])
}

func complexArrayCase() string {
	return lines(
		"(",
		"\t// comment",
		"\t((",
		"\t\t// comment",
		"\t\t(",
		"\t\t\t// comment",
		"\t\t\tfoo",
		"\t\t\t// comment",
		"\t\t)",
		"\t\t// comment",
		"\t))",
		"\t// comment",
		")",
		"// comment before instanceof\r      instanceof",
		"",
		"// comment after instanceof",
		"",
		"(",
		"\t// comment",
		"",
		"\t(",
		"",
		"\t\t// comment",
		"",
		"\t\tArray",
		"",
		"\t\t// comment",
		"\t)",
		"",
		"\t\t// comment",
		")",
		"",
		"\t// comment",
	)
}

func complexArrayTarget() string {
	code := complexArrayCase()
	return strings.TrimSuffix(code, "\n\n\t// comment")
}

func complexArrayOutput() string {
	left := lines(
		"(",
		"\t// comment",
		"\t((",
		"\t\t// comment",
		"\t\t(",
		"\t\t\t// comment",
		"\t\t\tfoo",
		"\t\t\t// comment",
		"\t\t)",
		"\t\t// comment",
		"\t))",
		"\t// comment",
		")",
	)
	return lines(
		"Array.isArray("+left+")",
		"// comment before instanceof\r",
		"",
		"// comment after instanceof",
		"",
		"\t// comment",
		"",
		"",
		"\t\t// comment",
		"",
		"",
		"\t\t// comment",
		"",
		"",
		"\t\t// comment",
		"",
		"",
		"\t// comment",
	)
}

func lines(parts ...string) string {
	return strings.Join(parts, "\n")
}
