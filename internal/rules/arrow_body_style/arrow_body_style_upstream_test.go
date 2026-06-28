// TestArrowBodyStyleUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/arrow-body-style.js 1:1. Position assertions cover line/column
// for every invalid case (and endLine/endColumn where upstream asserts them).
// rslint-specific lock-in cases live in the arrow_body_style_extras_test.go file.
package arrow_body_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrowBodyStyleUpstream(t *testing.T) {
	asNeeded := []any{"as-needed"}
	always := []any{"always"}
	never := []any{"never"}
	requireReturn := []any{"as-needed", map[string]any{"requireReturnForObjectLiteral": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ArrowBodyStyleRule,
		[]rule_tester.ValidTestCase{
			{Code: "var foo = () => {};"},
			{Code: "var foo = () => 0;"},
			{Code: "var addToB = (a) => { b =  b + a };"},
			{Code: "var foo = () => { /* do nothing */ };"},
			{Code: "var foo = () => {\n /* do nothing */ \n};"},
			{Code: "var foo = (data, name) => {\ndata[name] = true;\nreturn data;\n};"},
			{Code: "var foo = () => ({});"},
			{Code: "var foo = () => bar();"},
			{Code: "var foo = () => { bar(); };"},
			{Code: "var foo = () => { b = a };"},
			{Code: "var foo = () => { bar: 1 };"},
			{Code: "var foo = () => { return 0; };", Options: always},
			{Code: "var foo = () => { return bar(); };", Options: always},
			{Code: "var foo = () => 0;", Options: never},
			{Code: "var foo = () => ({ foo: 0 });", Options: never},
			{Code: "var foo = () => {};", Options: requireReturn},
			{Code: "var foo = () => 0;", Options: requireReturn},
			{Code: "var addToB = (a) => { b =  b + a };", Options: requireReturn},
			{Code: "var foo = () => { /* do nothing */ };", Options: requireReturn},
			{Code: "var foo = () => {\n /* do nothing */ \n};", Options: requireReturn},
			{Code: "var foo = (data, name) => {\ndata[name] = true;\nreturn data;\n};", Options: requireReturn},
			{Code: "var foo = () => bar();", Options: requireReturn},
			{Code: "var foo = () => { bar(); };", Options: requireReturn},
			{Code: "var foo = () => { return { bar: 0 }; };", Options: requireReturn},
		},
		[]rule_tester.InvalidTestCase{
			// ---- as-needed / default: `in`-operator inside for-loop init ----
			{
				Code:    "for (var foo = () => { return a in b ? bar : () => {} } ;;);",
				Output:  []string{"for (var foo = () => (a in b ? bar : () => {}) ;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 22, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "a in b; for (var f = () => { return c };;);",
				Output:  []string{"a in b; for (var f = () => c;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 28, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (a = b => { return c in d ? e : f } ;;);",
				Output:  []string{"for (a = b => (c in d ? e : f) ;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (var f = () => { return a };;);",
				Output:  []string{"for (var f = () => a;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 20, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (var f;f = () => { return a };);",
				Output:  []string{"for (var f;f = () => a;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 22, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (var f = () => { return a in c };;);",
				Output:  []string{"for (var f = () => (a in c);;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 20, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (var f;f = () => { return a in c };);",
				Output:  []string{"for (var f;f = () => a in c;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 22, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (;;){var f = () => { return a in c }}",
				Output:  []string{"for (;;){var f = () => a in c}"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 24, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (a = b => { return c = d in e } ;;);",
				Output:  []string{"for (a = b => (c = d in e) ;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "for (var a;;a = b => { return c = d in e } );",
				Output:  []string{"for (var a;;a = b => c = d in e );"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 22, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for (let a = (b, c, d) => { return vb && c in d; }; ;);",
				Output: []string{"for (let a = (b, c, d) => (vb && c in d); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for (let a = (b, c, d) => { return v in b && c in d; }; ;);",
				Output: []string{"for (let a = (b, c, d) => (v in b && c in d); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "function foo(){ for (let a = (b, c, d) => { return v in b && c in d; }; ;); }",
				Output: []string{"function foo(){ for (let a = (b, c, d) => (v in b && c in d); ;); }"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 43, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for ( a = (b, c, d) => { return v in b && c in d; }; ;);",
				Output: []string{"for ( a = (b, c, d) => (v in b && c in d); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 24, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for ( a = (b) => { return (c in d) }; ;);",
				Output: []string{"for ( a = (b) => (c in d); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for (let a = (b, c, d) => { return vb in dd ; }; ;);",
				Output: []string{"for (let a = (b, c, d) => (vb in dd ); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "for (let a = (b, c, d) => { return vb in c in dd ; }; ;);",
				Output: []string{"for (let a = (b, c, d) => (vb in c in dd ); ;);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "do{let a = () => {return f in ff}}while(true){}",
				Output: []string{"do{let a = () => f in ff}while(true){}"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "do{for (let a = (b, c, d) => { return vb in c in dd ; }; ;);}while(true){}",
				Output: []string{"do{for (let a = (b, c, d) => (vb in c in dd ); ;);}while(true){}"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 30, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "scores.map(score => { return x in +(score / maxScore).toFixed(2)});",
				Output: []string{"scores.map(score => x in +(score / maxScore).toFixed(2));"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 21, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "const fn = (a, b) => { return a + x in Number(b) };",
				Output: []string{"const fn = (a, b) => a + x in Number(b);"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 22, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- always: expected block ----
			{
				Code:    "var foo = () => 0",
				Output:  []string{"var foo = () => {return 0}"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, EndLine: 1, EndColumn: 18, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => 0;",
				Output:  []string{"var foo = () => {return 0};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => ({});",
				Output:  []string{"var foo = () => {return {}};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => (  {});",
				Output:  []string{"var foo = () => {return   {}};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 20, MessageId: "expectedBlock"}},
			},
			{
				Code:    "(() => ({}))",
				Output:  []string{"(() => {return {}})"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 9, MessageId: "expectedBlock"}},
			},
			{
				Code:    "(() => ( {}))",
				Output:  []string{"(() => {return  {}})"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 10, MessageId: "expectedBlock"}},
			},
			// ---- as-needed: unexpected single/object block ----
			{
				Code:    "var foo = () => { return 0; };",
				Output:  []string{"var foo = () => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return 0 };",
				Output:  []string{"var foo = () => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return bar(); };",
				Output:  []string{"var foo = () => bar();"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => {};",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedEmptyBlock"}},
			},
			{
				Code:    "var foo = () => {\nreturn 0;\n};",
				Output:  []string{"var foo = () => 0;"},
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return { bar: 0 }; };",
				Output:  []string{"var foo = () => ({ bar: 0 });"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedObjectBlock"}},
			},
			{
				Code:    "var foo = () => { return ({ bar: 0 }); };",
				Output:  []string{"var foo = () => ({ bar: 0 });"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "var foo = () => { return a, b }",
				Output: []string{"var foo = () => (a, b)"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return };",
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return; };",
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return ( /* a */ {ok: true} /* b */ ) };",
				Output:  []string{"var foo = () => ( /* a */ {ok: true} /* b */ );"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return '{' };",
				Output:  []string{"var foo = () => '{';"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return { bar: 0 }.bar; };",
				Output:  []string{"var foo = () => ({ bar: 0 }.bar);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedObjectBlock"}},
			},
			{
				Code:    "var foo = (data, name) => {\ndata[name] = true;\nreturn data;\n};",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedOtherBlock"}},
			},
			{
				Code:    "var foo = () => { bar };",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedOtherBlock"}},
			},
			{
				Code:    "var foo = () => { return 0; };",
				Output:  []string{"var foo = () => 0;"},
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => { return bar(); };",
				Output:  []string{"var foo = () => bar();"},
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = () => ({});",
				Output:  []string{"var foo = () => {return {}};"},
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => ({ bar: 0 });",
				Output:  []string{"var foo = () => {return { bar: 0 }};"},
				Options: requireReturn,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => (((((((5)))))));",
				Output:  []string{"var foo = () => {return (((((((5)))))))};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 24, MessageId: "expectedBlock"}},
			},
			{
				// Not fixed; fixing would cause ASI issues.
				Code:    "var foo = () => { return bar }\n[1, 2, 3].map(foo)",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				// Not fixed; fixing would cause ASI issues.
				Code:    "var foo = () => { return bar }\n(1).toString();",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				// Fixing here is ok because the arrow function has a semicolon afterwards.
				Code:    "var foo = () => { return bar };\n[1, 2, 3].map(foo)",
				Output:  []string{"var foo = () => bar;\n[1, 2, 3].map(foo)"},
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = /* a */ ( /* b */ ) /* c */ => /* d */ { /* e */ return /* f */ 5 /* g */ ; /* h */ } /* i */ ;",
				Output:  []string{"var foo = /* a */ ( /* b */ ) /* c */ => /* d */  /* e */  /* f */ 5 /* g */  /* h */  /* i */ ;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 50, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var foo = /* a */ ( /* b */ ) /* c */ => /* d */ ( /* e */ 5 /* f */ ) /* g */ ;",
				Output:  []string{"var foo = /* a */ ( /* b */ ) /* c */ => /* d */ {return ( /* e */ 5 /* f */ )} /* g */ ;"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 60, MessageId: "expectedBlock"}},
			},
			{
				Code:   "var foo = () => {\nreturn bar;\n};",
				Output: []string{"var foo = () => bar;"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, EndLine: 3, EndColumn: 2, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "var foo = () => {\nreturn bar;};",
				Output: []string{"var foo = () => bar;"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, EndLine: 2, EndColumn: 13, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:   "var foo = () => {return bar;\n};",
				Output: []string{"var foo = () => bar;"},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, EndLine: 2, EndColumn: 2, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code: "\n              var foo = () => {\n                return foo\n                  .bar;\n              };\n            ",
				Output: []string{
					"\n              var foo = () => foo\n                  .bar;\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 2, Column: 31, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code: "\n              var foo = () => {\n                return {\n                  bar: 1,\n                  baz: 2\n                };\n              };\n            ",
				Output: []string{
					"\n              var foo = () => ({\n                  bar: 1,\n                  baz: 2\n                });\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{Line: 2, Column: 31, EndLine: 7, EndColumn: 16, MessageId: "unexpectedObjectBlock"}},
			},
			{
				Code:    "var foo = () => ({foo: 1}).foo();",
				Output:  []string{"var foo = () => {return {foo: 1}.foo()};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => ({foo: 1}.foo());",
				Output:  []string{"var foo = () => {return {foo: 1}.foo()};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => ( {foo: 1} ).foo();",
				Output:  []string{"var foo = () => {return  {foo: 1} .foo()};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
			{
				Code: "\n              var foo = () => ({\n                  bar: 1,\n                  baz: 2\n                });\n            ",
				Output: []string{
					"\n              var foo = () => {return {\n                  bar: 1,\n                  baz: 2\n                }};\n            ",
				},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
			{
				Code: "\n              parsedYears = _map(years, (year) => (\n                  {\n                      index : year,\n                      title : splitYear(year)\n                  }\n              ));\n            ",
				Output: []string{
					"\n              parsedYears = _map(years, (year) => {\n                  return {\n                      index : year,\n                      title : splitYear(year)\n                  }\n              });\n            ",
				},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
			// https://github.com/eslint/eslint/issues/14633
			{
				Code:    "const createMarker = (color) => ({ latitude, longitude }, index) => {};",
				Output:  []string{"const createMarker = (color) => {return ({ latitude, longitude }, index) => {}};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedBlock"}},
			},
		},
	)
}
