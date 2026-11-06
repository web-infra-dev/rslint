package prefer_regexp_test_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_regexp_test"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Source: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-regexp-test.js
func TestPreferRegexpTestUpstream(t *testing.T) {
	t.Run("upstream group 1", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "function foo(value: `a` | number) { if (/a/.exec(value)) {} }", FileName: "file.ts", Output: []string{"function foo(value: `a` | number) { if (/a/.test(value)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function foo(value: `a` | RegExp) { if (/a/.exec(value)) {} }", FileName: "file.ts", Output: []string{"function foo(value: `a` | RegExp) { if (/a/.test(value)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 2", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "function foo() { if (string[0].match(/a/)) {} } const string = `abc`;", FileName: "case.js", Output: []string{"function foo() { if (/a/.test(string[0])) {} } const string = `abc`;"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 3", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "if (object?.value.match(/./).length) {}", FileName: "case.js"},
			{Code: "if (object?.value.match(/./).length > 0) {}", FileName: "case.js"},
			{Code: "if (getObject()?.value.match(getPattern())?.length) {}", FileName: "case.js"},
			{Code: "if (object?.getValue().match(/./).length) {}", FileName: "case.js"},
			{Code: "if (object?.value.match(/./)?.length > 0) {}", FileName: "case.js"},
			{Code: "if (object?.value!.match(/./).length) {}", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "if ((object?.value).match(/./).length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((/./).test(object?.value)) {}"}}},
			}},
			{Code: "if (object?.regexp.exec(value).length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (object?.regexp.test(value)) {}"}}},
			}},
		})
	})
	t.Run("upstream group 4", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "text.search({regexp: /x/}.regexp) !== -1;", FileName: "case.js", Output: []string{"({regexp: /x/}.regexp).test(text);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "text.search(({regexp: /x/}.regexp)) !== -1;", FileName: "case.js", Output: []string{"({regexp: /x/}.regexp).test((text));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "text.search({regexp: /x/}.regexp!) !== -1;", FileName: "file.ts", Output: []string{"({regexp: /x/}.regexp!).test(text);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 5", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }", FileName: "file.ts"},
			{Code: "/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}", FileName: "file.ts"},
			{Code: "/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }", FileName: "file.ts"},
			{Code: "/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }", FileName: "file.ts", Output: []string{"/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}", FileName: "file.ts", Output: []string{"/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 65, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }"}}},
			}},
			{Code: "/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 57, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 6", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }", FileName: "file.ts"},
			{Code: "/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}", FileName: "file.ts"},
			{Code: "/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }", FileName: "file.ts"},
			{Code: "/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }", FileName: "file.ts", Output: []string{"/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}", FileName: "file.ts", Output: []string{"/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 65, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }"}}},
			}},
			{Code: "/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 57, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 7", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }", FileName: "file.ts"},
			{Code: "/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}", FileName: "file.ts"},
			{Code: "/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }", FileName: "file.ts"},
			{Code: "/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }", FileName: "file.ts", Output: []string{"/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}", FileName: "file.ts", Output: []string{"/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 65, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }"}}},
			}},
			{Code: "/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 57, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 8", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "/** @param {boolean} pattern */ function run(pattern: string) { if (text.match(pattern)) {} }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/** @param {string} pattern */ function run(pattern: RegExp) { if (text.match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 68, EndLine: 1, EndColumn: 87, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @param {string} pattern */ function run(pattern: RegExp) { if (pattern.test(text)) {} }"}}},
			}},
		})
	})
	t.Run("upstream group 9", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "/** @param {boolean} pattern */ function run(pattern: string) { if (text.match(pattern)) {} }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/** @param {string} pattern */ function run(pattern: RegExp) { if (text.match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 68, EndLine: 1, EndColumn: 87, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @param {string} pattern */ function run(pattern: RegExp) { if (pattern.test(text)) {} }"}}},
			}},
		})
	})
	t.Run("upstream group 10", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const pattern = /a/; function run() { return\"a\".search(pattern) !== -1; }", FileName: "case.js", Output: []string{"const pattern = /a/; function run() { return pattern.test(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 45, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const pattern = /a/; function run() { return\"a\".search(pattern) === -1; }", FileName: "case.js", Output: []string{"const pattern = /a/; function run() { return !pattern.test(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 45, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo\n\"a\".search(/a/) === -1", FileName: "case.js", Output: []string{"foo\n;!/a/.test(\"a\")"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function run() { return\"a\".match(/a/) ? true : false; }", FileName: "case.js", Output: []string{"function run() { return /a/.test(\"a\") ? true : false; }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "for (const value of\"a\".match(/a/) ? [1] : [2]);", FileName: "case.js", Output: []string{"for (const value of /a/.test(\"a\") ? [1] : [2]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function run(pattern: RegExp) { return\"a\".search(pattern) !== -1; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 39, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function run(pattern: RegExp) { return pattern.test(\"a\"); }"}}},
			}},
		})
	})
	t.Run("upstream group 11", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "const bar = !re.test(foo)", FileName: "case.js"},
			{Code: "const matches = foo.match(re) || []", FileName: "case.js"},
			{Code: "const matches = foo.match(re)", FileName: "case.js"},
			{Code: "const matches = re.exec(foo)", FileName: "case.js"},
			{Code: "string.search(regex)", FileName: "case.js"},
			{Code: "const index = string.search(regex)", FileName: "case.js"},
			{Code: "if (string.search(regex)) {}", FileName: "case.js"},
			{Code: "while (foo = re.exec(bar)) {}", FileName: "case.js"},
			{Code: "while ((foo = re.exec(bar))) {}", FileName: "case.js"},
			{Code: "if (foo.notMatch(re)) {}", FileName: "case.js"},
			{Code: "if (re.notExec(foo)) {}", FileName: "case.js"},
			{Code: "if (foo.match) {}", FileName: "case.js"},
			{Code: "if (re.exec) {}", FileName: "case.js"},
			{Code: "if (foo[match](re)) {}", FileName: "case.js"},
			{Code: "if (re[exec](foo)) {}", FileName: "case.js"},
			{Code: "if (foo[\"match\"](re)) {}", FileName: "case.js"},
			{Code: "if (re[\"exec\"](foo)) {}", FileName: "case.js"},
			{Code: "if (match(re)) {}", FileName: "case.js"},
			{Code: "if (exec(foo)) {}", FileName: "case.js"},
			{Code: "if (foo.match()) {}", FileName: "case.js"},
			{Code: "if (re.exec()) {}", FileName: "case.js"},
			{Code: "if (foo.search() !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.match(re, another)) {}", FileName: "case.js"},
			{Code: "if (re.exec(foo, another)) {}", FileName: "case.js"},
			{Code: "if (foo.search(re, another) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.match(...[regexp])) {}", FileName: "case.js"},
			{Code: "if (re.exec(...[string])) {}", FileName: "case.js"},
			{Code: "if (foo.search(...[regexp]) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.match(1)) {}", FileName: "case.js"},
			{Code: "if (foo.match(\"1\")) {}", FileName: "case.js"},
			{Code: "if (foo.match(null)) {}", FileName: "case.js"},
			{Code: "if (foo.match(undefined)) {}", FileName: "case.js"},
			{Code: "if (foo.match(void 0)) {}", FileName: "case.js"},
			{Code: "const pattern = undefined; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = void 0; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "if (foo.match(1n)) {}", FileName: "case.js"},
			{Code: "if (foo.match(true)) {}", FileName: "case.js"},
			{Code: "if (foo.search(undefined) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(void 0) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(\"1\") !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(`1`) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(`${pattern}`) !== -1) {}", FileName: "case.js"},
			{Code: "const pattern = \"1\"; if (foo.search(pattern) !== -1) {}", FileName: "case.js"},
			{Code: "const pattern = `1`; if (foo.search(pattern) !== -1) {}", FileName: "case.js"},
			{Code: "const pattern = `${value}`;\nif (foo.search(pattern) !== -1) {}", FileName: "case.js"},
			{Code: "type Pattern = string; function foo(pattern: Pattern) { return text.search(pattern) !== -1; }", FileName: "file.ts"},
			{Code: "const pattern = 1; if (foo.search(pattern) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(1 + 2) !== -1) {}", FileName: "case.js"},
			{Code: "if (foo.search(/re/) <= -1) {}", FileName: "case.js"},
			{Code: "if (-1 !== foo.search(/re/)) {}", FileName: "case.js"},
			{Code: "if (-1 === foo.search(/re/)) {}", FileName: "case.js"},
			{Code: "if (0 <= foo.search(/re/)) {}", FileName: "case.js"},
			{Code: "if (0 > foo.search(/re/)) {}", FileName: "case.js"},
			{Code: "if (({}).match(action)) {}", FileName: "case.js"},
			{Code: "if ((() => {}).match(action)) {}", FileName: "case.js"},
			{Code: "if (/string/.match(/regexp/)) {}", FileName: "case.js"},
			{Code: "if (/string/.search(/regexp/) !== -1) {}", FileName: "case.js"},
			{Code: "const actionCreator = {}; if (actionCreator.match(action)) {}", FileName: "case.js"},
			{Code: "const actionCreator = () => {}; if (actionCreator.match(action)) {}", FileName: "case.js"},
			{Code: "const pattern = \"1\"; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = `1`; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = `${value}`;\nif (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = {}; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = () => {}; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const pattern = 1 + 2; if (foo.match(pattern)) {}", FileName: "case.js"},
			{Code: "const object = {pattern: /regexp/};\nObject.defineProperty(object, 'pattern', {get() { return 'regexp'; }});\nif (foo.match(object.pattern)) {}", FileName: "case.js"},
			{Code: "if (\"regexp\".exec(foo)) {}", FileName: "case.js"},
			{Code: "const regexp = \"regexp\"; if (regexp.exec(foo)) {}", FileName: "case.js"},
			{Code: "const regexp = {}; if (regexp.exec(foo)) {}", FileName: "case.js"},
			{Code: "const regexp = () => {}; if (regexp.exec(foo)) {}", FileName: "case.js"},
			{Code: "if (regexp.exec({})) {}", FileName: "case.js"},
			{Code: "if (regexp.exec(/string/)) {}", FileName: "case.js"},
			{Code: "if (regexp.exec(undefined)) {}", FileName: "case.js"},
			{Code: "if (regexp.exec(void 0)) {}", FileName: "case.js"},
			{Code: "const value = {}; if (regexp.exec(value)) {}", FileName: "case.js"},
			{Code: "const value = () => {}; if (regexp.exec(value)) {}", FileName: "case.js"},
			{Code: "if (uri.match(/unicorn/).length >= 1) {}", FileName: "case.js"},
			{Code: "if (uri.match(/unicorn/).length === 0) {}", FileName: "case.js"},
			{Code: "if (0 < uri.match(/unicorn/).length) {}", FileName: "case.js"},
			{Code: "const Boolean = value => value; const re = /a/; const bar = Boolean(foo.match(re))", FileName: "case.js"},
			{Code: "const Boolean = value => value; const isNotEmpty = Boolean(uri.match(/unicorn/).length)", FileName: "case.js"},
			{Code: "function unicorn(Boolean) { const re = /a/; if (Boolean(foo.match(re))) {} }", FileName: "case.js"},
			{Code: "if (uri.match?.(/unicorn/)?.length) {}", FileName: "case.js"},
			{Code: "if (!uri.match(/unicorn/).length) {}", FileName: "case.js"},
			{Code: "if (!uri.match(/unicorn/)?.length) {}", FileName: "case.js"},
			{Code: "if (!uri.match(/unicorn/).length > 0) {}", FileName: "case.js"},
			{Code: "if (!(uri.match(/unicorn/).length > 0)) {}", FileName: "case.js"},
			{Code: "if (!(foo || uri.match(/unicorn/).length)) {}", FileName: "case.js"},
			{Code: "if (!(foo || uri.match(/unicorn/).length > 0)) {}", FileName: "case.js"},
			{Code: "if (!Boolean(uri.match(/unicorn/).length)) {}", FileName: "case.js"},
			{Code: "if (!Boolean(uri.match(/unicorn/).length > 0)) {}", FileName: "case.js"},
			{Code: "if (!Boolean(foo || uri.match(/unicorn/).length)) {}", FileName: "case.js"},
			{Code: "if (foo.match(a ? (b ? \"x\" : null) : \"y\")) {}", FileName: "case.js"},
			{Code: "function foo(pattern: 1) { if (\"string\".match(pattern)) {} }", FileName: "file.ts"},
			{Code: "function foo(pattern: \"a\" | 1) { if (\"string\".match(pattern)) {} }", FileName: "file.ts"},
			{Code: "function foo(regexp: String, value: string) { if (regexp.exec(value)) {} }", FileName: "file.ts"},
			{Code: "function foo(regexp: (string & {brand: 1}) | number, value: string) { if (regexp.exec(value)) {} }", FileName: "file.ts"},
			{Code: "function foo(value: RegExp & {brand: 1}) { if (value.match(/a/)) {} }", FileName: "file.ts"},
			{Code: "function foo(pattern) { if (\"string\".match(<number>pattern)) {} }", FileName: "file.ts"},
			{Code: "function foo(pattern) { if (\"string\".match(pattern satisfies number)) {} }", FileName: "file.ts"},
			{Code: "function foo(regexp: String | \"a\", value: string) { if (regexp.exec(value)) {} }", FileName: "file.ts"},
			{Code: "if (foo.match(a ? 1 : null)) {}", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "const re = /a/; const bar = !foo.match(re)", FileName: "case.js", Output: []string{"const re = /a/; const bar = !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = Boolean(foo.match(re))", FileName: "case.js", Output: []string{"const re = /a/; const bar = Boolean(re.test(foo))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = !!Boolean(foo.match(re))", FileName: "case.js", Output: []string{"const re = /a/; const bar = !!Boolean(re.test(foo))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; if (foo.match(re)) {}", FileName: "case.js", Output: []string{"const re = /a/; if (re.test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.match(re) ? 1 : 2", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo) ? 1 : 2"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; while (foo.match(re)) foo = foo.slice(1);", FileName: "case.js", Output: []string{"const re = /a/; while (re.test(foo)) foo = foo.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; do {foo = foo.slice(1)} while (foo.match(re));", FileName: "case.js", Output: []string{"const re = /a/; do {foo = foo.slice(1)} while (re.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; for (; foo.match(re); ) foo = foo.slice(1);", FileName: "case.js", Output: []string{"const re = /a/; for (; re.test(foo); ) foo = foo.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (object.match(/regexp/)) {}", FileName: "case.js", Output: []string{"if (/regexp/.test(object)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const pattern = /regexp/; if (object.match(pattern)) {}", FileName: "case.js", Output: []string{"const pattern = /regexp/; if (pattern.test(object)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const pattern = /regexp/g; if (object.match(pattern)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const pattern = /regexp/g; if (pattern.test(object)) {}"}}},
			}},
			{Code: "if (uri.match(/unicorn/).length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "if (uri.match(/unicorn/).length > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "if (Boolean(uri.match(/unicorn/).length)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (Boolean(/unicorn/.test(uri))) {}"}}},
			}},
			{Code: "if (uri.match(/unicorn/)?.length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "if (uri.match(/unicorn/)?.length > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "if (uri.match(/unicorn/).length /* keep */ > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 2, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}"}}},
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 3, Column: 2, EndLine: 3, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (\n\turi.match(/unicorn/).length ||\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}"}}},
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 4, Column: 2, EndLine: 4, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/)?.length > 0\n) {}"}}},
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 5, Column: 2, EndLine: 5, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\t/unicorn/.test(uri)\n) {}"}}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) !== -1", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) != -1", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) > -1", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) >= 0", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) === -1", FileName: "case.js", Output: []string{"const re = /a/; const bar = !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) == -1", FileName: "case.js", Output: []string{"const re = /a/; const bar = !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = foo.search(re) < 0", FileName: "case.js", Output: []string{"const re = /a/; const bar = !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.search(/re/) !== -1) {}", FileName: "case.js", Output: []string{"if (/re/.test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((foo).search(/re/) !== -1) {}", FileName: "case.js", Output: []string{"if ((/re/).test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((foo.search(/re/)) === -1) {}", FileName: "case.js", Output: []string{"if (!(/re/.test(foo))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 6, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.search(bar.baz()) !== -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (bar.baz().test(foo)) {}"}}},
			}},
			{Code: "if (foo.search(new RegExp(\"re\", \"g\")) !== -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"re\", \"g\").test(foo)) {}"}}},
			}},
			{Code: "if (foo.search(unknown) !== -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (unknown.test(foo)) {}"}}},
			}},
			{Code: "if (foo.search(/re/) /* keep */ !== -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.search(/re/) /* keep */ === -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const regex = /weird/g;\nif (foo.search(regex) !== -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 5, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const regex = /weird/g;\nif (regex.test(foo));"}}},
			}},
			{Code: "const regex = /weird/g;\nif (foo.search(regex) === -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 5, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const regex = /weird/g;\nif (!regex.test(foo));"}}},
			}},
			{Code: "const re = /a/y;\nif (foo.search(re) !== -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 5, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const re = /a/y;\nif (re.test(foo));"}}},
			}},
			{Code: "const re = /a/; const bar = !re.exec(foo)", FileName: "case.js", Output: []string{"const re = /a/; const bar = !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = Boolean(re.exec(foo))", FileName: "case.js", Output: []string{"const re = /a/; const bar = Boolean(re.test(foo))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; if (re.exec(foo)) {}", FileName: "case.js", Output: []string{"const re = /a/; if (re.test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; const bar = re.exec(foo) ? 1 : 2", FileName: "case.js", Output: []string{"const re = /a/; const bar = re.test(foo) ? 1 : 2"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; while (re.exec(foo)) foo = foo.slice(1);", FileName: "case.js", Output: []string{"const re = /a/; while (re.test(foo)) foo = foo.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; do {foo = foo.slice(1)} while (re.exec(foo));", FileName: "case.js", Output: []string{"const re = /a/; do {foo = foo.slice(1)} while (re.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; for (; re.exec(foo); ) foo = foo.slice(1);", FileName: "case.js", Output: []string{"const re = /a/; for (; re.test(foo); ) foo = foo.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (/unicorn/.exec(uri).length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "if (/unicorn/.exec(uri)?.length > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(uri)) {}"}}},
			}},
			{Code: "const re = /a/; if ((0, foo).match(re)) {}", FileName: "case.js", Output: []string{"const re = /a/; if ((re).test((0, foo))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; if ((0, foo).match((re))) {}", FileName: "case.js", Output: []string{"const re = /a/; if ((re).test((0, foo))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; if ((foo).match(re)) {}", FileName: "case.js", Output: []string{"const re = /a/; if ((re).test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /a/; if ((foo).match((re))) {}", FileName: "case.js", Output: []string{"const re = /a/; if ((re).test((foo))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.match(/re/)) {}", FileName: "case.js", Output: []string{"if (/re/.test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const bar = {bar: /a/}; if (foo.match(bar.baz)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const bar = {bar: /a/}; if (bar.baz.test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(bar.baz())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (bar.baz().test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(new RegExp(\"re\", \"g\"))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"re\", \"g\").test(foo)) {}"}}},
			}},
			{Code: "const options = {toString() { sideEffect(); return 'a'; }};\nif (getString().match(new RegExp(options))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 5, EndLine: 2, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const options = {toString() { sideEffect(); return 'a'; }};\nif (new RegExp(options).test(getString())) {}"}}},
			}},
			{Code: "if (foo.match(new SomeRegExp())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new SomeRegExp().test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(new SomeRegExp)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((new SomeRegExp).test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(bar?.baz)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (bar?.baz.test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(bar?.baz())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (bar?.baz().test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(bar || baz)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((bar || baz).test(foo)) {}"}}},
			}},
			{Code: "async function a() {\n\tif (foo.match(await bar())) {}\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 6, EndLine: 2, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "async function a() {\n\tif ((await bar()).test(foo)) {}\n}"}}},
			}},
			{Code: "if ((foo).match(/re/)) {}", FileName: "case.js", Output: []string{"if ((/re/).test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((foo).match(new SomeRegExp)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((new SomeRegExp).test(foo)) {}"}}},
			}},
			{Code: "if ((foo).match(bar?.baz)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((bar?.baz).test(foo)) {}"}}},
			}},
			{Code: "if ((foo).match(bar?.baz())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((bar?.baz()).test(foo)) {}"}}},
			}},
			{Code: "const bar = false; const baz = /a/; if ((foo).match(bar || baz)) {}", FileName: "case.js", Output: []string{"const bar = false; const baz = /a/; if ((bar || baz).test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "async function a() {\n\tif ((foo).match(await bar())) {}\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 6, EndLine: 2, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "async function a() {\n\tif ((await bar()).test(foo)) {}\n}"}}},
			}},
			{Code: "const re = [/a/]; if (foo.match([re][0])) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const re = [/a/]; if ([re][0].test(foo)) {}"}}},
			}},
			{Code: "async function a() {\n\tif (\n\t\t/* 1 */ foo() /* 2 */\n\t\t\t./* 3 */ match /* 4 */ (\n\t\t\t\t/* 5 */ await /* 6 */ bar() /* 7 */\n\t\t\t\t,\n\t\t\t\t/* 8 */\n\t\t\t)\n\t) {}\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 3, Column: 11, EndLine: 8, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "async function a() {\n\tif (\n\t\t/* 1 */ (await /* 6 */ bar()) /* 2 */\n\t\t\t./* 3 */ test /* 4 */ (\n\t\t\t\t/* 5 */ foo() /* 7 */\n\t\t\t\t,\n\t\t\t\t/* 8 */\n\t\t\t)\n\t) {}\n}"}}},
			}},
			{Code: "const regex = new RegExp('[.!?]\\s*$');\nif (foo.match(regex)) {}", FileName: "case.js", Output: []string{"const regex = new RegExp('[.!?]\\s*$');\nif (regex.test(foo)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 5, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.match(unknown)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (unknown.test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(/a/g));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/a/g.test(foo));"}}},
			}},
			{Code: "if (foo.match(/a/y));", FileName: "case.js", Output: []string{"if (/a/y.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.match(/a/gy));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/a/gy.test(foo));"}}},
			}},
			{Code: "if (foo.match(/a/ig));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/a/ig.test(foo));"}}},
			}},
			{Code: "if (foo.match(new RegExp(\"a\", \"g\")));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"a\", \"g\").test(foo));"}}},
			}},
			{Code: "if (/a/g.exec(foo));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/a/g.test(foo));"}}},
			}},
			{Code: "if (/a/y.exec(foo));", FileName: "case.js", Output: []string{"if (/a/y.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (/a/gy.exec(foo));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/a/gy.test(foo));"}}},
			}},
			{Code: "if (/a/yi.exec(foo));", FileName: "case.js", Output: []string{"if (/a/yi.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (new RegExp(\"a\", \"g\").exec(foo));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"a\", \"g\").test(foo));"}}},
			}},
			{Code: "if (new RegExp(\"a\", \"y\").exec(foo));", FileName: "case.js", Output: []string{"if (new RegExp(\"a\", \"y\").test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const regex = /weird/g;\nif (foo.match(regex));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 5, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const regex = /weird/g;\nif (regex.test(foo));"}}},
			}},
			{Code: "const regex = /weird/g;\nif (regex.exec(foo));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 2, Column: 11, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const regex = /weird/g;\nif (regex.test(foo));"}}},
			}},
			{Code: "const regex = /weird/y;\nif (regex.exec(foo));", FileName: "case.js", Output: []string{"const regex = /weird/y;\nif (regex.test(foo));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 2, Column: 11, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const regex = /weird/gyi;\nif (regex.exec(foo));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 2, Column: 11, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const regex = /weird/gyi;\nif (regex.test(foo));"}}},
			}},
			{Code: "let re = new RegExp('foo', 'g');\nif(str.match(re));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 4, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "let re = new RegExp('foo', 'g');\nif(re.test(str));"}}},
			}},
			{Code: "!/a/u.exec(foo)", FileName: "case.js", Output: []string{"!/a/u.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "!/a/v.exec(foo)", FileName: "case.js", Output: []string{"!/a/v.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function foo() { if (string[0].match(/a/)) {} } const string = \"abc\";", FileName: "case.js", Output: []string{"function foo() { if (/a/.test(string[0])) {} } const string = \"abc\";"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (foo.match(a ? (b ? /x/ : null) : /y/)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((a ? (b ? /x/ : null) : /y/).test(foo)) {}"}}},
			}},
			{Code: "if (foo.match(a ? /x/ : /y/)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((a ? /x/ : /y/).test(foo)) {}"}}},
			}},
			{Code: "function foo(pattern: string | Foo) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: string | Foo) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern: Foo & RegExp) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: Foo & RegExp) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern: ({a: 1} & {b: 1}) | RegExp) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 57, EndLine: 1, EndColumn: 80, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: ({a: 1} & {b: 1}) | RegExp) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern: Foo.Bar) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: Foo.Bar) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern) { if (\"string\".match(pattern as RegExp)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern) { if ((pattern as RegExp).test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern) { if (\"string\".match(pattern as Foo)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern) { if ((pattern as Foo).test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern: RegExp | undefined) { if (\"string\".match(pattern!)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 49, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: RegExp | undefined) { if (pattern!.test(\"string\")) {} }"}}},
			}},
		})
	})
	t.Run("upstream group 12", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "function stringSearchPattern(pattern: string) {\n\tif (\"string\".search(pattern) !== -1) {}\n}\n\nfunction stringMatchPattern(pattern: string) {\n\tif (\"string\".match(pattern)) {}\n}\n\nfunction stringSearchAssertionPattern(pattern: string) {\n\tif (\"string\".search((`${pattern}` as string)) !== -1) {}\n}\n\nfunction stringMatchAssertionPattern(pattern: string) {\n\tif (\"string\".match((`${pattern}` as string))) {}\n}\n\nfunction actionCreatorMatcher(actionCreator: {match(action: unknown): boolean}, action: unknown) {\n\tif (actionCreator.match(action)) {}\n}\n\nfunction functionActionCreatorMatcher(actionCreator: (() => void) & {match(action: unknown): boolean}, action: unknown) {\n\tif (actionCreator.match(action)) {}\n}\n\nfunction stringExecReceiver(regexp: string, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction stringOrObjectExecReceiver(regexp: string | object, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectExecReceiver(regexp: {exec(value: string): unknown}, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectExecArgument(regexp: RegExp, value: object) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction regexpOrObjectExecArgument(regexp: RegExp, value: RegExp | object) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectMatchReceiver(value: object, pattern: RegExp) {\n\tif (value.match(pattern)) {}\n}\n\nfunction stringOrNumberMatchPattern(pattern: string | number) {\n\tif (\"string\".match(pattern)) {}\n}\n\nfunction stringOrRegExpMatchPattern(pattern: string | RegExp) {\n\tif (\"string\".match(pattern)) {}\n}\n\nfunction stringOrUndefinedSearchPattern(pattern: string | undefined) {\n\tif (\"string\".search(pattern) !== -1) {}\n}\n\nfunction arrayMatchPattern(pattern: string[]) {\n\tif (\"string\".match(pattern)) {}\n}\n\nfunction tupleSearchPattern(pattern: [string]) {\n\tif (\"string\".search(pattern) !== -1) {}\n}\n\nfunction genericArrayMatchPattern(pattern: Array<string>) {\n\tif (\"string\".match(pattern)) {}\n}\n\nfunction genericReadonlyArraySearchPattern(pattern: ReadonlyArray<string>) {\n\tif (\"string\".search(pattern) !== -1) {}\n}", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "function foo(pattern: unknown) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: unknown) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(pattern: Action) { if (\"string\".match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(pattern: Action) { if (pattern.test(\"string\")) {} }"}}},
			}},
			{Code: "function foo(regexp: RegExp, value: string) { if (regexp.exec(value)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 58, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(regexp: RegExp, value: string) { if (regexp.test(value)) {} }"}}},
			}},
		})
	})
	t.Run("upstream group 13", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "type UnknownAction = {\n\ttype: string;\n};\n\ntype ActionCreatorWithMatcher<Type extends string> = (() => {type: Type}) & {\n\tmatch(action: UnknownAction): action is {type: Type};\n};\n\ndeclare function createAction<Type extends string>(type: Type): ActionCreatorWithMatcher<Type>;\n\nconst increment = createAction('counter/increment');\ndeclare const action: UnknownAction;\n\nif (increment.match(action)) {}", FileName: "file.ts"},
			{Code: "declare function getPattern(): string | number;\nif (\"string\".match(getPattern())) {}", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "declare const pattern: any;\n\nif (\"string\".match(pattern)) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 3, Column: 5, EndLine: 3, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "declare const pattern: any;\n\nif (pattern.test(\"string\")) {}"}}},
			}},
			{Code: "declare const pattern: unknown;\n\nif (\"string\".match(pattern)) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 3, Column: 5, EndLine: 3, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "declare const pattern: unknown;\n\nif (pattern.test(\"string\")) {}"}}},
			}},
			{Code: "declare function getPattern(): RegExp | number; if (\"string\".match(getPattern())) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 53, EndLine: 1, EndColumn: 81, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "declare function getPattern(): RegExp | number; if (getPattern().test(\"string\")) {}"}}},
			}},
			{Code: "declare class Pattern extends RegExp {} declare function getPattern(): Pattern; if (\"string\".match(getPattern())) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 85, EndLine: 1, EndColumn: 113, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "declare class Pattern extends RegExp {} declare function getPattern(): Pattern; if (getPattern().test(\"string\")) {}"}}},
			}},
			{Code: "function foo<Type extends RegExp>(getPattern: () => Type | null) { if (\"string\".match(getPattern())) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 72, EndLine: 1, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo<Type extends RegExp>(getPattern: () => Type | null) { if (getPattern().test(\"string\")) {} }"}}},
			}},
			{Code: "function foo<Type>(getPattern: () => Type) { if (\"string\".match(getPattern())) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 50, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo<Type>(getPattern: () => Type) { if (getPattern().test(\"string\")) {} }"}}},
			}},
			{Code: "function foo<Type>(getPattern: () => NonNullable<Type>) { if (\"string\".match(getPattern())) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 63, EndLine: 1, EndColumn: 91, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo<Type>(getPattern: () => NonNullable<Type>) { if (getPattern().test(\"string\")) {} }"}}},
			}},
		})
	})
	t.Run("upstream group 14", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const re = /a/v; !re.exec(foo)", FileName: "case.js", Output: []string{"const re = /a/v; !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = new RegExp(\"a\", \"v\"); !re.exec(foo)", FileName: "case.js", Output: []string{"const re = new RegExp(\"a\", \"v\"); !re.test(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 15", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const re = /a/y; if (foo.search(re) !== -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 22, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const re = /a/y; if (re.test(foo));"}}},
			}},
			{Code: "const re = new RegExp(\"a\", \"y\"); if (foo.search(re) !== -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 38, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const re = new RegExp(\"a\", \"y\"); if (re.test(foo));"}}},
			}},
			{Code: "if (foo.search(new RegExp(\"a\", \"y\")) !== -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"a\", \"y\").test(foo));"}}},
			}},
		})
	})
	t.Run("upstream group 16", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "a\nfoo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"a\n;/b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo()\nfoo.search(/b/) > -1;", FileName: "case.js", Output: []string{"foo()\n;/b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (q) foo()\nfoo.match(/b/) ? 1 : 2;", FileName: "case.js", Output: []string{"if (q) foo()\n;/b/.test(foo) ? 1 : 2;"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const re = /b/;\nfoo.search(re) >= 0;", FileName: "case.js", Output: []string{"const re = /b/;\nre.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("upstream group 17", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "switch (foo.match(/b/)) { case null: break; }", FileName: "case.js"},
			{Code: "switch (foo.search(/b/)) { case -1: break; }", FileName: "case.js"},
			{Code: "switch (foo.match(/b/)) { case foo.match(/b/): break; }", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "switch (foo.match(/b/) ? 1 : 2) { case 1: break; }", FileName: "case.js", Output: []string{"switch (/b/.test(foo) ? 1 : 2) { case 1: break; }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Upstream runtime tests", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "const object = null; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];", FileName: "case.js"},
			{Code: "const object = undefined; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];", FileName: "case.js"},
			{Code: "const object = {value: \"x\"}; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "a;\nfoo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"a;\n/b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "  foo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"  /b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 3, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "a\r\nfoo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"a\r\n;/b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a = 1; foo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"const a = 1; /b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 14, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (q) foo();\nfoo.search(/b/) >= 0;", FileName: "case.js", Output: []string{"if (q) foo();\n/b/.test(foo);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Documentation examples", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "if (/unicorn/.test(string)) {}", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "if (string.match(/unicorn/)) {}", FileName: "case.js", Output: []string{"if (/unicorn/.test(string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (/unicorn/.exec(string)) {}", FileName: "case.js", Output: []string{"if (/unicorn/.test(string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (string.match(/unicorn/).length > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/unicorn/.test(string)) {}"}}},
			}},
			{Code: "if (string.search(/unicorn/) !== -1) {}", FileName: "case.js", Output: []string{"if (/unicorn/.test(string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	// Template parsers are not available in the native RuleTester. Keep every upstream case visible as an explained skip.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
		{Code: "<template><div v-if=\"/re/.exec(string)\"></div></template>", Skip: true},                                         // Vue parser is not supported.
		{Code: "<template><div v-if=\"'string'.match(/re/)\"></div></template>", Skip: true},                                      // Vue parser is not supported.
		{Code: "<template><div v-if=\"'string'.search(/re/) !== -1\"></div></template>", Skip: true},                              // Vue parser is not supported.
		{Code: "<script>if(/re/.exec(string));</script>", Skip: true},                                                             // Vue parser is not supported.
		{Code: "<template><button @click=\"setup()\nvalue.search(/x/) !== -1\"></button></template>", Skip: true},                 // Vue parser is not supported.
		{Code: "<template><button @click=\"setup()\r\nvalue.search(/x/) !== -1\"></button></template>", Skip: true},               // Vue parser is not supported.
		{Code: "<template><button @click=\"setup()\nvalue.match(/x/) ? found() : missing()\"></button></template>", Skip: true},   // Vue parser is not supported.
		{Code: "<template><button @click=\"if (q) run();\nvalue.search(/x/) !== -1\"></button></template>", Skip: true},           // Vue parser is not supported.
		{Code: "<template><div>{{ (() => { setup()\nvalue.search(/x/) !== -1 })() }}</div></template>", Skip: true},               // Vue parser is not supported.
		{Code: "<template><button @click=\"() => value.search(/x/) !== -1\"></button></template>", Skip: true},                    // Vue parser is not supported.
		{Code: "<template><button @click=\"setup(); if (/x/.exec(value)) run()\"></button></template>", Skip: true},               // Vue parser is not supported.
		{Code: "<script>const value = \"x\";</script>{value.search(/x/) !== -1}", Skip: true},                                     // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script><div>{value.match(/x/) ? yes : no}</div>", Skip: true},                       // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{#if value}{value.search(/x/) !== -1}{/if}", Skip: true},                     // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{ value.search(/x/) !== -1}", Skip: true},                                    // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{/* keep */ value.search(/x/) !== -1}", Skip: true},                          // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{@html value.search(/x/) !== -1}", Skip: true},                               // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script><div style:color={value.search(/x/) !== -1}></div>", Skip: true},             // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{(value.search(/x/) !== -1)}", Skip: true},                                   // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{value.search(/x/) === -1}", Skip: true},                                     // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script><div title={value.search(/x/) !== -1}></div>", Skip: true},                   // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{#if value.search(/x/) !== -1}yes{/if}", Skip: true},                         // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script><button onclick={() => value.search(/x/) !== -1}>test</button>", Skip: true}, // Svelte parser is not supported.
		{Code: "<script>const value = \"x\";</script>{value.search(/* keep */ /x/) !== -1}", Skip: true},                          // Svelte parser is not supported.
		{Code: "<template><div v-if=\"value.search(/x/) /* keep */ !== -1\"></div></template>", Skip: true},                       // Vue parser is not supported.
		{Code: "<template><div v-if=\"value.search(/x/) === /* keep */ -1\"></div></template>", Skip: true},                       // Vue parser is not supported.
		{Code: "<template><div v-if=\"value.match(/x/) /* keep */ .length\"></div></template>", Skip: true},                       // Vue parser is not supported.
		{Code: "<template><div v-if=\"value.match(/x/).length /* keep */ > 0\"></div></template>", Skip: true},                    // Vue parser is not supported.
		{Code: "<template><div v-if=\"/unicorn/.test(string)\">Vue</div></template>", Skip: true},                                 // Vue documentation parser is not supported.
		{Code: "<template><div v-if=\"/unicorn/.exec(string)\">Vue</div></template>", Skip: true},                                 // Vue documentation parser is not supported.
	}, nil)
}
