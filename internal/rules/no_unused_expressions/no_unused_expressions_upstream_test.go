// TestNoUnusedExpressionsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-unused-expressions.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in the no_unused_expressions_extras_test.go file.
package no_unused_expressions

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const unusedExpressionMessage = "Expected an assignment or function call and instead saw an expression."

func TestNoUnusedExpressionsUpstream(t *testing.T) {
	allowShortCircuit := map[string]interface{}{"allowShortCircuit": true}
	allowTernary := map[string]interface{}{"allowTernary": true}
	allowShortCircuitAndTernary := map[string]interface{}{"allowShortCircuit": true, "allowTernary": true}
	allowTaggedTemplates := map[string]interface{}{"allowTaggedTemplates": true}
	disallowTaggedTemplates := map[string]interface{}{"allowTaggedTemplates": false}
	enforceForJSX := map[string]interface{}{"enforceForJSX": true}
	ignoreDirectives := map[string]interface{}{"ignoreDirectives": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedExpressionsRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
			{Code: "function f(){}"},
			{Code: "a = b"},
			{Code: "new a"},
			{Code: "{}"},
			{Code: "f(); g()"},
			{Code: "i++"},
			{Code: "a()"},
			{Code: "a && a()", Options: allowShortCircuit},
			{Code: "a() || (b = c)", Options: allowShortCircuit},
			{Code: "a ? b() : c()", Options: allowTernary},
			{Code: "a ? b() || (c = d) : e()", Options: allowShortCircuitAndTernary},
			{Code: "delete foo.bar"},
			{Code: "void new C"},
			{Code: "\"use strict\";"},
			{Code: "\"directive one\"; \"directive two\"; f();"},
			{Code: "function foo() {\"use strict\"; return true; }"},
			{Code: "var foo = () => {\"use strict\"; return true; }"},
			{Code: "function foo() {\"directive one\"; \"directive two\"; f(); }"},
			{Code: "function foo() { var foo = \"use strict\"; return true; }"},
			{Code: "function* foo(){ yield 0; }"},
			{Code: "async function foo() { await 5; }"},
			{Code: "async function foo() { await foo.bar; }"},
			{Code: "async function foo() { bar && await baz; }", Options: allowShortCircuit},
			{Code: "async function foo() { foo ? await bar : await baz; }", Options: allowTernary},
			{Code: "tag`tagged template literal`", Options: allowTaggedTemplates},
			{Code: "shouldNotBeAffectedByAllowTemplateTagsOption()", Options: allowTaggedTemplates},
			{Code: "import(\"foo\")"},
			{Code: "func?.(\"foo\")"},
			{Code: "obj?.foo(\"bar\")"},
			// ---- upstream valid: JSX ----
			{Code: "<div />", Tsx: true},
			{Code: "<></>", Tsx: true},
			{Code: "var partial = <div />", Tsx: true},
			{Code: "var partial = <div />", Options: enforceForJSX, Tsx: true},
			{Code: "var partial = <></>", Options: enforceForJSX, Tsx: true},
			// ---- upstream valid: ignoreDirectives option ----
			{Code: "\"use strict\";", Options: ignoreDirectives},
			{Code: "\"directive one\"; \"directive two\"; f();", Options: ignoreDirectives},
			{Code: "function foo() {\"use strict\"; return true; }", Options: ignoreDirectives},
			{Code: "function foo() {\"directive one\"; \"directive two\"; f(); }", Options: ignoreDirectives},
			// ---- upstream valid: TypeScript-specific ----
			{Code: "test.age?.toLocaleString();"},
			{Code: "let a = (a?.b).c;"},
			{Code: "let b = a?.['b'];"},
			{Code: "let c = one[2]?.[3][4];"},
			{Code: "one[2]?.[3][4]?.();"},
			{Code: "a?.['b']?.c();"},
			{Code: "module Foo {\n  'use strict';\n}"},
			{Code: "namespace Foo {\n  'use strict';\n\n  export class Foo {}\n  export class Bar {}\n}"},
			{Code: "function foo() {\n  'use strict';\n\n  return null;\n}"},
			{Code: "import('./foo');"},
			{Code: "import('./foo').then(() => {});"},
			{Code: "class Foo<T> {}\nnew Foo<string>();"},
			{Code: "foo && foo?.();", Options: allowShortCircuit},
			{Code: "foo && import('./foo');", Options: allowShortCircuit},
			{Code: "foo ? import('./foo') : import('./bar');", Options: allowTernary},
			{Code: "<div/> as any", Tsx: true},
			{Code: "foo && foo()!;", Options: allowShortCircuit},
			{
				Code:    "declare const foo: Function | undefined;\n<any>(foo && foo()!)",
				Options: allowShortCircuit,
			},
			{Code: "(Foo && Foo())<string, number>;", Options: allowShortCircuit},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid ----
			invalidUnusedExpression("0", "0"),
			invalidUnusedExpression("a", "a"),
			invalidUnusedExpression("f(), 0", "f(), 0"),
			invalidUnusedExpression("{0}", "0"),
			invalidUnusedExpression("[]", "[]"),
			invalidUnusedExpression("a && b();", "a && b();"),
			invalidUnusedExpression("a() || false", "a() || false"),
			invalidUnusedExpression("a || (b = c)", "a || (b = c)"),
			invalidUnusedExpression("a ? b() || (c = d) : e", "a ? b() || (c = d) : e"),
			invalidUnusedExpression("`untagged template literal`", "`untagged template literal`"),
			invalidUnusedExpression("tag`tagged template literal`", "tag`tagged template literal`"),
			invalidUnusedExpressionWithOptions("a && b()", "a && b()", allowTernary),
			invalidUnusedExpressionWithOptions("a ? b() : c()", "a ? b() : c()", allowShortCircuit),
			invalidUnusedExpressionWithOptions("a || b", "a || b", allowShortCircuit),
			invalidUnusedExpressionWithOptions("a() && b", "a() && b", allowShortCircuit),
			invalidUnusedExpressionWithOptions("a ? b : 0", "a ? b : 0", allowTernary),
			invalidUnusedExpressionWithOptions("a ? b : c()", "a ? b : c()", allowTernary),
			invalidUnusedExpression("foo.bar;", "foo.bar;"),
			invalidUnusedExpression("!a", "!a"),
			invalidUnusedExpression("+a", "+a"),
			invalidUnusedExpression("\"directive one\"; f(); \"directive two\";", "\"directive two\";"),
			invalidUnusedExpression("function foo() {\"directive one\"; f(); \"directive two\"; }", "\"directive two\";"),
			invalidUnusedExpression("if (0) { \"not a directive\"; f(); }", "\"not a directive\";"),
			invalidUnusedExpression("function foo() { var foo = true; \"use strict\"; }", "\"use strict\";"),
			invalidUnusedExpression("var foo = () => { var foo = true; \"use strict\"; }", "\"use strict\";"),
			invalidUnusedExpressionWithOptions("`untagged template literal`", "`untagged template literal`", allowTaggedTemplates),
			invalidUnusedExpressionWithOptions("`untagged template literal`", "`untagged template literal`", disallowTaggedTemplates),
			invalidUnusedExpressionWithOptions("tag`tagged template literal`", "tag`tagged template literal`", disallowTaggedTemplates),
			// ---- upstream invalid: optional chaining ----
			invalidUnusedExpression("obj?.foo", "obj?.foo"),
			invalidUnusedExpression("obj?.foo.bar", "obj?.foo.bar"),
			invalidUnusedExpression("obj?.foo().bar", "obj?.foo().bar"),
			// ---- upstream invalid: JSX ----
			invalidUnusedExpressionWithOptionsTsx("<div />", "<div />", enforceForJSX),
			invalidUnusedExpressionWithOptionsTsx("<></>", "<></>", enforceForJSX),
			// ---- upstream invalid: class static blocks do not have directive prologues ----
			invalidUnusedExpression("class C { static { 'use strict'; } }", "'use strict';"),
			{
				Code: "class C { static {\n'foo'\n'bar'\n} }",
				Errors: []rule_tester.InvalidTestCaseError{
					unusedExpressionErrorAt("class C { static {\n'foo'\n'bar'\n} }", "'foo'"),
					unusedExpressionErrorAt("class C { static {\n'foo'\n'bar'\n} }", "'bar'"),
				},
			},
			invalidUnusedExpressionWithOptions("foo;", "foo;", ignoreDirectives),
			// SKIP: rslint does not support ESLint's ecmaVersion: 3 parser mode,
			// where string literals are not represented as directive prologues.
			{Code: "\"use strict\";", Skip: true},
			// SKIP: rslint does not support ESLint's ecmaVersion: 3 parser mode.
			{Code: "\"directive one\"; \"directive two\"; f();", Skip: true},
			// SKIP: rslint does not support ESLint's ecmaVersion: 3 parser mode.
			{Code: "function foo() {\"use strict\"; return true; }", Skip: true},
			// SKIP: rslint does not support ESLint's ecmaVersion: 3 parser mode.
			{Code: "function foo() {\"directive one\"; \"directive two\"; f(); }", Skip: true},
			// ---- upstream invalid: TypeScript-specific ----
			invalidUnusedExpression("\n  if (0) 0;\n", "0;"),
			invalidUnusedExpression("\n  f(0), {};\n", "f(0), {};"),
			invalidUnusedExpression("\n  a, b();\n", "a, b();"),
			invalidUnusedExpression("\n  a() &&\n\tfunction namedFunctionInExpressionContext() {\n\t  f();\n\t};\n", "a() &&"),
			invalidUnusedExpression("\n  a?.b;\n", "a?.b;"),
			invalidUnusedExpression("\n  (a?.b).c;\n", "(a?.b).c;"),
			invalidUnusedExpression("\n  a?.['b'];\n", "a?.['b'];"),
			invalidUnusedExpression("\n  (a?.['b']).c;\n", "(a?.['b']).c;"),
			invalidUnusedExpression("\n  a?.b()?.c;\n", "a?.b()?.c;"),
			invalidUnusedExpression("\n  (a?.b()).c;\n", "(a?.b()).c;"),
			invalidUnusedExpression("\n  one[2]?.[3][4];\n", "one[2]?.[3][4];"),
			invalidUnusedExpression("\n  one.two?.three.four;\n", "one.two?.three.four;"),
			invalidUnusedExpression("module Foo {\n  const foo = true;\n  'use strict';\n}", "'use strict';"),
			invalidUnusedExpression("namespace Foo {\n  export class Foo {}\n  export class Bar {}\n\n  'use strict';\n}", "'use strict';"),
			invalidUnusedExpression("function foo() {\n  const foo = true;\n\n  'use strict';\n}", "'use strict';"),
			invalidUnusedExpressionWithOptions("foo && foo?.bar;", "foo && foo?.bar;", allowShortCircuit),
			invalidUnusedExpressionWithOptions("foo ? foo?.bar : bar.baz;", "foo ? foo?.bar : bar.baz;", allowTernary),
			invalidUnusedExpression("\n  class Foo<T> {}\n  Foo<string>;\n", "Foo<string>;"),
			invalidUnusedExpression("Map<string, string>;", "Map<string, string>;"),
			invalidUnusedExpression("\n  declare const foo: number | undefined;\n  foo;\n", "foo;"),
			invalidUnusedExpression("\n  declare const foo: number | undefined;\n  foo as any;\n", "foo as any;"),
			invalidUnusedExpression("\n  declare const foo: number | undefined;\n  <any>foo;\n", "<any>foo;"),
			invalidUnusedExpression("\n  declare const foo: number | undefined;\n  foo!;\n", "foo!;"),
		},
	)
}

func invalidUnusedExpression(code string, snippet string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			unusedExpressionErrorAt(code, snippet),
		},
	}
}

func invalidUnusedExpressionWithOptions(code string, snippet string, options any) rule_tester.InvalidTestCase {
	tc := invalidUnusedExpression(code, snippet)
	tc.Options = options
	return tc
}

func invalidUnusedExpressionWithOptionsTsx(code string, snippet string, options any) rule_tester.InvalidTestCase {
	tc := invalidUnusedExpressionWithOptions(code, snippet, options)
	tc.Tsx = true
	return tc
}

func unusedExpressionErrorAt(code string, snippet string) rule_tester.InvalidTestCaseError {
	line, column := lineColumnForSnippet(code, snippet)
	return rule_tester.InvalidTestCaseError{
		MessageId: "unusedExpression",
		Message:   unusedExpressionMessage,
		Line:      line,
		Column:    column,
	}
}

func lineColumnForSnippet(code string, snippet string) (int, int) {
	index := strings.Index(code, snippet)
	if index < 0 {
		panic("snippet not found: " + snippet)
	}
	before := code[:index]
	line := strings.Count(before, "\n") + 1
	lastNewline := strings.LastIndex(before, "\n")
	if lastNewline < 0 {
		return line, len(before) + 1
	}
	return line, len(before[lastNewline+1:]) + 1
}
