// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/require-array-sort-compare.js
package require_array_sort_compare_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_sort_compare"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID           = "require-array-sort-compare"
	numericSuggestionID = "require-array-sort-compare/numeric"
	stringSuggestionID  = "require-array-sort-compare/string"
	numericCompare      = "(a, b) => a - b"
	stringCompare       = "(a, b) => a.localeCompare(b)"
	message             = "Pass a compare function to avoid sorting elements as strings."
)

func valid(code, filename string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: filename}
}

func replaceCompareArgument(code, method, compare string) string {
	if strings.Contains(code, method+"(undefined)") {
		return strings.Replace(code, "undefined", compare, 1)
	}
	needle := method + "()"
	if !strings.Contains(code, needle) {
		panic("empty sort call not found in require-array-sort-compare fixture: " + method)
	}
	return strings.Replace(code, needle, method+"("+compare+")", 1)
}

func invalid(code, method, filename string, suggestions bool) rule_tester.InvalidTestCase {
	needle := method + "()"
	index := strings.Index(code, needle)
	if index < 0 {
		index = strings.Index(code, method+"(undefined)")
	}
	if index < 0 {
		index = strings.Index(code, method)
	}
	if index < 0 {
		panic("method not found in require-array-sort-compare fixture: " + method)
	}
	line := strings.Count(code[:index], "\n") + 1
	lineStart := strings.LastIndex(code[:index], "\n") + 1
	column := index - lineStart + 1
	expectedSuggestions := []rule_tester.InvalidTestCaseSuggestion{}
	if suggestions {
		expectedSuggestions = []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: numericSuggestionID, Output: replaceCompareArgument(code, method, numericCompare)},
			{MessageId: stringSuggestionID, Output: replaceCompareArgument(code, method, stringCompare)},
		}
	}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: filename,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message: message,
			Line: line, Column: column, EndLine: line, EndColumn: column + len(method),
			Suggestions: expectedSuggestions,
		}},
	}
}

func TestRequireArraySortCompareUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid(`function f(foo: Int8Array) { foo.sort(); }`, "file.ts"),
		valid(`const foo = new Int8Array(); foo.sort();`, "file.ts"),
		valid(`declare function getBytes(): Int8Array; getBytes().sort();`, "file.ts"),
		valid(`array.sort(compareFunction)`, "file.js"),
		valid(`array.toSorted(compareFunction)`, "file.js"),
		valid(`array.sort((a, b) => a - b)`, "file.js"),
		valid(`array.sort((a, b) => a.localeCompare(b))`, "file.js"),
		valid(`array.toSorted((a, b) => a - b)`, "file.js"),
		valid(`array.toSorted((a, b) => a.localeCompare(b))`, "file.js"),
		valid(`array.sort(...[])`, "file.js"),
		valid(`array.toSorted(...[])`, "file.js"),
		valid(`array.sort?.()`, "file.js"),
		valid(`array?.sort?.()`, "file.js"),
		valid(`array["sort"]()`, "file.js"),
		valid(`array[sort]()`, "file.js"),
		valid(`Array.prototype.sort.call(array)`, "file.js"),
		valid(`Array.prototype.sort.apply(array)`, "file.js"),
		valid(`Array.prototype.toSorted.call(array)`, "file.js"),
		valid(`Array.prototype.toSorted.apply(array)`, "file.js"),
		valid(`({sort() {}}).sort()`, "file.js"),
		valid(`({toSorted() {}}).toSorted()`, "file.js"),
		valid(`(() => {}).sort()`, "file.js"),
		valid(`(class {}).sort()`, "file.js"),
		valid(`new Set().sort()`, "file.js"),
		valid(`new Set().toSorted()`, "file.js"),
		valid(`const collection = new Set(); collection.sort()`, "file.js"),
		valid(`const object = {}; object.sort()`, "file.js"),
		valid(`const object = {}; object.toSorted()`, "file.js"),
		valid(`const function_ = () => {}; function_.sort()`, "file.js"),
		valid(`const array: string = ""; array.sort()`, "file.ts"),
		valid(`declare function getCollection(): {sort(): void}; getCollection().sort();`, "file.ts"),
	}
	invalidCases := []rule_tester.InvalidTestCase{
		invalid(`array.sort()`, "sort", "file.js", true),
		invalid(`array.toSorted()`, "toSorted", "file.js", true),
		invalid(`array.sort(undefined)`, "sort", "file.js", true),
		invalid(`array.toSorted(undefined)`, "toSorted", "file.js", true),
		invalid(`array?.sort()`, "sort", "file.js", true),
		invalid(`array?.toSorted()`, "toSorted", "file.js", true),
		invalid(`[].sort()`, "sort", "file.js", true),
		invalid(`[].toSorted()`, "toSorted", "file.js", true),
		invalid(`[3, 2, 1].sort()`, "sort", "file.js", true),
		invalid(`Array.from(iterable).sort()`, "sort", "file.js", true),
		invalid(`Array.of(3, 2, 1).sort()`, "sort", "file.js", true),
		invalid(`new Array(3).sort()`, "sort", "file.js", true),
		invalid(`const array = []; array.sort()`, "sort", "file.js", true),
		invalid(`const array = Array.from(iterable); array.sort()`, "sort", "file.js", true),
		invalid(`const array = Array.of(3, 2, 1); array.toSorted()`, "toSorted", "file.js", true),
		invalid(`array.sort(/* comment */)`, "sort", "file.js", false),
		invalid(`const array: string[] = []; array.sort()`, "sort", "file.ts", true),
		invalid(`const array: Array<string> = []; array.toSorted()`, "toSorted", "file.ts", true),
		invalid(`(value as string[]).sort()`, "sort", "file.ts", true),
		invalid(`(<string[]>value).toSorted()`, "toSorted", "file.ts", true),
	}
	if len(validCases) != 31 || len(invalidCases) != 20 {
		t.Fatalf("upstream coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &require_array_sort_compare.RequireArraySortCompareRule, validCases, invalidCases)
}

func TestRequireArraySortCompareDocumentation(t *testing.T) {
	first := "const numbers = [3, 1, 10, 2, 20];\nnumbers.toSorted();\nnumbers.toSorted((a, b) => a - b);"
	second := "[5, 10, 15, 2, 25].sort();\n[5, 10, 15, 2, 25].sort((a, b) => a - b);"
	third := "const names = ['Alice', 'bob', 'Charlie'];\nnames.sort();\nnames.sort((a, b) => a.localeCompare(b, undefined, {sensitivity: 'base'}));"
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_array_sort_compare.RequireArraySortCompareRule,
		[]rule_tester.ValidTestCase{valid(`const numbers = [3, 1, 4, 1, 5, 9]; numbers.sort((a, b) => b - a);`, "file.js")},
		[]rule_tester.InvalidTestCase{
			invalid(first, "toSorted", "file.js", true),
			invalid(second, "sort", "file.js", true),
			invalid(third, "sort", "file.js", true),
		},
	)
}
