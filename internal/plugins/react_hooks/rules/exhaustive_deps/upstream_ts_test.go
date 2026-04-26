package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_UpstreamTs ports the small "upstream_ts_test" upstream group.

var upstreamTsValid = []rule_tester.ValidTestCase{
{
	Code: `
function MyComponent() {
  const ref = useRef() as React.MutableRefObject<HTMLDivElement>;
  useEffect(() => {
    console.log(ref.current);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber);
  }, [])
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useSpecialEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber);
  })
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useSpecialEffect", "experimental_autoDependenciesHooks": []interface{}{"useSpecialEffect"}},
},

{
	Code: `
function App() {
  const foo = {x: 1};
  React.useEffect(() => {
    const bar = {x: 2};
    const baz = bar as typeof foo;
    console.log(baz);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function App(props) {
  React.useEffect(() => {
    console.log(props.test);
  }, [props.test] as const);
}
`,
	Tsx:  true,
},

{
	Code: `
function App(props) {
  React.useEffect(() => {
    console.log(props.test);
  }, [props.test] as any);
}
`,
	Tsx:  true,
},

{
	Code: `
function App(props) {
  React.useEffect((() => {
    console.log(props.test);
  }) as any, [props.test]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing<T>(): void {
  useEffect(() => {
    let foo: T;
    console.log(foo);
  }, []);
}
`,
	Tsx:  true,
},
}

var upstreamTsInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent() {
  const local = {} as string;
  useEffect(() => {
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {} as string;
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function App() {
  const foo = {x: 1};
  const bar = {x: 2};
  useEffect(() => {
    const baz = bar as typeof foo;
    console.log(baz);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'bar'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function App() {
  const foo = {x: 1};
  const bar = {x: 2};
  useEffect(() => {
    const baz = bar as typeof foo;
    console.log(baz);
  }, [bar]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza.crust,
    toppings: pizza?.toppings,
  }), []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'pizza.crust' and 'pizza?.toppings'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza.crust,
    toppings: pizza?.toppings,
  }), [pizza.crust, pizza?.toppings]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza?.crust,
    density: pizza.crust.density,
  }), []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'pizza.crust'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza?.crust,
    density: pizza.crust.density,
  }), [pizza.crust]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza.crust,
    density: pizza?.crust.density,
  }), []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'pizza.crust'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza.crust,
    density: pizza?.crust.density,
  }), [pizza.crust]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza?.crust,
    density: pizza?.crust.density,
  }), []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'pizza?.crust'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const pizza = {};

  useEffect(() => ({
    crust: pizza?.crust,
    density: pizza?.crust.density,
  }), [pizza?.crust]);
}
`}}},
	},
},

{
	Code: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props.upperViewHeight;
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props.upperViewHeight'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props.upperViewHeight;
  }, [props.upperViewHeight]);
}
`}}},
	},
},

{
	Code: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props?.upperViewHeight;
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props?.upperViewHeight'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props?.upperViewHeight;
  }, [props?.upperViewHeight]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber + state);
  }, [])
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'state'. Either include it or remove the dependency array. You can also do a functional update 'setState(s => ...)' if you only need 'state' in the 'setState' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber + state);
  }, [state])
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useSpecialEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber + state);
  }, [])
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useSpecialEffect", "experimental_autoDependenciesHooks": []interface{}{"useSpecialEffect"}},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useSpecialEffect has a missing dependency: 'state'. Either include it or remove the dependency array. You can also do a functional update 'setState(s => ...)' if you only need 'state' in the 'setState' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useSpecialEffect(() => {
    const someNumber: typeof state = 2;
    setState(prevState => prevState + someNumber + state);
  }, [state])
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useMemo(() => {
    const someNumber: typeof state = 2;
    console.log(someNumber);
  }, [state])
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useMemo has an unnecessary dependency: 'state'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const [state, setState] = React.useState<number>(0);

  useMemo(() => {
    const someNumber: typeof state = 2;
    console.log(someNumber);
  }, [])
}
`}}},
	},
},

{
	Code: `
function Foo() {
  const foo = {} as any;
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

{
	Code: `
function useCustomCallback(callback, deps) {
  return useCallback(callback as any, deps)
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback received a function whose dependencies are unknown. Pass an inline function instead."},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  });
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"requireExplicitEffectDeps": true},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect always requires dependencies. Please add a dependency array or an explicit `undefined`"},
	},
},
}

func TestExhaustiveDeps_UpstreamTs(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamTsValid,
		upstreamTsInvalid,
	)
}
