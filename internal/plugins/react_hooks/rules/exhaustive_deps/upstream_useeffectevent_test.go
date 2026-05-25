package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_UseEffectEvent holds the upstream cases whose primary
// diagnostic falls into the "useeffectevent" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamUseEffectEventValid = []rule_tester.ValidTestCase{

}

var upstreamUseEffectEventInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent({ theme }) {
  const onStuff = useEffectEvent(() => {
    showNotification(theme);
  });
  useEffect(() => {
    onStuff();
  }, [onStuff]);
  React.useEffect(() => {
    onStuff();
  }, [onStuff]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onStuff` from the list.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ theme }) {
  const onStuff = useEffectEvent(() => {
    showNotification(theme);
  });
  useEffect(() => {
    onStuff();
  }, []);
  React.useEffect(() => {
    onStuff();
  }, [onStuff]);
}
`}}},
		{Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onStuff` from the list.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ theme }) {
  const onStuff = useEffectEvent(() => {
    showNotification(theme);
  });
  useEffect(() => {
    onStuff();
  }, [onStuff]);
  React.useEffect(() => {
    onStuff();
  }, []);
}
`}}},
	},
},
}

func TestExhaustiveDeps_Upstream_UseEffectEvent(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamUseEffectEventValid,
		upstreamUseEffectEventInvalid,
	)
}
