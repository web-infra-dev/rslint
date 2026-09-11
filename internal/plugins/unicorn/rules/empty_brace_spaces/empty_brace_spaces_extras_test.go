// rslint-specific coverage for empty-brace-spaces. The cases here exercise the
// TypeScript-only AST shapes and the tsgo/ESTree structural differences the Go
// port has to bridge; the upstream suite lives in
// empty_brace_spaces_upstream_test.go.
package empty_brace_spaces_test

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/empty_brace_spaces"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// sourcePosition returns the 1-based line and UTF-16 column for byte offsets
// within code. `start` is the first byte of the reported span and `end` is one
// past its last byte.
func sourcePosition(code string, start, end int) (int, int, int, int) {
	locate := func(offset int) (int, int) {
		prefix := code[:offset]
		lineStart := strings.LastIndex(prefix, "\n") + 1
		return strings.Count(prefix, "\n") + 1, len(utf16.Encode([]rune(prefix[lineStart:]))) + 1
	}
	line, column := locate(start)
	endLine, endColumn := locate(end)
	return line, column, endLine, endColumn
}

// reportAt builds an invalid case reporting the whitespace between braces that
// spans `code[start:end]` (byte offsets, end exclusive). Upstream reports
// `loc: {start: openingBrace.end, end: closingBrace.start}` and removes the
// same span, so `output` drops that slice.
func reportAt(code, file string, start, end int) rule_tester.InvalidTestCase {
	line, column, endLine, endColumn := sourcePosition(code, start, end)

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

// emptyPairSpans reports the whitespace span of each empty brace pair, keyed by
// the position of its opening brace so that iteration in key order is source
// order. Braces are matched literally, which is enough for the snippets below:
// none of them hides a brace inside a string or a comment.
func emptyPairSpans(code string) map[int][2]int {
	spans := make(map[int][2]int)
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
			// Only a whitespace-only run is reported; an already-compact pair
			// such as `{}` produces no diagnostic.
			if spaces := code[open+1 : i]; spaces != "" && strings.Trim(spaces, " \t\n\r\v\f\u00a0\u3000\ufeff") == "" {
				spans[open] = [2]int{open + 1, i}
			}
		}
	}
	return spans
}

// invalidTS reports the whitespace run inside the last empty brace pair of a
// `.ts` case. Every case here is written so that the pair the rule reports is
// the final empty one in the snippet, which keeps the assertion unambiguous
// without restating byte offsets. It fails the test when that pair is missing,
// so a case can never silently assert the wrong brace.
func invalidTS(t *testing.T, code string) rule_tester.InvalidTestCase {
	t.Helper()
	spans := emptyPairSpans(code)
	openings := sortedOpenings(spans)
	if len(openings) == 0 {
		t.Fatalf("no empty brace pair in %q", code)
	}
	span := spans[openings[len(openings)-1]]
	return reportAt(code, "file.ts", span[0], span[1])
}

// invalidTSPairs reports every empty brace pair of a `.ts` case. Diagnostics
// are ordered the way the rule visits the nodes — a class body before the
// header expression nested in it — because the RuleTester matches the expected
// errors by index.
func invalidTSPairs(t *testing.T, code string) rule_tester.InvalidTestCase {
	t.Helper()
	spans := emptyPairSpans(code)
	openings := sortedOpenings(spans)
	if len(openings) == 0 {
		t.Fatalf("no empty brace pair in %q", code)
	}
	// The rule reports a class body before the header braces nested inside it.
	reversed := make([]int, 0, len(openings))
	for i := len(openings) - 1; i >= 0; i-- {
		reversed = append(reversed, openings[i])
	}

	testCase := reportAt(code, "file.ts", spans[openings[0]][0], spans[openings[0]][1])
	testCase.Errors = nil
	for _, opening := range reversed {
		span := spans[opening]
		line, column, endLine, endColumn := sourcePosition(code, span[0], span[1])
		testCase.Errors = append(testCase.Errors, rule_tester.InvalidTestCaseError{
			MessageId: messageID,
			Message:   messageText,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		})
	}
	testCase.Output = []string{removeSpans(code, spans, openings)}
	return testCase
}

// removeSpans deletes every listed span from code, right to left so the
// remaining offsets stay valid.
func removeSpans(code string, spans map[int][2]int, openings []int) string {
	out := code
	for i := len(openings) - 1; i >= 0; i-- {
		span := spans[openings[i]]
		out = out[:span[0]] + out[span[1]:]
	}
	return out
}

// invalidTSNested reports the innermost empty brace pair of a nested snippet
// such as `const o = { a: { } };`, where the outer pair is not empty.
func invalidTSNested(t *testing.T, code string) rule_tester.InvalidTestCase {
	t.Helper()
	spans := emptyPairSpans(code)
	openings := sortedOpenings(spans)
	if len(openings) == 0 {
		t.Fatalf("no empty brace pair in %q", code)
	}
	// The innermost pair is the one whose opening brace comes last.
	span := spans[openings[len(openings)-1]]
	return reportAt(code, "file.ts", span[0], span[1])
}

func sortedOpenings(spans map[int][2]int) []int {
	openings := make([]int, 0, len(spans))
	for opening := range spans {
		openings = append(openings, opening)
	}
	slices.Sort(openings)
	return openings
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
		// A brace in the class header must not be mistaken for the body's
		// opener: `mixin({})` and a `T extends {}` bound both contribute a
		// brace pair before the body starts.
		invalidTS(t, `class A extends mixin({}) { }`),
		invalidTS(t, `class A<T extends {}> { }`),
		invalidTS(t, `@decorator({}) class A { }`),
		// A header brace pair that holds a property is not empty, so only the
		// class body is reported.
		invalidTS(t, `class A extends mixin({ a: 1 }) { }`),
		// A header brace pair that is itself empty is reported too, and both
		// fixes are applied in a single pass.
		invalidTSPairs(t, `class A extends mixin({ }) { }`),

		// ---- Decorated empty class bodies ----
		invalidTS(t, "@sealed\nclass A { }"),

		// ---- Static blocks ----
		// The class body's own pair is already compact, so only the static
		// block's whitespace is reported.
		invalidTSPairs(t, "class A {static { }}"),
		invalidTSPairs(t, "class A { static {   } }"),
		invalidTSPairs(t, "class A {static {\n}}"),

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
		invalidTSNested(t, `const o = { a: { } };`),

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
