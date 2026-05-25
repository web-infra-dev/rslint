package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_Unnecessary holds the upstream cases whose primary
// diagnostic falls into the "unnecessary" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamUnnecessaryValid = []rule_tester.ValidTestCase{

}

var upstreamUnnecessaryInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent() {
  const local1 = {};
  const local2 = {};
  useMemo(() => {
    console.log(local1);
  }, [local1, local2]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useMemo has an unnecessary dependency: 'local2'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = {};
  const local2 = {};
  useMemo(() => {
    console.log(local1);
  }, [local1]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  useCallback(() => {}, [window]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'window'. Either exclude it or remove the dependency array. Outer scope values like 'window' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  useCallback(() => {}, []);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let local = props.foo;
  useCallback(() => {}, [local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'local'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let local = props.foo;
  useCallback(() => {}, []);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const local = {};
  useCallback(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props, props.foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'props.foo'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const local = {};
  useCallback(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props]);
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
  }, [local.id, local]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'local.id'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
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
  }, [props.foo.bar.baz, props.foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'props.foo.bar.baz'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = {};
  useCallback(() => {
    const local1 = {};
    console.log(local1);
  }, [local1]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'local1'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = {};
  useCallback(() => {
    const local1 = {};
    console.log(local1);
  }, []);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const local1 = {};
  useCallback(() => {}, [local1]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'local1'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const local1 = {};
  useCallback(() => {}, []);
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
  }, [ref1.current, ref2.current, props.someOtherRefs, props.color]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has unnecessary dependencies: 'ref1.current' and 'ref2.current'. Either exclude them or remove the dependency array. Mutable values like 'ref1.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1.current.focus();
    console.log(ref2.current.textContent);
    alert(props.someOtherRefs.current.innerHTML);
    fetch(props.color);
  }, [props.someOtherRefs, props.color]);
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
    ref1?.current?.focus();
    console.log(ref2?.current?.textContent);
    alert(props.someOtherRefs.current.innerHTML);
    fetch(props.color);
  }, [ref1?.current, ref2?.current, props.someOtherRefs, props.color]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has unnecessary dependencies: 'ref1.current' and 'ref2.current'. Either exclude them or remove the dependency array. Mutable values like 'ref1.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1?.current?.focus();
    console.log(ref2?.current?.textContent);
    alert(props.someOtherRefs.current.innerHTML);
    fetch(props.color);
  }, [props.someOtherRefs, props.color]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, [ref.current]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has an unnecessary dependency: 'ref.current'. Either exclude it or remove the dependency array. Mutable values like 'ref.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, []);
}
`}}},
	},
},

{
	Code: `
function MyComponent({ activeTab }) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1.current.scrollTop = 0;
    ref2.current.scrollTop = 0;
  }, [ref1.current, ref2.current, activeTab]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has unnecessary dependencies: 'ref1.current' and 'ref2.current'. Either exclude them or remove the dependency array. Mutable values like 'ref1.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ activeTab }) {
  const ref1 = useRef();
  const ref2 = useRef();
  useEffect(() => {
    ref1.current.scrollTop = 0;
    ref2.current.scrollTop = 0;
  }, [activeTab]);
}
`}}},
	},
},

{
	Code: `
function MyComponent({ activeTab, initY }) {
  const ref1 = useRef();
  const ref2 = useRef();
  const fn = useCallback(() => {
    ref1.current.scrollTop = initY;
    ref2.current.scrollTop = initY;
  }, [ref1.current, ref2.current, activeTab, initY]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has unnecessary dependencies: 'activeTab', 'ref1.current', and 'ref2.current'. Either exclude them or remove the dependency array. Mutable values like 'ref1.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent({ activeTab, initY }) {
  const ref1 = useRef();
  const ref2 = useRef();
  const fn = useCallback(() => {
    ref1.current.scrollTop = initY;
    ref2.current.scrollTop = initY;
  }, [initY]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, [ref.current, ref]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has an unnecessary dependency: 'ref.current'. Either exclude it or remove the dependency array. Mutable values like 'ref.current' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, [ref]);
}
`}}},
	},
},

{
	Code: `
function MyComponent() {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [window]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has an unnecessary dependency: 'window'. Either exclude it or remove the dependency array. Outer scope values like 'window' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent() {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);
}
`}}},
	},
},

{
	Code: `
import MutableStore from 'store';

function MyComponent() {
  useEffect(() => {
    console.log(MutableStore.hello);
  }, [MutableStore.hello]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has an unnecessary dependency: 'MutableStore.hello'. Either exclude it or remove the dependency array. Outer scope values like 'MutableStore.hello' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
import MutableStore from 'store';

function MyComponent() {
  useEffect(() => {
    console.log(MutableStore.hello);
  }, []);
}
`}}},
	},
},

{
	Code: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    useEffect(() => {
      console.log(MutableStore.hello.world, props.foo, x, y, z, global.stuff);
    }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has unnecessary dependencies: 'MutableStore.hello.world', 'global.stuff', and 'z'. Either exclude them or remove the dependency array. Outer scope values like 'MutableStore.hello.world' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    useEffect(() => {
      console.log(MutableStore.hello.world, props.foo, x, y, z, global.stuff);
    }, [props.foo, x, y]);
  }
}
`}}},
	},
},

{
	Code: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    useEffect(() => {
      // nothing
    }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has unnecessary dependencies: 'MutableStore.hello.world', 'global.stuff', and 'z'. Either exclude them or remove the dependency array. Outer scope values like 'MutableStore.hello.world' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    useEffect(() => {
      // nothing
    }, [props.foo, x, y]);
  }
}
`}}},
	},
},

{
	Code: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    const fn = useCallback(() => {
      // nothing
    }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has unnecessary dependencies: 'MutableStore.hello.world', 'global.stuff', 'props.foo', 'x', 'y', and 'z'. Either exclude them or remove the dependency array. Outer scope values like 'MutableStore.hello.world' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    const fn = useCallback(() => {
      // nothing
    }, []);
  }
}
`}}},
	},
},

{
	Code: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    const fn = useCallback(() => {
      // nothing
    }, [MutableStore?.hello?.world, props.foo, x, y, z, global?.stuff]);
  }
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has unnecessary dependencies: 'MutableStore.hello.world', 'global.stuff', 'props.foo', 'x', 'y', and 'z'. Either exclude them or remove the dependency array. Outer scope values like 'MutableStore.hello.world' aren't valid dependencies because mutating them doesn't re-render the component.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
import MutableStore from 'store';
let z = {};

function MyComponent(props) {
  let x = props.foo;
  {
    let y = props.bar;
    const fn = useCallback(() => {
      // nothing
    }, []);
  }
}
`}}},
	},
},

{
	Code: `
function Thing() {
  useEffect(() => {
    const fetchData = async () => {};
    fetchData();
  }, [fetchData]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useEffect has an unnecessary dependency: 'fetchData'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Thing() {
  useEffect(() => {
    const fetchData = async () => {};
    fetchData();
  }, []);
}
`}}},
	},
},

{
	Code: `
function Example() {
  const foo = useCallback(() => {
    foo();
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "React Hook useCallback has an unnecessary dependency: 'foo'. Either exclude it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function Example() {
  const foo = useCallback(() => {
    foo();
  }, []);
}
`}}},
	},
},
}

func TestExhaustiveDeps_Upstream_Unnecessary(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamUnnecessaryValid,
		upstreamUnnecessaryInvalid,
	)
}
