// TestRequireMockTypeParametersUpstream migrates the @vitest/eslint-plugin
// v1.6.27 require-mock-type-parameters suite, with the namespace rewritten to
// Rstest's `rs`. Every case is carried over; the two Rstest module loaders
// that have no counterpart in that suite, `requireActual` and `requireMock`,
// are covered in require_mock_type_parameters_extras_test.go.
package require_mock_type_parameters

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireMockTypeParametersUpstream(t *testing.T) {
	checkImports := []any{map[string]any{"checkImportFunctions": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireMockTypeParametersRule,
		[]rule_tester.ValidTestCase{
			{Code: `rs.fn<(...args: any[]) => any>()`},
			{Code: `rs.fn<(...args: string[]) => any>()`},
			{Code: `rs.fn<(arg1: string) => string>()`},
			{Code: `rs.fn<(arg1: any) => string>()`},
			{Code: `rs.fn<(arg1: string) => void>()`},
			{Code: `rs.fn<(arg1: string, arg2: boolean) => string>()`},
			{Code: `rs.fn<(arg1: string, arg2: boolean, ...args: string[]) => string>()`},
			{Code: `rs.fn<MyProcedure>()`},
			{Code: `rs.fn<any>()`},
			{Code: `rs.fn<(...args: any[]) => any>(() => {})`},
			{Code: `rs.fn<() => string | undefined>().mockReturnValue("some error message");`},
			{Code: `rs.importActual<{ default: boolean }>("./example.js")`},
			{Code: `rs.importActual<MyModule>("./example.js")`},
			{Code: `rs.importActual<any>("./example.js")`},
			{Code: `rs.importMock<{ default: boolean }>("./example.js")`},
			{Code: `rs.importMock<MyModule>("./example.js")`},
			{Code: `rs.importMock<any>("./example.js")`},
			// The module loaders are only checked when the option asks for it.
			{Code: `rs.importActual("./example.js")`},
			{Code: `rs.importMock("./example.js")`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `rs.fn()`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingTypeParameter",
					Message:   "'fn' is called without a type parameter, so its result is untyped.",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 6,
				}},
			},
			{
				Code: `rs.fn(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingTypeParameter",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 6,
				}},
			},
			{
				Code:    `rs.importActual("./example.js")`,
				Options: checkImports,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingTypeParameter",
					Message:   "'importActual' is called without a type parameter, so its result is untyped.",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code:    `rs.importMock("./example.js")`,
				Options: checkImports,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingTypeParameter",
					Message:   "'importMock' is called without a type parameter, so its result is untyped.",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
		},
	)
}
