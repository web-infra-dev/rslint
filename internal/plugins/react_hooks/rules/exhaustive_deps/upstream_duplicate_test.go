package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_Duplicate holds the upstream cases whose primary
// diagnostic falls into the "duplicate" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamDuplicateValid = []rule_tester.ValidTestCase{

}

var upstreamDuplicateInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
    console.log(local);
  }, [local, local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a duplicate dependency: 'local'. Either omit it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local, local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a duplicate dependency: 'local'. Either omit it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},
}

func TestExhaustiveDeps_Upstream_Duplicate(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamDuplicateValid,
		upstreamDuplicateInvalid,
	)
}
