package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Logical assignments short-circuit like || && and ??: the right side runs only for a falsy, truthy or nullish left side.
// Each row is an expression placed after `x:` in a magic comment that also
// names the chunk. Validity was recorded from the pinned upstream rule and
// checked against Node.js's vm.runInNewContext, which the rule uses.

var logicalAssignmentRows = []classSyntaxRow{
	{"Object ||= missing", true},
	{"Object ??= missing", true},
	{"Object ||= a?.p", true},
	{"console ??= fallback?.p", true},
	{"undefined ??= missing", false},
	{"undefined &&= missing", true},
	{"NaN &&= missing", true},
	{"NaN ||= missing", false},
	{"null ??= missing", false},
	{"missing ||= 1", false},
	{"missing &&= 1", false},
	{"missing ??= 1", false},
	{"({p: 1}).p ||= missing", true},
	{"({p: 1}).p &&= missing", false},
	{"({p: 0}).p ||= missing", false},
	{"({p: null}).p ??= missing", false},
	{"({p: 1}).p ??= missing", true},
	{"[1][0] ||= missing", true},
	{"[0][0] ||= missing", false},
	{"({a: x} = Object ||= {a: 1})", true},
	{"({a: x} = Object ??= missing)", true},
	{"[Object ||= missing]", true},
	{"Object ||= (Object &&= missing)", true},
	{"(Object ||= missing) + 1", true},
}

func TestDynamicImportChunknameLogicalAssignmentTables(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, rows := range [][]classSyntaxRow{logicalAssignmentRows} {
		for _, row := range rows {
			code := "import(\n  /* webpackChunkName: \"a\", x: " + row.expression + " */\n  'm',\n)"
			if row.valid {
				valid = append(valid, rule_tester.ValidTestCase{Code: code})
				continue
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: classInvalidSyntaxMessage, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}
