package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_SetStateNoDeps holds the upstream cases whose primary
// diagnostic falls into the "setstate_no_deps" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamSetStateNoDepsValid = []rule_tester.ValidTestCase{

}

var upstreamSetStateNoDepsInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function Hello() {
  const [state, setState] = useState(0);
  useEffect(() => {
    setState({});
  });
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect contains a call to 'setState'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [] as a second argument to the useEffect Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Hello() {
  const [state, setState] = useState(0);
  useEffect(() => {
    setState({});
  }, []);
}
`}}},
	},
},

{
	Code: `
function Hello() {
  const [data, setData] = useState(0);
  useEffect(() => {
    fetchData.then(setData);
  });
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect contains a call to 'setData'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [] as a second argument to the useEffect Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Hello() {
  const [data, setData] = useState(0);
  useEffect(() => {
    fetchData.then(setData);
  }, []);
}
`}}},
	},
},

{
	Code: `
function Hello({ country }) {
  const [data, setData] = useState(0);
  useEffect(() => {
    fetchData(country).then(setData);
  });
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect contains a call to 'setData'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [country] as a second argument to the useEffect Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Hello({ country }) {
  const [data, setData] = useState(0);
  useEffect(() => {
    fetchData(country).then(setData);
  }, [country]);
}
`}}},
	},
},

{
	Code: `
function Hello({ prop1, prop2 }) {
  const [state, setState] = useState(0);
  useEffect(() => {
    if (prop1) {
      setState(prop2);
    }
  });
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect contains a call to 'setState'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [prop1, prop2] as a second argument to the useEffect Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Hello({ prop1, prop2 }) {
  const [state, setState] = useState(0);
  useEffect(() => {
    if (prop1) {
      setState(prop2);
    }
  }, [prop1, prop2]);
}
`}}},
	},
},
}

func TestExhaustiveDeps_Upstream_SetStateNoDeps(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamSetStateNoDepsValid,
		upstreamSetStateNoDepsInvalid,
	)
}
