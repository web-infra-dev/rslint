package switch_exhaustiveness_check

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestSwitchExhaustivenessCheckRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		[]rule_tester.ValidTestCase{
			// All branches matched
			{Code: `
type Bool = true | false;

function test(value: Bool): number {
  switch (value) {
    case true:
      return 1;
    case false:
      return 0;
  }
}
`},
			// Non-union types don't require exhaustiveness
			{Code: `
const day = 'Monday' as string;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case 'Tuesday': {
    result = 2;
    break;
  }
}
`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases once exhaustiveness checking is implemented
			// The rule implementation needs type checking to detect non-exhaustive switches
		},
	)
}
