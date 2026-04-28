package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_Missing holds the upstream cases whose primary
// diagnostic falls into the "missing" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamMissingValid = []rule_tester.ValidTestCase{

}

var upstreamMissingInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.toString());
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.toString());
  }, [props.foo]);
}
`}}},
	},
},

{
	Code: `
function ComponentUsingFormState(props) {
  const [state7, dispatch3] = useFormState();
  const [state8, dispatch4] = ReactDOM.useFormState();
  useEffect(() => {
    dispatch3();
    dispatch4();

    // dynamic
    console.log(state7);
    console.log(state8);

  }, [state7, state8]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'dispatch3' and 'dispatch4'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function ComponentUsingFormState(props) {
  const [state7, dispatch3] = useFormState();
  const [state8, dispatch4] = ReactDOM.useFormState();
  useEffect(() => {
    dispatch3();
    dispatch4();

    // dynamic
    console.log(state7);
    console.log(state8);

  }, [dispatch3, dispatch4, state7, state8]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar.baz);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo?.bar.baz'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar.baz);
  }, [props.foo?.bar.baz]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar?.baz);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo?.bar?.baz'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar?.baz);
  }, [props.foo?.bar?.baz]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar.toString());
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo?.bar'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar.toString());
  }, [props.foo?.bar]);
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
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function Counter(unstableProp) {
  let [count, setCount] = useState(0);
  setCount = unstableProp
  useEffect(() => {
    let id = setInterval(() => {
      setCount(c => c + 1);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'setCount'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter(unstableProp) {
  let [count, setCount] = useState(0);
  setCount = unstableProp
  useEffect(() => {
    let id = setInterval(() => {
      setCount(c => c + 1);
    }, 1000);
    return () => clearInterval(id);
  }, [setCount]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  let local = 42;
  useEffect(() => {
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  let local = 42;
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
  const local = /foo/;
  useEffect(() => {
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = /foo/;
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
    if (true) {
      console.log(local);
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    if (true) {
      console.log(local);
    }
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
    try {
      console.log(local);
    } finally {}
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    try {
      console.log(local);
    } finally {}
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
    function inner() {
      console.log(local);
    }
    inner();
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    function inner() {
      console.log(local);
    }
    inner();
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = someFunc();
  {
    const local2 = someFunc();
    useEffect(() => {
      console.log(local1);
      console.log(local2);
    }, []);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'local1' and 'local2'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = someFunc();
  {
    const local2 = someFunc();
    useEffect(() => {
      console.log(local1);
      console.log(local2);
    }, [local1, local2]);
  }
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = {};
  const local2 = {};
  useEffect(() => {
    console.log(local1);
    console.log(local2);
  }, [local1]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local2'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = {};
  const local2 = {};
  useEffect(() => {
    console.log(local1);
    console.log(local2);
  }, [local1, local2]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = someFunc();
  function MyNestedComponent() {
    const local2 = {};
    useCallback(() => {
      console.log(local1);
      console.log(local2);
    }, [local1]);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'local2'. Either include it or remove the dependency array. Outer scope values like 'local1' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = someFunc();
  function MyNestedComponent() {
    const local2 = {};
    useCallback(() => {
      console.log(local1);
      console.log(local2);
    }, [local2]);
  }
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
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
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
function MyComponent({ history }) {
  useEffect(() => {
    return history.listen();
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'history'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ history }) {
  useEffect(() => {
    return history.listen();
  }, [history]);
}
`}}},
	},
},

{
	Code: `
function MyComponent({ history }) {
  useEffect(() => {
    return [
      history.foo.bar[2].dobedo.listen(),
      history.foo.bar().dobedo.listen[2]
    ];
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'history.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ history }) {
  useEffect(() => {
    return [
      history.foo.bar[2].dobedo.listen(),
      history.foo.bar().dobedo.listen[2]
    ];
  }, [history.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent({ history }) {
  useEffect(() => {
    return [
      history?.foo
    ];
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'history?.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ history }) {
  useEffect(() => {
    return [
      history?.foo
    ];
  }, [history?.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent({ foo, bar, baz }) {
  useEffect(() => {
    console.log(foo, bar, baz);
  }, ['foo', 'bar']);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'bar', 'baz', and 'foo'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ foo, bar, baz }) {
  useEffect(() => {
    console.log(foo, bar, baz);
  }, [bar, baz, foo]);
}
`}}},
		{Message: "The 'foo' literal is not a valid dependency because it never changes. Did you mean to include foo in the array instead?"},
		{Message: "The 'bar' literal is not a valid dependency because it never changes. Did you mean to include bar in the array instead?"},
	},
},

{
	Code: `
function MyComponent({ foo, bar, baz }) {
  useEffect(() => {
    console.log(foo, bar, baz);
  }, [42, false, null]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'bar', 'baz', and 'foo'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ foo, bar, baz }) {
  useEffect(() => {
    console.log(foo, bar, baz);
  }, [bar, baz, foo]);
}
`}}},
		{Message: "The 42 literal is not a valid dependency because it never changes. You can safely remove it."},
		{Message: "The false literal is not a valid dependency because it never changes. You can safely remove it."},
		{Message: "The null literal is not a valid dependency because it never changes. You can safely remove it."},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const dependencies = [local];
  useEffect(() => {
    console.log(local);
  }, [...dependencies]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const dependencies = [local];
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
		{Message: "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies."},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [computeCacheKey(local)]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.items[0]);
  }, [props.items[0]]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props.items'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.items[0]);
  }, [props.items]);
}
`}}},
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent({ items }) {
  useEffect(() => {
    console.log(items[0]);
  }, [items[0]]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'items'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ items }) {
  useEffect(() => {
    console.log(items[0]);
  }, [items]);
}
`}}},
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent(props) {
  const local = {};
  useCallback(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has missing dependencies: 'props.bar' and 'props.foo'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const local = {};
  useCallback(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props.bar, props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {id: 42};
  useEffect(() => {
    console.log(local);
  }, [local.id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {id: 42};
  useEffect(() => {
    console.log(local);
  }, [local, local.id]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {id: 42};
  const fn = useCallback(() => {
    console.log(local);
  }, [local.id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {id: 42};
  const fn = useCallback(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo.bar.baz'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo.bar.baz]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let color = {}
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
    console.log(color);
  }, [props.foo, props.foo.bar.baz]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'color'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let color = {}
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
    console.log(color);
  }, [color, props.foo.bar.baz]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
    console.log(props.foo.fizz.bizz);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has missing dependencies: 'props.foo.bar.baz' and 'props.foo.fizz.bizz'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
    console.log(props.foo.fizz.bizz);
  }, [props.foo.bar.baz, props.foo.fizz.bizz]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar);
  }, [props.foo.bar.baz]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props.foo.bar'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar);
  }, [props.foo.bar]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props);
    console.log(props.hello);
  }, [props.foo.bar.baz]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'props'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props);
    console.log(props.hello);
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'props.bar' and 'props.foo'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props.bar, props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, [c, a, g]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'b', 'd', 'e', and 'f'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, [c, a, g, b, e, d, f]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, [a, c, g]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'b', 'd', 'e', and 'f'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, [a, b, c, d, e, f, g]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'a', 'b', 'c', 'd', 'e', 'f', and 'g'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let a, b, c, d, e, f, g;
  useEffect(() => {
    console.log(b, e, d, c, a, g, f);
  }, [a, b, c, d, e, f, g]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const local = {};
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'local', 'props.bar', and 'props.foo'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const local = {};
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
    console.log(local);
  }, [local, props.bar, props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const local = {};
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
    console.log(local);
  }, [props]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const local = {};
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
    console.log(local);
  }, [local, props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook useCallback has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, [props.foo]);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook useMemo has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook React.useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook React.useCallback has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.useMemo(() => {
    console.log(props.foo);
  }, []);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook React.useMemo has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, []);
  useCallback(() => {
    console.log(props.foo);
  }, []);
  useMemo(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCallback(() => {
    console.log(props.foo);
  }, []);
  React.useMemo(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.notReactiveHook(() => {
    console.log(props.foo);
  }, []);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
  useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useCustomEffect"},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCustomEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
  useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`}}},
		{Message: "React Hook React.useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
  useEffect(() => {
    console.log(props.foo);
  }, []);
  React.useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  React.useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`}}},
	},
},

// SKIP: unsupported settings shape
{
	Skip: true,
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCustomEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
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
  }, [a ? local : b]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [a && local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
		{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
	},
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  const [state, setState] = useState();
  useEffect(() => {
    ref.current = {};
    setState(state + 1);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'state'. Either include it or remove the dependency array. You can also do a functional update 'setState(s => ...)' if you only need 'state' in the 'setState' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const ref = useRef();
  const [state, setState] = useState();
  useEffect(() => {
    ref.current = {};
    setState(state + 1);
  }, [state]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  const [state, setState] = useState();
  useEffect(() => {
    ref.current = {};
    setState(state + 1);
  }, [ref]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'state'. Either include it or remove the dependency array. You can also do a functional update 'setState(s => ...)' if you only need 'state' in the 'setState' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const ref = useRef();
  const [state, setState] = useState();
  useEffect(() => {
    ref.current = {};
    setState(state + 1);
  }, [ref, state]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1.current.focus();
    console.log(ref2.current.textContent);
    alert(props.someOtherRefs.current.innerHTML);
    fetch(props.color);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'props.color' and 'props.someOtherRefs'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1.current.focus();
    console.log(ref2.current.textContent);
    alert(props.someOtherRefs.current.innerHTML);
    fetch(props.color);
  }, [props.color, props.someOtherRefs]);
}
`}}},
	},
},

{
	Code: `
const MyComponent = forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    focus() {
      alert(props.hello);
    }
  }), [])
});
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useImperativeHandle has a missing dependency: 'props.hello'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
const MyComponent = forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    focus() {
      alert(props.hello);
    }
  }), [props.hello])
});
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    if (props.onChange) {
      props.onChange();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    if (props.onChange) {
      props.onChange();
    }
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    if (props?.onChange) {
      props?.onChange();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    if (props?.onChange) {
      props?.onChange();
    }
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    function play() {
      props.onPlay();
    }
    function pause() {
      props.onPause();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    function play() {
      props.onPlay();
    }
    function pause() {
      props.onPause();
    }
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    if (props.foo.onChange) {
      props.foo.onChange();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    if (props.foo.onChange) {
      props.foo.onChange();
    }
  }, [props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    props.onChange();
    if (props.foo.onChange) {
      props.foo.onChange();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    props.onChange();
    if (props.foo.onChange) {
      props.foo.onChange();
    }
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const [skillsCount] = useState();
  useEffect(() => {
    if (skillsCount === 0 && !props.isEditMode) {
      props.toggleEditMode();
    }
  }, [skillsCount, props.isEditMode, props.toggleEditMode]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const [skillsCount] = useState();
  useEffect(() => {
    if (skillsCount === 0 && !props.isEditMode) {
      props.toggleEditMode();
    }
  }, [skillsCount, props.isEditMode, props.toggleEditMode, props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const [skillsCount] = useState();
  useEffect(() => {
    if (skillsCount === 0 && !props.isEditMode) {
      props.toggleEditMode();
    }
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'props' and 'skillsCount'. Either include them or remove the dependency array. However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the useEffect call and refer to those specific props inside useEffect.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const [skillsCount] = useState();
  useEffect(() => {
    if (skillsCount === 0 && !props.isEditMode) {
      props.toggleEditMode();
    }
  }, [props, skillsCount]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    externalCall(props);
    props.onChange();
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    externalCall(props);
    props.onChange();
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    props.onChange();
    externalCall(props);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'props'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  useEffect(() => {
    props.onChange();
    externalCall(props);
  }, [props]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = 42;
  const local2 = '42';
  const local3 = null;
  const local4 = {};
  useEffect(() => {
    console.log(local1);
    console.log(local2);
    console.log(local3);
    console.log(local4);
  }, [local1, local3]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local4'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = 42;
  const local2 = '42';
  const local3 = null;
  const local4 = {};
  useEffect(() => {
    console.log(local1);
    console.log(local2);
    console.log(local3);
    console.log(local4);
  }, [local1, local3, local4]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    setTimeout(() => console.log(taint));
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'handleNext1'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    setTimeout(() => console.log(taint));
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, [handleNext1]);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useLayoutEffect has a missing dependency: 'handleNext2'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    setTimeout(() => console.log(taint));
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, [handleNext2]);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useMemo has a missing dependency: 'handleNext3'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    setTimeout(() => console.log(taint));
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, [handleNext3]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  function handleChange() {}

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'handleNext1'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  function handleChange() {}

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, [handleNext1]);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useLayoutEffect has a missing dependency: 'handleNext2'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  function handleChange() {}

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, [handleNext2]);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useMemo has a missing dependency: 'handleNext3'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  function handleChange() {}

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, [handleNext3]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  const handleChange = () => {};

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'handleNext1'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  const handleChange = () => {};

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, [handleNext1]);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useLayoutEffect has a missing dependency: 'handleNext2'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  const handleChange = () => {};

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, [handleNext2]);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`}}},
		{Message: "React Hook useMemo has a missing dependency: 'handleNext3'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();
  let taint = props.foo;

  // Shouldn't affect anything
  const handleChange = () => {};

  function handleNext1(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(taint(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(taint);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, [handleNext3]);
}
`}}},
	},
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count + 1);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'count'. Either include it or remove the dependency array. You can also do a functional update 'setCount(c => ...)' if you only need 'count' in the 'setCount' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter() {
  let [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count + 1);
    }, 1000);
    return () => clearInterval(id);
  }, [count]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);
  let [increment, setIncrement] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'count' and 'increment'. Either include them or remove the dependency array. You can also do a functional update 'setCount(c => ...)' if you only need 'count' in the 'setCount' call.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter() {
  let [count, setCount] = useState(0);
  let [increment, setIncrement] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, [count, increment]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);
  let [increment, setIncrement] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'increment'. Either include it or remove the dependency array. You can also replace multiple useState variables with useReducer if 'setCount' needs the current value of 'increment'.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter() {
  let [count, setCount] = useState(0);
  let [increment, setIncrement] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, [increment]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);
  let increment = useCustomHook();

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'increment'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter() {
  let [count, setCount] = useState(0);
  let increment = useCustomHook();

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, [increment]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter({ step }) {
  let [count, setCount] = useState(0);

  function increment(x) {
    return x + step;
  }

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => increment(count));
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'increment'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter({ step }) {
  let [count, setCount] = useState(0);

  function increment(x) {
    return x + step;
  }

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => increment(count));
    }, 1000);
    return () => clearInterval(id);
  }, [increment]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter({ increment }) {
  let [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'increment'. Either include it or remove the dependency array. If 'setCount' needs the current value of 'increment', you can also switch to useReducer instead of useState and read 'increment' in the reducer.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter({ increment }) {
  let [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, [increment]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Counter() {
  const [count, setCount] = useState(0);

  function tick() {
    setCount(count + 1);
  }

  useEffect(() => {
    let id = setInterval(() => {
      tick();
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'tick'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Counter() {
  const [count, setCount] = useState(0);

  function tick() {
    setCount(count + 1);
  }

  useEffect(() => {
    let id = setInterval(() => {
      tick();
    }, 1000);
    return () => clearInterval(id);
  }, [tick]);

  return <h1>{count}</h1>;
}
`}}},
	},
},

{
	Code: `
function Podcasts() {
  useEffect(() => {
    alert(podcasts);
  }, []);
  let [podcasts, setPodcasts] = useState(null);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'podcasts'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts() {
  useEffect(() => {
    alert(podcasts);
  }, [podcasts]);
  let [podcasts, setPodcasts] = useState(null);
}
`}}},
	},
},

{
	Code: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    fetchPodcasts(id).then(setPodcasts);
  }, [id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'fetchPodcasts'. Either include it or remove the dependency array. If 'fetchPodcasts' changes too often, find the parent component that defines it and wrap that definition in useCallback.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    fetchPodcasts(id).then(setPodcasts);
  }, [fetchPodcasts, id]);
}
`}}},
	},
},

{
	Code: `
function Podcasts({ api: { fetchPodcasts }, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    fetchPodcasts(id).then(setPodcasts);
  }, [id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'fetchPodcasts'. Either include it or remove the dependency array. If 'fetchPodcasts' changes too often, find the parent component that defines it and wrap that definition in useCallback.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts({ api: { fetchPodcasts }, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    fetchPodcasts(id).then(setPodcasts);
  }, [fetchPodcasts, id]);
}
`}}},
	},
},

{
	Code: `
function Podcasts({ fetchPodcasts, fetchPodcasts2, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    setTimeout(() => {
      console.log(id);
      fetchPodcasts(id).then(setPodcasts);
      fetchPodcasts2(id).then(setPodcasts);
    });
  }, [id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'fetchPodcasts' and 'fetchPodcasts2'. Either include them or remove the dependency array. If 'fetchPodcasts' changes too often, find the parent component that defines it and wrap that definition in useCallback.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts({ fetchPodcasts, fetchPodcasts2, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    setTimeout(() => {
      console.log(id);
      fetchPodcasts(id).then(setPodcasts);
      fetchPodcasts2(id).then(setPodcasts);
    });
  }, [fetchPodcasts, fetchPodcasts2, id]);
}
`}}},
	},
},

{
	Code: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    console.log(fetchPodcasts);
    fetchPodcasts(id).then(setPodcasts);
  }, [id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'fetchPodcasts'. Either include it or remove the dependency array. If 'fetchPodcasts' changes too often, find the parent component that defines it and wrap that definition in useCallback.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    console.log(fetchPodcasts);
    fetchPodcasts(id).then(setPodcasts);
  }, [fetchPodcasts, id]);
}
`}}},
	},
},

{
	Code: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    console.log(fetchPodcasts);
    fetchPodcasts?.(id).then(setPodcasts);
  }, [id]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'fetchPodcasts'. Either include it or remove the dependency array. If 'fetchPodcasts' changes too often, find the parent component that defines it and wrap that definition in useCallback.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Podcasts({ fetchPodcasts, id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    console.log(fetchPodcasts);
    fetchPodcasts?.(id).then(setPodcasts);
  }, [fetchPodcasts, id]);
}
`}}},
	},
},

{
	Code: `
function Example({ prop }) {
  const foo = useCallback(() => {
    prop.hello(foo);
  }, [foo]);
  const bar = useCallback(() => {
    foo();
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has a missing dependency: 'prop'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Example({ prop }) {
  const foo = useCallback(() => {
    prop.hello(foo);
  }, [prop]);
  const bar = useCallback(() => {
    foo();
  }, [foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  function myEffect() {
    console.log(local);
  }
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  function myEffect() {
    console.log(local);
  }
  useEffect(myEffect, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const myEffect = () => {
    console.log(local);
  };
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const myEffect = () => {
    console.log(local);
  };
  useEffect(myEffect, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const myEffect = function() {
    console.log(local);
  };
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const myEffect = function() {
    console.log(local);
  };
  useEffect(myEffect, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const myEffect = () => {
    otherThing();
  };
  const otherThing = () => {
    console.log(local);
  };
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'otherThing'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const myEffect = () => {
    otherThing();
  };
  const otherThing = () => {
    console.log(local);
  };
  useEffect(myEffect, [otherThing]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const myEffect = debounce(() => {
    console.log(local);
  }, delay);
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'myEffect'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const myEffect = debounce(() => {
    console.log(local);
  }, delay);
  useEffect(myEffect, [myEffect]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local = {};
  const myEffect = debounce(() => {
    console.log(local);
  }, delay);
  useEffect(myEffect, [local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'myEffect'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  const myEffect = debounce(() => {
    console.log(local);
  }, delay);
  useEffect(myEffect, [myEffect]);
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
  }, []);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"enableDangerousAutofixThisMayCauseInfiniteLoops": true},
	Output: []string{`
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has a missing dependency: 'local'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let foo = {}
  useEffect(() => {
    foo.bar.baz = 43;
    props.foo.bar.baz = 1;
  }, []);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has missing dependencies: 'foo.bar' and 'props.foo.bar'. Either include them or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let foo = {}
  useEffect(() => {
    foo.bar.baz = 43;
    props.foo.bar.baz = 1;
  }, [foo.bar, props.foo.bar]);
}
`}}},
	},
},
}

func TestExhaustiveDeps_Upstream_Missing(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamMissingValid,
		upstreamMissingInvalid,
	)
}
