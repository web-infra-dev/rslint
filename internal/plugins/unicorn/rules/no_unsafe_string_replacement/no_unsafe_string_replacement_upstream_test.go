// TestNoUnsafeStringReplacementUpstream migrates the full valid/invalid suite
// from upstream v73.0.0 test/no-unsafe-string-replacement.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in no_unsafe_string_replacement_extras_test.go.
package no_unsafe_string_replacement_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unsafe_string_replacement"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-unsafe-string-replacement"

func replacementMessage(method string) string {
	return "Do not use a non-literal replacement value with `String#" + method + "()`."
}

func TestNoUnsafeStringReplacementUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unsafe_string_replacement.NoUnsafeStringReplacementRule,
		[]rule_tester.ValidTestCase{
			// ---- Allowed replacement values ----
			jsValid(`template.replace("{url}", "https://example.com")`),
			jsValid("template.replace(\"{url}\", `https://example.com`)"),
			jsValid("template.replace(\"{url}\", String.raw`https://example.com`)"),
			jsValid(`template.replace("{url}", () => htmlEscape(url))`),
			jsValid(`template.replace("{url}", function () { return htmlEscape(url); })`),
			tsValid(`template.replace("{url}", "https://example.com" as string)`),
			tsValid(`template.replace("{url}", "https://example.com" satisfies string)`),
			tsValid(`template.replace("{url}", "https://example.com"!)`),
			tsValid(`template.replace("{url}", <string>"https://example.com")`),
			jsValid(`template.replaceAll("{url}", "https://example.com")`),
			jsValid("template.replaceAll(\"{url}\", `https://example.com`)"),
			jsValid("string.replaceAll(/(?<symbol>`|\\$(?={))/g, String.raw`\\$<symbol>`)"),
			jsValid(`template.replaceAll("{url}", () => htmlEscape(url))`),
			jsValid(`template.replaceAll("{url}", function () { return htmlEscape(url); })`),
			jsValid("template.replace(\"{url}\", \"$` onerror=alert(1) \")"),

			// ---- Calls outside the matched shape ----
			jsValid(`template.replace("{url}")`),
			jsValid(`template.replace("{url}", replacement, extraArgument)`),
			jsValid(`template.replace(...argumentsArray)`),
			jsValid(`template.replace("{url}", ...replacement)`),
			jsValid(`template[replace]("{url}", replacement)`),
			jsValid(`template["replace"]("{url}", replacement)`),
			jsValid(`template.notReplace("{url}", replacement)`),
			jsValid(`replace("{url}", replacement)`),

			// ---- Known non-string receivers and plain object replacements ----
			jsValid(`const router = useRouter(); router.replace(pathname, {locale});`),
			jsValid(`const router = useRouter(); const options = {locale}; router.replace(pathname, options);`),
			jsValid(`router.replace(pathname, {locale: nextLocale});`),
			tsValid(`router.replace(pathname, {locale: nextLocale} as RouterOptions);`),
			tsValid(`const options = {locale: nextLocale}; router.replace(pathname, options as RouterOptions);`),
			tsValid(`declare const router: {replace(href: string, options: object): void}; router.replace(pathname, {locale});`),
			tsValid(`function foo(object: {replaceAll(a: string, b: object): void}) { object.replaceAll("{url}", {}); }`),
			tsValid(`function foo(value: number) { value.replace("{url}", replacement); }`),
			tsValid(`declare const pathname: string;
declare const options: unknown;
declare function useRouter(): {replace(href: string, options: unknown): void};
useRouter().replace(pathname, options);`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dynamic replacements ----
			invalid(`template.replace("{url}", htmlEscape(url))`, `htmlEscape(url)`, "replace"),
			invalid(`template.replaceAll("{url}", htmlEscape(url))`, `htmlEscape(url)`, "replaceAll"),
			invalid(`template.replace("{url}", replacement)`, `replacement`, "replace"),
			invalid(`template.replace("{url}", options.replacement)`, `options.replacement`, "replace"),
			invalid(`template.replace("{url}", options?.replacement)`, `options?.replacement`, "replace"),
			invalid(`template.replace("{url}", String(url))`, `String(url)`, "replace"),
			invalid("template.replace(\"{url}\", String.raw`${url}`)", "String.raw`${url}`", "replace"),
			invalid("template.replace(\"{url}\", css`safe string`)", "css`safe string`", "replace"),
			invalid("const String = {raw: () => replacement}; template.replace(\"{url}\", String.raw`ignored`)", "String.raw`ignored`", "replace"),
			tsInvalid(`template.replace("{url}", htmlEscape(url) as string)`, `htmlEscape(url) as string`, "replace"),
			invalid("template.replace(\"{url}\", `${url}`)", "`${url}`", "replace"),
			invalid(`template.replace("{url}", url ? htmlEscape(url) : "")`, `url ? htmlEscape(url) : ""`, "replace"),

			// ---- Object coercion and other known dynamic values ----
			invalid(`template.replace("{url}", {toString() { return url; }})`, `{toString() { return url; }}`, "replace"),
			invalid(`template.replace("{url}", {toString: () => url})`, `{toString: () => url}`, "replace"),
			invalid(`template.replace("{url}", {valueOf: () => url})`, `{valueOf: () => url}`, "replace"),
			invalid(`template.replace("{url}", {__proto__: {toString() { return url; }}})`, `{__proto__: {toString() { return url; }}}`, "replace"),
			invalid(`template.replace("{url}", [url])`, `[url]`, "replace"),
			invalid(`template.replace("{url}", 1)`, `1`, "replace"),
			invalid(`template.replace("{url}", (htmlEscape(url), url))`, `htmlEscape(url), url`, "replace"),
			invalid(`template.replaceAll("{url}", String(++count))`, `String(++count)`, "replaceAll"),

			// ---- Optional calls and report locations ----
			invalid(`template?.replace("{url}", replacement)`, `replacement`, "replace"),
			invalid(`template.replace?.("{url}", replacement)`, `replacement`, "replace"),
			invalid("template.replace(\n\t\"{url}\",\n\t/* comment */ htmlEscape(url)\n)", `htmlEscape(url)`, "replace"),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func invalid(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCase {
	return invalidWithFileName(code, target, method, "file.js", occurrence...)
}

func tsInvalid(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCase {
	return invalidWithFileName(code, target, method, "file.ts", occurrence...)
}

func invalidWithFileName(code string, target string, method string, fileName string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, method, occurrence...),
		},
	}
}

func expectedError(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCaseError {
	nth := 1
	if len(occurrence) > 0 {
		nth = occurrence[0]
	}
	offset := nthIndex(code, target, nth)
	if offset < 0 {
		panic("target not found in no-unsafe-string-replacement test: " + target)
	}
	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   replacementMessage(method),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(value string, target string, nth int) int {
	searchStart := 0
	for index := range nth {
		offset := strings.Index(value[searchStart:], target)
		if offset < 0 {
			return -1
		}
		searchStart += offset
		if index == nth-1 {
			return searchStart
		}
		searchStart += len(target)
	}
	return -1
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index := range offset {
		if code[index] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
