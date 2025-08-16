package rules_of_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/react-hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRulesOfHooks(t *testing.T) {
	errors := make([]rule_tester.InvalidTestCaseError, 6)
	for i, err := range errors {
		err.MessageId = "import/no-self-import"
		err.Line = i + 2
		err.Column = 1
		errors[i] = err
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&rules_of_hooks.RulesOfHooksRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
// Valid because components can use hooks.
function ComponentWithHook() {
	useHook();
}
				`,
			},
			{
				Code: `
// Valid because components can use hooks.
function createComponentWithHook() {
  return function ComponentWithHook() {
    useHook();
  };
}
				`,
			},
			{
				Code: `
// Valid because hooks can use hooks.
function useHookWithHook() {
  useHook();
}
				`,
			},
			{
				Code: `
// Valid because hooks can use hooks.
function createHook() {
  return function useHookWithHook() {
    useHook();
  }
}
				`,
			},
			{
				Code: `
// Valid because components can call functions.
function ComponentWithNormalFunction() {
  doSomething();
}
				`,
			},
			{
				Code: `
// Valid because functions can call functions.
function normalFunctionWithNormalFunction() {
  doSomething();
}
				`,
			},
			{
				Code: `
// Valid because functions can call functions.
function normalFunctionWithConditionalFunction() {
  if (cond) {
    doSomething();
  }
}
				`,
			},
			{
				Code: `
// Valid because functions can call functions.
function functionThatStartsWithUseButIsntAHook() {
  if (cond) {
    userFetch();
  }
}
				`,
			},
			{
				Code: `
// Valid although unconditional return doesn't make sense and would fail other rules.
// We could make it invalid but it doesn't matter.
function useUnreachable() {
  return;
  useHook();
}
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function useHook() { useState(); }
const whatever = function useHook() { useState(); };
const useHook1 = () => { useState(); };
let useHook2 = () => useState();
useHook2 = () => { useState(); };
({useHook: () => { useState(); }});
({useHook() { useState(); }});
const {useHook3 = () => { useState(); }} = {};
({useHook = () => { useState(); }} = {});
Namespace.useHook = () => { useState(); };
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function useHook() {
  useHook1();
  useHook2();
}
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function createHook() {
  return function useHook() {
    useHook1();
    useHook2();
  };
}
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function useHook() {
  useState() && a;
}
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function useHook() {
  return useHook1() + useHook2();
}
				`,
			},
			{
				Code: `
// Valid because hooks can call hooks.
function useHook() {
  return useHook1(useHook2());
}
				`,
			},
			{
				Code: `
// Valid because hooks can be used in anonymous arrow-function arguments
// to forwardRef.
const FancyButton = React.forwardRef((props, ref) => {
  useHook();
  return <button {...props} ref={ref} />
});
				`,
			},
			{
				Code: `
// Valid because hooks can be used in anonymous function arguments to
// forwardRef.
const FancyButton = React.forwardRef(function (props, ref) {
  useHook();
  return <button {...props} ref={ref} />
});
				`,
			},
			{
				Code: `
// Valid because hooks can be used in anonymous function arguments to
// forwardRef.
const FancyButton = forwardRef(function (props, ref) {
  useHook();
  return <button {...props} ref={ref} />
});
				`,
			},
			{
				Code: `
// Valid because hooks can be used in anonymous function arguments to
// React.memo.
const MemoizedFunction = React.memo(props => {
  useHook();
  return <button {...props} />
});
				`,
			},
			{
				Code: `
// Valid because hooks can be used in anonymous function arguments to
// memo.
const MemoizedFunction = memo(function (props) {
  useHook();
  return <button {...props} />
});
				`,
			},
			{
				Code: `
// Valid because classes can call functions.
// We don't consider these to be hooks.
class C {
  m() {
    this.useHook();
    super.useHook();
  }
}
				`,
			},
			{
				Code: `
// Valid -- this is a regression test.
jest.useFakeTimers();
beforeEach(() => {
  jest.useRealTimers();
})
				`,
			},
			{
				Code: `
// Valid because they're not matching use[A-Z].
fooState();
_use();
_useState();
use_hook();
// also valid because it's not matching the PascalCase namespace
jest.useFakeTimer()
				`,
			},
			{
				Code: `
// Regression test for some internal code.
// This shows how the "callback rule" is more relaxed,
// and doesn't kick in unless we're confident we're in
// a component or a hook.
function makeListener(instance) {
  each(pixelsWithInferredEvents, pixel => {
    if (useExtendedSelector(pixel.id) && extendedButton) {
      foo();
    }
  });
}
				`,
			},
			{
				Code: `
// This is valid because "use"-prefixed functions called in
// unnamed function arguments are not assumed to be hooks.
React.unknownFunction((foo, bar) => {
  if (foo) {
    useNotAHook(bar)
  }
});
				`,
			},
			{
				Code: `
// This is valid because "use"-prefixed functions called in
// unnamed function arguments are not assumed to be hooks.
unknownFunction(function(foo, bar) {
  if (foo) {
    useNotAHook(bar)
  }
});
				`,
			},
			{
				Code: `
// Regression test for incorrectly flagged valid code.
function RegressionTest() {
  const foo = cond ? a : b;
  useState();
}
				`,
			},
			{
				Code: `
// Valid because exceptions abort rendering
function RegressionTest() {
  if (page == null) {
    throw new Error('oh no!');
  }
  useState();
}
				`,
			},
			{
				Code: `
// Valid because the loop doesn't change the order of hooks calls.
function RegressionTest() {
  const res = [];
  const additionalCond = true;
  for (let i = 0; i !== 10 && additionalCond; ++i ) {
    res.push(i);
  }
  React.useLayoutEffect(() => {});
}
				`,
			},
			{
				Code: `
// Is valid but hard to compute by brute-forcing
function MyComponent() {
  // 40 conditions
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}
  if (c) {} else {}

  // 10 hooks
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
  useHook();
}
				`,
			},
			{
				Code: `
// Valid because the neither the conditions before or after the hook affect the hook call
// Failed prior to implementing BigInt because pathsFromStartToEnd and allPathsFromStartToEnd were too big and had rounding errors
const useSomeHook = () => {};

const SomeName = () => {
  const filler = FILLER ?? FILLER ?? FILLER;
  const filler2 = FILLER ?? FILLER ?? FILLER;
  const filler3 = FILLER ?? FILLER ?? FILLER;
  const filler4 = FILLER ?? FILLER ?? FILLER;
  const filler5 = FILLER ?? FILLER ?? FILLER;
  const filler6 = FILLER ?? FILLER ?? FILLER;
  const filler7 = FILLER ?? FILLER ?? FILLER;
  const filler8 = FILLER ?? FILLER ?? FILLER;

  useSomeHook();

  if (anyConditionCanEvenBeFalse) {
    return null;
  }

  return (
    <React.Fragment>
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
      {FILLER ? FILLER : FILLER}
    </React.Fragment>
  );
};
				`,
			},
			{
				Code: `
// Valid because the neither the condition nor the loop affect the hook call.
function App(props) {
  const someObject = {propA: true};
  for (const propName in someObject) {
    if (propName === true) {
    } else {
    }
  }
  const [myState, setMyState] = useState(null);
}
				`,
			},
			{
				Code: `
function App() {
  const text = use(Promise.resolve('A'));
  return <Text text={text} />
}
				`,
			},
			{
				Code: `
import * as React from 'react';
function App() {
  if (shouldShowText) {
    const text = use(query);
    const data = React.use(thing);
    const data2 = react.use(thing2);
    return <Text text={text} />
  }
  return <Text text={shouldFetchBackupText ? use(backupQuery) : "Nothing to see here"} />
}
				`,
			},
			{
				Code: `
function App() {
  let data = [];
  for (const query of queries) {
    const text = use(item);
    data.push(text);
  }
  return <Child data={data} />
}
				`,
			},
			{
				Code: `
function App() {
  const data = someCallback((x) => use(x));
  return <Child data={data} />
}
				`,
			},
			{
				Code: `
export const notAComponent = () => {
   return () => {
    useState();
  }
}
				`,
			},
			{
				Code: `
export default () => {
  if (isVal) {
    useState(0);
  }
}
				`,
			},
			{
				Code: `
function notAComponent() {
  return new Promise.then(() => {
    useState();
  });
}
				`,
			},
			{
				Code: `
// Valid because the hook is outside of the loop
const Component = () => {
  const [state, setState] = useState(0);
  for (let i = 0; i < 10; i++) {
    console.log(i);
  }
  return <div></div>;
};
				`,
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
