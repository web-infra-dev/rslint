package prefer_regexp_test_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_regexp_test"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases checked against eslint-plugin-unicorn v79.0.0.
func TestPreferRegexpTestExtras(t *testing.T) {
	t.Run("Additional AST and edit cases", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
			{Code: "if (text.match?.(/x/)) {}", FileName: "case.js"},
			{Code: "if (text?.match(/x/)) {}", FileName: "case.js"},
			{Code: "if ((text?.match)(/x/)) {}", FileName: "case.js"},
			{Code: "if (text[\"match\"](/x/)) {}", FileName: "case.js"},
			{Code: "class C { #match() {} f() { if (this.#match(/x/)) {} } }", FileName: "case.js"},
			{Code: "if (object?.value.match(/x/)) {}", FileName: "case.js"},
			{Code: "if (object?.regexp.exec(text)) {}", FileName: "case.js"},
			{Code: "if (text.match(/x/)[\"length\"]) {}", FileName: "case.js"},
			{Code: "const result = text.match(/x/).length > 0;", FileName: "case.js"},
			{Code: "if (Boolean(!text.match(/x/).length)) {}", FileName: "case.js"},
			{Code: "if (text.match(/x/).length ?? fallback) {}", FileName: "case.js"},
			{Code: "if (text.match(/x/) ?? fallback) {}", FileName: "case.js"},
			{Code: "if (text.search(/x/) < -0) {}", FileName: "case.js"},
			{Code: "const holder = {pattern:/x/}; if (text.match(holder.pattern)) {}", FileName: "case.js"},
			{Code: "if (text.match(Symbol.iterator)) {}", FileName: "case.js"},
			{Code: "const result = text.match(/x/) || fallback;", FileName: "case.js"},
			{Code: "switch (text.match(/x/)) {}", FileName: "case.js"},
			{Code: "if ((text.match(/x/)) as boolean) {}", FileName: "file.ts"},
			{Code: "if (text.match(/x/)!) {}", FileName: "file.ts"},
			{Code: "if ((text.match as Function)(/x/)) {}", FileName: "file.ts"},
			{Code: "function f(pattern: string | RegExp) { if (text.match(pattern)) {} }", FileName: "file.ts"},
			{Code: "function f(text: RegExp | number) { if (text.match(/x/)) {} }", FileName: "file.ts"},
			{Code: "const view = <div match={/x/} />;", FileName: "case.tsx"},
		}, []rule_tester.InvalidTestCase{
			{Code: "if ((text.match)(/x/)) {}", FileName: "case.js", Output: []string{"if ((/x/.test)(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((object?.value).match(/x/)) {}", FileName: "case.js", Output: []string{"if ((/x/).test(object?.value)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((object?.regexp).exec(text)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((object?.regexp).test(text)) {}"}}},
			}},
			{Code: "if ((text.match(/x/)).length > (0x0)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((/x/.test(text))) {}"}}},
			}},
			{Code: "if ((text.match(/x/).length) /* keep */ > 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(/x/) /* keep */ .length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(/x/).length > /* keep */ 0) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(/x/)?. /* keep */ length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((text.search(/x/)) >= (0x0)) {}", FileName: "case.js", Output: []string{"if ((/x/.test(text))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 6, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.search(/x/) > -(0x1)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.search(/x/) == (-1)) {}", FileName: "case.js", Output: []string{"if (!/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const result = text.search(/x/) === -1;", FileName: "case.js", Output: []string{"const result = !/x/.test(text);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((text.search(/x/)) /* keep */ !== -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 6, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(new RegExp())) {}", FileName: "case.js", Output: []string{"if (new RegExp().test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(new RegExp(/x/g, \"i\"))) {}", FileName: "case.js", Output: []string{"if (new RegExp(/x/g, \"i\").test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(new RegExp(\"[\", \"u\"))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"[\", \"u\").test(text)) {}"}}},
			}},
			{Code: "if (text.match(new RegExp(\"x\", \"gg\"))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"x\", \"gg\").test(text)) {}"}}},
			}},
			{Code: "const flags = \"i\"; let re = new RegExp(\"x\", flags); if (text.match(re)) {}", FileName: "case.js", Output: []string{"const flags = \"i\"; let re = new RegExp(\"x\", flags); if (re.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 57, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "let re = /x/; re = /x/g; if (text.match(re)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "let re = /x/; re = /x/g; if (re.test(text)) {}"}}},
			}},
			{Code: "const re = /x/y; if (re.exec(text)) {}", FileName: "case.js", Output: []string{"const re = /x/y; if (re.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const RegExp = Custom; if (text.match(new RegExp(\"x\"))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const RegExp = Custom; if (new RegExp(\"x\").test(text)) {}"}}},
			}},
			{Code: "const a = b; const b = a; if (text.match(a)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a = b; const b = a; if (a.test(text)) {}"}}},
			}},
			{Code: "if (text.match(({pattern:/x/}).pattern)) {}", FileName: "case.js", Output: []string{"if (({pattern:/x/}).pattern.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match((flag ? /x/ : null))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((flag ? /x/ : null).test((text))) {}"}}},
			}},
			{Code: "if (text.match(flag ? /x/ : unknown)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((flag ? /x/ : unknown).test(text)) {}"}}},
			}},
			{Code: "if (text.match(void effect())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((void effect()).test(text)) {}"}}},
			}},
			{Code: "if (text.match((effect(), /x/))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((effect(), /x/).test((text))) {}"}}},
			}},
			{Code: "if (text.search((effect(), /x/)) === -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (!(effect(), /x/).test((text))) {}"}}},
			}},
			{Code: "\"😀\"; if (text.match(/x/)) {}", FileName: "case.js", Output: []string{"\"😀\"; if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (\n /x/.exec(\n text\n )\n) {}", FileName: "case.js", Output: []string{"if (\n /x/.test(\n text\n )\n) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const result = () => text.search(({pattern:/x/}).pattern) !== -1;", FileName: "case.js", Output: []string{"const result = () => ({pattern:/x/}).pattern.test(text);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 22, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const result = () => text.search(({pattern:/x/}).pattern) === -1;", FileName: "case.js", Output: []string{"const result = () => !({pattern:/x/}).pattern.test(text);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 22, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (/** @type {string} */ (text).match(/x/)) {}", FileName: "case.js", Output: []string{"if (/** @type {string} */ (/x/).test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((/** @type {RegExp} */ (/x/)).exec(text)) {}", FileName: "case.js", Output: []string{"if ((/** @type {RegExp} */ (/x/)).test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match(/** @type {RegExp} */ (/x/))) {}", FileName: "case.js", Output: []string{"if (/x/.test(/** @type {RegExp} */ (text))) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "for (const x of\"x\".search(/x/) === -1 ? [] : []) {}", FileName: "case.js", Output: []string{"for (const x of !/x/.test(\"x\") ? [] : []) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 16, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function f(){return\"x\".search(({pattern:/x/}).pattern) === -1}", FileName: "case.js", Output: []string{"function f(){return !({pattern:/x/}).pattern.test(\"x\")}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 20, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if ((text as string).match(/x/)) {}", FileName: "file.ts", Output: []string{"if ((/x/).test(text as string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (text.match((pattern as RegExp))) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((pattern as RegExp).test((text))) {}"}}},
			}},
			{Code: "if (text.match((pattern satisfies RegExp))) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((pattern satisfies RegExp).test((text))) {}"}}},
			}},
			{Code: "if (text.match(<RegExp>pattern)) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((<RegExp>pattern).test(text)) {}"}}},
			}},
			{Code: "function f(pattern: RegExp | null) { if (text.match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 42, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: RegExp | null) { if (pattern.test(text)) {} }"}}},
			}},
			{Code: "function f(text: string | number) { if (/x/.exec(text)) {} }", FileName: "file.ts", Output: []string{"function f(text: string | number) { if (/x/.test(text)) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function f(pattern: Foo & RegExp) { if (text.match(pattern)) {} }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: Foo & RegExp) { if (pattern.test(text)) {} }"}}},
			}},
			{Code: "if (text.search(({pattern:/x/}).pattern!) !== -1) {}", FileName: "file.ts", Output: []string{"if (({pattern:/x/}).pattern!.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const view = <div>{text.match(/x/) ? \"yes\" : \"no\"}</div>;", FileName: "case.tsx", Output: []string{"const view = <div>{/x/.test(text) ? \"yes\" : \"no\"}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
}

func TestPreferRegexpTestArtifactsFollowDemand(t *testing.T) {
	for _, tc := range []struct {
		code, output    string
		fix, suggestion bool
	}{
		{`if (text.match(/x/)) {}`, `if (/x/.test(text)) {}`, true, false},
		{`const re = /x/y; if (re.exec(text)) {}`, `const re = /x/y; if (re.test(text)) {}`, true, false},
		{`text.search(/x/) === -1;`, `!/x/.test(text);`, true, false},
		{`if (text.match(/x/g)) {}`, `if (/x/g.test(text)) {}`, false, true},
		{`if (text.match(/x/)?.length > 0) {}`, `if (/x/.test(text)) {}`, false, true},
		{`if (re.exec(text)) {}`, `if (re.test(text)) {}`, false, true},
		{`text.search(/x/) /* keep */ !== -1;`, `text.search(/x/) /* keep */ !== -1;`, false, false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, source, err := helper.CreateTestProgram(tc.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_regexp_test.PreferRegexpTestRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_regexp_test.PreferRegexpTestRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity || diagnostic.RuleName != baseline.RuleName {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := tc.fix && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				wantSuggestion := tc.suggestion && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix || (diagnostic.Suggestions != nil) != wantSuggestion {
					t.Errorf("demand %d produced unexpected edit categories", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fixes", demand)
				}
				if wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Errorf("demand %d changed suggestions", demand)
				}
				if wantSuggestion {
					suggestions := *diagnostic.Suggestions
					if len(suggestions) != 1 || suggestions[0].Message.Id != "suggestion" || suggestions[0].Message.Description != "Switch to `RegExp#test(…)`." {
						t.Fatal("unexpected suggestion")
					}
					diagnostic.FixesPtr = &suggestions[0].FixesArr
				}
				// Applying edits sorts them in place; preserve the captured artifacts.
				if diagnostic.FixesPtr != nil {
					fixes := slices.Clone(*diagnostic.FixesPtr)
					diagnostic.FixesPtr = &fixes
				}
				output, _, _ := linter.ApplyRuleFixes(tc.code, []rule.RuleDiagnostic{diagnostic})
				want := tc.code
				if wantFix || wantSuggestion {
					want = tc.output
				}
				if output != want {
					t.Errorf("demand %d output = %q, want %q", demand, output, want)
				}
			}
		})
	}
}

func TestPreferRegexpTestStaticValueEdges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
		{Code: "const pattern = undefined; if (text.match(pattern)) {}", FileName: "case.js"},
		{Code: "if (object?.value.search(/x/) !== -1) {}", FileName: "case.js"},
		{Code: "if (text.match(Symbol.iterator)) {}", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "if (text.match({pattern: undefined}.pattern)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ({pattern: undefined}.pattern.test(text)) {}"}}},
		}},
		{Code: "let pattern = undefined; if (text.match(pattern)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "let pattern = undefined; if (pattern.test(text)) {}"}}},
		}},
		{Code: "if ((object?.value).search(/x/) !== -1) {}", FileName: "case.js", Output: []string{"if ((/x/).test(object?.value)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (object?.regexp.exec(text)?.length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (object?.regexp.test(text)) {}"}}},
		}},
		{Code: "if (text.match(Boolean)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (Boolean.test(text)) {}"}}},
		}},
		{Code: "if (text.search((async function() {}).pattern) === -1) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (!(async function() {}).pattern.test(text)) {}"}}},
		}},
		{Code: "const result = () => text.search((async function() {}).pattern) !== -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 22, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = () => (async function() {}).pattern.test(text);"}}},
		}},
	})
}

func TestPreferRegexpTestMutableMemberAliases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
		{Code: "const object = {pattern:/x/}; const pattern = object.pattern; if (text.match(pattern || /y/)) {}", FileName: "case.js"},
		{Code: "const object = {pattern:/x/}; const pattern = object.pattern; if ((pattern || /y/).exec(text)) {}", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "const pattern = parseInt(\"1\"); if (text.match(pattern || /y/)) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const pattern = parseInt(\"1\"); if ((pattern || /y/).test(text)) {}"}}},
		}},
	})
}

// Reviewed against eslint-plugin-unicorn v79.0.0: JS file kinds, numeric
// spellings, optional chains, type aliases, coercion, and disable directives.
// RuleTester registers the selected rule as "test" for disable directives.
func TestPreferRegexpTestReviewEdges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_regexp_test.PreferRegexpTestRule, []rule_tester.ValidTestCase{
		{Code: "if (text.search(/x/) >= (-0)) {}", FileName: "case.js"},
		{Code: "if (text.search(/x/) >= (0n)) {}", FileName: "case.js"},
		{Code: "if (text.search(/x/) !== -(1_0)) {}", FileName: "case.js"},
		{Code: "if (text.search(/x/) !== -(1n)) {}", FileName: "case.js"},
		{Code: "if (!Boolean(text.match(/x/).length > 0)) {}", FileName: "case.js"},
		{Code: "if (object?.value.match(/x/)) {}", FileName: "case.js"},
		{Code: "if (object?.value.match(/x/)?.length) {}", FileName: "case.js"},
		{Code: "if (!Boolean(object?.value.match(/x/).length > 0)) {}", FileName: "case.js"},
		{Code: "if (object?.regexp.exec(text)) {}", FileName: "case.js"},
		{Code: "if (!Boolean(object?.regexp.exec(text).length > 0)) {}", FileName: "case.js"},
		{Code: "if (!Boolean((object?.value).match(/x/).length > 0)) {}", FileName: "case.js"},
		{Code: "function f(pattern: never) {if (text.match(pattern)) {}}", FileName: "file.ts"},
		// Documented difference: the shared string proof recognizes string & unknown as string.
		{Code: "function f(pattern: string & unknown) {if (text.match(pattern)) {}}", FileName: "file.ts"},
		{Code: "function f(pattern: Array<string>) {if (text.match(pattern)) {}}", FileName: "file.ts"},
		{Code: "function f(pattern: ReadonlyArray<string>) {if (text.match(pattern)) {}}", FileName: "file.ts"},
		{Code: "function f<T extends string>(pattern: T) {if (text.match(pattern)) {}}", FileName: "file.ts"},
		{Code: "const object = {pattern:/x/}; Object.defineProperty(object, \"pattern\", {get(){return \"x\";}}); if(text.match(object.pattern)) {}", FileName: "case.js"},
		{Code: "/* eslint-disable test */ if (text.match(/x/)) {}", FileName: "case.js"},
		{Code: "if (text.match(/x/)) {} // eslint-disable-line test", FileName: "case.js"},
		{Code: "// eslint-disable-next-line test\nif (text.match(/x/)) {}", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: `if (text.match(new RegExp("x"))) {}`, FileName: "case.js", Globals: map[string]any{"RegExp": "off"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: `if (new RegExp("x").test(text)) {}`}}},
		}},
		{Code: `/* global RegExp:off */ if (text.match(new RegExp("x"))) {}`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: `/* global RegExp:off */ if (new RegExp("x").test(text)) {}`}}},
		}},
		{Code: "function read() {return 1;} if (text.match(read())) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function read() {return 1;} if (read().test(text)) {}"}}},
		}},
		{Code: "function read() {return 1;} if (read().match(/x/)) {}", FileName: "case.js", Output: []string{"function read() {return 1;} if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function read() {return 1;} if (text.match(read())) {}", FileName: "case.jsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function read() {return 1;} if (read().test(text)) {}"}}},
		}},
		{Code: "function read() {return 1;} if (read().match(/x/)) {}", FileName: "case.jsx", Output: []string{"function read() {return 1;} if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function read() {return 1;} if (text.match(read())) {}", FileName: "case.mjs", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function read() {return 1;} if (read().test(text)) {}"}}},
		}},
		{Code: "function read() {return 1;} if (read().match(/x/)) {}", FileName: "case.mjs", Output: []string{"function read() {return 1;} if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function read() {return 1;} if (text.match(read())) {}", FileName: "case.cjs", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function read() {return 1;} if (read().test(text)) {}"}}},
		}},
		{Code: "function read() {return 1;} if (read().match(/x/)) {}", FileName: "case.cjs", Output: []string{"function read() {return 1;} if (/x/.test(read())) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0x0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0b0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0o0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0.0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0e0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) >= (0x0_0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) !== -(0x1)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) !== -(0b1)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) !== -(0o1)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) !== -(1.0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.search(/x/) !== -(1e0)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-search", Message: "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(/x/)) {}", FileName: "case.js", Output: []string{"if (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(/x/)?.length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (/x/.test(text)) {}"}}},
		}},
		{Code: "if (object?.regexp.exec(text)?.length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "regexp-exec", Message: "Prefer `.test(…)` over `.exec(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (object?.regexp.test(text)) {}"}}},
		}},
		{Code: "if ((object?.value).match(/x/)) {}", FileName: "case.js", Output: []string{"if ((/x/).test(object?.value)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if ((object?.value).match(/x/)?.length) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if ((/x/).test(object?.value)) {}"}}},
		}},
		{Code: "function f(pattern: unknown) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: unknown) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "function f(pattern: any) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: any) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "function f(pattern: RegExp | undefined) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: RegExp | undefined) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "function f(pattern: RegExp & unknown) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(pattern: RegExp & unknown) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "type Pattern = RegExp; function f(pattern: Pattern) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 58, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "type Pattern = RegExp; function f(pattern: Pattern) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "type Pattern<T> = T; function f(pattern: Pattern<RegExp>) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 64, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "type Pattern<T> = T; function f(pattern: Pattern<RegExp>) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "class Pattern extends RegExp {} function f(pattern: Pattern) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 67, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class Pattern extends RegExp {} function f(pattern: Pattern) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "function f<T extends RegExp>(pattern: T) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f<T extends RegExp>(pattern: T) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "interface Pattern extends RegExp {} function f(pattern: Pattern) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 71, EndLine: 1, EndColumn: 90, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "interface Pattern extends RegExp {} function f(pattern: Pattern) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "type A = B; type B = A; function f(pattern: A) {if (text.match(pattern)) {}}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 53, EndLine: 1, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "type A = B; type B = A; function f(pattern: A) {if (pattern.test(text)) {}}"}}},
		}},
		{Code: "const a = b; const b = a; if (text.match(a)) {}", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a = b; const b = a; if (a.test(text)) {}"}}},
		}},
		{Code: "const 𝒜 = /x/; if (text.match(𝒜)) {}", FileName: "case.js", Output: []string{"const 𝒜 = /x/; if (𝒜.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// 😀\nif (text.match(/x/)) {}", FileName: "case.js", Output: []string{"// 😀\nif (/x/.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 2, Column: 5, EndLine: 2, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(new RegExp(\"x\", undefined))) {}", FileName: "case.js", Output: []string{"if (new RegExp(\"x\", undefined).test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(new RegExp(/x/g, undefined))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(/x/g, undefined).test(text)) {}"}}},
		}},
		{Code: "if (text.match(new RegExp(/x/g, \"\"))) {}", FileName: "case.js", Output: []string{"if (new RegExp(/x/g, \"\").test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(new RegExp(Symbol.iterator))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(Symbol.iterator).test(text)) {}"}}},
		}},
		{Code: "if (text.match(new RegExp(\"[a&&b]\", \"v\"))) {}", FileName: "case.js", Output: []string{"if (new RegExp(\"[a&&b]\", \"v\").test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (text.match(new RegExp(\"x\", \"uv\"))) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (new RegExp(\"x\", \"uv\").test(text)) {}"}}},
		}},
		{Code: "const pattern = /x/; use(pattern); if (text.match(pattern)) {}", FileName: "case.js", Output: []string{"const pattern = /x/; use(pattern); if (pattern.test(text)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "string-match", Message: "Prefer `RegExp#test(…)` over `String#match(…)`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
