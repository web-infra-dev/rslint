// TestNoEmptyFunctionUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-empty-function.js 1:1. Position assertions
// cover line/column/endLine/endColumn for every invalid case. rslint-specific
// lock-in cases live in the no_empty_function_extras_test.go file.
package no_empty_function

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var allAllowOptions = []string{
	"functions",
	"arrowFunctions",
	"generatorFunctions",
	"methods",
	"generatorMethods",
	"getters",
	"setters",
	"constructors",
	"asyncFunctions",
	"asyncMethods",
	"privateConstructors",
	"protectedConstructors",
	"decoratedFunctions",
	"overrideMethods",
}

type upstreamSeed struct {
	code  string
	name  string
	allow []string
}

func TestNoEmptyFunctionUpstream(t *testing.T) {
	valid, invalid := upstreamCases(
		[]rule_tester.ValidTestCase{{Code: "var foo = () => 0;"}},
		[]rule_tester.InvalidTestCase{
			locationCase("function foo() {}", "function 'foo'", 1, 16, 1, 18, "function foo() { /* empty */ }"),
			locationCase("var foo = function () {\n}", "function", 1, 23, 2, 2, "var foo = function () { /* empty */ }"),
			locationCase("var foo = () => { \n\n  }", "arrow function", 1, 17, 3, 4, "var foo = () => { /* empty */ }"),
			locationCase("var obj = {\n\tfoo() {\n\t}\n}", "method 'foo'", 2, 8, 3, 3, "var obj = {\n\tfoo() { /* empty */ }\n}"),
			locationCase("class A { foo() { } }", "method 'foo'", 1, 17, 1, 20, "class A { foo() { /* empty */ } }"),
		},
		upstreamJSSSeeds,
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		valid,
		invalid,
	)
}

func TestNoEmptyFunctionUpstreamTypeScript(t *testing.T) {
	valid, invalid := upstreamCases(
		[]rule_tester.ValidTestCase{
			{Code: "class A { constructor(public param: string) {} }"},
			{Code: "class A { constructor(private param: string) {} }"},
			{Code: "class A { constructor(protected param: string) {} }"},
			{Code: "class A { constructor(readonly param: string) {} }"},
		},
		nil,
		upstreamTSSeeds,
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		valid,
		invalid,
	)
}

func upstreamCases(initialValid []rule_tester.ValidTestCase, initialInvalid []rule_tester.InvalidTestCase, seeds []upstreamSeed) ([]rule_tester.ValidTestCase, []rule_tester.InvalidTestCase) {
	valid := append([]rule_tester.ValidTestCase{}, initialValid...)
	invalid := append([]rule_tester.InvalidTestCase{}, initialInvalid...)

	for _, seed := range seeds {
		valid = append(valid,
			rule_tester.ValidTestCase{Code: strings.Replace(seed.code, "{}", "{ bar(); }", 1)},
			rule_tester.ValidTestCase{Code: strings.Replace(seed.code, "{}", "{ /* empty */ }", 1)},
			rule_tester.ValidTestCase{Code: strings.Replace(seed.code, "{}", "{\n    // empty\n}", 1)},
		)
		for _, allow := range seed.allow {
			valid = append(valid, rule_tester.ValidTestCase{
				Code:    seed.code + " // allow: " + allow,
				Options: allowOption(allow),
			})
		}

		invalid = append(invalid, generatedInvalidCase(seed.code, seed.name, nil))
		for _, allow := range allAllowOptions {
			if containsString(seed.allow, allow) {
				continue
			}
			code := seed.code + " // allow: " + allow
			invalid = append(invalid, generatedInvalidCase(code, seed.name, allowOption(allow)))
		}
	}

	return valid, invalid
}

func allowOption(allow string) []interface{} {
	return []interface{}{map[string]interface{}{"allow": []interface{}{allow}}}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func generatedInvalidCase(code string, name string, options any) rule_tester.InvalidTestCase {
	tc := rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Errors: []rule_tester.InvalidTestCaseError{
			generatedInvalidError(code, name, 0),
		},
	}
	return tc
}

func generatedInvalidCaseWithNames(code string, names ...string) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(names))
	for i, name := range names {
		errors = append(errors, generatedInvalidError(code, name, i))
	}
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: errors,
	}
}

func generatedInvalidError(code string, name string, occurrence int) rule_tester.InvalidTestCaseError {
	start := nthIndex(code, "{}", occurrence)
	if start < 0 {
		panic("generated no-empty-function case must contain enough {} bodies: " + code)
	}
	line, column := lineColumnForOffset(code, start)
	endLine, endColumn := lineColumnForOffset(code, start+2)
	output := code[:start] + "{ /* empty */ }" + code[start+2:]
	return emptyFunctionError(name, line, column, endLine, endColumn, output)
}

func nthIndex(s string, substr string, occurrence int) int {
	if occurrence < 0 {
		return -1
	}
	offset := 0
	for i := 0; i <= occurrence; i++ {
		idx := strings.Index(s[offset:], substr)
		if idx < 0 {
			return -1
		}
		if i == occurrence {
			return offset + idx
		}
		offset += idx + len(substr)
	}
	return -1
}

func locationCase(code string, name string, line int, column int, endLine int, endColumn int, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			emptyFunctionError(name, line, column, endLine, endColumn, output),
		},
	}
}

func emptyFunctionError(name string, line int, column int, endLine int, endColumn int, output string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpected",
		Message:   "Unexpected empty " + name + ".",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "suggestComment", Output: output},
		},
	}
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line, column := 1, 1
	for i, r := range code {
		if i >= offset {
			break
		}
		if r == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

var upstreamJSSSeeds = []upstreamSeed{
	{code: `function foo() {}`, name: "function 'foo'", allow: []string{"functions"}},
	{code: `var foo = function() {};`, name: "function", allow: []string{"functions"}},
	{code: `var obj = {foo: function() {}};`, name: "method 'foo'", allow: []string{"functions"}},
	{code: `var foo = () => {};`, name: "arrow function", allow: []string{"arrowFunctions"}},
	{code: `function* foo() {}`, name: "generator function 'foo'", allow: []string{"generatorFunctions"}},
	{code: `var foo = function*() {};`, name: "generator function", allow: []string{"generatorFunctions"}},
	{code: `var obj = {foo: function*() {}};`, name: "generator method 'foo'", allow: []string{"generatorFunctions"}},
	{code: `var obj = {foo() {}};`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A {foo() {}}`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A {static foo() {}}`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `var A = class {foo() {}};`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `var A = class {static foo() {}};`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `var obj = {*foo() {}};`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A {*foo() {}}`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A {static *foo() {}}`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `var A = class {*foo() {}};`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `var A = class {static *foo() {}};`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `var obj = {get foo() {}};`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `class A {get foo() {}}`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `class A {static get foo() {}}`, name: "static getter 'foo'", allow: []string{"getters"}},
	{code: `var A = class {get foo() {}};`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `var A = class {static get foo() {}};`, name: "static getter 'foo'", allow: []string{"getters"}},
	{code: `var obj = {set foo(value) {}};`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `class A {set foo(value) {}}`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `class A {static set foo(value) {}}`, name: "static setter 'foo'", allow: []string{"setters"}},
	{code: `var A = class {set foo(value) {}};`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `var A = class {static set foo(value) {}};`, name: "static setter 'foo'", allow: []string{"setters"}},
	{code: `class A {constructor() {}}`, name: "constructor", allow: []string{"constructors"}},
	{code: `var A = class {constructor() {}};`, name: "constructor", allow: []string{"constructors"}},
	{code: `const foo = { async method() {} }`, name: "async method 'method'", allow: []string{"asyncMethods"}},
	{code: `async function a(){}`, name: "async function 'a'", allow: []string{"asyncFunctions"}},
	{code: `const foo = async function () {}`, name: "async function", allow: []string{"asyncFunctions"}},
	{code: `class Foo { async bar() {} }`, name: "async method 'bar'", allow: []string{"asyncMethods"}},
	{code: `const foo = async () => {};`, name: "async arrow function", allow: []string{"arrowFunctions"}},
}

var upstreamTSSeeds = []upstreamSeed{
	{code: `function foo() {}`, name: "function 'foo'", allow: []string{"functions"}},
	{code: `const foo = function(param: string) {};`, name: "function", allow: []string{"functions"}},
	{code: `const obj = {foo: function(param: string) {}};`, name: "method 'foo'", allow: []string{"functions"}},
	{code: `const foo = (param: string) => {};`, name: "arrow function", allow: []string{"arrowFunctions"}},
	{code: `function* foo(param: string) {}`, name: "generator function 'foo'", allow: []string{"generatorFunctions"}},
	{code: `const foo = function*(param: string) {};`, name: "generator function", allow: []string{"generatorFunctions"}},
	{code: `const obj = {foo: function*(param: string) {}};`, name: "generator method 'foo'", allow: []string{"generatorFunctions"}},
	{code: `const obj = {foo(param: string) {}};`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A { foo(param: string) {} }`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A { private foo() {} }`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A { protected foo() {} }`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `class A { static foo(param: string) {} }`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `class A { private static foo() {} }`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `class A { protected static foo() {} }`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `const A = class {foo(param: string) {}};`, name: "method 'foo'", allow: []string{"methods"}},
	{code: `const A = class {static foo(param: string) {}};`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `const A = class {private static foo() {}};`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `const A = class {protected static foo() {}};`, name: "static method 'foo'", allow: []string{"methods"}},
	{code: `class B { @decorator() foo() {} }`, name: "method 'foo'", allow: []string{"methods", "decoratedFunctions"}},
	{code: `const B = class { @decorator() foo() {} }`, name: "method 'foo'", allow: []string{"methods", "decoratedFunctions"}},
	{code: `class B extends C { override foo() {} }`, name: "method 'foo'", allow: []string{"methods", "overrideMethods"}},
	{code: `class B extends C { @decorator() override foo() {} }`, name: "method 'foo'", allow: []string{"methods", "decoratedFunctions", "overrideMethods"}},
	{code: `const obj = {*foo(param: string) {}};`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A { *foo(param: string) {} }`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A {static *foo(param: string) {}}`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A {private static *foo() {}}`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `class A {protected static *foo() {}}`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `const A = class {*foo(param: string) {}};`, name: "generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `const A = class {static *foo(param: string) {}};`, name: "static generator method 'foo'", allow: []string{"generatorMethods"}},
	{code: `const obj = {get foo(): string {}};`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `class A {get foo(): string {}}`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `class A {static get foo(): string {}}`, name: "static getter 'foo'", allow: []string{"getters"}},
	{code: `const A = class {get foo(): string {}};`, name: "getter 'foo'", allow: []string{"getters"}},
	{code: `const A = class {static get foo(): string {}};`, name: "static getter 'foo'", allow: []string{"getters"}},
	{code: `class A {@decorator() get foo(): string {}}`, name: "getter 'foo'", allow: []string{"getters", "decoratedFunctions"}},
	{code: `class A {@decorator() static get foo(): string {}}`, name: "static getter 'foo'", allow: []string{"getters", "decoratedFunctions"}},
	{code: `const A = class {@decorator() get foo(): string {}};`, name: "getter 'foo'", allow: []string{"getters", "decoratedFunctions"}},
	{code: `const A = class {@decorator() static get foo(): string {}};`, name: "static getter 'foo'", allow: []string{"getters", "decoratedFunctions"}},
	{code: `class A extends B {override get foo(): string {}}`, name: "getter 'foo'", allow: []string{"getters", "overrideMethods"}},
	{code: `class A extends B {static override get foo(): string {}}`, name: "static getter 'foo'", allow: []string{"getters", "overrideMethods"}},
	{code: `const A = class extends B {override get foo(): string {}};`, name: "getter 'foo'", allow: []string{"getters", "overrideMethods"}},
	{code: `const A = class extends B {static override get foo(): string {}};`, name: "static getter 'foo'", allow: []string{"getters", "overrideMethods"}},
	{code: `const obj = {set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `class A {set foo(value: string) {}}`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `class A {static set foo(value: string) {}}`, name: "static setter 'foo'", allow: []string{"setters"}},
	{code: `const A = class {set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"setters"}},
	{code: `const A = class {static set foo(value: string) {}};`, name: "static setter 'foo'", allow: []string{"setters"}},
	{code: `class A {@decorator() set foo(value: string) {}}`, name: "setter 'foo'", allow: []string{"setters", "decoratedFunctions"}},
	{code: `class A {@decorator() static set foo(value: string) {}}`, name: "static setter 'foo'", allow: []string{"setters", "decoratedFunctions"}},
	{code: `const A = class {@decorator() set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"setters", "decoratedFunctions"}},
	{code: `const A = class {@decorator() static set foo(value: string) {}};`, name: "static setter 'foo'", allow: []string{"setters", "decoratedFunctions"}},
	{code: `class A extends B {override set foo(value: string) {}}`, name: "setter 'foo'", allow: []string{"setters", "overrideMethods"}},
	{code: `class A extends B {static override set foo(value: string) {}}`, name: "static setter 'foo'", allow: []string{"setters", "overrideMethods"}},
	{code: `const A = class extends B {override set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"setters", "overrideMethods"}},
	{code: `const A = class extends B {static override set foo(value: string) {}};`, name: "static setter 'foo'", allow: []string{"setters", "overrideMethods"}},
	{code: `class A { constructor(param: string) {} }`, name: "constructor", allow: []string{"constructors"}},
	{code: `class B { private constructor() {} }`, name: "constructor", allow: []string{"constructors", "privateConstructors"}},
	{code: `class B { protected constructor() {} }`, name: "constructor", allow: []string{"constructors", "protectedConstructors"}},
	{code: `const A = class {constructor(param: string) {}};`, name: "constructor", allow: []string{"constructors"}},
	{code: `const B = class { private constructor() {} }`, name: "constructor", allow: []string{"constructors", "privateConstructors"}},
	{code: `const B = class { protected constructor() {} }`, name: "constructor", allow: []string{"constructors", "protectedConstructors"}},
	{code: `const foo = { async method(param: string) {} }`, name: "async method 'method'", allow: []string{"asyncMethods"}},
	{code: `async function a(param: string){}`, name: "async function 'a'", allow: []string{"asyncFunctions"}},
	{code: `const foo = async function(param: string) {}`, name: "async function", allow: []string{"asyncFunctions"}},
	{code: `class A { async foo(param: string) {} }`, name: "async method 'foo'", allow: []string{"asyncMethods"}},
	{code: `class A { @decorator() async foo(param: string) {} }`, name: "async method 'foo'", allow: []string{"asyncMethods", "decoratedFunctions"}},
	{code: `class A extends B { override async foo(param: string) {} }`, name: "async method 'foo'", allow: []string{"asyncMethods", "overrideMethods"}},
	{code: `const foo = async (): Promise<void> => {};`, name: "async arrow function", allow: []string{"arrowFunctions"}},
	{code: `class A { private constructor() {} }`, name: "constructor", allow: []string{"privateConstructors", "constructors"}},
	{code: `const A = class { private constructor() {} };`, name: "constructor", allow: []string{"privateConstructors", "constructors"}},
	{code: `class A { protected constructor() {} }`, name: "constructor", allow: []string{"protectedConstructors", "constructors"}},
	{code: `const A = class { protected constructor() {} };`, name: "constructor", allow: []string{"protectedConstructors", "constructors"}},
	{code: `class A { @decorator() foo() {} }`, name: "method 'foo'", allow: []string{"decoratedFunctions", "methods"}},
	{code: `const A = class { @decorator() foo() {} }`, name: "method 'foo'", allow: []string{"decoratedFunctions", "methods"}},
	{code: `class B {@decorator() get foo(): string {}}`, name: "getter 'foo'", allow: []string{"decoratedFunctions", "getters"}},
	{code: `class B {@decorator() static get foo(): string {}}`, name: "static getter 'foo'", allow: []string{"decoratedFunctions", "getters"}},
	{code: `const B = class {@decorator() get foo(): string {}};`, name: "getter 'foo'", allow: []string{"decoratedFunctions", "getters"}},
	{code: `const B = class {@decorator() static get foo(): string {}};`, name: "static getter 'foo'", allow: []string{"decoratedFunctions", "getters"}},
	{code: `class B {@decorator() set foo(value: string) {}}`, name: "setter 'foo'", allow: []string{"decoratedFunctions", "setters"}},
	{code: `class B {@decorator() static set foo(value: string) {}}`, name: "static setter 'foo'", allow: []string{"decoratedFunctions", "setters"}},
	{code: `const B = class {@decorator() set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"decoratedFunctions", "setters"}},
	{code: `const B = class {@decorator() static set foo(value: string) {}};`, name: "static setter 'foo'", allow: []string{"decoratedFunctions", "setters"}},
	{code: `class B { @decorator() async foo(param: string) {} }`, name: "async method 'foo'", allow: []string{"decoratedFunctions", "asyncMethods"}},
	{code: `class A extends B { @decorator() override foo() {} }`, name: "method 'foo'", allow: []string{"decoratedFunctions", "methods", "overrideMethods"}},
	{code: `class B extends C {override get foo(): string {}}`, name: "getter 'foo'", allow: []string{"overrideMethods", "getters"}},
	{code: `class B extends C {static override get foo(): string {}}`, name: "static getter 'foo'", allow: []string{"overrideMethods", "getters"}},
	{code: `const B = class extends C {override get foo(): string {}};`, name: "getter 'foo'", allow: []string{"overrideMethods", "getters"}},
	{code: `const B = class extends C {static override get foo(): string {}};`, name: "static getter 'foo'", allow: []string{"overrideMethods", "getters"}},
	{code: `class B extends C {override set foo(value: string) {}}`, name: "setter 'foo'", allow: []string{"overrideMethods", "setters"}},
	{code: `class B extends C {static override set foo(value: string) {}}`, name: "static setter 'foo'", allow: []string{"overrideMethods", "setters"}},
	{code: `const B = class extends C {override set foo(value: string) {}};`, name: "setter 'foo'", allow: []string{"overrideMethods", "setters"}},
	{code: `const B = class extends C {static override set foo(value: string) {}};`, name: "static setter 'foo'", allow: []string{"overrideMethods", "setters"}},
	{code: `class B extends C { override async foo(param: string) {} }`, name: "async method 'foo'", allow: []string{"overrideMethods", "asyncMethods"}},
	{code: `class C extends D { @decorator() override foo() {} }`, name: "method 'foo'", allow: []string{"overrideMethods", "methods", "decoratedFunctions"}},
	{code: `class A extends B { override foo() {} }`, name: "method 'foo'", allow: []string{"overrideMethods", "methods"}},
}
