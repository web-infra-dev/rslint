// rslint-specific coverage for empty-brace-spaces. The cases here exercise the
// TypeScript-only AST shapes and the tsgo/ESTree structural differences the Go
// port has to bridge; the upstream suite lives in
// empty_brace_spaces_upstream_test.go.
package empty_brace_spaces_test

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/empty_brace_spaces"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// reportAt builds an invalid case reporting the whitespace between braces that
// spans `code[start:end]` (byte offsets, end exclusive). Upstream reports
// `loc: {start: openingBrace.end, end: closingBrace.start}` and removes the
// same span, so `output` drops that slice.
func reportAt(code, file string, start, end int) rule_tester.InvalidTestCase {
	locate := func(offset int) (int, int) {
		prefix := code[:offset]
		lineStart := strings.LastIndex(prefix, "\n") + 1
		return strings.Count(prefix, "\n") + 1, len(utf16.Encode([]rune(prefix[lineStart:]))) + 1
	}
	line, column := locate(start)
	endLine, endColumn := locate(end)

	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: file,
		Output:   []string{code[:start] + code[end:]},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   messageText,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		}},
	}
}

// emptyPairSpans lists the whitespace spans of every empty brace pair in code,
// in source order. Braces are matched literally, which is enough for the
// snippets below: none of them hides a brace inside a string or a comment.
func emptyPairSpans(code string) [][2]int {
	var spans [][2]int
	var stack []int
	for i := range len(code) {
		switch code[i] {
		case '{':
			stack = append(stack, i)
		case '}':
			if len(stack) == 0 {
				continue
			}
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if spaces := code[open+1 : i]; strings.Trim(spaces, " \t\n\r\v\f\u00a0\u3000\ufeff") == "" {
				spans = append(spans, [2]int{open + 1, i})
			}
		}
	}
	return spans
}

// invalidTS reports the only empty brace pair of a `.ts` case. It fails the
// test when the snippet does not have exactly one pair, so a case can never
// silently assert the wrong brace.
func invalidTS(t *testing.T, code string) rule_tester.InvalidTestCase {
	t.Helper()
	spans := emptyPairSpans(code)
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 empty brace pair in %q, found %d", code, len(spans))
	}
	return reportAt(code, "file.ts", spans[0][0], spans[0][1])
}

func validTS(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func TestEmptyBraceSpacesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Empty pairs without whitespace ----
		validTS(`class A {}`),
		validTS(`function f() {}`),
		validTS(`const o = {};`),
		validTS(`class A {static {}}`),

		// ---- Comment-only bodies keep the pair non-empty ----
		// Upstream's `/^\s+$/` test rejects a comment, so the whitespace
		// around it must survive untouched.
		validTS("class A { /* c */ }"),
		validTS("function f() {\n\t// c\n}"),
		validTS("const o = { /* c */ };"),
		validTS("class A {static { /* c */ }}"),
		validTS("try {} finally { /* c */ }"),

		// ---- Non-empty bodies ----
		validTS(`class A { m() {} }`),
		validTS(`class A { static m() {} }`),
		validTS("class A {\n\tm(): void {}\n}"),
		validTS(`function f() { return 1; }`),
		validTS(`const o = {a: 1};`),
		// A stray `;` is content between the braces, so it is not whitespace.
		validTS(`class A {;}`),
		validTS("class A { ; }"),
		validTS("const o = { ...s, };"),

		// ---- AST shapes the rule must not inspect ----
		// A switch CaseBlock, an object binding pattern, an import clause, a
		// type literal, and an enum body are not BlockStatement / ClassBody /
		// StaticBlock / ObjectExpression upstream either.
		validTS("switch (foo) { }"),
		validTS("const { } = foo;"),
		validTS(`import { } from "foo";`),
		validTS("export { };"),
		validTS("import { } from \"foo\";\nexport type T = { };\ninterface I { }\nenum E { }"),
		validTS("const x: { } = {};"),
		validTS("function f(): { } { return {}; }"),
		validTS("namespace N { }"),
		validTS("declare namespace N { }"),
		validTS("type T = { };"),
		validTS("declare module 'm' { }"),

		// ---- JSX: an attribute brace pair is not an object literal ----
		{Code: "const a = <div style={ } />;", FileName: "file.tsx", Tsx: true},

		// ---- Whitespace the rule does not treat as `\s` ----
		// U+200B ZERO WIDTH SPACE is not ECMAScript whitespace.
		validTS("class A {\u200b}"),

		// ---- `with` keeps the script source goal ----
		{
			Code:            "with (foo) {}",
			FileName:        "file.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- TypeScript-only class declarations ----
		invalidTS(t, `class A { }`),
		invalidTS(t, "abstract class A {\n}"),
		invalidTS(t, "declare class A { }"),
		invalidTS(t, `export default class { }`),
		invalidTS(t, `export class A { }`),
		invalidTS(t, `const A = class { }`),
		// Ambience, type parameters, and heritage sit between `class` and the
		// body, so the brace scan has to skip them.
		invalidTS(t, `class A<T> { }`),
		invalidTS(t, `class A extends B { }`),
		invalidTS(t, `class A implements I { }`),

		// ---- Decorated empty class bodies ----
		invalidTS(t, "@sealed\nclass A { }"),

		// ---- Static blocks ----
		// The class body itself is not empty, so only the block's pair qualifies.
		invalidTS(t, "class A {static { }}"),
		invalidTS(t, "class A { static {   } }"),
		invalidTS(t, "class A {static {\n}}"),

		// ---- Block statements ----
		invalidTS(t, `if (a) { }`),
		invalidTS(t, `while (a) { }`),
		invalidTS(t, `for (;;) { }`),
		invalidTS(t, `do { } while (a)`),
		invalidTS(t, `const f = () => { }`),
		// A labeled statement is still a BlockStatement.
		invalidTS(t, `label: { }`),
		// A destructuring parameter introduces braces before the body; the scan
		// must still land on the function body's brace.
		invalidTS(t, `const f = ({a}) => { }`),
		invalidTS(t, `const f = ([a]) => { }`),
		invalidTS(t, `function f({a}) { }`),
		invalidTS(t, `function f({a}: {a: number}) { }`),
		// A default value and the body are independent empty pairs, so both are
		// reported and both are fixed in one pass.
		{
			Code:     `function f(x = { }) { }`,
			FileName: "file.ts",
			Output:   []string{`function f(x = {}) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: messageID,
					Message:   messageText,
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 18,
				},
				{
					MessageId: messageID,
					Message:   messageText,
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 23,
				},
			},
		},
		// A nested empty block reports once per pair.
		{
			Code:     "function f() {\n\tif (a) { }\n}",
			FileName: "file.ts",
			Output:   []string{"function f() {\n\tif (a) {}\n}"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: messageID,
				Message:   messageText,
				Line:      2,
				Column:    10,
				EndLine:   2,
				EndColumn: 11,
			}},
		},

		// ---- Object literals ----
		invalidTS(t, `const o = { };`),
		invalidTS(t, "const o = {\n};"),
		invalidTS(t, `f({ });`),
		// The inner pair is the empty one; the outer literal holds a property.
		invalidTS(t, `const o = { a: { } };`),

		// ---- ECMAScript whitespace beyond ASCII ----
		// U+00A0 NBSP spans two UTF-8 bytes while the end column counts the
		// single decoded character.
		invalidTS(t, "class A {\u00a0}"),
		invalidTS(t, "class A {\u3000}"),
		invalidTS(t, "class A {\ufeff}"),
		invalidTS(t, "class A {\v}"),
		invalidTS(t, "class A {\f}"),
		invalidTS(t, "class A {\r\n}"),
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &empty_brace_spaces.EmptyBraceSpacesRule, valid, invalid)
}
