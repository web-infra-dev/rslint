package no_duplicate_enum_values

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoDuplicateEnumValuesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateEnumValuesRule, []rule_tester.ValidTestCase{
		{
			Code: `
enum E {
  A,
  B,
}
			`,
		},
		{
			Code: `
enum E {
  A = 1,
  B,
}
			`,
		},
		{
			Code: `
enum E {
  A = 1,
  B = 2,
}
			`,
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'B',
}
			`,
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'B',
  C,
}
			`,
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'B',
  C = 2,
  D = 1 + 1,
}
			`,
		},
		{
			Code: `
enum E {
  A = 3,
  B = 2,
  C,
}
			`,
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'B',
  C = 2,
  D = foo(),
}
			`,
		},
		{
			Code: `
enum E {
  A = '',
  B = 0,
}
			`,
		},
		{
			Code: `
enum E {
  A = 0,
  B = -0,
  C = NaN,
}
			`,
		},
		{
			Code: `
const A = 'A';
enum E {
  A = 'A',
  B = ` + "`${A}`" + `,
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
enum E {
  A = 1,
  B = 1,
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicateValue",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'A',
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicateValue",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum E {
  A = 'A',
  B = 'A',
  C = 1,
  D = 1,
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicateValue",
					Line:      4,
					Column:    3,
				},
				{
					MessageId: "duplicateValue",
					Line:      6,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum E {
  A = 'A',
  B = ` + "`A`" + `,
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicateValue",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum E {
  A = ` + "`A`" + `,
  B = ` + "`A`" + `,
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicateValue",
					Line:      4,
					Column:    3,
				},
			},
		},
	})
}
