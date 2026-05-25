// TestBraceStyleUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/brace-style/brace-style._js_.test.ts and
// brace-style._ts_.test.ts 1:1. Position assertions cover line/column for
// every invalid case. rslint-specific lock-in cases live in
// brace_style_extras_test.go.
//
// `with (...) { ... }` cases from the TS suite ARE included — tsgo parses
// the construct even under strict mode (the strict-mode rejection is a
// downstream diagnostic, not a parse failure), and the WithStatement node's
// Block child is checked by the same KindBlock listener that handles every
// other statement body.
package brace_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/brace_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func opts1tbs() []any       { return []any{"1tbs"} }
func optsStroustrup() []any { return []any{"stroustrup"} }
func optsAllman() []any     { return []any{"allman"} }
func opts1tbsSingle() []any {
	return []any{"1tbs", map[string]any{"allowSingleLine": true}}
}
func optsStroustrupSingle() []any {
	return []any{"stroustrup", map[string]any{"allowSingleLine": true}}
}
func optsAllmanSingle() []any {
	return []any{"allman", map[string]any{"allowSingleLine": true}}
}

func TestBraceStyleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&brace_style.BraceStyleRule,
		[]rule_tester.ValidTestCase{
			// ---- default (1tbs) baseline ----
			{Code: "function f() {\n    if (true)\n        return {x: 1}\n    else {\n        var y = 2\n        return y\n    }\n}"},
			{Code: "if (tag === 1) glyph.id = pbf.readVarint();\nelse if (tag === 2) glyph.bitmap = pbf.readBytes();"},
			{Code: "function foo () {\n  return;\n}"},
			{Code: "function a(b,\nc,\nd) { }"},
			{Code: "!function foo () {\n  return;\n}"},
			{Code: "!function a(b,\nc,\nd) { }"},
			{Code: "if (foo) {\n  bar();\n}"},
			{Code: "if (a) {\n  b();\n} else {\n  c();\n}"},
			{Code: "while (foo) {\n  bar();\n}"},
			{Code: "for (;;) {\n  bar();\n}"},
			{Code: "switch (foo) {\n  case 'bar': break;\n}"},
			{Code: "try {\n  bar();\n} catch (e) {\n  baz();\n}"},
			{Code: "do {\n  bar();\n} while (true)"},
			{Code: "for (foo in bar) {\n  baz();\n}"},
			{Code: "if (a &&\n  b &&\n  c) {\n}"},
			{Code: "switch(0) {\n}"},
			{Code: "class Foo {\n}"},
			{Code: "(class {\n})"},
			{Code: "class\nFoo {\n}"},
			{Code: "class Foo {\n    bar() {\n    }\n}"},

			// ---- stroustrup ----
			{Code: "if (foo) {\n}\nelse {\n}", Options: optsStroustrup()},
			{Code: "try {\n  bar();\n}\ncatch (e) {\n  baz();\n}", Options: optsStroustrup()},
			{Code: "if (tag === 1) fontstack.name = pbf.readString();\nelse if (tag === 2) fontstack.range = pbf.readString();\nelse if (tag === 3) {\n  var glyph = pbf.readMessage(readGlyph, {});\n  fontstack.glyphs[glyph.id] = glyph;\n}", Options: optsStroustrup()},
			{Code: "class Foo {\n}", Options: optsStroustrup()},
			{Code: "(class {\n})", Options: optsStroustrup()},

			// ---- allman ----
			{Code: "if (foo)\n{\n}\nelse\n{\n}", Options: optsAllman()},
			{Code: "try\n{\n  bar();\n}\ncatch (e)\n{\n  baz();\n}", Options: optsAllman()},
			{Code: "switch(x)\n{\n  case 1:\n    bar();\n}", Options: optsAllman()},
			{Code: "class Foo\n{\n}", Options: optsAllman()},
			{Code: "(class\n{\n})", Options: optsAllman()},
			{Code: "class\nFoo\n{\n}", Options: optsAllman()},

			// ---- 1tbs + allowSingleLine: true ----
			{Code: "function foo () { return; }", Options: opts1tbsSingle()},
			{Code: "function foo () { a(); b(); return; }", Options: opts1tbsSingle()},
			{Code: "function a(b,c,d) { }", Options: opts1tbsSingle()},
			{Code: "!function foo () { return; }", Options: opts1tbsSingle()},
			{Code: "!function a(b,c,d) { }", Options: opts1tbsSingle()},
			{Code: "if (foo) {  bar(); }", Options: opts1tbsSingle()},
			{Code: "if (a) { b(); } else { c(); }", Options: opts1tbsSingle()},
			{Code: "while (foo) {  bar(); }", Options: opts1tbsSingle()},
			{Code: "for (;;) {  bar(); }", Options: opts1tbsSingle()},
			{Code: "switch (foo) {  case \"bar\": break; }", Options: opts1tbsSingle()},
			{Code: "try {  bar(); } catch (e) { baz();  }", Options: opts1tbsSingle()},
			{Code: "do {  bar(); } while (true)", Options: opts1tbsSingle()},
			{Code: "for (foo in bar) {  baz();  }", Options: opts1tbsSingle()},
			{Code: "if (a && b && c) {  }", Options: opts1tbsSingle()},
			{Code: "switch(0) {}", Options: opts1tbsSingle()},
			{Code: "if (foo) {}\nelse {}", Options: optsStroustrupSingle()},
			{Code: "try {  bar(); }\ncatch (e) { baz();  }", Options: optsStroustrupSingle()},
			{Code: "var foo = () => { return; }", Options: optsStroustrupSingle()},
			{Code: "if (foo) {}\nelse {}", Options: optsAllmanSingle()},
			{Code: "try {  bar(); }\ncatch (e) { baz();  }", Options: optsAllmanSingle()},
			{Code: "var foo = () => { return; }", Options: optsAllmanSingle()},
			{Code: "if (foo) { baz(); } else {\n  boom();\n}", Options: opts1tbsSingle()},
			{Code: "if (foo) { baz(); } else if (bar) {\n  boom();\n}", Options: opts1tbsSingle()},
			{Code: "if (foo) { baz(); } else\nif (bar) {\n  boom();\n}", Options: opts1tbsSingle()},
			{Code: "try { somethingRisky(); } catch(e) {\n  handleError();\n}", Options: opts1tbsSingle()},
			{Code: "if (tag === 1) fontstack.name = pbf.readString();\nelse if (tag === 2) fontstack.range = pbf.readString();\nelse if (tag === 3) {\n  var glyph = pbf.readMessage(readGlyph, {});\n  fontstack.glyphs[glyph.id] = glyph;\n}"},
			{Code: "switch(x) {}", Options: optsAllmanSingle()},
			{Code: "class Foo {}", Options: opts1tbsSingle()},
			{Code: "class Foo {}", Options: optsAllmanSingle()},
			{Code: "(class {})", Options: opts1tbsSingle()},
			{Code: "(class {})", Options: optsAllmanSingle()},

			// ---- standalone / Program / SwitchCase block parents — skipped ----
			// https://github.com/eslint/eslint/issues/7908
			{Code: "{}"},
			{Code: "if (foo) {\n}\n{\n}"},
			{Code: "switch (foo) {\n  case bar:\n    baz();\n    {\n      qux();\n    }\n}"},
			{Code: "{\n}"},
			{Code: "{\n  {\n  }\n}"},

			// https://github.com/eslint/eslint/issues/7974
			{Code: "class Ball {\n  throw() {}\n  catch() {}\n}"},
			{Code: "({\n  and() {},\n  finally() {}\n})"},
			{Code: "(class {\n  or() {}\n  else() {}\n})"},
			{Code: "if (foo) bar = function() {}\nelse baz()"},

			// ---- class static blocks ----
			{Code: "class C {\n    static {\n        foo;\n    }\n}", Options: opts1tbs()},
			{Code: "class C {\n    static {}\n\n    static {\n    }\n}", Options: opts1tbs()},
			{Code: "class C {\n    static { foo; }\n}", Options: opts1tbsSingle()},
			{Code: "class C {\n    static {\n        foo;\n    }\n}", Options: optsStroustrup()},
			{Code: "class C {\n    static {}\n\n    static {\n    }\n}", Options: optsStroustrup()},
			{Code: "class C {\n    static { foo; }\n}", Options: optsStroustrupSingle()},
			{Code: "class C\n{\n    static\n    {\n        foo;\n    }\n}", Options: optsAllman()},
			{Code: "class C\n{\n    static\n    {}\n}", Options: optsAllman()},
			{Code: "class C\n{\n    static {}\n\n    static { foo; }\n\n    static\n    { foo; }\n}", Options: optsAllmanSingle()},
			{Code: "class C {\n    static {\n        {\n            foo;\n        }\n    }\n}", Options: opts1tbs()},

			// ---- TS `with` statement (parsed even under strict mode) ----
			{Code: "with (foo) {\n  bar();\n}"},
			{Code: "with (foo) {  bar(); }", Options: opts1tbsSingle()},

			// ---- TS namespace/module bodies ----
			{Code: "module \"Foo\" {\n}", Options: opts1tbs()},
			{Code: "module \"Foo\" {\n}", Options: optsStroustrup()},
			{Code: "module \"Foo\"\n{\n}", Options: optsAllman()},
			{Code: "namespace Foo {\n}", Options: opts1tbs()},
			{Code: "namespace Foo {\n}", Options: optsStroustrup()},
			{Code: "namespace Foo\n{\n}", Options: optsAllman()},
		},
		[]rule_tester.InvalidTestCase{
			// ---- nextLineClose (1tbs, } not on same line as keyword) ----
			{
				Code:   "if (f) {\n  bar;\n}\nelse\n  baz;",
				Output: []string{"if (f) {\n  bar;\n} else\n  baz;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
				},
			},

			// ---- blockSameLine + singleLineClose (single-line block in multiline mode) ----
			{
				Code:   "var foo = () => { return; }",
				Output: []string{"var foo = () => {\n return; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 17},
					{MessageId: "singleLineClose", Line: 1, Column: 27},
				},
			},
			{
				Code:   "function foo() { return; }",
				Output: []string{"function foo() {\n return; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 16},
					{MessageId: "singleLineClose", Line: 1, Column: 26},
				},
			},

			// ---- nextLineOpen + singleLineClose ----
			{
				Code:   "function foo() \n { \n return; }",
				Output: []string{"function foo() { \n return; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 10},
				},
			},
			{
				Code:   "!function foo() \n { \n return; }",
				Output: []string{"!function foo() { \n return; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 10},
				},
			},
			{
				Code:   "if (foo) \n { \n bar(); }",
				Output: []string{"if (foo) { \n bar(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 9},
				},
			},
			{
				Code:   "if (a) { \nb();\n } else \n { c(); }",
				Output: []string{"if (a) { \nb();\n } else {\n c(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 4, Column: 2},
					{MessageId: "blockSameLine", Line: 4, Column: 2},
					{MessageId: "singleLineClose", Line: 4, Column: 9},
				},
			},
			{
				Code:   "while (foo) \n { \n bar(); }",
				Output: []string{"while (foo) { \n bar(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 9},
				},
			},
			{
				Code:   "for (;;) \n { \n bar(); }",
				Output: []string{"for (;;) { \n bar(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 9},
				},
			},
			{
				Code:   "switch (foo) \n { \n case \"bar\": break; }",
				Output: []string{"switch (foo) { \n case \"bar\": break; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 21},
				},
			},
			{
				Code:   "switch (foo) \n { }",
				Output: []string{"switch (foo) { }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "try \n { \n bar(); \n } catch (e) {}",
				Output: []string{"try { \n bar(); \n } catch (e) {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) \n {}",
				Output: []string{"try { \n bar(); \n } catch (e) {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 4, Column: 2},
				},
			},
			{
				Code:   "do \n { \n bar(); \n} while (true)",
				Output: []string{"do { \n bar(); \n} while (true)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "for (foo in bar) \n { \n baz(); \n }",
				Output: []string{"for (foo in bar) { \n baz(); \n }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "for (foo of bar) \n { \n baz(); \n }",
				Output: []string{"for (foo of bar) { \n baz(); \n }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},

			// ---- nextLineClose on try/catch/finally ----
			{
				Code:   "try { \n bar(); \n }\ncatch (e) {\n}",
				Output: []string{"try { \n bar(); \n } catch (e) {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
				Output: []string{"try { \n bar(); \n } catch (e) {\n} finally {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 4, Column: 1},
				},
			},
			{
				Code:   "if (a) { \nb();\n } \n else { \nc();\n }",
				Output: []string{"if (a) { \nb();\n } else { \nc();\n }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 2},
				},
			},

			// ---- stroustrup sameLineClose ----
			{
				Code:   "try { \n bar(); \n }\ncatch (e) {\n} finally {\n}",
				Output: []string{"try { \n bar(); \n }\ncatch (e) {\n}\n finally {\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 5, Column: 1},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
				Output: []string{"try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "if (a) { \nb();\n } else { \nc();\n }",
				Output: []string{"if (a) { \nb();\n }\n else { \nc();\n }"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}",
				Output: []string{"if (foo) {\nbaz();\n}\n else if (bar) {\nbaz();\n}\nelse {\nqux();\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
				},
			},
			{
				Code:   "if (foo) {\npoop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
				Output: []string{"if (foo) {\npoop();\n} \nelse if (bar) {\nbaz();\n}\n else if (thing) {\nboom();\n}\nelse {\nqux();\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 6, Column: 1},
				},
			},

			// ---- allman sameLineOpen / sameLineClose ----
			{
				Code:   "try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}",
				Output: []string{"try \n{ \n bar(); \n }\n catch (e) \n{\n}\n finally \n{\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 5},
					{MessageId: "sameLineOpen", Line: 4, Column: 12},
					{MessageId: "sameLineOpen", Line: 6, Column: 10},
				},
			},
			{
				Code:   "switch(x) { case 1: \nbar(); }\n ",
				Output: []string{"switch(x) \n{\n case 1: \nbar(); \n}\n "},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 11},
					{MessageId: "blockSameLine", Line: 1, Column: 11},
					{MessageId: "singleLineClose", Line: 2, Column: 8},
				},
			},
			{
				Code:   "if (a) { \nb();\n } else { \nc();\n }",
				Output: []string{"if (a) \n{ \nb();\n }\n else \n{ \nc();\n }"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 8},
					{MessageId: "sameLineClose", Line: 3, Column: 2},
					{MessageId: "sameLineOpen", Line: 3, Column: 9},
				},
			},
			{
				Code:   "if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}",
				Output: []string{"if (foo) \n{\nbaz();\n}\n else if (bar) \n{\nbaz();\n}\nelse \n{\nqux();\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 10},
					{MessageId: "sameLineClose", Line: 3, Column: 1},
					{MessageId: "sameLineOpen", Line: 3, Column: 17},
					{MessageId: "sameLineOpen", Line: 6, Column: 6},
				},
			},
			{
				Code:   "if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
				Output: []string{"if (foo)\n{\n poop();\n} \nelse if (bar) \n{\nbaz();\n}\n else if (thing) \n{\nboom();\n}\nelse \n{\nqux();\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 1},
					{MessageId: "sameLineOpen", Line: 4, Column: 15},
					{MessageId: "sameLineClose", Line: 6, Column: 1},
					{MessageId: "sameLineOpen", Line: 6, Column: 19},
					{MessageId: "sameLineOpen", Line: 9, Column: 6},
				},
			},
			{
				Code:   "if (foo)\n{\n  bar(); }",
				Output: []string{"if (foo)\n{\n  bar(); \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 10},
				},
			},
			{
				Code:   "try\n{\n  somethingRisky();\n} catch (e)\n{\n  handleError()\n}",
				Output: []string{"try\n{\n  somethingRisky();\n}\n catch (e)\n{\n  handleError()\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 4, Column: 1},
				},
			},

			// ---- 1tbs + allowSingleLine ----
			{
				Code:   "function foo() { return; \n}",
				Output: []string{"function foo() {\n return; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 16},
				},
			},
			{
				Code:   "function foo() { a(); b(); return; \n}",
				Output: []string{"function foo() {\n a(); b(); return; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 16},
				},
			},
			{
				Code:   "function foo() { \n return; }",
				Output: []string{"function foo() { \n return; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 10},
				},
			},
			{
				Code:   "function foo() {\na();\nb();\nreturn; }",
				Output: []string{"function foo() {\na();\nb();\nreturn; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 4, Column: 9},
				},
			},
			{
				Code:   "!function foo() { \n return; }",
				Output: []string{"!function foo() { \n return; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 10},
				},
			},
			{
				Code:   "if (a) { b();\n } else { c(); }",
				Output: []string{"if (a) {\n b();\n } else { c(); }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 8},
				},
			},
			{
				Code:   "if (a) { b(); }\nelse { c(); }",
				Output: []string{"if (a) { b(); } else { c(); }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 1, Column: 15},
				},
			},
			{
				Code:   "while (foo) { \n bar(); }",
				Output: []string{"while (foo) { \n bar(); \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 9},
				},
			},
			{
				Code:   "for (;;) { bar(); \n }",
				Output: []string{"for (;;) {\n bar(); \n }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 10},
				},
			},
			{
				Code:   "switch (foo) \n { \n case \"bar\": break; }",
				Output: []string{"switch (foo) { \n case \"bar\": break; \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 21},
				},
			},
			{
				Code:   "switch (foo) \n { }",
				Output: []string{"switch (foo) { }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "try {  bar(); }\ncatch (e) { baz();  }",
				Output: []string{"try {  bar(); } catch (e) { baz();  }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 1, Column: 15},
				},
			},
			{
				Code:   "try \n { \n bar(); \n } catch (e) {}",
				Output: []string{"try { \n bar(); \n } catch (e) {}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) \n {}",
				Output: []string{"try { \n bar(); \n } catch (e) {}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 4, Column: 2},
				},
			},
			{
				Code:   "do \n { \n bar(); \n} while (true)",
				Output: []string{"do { \n bar(); \n} while (true)"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "for (foo in bar) \n { \n baz(); \n }",
				Output: []string{"for (foo in bar) { \n baz(); \n }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
				},
			},
			{
				Code:   "try { \n bar(); \n }\ncatch (e) {\n}",
				Output: []string{"try { \n bar(); \n } catch (e) {\n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
				Output: []string{"try { \n bar(); \n } catch (e) {\n} finally {\n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 4, Column: 1},
				},
			},
			{
				Code:   "if (a) { \nb();\n } \n else { \nc();\n }",
				Output: []string{"if (a) { \nb();\n } else { \nc();\n }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 2},
				},
			},

			// ---- stroustrup + allowSingleLine ----
			{
				Code:   "try { \n bar(); \n }\ncatch (e) {\n} finally {\n}",
				Output: []string{"try { \n bar(); \n }\ncatch (e) {\n}\n finally {\n}"},
				Options: optsStroustrupSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 5, Column: 1},
				},
			},
			{
				Code:   "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
				Output: []string{"try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}"},
				Options: optsStroustrupSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "if (a) { \nb();\n } else { \nc();\n }",
				Output: []string{"if (a) { \nb();\n }\n else { \nc();\n }"},
				Options: optsStroustrupSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 2},
				},
			},

			// ---- allman + allowSingleLine ----
			{
				Code:   "if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
				Output: []string{"if (foo)\n{\n poop();\n} \nelse if (bar) \n{\nbaz();\n}\n else if (thing) \n{\nboom();\n}\nelse \n{\nqux();\n}"},
				Options: optsAllmanSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 1},
					{MessageId: "sameLineOpen", Line: 4, Column: 15},
					{MessageId: "sameLineClose", Line: 6, Column: 1},
					{MessageId: "sameLineOpen", Line: 6, Column: 19},
					{MessageId: "sameLineOpen", Line: 9, Column: 6},
				},
			},

			// ---- Comment interferes with fix — no output expected ----
			{
				Code: "if (foo) // comment \n{\nbar();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- https://github.com/eslint/eslint/issues/7493 ----
			{
				Code:   "if (foo) {\n bar\n.baz }",
				Output: []string{"if (foo) {\n bar\n.baz \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 6},
				},
			},
			{
				Code:   "if (foo)\n{\n bar\n.baz }",
				Output: []string{"if (foo)\n{\n bar\n.baz \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 4, Column: 6},
				},
			},
			{
				Code:   "if (foo) { bar\n.baz }",
				Output: []string{"if (foo) {\n bar\n.baz \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 10},
					{MessageId: "singleLineClose", Line: 2, Column: 6},
				},
			},
			{
				Code:   "if (foo) { bar\n.baz }",
				Output: []string{"if (foo) \n{\n bar\n.baz \n}"},
				Options: optsAllmanSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 10},
					{MessageId: "blockSameLine", Line: 1, Column: 10},
					{MessageId: "singleLineClose", Line: 2, Column: 6},
				},
			},
			{
				Code:   "switch (x) {\n case 1: foo() }",
				Output: []string{"switch (x) {\n case 1: foo() \n}"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 16},
				},
			},

			// ---- class declarations / expressions ----
			{
				Code:   "class Foo\n{\n}",
				Output: []string{"class Foo {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "(class\n{\n})",
				Output: []string{"(class {\n})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "class Foo{\n}",
				Output: []string{"class Foo\n{\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 10},
				},
			},
			{
				Code:   "(class {\n})",
				Output: []string{"(class \n{\n})"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 8},
				},
			},
			{
				Code:   "class Foo {\nbar() {\n}}",
				Output: []string{"class Foo {\nbar() {\n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "(class Foo {\nbar() {\n}})",
				Output: []string{"(class Foo {\nbar() {\n}\n})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 2},
				},
			},
			{
				Code:   "class\nFoo{}",
				Output: []string{"class\nFoo\n{}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 2, Column: 4},
				},
			},

			// ---- https://github.com/eslint/eslint/issues/7621 ----
			{
				Code:   "if (foo)\n{\n    bar\n}\nelse {\n    baz\n}",
				Output: []string{"if (foo) {\n    bar\n} else {\n    baz\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
					{MessageId: "nextLineClose", Line: 4, Column: 1},
				},
			},

			// ---- class static blocks ----
			{
				Code:   "class C {\n    static\n    {\n        foo;\n    }\n}",
				Output: []string{"class C {\n    static {\n        foo;\n    }\n}"},
				Options: opts1tbs(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
				},
			},
			{
				Code:   "class C {\n    static {foo;\n    }\n}",
				Output: []string{"class C {\n    static {\nfoo;\n    }\n}"},
				Options: opts1tbs(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 12},
				},
			},
			{
				Code:   "class C {\n    static {\n        foo;}\n}",
				Output: []string{"class C {\n    static {\n        foo;\n}\n}"},
				Options: opts1tbs(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 13},
				},
			},
			{
				Code:   "class C {\n    static\n    {foo;}\n}",
				Output: []string{"class C {\n    static {\nfoo;\n}\n}"},
				Options: opts1tbs(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
					{MessageId: "blockSameLine", Line: 3, Column: 5},
					{MessageId: "singleLineClose", Line: 3, Column: 10},
				},
			},
			{
				Code:   "class C {\n    static\n    {}\n}",
				Output: []string{"class C {\n    static {}\n}"},
				Options: opts1tbs(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
				},
			},
			{
				Code:   "class C {\n    static\n    {\n        foo;\n    }\n}",
				Output: []string{"class C {\n    static {\n        foo;\n    }\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
				},
			},
			{
				Code:   "class C {\n    static {foo;\n    }\n}",
				Output: []string{"class C {\n    static {\nfoo;\n    }\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 12},
				},
			},
			{
				Code:   "class C {\n    static {\n        foo;}\n}",
				Output: []string{"class C {\n    static {\n        foo;\n}\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 13},
				},
			},
			{
				Code:   "class C {\n    static\n    {foo;}\n}",
				Output: []string{"class C {\n    static {\nfoo;\n}\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
					{MessageId: "blockSameLine", Line: 3, Column: 5},
					{MessageId: "singleLineClose", Line: 3, Column: 10},
				},
			},
			{
				Code:   "class C {\n    static\n    {}\n}",
				Output: []string{"class C {\n    static {}\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 5},
				},
			},
			{
				Code:   "class C\n{\n    static{\n        foo;\n    }\n}",
				Output: []string{"class C\n{\n    static\n{\n        foo;\n    }\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 3, Column: 11},
				},
			},
			{
				Code:   "class C\n{\n    static\n    {foo;\n    }\n}",
				Output: []string{"class C\n{\n    static\n    {\nfoo;\n    }\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 4, Column: 5},
				},
			},
			{
				Code:   "class C\n{\n    static\n    {\n        foo;}\n}",
				Output: []string{"class C\n{\n    static\n    {\n        foo;\n}\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 5, Column: 13},
				},
			},
			{
				Code:   "class C\n{\n    static{foo;}\n}",
				Output: []string{"class C\n{\n    static\n{\nfoo;\n}\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 3, Column: 11},
					{MessageId: "blockSameLine", Line: 3, Column: 11},
					{MessageId: "singleLineClose", Line: 3, Column: 16},
				},
			},
			{
				Code:   "class C\n{\n    static{}\n}",
				Output: []string{"class C\n{\n    static\n{}\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 3, Column: 11},
				},
			},

			// ---- TS `with` statement ----
			{
				Code:   "with (foo) \n { \n bar(); }",
				Output: []string{"with (foo) { \n bar(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 2},
					{MessageId: "singleLineClose", Line: 3, Column: 9},
				},
			},
			{
				Code:    "with (foo) { bar(); \n }",
				Output:  []string{"with (foo) {\n bar(); \n }"},
				Options: opts1tbsSingle(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 12},
				},
			},

			// ---- TS namespace/module bodies ----
			{
				Code:   "module \"Foo\"\n{\n}",
				Output: []string{"module \"Foo\" {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "module \"Foo\"\n{\n}",
				Output: []string{"module \"Foo\" {\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "module \"Foo\" { \n }",
				Output: []string{"module \"Foo\" \n{ \n }"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 14},
				},
			},
			{
				Code:   "namespace Foo\n{\n}",
				Output: []string{"namespace Foo {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "namespace Foo\n{\n}",
				Output: []string{"namespace Foo {\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},
			{
				Code:   "namespace Foo { \n }",
				Output: []string{"namespace Foo \n{ \n }"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 15},
				},
			},
		},
	)
}
