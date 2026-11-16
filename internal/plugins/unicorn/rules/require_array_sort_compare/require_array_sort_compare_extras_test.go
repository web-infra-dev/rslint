package require_array_sort_compare_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_sort_compare"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireArraySortCompareExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_array_sort_compare.RequireArraySortCompareRule,
		[]rule_tester.ValidTestCase{
			valid(`new Uint8Array().toSorted()`, "file.js"),
			valid(`array["toSorted"]()`, "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`((array)).sort()`, "sort", "file.js", true),
			invalid(`((array)).toSorted()`, "toSorted", "file.js", true),
			// Comments outside the replaced argument range are preserved.
			commentCase(`array.sort(/* before */ undefined /* after */)`, "sort", "file.js", 1, 7, `array.sort(/* before */ COMPARATOR /* after */)`),
			commentCase("array.sort(\n\t// keep\n)", "sort", "file.js", 1, 7, "array.sort(\n\t// keep\nCOMPARATOR)"),
			// The JSDoc cast comment sits outside the parenthesized argument.
			commentCase(`array.sort(/** @type {any} */ (undefined))`, "sort", "file.js", 1, 7, `array.sort(/** @type {any} */ COMPARATOR)`),
			commentCase(`array.sort((undefined /* keep */))`, "sort", "file.js", 1, 7, ""),
			{
				// Upstream replaceArgument replaces the parenthesized argument range.
				Code:     `array.sort(((undefined)))`,
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1, Column: 7, EndLine: 1, EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: numericSuggestionID, Output: "array.sort(" + numericCompare + ")"},
						{MessageId: stringSuggestionID, Output: "array.sort(" + stringCompare + ")"},
					},
				}},
			},
		},
	)
}
