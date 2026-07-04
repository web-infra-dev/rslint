package new_for_builtins_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/new_for_builtins"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDEnforce           = "enforce"
	messageIDDisallow          = "disallow"
	messageIDDisallowCallOrNew = "disallowCallOrNew"
	messageIDErrorDate         = "error-date"
	messageIDSuggestionDate    = "suggestion-date"
	messageDate                = "Use `String(new Date())` instead of `Date()`."
	messageDateSuggestion      = "Switch to `String(new Date())`."
)

// TestNewForBuiltinsUpstream migrates the full valid/invalid suite from
// upstream test/new-for-builtins.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// new_for_builtins_extras_test.go file.
func TestNewForBuiltinsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- test.snapshot valid ----
		jsValid("const foo = new Object()"),
		jsValid("const foo = new Array()"),
		// Optional calls can't become `new` expressions.
		jsValid("const foo = Array?.()"),
		jsValid("const foo = Map?.()"),
		jsValid("const foo = Date?.()"),
		jsValid("const foo = globalThis?.Date()"),
		jsValid("const foo = Intl.DateTimeFormat?.()"),
		jsValid("const foo = Intl?.DateTimeFormat()"),
		jsValid("const foo = Temporal.PlainDate?.(2024, 1, 1)"),
		jsValid("const foo = WebAssembly.Module?.(buffer)"),
		jsValid("const foo = WebAssembly?.Module(buffer)"),
	}

	for _, name := range []string{
		"ArrayBuffer", "BigInt64Array", "BigUint64Array", "DataView", "Error",
		"Float16Array", "Float32Array", "Float64Array", "Function", "Int8Array",
		"Int16Array", "Int32Array", "Map", "WeakMap", "Set", "WeakSet",
		"Promise", "RegExp", "Uint8Array", "Uint16Array", "Uint32Array",
		"Uint8ClampedArray", "AggregateError", "TypeError", "SuppressedError",
		"DisposableStack", "AsyncDisposableStack",
	} {
		valid = append(valid, jsValid(fmt.Sprintf("const foo = new %s()", name)))
	}
	valid = append(valid,
		jsValid("const foo = new Map([['foo', 'bar'], ['unicorn', 'rainbow']])"),
		jsValid("const foo = new AggregateError([])"),
		jsValid("const foo = new SuppressedError(error, suppressed)"),
		jsValid("const foo = BigInt()"),
		jsValid("const foo = Boolean()"),
		jsValid("const foo = Number()"),
		jsValid("const foo = String()"),
		jsValid("const foo = Symbol()"),
		jsValid("const foo = new Intl.DateTimeFormat()"),
		jsValid("const foo = new globalThis.Intl.DateTimeFormat()"),
		jsValid("const foo = new Intl.DisplayNames('en', {type: 'language'})"),
		jsValid("const foo = new Intl.Locale('en')"),
		jsValid("const foo = new Intl.Segmenter()"),
		jsValid("const foo = new Temporal.PlainDate(2024, 1, 1)"),
		jsValid("const foo = new globalThis.Temporal.PlainDate(2024, 1, 1)"),
		jsValid("const foo = new Temporal.ZonedDateTime(0n, 'UTC')"),
		jsValid("const foo = Temporal.Now.instant()"),
		jsValid("const foo = new WebAssembly.Module(buffer)"),
		jsValid("const foo = new globalThis.WebAssembly.Module(buffer)"),
		jsValid("const foo = new WebAssembly.Memory({initial: 1})"),
		jsValid("const foo = new WebAssembly.CompileError()"),
		jsValid("const foo = WebAssembly.instantiate(buffer)"),
		jsValid("const foo = WebAssembly.JSTag"),
	)

	shadowedCallObjects := append([]string{}, enforceNewObjects...)
	shadowedCallObjects = append(shadowedCallObjects, disallowCallOrNewObjects...)
	shadowedNewObjects := append([]string{}, disallowNewObjects...)
	shadowedNewObjects = append(shadowedNewObjects, disallowCallOrNewObjects...)
	for _, object := range shadowedCallObjects {
		valid = append(valid,
			jsValid(createShadowedCallTest(object)),
			jsValid(createNestedShadowedCallTest(object)),
		)
	}
	for _, object := range shadowedNewObjects {
		valid = append(valid,
			jsValid(createShadowedNewTest(object)),
			jsValid(createNestedShadowedNewTest(object)),
		)
	}

	valid = append(valid,
		// #122
		jsValid(lines(
			"import { Map } from 'immutable';",
			"const m = Map();",
		)),
		jsValid(lines(
			"const {Map} = require('immutable');",
			"const foo = Map();",
		)),
		jsValid(lines(
			"const {String} = require('guitar');",
			"const lowE = new String();",
		)),
		jsValid(lines(
			"import {String} from 'guitar';",
			"const lowE = new String();",
		)),
		// Not builtin
		jsValid("new Foo();Bar();"),
		jsValid("Foo();new Bar();"),
		// Ignored
		jsValid("const isObject = v => Object(v) === v;"),
		jsValid("const isObject = v => globalThis.Object(v) === v;"),
		jsValid("(x) !== Object(x)"),
		// SKIP: rslint does not support ESLint's languageOptions.globals override.
		rule_tester.ValidTestCase{Code: `new Symbol("")`, Skip: true},
	)

	invalid := []rule_tester.InvalidTestCase{
		// ---- test.snapshot invalid ----
		enforceInvalid("const object = (Object)();", "(Object)()", "Object"),
		enforceInvalid("const isObject = v => Object(v) == v;", "Object(v)", "Object"),
		disallowInvalid(`const symbol = new (Symbol)("");`, `new (Symbol)("")`, "Symbol", `const symbol = (Symbol)("");`),
		disallowInvalid(`const symbol = new /* comment */ Symbol("");`, `new /* comment */ Symbol("")`, "Symbol", `const symbol = /* comment */ Symbol("");`),
		disallowInvalid("const symbol = new Symbol;", "new Symbol", "Symbol", "const symbol = Symbol();"),
		tsDisallowInvalid("const s = new Symbol()!;", "new Symbol()", "Symbol", "const s = Symbol()!;"),
		disallowInvalid(lines(
			"() => {",
			"\treturn new // 1",
			"\t\tSymbol();",
			"}",
		), "new // 1\n\t\tSymbol()", "Symbol", lines(
			"() => {",
			"\treturn ( // 1",
			"\t\tSymbol());",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn (",
			"\t\tnew // 2",
			"\t\t\tSymbol()",
			"\t);",
			"}",
		), "new // 2\n\t\t\tSymbol()", "Symbol", lines(
			"() => {",
			"\treturn (",
			"\t\t// 2",
			"\t\t\tSymbol()",
			"\t);",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn new // 3",
			"\t\t(Symbol);",
			"}",
		), "new // 3\n\t\t(Symbol)", "Symbol", lines(
			"() => {",
			"\treturn ( // 3",
			"\t\t(Symbol)());",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn new // 4",
			"\t\tSymbol;",
			"}",
		), "new // 4\n\t\tSymbol", "Symbol", lines(
			"() => {",
			"\treturn ( // 4",
			"\t\tSymbol());",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn (",
			"\t\tnew // 5",
			"\t\t\tSymbol",
			"\t);",
			"}",
		), "new // 5\n\t\t\tSymbol", "Symbol", lines(
			"() => {",
			"\treturn (",
			"\t\t// 5",
			"\t\t\tSymbol()",
			"\t);",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn (",
			"\t\tnew // 6",
			"\t\t\t(Symbol)",
			"\t);",
			"}",
		), "new // 6\n\t\t\t(Symbol)", "Symbol", lines(
			"() => {",
			"\treturn (",
			"\t\t// 6",
			"\t\t\t(Symbol)()",
			"\t);",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\tthrow new // 1",
			"\t\tSymbol();",
			"}",
		), "new // 1\n\t\tSymbol()", "Symbol", lines(
			"() => {",
			"\tthrow ( // 1",
			"\t\tSymbol());",
			"}",
		)),
		disallowInvalid(lines(
			"() => {",
			"\treturn new /**/ Symbol;",
			"}",
		), "new /**/ Symbol", "Symbol", lines(
			"() => {",
			"\treturn /**/ Symbol();",
			"}",
		)),
		disallowNoFixInvalid("new globalThis.String()", "new globalThis.String()", "String"),
		disallowNoFixInvalid("new global.String()", "new global.String()", "String"),
		disallowNoFixInvalid("new self.String()", "new self.String()", "String"),
		disallowNoFixInvalid("new window.String()", "new window.String()", "String"),
		disallowNoFixInvalid(lines(
			"const {String} = globalThis;",
			"new String();",
		), "new String()", "String"),
		disallowNoFixInvalid(lines(
			"const {String: RenamedString} = globalThis;",
			"new RenamedString();",
		), "new RenamedString()", "String"),
		disallowNoFixInvalid(lines(
			"const RenamedString = globalThis.String;",
			"new RenamedString();",
		), "new RenamedString()", "String"),
		enforceInvalid("globalThis.Array()", "globalThis.Array()", "Array"),
		enforceInvalid("global.Array()", "global.Array()", "Array"),
		enforceInvalid("self.Array()", "self.Array()", "Array"),
		enforceInvalid("window.Array()", "window.Array()", "Array"),
		enforceInvalid(lines(
			"const {Array: RenamedArray} = globalThis;",
			"RenamedArray();",
		), "RenamedArray()", "Array"),
		// SKIP: rslint does not support ESLint's languageOptions.globals override.
		{Code: "globalThis.Array()", Skip: true},
		// SKIP: rslint does not support ESLint's languageOptions.globals override.
		{Code: lines("const {Array} = globalThis;", "Array();"), Skip: true},
	}

	for _, name := range []string{
		"Object", "Array", "ArrayBuffer", "BigInt64Array", "BigUint64Array",
		"DataView", "Error", "AggregateError", "EvalError", "RangeError",
		"ReferenceError", "SuppressedError", "SyntaxError", "TypeError", "URIError",
		"DisposableStack", "AsyncDisposableStack", "Float16Array", "Float32Array",
		"Float64Array", "Function", "Int8Array", "Int16Array", "Int32Array",
		"WeakMap", "Set", "WeakSet", "Promise", "RegExp", "Uint8Array",
		"Uint16Array", "Uint32Array", "Uint8ClampedArray", "Intl.Collator",
		"Intl.DateTimeFormat", "Intl.DurationFormat", "Intl.ListFormat",
		"Intl.NumberFormat", "Intl.PluralRules", "Intl.RelativeTimeFormat",
		"Intl.Segmenter", "Temporal.Duration", "Temporal.Instant",
		"Temporal.PlainDateTime", "Temporal.PlainMonthDay", "Temporal.PlainTime",
		"Temporal.PlainYearMonth", "Temporal.ZonedDateTime", "WebAssembly.Module",
		"WebAssembly.Instance", "WebAssembly.Memory", "WebAssembly.Table",
		"WebAssembly.Global", "WebAssembly.Tag", "WebAssembly.Exception",
		"WebAssembly.CompileError", "WebAssembly.LinkError", "WebAssembly.RuntimeError",
	} {
		code := fmt.Sprintf("const foo = %s()", name)
		invalid = append(invalid, enforceInvalid(code, name+"()", name))
	}
	invalid = append(invalid,
		enforceInvalid("const foo = Error('Foo bar')", "Error('Foo bar')", "Error"),
		enforceInvalid("const foo = AggregateError([])", "AggregateError([])", "AggregateError"),
		enforceInvalid("const foo = SuppressedError(error, suppressed)", "SuppressedError(error, suppressed)", "SuppressedError"),
		enforceInvalid("const foo = (( Map ))()", "(( Map ))()", "Map"),
		enforceInvalid("const foo = Map([['foo', 'bar'], ['unicorn', 'rainbow']])", "Map([['foo', 'bar'], ['unicorn', 'rainbow']])", "Map"),
		enforceInvalid("const foo = Intl.DisplayNames('en', {type: 'language'})", "Intl.DisplayNames('en', {type: 'language'})", "Intl.DisplayNames"),
		enforceInvalid("const foo = Intl.Locale('en')", "Intl.Locale('en')", "Intl.Locale"),
		enforceInvalid("const foo = Temporal.Instant(0n)", "Temporal.Instant(0n)", "Temporal.Instant"),
		enforceInvalid("const foo = Temporal.PlainDate(2024, 1, 1)", "Temporal.PlainDate(2024, 1, 1)", "Temporal.PlainDate"),
		enforceInvalid("const foo = globalThis.Temporal.PlainDate(2024, 1, 1)", "globalThis.Temporal.PlainDate(2024, 1, 1)", "Temporal.PlainDate"),
		disallowCallOrNewInvalid("const foo = Temporal.Now()", "Temporal.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = Temporal.Now?.()", "Temporal.Now?.()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = Temporal?.Now()", "Temporal?.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = new Temporal.Now()", "new Temporal.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = globalThis.Temporal.Now()", "globalThis.Temporal.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = globalThis.Temporal?.Now()", "globalThis.Temporal?.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = globalThis?.Temporal.Now()", "globalThis?.Temporal.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = new globalThis.Temporal.Now()", "new globalThis.Temporal.Now()", "Temporal.Now"),
		disallowCallOrNewInvalid("const foo = WebAssembly()", "WebAssembly()", "WebAssembly"),
		disallowCallOrNewInvalid("const foo = new WebAssembly()", "new WebAssembly()", "WebAssembly"),
		disallowCallOrNewInvalid("const foo = globalThis.WebAssembly()", "globalThis.WebAssembly()", "WebAssembly"),
		disallowCallOrNewInvalid("const foo = new globalThis.WebAssembly()", "new globalThis.WebAssembly()", "WebAssembly"),
		disallowCallOrNewInvalid("const foo = WebAssembly.JSTag()", "WebAssembly.JSTag()", "WebAssembly.JSTag"),
		disallowCallOrNewInvalid("const foo = new WebAssembly.JSTag()", "new WebAssembly.JSTag()", "WebAssembly.JSTag"),
		disallowCallOrNewInvalid("const foo = globalThis.WebAssembly.JSTag()", "globalThis.WebAssembly.JSTag()", "WebAssembly.JSTag"),
		disallowCallOrNewInvalid("const foo = globalThis.WebAssembly?.JSTag()", "globalThis.WebAssembly?.JSTag()", "WebAssembly.JSTag"),
		disallowCallOrNewInvalid("const foo = new globalThis.WebAssembly.JSTag()", "new globalThis.WebAssembly.JSTag()", "WebAssembly.JSTag"),
		enforceInvalid("const foo = WebAssembly.Module(buffer)", "WebAssembly.Module(buffer)", "WebAssembly.Module"),
		enforceInvalid("const foo = globalThis.WebAssembly.Module(buffer)", "globalThis.WebAssembly.Module(buffer)", "WebAssembly.Module"),
		enforceInvalid("const foo = WebAssembly.Instance(module, imports)", "WebAssembly.Instance(module, imports)", "WebAssembly.Instance"),
		enforceInvalid("const foo = WebAssembly.Table({initial: 1, element: 'anyfunc'})", "WebAssembly.Table({initial: 1, element: 'anyfunc'})", "WebAssembly.Table"),
		enforceInvalid("const foo = WebAssembly.Global({value: 'i32', mutable: true}, 0)", "WebAssembly.Global({value: 'i32', mutable: true}, 0)", "WebAssembly.Global"),
		enforceInvalid("const foo = WebAssembly.Tag({parameters: ['i32']})", "WebAssembly.Tag({parameters: ['i32']})", "WebAssembly.Tag"),
		enforceInvalid("const foo = WebAssembly.Exception(tag, [1])", "WebAssembly.Exception(tag, [1])", "WebAssembly.Exception"),
		disallowInvalid("const foo = new BigInt(123)", "new BigInt(123)", "BigInt", "const foo = BigInt(123)"),
		disallowNoFixInvalid("const foo = new Boolean()", "new Boolean()", "Boolean"),
		disallowNoFixInvalid("const foo = new Number()", "new Number()", "Number"),
		disallowNoFixInvalid("const foo = new Number('123')", "new Number('123')", "Number"),
		disallowNoFixInvalid("const foo = new String()", "new String()", "String"),
		disallowInvalid("const foo = new Symbol()", "new Symbol()", "Symbol", "const foo = Symbol()"),
	)

	blockShadowCode := lines(
		"function varCheck() {",
		"\t{",
		"\t\tvar WeakMap = function() {};",
		"\t}",
		"\t// This should not be reported",
		"\treturn WeakMap()",
		"}",
		"function constCheck() {",
		"\t{",
		"\t\tconst Array = function() {};",
		"\t}",
		"\treturn Array()",
		"}",
		"function letCheck() {",
		"\t{",
		"\t\tlet Map = function() {};",
		"\t}",
		"\treturn Map()",
		"}",
	)
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   blockShadowCode,
		Output: []string{strings.Replace(strings.Replace(blockShadowCode, "return Array()", "return new Array()", 1), "return Map()", "return new Map()", 1)},
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(blockShadowCode, "Array()", messageIDEnforce, enforceMessage("Array")),
			expectedErrorAfter(blockShadowCode, "function letCheck()", "Map()", messageIDEnforce, enforceMessage("Map")),
		},
	})
	invalid = append(invalid,
		enforceInvalid(lines(
			"function foo() {",
			"\treturn(globalThis).Map()",
			"}",
		), "(globalThis).Map()", "Map"),
	)

	invalid = append(invalid,
		// ---- `Date` invalid ----
		dateFixInvalid("const foo = Date();", "Date()", "const foo = String(new Date());"),
		dateFixInvalid("const foo = globalThis.Date();", "globalThis.Date()", "const foo = String(new Date());"),
		dateFixInvalid(lines(
			"function foo() {",
			"\treturn(globalThis).Date();",
			"}",
		), "(globalThis).Date()", lines(
			"function foo() {",
			"\treturn String(new Date());",
			"}",
		)),
		dateSuggestionInvalid("const foo = Date(/*comment*/);", "Date(/*comment*/)", "const foo = String(new Date());"),
		dateSuggestionInvalid("const foo = globalThis/*comment*/.Date();", "globalThis/*comment*/.Date()", "const foo = String(new Date());"),
		dateSuggestionInvalid("const foo = Date(bar);", "Date(bar)", "const foo = String(new Date());"),
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&new_for_builtins.NewForBuiltinsRule,
		valid,
		invalid,
	)
}

var enforceNewObjects = []string{
	"Object", "Array", "ArrayBuffer", "DataView", "Date", "Function", "Map",
	"WeakMap", "Set", "WeakSet", "Promise", "RegExp", "SharedArrayBuffer",
	"Proxy", "WeakRef", "FinalizationRegistry", "DisposableStack",
	"AsyncDisposableStack", "Error", "AggregateError", "EvalError", "RangeError",
	"ReferenceError", "SuppressedError", "SyntaxError", "TypeError", "URIError",
	"Intl.Collator", "Intl.DateTimeFormat", "Intl.DisplayNames",
	"Intl.DurationFormat", "Intl.ListFormat", "Intl.Locale", "Intl.NumberFormat",
	"Intl.PluralRules", "Intl.RelativeTimeFormat", "Intl.Segmenter",
	"Temporal.Duration", "Temporal.Instant", "Temporal.PlainDate",
	"Temporal.PlainDateTime", "Temporal.PlainMonthDay", "Temporal.PlainTime",
	"Temporal.PlainYearMonth", "Temporal.ZonedDateTime", "WebAssembly.Module",
	"WebAssembly.Instance", "WebAssembly.Memory", "WebAssembly.Table",
	"WebAssembly.Global", "WebAssembly.Tag", "WebAssembly.Exception",
	"WebAssembly.CompileError", "WebAssembly.LinkError", "WebAssembly.RuntimeError",
	"Float16Array", "Float32Array", "Float64Array", "Int8Array", "Int16Array",
	"Int32Array", "BigInt64Array", "BigUint64Array", "Uint8Array", "Uint16Array",
	"Uint32Array", "Uint8ClampedArray",
}

var disallowNewObjects = []string{"BigInt", "Boolean", "Number", "String", "Symbol"}
var disallowCallOrNewObjects = []string{"Temporal.Now", "WebAssembly", "WebAssembly.JSTag"}

func createShadowedCallTest(object string) string {
	objectName, propertyName, ok := strings.Cut(object, ".")
	if !ok {
		return lines(
			fmt.Sprintf("const %s = function() {};", object),
			fmt.Sprintf("const foo = %s();", object),
		)
	}
	return lines(
		fmt.Sprintf("const %s = {%s() {}};", objectName, propertyName),
		fmt.Sprintf("const foo = %s();", object),
	)
}

func createShadowedNewTest(object string) string {
	objectName, propertyName, ok := strings.Cut(object, ".")
	if !ok {
		return lines(
			fmt.Sprintf("const %s = function() {};", object),
			fmt.Sprintf("const foo = new %s();", object),
		)
	}
	return lines(
		fmt.Sprintf("const %s = {%s: class {}};", objectName, propertyName),
		fmt.Sprintf("const foo = new %s();", object),
	)
}

func createNestedShadowedCallTest(object string) string {
	objectName, propertyName, ok := strings.Cut(object, ".")
	if !ok {
		return lines(
			"function outer() {",
			fmt.Sprintf("\tconst %s = function() {};", object),
			"\tfunction inner() {",
			fmt.Sprintf("\t\tconst foo = %s();", object),
			"\t}",
			"}",
		)
	}
	return lines(
		"function outer() {",
		fmt.Sprintf("\tconst %s = {%s() {}};", objectName, propertyName),
		"\tfunction inner() {",
		fmt.Sprintf("\t\tconst foo = %s();", object),
		"\t}",
		"}",
	)
}

func createNestedShadowedNewTest(object string) string {
	objectName, propertyName, ok := strings.Cut(object, ".")
	if !ok {
		return lines(
			"function insideFunction() {",
			fmt.Sprintf("\tconst %s = function() {};", object),
			"\tfunction inner() {",
			fmt.Sprintf("\t\tconst foo = new %s();", object),
			"\t}",
			"}",
		)
	}
	return lines(
		"function insideFunction() {",
		fmt.Sprintf("\tconst %s = {%s: class {}};", objectName, propertyName),
		"\tfunction inner() {",
		fmt.Sprintf("\t\tconst foo = new %s();", object),
		"\t}",
		"}",
	)
}

func lines(parts ...string) string {
	return strings.Join(parts, "\n")
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func enforceInvalid(code string, target string, name string) rule_tester.InvalidTestCase {
	output := insertBeforeFirst(code, target, enforceFixText(code, target))
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageIDEnforce, enforceMessage(name)),
		},
	}
}

func disallowInvalid(code string, target string, name string, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageIDDisallow, disallowMessage(name)),
		},
	}
}

func tsDisallowInvalid(code string, target string, name string, output string) rule_tester.InvalidTestCase {
	testCase := disallowInvalid(code, target, name, output)
	testCase.FileName = "file.ts"
	return testCase
}

func disallowNoFixInvalid(code string, target string, name string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageIDDisallow, disallowMessage(name)),
		},
	}
}

func disallowCallOrNewInvalid(code string, target string, name string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageIDDisallowCallOrNew, disallowCallOrNewMessage(name)),
		},
	}
}

func dateFixInvalid(code string, target string, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageIDErrorDate, messageDate),
		},
	}
}

func dateSuggestionInvalid(code string, target string, output string) rule_tester.InvalidTestCase {
	err := expectedError(code, target, messageIDErrorDate, messageDate)
	err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
		MessageId: messageIDSuggestionDate,
		Output:    output,
	}}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors:   []rule_tester.InvalidTestCaseError{err},
	}
}

func expectedError(code string, target string, messageID string, message string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, target)
	if start < 0 {
		panic(fmt.Sprintf("target %q not found in %q", target, code))
	}
	end := start + len(target)
	line, column := lineColumnForOffset(code, start)
	endLine, endColumn := lineColumnForOffset(code, end)
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func expectedErrorAfter(code string, anchor string, target string, messageID string, message string) rule_tester.InvalidTestCaseError {
	anchorIndex := strings.Index(code, anchor)
	if anchorIndex < 0 {
		panic(fmt.Sprintf("anchor %q not found in %q", anchor, code))
	}
	relativeIndex := strings.Index(code[anchorIndex:], target)
	if relativeIndex < 0 {
		panic(fmt.Sprintf("target %q not found after %q in %q", target, anchor, code))
	}
	start := anchorIndex + relativeIndex
	end := start + len(target)
	line, column := lineColumnForOffset(code, start)
	endLine, endColumn := lineColumnForOffset(code, end)
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := range offset {
		if code[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func insertBeforeFirst(code string, target string, insertion string) string {
	index := strings.Index(code, target)
	if index < 0 {
		panic(fmt.Sprintf("target %q not found in %q", target, code))
	}
	return code[:index] + insertion + code[index:]
}

func enforceFixText(code string, target string) string {
	index := strings.Index(code, target)
	if utils.NeedsLeadingSpaceForReplacement(code, index, "new ") {
		return " new "
	}
	return "new "
}

func enforceMessage(name string) string {
	return fmt.Sprintf("Use `new %s()` instead of `%s()`.", name, name)
}

func disallowMessage(name string) string {
	return fmt.Sprintf("Use `%s()` instead of `new %s()`.", name, name)
}

func disallowCallOrNewMessage(name string) string {
	return fmt.Sprintf("`%s` is not a function or constructor.", name)
}
