package rules_of_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRulesOfHooks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&rules_of_hooks.RulesOfHooksRule,
		[]rule_tester.ValidTestCase{
			// 			{
			// 				Code: `
			// // Valid because components can use hooks.
			// function ComponentWithHook() {
			// 	useHook();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because components can use hooks.
			// function createComponentWithHook() {
			//   return function ComponentWithHook() {
			//     useHook();
			//   };
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can use hooks.
			// function useHookWithHook() {
			//   useHook();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can use hooks.
			// function createHook() {
			//   return function useHookWithHook() {
			//     useHook();
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because components can call functions.
			// function ComponentWithNormalFunction() {
			//   doSomething();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because functions can call functions.
			// function normalFunctionWithNormalFunction() {
			//   doSomething();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because functions can call functions.
			// function normalFunctionWithConditionalFunction() {
			//   if (cond) {
			//     doSomething();
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because functions can call functions.
			// function functionThatStartsWithUseButIsntAHook() {
			//   if (cond) {
			//     userFetch();
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid although unconditional return doesn't make sense and would fail other rules.
			// // We could make it invalid but it doesn't matter.
			// function useUnreachable() {
			//   return;
			//   useHook();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function useHook() { useState(); }
			// const whatever = function useHook() { useState(); };
			// const useHook1 = () => { useState(); };
			// let useHook2 = () => useState();
			// useHook2 = () => { useState(); };
			// ({useHook: () => { useState(); }});
			// ({useHook() { useState(); }});
			// const {useHook3 = () => { useState(); }} = {};
			// ({useHook = () => { useState(); }} = {});
			// Namespace.useHook = () => { useState(); };
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function useHook() {
			//   useHook1();
			//   useHook2();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function createHook() {
			//   return function useHook() {
			//     useHook1();
			//     useHook2();
			//   };
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function useHook() {
			//   useState() && a;
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function useHook() {
			//   return useHook1() + useHook2();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can call hooks.
			// function useHook() {
			//   return useHook1(useHook2());
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can be used in anonymous arrow-function arguments
			// // to forwardRef.
			// const FancyButton = React.forwardRef((props, ref) => {
			//   useHook();
			//   return <button {...props} ref={ref} />
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can be used in anonymous function arguments to
			// // forwardRef.
			// const FancyButton = React.forwardRef(function (props, ref) {
			//   useHook();
			//   return <button {...props} ref={ref} />
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can be used in anonymous function arguments to
			// // forwardRef.
			// const FancyButton = forwardRef(function (props, ref) {
			//   useHook();
			//   return <button {...props} ref={ref} />
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can be used in anonymous function arguments to
			// // React.memo.
			// const MemoizedFunction = React.memo(props => {
			//   useHook();
			//   return <button {...props} />
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because hooks can be used in anonymous function arguments to
			// // memo.
			// const MemoizedFunction = memo(function (props) {
			//   useHook();
			//   return <button {...props} />
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because classes can call functions.
			// // We don't consider these to be hooks.
			// class C {
			//   m() {
			//     this.useHook();
			//     super.useHook();
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid -- this is a regression test.
			// jest.useFakeTimers();
			// beforeEach(() => {
			//   jest.useRealTimers();
			// })
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because they're not matching use[A-Z].
			// fooState();
			// _use();
			// _useState();
			// use_hook();
			// // also valid because it's not matching the PascalCase namespace
			// jest.useFakeTimer()
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Regression test for some internal code.
			// // This shows how the "callback rule" is more relaxed,
			// // and doesn't kick in unless we're confident we're in
			// // a component or a hook.
			// function makeListener(instance) {
			//   each(pixelsWithInferredEvents, pixel => {
			//     if (useExtendedSelector(pixel.id) && extendedButton) {
			//       foo();
			//     }
			//   });
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // This is valid because "use"-prefixed functions called in
			// // unnamed function arguments are not assumed to be hooks.
			// React.unknownFunction((foo, bar) => {
			//   if (foo) {
			//     useNotAHook(bar)
			//   }
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // This is valid because "use"-prefixed functions called in
			// // unnamed function arguments are not assumed to be hooks.
			// unknownFunction(function(foo, bar) {
			//   if (foo) {
			//     useNotAHook(bar)
			//   }
			// });
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Regression test for incorrectly flagged valid code.
			// function RegressionTest() {
			//   const foo = cond ? a : b;
			//   useState();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because exceptions abort rendering
			// function RegressionTest() {
			//   if (page == null) {
			//     throw new Error('oh no!');
			//   }
			//   useState();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because the loop doesn't change the order of hooks calls.
			// function RegressionTest() {
			//   const res = [];
			//   const additionalCond = true;
			//   for (let i = 0; i !== 10 && additionalCond; ++i ) {
			//     res.push(i);
			//   }
			//   React.useLayoutEffect(() => {});
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Is valid but hard to compute by brute-forcing
			// function MyComponent() {
			//   // 40 conditions
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}
			//   if (c) {} else {}

			//   // 10 hooks
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			//   useHook();
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because the neither the conditions before or after the hook affect the hook call
			// // Failed prior to implementing BigInt because pathsFromStartToEnd and allPathsFromStartToEnd were too big and had rounding errors
			// const useSomeHook = () => {};

			// const SomeName = () => {
			//   const filler = FILLER ?? FILLER ?? FILLER;
			//   const filler2 = FILLER ?? FILLER ?? FILLER;
			//   const filler3 = FILLER ?? FILLER ?? FILLER;
			//   const filler4 = FILLER ?? FILLER ?? FILLER;
			//   const filler5 = FILLER ?? FILLER ?? FILLER;
			//   const filler6 = FILLER ?? FILLER ?? FILLER;
			//   const filler7 = FILLER ?? FILLER ?? FILLER;
			//   const filler8 = FILLER ?? FILLER ?? FILLER;

			//   useSomeHook();

			//   if (anyConditionCanEvenBeFalse) {
			//     return null;
			//   }

			//   return (
			//     <React.Fragment>
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//       {FILLER ? FILLER : FILLER}
			//     </React.Fragment>
			//   );
			// };
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because the neither the condition nor the loop affect the hook call.
			// function App(props) {
			//   const someObject = {propA: true};
			//   for (const propName in someObject) {
			//     if (propName === true) {
			//     } else {
			//     }
			//   }
			//   const [myState, setMyState] = useState(null);
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// function App() {
			//   const text = use(Promise.resolve('A'));
			//   return <Text text={text} />
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// import * as React from 'react';
			// function App() {
			//   if (shouldShowText) {
			//     const text = use(query);
			//     const data = React.use(thing);
			//     const data2 = react.use(thing2);
			//     return <Text text={text} />
			//   }
			//   return <Text text={shouldFetchBackupText ? use(backupQuery) : "Nothing to see here"} />
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// function App() {
			//   let data = [];
			//   for (const query of queries) {
			//     const text = use(item);
			//     data.push(text);
			//   }
			//   return <Child data={data} />
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// function App() {
			//   const data = someCallback((x) => use(x));
			//   return <Child data={data} />
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// export const notAComponent = () => {
			//    return () => {
			//     useState();
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// export default () => {
			//   if (isVal) {
			//     useState(0);
			//   }
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// function notAComponent() {
			//   return new Promise.then(() => {
			//     useState();
			//   });
			// }
			// 				`,
			// 			},
			// 			{
			// 				Code: `
			// // Valid because the hook is outside of the loop
			// const Component = () => {
			//   const [state, setState] = useState(0);
			//   for (let i = 0; i < 10; i++) {
			//     console.log(i);
			//   }
			//   return <div></div>;
			// };
			// 				`,
			// 			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
// Invalid because it's dangerous and might not warn otherwise.
// This *must* be invalid.
function ComponentWithConditionalHook() {
  if (cond) {
    useConditionalHook();
  }
}
				`,

				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "conditionalHook",
						Line:      6,
					},
				},
			},
			// 			{
			// 				Code: `
			// Hook.useState();
			// Hook._useState();
			// Hook.use42();
			// Hook.useHook();
			// Hook.use_hook();
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      2,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      4,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// class C {
			//   m() {
			//     This.useHook();
			//     Super.useHook();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      4,
			// 					},
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // This is a false positive (it's valid) that unfortunately
			// // we cannot avoid. Prefer to rename it to not start with "use"
			// class Foo extends Component {
			//   render() {
			//     if (cond) {
			//       FooStore.useFeatureFlag();
			//     }
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function ComponentWithConditionalHook() {
			//   if (cond) {
			//     Namespace.useConditionalHook();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function createComponent() {
			//   return function ComponentWithConditionalHook() {
			//     if (cond) {
			//       useConditionalHook();
			//     }
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHookWithConditionalHook() {
			//   if (cond) {
			//     useConditionalHook();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function createHook() {
			//   return function useHookWithConditionalHook() {
			//     if (cond) {
			//       useConditionalHook();
			//     }
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function ComponentWithTernaryHook() {
			//   cond ? useTernaryHook() : null;
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// function ComponentWithHookInsideCallback() {
			//   useEffect(() => {
			//     useHookInsideCallback();
			//   });
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "genericHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// function createComponent() {
			//   return function ComponentWithHookInsideCallback() {
			//     useEffect(() => {
			//       useHookInsideCallback();
			//     });
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "genericHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// const ComponentWithHookInsideCallback = React.forwardRef((props, ref) => {
			//   useEffect(() => {
			//     useHookInsideCallback();
			//   });
			//   return <button {...props} ref={ref} />
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "genericHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// const ComponentWithHookInsideCallback = React.memo(props => {
			//   useEffect(() => {
			//     useHookInsideCallback();
			//   });
			//   return <button {...props} />
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "genericHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// function ComponentWithHookInsideCallback() {
			//   function handleClick() {
			//     useState();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's a common misunderstanding.
			// // We *could* make it valid but the runtime error could be confusing.
			// function createComponent() {
			//   return function ComponentWithHookInsideCallback() {
			//     function handleClick() {
			//       useState();
			//     }
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function ComponentWithHookInsideLoop() {
			//   while (cond) {
			//     useHookInsideLoop();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function ComponentWithHookInsideLoop() {
			//   do {
			//     useHookInsideLoop();
			//   } while (cond);
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function ComponentWithHookInsideLoop() {
			//   do {
			//     foo();
			//   } while (useHookInsideLoop());
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function renderItem() {
			//   useState();
			// }

			// function List(props) {
			//   return props.items.map(renderItem);
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Currently invalid because it violates the convention and removes the "taint"
			// // from a hook. We *could* make it valid to avoid some false positives but let's
			// // ensure that we don't break the "renderItem" and "normalFunctionWithConditionalHook"
			// // cases which must remain invalid.
			// function normalFunctionWithHook() {
			//   useHookInsideNormalFunction();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // These are neither functions nor hooks.
			// function _normalFunctionWithHook() {
			//   useHookInsideNormalFunction();
			// }
			// function _useNotAHook() {
			//   useHookInsideNormalFunction();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      4,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function normalFunctionWithConditionalHook() {
			//   if (cond) {
			//     useHookInsideNormalFunction();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHookInLoops() {
			//   while (a) {
			//     useHook1();
			//     if (b) return;
			//     useHook2();
			//   }
			//   while (c) {
			//     useHook3();
			//     if (d) return;
			//     useHook4();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      8,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      11,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      13,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHookInLoops() {
			//   while (a) {
			//     useHook1();
			//     if (b) continue;
			//     useHook2();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      8,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHookInLoops() {
			//   do {
			//     useHook1();
			//     if (a) return;
			//     useHook2();
			//   } while (b);

			//   do {
			//     useHook3();
			//     if (c) return;
			//     useHook4();
			//   } while (d)
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      8,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      12,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      14,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHookInLoops() {
			//   do {
			//     useHook1();
			//     if (a) continue;
			//     useHook2();
			//   } while (b);
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "loopHook",
			// 						Line:      8,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useLabeledBlock() {
			//   label: {
			//     if (a) break label;
			//     useHook();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Currently invalid.
			// // These are variations capturing the current heuristic--
			// // we only allow hooks in PascalCase or useFoo functions.
			// // We *could* make some of these valid. But before doing it,
			// // consider specific cases documented above that contain reasoning.
			// function a() { useState(); }
			// const whatever = function b() { useState(); };
			// const c = () => { useState(); };
			// let d = () => useState();
			// e = () => { useState(); };
			// ({f: () => { useState(); }});
			// ({g() { useState(); }});
			// const {j = () => { useState(); }} = {};
			// ({k = () => { useState(); }} = {});
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      7,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      8,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      9,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      10,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      11,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      12,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      13,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      14,
			// 					},
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      15,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook() {
			//   if (a) return;
			//   useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook() {
			//   if (a) return;
			//   if (b) {
			//     console.log('true');
			//   } else {
			//     console.log('false');
			//   }
			//   useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      11,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook() {
			//   if (b) {
			//     console.log('true');
			//   } else {
			//     console.log('false');
			//   }
			//   if (a) return;
			//   useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      11,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook() {
			//   a && useHook1();
			//   b && useHook2();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      5,
			// 					},
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook() {
			//   try {
			//     f();
			//     useState();
			//   } catch {}
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// function useHook({ bar }) {
			//   let foo1 = bar && useState();
			//   let foo2 = bar || useState();
			//   let foo3 = bar ?? useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      5,
			// 					},
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// const FancyButton = React.forwardRef((props, ref) => {
			//   if (props.fancy) {
			//     useCustomHook();
			//   }
			//   return <button ref={ref}>{props.children}</button>;
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// const FancyButton = forwardRef(function(props, ref) {
			//   if (props.fancy) {
			//     useCustomHook();
			//   }
			//   return <button ref={ref}>{props.children}</button>;
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous and might not warn otherwise.
			// // This *must* be invalid.
			// const MemoizedButton = memo(function(props) {
			//   if (props.fancy) {
			//     useCustomHook();
			//   }
			//   return <button>{props.children}</button>;
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "conditionalHook",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // This is invalid because "use"-prefixed functions used in named
			// // functions are assumed to be hooks.
			// React.unknownFunction(function notAComponent(foo, bar) {
			//   useProbablyAHook(bar)
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Invalid because it's dangerous.
			// // Normally, this would crash, but not if you use inline requires.
			// // This *must* be invalid.
			// // It's expected to have some false positives, but arguably
			// // they are confusing anyway due to the use*() convention
			// // already being associated with Hooks.
			// useState();
			// if (foo) {
			//   const foo = React.useCallback(() => {});
			// }
			// useCustomHook();
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      8,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      10,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      12,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// // Technically this is a false positive.
			// // We *could* make it valid (and it used to be).
			// //
			// // However, top-level Hook-like calls can be very dangerous
			// // in environments with inline requires because they can mask
			// // the runtime error by accident.
			// // So we prefer to disallow it despite the false positive.

			// const {createHistory, useBasename} = require('history-2.1.2');
			// const browserHistory = useBasename(createHistory)({
			//   basename: '/',
			// });
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      11,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// class ClassComponentWithFeatureFlag extends React.Component {
			//   render() {
			//     if (foo) {
			//       useFeatureFlag();
			//     }
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      5,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// class ClassComponentWithHook extends React.Component {
			//   render() {
			//     React.useState();
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      4,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// (class {useHook = () => { useState(); }});
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      2,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// (class {useHook() { useState(); }});
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      2,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// (class {h = () => { useState(); }});
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      2,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// (class {i() { useState(); }});
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      2,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function AsyncComponent() {
			//   useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function useAsyncHook() {
			//   useState();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function Page() {
			//   useId();
			//   React.useId();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      3,
			// 					},
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      4,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function useAsyncHook() {
			//   useId();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function notAHook() {
			//   useId();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// Hook.use();
			// Hook._use();
			// Hook.useState();
			// Hook._useState();
			// Hook.use42();
			// Hook.useHook();
			// Hook.use_hook();
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      2,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      4,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      6,
			// 					},
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      7,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// function notAComponent() {
			//   use(promise);
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "functionHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// const text = use(promise);
			// function App() {
			//   return <Text text={text} />
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "topLevelHook",
			// 						Line:      2,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// class C {
			//   m() {
			//     use(promise);
			//   }
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "classHook",
			// 						Line:      4,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// async function AsyncComponent() {
			//   use();
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "asyncComponentHook",
			// 						Line:      3,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// function App({p1, p2}) {
			//   try {
			//     use(p1);
			//   } catch (error) {
			//     console.error(error);
			//   }
			//   use(p2);
			//   return <div>App</div>;
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "tryCatchUse",
			// 						Line:      4,
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Code: `
			// function App({p1, p2}) {
			//   try {
			//     doSomething();
			//   } catch {
			//     use(p1);
			//   }
			//   use(p2);
			//   return <div>App</div>;
			// }
			// 				`,

			// 				Errors: []rule_tester.InvalidTestCaseError{
			// 					{
			// 						MessageId: "tryCatchUse",
			// 						Line:      6,
			// 					},
			// 				},
			// 			},
		},
	)
}
