package rules_of_hooks

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRulesOfHooksRule_Upstream runs every valid + invalid case ported from
// upstream `RulesOfHooks-test.js`. Flow-syntax cases (`component` / `hook`
// keywords) carry `Skip: true` with a `// SKIP:` reason and stay in the
// suite — they will be re-enabled if rslint's parser ever supports Flow.

var rulesOfHooksUpstreamValid = []rule_tester.ValidTestCase{
		// ---- Upstream: Components and hooks may call hooks ----
		{Code: `
			function ComponentWithHook() {
				useHook();
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component Button() {
				useHook();
				return <div>Button!</div>;
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `hook` syntax.
		{Skip: true, Code: `
			hook useSampleHook() {
				useHook();
			}
		`, Tsx: true},
		{Code: `
			function createComponentWithHook() {
				return function ComponentWithHook() {
					useHook();
				};
			}
		`, Tsx: true},
		{Code: `
			function useHookWithHook() {
				useHook();
			}
		`, Tsx: true},
		{Code: `
			function createHook() {
				return function useHookWithHook() {
					useHook();
				}
			}
		`, Tsx: true},
		{Code: `
			function ComponentWithNormalFunction() {
				doSomething();
			}
		`, Tsx: true},
		{Code: `
			function normalFunctionWithNormalFunction() {
				doSomething();
			}
		`, Tsx: true},
		{Code: `
			function normalFunctionWithConditionalFunction() {
				if (cond) {
					doSomething();
				}
			}
		`, Tsx: true},
		{Code: `
			function functionThatStartsWithUseButIsntAHook() {
				if (cond) {
					userFetch();
				}
			}
		`, Tsx: true},
		{Code: `
			function useUnreachable() {
				return;
				useHook();
			}
		`, Tsx: true},

		// ---- Upstream: Various binding shapes for hooks ----
		{Code: `
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
		`, Tsx: true},
		{Code: `
			function useHook() {
				useHook1();
				useHook2();
			}
		`, Tsx: true},
		{Code: `
			function createHook() {
				return function useHook() {
					useHook1();
					useHook2();
				};
			}
		`, Tsx: true},
		{Code: `
			function useHook() {
				useState() && a;
			}
		`, Tsx: true},
		{Code: `
			function useHook() {
				return useHook1() + useHook2();
			}
		`, Tsx: true},
		{Code: `
			function useHook() {
				return useHook1(useHook2());
			}
		`, Tsx: true},

		// ---- Upstream: forwardRef / memo callbacks ----
		{Code: `
			const FancyButton = React.forwardRef((props, ref) => {
				useHook();
				return <button {...props} ref={ref} />
			});
		`, Tsx: true},
		{Code: `
			const FancyButton = React.forwardRef(function (props, ref) {
				useHook();
				return <button {...props} ref={ref} />
			});
		`, Tsx: true},
		{Code: `
			const FancyButton = forwardRef(function (props, ref) {
				useHook();
				return <button {...props} ref={ref} />
			});
		`, Tsx: true},
		{Code: `
			const MemoizedFunction = React.memo(props => {
				useHook();
				return <button {...props} />
			});
		`, Tsx: true},
		{Code: `
			const MemoizedFunction = memo(function (props) {
				useHook();
				return <button {...props} />
			});
		`, Tsx: true},

		// ---- Upstream: classes calling functions are not hooks ----
		{Code: `
			class C {
				m() {
					this.useHook();
					super.useHook();
				}
			}
		`, Tsx: true},

		// ---- Upstream: jest.useFakeTimers etc are not React hooks ----
		{Code: `
			jest.useFakeTimers();
			beforeEach(() => {
				jest.useRealTimers();
			})
		`, Tsx: true},
		{Code: `
			fooState();
			_use();
			_useState();
			use_hook();
			jest.useFakeTimer()
		`, Tsx: true},

		// ---- Upstream: "use"-prefixed callbacks in unnamed function args ----
		{Code: `
			function makeListener(instance) {
				each(pixelsWithInferredEvents, pixel => {
					if (useExtendedSelector(pixel.id) && extendedButton) {
						foo();
					}
				});
			}
		`, Tsx: true},
		{Code: `
			React.unknownFunction((foo, bar) => {
				if (foo) {
					useNotAHook(bar)
				}
			});
		`, Tsx: true},
		{Code: `
			unknownFunction(function(foo, bar) {
				if (foo) {
					useNotAHook(bar)
				}
			});
		`, Tsx: true},

		// ---- Upstream: regression cases ----
		{Code: `
			function RegressionTest() {
				const foo = cond ? a : b;
				useState();
			}
		`, Tsx: true},
		{Code: `
			function RegressionTest() {
				if (page == null) {
					throw new Error('oh no!');
				}
				useState();
			}
		`, Tsx: true},
		{Code: `
			function RegressionTest() {
				const res = [];
				const additionalCond = true;
				for (let i = 0; i !== 10 && additionalCond; ++i ) {
					res.push(i);
				}
				React.useLayoutEffect(() => {});
			}
		`, Tsx: true},
		{Code: `
			function MyComponent() {
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
		`, Tsx: true},
		{Code: `
			const useSomeHook = () => {};
			const SomeName = () => {
				const filler = FILLER ?? FILLER ?? FILLER;
				const filler2 = FILLER ?? FILLER ?? FILLER;
				useSomeHook();
				if (anyConditionCanEvenBeFalse) {
					return null;
				}
				return null;
			};
		`, Tsx: true},
		{Code: `
			function App(props) {
				const someObject = {propA: true};
				for (const propName in someObject) {
					if (propName === true) {
					} else {
					}
				}
				const [myState, setMyState] = useState(null);
			}
		`, Tsx: true},

		// ---- Upstream: React `use()` may be called conditionally / in loops ----
		{Code: `
			function App() {
				const text = use(Promise.resolve('A'));
				return <Text text={text} />
			}
		`, Tsx: true},
		{Code: `
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
		`, Tsx: true},
		{Code: `
			function App() {
				let data = [];
				for (const query of queries) {
					const text = use(item);
					data.push(text);
				}
				return <Child data={data} />
			}
		`, Tsx: true},
		{Code: `
			function App() {
				const data = someCallback((x) => use(x));
				return <Child data={data} />
			}
		`, Tsx: true},

		// ---- Upstream: TODO-style false-negatives upstream documents ----
		{Code: `
			export const notAComponent = () => {
				return () => {
					useState();
				}
			}
		`, Tsx: true},
		{Code: `
			export default () => {
				if (isVal) {
					useState(0);
				}
			}
		`, Tsx: true},
		{Code: `
			function notAComponent() {
				return new Promise.then(() => {
					useState();
				});
			}
		`, Tsx: true},

		// ---- Upstream: hook outside loop ----
		{Code: `
			const Component = () => {
				const [state, setState] = useState(0);
				for (let i = 0; i < 10; i++) {
					console.log(i);
				}
				return <div></div>;
			};
		`, Tsx: true},

		// ---- Upstream: useEffectEvent valid ----
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useMyEffect(() => {
					onClick();
				});
				useServerEffect(() => {
					onClick();
				});
			}
		`, Tsx: true, Settings: map[string]interface{}{
			"react-hooks": map[string]interface{}{
				"additionalEffectHooks": "(useMyEffect|useServerEffect)",
			},
		}},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useMyEffect(() => {
					onClick();
				});
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useEffect(() => {
					onClick();
				});
				React.useEffect(() => {
					onClick();
				});
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useEffect(() => {
					onClick();
				});
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				const onClick2 = useEffectEvent(() => {
					debounce(onClick);
					debounce(() => onClick());
					debounce(() => { onClick() });
					deboucne(() => debounce(onClick));
				});
				useEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				React.useEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				return null;
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				useEffect(() => { onClick() });
				return null;
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				useEffect(() => {
					onClick();
				});
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				useEffect(() => { onClick() });
				const onClick = useEffectEvent(() => {});
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onEvent = useEffectEvent((text) => {
					console.log(text);
				});
				useEffect(() => {
					onEvent('Hello world');
				});
				React.useEffect(() => {
					onEvent('Hello world');
				});
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onEvent = useEffectEvent((text) => {});
				useEffect(() => { onEvent('Hello world'); });
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useLayoutEffect(() => {
					onClick();
				});
				React.useLayoutEffect(() => {
					onClick();
				});
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				useLayoutEffect(() => { onClick() });
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				useInsertionEffect(() => {
					onClick();
				});
				React.useInsertionEffect(() => {
					onClick();
				});
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme) {
				const onClick = useEffectEvent(() => {});
				useInsertionEffect(() => { onClick() });
			}
		`, Tsx: true},
		{Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => {
					showNotification(theme);
				});
				const onClick2 = useEffectEvent(() => {
					debounce(onClick);
					debounce(() => onClick());
					debounce(() => { onClick() });
					deboucne(() => debounce(onClick));
				});
				useLayoutEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				React.useLayoutEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				useInsertionEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				React.useInsertionEffect(() => {
					let id = setInterval(() => onClick(), 100);
					return () => clearInterval(onClick);
				}, []);
				return null;
			}
		`, Tsx: true},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				useLayoutEffect(() => { onClick(); }, []);
				useInsertionEffect(() => { onClick(); }, []);
				return null;
			}
		`, Tsx: true},
}

var rulesOfHooksUpstreamInvalid = []rule_tester.InvalidTestCase{
		// ---- Upstream: SKIP'd flow invalid tests ----
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component Button(cond: boolean) {
				if (cond) {
					useConditionalHook();
				}
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
		}},
		// SKIP: rslint does not parse Flow `hook` syntax.
		{Skip: true, Code: `
			hook useTest(cond: boolean) {
				if (cond) {
					useConditionalHook();
				}
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
		}},

		// ---- Upstream: conditional hooks ----
		// Lock the full range (Line+Column+EndLine+EndColumn) on the
		// representative conditional case so a refactor that drifts the
		// reported node (e.g. switching from callee Identifier to whole
		// CallExpression) immediately fails. Other conditional cases
		// elsewhere in this file omit EndLine/EndColumn for brevity but
		// rely on this anchor + Message text for semantic equivalence.
		{
			Code: `
				function ComponentWithConditionalHook() {
					if (cond) {
						useConditionalHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7, EndLine: 4, EndColumn: 25},
			},
		},
		{
			Code: `
				Hook.useState();
				Hook._useState();
				Hook.use42();
				Hook.useHook();
				Hook.use_hook();
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "Hook.useState" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 5},
				{Message: `React Hook "Hook.use42" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 5},
				{Message: `React Hook "Hook.useHook" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 5, Column: 5},
			},
		},
		{
			Code: `
				class C {
					m() {
						This.useHook();
						Super.useHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Range covers the full PropertyAccessExpression callee
				// (`This.useHook`), 12 chars long.
				{Message: `React Hook "This.useHook" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7, EndLine: 4, EndColumn: 19},
				{Message: `React Hook "Super.useHook" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 5, Column: 7, EndLine: 5, EndColumn: 20},
			},
		},
		{
			Code: `
				class Foo extends Component {
					render() {
						if (cond) {
							FooStore.useFeatureFlag();
						}
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "FooStore.useFeatureFlag" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 5, Column: 8},
			},
		},
		{
			Code: `
				function ComponentWithConditionalHook() {
					if (cond) {
						Namespace.useConditionalHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "Namespace.useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function createComponent() {
					return function ComponentWithConditionalHook() {
						if (cond) {
							useConditionalHook();
						}
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 8},
			},
		},
		{
			Code: `
				function useHookWithConditionalHook() {
					if (cond) {
						useConditionalHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function createHook() {
					return function useHookWithConditionalHook() {
						if (cond) {
							useConditionalHook();
						}
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 8},
			},
		},
		{
			Code: `
				function ComponentWithTernaryHook() {
					cond ? useTernaryHook() : null;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useTernaryHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 3, Column: 13},
			},
		},
		{
			Code: `
				function ComponentWithHookInsideCallback() {
					useEffect(() => {
						useHookInsideCallback();
					});
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideCallback" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function createComponent() {
					return function ComponentWithHookInsideCallback() {
						useEffect(() => {
							useHookInsideCallback();
						});
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideCallback" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 5, Column: 8},
			},
		},
		{
			Code: `
				const ComponentWithHookInsideCallback = React.forwardRef((props, ref) => {
					useEffect(() => {
						useHookInsideCallback();
					});
					return <button {...props} ref={ref} />
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideCallback" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				const ComponentWithHookInsideCallback = React.memo(props => {
					useEffect(() => {
						useHookInsideCallback();
					});
					return <button {...props} />
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideCallback" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function ComponentWithHookInsideCallback() {
					function handleClick() {
						useState();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "handleClick" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function createComponent() {
					return function ComponentWithHookInsideCallback() {
						function handleClick() {
							useState();
						}
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "handleClick" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 5, Column: 8},
			},
		},

		// ---- Upstream: loop hooks ----
		// Anchor full range here for the loop diagnostic family.
		{
			Code: `
				function ComponentWithHookInsideLoop() {
					while (cond) {
						useHookInsideLoop();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideLoop" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7, EndLine: 4, EndColumn: 24},
			},
		},
		{
			Code: `
				function ComponentWithHookInsideLoop() {
					do {
						useHookInsideLoop();
					} while (cond);
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideLoop" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				function ComponentWithHookInsideLoop() {
					do {
						foo();
					} while (useHookInsideLoop());
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideLoop" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 15},
			},
		},
		{
			Code: `
				function renderItem() {
					useState();
				}
				function List(props) {
					return props.items.map(renderItem);
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "renderItem" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
			},
		},
		{
			Code: `
				function normalFunctionWithHook() {
					useHookInsideNormalFunction();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideNormalFunction" is called in function "normalFunctionWithHook" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
			},
		},
		{
			Code: `
				function _normalFunctionWithHook() {
					useHookInsideNormalFunction();
				}
				function _useNotAHook() {
					useHookInsideNormalFunction();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideNormalFunction" is called in function "_normalFunctionWithHook" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
				{Message: `React Hook "useHookInsideNormalFunction" is called in function "_useNotAHook" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 6, Column: 6},
			},
		},
		{
			Code: `
				function normalFunctionWithConditionalHook() {
					if (cond) {
						useHookInsideNormalFunction();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHookInsideNormalFunction" is called in function "normalFunctionWithConditionalHook" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 4, Column: 7},
			},
		},

		// ---- Upstream: useHookInLoops with various return / continue ----
		{
			Code: `
				function useHookInLoops() {
					while (a) {
						useHook1();
						if (b) return;
						useHook2();
					}
					while (c) {
						useHook3();
						if (d) return;
						useHook4();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook1" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
				{Message: `React Hook "useHook2" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 6, Column: 7},
				{Message: `React Hook "useHook3" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 9, Column: 7},
				{Message: `React Hook "useHook4" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 11, Column: 7},
			},
		},
		{
			Code: `
				function useHookInLoops() {
					while (a) {
						useHook1();
						if (b) continue;
						useHook2();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook1" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
				{Message: `React Hook "useHook2" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 6, Column: 7},
			},
		},
		{
			Code: `
				function useHookInLoops() {
					do {
						useHook1();
						if (a) return;
						useHook2();
					} while (b);

					do {
						useHook3();
						if (c) return;
						useHook4();
					} while (d)
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook1" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
				{Message: `React Hook "useHook2" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 6, Column: 7},
				{Message: `React Hook "useHook3" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 10, Column: 7},
				{Message: `React Hook "useHook4" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 12, Column: 7},
			},
		},
		{
			Code: `
				function useHookInLoops() {
					do {
						useHook1();
						if (a) continue;
						useHook2();
					} while (b);
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook1" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
				{Message: `React Hook "useHook2" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`, Line: 6, Column: 7},
			},
		},

		// ---- Upstream: labeled break ----
		{
			Code: `
				function useLabeledBlock() {
					label: {
						if (a) break label;
						useHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 7},
			},
		},

		// ---- Upstream: lowercase / non-component / non-hook function names ----
		{
			Code: `
				function a() { useState(); }
				const whatever = function b() { useState(); };
				const c = () => { useState(); };
				let d = () => useState();
				e = () => { useState(); };
				({f: () => { useState(); }});
				({g() { useState(); }});
				const {j = () => { useState(); }} = {};
				({k = () => { useState(); }} = {});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "a" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "b" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "c" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "d" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "e" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "f" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "g" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "j" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
				{Message: `React Hook "useState" is called in function "k" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
			},
		},

		// ---- Upstream: early-return variants ----
		// Anchor full range here for the early-return suffix variant.
		{
			Code: `
				function useHook() {
					if (a) return;
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`, Line: 4, Column: 6, EndLine: 4, EndColumn: 14},
			},
		},
		{
			Code: `
				function useHook() {
					if (a) return;
					if (b) {
						console.log('true');
					} else {
						console.log('false');
					}
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`, Line: 9, Column: 6},
			},
		},
		{
			Code: `
				function useHook() {
					if (b) {
						console.log('true');
					} else {
						console.log('false');
					}
					if (a) return;
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`, Line: 9, Column: 6},
			},
		},

		// ---- Upstream: short-circuit variants ----
		{
			Code: `
				function useHook() {
					a && useHook1();
					b && useHook2();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook1" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 3, Column: 11},
				{Message: `React Hook "useHook2" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 11},
			},
		},
		{
			Code: `
				function useHook() {
					try {
						f();
						useState();
					} catch {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 7},
			},
		},
		{
			Code: `
				function useHook({ bar }) {
					let foo1 = bar && useState();
					let foo2 = bar || useState();
					let foo3 = bar ?? useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 3, Column: 24},
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 24},
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 5, Column: 24},
			},
		},

		// ---- Upstream: forwardRef / memo with cond ----
		{
			Code: `
				const FancyButton = React.forwardRef((props, ref) => {
					if (props.fancy) {
						useCustomHook();
					}
					return <button ref={ref}>{props.children}</button>;
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useCustomHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				const FancyButton = forwardRef(function(props, ref) {
					if (props.fancy) {
						useCustomHook();
					}
					return <button ref={ref}>{props.children}</button>;
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useCustomHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				const MemoizedButton = memo(function(props) {
					if (props.fancy) {
						useCustomHook();
					}
					return <button>{props.children}</button>;
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useCustomHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				React.unknownFunction(function notAComponent(foo, bar) {
					useProbablyAHook(bar)
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useProbablyAHook" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
			},
		},

		// ---- Upstream: top-level hooks ----
		{
			Code: `
				useState();
				if (foo) {
					const foo = React.useCallback(() => {});
				}
				useCustomHook();
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 5},
				{Message: `React Hook "React.useCallback" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 18},
				{Message: `React Hook "useCustomHook" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 6, Column: 5},
			},
		},
		{
			Code: `
				const {createHistory, useBasename} = require('history-2.1.2');
				const browserHistory = useBasename(createHistory)({
					basename: '/',
				});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useBasename" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 3, Column: 28},
			},
		},

		// ---- Upstream: class components ----
		{
			Code: `
				class ClassComponentWithFeatureFlag extends React.Component {
					render() {
						if (foo) {
							useFeatureFlag();
						}
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFeatureFlag" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 5, Column: 8},
			},
		},
		{
			Code: `
				class ClassComponentWithHook extends React.Component {
					render() {
						React.useState();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "React.useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				(class {useHook = () => { useState(); }});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 31},
			},
		},
		{
			Code: `
				(class {useHook() { useState(); }});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 25},
			},
		},
		{
			Code: `
				(class {h = () => { useState(); }});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 25},
			},
		},
		{
			Code: `
				(class {i() { useState(); }});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 19},
			},
		},

		// ---- Upstream: async function ----
		// Anchor full range here; remaining async cases below stay
		// Line+Column-only.
		{
			Code: `
				async function AsyncComponent() {
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in an async function.`, Line: 3, Column: 6, EndLine: 3, EndColumn: 14},
			},
		},
		{
			Code: `
				async function useAsyncHook() {
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in an async function.`, Line: 3, Column: 6},
			},
		},
		{
			Code: `
				async function Page() {
					useId();
					React.useId();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useId" cannot be called in an async function.`, Line: 3, Column: 6},
				{Message: `React Hook "React.useId" cannot be called in an async function.`, Line: 4, Column: 6},
			},
		},
		{
			Code: `
				async function useAsyncHook() {
					useId();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useId" cannot be called in an async function.`, Line: 3, Column: 6},
			},
		},
		{
			Code: `
				async function notAHook() {
					useId();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useId" is called in function "notAHook" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
			},
		},

		// ---- Upstream: use() specifics ----
		{
			Code: `
				Hook.use();
				Hook._use();
				Hook.useState();
				Hook._useState();
				Hook.use42();
				Hook.useHook();
				Hook.use_hook();
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "Hook.use" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 5},
				{Message: `React Hook "Hook.useState" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 5},
				{Message: `React Hook "Hook.use42" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 6, Column: 5},
				{Message: `React Hook "Hook.useHook" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 7, Column: 5},
			},
		},
		{
			Code: `
				function notAComponent() {
					use(promise);
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`, Line: 3, Column: 6},
			},
		},
		{
			Code: `
				const text = use(promise);
				function App() {
					return <Text text={text} />
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 2, Column: 18},
			},
		},
		{
			Code: `
				class C {
					m() {
						use(promise);
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`, Line: 4, Column: 7},
			},
		},
		{
			Code: `
				async function AsyncComponent() {
					use();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called in an async function.`, Line: 3, Column: 6},
			},
		},
		// Anchor full range for the tryCatchUseError diagnostic family.
		{
			Code: `
				function App({p1, p2}) {
					try {
						use(p1);
					} catch (error) {
						console.error(error);
					}
					use(p2);
					return <div>App</div>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called in a try/catch block.`, Line: 4, Column: 7, EndLine: 4, EndColumn: 10},
			},
		},
		{
			Code: `
				function App({p1, p2}) {
					try {
						doSomething();
					} catch {
						use(p1);
					}
					use(p2);
					return <div>App</div>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called in a try/catch block.`, Line: 6, Column: 7},
			},
		},

		// ---- Upstream: useEffectEvent invalid usages ----
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					useCustomHook(() => {
						onClick();
					});
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent() {
				const onClick = useEffectEvent(() => {});
				useCustomHook(() => { onClick(); });
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
		}},
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					useWrongHook(() => {
						onClick();
					});
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react-hooks": map[string]interface{}{
					"additionalEffectHooks": "useMyEffect",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				useWrongHook(() => { onClick(); });
			}
		`, Tsx: true, Settings: map[string]interface{}{
			"react-hooks": map[string]interface{}{"additionalEffectHooks": "useMyEffect"},
		}, Errors: []rule_tester.InvalidTestCaseError{
			{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
		}},
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					return <Child onClick={onClick}></Child>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				return <Child onClick={onClick}></Child>;
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
		}},
		{
			Code: `
				function MyComponent({ theme }) {
					return <Child onClick={useEffectEvent(() => {
						showNotification(theme);
					})} />;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`, Line: 3, Column: 29},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				return <Child onClick={useEffectEvent(() => {})} />;
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`},
		}},
		{
			Code: `
				const MyComponent = ({ theme }) => {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					return <Child onClick={onClick}></Child>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
			},
		},
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					let foo = onClick;
					return <Bar onClick={foo} />
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				let foo = onClick;
				return <Bar onClick={foo} />
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
		}},
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(them);
					});
					useEffect(() => {
						setTimeout(onClick, 100);
					});
					return <Child onClick={onClick} />
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
			},
		},
		// SKIP: rslint does not parse Flow `component` syntax.
		{Skip: true, Code: `
			component MyComponent(theme: any) {
				const onClick = useEffectEvent(() => {});
				useEffect(() => { setTimeout(onClick, 100); });
				return <Child onClick={onClick} />
			}
		`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
		}},
}

func TestRulesOfHooksRule_Upstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &RulesOfHooksRule,
		rulesOfHooksUpstreamValid,
		rulesOfHooksUpstreamInvalid,
	)
}
