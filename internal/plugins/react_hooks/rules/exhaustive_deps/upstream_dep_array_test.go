package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_DepArray holds the upstream cases whose primary
// diagnostic falls into the "dep_array" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamDepArrayValid = []rule_tester.ValidTestCase{

}

var upstreamDepArrayInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent(props) {
  useSpecialEffect(() => {
    console.log(props.foo);
  }, null);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useSpecialEffect"},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useSpecialEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies."},
		{Message: "React Hook useSpecialEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useSpecialEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  useEffect(() => {}, ['foo']);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' literal is not a valid dependency because it never changes. You can safely remove it."},
	},
},

{
	Code: `
function MyComponent() {
  const dependencies = [];
  useEffect(() => {}, dependencies);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies."},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const dependencies = [local];
  useEffect(() => {
    console.log(local);
  }, dependencies);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies."},
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const dependencies = [local];
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    console.log(local);
  }, [local, ...dependencies]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies."},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.items[0]);
  }, [props.items, props.items[0]]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent({ items }) {
  useEffect(() => {
    console.log(items[0]);
  }, [items, items[0]]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {}, [props?.attribute.method()]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {}, [props.method()]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},
}

func TestExhaustiveDeps_Upstream_DepArray(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamDepArrayValid,
		upstreamDepArrayInvalid,
	)
}
