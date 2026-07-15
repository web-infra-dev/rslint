package no_mocks_import_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_mocks_import"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMocksImport(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_mocks_import.NoMocksImportRule,
		[]rule_tester.ValidTestCase{
			{Code: `import something from "something"`},
			{Code: `require("somethingElse")`},
			{Code: `require("./__mocks__.js")`},
			{Code: `require("./__mocks__x")`},
			{Code: `require("./__mocks__x/x")`},
			{Code: `require("./x__mocks__")`},
			{Code: `require("./x__mocks__/x")`},
			{Code: `require()`},
			{Code: `var path = "./__mocks__.js"; require(path)`},
			{Code: `entirelyDifferent(fn)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `require("./__mocks__")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noManualImport",
						Message:   "Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path",
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			{
				Code: `require("./__mocks__/")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `require("./__mocks__/index")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 9, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: `require("__mocks__")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 9, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `require("__mocks__/")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 9, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `require("__mocks__/index")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 9, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `import thing from "./__mocks__/index"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noManualImport", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
		},
	)
}
