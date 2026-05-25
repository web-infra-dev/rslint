package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_UpstreamFlow ports the small "upstream_flow_test" upstream group.

var upstreamFlowValid = []rule_tester.ValidTestCase{
// SKIP: rslint does not parse Flow
{
	Skip: true,
	Code: `
function Example({ prop }) {
  const bar = useEffect(<T>(a: T): Hello => {
    prop();
  }, [prop]);
}
`,
	Tsx:  true,
},

// SKIP: rslint does not parse Flow
{
	Skip: true,
	Code: `
function MyComponent() {
  type ColumnKey = 'id' | 'name';
  type Item = {id: string, name: string};

  const columns = useMemo(
    () => [
      {
        type: 'text',
        key: 'id',
      } as TextColumn<ColumnKey, Item>,
    ],
    [],
  );
}
`,
	Tsx:  true,
},
}

var upstreamFlowInvalid = []rule_tester.InvalidTestCase{
// SKIP: rslint does not parse Flow
{
	Skip: true,
	Code: `
hook useExample(a) {
  useEffect(() => {
    console.log(a);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'a'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
hook useExample(a) {
  useEffect(() => {
    console.log(a);
  }, [a]);
}
`}}},
	},
},

// SKIP: rslint does not parse Flow
{
	Skip: true,
	Code: `
function Foo() {
  const foo = ({}: any);
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},
}

func TestExhaustiveDeps_UpstreamFlow(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamFlowValid,
		upstreamFlowInvalid,
	)
}
