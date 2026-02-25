package prefer_string_starts_ends_with

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStringStartsEndsWithRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferStringStartsEndsWithRule, []rule_tester.ValidTestCase{
		// String array - not string type
		{Code: `
function f(s: string[]) {
  s[0] === 'a';
}
`},
		// Optional chaining on string array | null
		{Code: `
function f(s: string[] | null) {
  s?.[0] === 'a';
}
`},
		// Optional chaining on string array | undefined
		{Code: `
function f(s: string[] | undefined) {
  s?.[0] === 'a';
}
`},
		// Not comparison operator
		{Code: `
function f(s: string) {
  s[0] + 'a';
}
`},
		// Index is not 0
		{Code: `
function f(s: string) {
  s[1] === 'a';
}
`},
		// Optional chaining with non-0 index
		{Code: `
function f(s: string | undefined) {
  s?.[1] === 'a';
}
`},
		// Union with non-string
		{Code: `
function f(s: string | string[]) {
  s[0] === 'a';
}
`},
		// Any type
		{Code: `
function f(s: any) {
  s[0] === 'a';
}
`},
		// Generic type
		{Code: `
function f<T>(s: T) {
  s[0] === 'a';
}
`},
		// String array - length - 1
		{Code: `
function f(s: string[]) {
  s[s.length - 1] === 'a';
}
`},
		// Optional chaining on string array | undefined - length - 1
		{Code: `
function f(s: string[] | undefined) {
  s?.[s.length - 1] === 'a';
}
`},
		// Different length offset
		{Code: `
function f(s: string) {
  s[s.length - 2] === 'a';
}
`},
		// Optional chaining - different length offset
		{Code: `
function f(s: string | undefined) {
  s?.[s.length - 2] === 'a';
}
`},
		// charAt on array
		{Code: `
function f(s: string[]) {
  s.charAt(0) === 'a';
}
`},
		// Optional chaining charAt on array
		{Code: `
function f(s: string[] | undefined) {
  s?.charAt(0) === 'a';
}
`},
		// charAt not comparison
		{Code: `
function f(s: string) {
  s.charAt(0) + 'a';
}
`},
		// charAt with non-0
		{Code: `
function f(s: string) {
  s.charAt(1) === 'a';
}
`},
		// Optional chaining charAt with non-0
		{Code: `
function f(s: string | undefined) {
  s?.charAt(1) === 'a';
}
`},
		// charAt without argument
		{Code: `
function f(s: string) {
  s.charAt() === 'a';
}
`},
		// charAt on array with length - 1
		{Code: `
function f(s: string[]) {
  s.charAt(s.length - 1) === 'a';
}
`},
		// charAt with different object in length
		{Code: `
function f(a: string, b: string, c: string) {
  (a + b).charAt((a + c).length - 1) === 'a';
}
`},
		// charAt with unrelated length property
		{Code: `
function f(a: string, b: string, c: string) {
  (a + b).charAt(c.length - 1) === 'a';
}
`},
		// indexOf on array
		{Code: `
function f(s: string[]) {
  s.indexOf(needle) === 0;
}
`},
		// indexOf on union
		{Code: `
function f(s: string | string[]) {
  s.indexOf(needle) === 0;
}
`},
		// indexOf with wrong comparison
		{Code: `
function f(s: string) {
  s.indexOf(needle) === s.length - needle.length;
}
`},
		// lastIndexOf on array
		{Code: `
function f(s: string[]) {
  s.lastIndexOf(needle) === s.length - needle.length;
}
`},
		// lastIndexOf with 0
		{Code: `
function f(s: string) {
  s.lastIndexOf(needle) === 0;
}
`},
		// match without null comparison
		{Code: `
function f(s: string) {
  s.match(/^foo/);
}
`},
		// match endsWith without null comparison
		{Code: `
function f(s: string) {
  s.match(/foo$/);
}
`},
		// match startsWith with addition
		{Code: `
function f(s: string) {
  s.match(/^foo/) + 1;
}
`},
		// match endsWith with addition
		{Code: `
function f(s: string) {
  s.match(/foo$/) + 1;
}
`},
		// match on non-string type (startsWith)
		{Code: `
function f(s: { match(x: any): boolean }) {
  s.match(/^foo/) !== null;
}
`},
		// match on non-string type (endsWith)
		{Code: `
function f(s: { match(x: any): boolean }) {
  s.match(/foo$/) !== null;
}
`},
		// match with non-start/end regex
		{Code: `
function f(s: string) {
  s.match(/foo/) !== null;
}
`},
		// match with both ^ and $
		{Code: `
function f(s: string) {
  s.match(/^foo$/) !== null;
}
`},
		// match with . in regex
		{Code: `
function f(s: string) {
  s.match(/^foo./) !== null;
}
`},
		// match with alternation
		{Code: `
function f(s: string) {
  s.match(/^foo|bar/) !== null;
}
`},
		// match with empty RegExp constructor
		{Code: `
function f(s: string) {
  s.match(new RegExp('')) !== null;
}
`},
		// match with unknown pattern variable
		{Code: `
function f(s: string) {
  s.match(pattern) !== null;
}
`},
		// match with RegExp syntax error
		{Code: `
function f(s: string) {
  s.match(new RegExp('^/!{[', 'u')) !== null;
}
`},
		// match without arguments
		{Code: `
function f(s: string) {
  s.match() !== null;
}
`},
		// match with non-regex argument
		{Code: `
function f(s: string) {
  s.match(777) !== null;
}
`},
		// slice on array
		{Code: `
function f(s: string[]) {
  s.slice(0, needle.length) === needle;
}
`},
		// slice endsWith on array
		{Code: `
function f(s: string[]) {
  s.slice(-needle.length) === needle;
}
`},
		// slice with wrong start index
		{Code: `
function f(s: string) {
  s.slice(1, 4) === 'bar';
}
`},
		// slice with negative start and end
		{Code: `
function f(s: string) {
  s.slice(-4, -1) === 'bar';
}
`},
		// slice with non-0 start
		{Code: `
function f(s: string) {
  s.slice(1) === 'bar';
}
`},
		// optional chaining slice with non-0 start
		{Code: `
function f(s: string | null) {
  s?.slice(1) === 'bar';
}
`},
		// pattern.test with unknown pattern
		{Code: `
function f(s: string) {
  pattern.test(s);
}
`},
		// regex test without arguments
		{Code: `
function f(s: string) {
  /^bar/.test();
}
`},
		// non-regex test call
		{Code: `
function f(x: { test(): void }, s: string) {
  x.test(s);
}
`},
		// slice with negative end
		{Code: `
function f(s: string) {
  s.slice(0, -4) === 'car';
}
`},
		// slice with negative end in compound expression
		{Code: `
function f(x: string, s: string) {
  x.endsWith('foo') && x.slice(0, -4) === 'bar';
}
`},
		// slice with non-matching length
		{Code: `
function f(s: string) {
  s.slice(0, length) === needle;
}
`},
		// slice with non-matching negative length
		{Code: `
function f(s: string) {
  s.slice(-length) === needle;
}
`},
		// slice with literal length not matching needle
		{Code: `
function f(s: string) {
  s.slice(0, 3) === needle;
}
`},
		// allowSingleElementEquality: 'always'
		{
			Code: `
declare const s: string;
s[0] === 'a';
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowSingleElementEquality": "always",
				},
			},
		},
		{
			Code: `
declare const s: string;
s[s.length - 1] === 'a';
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowSingleElementEquality": "always",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// === String indexing ===
		// s[0] === 'a'
		{
			Code: `
function f(s: string) {
  s[0] === 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('a');
}
`},
		},
		// s?.[0] === 'a'
		{
			Code: `
function f(s: string) {
  s?.[0] === 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s?.startsWith('a');
}
`},
		},
		// s[0] !== 'a'
		{
			Code: `
function f(s: string) {
  s[0] !== 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('a');
}
`},
		},
		// s?.[0] !== 'a'
		{
			Code: `
function f(s: string) {
  s?.[0] !== 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s?.startsWith('a');
}
`},
		},
		// s[0] == 'a'
		{
			Code: `
function f(s: string) {
  s[0] == 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('a');
}
`},
		},
		// s[0] != 'a'
		{
			Code: `
function f(s: string) {
  s[0] != 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('a');
}
`},
		},
		// s[0] === 'あ' (single character)
		{
			Code: "function f(s: string) {\n  s[0] === '\u3042';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{"function f(s: string) {\n  s.startsWith('\u3042');\n}\n"},
		},
		// s[0] === '👍' (emoji - length is 2, no fix)
		{
			Code: "function f(s: string) {\n  s[0] === '\U0001F44D';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
		},
		// s[0] === t (unknown length, no fix)
		{
			Code: `
function f(s: string, t: string) {
  s[0] === t;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
		},
		// s[s.length - 1] === 'a'
		{
			Code: `
function f(s: string) {
  s[s.length - 1] === 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('a');
}
`},
		},
		// noFormat: (s)[0] === ("a")
		{
			Code: "function f(s: string) {\n  (s)[0] === (\"a\")\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{"function f(s: string) {\n  (s).startsWith(\"a\")\n}\n"},
		},

		// === String#charAt ===
		// s.charAt(0) === 'a'
		{
			Code: `
function f(s: string) {
  s.charAt(0) === 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('a');
}
`},
		},
		// s.charAt(0) !== 'a'
		{
			Code: `
function f(s: string) {
  s.charAt(0) !== 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('a');
}
`},
		},
		// s.charAt(0) == 'a'
		{
			Code: `
function f(s: string) {
  s.charAt(0) == 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('a');
}
`},
		},
		// s.charAt(0) != 'a'
		{
			Code: `
function f(s: string) {
  s.charAt(0) != 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('a');
}
`},
		},
		// s.charAt(0) === 'あ'
		{
			Code: "function f(s: string) {\n  s.charAt(0) === '\u3042';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{"function f(s: string) {\n  s.startsWith('\u3042');\n}\n"},
		},
		// s.charAt(0) === '👍' (no fix)
		{
			Code: "function f(s: string) {\n  s.charAt(0) === '\U0001F44D';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
		},
		// s.charAt(0) === t (no fix)
		{
			Code: `
function f(s: string, t: string) {
  s.charAt(0) === t;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
		},
		// s.charAt(s.length - 1) === 'a'
		{
			Code: `
function f(s: string) {
  s.charAt(s.length - 1) === 'a';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('a');
}
`},
		},
		// noFormat: (s).charAt(0) === "a"
		{
			Code: "function f(s: string) {\n  (s).charAt(0) === \"a\";\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{"function f(s: string) {\n  (s).startsWith(\"a\");\n}\n"},
		},

		// === String#indexOf ===
		// s.indexOf(needle) === 0
		{
			Code: `
function f(s: string) {
  s.indexOf(needle) === 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith(needle);
}
`},
		},
		// s?.indexOf(needle) === 0
		{
			Code: `
function f(s: string) {
  s?.indexOf(needle) === 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s?.startsWith(needle);
}
`},
		},
		// s.indexOf(needle) !== 0
		{
			Code: `
function f(s: string) {
  s.indexOf(needle) !== 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith(needle);
}
`},
		},
		// s.indexOf(needle) == 0
		{
			Code: `
function f(s: string) {
  s.indexOf(needle) == 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith(needle);
}
`},
		},
		// s.indexOf(needle) != 0
		{
			Code: `
function f(s: string) {
  s.indexOf(needle) != 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith(needle);
}
`},
		},

		// === String#lastIndexOf ===
		// s.lastIndexOf('bar') === s.length - 3
		{
			Code: `
function f(s: string) {
  s.lastIndexOf('bar') === s.length - 3;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('bar');
}
`},
		},
		// s.lastIndexOf('bar') !== s.length - 3
		{
			Code: `
function f(s: string) {
  s.lastIndexOf('bar') !== s.length - 3;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.endsWith('bar');
}
`},
		},
		// s.lastIndexOf('bar') == s.length - 3
		{
			Code: `
function f(s: string) {
  s.lastIndexOf('bar') == s.length - 3;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('bar');
}
`},
		},
		// s.lastIndexOf('bar') != s.length - 3
		{
			Code: `
function f(s: string) {
  s.lastIndexOf('bar') != s.length - 3;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.endsWith('bar');
}
`},
		},
		// lastIndexOf with .length
		{
			Code: `
function f(s: string) {
  s.lastIndexOf('bar') === s.length - 'bar'.length;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('bar');
}
`},
		},
		// lastIndexOf with variable
		{
			Code: `
function f(s: string) {
  s.lastIndexOf(needle) === s.length - needle.length;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith(needle);
}
`},
		},

		// === String#match ===
		// s.match(/^bar/) !== null
		{
			Code: `
function f(s: string) {
  s.match(/^bar/) !== null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// s?.match(/^bar/) !== null
		{
			Code: `
function f(s: string) {
  s?.match(/^bar/) !== null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s?.startsWith("bar");
}
`},
		},
		// s.match(/^bar/) != null
		{
			Code: `
function f(s: string) {
  s.match(/^bar/) != null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// s.match(/bar$/) !== null
		{
			Code: `
function f(s: string) {
  s.match(/bar$/) !== null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith("bar");
}
`},
		},
		// s.match(/bar$/) != null
		{
			Code: `
function f(s: string) {
  s.match(/bar$/) != null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith("bar");
}
`},
		},
		// s.match(/^bar/) === null
		{
			Code: `
function f(s: string) {
  s.match(/^bar/) === null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith("bar");
}
`},
		},
		// s.match(/^bar/) == null
		{
			Code: `
function f(s: string) {
  s.match(/^bar/) == null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith("bar");
}
`},
		},
		// s.match(/bar$/) === null
		{
			Code: `
function f(s: string) {
  s.match(/bar$/) === null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.endsWith("bar");
}
`},
		},
		// s.match(/bar$/) == null
		{
			Code: `
function f(s: string) {
  s.match(/bar$/) == null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.endsWith("bar");
}
`},
		},
		// match with variable regex pattern
		{
			Code: `
const pattern = /^bar/;
function f(s: string) {
  s.match(pattern) != null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = /^bar/;
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// match with new RegExp variable
		{
			Code: `
const pattern = new RegExp('^bar');
function f(s: string) {
  s.match(pattern) != null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = new RegExp('^bar');
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// match with quoted regex
		{
			Code: `
const pattern = /^"quoted"/;
function f(s: string) {
  s.match(pattern) != null;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = /^"quoted"/;
function f(s: string) {
  s.startsWith("\"quoted\"");
}
`},
		},

		// === String#slice ===
		// s.slice(0, 3) === 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(0, 3) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('bar');
}
`},
		},
		// s?.slice(0, 3) === 'bar'
		{
			Code: `
function f(s: string) {
  s?.slice(0, 3) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s?.startsWith('bar');
}
`},
		},
		// s.slice(0, 3) !== 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(0, 3) !== 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('bar');
}
`},
		},
		// s.slice(0, 3) == 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(0, 3) == 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('bar');
}
`},
		},
		// s.slice(0, 3) != 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(0, 3) != 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.startsWith('bar');
}
`},
		},
		// s.slice(0, needle.length) === needle
		{
			Code: `
function f(s: string) {
  s.slice(0, needle.length) === needle;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith(needle);
}
`},
		},
		// s.slice(0, needle.length) == needle (no fix - loose equality with non-literal)
		{
			Code: `
function f(s: string) {
  s.slice(0, needle.length) == needle;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
		},
		// s.slice(-3) === 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(-3) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('bar');
}
`},
		},
		// s.slice(-3) !== 'bar'
		{
			Code: `
function f(s: string) {
  s.slice(-3) !== 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  !s.endsWith('bar');
}
`},
		},
		// s.slice(-needle.length) === needle
		{
			Code: `
function f(s: string) {
  s.slice(-needle.length) === needle;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith(needle);
}
`},
		},
		// s.slice(s.length - needle.length) === needle
		{
			Code: `
function f(s: string) {
  s.slice(s.length - needle.length) === needle;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith(needle);
}
`},
		},
		// s.substring(0, 3) === 'bar'
		{
			Code: `
function f(s: string) {
  s.substring(0, 3) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith('bar');
}
`},
		},
		// s.substring(-3) === 'bar' (no fix - probable mistake)
		{
			Code: `
function f(s: string) {
  s.substring(-3) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
		},
		// s.substring(s.length - 3, s.length) === 'bar'
		{
			Code: `
function f(s: string) {
  s.substring(s.length - 3, s.length) === 'bar';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith('bar');
}
`},
		},

		// === RegExp#test ===
		// /^bar/.test(s)
		{
			Code: `
function f(s: string) {
  /^bar/.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// /^bar/?.test(s)
		{
			Code: `
function f(s: string) {
  /^bar/?.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  s?.startsWith("bar");
}
`},
		},
		// /bar$/.test(s)
		{
			Code: `
function f(s: string) {
  /bar$/.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferEndsWith"},
			},
			Output: []string{`
function f(s: string) {
  s.endsWith("bar");
}
`},
		},
		// pattern.test(s) with variable regex
		{
			Code: `
const pattern = /^bar/;
function f(s: string) {
  pattern.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = /^bar/;
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// pattern.test(s) with new RegExp variable
		{
			Code: `
const pattern = new RegExp('^bar');
function f(s: string) {
  pattern.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = new RegExp('^bar');
function f(s: string) {
  s.startsWith("bar");
}
`},
		},
		// pattern.test(s) with quoted regex
		{
			Code: `
const pattern = /^"quoted"/;
function f(s: string) {
  pattern.test(s);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
const pattern = /^"quoted"/;
function f(s: string) {
  s.startsWith("\"quoted\"");
}
`},
		},
		// /^bar/.test(a + b) - binary expression arg
		{
			Code: `
function f(s: string) {
  /^bar/.test(a + b);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: string) {
  (a + b).startsWith("bar");
}
`},
		},

		// === Variation of string types ===
		// String literal union type
		{
			Code: `
function f(s: 'a' | 'b') {
  s.indexOf(needle) === 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f(s: 'a' | 'b') {
  s.startsWith(needle);
}
`},
		},
		// Generic type constraint
		{
			Code: `
function f<T extends 'a' | 'b'>(s: T) {
  s.indexOf(needle) === 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
function f<T extends 'a' | 'b'>(s: T) {
  s.startsWith(needle);
}
`},
		},
		// Intersection type
		{
			Code: `
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  s.indexOf(needle) === 0;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStartsWith"},
			},
			Output: []string{`
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  s.startsWith(needle);
}
`},
		},
	})
}
