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
			invalid(`array.sort(/* before */ undefined /* after */)`, "sort", "file.js", false),
		},
	)
}
