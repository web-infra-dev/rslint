package prefer_jest_mocked_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_jest_mocked"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferJestMockedRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_jest_mocked.PreferJestMockedRule,
		[]rule_tester.ValidTestCase{
			{Code: `foo();`},
			{Code: `jest.mocked(foo).mockReturnValue(1);`},
			{Code: `bar.mockReturnValue(1);`},
			{Code: `sinon.stub(foo).returns(1);`},
			{Code: `foo.mockImplementation(() => 1);`},
			{Code: `obj.foo();`},
			{Code: `mockFn.mockReturnValue(1);`},
			{Code: `arr[0]();`},
			{Code: `obj.foo.mockReturnValue(1);`},
			{Code: `jest.spyOn(obj, 'foo').mockReturnValue(1);`},
			{Code: `(foo as Mock.jest).mockReturnValue(1);`},
			{Code: `
      type MockType = jest.Mock;
      const mockFn = jest.fn();
      (mockFn as MockType).mockReturnValue(1);
    `},
			{Code: `((foo as jest.Mock) as unknown).mockReturnValue(1);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `(foo as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   `(foo as unknown as string as unknown as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code:   `((foo as unknown) as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:   `(foo as unknown as jest.Mock as unknown as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code:   `(<jest.Mock>foo).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `(((foo)) as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:   `(foo as jest.Mock).mockImplementation(1);`,
				Output: []string{`(jest.mocked(foo)).mockImplementation(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   `(foo as unknown as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:   `(<jest.Mock>foo as unknown).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo) as unknown).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `(Obj.foo as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(Obj.foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:   `([].foo as jest.Mock).mockReturnValue(1);`,
				Output: []string{`(jest.mocked([].foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:   `(foo as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:   `(foo as jest.MockedFunction).mockImplementation(1);`,
				Output: []string{`(jest.mocked(foo)).mockImplementation(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:   `(foo as unknown as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:   `(Obj.foo as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(Obj.foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:   `(new Array(0).fill(null).foo as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(new Array(0).fill(null).foo)).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code:   `(jest.fn(() => foo) as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(jest.fn(() => foo))).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:   `const mockedUseFocused = useFocused as jest.MockedFunction<typeof useFocused>;`,
				Output: []string{`const mockedUseFocused = jest.mocked(useFocused);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 26, EndLine: 1, EndColumn: 78},
				},
			},
			{
				Code:   `const filter = (MessageService.getMessage as jest.Mock).mock.calls[0][0];`,
				Output: []string{`const filter = (jest.mocked(MessageService.getMessage)).mock.calls[0][0];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 17, EndLine: 1, EndColumn: 55},
				},
			},
			{
				Code: `class A {}
(foo as jest.MockedClass<A>)
`,
				Output: []string{`class A {}
(jest.mocked(foo))
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 2, Column: 2, EndLine: 2, EndColumn: 28},
				},
			},
			{
				Code:   `(foo as jest.MockedObject<{method: () => void}>)`,
				Output: []string{`(jest.mocked(foo))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code:   `(Obj['foo'] as jest.MockedFunction).mockReturnValue(1);`,
				Output: []string{`(jest.mocked(Obj['foo'])).mockReturnValue(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 1, Column: 2, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `(
  new Array(100)
    .fill(undefined)
    .map(x => x.value)
    .filter(v => !!v).myProperty as jest.MockedFunction<{
    method: () => void;
  }>
).mockReturnValue(1);
`,
				Output: []string{`(
  jest.mocked(new Array(100)
    .fill(undefined)
    .map(x => x.value)
    .filter(v => !!v).myProperty)
).mockReturnValue(1);
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestMocked", Line: 2, Column: 3, EndLine: 7, EndColumn: 5},
				},
			},
		},
	)
}
