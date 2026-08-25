package no_array_constructor

// TestNoArrayConstructorUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-array-constructor.js 1:1 — both the plain
// RuleTester (Espree) and the ruleTesterTypeScript (@typescript-eslint/parser)
// blocks, since rslint parses every file with a single TS-aware parser and
// has no equivalent split. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in the
// no_array_constructor_extras_test.go file.

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func lineColAt(code string, idx int) (line, col int) {
	before := code[:idx]
	line = 1 + strings.Count(before, "\n")
	lastNL := strings.LastIndex(before, "\n")
	col = idx - lastNL
	return
}

// rangeOf locates the first occurrence of marker in code and returns its
// 1-based line/column span (EndColumn is exclusive), matching the
// diagnostic's reported node range.
func rangeOf(code, marker string) (startLine, startCol, endLine, endCol int) {
	idx := strings.Index(code, marker)
	if idx < 0 {
		panic("rangeOf: marker not found: " + marker)
	}
	startLine, startCol = lineColAt(code, idx)
	endLine, endCol = lineColAt(code, idx+len(marker))
	return
}

// directFixCase builds an InvalidTestCase for a report that upstream applies
// as a direct autofix (shouldSuggest is false: no optional chain, no leading
// comment, and either zero arguments or ≥2 non-spread arguments).
func directFixCase(code, marker, output string) rule_tester.InvalidTestCase {
	l1, c1, l2, c2 := rangeOf(code, marker)
	return rule_tester.InvalidTestCase{
		Code:   code,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferLiteral",
			Line:      l1, Column: c1, EndLine: l2, EndColumn: c2,
		}},
	}
}

// suggestCase builds an InvalidTestCase for a report upstream only offers as
// a suggestion (shouldSuggest is true).
func suggestCase(code, marker, messageId, output string) rule_tester.InvalidTestCase {
	l1, c1, l2, c2 := rangeOf(code, marker)
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferLiteral",
			Line:      l1, Column: c1, EndLine: l2, EndColumn: c2,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: messageId,
				Output:    output,
			}},
		}},
	}
}

// asiCase builds a directFixCase for one of upstream's ASI-safety fixtures:
// code containing exactly one occurrence of pattern (either "Array(...)" or
// "new Array(...)"). All of upstream's ASI fixtures are argument-free or have
// ≥2 non-spread arguments, so the fix is always a direct autofix, never a
// suggestion.
func asiCase(code, pattern string, needsSemicolon bool) rule_tester.InvalidTestCase {
	idx := strings.Index(code, pattern)
	if idx < 0 {
		panic("asiCase: pattern not found in code: " + pattern)
	}
	open := strings.Index(pattern, "(")
	argsText := pattern[open+1 : len(pattern)-1]
	fixText := "[" + argsText + "]"
	if needsSemicolon {
		fixText = ";" + fixText
	}
	output := code[:idx] + fixText + code[idx+len(pattern):]
	return directFixCase(code, pattern, output)
}

// tsAsiCase builds a directFixCase for the "no semicolon required after
// TypeScript syntax" block: prefix is a TS declaration/statement that never
// leaves an ASI hazard behind it, followed by `Array(0, 1)` on the next line.
//
// weAddSemicolon locks in a documented, safe divergence from upstream: our
// utils.NeedsPrecedingSemicolon (see its doc comment) omits ESLint's
// TypeScript-only exemptions for type positions, ambient/overload function
// declarations, and import-equals declarations, so it falls back to the
// conservative "needs a semicolon" answer for the token kinds these prefixes
// end with — upstream needs none. The extra `;` never changes behavior, only
// adds a redundant character, so it's a documented difference (see the rule's
// "Differences from ESLint" section) rather than a bug.
func tsAsiCase(prefix string, weAddSemicolon bool) rule_tester.InvalidTestCase {
	code := prefix + "\nArray(0, 1)"
	fix := "[0, 1]"
	if weAddSemicolon {
		fix = ";" + fix
	}
	output := prefix + "\n" + fix
	return directFixCase(code, "Array(0, 1)", output)
}

func TestNoArrayConstructorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoArrayConstructorRule,
		[]rule_tester.ValidTestCase{
			// ---- plain RuleTester valid ----
			{Code: `new Array(x)`},
			{Code: `Array(x)`},
			{Code: `new Array(9)`},
			{Code: `Array(9)`},
			{Code: `new foo.Array()`},
			{Code: `foo.Array()`},
			{Code: `new Array.foo`},
			{Code: `Array.foo()`},
			{Code: `new globalThis.Array`},
			{Code: `const createArray = Array => new Array()`},
			{Code: `var Array; new Array;`},
			{
				Code: `new Array()`, FileName: "no-array-constructor-espree.js", TSConfig: "tsconfig.allow-js.json",
				Globals: map[string]any{"Array": "off"},
			},

			// ---- ruleTesterTypeScript valid ----
			{Code: `new Array(x);`},
			{Code: `Array(x);`},
			{Code: `new Array(9);`},
			{Code: `Array(9);`},
			{Code: `new foo.Array();`},
			{Code: `foo.Array();`},
			{Code: `new Array.foo();`},
			{Code: `Array.foo();`},
			// TypeScript: explicit type arguments mean this isn't the bare
			// constructor call the rule targets.
			{Code: `new Array<Foo>(1, 2, 3);`},
			{Code: `new Array<Foo>();`},
			{Code: `Array<Foo>(1, 2, 3);`},
			{Code: `Array<Foo>();`},
			{Code: `Array<Foo>(3);`},
			// optional chain — still a single non-spread argument or a
			// member/type-argument form the rule doesn't target.
			{Code: `Array?.(x);`},
			{Code: `Array?.(9);`},
			{Code: `foo?.Array();`},
			{Code: `Array?.foo();`},
			{Code: `foo.Array?.();`},
			{Code: `Array.foo?.();`},
			{Code: `Array?.<Foo>(1, 2, 3);`},
			{Code: `Array?.<Foo>();`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- plain RuleTester: explicit cases ----
			directFixCase(`new Array()`, `new Array()`, `[]`),
			directFixCase(`new Array`, `new Array`, `[]`),
			directFixCase(`new Array(x, y)`, `new Array(x, y)`, `[x, y]`),
			directFixCase(`new Array(0, 1, 2)`, `new Array(0, 1, 2)`, `[0, 1, 2]`),
			suggestCase(`const array = Array?.();`, `Array?.()`, "useLiteral", `const array = [];`),
			directFixCase(
				"const array = (Array)(\n    /* foo */ a,\n    b = c() // bar\n);",
				"(Array)(\n    /* foo */ a,\n    b = c() // bar\n)",
				"const array = [\n    /* foo */ a,\n    b = c() // bar\n];",
			),
			suggestCase(`const array = Array(...args);`, `Array(...args)`, "useLiteral", `const array = [...args];`),
			suggestCase(`const array = Array(...foo, ...bar);`, `Array(...foo, ...bar)`, "useLiteral", `const array = [...foo, ...bar];`),
			suggestCase(`const array = new Array(...args);`, `new Array(...args)`, "useLiteral", `const array = [...args];`),
			suggestCase(`const array = Array(5, ...args);`, `Array(5, ...args)`, "useLiteral", `const array = [5, ...args];`),
			directFixCase(`const array = Array(5, 6, ...args);`, `Array(5, 6, ...args)`, `const array = [5, 6, ...args];`),
			directFixCase(`a = new (Array);`, `new (Array)`, `a = [];`),
			directFixCase(`a = new (Array) && (foo);`, `new (Array)`, `a = [] && (foo);`),

			// ---- Semicolon required before array literal to compensate for ASI ----
			asiCase("foo\nArray()", "Array()", true),
			asiCase("foo()\nArray(bar, baz)", "Array(bar, baz)", true),
			asiCase("new foo\nArray()", "Array()", true),
			asiCase("(a++)\nArray()", "Array()", true),
			asiCase("++a\nArray()", "Array()", true),
			asiCase("const foo = function() {}\nArray()", "Array()", true),
			asiCase(`const foo = class {}
Array("a", "b", "c")`, `Array("a", "b", "c")`, true),
			asiCase("foo = this.return\nArray()", "Array()", true),
			asiCase("var yield = bar.yield\nArray()", "Array()", true),
			asiCase("var foo = { bar: baz }\nArray()", "Array()", true),

			// ---- No semicolon required before array literal because ASI does not occur ----
			asiCase("Array()", "Array()", false),
			asiCase("{}\nArray()", "Array()", false),
			asiCase("function foo() {}\nArray()", "Array()", false),
			asiCase("class Foo {}\nArray()", "Array()", false),
			asiCase("foo: Array();", "Array()", false),
			asiCase("foo();Array();", "Array()", false),
			asiCase("{ Array(); }", "Array()", false),
			asiCase("if (a) Array();", "Array()", false),
			asiCase("if (a); else Array();", "Array()", false),
			asiCase("while (a) Array();", "Array()", false),
			asiCase("do Array();\nwhile (a);", "Array()", false),
			asiCase("for (let i = 0; i < 10; i++) Array();", "Array()", false),
			asiCase("for (const prop in obj) Array();", "Array()", false),
			asiCase("for (const element of iterable) Array();", "Array()", false),
			asiCase("with (obj) Array();", "Array()", false),

			// ---- No semicolon required before array literal because ASI still occurs ----
			asiCase("const foo = () => {}\nArray()", "Array()", false),
			asiCase("a++\nArray()", "Array()", false),
			asiCase("a--\nArray()", "Array()", false),
			asiCase("function foo() {\n    return\n    Array();\n}", "Array()", false),
			asiCase("function * foo() {\n    yield\n    Array();\n}", "Array()", false),
			asiCase("do {}\nwhile (a)\nArray()", "Array()", false),
			asiCase("debugger\nArray()", "Array()", false),
			asiCase("for (;;) {\n    break\n    Array()\n}", "Array()", false),
			asiCase("for (;;) {\n    continue\n    Array()\n}", "Array()", false),
			asiCase("foo: break foo\nArray()", "Array()", false),
			asiCase("foo: while (true) continue foo\nArray()", "Array()", false),
			asiCase("const foo = bar\nexport { foo }\nArray()", "Array()", false),
			asiCase("export { foo } from 'bar'\nArray()", "Array()", false),
			asiCase(`export { foo } from 'bar' with { type: "json" }
Array()`, "Array()", false),
			asiCase("export * as foo from 'bar'\nArray()", "Array()", false),
			asiCase(`export * as foo from 'bar' with { type: "json" }
Array()`, "Array()", false),
			asiCase("import foo from 'bar'\nArray()", "Array()", false),
			asiCase(`import foo from 'bar' with { type: "json" }
Array()`, "Array()", false),
			asiCase("var yield = 5;\n\nyield: while (foo) {\n    if (bar)\n        break yield\n    new Array();\n}", "new Array()", false),
			asiCase("var foo\nArray()", "Array()", false),
			asiCase("let bar\nArray()", "Array()", false),

			// ---- Comment preservation around the callee / parentheses ----
			directFixCase("/*a*/Array()", "Array()", "/*a*/[]"),
			directFixCase("/*a*/Array()/*b*/", "Array()", "/*a*/[]/*b*/"),
			suggestCase("Array/*a*/()", "Array/*a*/()", "useLiteral", "[]"),
			suggestCase(
				"/*a*//*b*/Array/*c*//*d*/()/*e*//*f*/;/*g*//*h*/",
				"Array/*c*//*d*/()",
				"useLiteral",
				"/*a*//*b*/[]/*e*//*f*/;/*g*//*h*/",
			),
			directFixCase("Array(/*a*/ /*b*/)", "Array(/*a*/ /*b*/)", "[/*a*/ /*b*/]"),
			directFixCase(
				"Array(/*a*/ x /*b*/, /*c*/ y /*d*/)",
				"Array(/*a*/ x /*b*/, /*c*/ y /*d*/)",
				"[/*a*/ x /*b*/, /*c*/ y /*d*/]",
			),
			directFixCase(
				"/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)/*f*/;/*g*/",
				"Array(/*b*/ x /*c*/, /*d*/ y /*e*/)",
				"/*a*/[/*b*/ x /*c*/, /*d*/ y /*e*/]/*f*/;/*g*/",
			),
			directFixCase("/*a*/new Array", "new Array", "/*a*/[]"),
			directFixCase("/*a*/new Array/*b*/", "new Array", "/*a*/[]/*b*/"),
			suggestCase("new/*a*/Array", "new/*a*/Array", "useLiteral", "[]"),
			suggestCase(
				"new/*a*//*b*/Array/*c*//*d*/()/*e*//*f*/;/*g*//*h*/",
				"new/*a*//*b*/Array/*c*//*d*/()",
				"useLiteral",
				"[]/*e*//*f*/;/*g*//*h*/",
			),
			directFixCase("new Array(/*a*/ /*b*/)", "new Array(/*a*/ /*b*/)", "[/*a*/ /*b*/]"),
			directFixCase(
				"new Array(/*a*/ x /*b*/, /*c*/ y /*d*/)",
				"new Array(/*a*/ x /*b*/, /*c*/ y /*d*/)",
				"[/*a*/ x /*b*/, /*c*/ y /*d*/]",
			),
			suggestCase(
				"new/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)/*f*/;/*g*/",
				"new/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)",
				"useLiteral",
				"[/*b*/ x /*c*/, /*d*/ y /*e*/]/*f*/;/*g*/",
			),
			suggestCase("// a\nArray // b\n()", "Array // b\n()", "useLiteral", "// a\n[]"),
			suggestCase("// a\nArray // b\n() // c", "Array // b\n()", "useLiteral", "// a\n[] // c"),
			suggestCase("new // a\nArray // b\n()", "new // a\nArray // b\n()", "useLiteral", "[]"),
			suggestCase("new (Array /* a */);", "new (Array /* a */)", "useLiteral", "[];"),
			suggestCase("(/* a */ Array)(1, 2, 3);", "(/* a */ Array)(1, 2, 3)", "useLiteral", "[1, 2, 3];"),
			suggestCase("(Array /* a */)(1, 2, 3);", "(Array /* a */)(1, 2, 3)", "useLiteral", "[1, 2, 3];"),
			suggestCase("(Array) /* a */ (1, 2, 3);", "(Array) /* a */ (1, 2, 3)", "useLiteral", "[1, 2, 3];"),
			suggestCase("(/* a */(Array))();", "(/* a */(Array))()", "useLiteral", "[];"),
			suggestCase(
				"Array?.(0, 1, 2).forEach(doSomething);",
				"Array?.(0, 1, 2)",
				"useLiteral",
				"[0, 1, 2].forEach(doSomething);",
			),

			// ---- ruleTesterTypeScript: explicit cases ----
			directFixCase(`new Array();`, `new Array()`, `[];`),
			directFixCase(`Array();`, `Array()`, `[];`),
			directFixCase(`new Array(x, y);`, `new Array(x, y)`, `[x, y];`),
			directFixCase(`Array(x, y);`, `Array(x, y)`, `[x, y];`),
			directFixCase(`new Array(0, 1, 2);`, `new Array(0, 1, 2)`, `[0, 1, 2];`),
			directFixCase(`Array(0, 1, 2);`, `Array(0, 1, 2)`, `[0, 1, 2];`),
			suggestCase(`Array?.(0, 1, 2);`, `Array?.(0, 1, 2)`, "useLiteral", `[0, 1, 2];`),
			suggestCase(`Array?.(x, y);`, `Array?.(x, y)`, "useLiteral", `[x, y];`),
			suggestCase(`Array /*a*/ ?.();`, `Array /*a*/ ?.()`, "useLiteral", `[];`),
			suggestCase(`Array?./*a*/();`, `Array?./*a*/()`, "useLiteral", `[];`),

			// ---- No semicolon required after TypeScript syntax ----
			// The five `type T = ...` cases, the six function-overload/ambient
			// cases, and the three `import ... =` cases each lock in the
			// documented divergence described on tsAsiCase: rslint adds a
			// redundant `;` that upstream omits. The uninitialized-declarator
			// cases (`const foo`, `declare const foo`, `let foo: bar`) match
			// upstream exactly since that path is fully implemented.
			tsAsiCase("type T = Foo", true),
			tsAsiCase("type T = Foo<Bar>", true),
			tsAsiCase("type T = (A | B)", true),
			tsAsiCase("type T = -1", true),
			tsAsiCase("type T = 'foo'", true),
			tsAsiCase("const foo", false),
			tsAsiCase("declare const foo", false),
			tsAsiCase("function foo()", true),
			tsAsiCase("declare function foo()", true),
			tsAsiCase("function foo(): []", true),
			tsAsiCase("declare function foo(): []", true),
			tsAsiCase("function foo(): (Foo)", true),
			tsAsiCase("declare function foo(): (Foo)", true),
			tsAsiCase("let foo: bar", false),
			tsAsiCase("import Foo = require('foo')", true),
			tsAsiCase("import Foo = Bar", true),
			tsAsiCase("import Foo = Bar.Baz.Qux", true),

			// ---- Multiple reports in one file, mixed semicolon requirement ----
			// The second report in each case (`Array() // ";" not required`)
			// follows an `as Fn` / `as Object` type-cast identifier — the same
			// documented divergence as the tsAsiCase block above, so rslint's
			// fix adds a `;` there too.
			{
				Code: "(function () {\n\tFn\n\tArray() // \";\" required\n}) as Fn\nArray() // \";\" not required",
				Output: []string{
					"(function () {\n\tFn\n\t;[] // \";\" required\n}) as Fn\n;[] // \";\" not required",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 3, Column: 2, EndLine: 3, EndColumn: 9},
					{MessageId: "preferLiteral", Line: 5, Column: 1, EndLine: 5, EndColumn: 8},
				},
			},
			{
				Code: "({\n\tfoo() {\n\t\tObject\n\t\tArray() // \";\" required\n\t}\n}) as Object\nArray() // \";\" not required",
				Output: []string{
					"({\n\tfoo() {\n\t\tObject\n\t\t;[] // \";\" required\n\t}\n}) as Object\n;[] // \";\" not required",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 4, Column: 3, EndLine: 4, EndColumn: 10},
					{MessageId: "preferLiteral", Line: 7, Column: 1, EndLine: 7, EndColumn: 8},
				},
			},
		},
	)
}
