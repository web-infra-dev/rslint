package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_StaleAssignment holds the upstream cases whose primary
// diagnostic falls into the "stale_assignment" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamStaleAssignmentValid = []rule_tester.ValidTestCase{

}

var upstreamStaleAssignmentInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent(props) {
  let value;
  let value2;
  let value3;
  let value4;
  let asyncValue;
  useEffect(() => {
    if (value4) {
      value = {};
    }
    value2 = 100;
    value = 43;
    value4 = true;
    console.log(value2);
    console.log(value3);
    setTimeout(() => {
      asyncValue = 100;
    });
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "Assignments to the 'value' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
		{Message: "Assignments to the 'value2' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
		{Message: "Assignments to the 'value4' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
		{Message: "Assignments to the 'asyncValue' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
	},
},

{
	Code: `
function MyComponent(props) {
  let value;
  let value2;
  let value3;
  let asyncValue;
  useEffect(() => {
    value = {};
    value2 = 100;
    value = 43;
    console.log(value2);
    console.log(value3);
    setTimeout(() => {
      asyncValue = 100;
    });
  }, [value, value2, value3]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "Assignments to the 'value' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
		{Message: "Assignments to the 'value2' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
		{Message: "Assignments to the 'asyncValue' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect."},
	},
},
}

func TestExhaustiveDeps_Upstream_StaleAssignment(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamStaleAssignmentValid,
		upstreamStaleAssignmentInvalid,
	)
}
