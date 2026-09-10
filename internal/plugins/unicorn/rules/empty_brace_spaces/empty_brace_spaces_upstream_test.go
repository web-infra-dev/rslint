// TestEmptyBraceSpacesUpstream migrates the valid/invalid suite from
// eslint-plugin-unicorn v73.0.0/test/empty-brace-spaces.js 1:1. The upstream
// suite builds its cases from a small set of templates, so the expansion is
// reproduced here with the same grouping and placeholder substitution.
// rslint-specific lock-in cases live in empty_brace_spaces_extras_test.go.
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

const (
	messageID   = "empty-brace-spaces"
	messageText = "Do not add spaces between braces."
	placeholder = "/* */"
)

// cases holds every upstream template that produces a BlockStatement,
// StaticBlock, or ObjectExpression brace pair.
var cases = []string{
	"{/* */}",
	"function foo(){/* */}",
	"if(foo) {/* */}",
	"if(foo) {} else if (bar) {/* */}",
	"if(foo) {} else {/* */}",
	"for(;;){/* */}",
	"for(foo in bar){/* */}",
	"for(foo of bar){/* */}",
	"switch (foo) {case bar: {/* */}}",
	"switch (foo) {default: {/* */}}",
	"try {/* */} catch(foo){}",
	"try {} catch(bar){/* */}",
	"try {} catch(foo){} finally {/* */}",
	"do {/* */} while (foo)",
	"while (foo){/* */}",
	"foo = () => {/* */}",
	"foo = function (){/* */}",
	"foo = {/* */}",
	"class Foo {bar() {/* */}}",
	"foo = class {bar() {/* */}}",
	"class Foo {static  {/* */}}",
}

// classBodyCases are the templates whose placeholder sits between a class's
// braces, where replacing it with a method keeps the class body non-empty.
var classBodyCases = []string{
	"class Foo {/* */}",
	"foo = class {/* */}",
}

var allCases = append(append([]string{}, cases...), classBodyCases...)

// ignoredCases are empty brace pairs upstream does not inspect: a switch
// CaseBlock, an object binding pattern, and an import clause.
var ignoredCases = []string{
	"switch (foo) {/* */}",
	"const {/* */} = foo",
	`import {/* */} from "foo"`,
}

func replacePlaceholder(code, replacement string) string {
	return strings.Replace(code, placeholder, replacement, 1)
}

// replaceAllPlaceholders mimics upstream's `String.prototype.replaceAll`, which
// feeds the same replacement to every placeholder occurrence in a template
// such as `try {/* */} catch(bar){/* */}`.
func replaceAllPlaceholders(code, replacement string) string {
	return strings.ReplaceAll(code, placeholder, replacement)
}

// position returns the 1-based line and UTF-16 column for byte offsets within
// code. `start` is the first byte of the reported span and `end` is one past
// its last byte.
func position(code string, start, end int) (int, int, int, int) {
	locate := func(offset int) (int, int) {
		prefix := code[:offset]
		lineStart := strings.LastIndex(prefix, "\n") + 1
		return strings.Count(prefix, "\n") + 1, len(utf16.Encode([]rune(prefix[lineStart:]))) + 1
	}
	line, column := locate(start)
	endLine, endColumn := locate(end)
	return line, column, endLine, endColumn
}

// invalidCase builds an invalid test case for a template whose placeholder is
// replaced with whitespace. Every placeholder site marks the inside of a brace
// pair, so each one reports the whitespace span between its braces.
func invalidCase(template, spaces string) rule_tester.InvalidTestCase {
	code := replaceAllPlaceholders(template, spaces)
	output := replaceAllPlaceholders(template, "")

	var errors []rule_tester.InvalidTestCaseError
	searchFrom := 0
	for {
		site := strings.Index(template[searchFrom:], placeholder)
		if site < 0 {
			break
		}
		site += searchFrom

		// Byte offsets of the reported span in the substituted code: from the
		// end of the opening brace through the start of the closing brace.
		start := site + strings.Count(template[:site], placeholder)*(len(spaces)-len(placeholder))
		end := start + len(spaces)

		// Upstream reports `loc: {start: openingBrace.end, end: closingBrace.start}`
		// and removes the same span, so a template placeholder that is directly
		// between the braces (no intervening whitespace) yields an empty range.
		line, column, endLine, endColumn := position(code, start, end)
		errors = append(errors, rule_tester.InvalidTestCaseError{
			MessageId: messageID,
			Message:   messageText,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		})
		searchFrom = site + len(placeholder)
	}
	if len(errors) == 0 {
		panic("invalid case requires a brace pair: " + template)
	}

	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors:   errors,
	}
}

func validCase(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

// TestEmptyBraceSpacesUpstream covers every case from
// eslint-plugin-unicorn v73.0.0/test/empty-brace-spaces.js.
func TestEmptyBraceSpacesUpstream(t *testing.T) {
	var valid []rule_tester.ValidTestCase

	// Empty bodies and comment-only bodies are already compact.
	for _, body := range []string{"", "/* comment */", "\n\t// comment \n"} {
		for _, code := range allCases {
			valid = append(valid, validCase(replacePlaceholder(code, body)))
		}
	}
	// Not empty.
	for _, code := range cases {
		valid = append(valid, validCase(replacePlaceholder(code, "unicorn")))
	}
	for _, code := range classBodyCases {
		valid = append(valid, validCase(replacePlaceholder(code, "baz() {}")))
	}
	// `with` uses a WithStatement rather than a plain BlockStatement upstream.
	valid = append(valid, rule_tester.ValidTestCase{
		Code:            "with (foo) {}",
		FileName:        "file.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "script"},
	})
	// We don't check these cases.
	for _, code := range ignoredCases {
		valid = append(valid, validCase(replacePlaceholder(code, strings.Repeat(" ", 3))))
	}

	var invalid []rule_tester.InvalidTestCase
	for _, spaces := range []string{" ", "\t", " \t \t ", "\n\n", "\r\n"} {
		for _, code := range allCases {
			invalid = append(invalid, invalidCase(code, spaces))
		}
	}
	invalid = append(invalid, invalidCase("with (foo) {/* */}", strings.Repeat(" ", 5)))

	// Upstream's standalone `test.snapshot` case, kept outside the generated
	// table. Its brace pair opens with a whitespace-only line, so the report
	// covers the whole run of spaces and the fix removes the line outright.
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:     "try {\n\tfoo();\n} catch (error) {\n" + strings.Repeat(" ", 7) + "\n}",
		FileName: "file.js",
		Output:   []string{"try {\n\tfoo();\n} catch (error) {}"},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   messageText,
			// `{` ends on line 3, and the closing `}` begins on line 5; the
			// reported run therefore crosses the indented blank line.
			Line:      3,
			Column:    18,
			EndLine:   5,
			EndColumn: 1,
		}},
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &empty_brace_spaces.EmptyBraceSpacesRule, valid, invalid)
}
