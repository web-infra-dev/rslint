// TestMaxNestedDescribeUpstream ports the complete eslint-plugin-jest v29.16.0
// suite to Rstest API spellings. Rstest-specific syntax and regressions live in
// max_nested_describe_extras_test.go.
package max_nested_describe_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/max_nested_describe"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func maxOption(maxAllowed int) []any {
	return []any{map[string]any{"max": maxAllowed}}
}

func exceededDepthError(depth, maxAllowed, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "exceededMaxDepth",
		Message:   fmt.Sprintf("Too many nested describe calls (%d) - maximum allowed is %d", depth, maxAllowed),
		Line:      line,
		Column:    column,
	}
}

func runMaxNestedDescribeRuleTester(
	t *testing.T,
	valid []rule_tester.ValidTestCase,
	invalid []rule_tester.InvalidTestCase,
) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_nested_describe.MaxNestedDescribeRule,
		valid,
		invalid,
	)
}

func TestMaxNestedDescribeUpstream(t *testing.T) {
	runMaxNestedDescribeRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			{Code: `describe('one', function () {
  describe('two', function () {
    describe('three', function () {
      describe('four', function () {
        describe('five', function () {});
      });
    });
  });
});`},
			{Code: `describe('first', () => {
  describe('child', () => {});
});
describe('second', () => {
  describe('child', () => {});
});`},
			{
				Code: `describe('one', () => {
  describe.only('two', () => {
    describe.skip('three', () => {});
  });
});`,
				Options: maxOption(3),
			},
			{Code: `test('case', () => {});`, Options: maxOption(0)},
			// Rstest has no fdescribe/xdescribe aliases; the analogous upstream
			// valid case therefore remains non-Rstest code and is ignored.
			{Code: `fdescribe('one', () => { describe('two', () => {}); }); xdescribe('three', () => {});`, Options: maxOption(1)},
			{Code: `describe('one', () => {
  describe.each(['two'])('%s', () => {});
});`},
			{Code: "describe('one', () => {\n  describe.each`value\n  ${'two'}\n  `('$value', () => {});\n});"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `describe('one', () => {});`,
				Options: maxOption(0),
				Errors:  []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 1)},
			},
			{
				Code: `describe('one', () => {
  describe('two', () => {
    describe('three', () => {
      describe('four', () => {
        describe('five', () => {
          describe('six', () => {});
        });
      });
    });
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(6, 5, 6, 11)},
			},
			{
				Code: `describe('one', () => {
  describe('two', () => {
    describe('three', () => {
      describe('four', () => {
        describe('five', () => {
          describe('six-a', () => {});
          describe('six-b', () => {});
        });
      });
    });
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededDepthError(6, 5, 6, 11),
					exceededDepthError(6, 5, 7, 11),
				},
			},
			{
				Code: `describe('one', () => {
  describe.only('two', () => {
    describe.skip('three-a', () => {});
    describe('three-b', () => {});
  });
});
describe('separate', () => {});`,
				Options: maxOption(2),
				Errors: []rule_tester.InvalidTestCaseError{
					exceededDepthError(3, 2, 3, 5),
					exceededDepthError(3, 2, 4, 5),
				},
			},
			{
				Code: `describe('one', () => {
  describe.each(['two'])('%s', () => {});
});`,
				Options: maxOption(1),
				Errors:  []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 2, 3)},
			},
			{
				Code:    "describe('one', () => {\n  describe.each`value\n  ${'two'}\n  `('$value', () => {});\n});",
				Options: maxOption(1),
				Errors:  []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 2, 3)},
			},
		},
	)
}
