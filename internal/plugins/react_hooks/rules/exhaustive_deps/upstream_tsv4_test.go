package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_UpstreamTsv4 ports the small "upstream_tsv4_test" upstream group.

var upstreamTsv4Valid = []rule_tester.ValidTestCase{

}

var upstreamTsv4Invalid = []rule_tester.InvalidTestCase{
{
	Code: `
function Foo({ Component }) {
  React.useEffect(() => {
    console.log(<Component />);
  }, []);
};
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook React.useEffect has a missing dependency: 'Component'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Foo({ Component }) {
  React.useEffect(() => {
    console.log(<Component />);
  }, [Component]);
};
`}}},
	},
},
}

func TestExhaustiveDeps_UpstreamTsv4(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamTsv4Valid,
		upstreamTsv4Invalid,
	)
}
