package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_CallbackArg holds the upstream cases whose primary
// diagnostic falls into the "callback_arg" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamCallbackArgValid = []rule_tester.ValidTestCase{

}

var upstreamCallbackArgInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent(props) {
  const value = useMemo(() => { return 2*2; });
  const fn = useCallback(() => { alert('foo'); });
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
		{Message: "React Hook useCallback does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
	},
},

{
	Code: `
function MyComponent({ fn1, fn2 }) {
  const value = useMemo(fn1);
  const fn = useCallback(fn2);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
		{Message: "React Hook useCallback does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
	},
},

{
	Code: `
function MyComponent() {
  useEffect()
  useLayoutEffect()
  useCallback()
  useMemo()
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect requires an effect callback. Did you forget to pass a callback to the hook?"},
		{Message: "React Hook useLayoutEffect requires an effect callback. Did you forget to pass a callback to the hook?"},
		{Message: "React Hook useCallback requires an effect callback. Did you forget to pass a callback to the hook?"},
		{Message: "React Hook useMemo requires an effect callback. Did you forget to pass a callback to the hook?"},
	},
},

{
	Code: `
function Thing() {
  useEffect(async () => {}, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "Effect callbacks are synchronous to prevent race conditions. Put the async function inside:\n\nuseEffect(() => {\n  async function fetchData() {\n    // You can await here\n    const response = await MyAPI.getData(someId);\n    // ...\n  }\n  fetchData();\n}, [someId]); // Or [] if effect doesn't need props or state\n\nLearn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching"},
	},
},

{
	Code: `
function Thing() {
  useEffect(async () => {});
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "Effect callbacks are synchronous to prevent race conditions. Put the async function inside:\n\nuseEffect(() => {\n  async function fetchData() {\n    // You can await here\n    const response = await MyAPI.getData(someId);\n    // ...\n  }\n  fetchData();\n}, [someId]); // Or [] if effect doesn't need props or state\n\nLearn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching"},
	},
},

{
	Code: `
function MyComponent({myEffect}) {
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect received a function whose dependencies are unknown. Pass an inline function instead."},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(debounce(() => {
    console.log(local);
  }, delay), []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect received a function whose dependencies are unknown. Pass an inline function instead."},
	},
},

{
	Code: `
function useCustomCallback(callback, deps) {
  return useCallback(callback, deps)
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback received a function whose dependencies are unknown. Pass an inline function instead."},
	},
},
}

func TestExhaustiveDeps_Upstream_CallbackArg(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamCallbackArgValid,
		upstreamCallbackArgInvalid,
	)
}
