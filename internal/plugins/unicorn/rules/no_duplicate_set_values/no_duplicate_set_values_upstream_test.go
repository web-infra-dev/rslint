// Ported from eslint-plugin-unicorn v80.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-duplicate-set-values.js
package no_duplicate_set_values_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_duplicate_set_values"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-duplicate-set-values"

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func validTS(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

// duplicate describes one diagnostic on a single line. value is the reported
// element's source text, or "undefined" for an array hole.
func duplicate(line, column, endColumn int, value string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   "Remove duplicate value `" + value + "` from the Set.",
		Line:      line, Column: column, EndLine: line, EndColumn: endColumn,
	}
}

func invalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.js", Errors: errors}
}

func invalidTS(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.ts", Errors: errors}
}

func TestNoDuplicateSetValuesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_set_values.NoDuplicateSetValuesRule,
		[]rule_tester.ValidTestCase{
			valid(`new Set([1, 2, "1"])`),
			valid(`new Set([{}, {}])`),
			valid(`new Set([[], []])`),
			valid(`new Set([/foo/, /foo/])`),
			valid(`new Set([Symbol(), Symbol()])`),
			valid(`new Set(foo)`),
			valid(`Set([1, 1])`),
			valid(`new NotSet([1, 1])`),
			valid(`new globalThis.Set([1, 1])`),
			valid(`new Set([foo(), foo()])`),
			valid(`new Set([foo.bar, foo.baz])`),
			valid("const object = {value: 1};\nObject.defineProperty(object, 'value', {get() { return 2; }});\nnew Set([object.value, 1]);"),
			valid(`new Set([foo, ...bar, baz])`),
			valid("const foo = {};\nnew Set([foo, {}]);"),
			valid("const foo = 1;\nconst bar = 2;\nnew Set([foo, bar]);"),

			// Documentation examples.
			valid("const set = new Set([\n\t'foo',\n\t'bar',\n]);"),
			valid("const foo = 2;\nconst set = new Set([foo, 3]);"),
			valid(`const set = new Set([foo.bar, foo.baz]);`),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`new Set([1, 2, 1])`, duplicate(1, 16, 17, "1")),
			invalid(`new Set(["foo", "bar", "foo"])`, duplicate(1, 24, 29, `"foo"`)),
			invalid(`new Set([null, null])`, duplicate(1, 16, 20, "null")),
			invalid(`new Set([undefined, undefined])`, duplicate(1, 21, 30, "undefined")),
			invalid(`new Set([, undefined])`, duplicate(1, 12, 21, "undefined")),
			// A duplicate hole is reported on the array literal.
			invalid(`new Set([undefined, ,])`, duplicate(1, 9, 23, "undefined")),
			invalid(`new Set([,,])`, duplicate(1, 9, 13, "undefined")),
			invalid(`new Set([-1, -1])`, duplicate(1, 14, 16, "-1")),
			invalid(`new Set([NaN, NaN])`, duplicate(1, 15, 18, "NaN")),
			invalid(`new Set([0, -0])`, duplicate(1, 13, 15, "-0")),
			invalid(`new Set([1, 1, 2, 2])`, duplicate(1, 13, 14, "1"), duplicate(1, 19, 20, "2")),
			invalid(`new Set([1 + 1, 2])`, duplicate(1, 17, 18, "2")),
			invalid("new Set([`foo`, \"foo\"])", duplicate(1, 17, 22, `"foo"`)),
			invalid(`new Set([foo, foo])`, duplicate(1, 15, 18, "foo")),
			invalid(`new Set([foo.bar, foo.bar])`, duplicate(1, 19, 26, "foo.bar")),
			invalid(`new Set([foo["bar"], foo.bar])`, duplicate(1, 22, 29, "foo.bar")),
			invalid(`new Set([this.foo, this.foo])`, duplicate(1, 20, 28, "this.foo")),
			invalid(`new Set([foo, ...bar, foo])`, duplicate(1, 23, 26, "foo")),
			invalid("const foo = 2;\nnew Set([foo, 2]);", duplicate(2, 15, 16, "2")),
			invalid("const foo = 'bar';\nnew Set([foo, 'bar']);", duplicate(2, 15, 20, "'bar'")),
			invalid("const foo = undefined;\nnew Set([foo, undefined]);", duplicate(2, 15, 24, "undefined")),
			invalid("const foo = {};\nnew Set([foo, foo]);", duplicate(2, 15, 18, "foo")),

			// TypeScript wrappers around the array literal.
			invalidTS(`new Set([1, 2, 1] as const);`, duplicate(1, 16, 17, "1")),
			invalidTS(`new Set([1, 1]!);`, duplicate(1, 13, 14, "1")),
			invalidTS(`new Set([1, 2, 1] satisfies number[]);`, duplicate(1, 16, 17, "1")),

			// Documentation examples.
			invalid("const set = new Set([\n\t'foo',\n\t'bar',\n\t'foo',\n]);", duplicate(4, 2, 7, "'foo'")),
			invalid("const foo = 2;\nconst set = new Set([foo, 2]);", duplicate(2, 27, 28, "2")),
			invalid(`const set = new Set([foo.bar, foo.bar]);`, duplicate(1, 31, 38, "foo.bar")),
		},
	)
}
