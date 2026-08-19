package no_object_constructor

// TestNoObjectConstructorUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-object-constructor.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases
// live in the no_object_constructor_extras_test.go file.

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// asiCase builds an InvalidTestCase for one of upstream's ASI-safety
// fixtures, whose replacement is always parenthesized, mirroring upstream's
// `props.code.replace(/(new )?Object\(\)/u, fixText)`.
func asiCase(code, pattern string, needsSemicolon bool, tsx bool) rule_tester.InvalidTestCase {
	if needsSemicolon {
		return suggestionCase(code, pattern, ";({})", "useLiteralAfterSemicolon", tsx)
	}
	return suggestionCase(code, pattern, "({})", "useLiteral", tsx)
}

// suggestionCase builds an InvalidTestCase for code containing exactly one
// occurrence of pattern (either "Object()" or "new Object()"), reported and
// replaced as a whole by fixText.
func suggestionCase(code, pattern, fixText, messageId string, tsx bool) rule_tester.InvalidTestCase {
	idx := strings.Index(code, pattern)
	if idx < 0 {
		panic("suggestionCase: pattern not found in code: " + pattern)
	}
	before := code[:idx]
	line := 1 + strings.Count(before, "\n")
	lastNL := strings.LastIndex(before, "\n")
	col := idx - lastNL
	endCol := col + len(pattern)
	output := before + fixText + code[idx+len(pattern):]

	return rule_tester.InvalidTestCase{
		Code: code,
		Tsx:  tsx,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferLiteral",
			Line:      line,
			Column:    col,
			EndLine:   line,
			EndColumn: endCol,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: messageId,
				Output:    output,
			}},
		}},
	}
}

func TestNoObjectConstructorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjectConstructorRule,
		[]rule_tester.ValidTestCase{
			{Code: `new Object(x)`},
			{Code: `Object(x)`},
			{Code: `new globalThis.Object`},
			{Code: `const createObject = Object => new Object()`},
			{Code: `var Object; new Object;`},
			{Code: `new Object()`, Globals: map[string]any{"Object": "off"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Explicit cases ----
			{
				Code: `new Object`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `({})`,
					}},
				}},
			},
			{
				Code: `Object()`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `({})`,
					}},
				}},
			},
			{
				Code: `const fn = () => Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `const fn = () => ({});`,
					}},
				}},
			},
			{
				Code: `Object() instanceof Object;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `({}) instanceof Object;`,
					}},
				}},
			},
			{
				Code: `const obj = Object?.();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `const obj = {};`,
					}},
				}},
			},
			{
				Code: `(new Object() instanceof Object);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Line:      1,
					Column:    2,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `({} instanceof Object);`,
					}},
				}},
			},

			// ---- Semicolon required before `({})` to compensate for ASI ----
			asiCase("\nfoo\nObject()\n", "Object()", true, false),
			asiCase("\nfoo()\nObject()\n", "Object()", true, false),
			asiCase("\nnew foo\nObject()\n", "Object()", true, false),
			asiCase("\n(a++)\nObject()\n", "Object()", true, false),
			asiCase("\n++a\nObject()\n", "Object()", true, false),
			asiCase("\nconst foo = function() {}\nObject()\n", "Object()", true, false),
			asiCase("\nconst foo = class {}\nObject()\n", "Object()", true, false),
			asiCase("\nfoo = this.return\nObject()\n", "Object()", true, false),
			asiCase("\nvar yield = bar.yield\nObject()\n", "Object()", true, false),
			asiCase("\nvar foo = { bar: baz }\nObject()\n", "Object()", true, false),
			asiCase("\n<foo />\nObject()\n", "Object()", true, true),
			asiCase("\n<foo></foo>\nObject()\n", "Object()", true, true),

			// ---- No semicolon required before `({})` because ASI does not occur ----
			asiCase("\n{}\nObject()\n", "Object()", false, false),
			asiCase("\nfunction foo() {}\nObject()\n", "Object()", false, false),
			asiCase("\nclass Foo {}\nObject()\n", "Object()", false, false),
			asiCase("foo: Object();", "Object()", false, false),
			asiCase("foo();Object();", "Object()", false, false),
			asiCase("{ Object(); }", "Object()", false, false),
			asiCase("if (a) Object();", "Object()", false, false),
			asiCase("if (a); else Object();", "Object()", false, false),
			asiCase("while (a) Object();", "Object()", false, false),
			asiCase("\ndo Object();\nwhile (a);\n", "Object()", false, false),
			asiCase("for (let i = 0; i < 10; i++) Object();", "Object()", false, false),
			asiCase("for (const prop in obj) Object();", "Object()", false, false),
			asiCase("for (const element of iterable) Object();", "Object()", false, false),
			asiCase("with (obj) Object();", "Object()", false, false),

			// ---- No semicolon required before `({})` because ASI still occurs ----
			asiCase("\nconst foo = () => {}\nObject()\n", "Object()", false, false),
			asiCase("\na++\nObject()\n", "Object()", false, false),
			asiCase("\na--\nObject()\n", "Object()", false, false),
			asiCase("\nfunction foo() {\n    return\n    Object();\n}\n", "Object()", false, false),
			asiCase("\nfunction * foo() {\n    yield\n    Object();\n}\n", "Object()", false, false),
			asiCase("\ndo {}\nwhile (a)\nObject()\n", "Object()", false, false),
			asiCase("\ndebugger\nObject()\n", "Object()", false, false),
			asiCase("\nfor (;;) {\n    break\n    Object()\n}\n", "Object()", false, false),
			asiCase("\nfor (;;) {\n    continue\n    Object()\n}\n", "Object()", false, false),
			asiCase("\nfoo: break foo\nObject()\n", "Object()", false, false),
			asiCase("\nfoo: while (true) continue foo\nObject()\n", "Object()", false, false),
			asiCase("\nconst foo = bar\nexport { foo }\nObject()\n", "Object()", false, false),
			asiCase("\nexport { foo } from 'bar'\nObject()\n", "Object()", false, false),
			asiCase("\nexport * as foo from 'bar'\nObject()\n", "Object()", false, false),
			asiCase("\nimport foo from 'bar'\nObject()\n", "Object()", false, false),
			asiCase("\nvar yield = 5;\n\nyield: while (foo) {\n    if (bar)\n        break yield\n    new Object();\n}\n", "new Object()", false, false),
			asiCase("\nvar foo\nObject()\n", "Object()", false, false),
			asiCase("\nlet bar\nObject()\n", "Object()", false, false),
		},
	)
}
