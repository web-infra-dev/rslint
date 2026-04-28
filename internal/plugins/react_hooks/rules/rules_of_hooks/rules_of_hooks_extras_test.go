package rules_of_hooks

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRulesOfHooksRule_Extras runs rslint-specific edge cases that have no
// upstream analog: tsgo AST quirks, naming / container boundaries,
// path-counting edges, and additional useEffectEvent / settings shapes.

var rulesOfHooksExtrasValid = []rule_tester.ValidTestCase{
		// ---- Extra edge cases (rslint-only): paren-wrapped callees, optional chains, TS wrappers ----
		// Paren-wrapped hook: `(useHook)()` — tsgo preserves the
		// ParenthesizedExpression that ESTree flattens. Hook detection must
		// see through it.
		{Code: `
			function ComponentWithHook() {
				(useHook)();
			}
		`, Tsx: true},
		// Paren-wrapped namespace: `(Namespace).useHook()` — same shape, two
		// hops of paren peel.
		{Code: `
			function ComponentWithHook() {
				(Namespace).useHook();
			}
		`, Tsx: true},
		// Optional-chain hook in a component is allowed structurally — the
		// rule recognizes it as a hook and the caller is a component.
		{Code: `
			function ComponentWithHook(obj) {
				obj?.useFoo;
			}
		`, Tsx: true},
		// `default:` clause in switch — fall-through to a hook. Switch-case
		// bodies are treated as conditional, but a non-component caller
		// trumps the conditional check via container-error path.
		{Code: `
			function ComponentWithSwitch(x) {
				switch (x) {
					case 1:
						break;
					default:
						break;
				}
				useFoo();
			}
		`, Tsx: true},
		// Hook inside a `finally` block — finally runs on every path, so the
		// hook is unconditional. (Upstream's CFG includes finally on the
		// natural exit.)
		{Code: `
			function ComponentWithFinally() {
				try { mayThrow(); } finally { /* no hook here */ }
				useFoo();
			}
		`, Tsx: true},
		// `$FlowFixMe[react-rule-hook]` suppresses the report on the next line.
		{Code: `
			function notAComponent() {
				// $FlowFixMe[react-rule-hook]
				useState();
			}
		`, Tsx: true},

		// ---- Path-counting edge cases (where AST heuristic vs upstream CFG could diverge) ----
		// for-loop initializer: hook runs once before the loop starts.
		// Upstream CFG places the init segment outside the loop; we must
		// not trigger loopError here.
		{Code: `
			function Component() {
				for (let x = useHook(); cond; ) {
					log();
				}
			}
		`, Tsx: true},
		// for-of right-hand side: evaluated once before iteration begins.
		// Upstream's CFG keeps the RHS in the pre-loop segment.
		{Code: `
			function Component() {
				for (const x of useArr()) {
					log(x);
				}
			}
		`, Tsx: true},
		// for-in right-hand side: same shape as for-of.
		{Code: `
			function Component() {
				for (const k in useObj()) {
					log(k);
				}
			}
		`, Tsx: true},
		// try block first statement: nothing in the try block can throw
		// before the hook, so upstream's path counting concludes the hook
		// is on every path.
		{Code: `
			function Component() {
				try {
					useState();
				} catch {}
			}
		`, Tsx: true},
		// finally block first statement: finally runs on every code path,
		// so the hook is unconditional.
		{Code: `
			function Component() {
				try {
					foo();
				} finally {
					useHook();
				}
			}
		`, Tsx: true},
		// Ternary condition position: `useHook() ? a : b` evaluates the
		// hook unconditionally; the conditional branches are the
		// consequent/alternate, not the test.
		{Code: `
			function Component() {
				return useHook() ? a : b;
			}
		`, Tsx: true},
		// Logical-and LEFT operand: always evaluated.
		{Code: `
			function Component() {
				return useHook() && extra;
			}
		`, Tsx: true},
		// Logical-or LEFT operand: always evaluated.
		{Code: `
			function Component() {
				return useHook() || fallback;
			}
		`, Tsx: true},
		// Nullish-coalescing LEFT operand: always evaluated.
		{Code: `
			function Component() {
				return useHook() ?? fallback;
			}
		`, Tsx: true},
		// Nested hook call as argument: both inner and outer are evaluated
		// unconditionally on the call path.
		{Code: `
			function useFoo() {
				return useBar(useBaz());
			}
		`, Tsx: true},
		// Hook AFTER a switch where every clause falls through cleanly.
		// The merge point post-switch is on every path.
		{Code: `
			function Component(x) {
				switch (x) {
					case 1: log(); break;
					case 2: log(); break;
					default: log(); break;
				}
				useHook();
			}
		`, Tsx: true},
		// Unconditional `throw useHook()`: the hook is evaluated before the
		// throw, on every path. The function has no natural exit, but
		// upstream's path-counting reasoning still treats the hook segment
		// as reached on every reachable path.
		{Code: `
			function Component() {
				throw useHook();
			}
		`, Tsx: true},
		// Spread argument: hook evaluated as part of normal call sequence.
		{Code: `
			function Component() {
				foo(...useHooks());
			}
		`, Tsx: true},
		// TS `as` cast around hook call — not a conditional wrapper.
		{Code: `
			function Component() {
				return useHook() as any;
			}
		`, Tsx: true},
		// Comma operator: every operand evaluates left-to-right.
		{Code: `
			function Component() {
				return (a, useHook());
			}
		`, Tsx: true},
		// for-loop CONDITION position: every iteration evaluates the
		// condition, including the very first one — this DOES make the
		// hook cycled. (Sanity check: confirm we still flag it as a loop.)
		// — moved to invalid below.

		// ---- Callee-shape edge cases (rslint-only) ----
		// Element access form: `Foo['useBar']()` — upstream rejects
		// `node.computed`, we don't recognize ElementAccessExpression as a
		// hook. Not a hook call → no report regardless of caller context.
		{Code: `
			function notAComponent() {
				Foo['useBar']();
			}
		`, Tsx: true},
		// Numeric-key access: `Foo[0]()` — same path; not a hook.
		{Code: `
			function notAComponent() {
				Foo[0]();
			}
		`, Tsx: true},
		// PrivateIdentifier in property access: `Foo.#useBar()` — `#useBar`
		// is a PrivateIdentifier, not a regular Identifier; we never look up
		// hook names on it.
		{Code: `
			class C {
				#useBar() {}
				m() { this.#useBar(); }
			}
		`, Tsx: true},
		// TS `as` cast wrapping the callee: `(useFoo as any)()` — ESTree
		// (and upstream) never sees the AsExpression wrapper, so the hook
		// name is hidden behind it. We mirror that observably: the
		// AsExpression wrapper makes the callee an AsExpression, not an
		// Identifier — so we don't recognize it as a hook either. (Both
		// implementations are uniformly silent here.)
		{Code: `
			function notAComponent() {
				(useFoo as any)();
			}
		`, Tsx: true},
		// Non-null assertion on the callee: `useFoo!()`. Same as `as`-cast
		// — the assertion node hides the Identifier from the rule.
		{Code: `
			function notAComponent() {
				useFoo!();
			}
		`, Tsx: true},
		// `<T>x` type-assertion form (only legal in .ts, not .tsx — in .tsx
		// `<any>` would parse as a JSX tag). Same family as the two above:
		// the wrapper hides the Identifier from the rule.
		{Code: `
			function notAComponent() {
				(<any>useFoo)();
			}
		`, Tsx: false},
		// Optional-chain hook callee in a component: still a hook.
		{Code: `
			function Component() {
				Foo?.useBar();
			}
		`, Tsx: true},

		// ---- Hook-naming boundary cases ----
		// Digit-suffix hook names: use[0-9] are valid hooks.
		{Code: `
			function Component() {
				use1();
				use9();
			}
		`, Tsx: true},
		// All-uppercase suffix is a hook name (regex says ^use[A-Z0-9]).
		{Code: `
			function Component() {
				useFOO();
				useFooBAR();
			}
		`, Tsx: true},
		// `_use*` is NOT a hook name (regex anchors on first char).
		{Code: `
			function notAComponent() {
				_use();
				_useState();
				_useFoo();
			}
		`, Tsx: true},
		// `use_*` (snake_case after `use`) is NOT a hook.
		{Code: `
			function notAComponent() {
				use_();
				use_hook();
				use_Foo();
			}
		`, Tsx: true},
		// Single-letter PascalCase namespace counts.
		{Code: `
			function Component() {
				A.useFoo();
			}
		`, Tsx: true},
		// Underscore-prefixed namespace is NOT PascalCase per upstream
		// regex `^[A-Z]` — so `_Foo.useBar()` is not a hook.
		{Code: `
			function notAComponent() {
				_Foo.useBar();
			}
		`, Tsx: true},

		// ---- Component-name boundary ----
		// Function declared as a component but returns no JSX — naming
		// alone classifies it; rule does not inspect return value.
		{Code: `
			function Foo() {
				useState();
			}
		`, Tsx: true},
		// Lowercase factory then named-Component arrow inside.
		{Code: `
			const componentFactory = () => {
				const Inner = () => { useState(); return null; };
				return Inner;
			};
		`, Tsx: true},

		// ---- Control-flow edge nests ----
		// Inner try first stmt — upstream's path counting: hook segment
		// reached on every reachable path inside inner try.
		{Code: `
			function Component() {
				try {
					try {
						useHook();
					} catch {}
				} catch {}
			}
		`, Tsx: true},
		// Try-finally with return in try: finally still runs unconditionally.
		{Code: `
			function Component() {
				try {
					return 1;
				} finally {
					useHook();
				}
			}
		`, Tsx: true},
		// Switch with no default + hook AFTER the switch: post-switch is
		// the merge point on every path.
		{Code: `
			function Component(x) {
				switch (x) {
					case 1: log(); break;
					case 2: log(); break;
				}
				useHook();
			}
		`, Tsx: true},
		// For-of with `break` inside body, hook AFTER the loop: hook is on
		// every path leaving the loop.
		{Code: `
			function Component() {
				for (const x of arr) {
					if (a) break;
					log(x);
				}
				useHook();
			}
		`, Tsx: true},
		// `while (true) { useHook(); }` — body is cycled. Tested as INVALID
		// below. Here we add a sanity-check that an outer hook OUTSIDE the
		// loop is fine even if the loop is unreachable from below.
		{Code: `
			function Component() {
				useHook();
				while (true) { /* infinite */ break; }
			}
		`, Tsx: true},
		// Labeled `continue` inside a labeled loop: hook after is still
		// reached on every iteration's tail path.
		{Code: `
			function Component() {
				outer: for (const x of arr) {
					if (a) continue outer;
					log(x);
				}
				useHook();
			}
		`, Tsx: true},

		// ---- useEffectEvent: nested effects, React.* form, BindingElement ----
		// React.useEffect form referencing useEffectEvent.
		{Code: `
			function MyComponent() {
				const onClick = useEffectEvent(() => {});
				React.useEffect(() => { onClick(); }, []);
			}
		`, Tsx: true},
		// Same identifier name in two unrelated components, each binding a
		// useEffectEvent and using it INSIDE useEffect → both valid.
		{Code: `
			function A() {
				const onClick = useEffectEvent(() => {});
				useEffect(() => { onClick(); });
			}
			function B() {
				const onClick = useEffectEvent(() => {});
				useEffect(() => { onClick(); });
			}
		`, Tsx: true},
		// Same identifier name, but in B it's a parameter — NOT a binding.
		// The fallback (name-based) resolver could be confused; the
		// TypeChecker path resolves correctly. Test ensures B's onClick
		// doesn't get falsely reported.
		{Code: `
			function A({ theme }) {
				const onClick = useEffectEvent(() => {});
				useEffect(() => { onClick(); });
			}
			function B({ onClick }) {
				return <Child onClick={onClick} />;
			}
		`, Tsx: true},

		// ---- use() boundary ----
		// `use()` inside an arbitrary callback that is NOT a component or
		// hook — upstream's logic skips genericError for `use()`.
		{Code: `
			function App() {
				callback(() => use(p));
			}
		`, Tsx: true},
		// React.use() in component conditional position — allowed.
		{Code: `
			function App() {
				if (cond) {
					return React.use(promise);
				}
				return null;
			}
		`, Tsx: true},

		// ---- settings edge: invalid regex silently ignored ----
		{
			Code: `
				function MyComponent() {
					const onClick = useEffectEvent(() => {});
					useEffect(() => { onClick(); });
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react-hooks": map[string]interface{}{
					"additionalEffectHooks": "(unclosed",
				},
			},
		},
		// settings present but `react-hooks` key missing — should be ignored.
		{
			Code: `
				function MyComponent() {
					const onClick = useEffectEvent(() => {});
					useEffect(() => { onClick(); });
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"some-other-plugin": map[string]interface{}{"key": "value"},
			},
		},

		// ---- Multi-return hook reachability ----
		// `if (a) return; if (b) return; useFoo();` — two early returns.
		// (We keep this as INVALID below since either gate triggers
		// "after early return".)

		// ---- Multi-level namespace hook callee (NOT recognized as hook) ----
		// `Foo?.Bar.useBaz()` — `Bar.useBaz` makes the object of the outer
		// PropertyAccess a PropertyAccessExpression (not Identifier).
		// Upstream's `isHook` recursion requires obj to be an Identifier;
		// we mirror that → not a hook call.
		{Code: `
			function notAComponent() {
				Foo?.Bar.useBaz();
			}
		`, Tsx: true},
		// `A.B.useFoo()` — three-level path; same shape, same outcome.
		{Code: `
			function notAComponent() {
				A.B.useFoo();
			}
		`, Tsx: true},

		// ---- Receiver shapes that are NOT Identifier (no hook) ----
		// `this.useFoo()` — `this` is ThisKeyword, not Identifier.
		{Code: `
			function Component() {
				this.useFoo();
			}
		`, Tsx: true},
		// `super.useFoo()` — `super` is SuperKeyword (only valid inside
		// class methods, but we test the rule's view of it).
		{Code: `
			class C extends X {
				m() { super.useFoo(); }
			}
		`, Tsx: true},

		// ---- Non-CallExpression contexts (no rule trigger) ----
		// Tagged template: `useFoo\`...\`` is TaggedTemplateExpression,
		// not CallExpression — rule's listener doesn't fire.
		{Code: `
			function notAComponent() {
				useFoo` + "`hello`" + `;
			}
		`, Tsx: true},
		// `new useFoo()` — NewExpression, not CallExpression.
		{Code: `
			function notAComponent() {
				new useFoo();
			}
		`, Tsx: true},
		// Plain identifier reference, not invoked.
		{Code: `
			function notAComponent() {
				const f = useFoo;
				return f;
			}
		`, Tsx: true},

		// ---- Naming boundary: `$` and other non-[A-Z0-9] chars ----
		// `useTYPESCRIPT` — all-caps suffix, valid hook name.
		// `use$` — `$` is not in [A-Z0-9] → NOT a hook.
		{Code: `
			function notAComponent() {
				use$();
			}
		`, Tsx: true},
		// `Foo9.useBar()` — digit-suffix PascalCase namespace is valid.
		{Code: `
			function Component() {
				Foo9.useBar();
			}
		`, Tsx: true},

		// ---- Component / hook container variants ----
		// Named default export — uses the function's own name.
		{Code: `
			export default function Component() {
				useState();
			}
		`, Tsx: true},
		// Named function expression as variable initializer — function's
		// own name wins over the LHS const.
		{Code: `
			const X = function ComponentName() {
				useState();
			};
		`, Tsx: true},
		// Class field initializer with named arrow assignment — name
		// resolved from the class field name should not affect a
		// component-named outer.
		{Code: `
			function Component() {
				const useNested = () => {
					useState();
				};
				return null;
			}
		`, Tsx: true},
		// Object getter named like a hook — getter inside ObjectLiteral
		// resolves its own name.
		{Code: `
			const obj = {
				get useFoo() {
					useState();
					return 0;
				}
			};
		`, Tsx: true},
		// Generator hook — function* with hook name.
		{Code: `
			function* useGenerator() {
				useState();
			}
		`, Tsx: true},

		// ---- Nested forwardRef / memo (component recognition still applies) ----
		{Code: `
			const C = memo(forwardRef((props, ref) => {
				useHook();
				return <button {...props} ref={ref} />;
			}));
		`, Tsx: true},
		{Code: `
			const C = React.memo(React.forwardRef((props, ref) => {
				useHook();
				return <button {...props} ref={ref} />;
			}));
		`, Tsx: true},
		// forwardRef wrapped in memo with React.memo + bare forwardRef.
		{Code: `
			const C = React.memo(forwardRef(function Inner(props, ref) {
				useHook();
				return <button />;
			}));
		`, Tsx: true},

		// ---- JSX expression positions in components ----
		// JSX attribute value calling a hook — unconditional.
		{Code: `
			function Component() {
				return <Child x={useState()} />;
			}
		`, Tsx: true},
		// JSX fragment short syntax with hook expression.
		{Code: `
			function Component() {
				return <>{useState()}</>;
			}
		`, Tsx: true},
		// React.Fragment with hook expression.
		{Code: `
			function Component() {
				return <React.Fragment>{useState()}</React.Fragment>;
			}
		`, Tsx: true},

		// ---- Multi-level throw inside if (no early-return signal) ----
		// `if (a) { if (b) throw new Error(); } useFoo();` — throws don't
		// count as early-return; useFoo reachable on every non-thrown path.
		{Code: `
			function Component() {
				if (a) {
					if (b) {
						throw new Error();
					}
				}
				useFoo();
			}
		`, Tsx: true},

		// ---- continue / break out of loop, hook AFTER loop ----
		// `for (...) { if (a) break; } useFoo()` — useFoo on every path
		// leaving the loop.
		{Code: `
			function Component() {
				for (const x of arr) {
					if (a) break;
				}
				useFoo();
			}
		`, Tsx: true},
		// continue with no label, hook after loop.
		{Code: `
			function Component() {
				for (const x of arr) {
					if (a) continue;
					log(x);
				}
				useFoo();
			}
		`, Tsx: true},

		// ---- settings malformed ----
		// settings.react-hooks is a string (wrong type) → silently ignored.
		{
			Code: `
				function MyComponent() {
					const onClick = useEffectEvent(() => {});
					useEffect(() => { onClick(); });
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react-hooks": "should-be-object",
			},
		},
		// settings.react-hooks.additionalEffectHooks is a number (wrong type).
		{
			Code: `
				function MyComponent() {
					const onClick = useEffectEvent(() => {});
					useEffect(() => { onClick(); });
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react-hooks": map[string]interface{}{
					"additionalEffectHooks": 42,
				},
			},
		},

		// ---- Edge cases of "AST heuristic vs BigInt path counting" ----
		// #5 multiple `break outer` jumping to the same label, hook AFTER
		// the outer loop. useBar is on every reachable path.
		{Code: `
			function useFoo() {
				outer: for (const x of arr) {
					for (const y of inner) {
						if (a) break outer;
						if (b) break outer;
						if (c) break outer;
					}
				}
				useBar();
			}
		`, Tsx: true},
		// #9 try with return + finally with hook: finally always runs.
		{Code: `
			function Component() {
				try {
					return foo();
				} finally {
					useBar();
				}
			}
		`, Tsx: true},
		// #16 with-statement (only legal in non-strict mode .ts files).
		// `with(obj) { useFoo(); }` — body is unconditional, hook reached.
		{Code: `
			function Component() {
				with (obj) {
					useFoo();
				}
			}
		`, Tsx: false},
		// #20 optional-chain on the hook callee itself: `Foo?.useBar?.()`
		// — still recognized as a hook (object is Identifier, prop is hook).
		{Code: `
			function Component() {
				Foo?.useBar?.();
			}
		`, Tsx: true},
		// #21 chained optional access: object is PropertyAccess, not
		// Identifier → not a hook callee.
		{Code: `
			function notAComponent() {
				Foo?.Bar?.useBaz();
			}
		`, Tsx: true},
		// #22 NOTE: a "block-scoped useEffectEvent binding referenced in
		// the same block" is a contradiction in upstream's model: any
		// useEffectEvent declared inside an `if` is itself a conditional
		// hook (and inline-passed-down). The valid scenario is covered by
		// the cross-function test below.
		// Cross-function useEffectEvent: binding in Outer, referenced in
		// nested Inner component's useEffect callback.
		{Code: `
			function Outer() {
				const onClick = useEffectEvent(() => {});
				function Inner() {
					useEffect(() => { onClick(); });
				}
			}
		`, Tsx: true},
		// #29 trailing-line $FlowFixMe (same line as hook): does NOT
		// suppress (upstream requires `endLine === hookLine - 1`).
		// Already exercised structurally elsewhere; here we add the
		// "trailing on same line is NOT a suppression" expectation as
		// INVALID below. Here we lock in the BLOCK-comment leading variant.
		// #30 block comment on the line above suppresses.
		{Code: `
			function notAComponent() {
				/* $FlowFixMe[react-rule-hook] */
				useState();
			}
		`, Tsx: true},
		// Multi-line block comment ending one line above the hook also suppresses.
		{Code: `
			function notAComponent() {
				/*
				 * $FlowFixMe[react-rule-hook]
				 */
				useState();
			}
		`, Tsx: true},
}

var rulesOfHooksExtrasInvalid = []rule_tester.InvalidTestCase{
		// ---- Extra edges (rslint-only): paren-wrapped callees, optional / TS wrappers ----
		// Paren-wrapped hook callee in a non-component named function: still
		// detected as a hook call → functionError.
		{
			Code: `
				function notAComponent() {
					(useState)();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
			},
		},
		// Paren-wrapped namespace call: `(Hook).useState()` at top level.
		{
			Code: `
				(Hook).useState();
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "(Hook).useState" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},
		// Switch case body: hook is conditional.
		{
			Code: `
				function Component(x) {
					switch (x) {
						case 1:
							useFoo();
							break;
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// Hook inside catch block: conditional.
		{
			Code: `
				function Component() {
					try {
						foo();
					} catch (e) {
						useState();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// Hook nested via labeled break inside else of an if: still
		// recognized via the labeled-break sibling walk.
		{
			Code: `
				function useBlock() {
					outer: {
						if (cond) {
							break outer;
						}
						useFoo();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// `$FlowFixMe[react-rule-hook]` on a different line does NOT suppress.
		{
			Code: `
				// $FlowFixMe[react-rule-hook]

				function notAComponent() {
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
			},
		},
		// for-loop CONDITION: hook in `for (; useFoo(); ) {}` IS cycled
		// (re-evaluated each iteration) — should still report loopError.
		{
			Code: `
				function Component() {
					for (; useHook(); ) {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// for-loop INCREMENTOR: same as condition — runs each iteration.
		{
			Code: `
				function Component() {
					for (let i = 0; i < 10; useHook()) {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// Try block with prior throwable statement: still conditional (mirrors
		// upstream's existing test, kept here too as a position-aware check
		// for our `isInsideTryBlockWithPriorStmt` helper).
		{
			Code: `
				function Component() {
					try {
						doSomething();
						useHook();
					} catch {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// Hook reached only via a labeled-break inside an else branch.
		{
			Code: `
				function useFoo() {
					outer: {
						if (a) {
							log();
						} else {
							break outer;
						}
						useHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},
		// Nested labeled blocks where the break target is the outer label.
		{
			Code: `
				function useFoo() {
					outer: {
						inner: {
							if (a) break outer;
						}
						useHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- More callee-shape invalid cases ----
		// Optional-chain hook callee at top level (no enclosing function):
		// still detected as a hook, reports topLevelError.
		{
			Code: `
				Foo?.useBar();
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "Foo?.useBar" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},
		// Numeric / digit-only suffix hook in a non-component function.
		{
			Code: `
				function notAComponent() {
					use1();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use1" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
			},
		},

		// ---- Switch fall-through: hook in case body that follows another. ----
		{
			Code: `
				function Component(x) {
					switch (x) {
						case 1: log();
						case 2: useHook(); break;
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- Multi-return reachability ----
		{
			Code: `
				function useHook() {
					if (a) return;
					if (b) return;
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`},
			},
		},

		// ---- Nested if both return ----
		{
			Code: `
				function useHook() {
					if (a) {
						if (b) return;
					}
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`},
			},
		},

		// ---- Inner try block N-th statement (with prior throwable) ----
		{
			Code: `
				function Component() {
					try {
						foo();
						try {
							bar();
							useHook();
						} catch {}
					} catch {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- use() in finally → tryCatchUseError (TryStatement covers
		// finally per upstream's isInsideTryCatch) ----
		{
			Code: `
				function App() {
					try { foo(); } finally { use(p); }
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "use" cannot be called in a try/catch block.`},
			},
		},

		// ---- React.use() in try block ----
		{
			Code: `
				function App({p}) {
					try { React.use(p); } catch {}
					return <div/>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "React.use" cannot be called in a try/catch block.`},
			},
		},

		// ---- For-of loop body break + hook in same body before break ----
		{
			Code: `
				function Component() {
					for (const x of arr) {
						useHook();
						if (a) break;
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- while(true) infinite loop with hook in body ----
		{
			Code: `
				function Component() {
					while (true) {
						useHook();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- useEffectEvent: same-name across components (line 1808 upstream) ----
		{
			Code: `
				function MyComponent({theme}) {
					const onClick = useEffectEvent(() => {
						showNotification(theme)
					});
					return <Child onClick={onClick} />
				}

				function MyOtherComponent({theme}) {
					const onClick = useEffectEvent(() => {
						showNotification(theme)
					});
					return <Child onClick={() => onClick()} />
				}

				function MyLastComponent({theme}) {
					const onClick = useEffectEvent(() => {
						showNotification(theme)
					});
					useEffect(() => {
						onClick();
						onClick;
					})
					return <Child />
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
			},
		},

		// ---- useEffectEvent: multi-error one component (line 1959 upstream) ----
		{
			Code: `
				function MyComponent({ theme }) {
					const onClick = useEffectEvent(() => {
						showNotification(theme);
					});
					const onClick2 = () => { onClick() };
					const onClick3 = useCallback(() => onClick(), []);
					const onClick4 = onClick;
					return <>
						<Child onClick={onClick}></Child>
						<Child onClick={onClick2}></Child>
						<Child onClick={onClick3}></Child>
					</>;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down."},
			},
		},

		// ---- additionalEffectHooks partial-match behavior ----
		// The regex `useM` matches any name starting with `useM` because
		// upstream uses `RegExp.test`, not exact match. So `useMyEffect`
		// AND `useMaintenance` both qualify. We don't currently test the
		// negative side; here we lock in that a non-matching name still
		// triggers the report.
		{
			Code: `
				function MyComponent() {
					const onClick = useEffectEvent(() => {});
					useUnrelatedHook(() => { onClick(); });
				}
			`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react-hooks": map[string]interface{}{
					"additionalEffectHooks": "useM",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "`onClick` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component."},
			},
		},

		// ---- Class expression component — hook in render ----
		{
			Code: `
				const Foo = class extends React.Component {
					render() {
						useState();
						return null;
					}
				};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},
		// Class constructor calling hook → classError.
		{
			Code: `
				class C extends X {
					constructor() {
						super();
						useState();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},

		// ---- Async generator function: hook + async modifier → asyncError ----
		{
			Code: `
				async function* Page() {
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" cannot be called in an async function.`},
			},
		},

		// ---- useEffectEvent inline used as expression in another hook ----
		// `const x = useMemo(() => useEffectEvent(...), [])` — useEffectEvent
		// call's parent is an ArrowFunction (concise body), not a
		// VariableDeclaration / ExpressionStatement → inline-passed-down.
		// Upstream ALSO reports "called inside a callback" because
		// `useEffectEvent` matches `use[A-Z]` and so passes through
		// `isHook` — so the call gets the generic-callback report on top
		// of the inline-passed-down report.
		{
			Code: `
				function MyComponent() {
					const x = useMemo(() => useEffectEvent(() => {}), []);
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`},
				{Message: `React Hook "useEffectEvent" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},
		// useEffectEvent immediately invoked: `useEffectEvent(() => {})()`
		// — the inner CE's parent is the outer CE → not an allowed parent.
		{
			Code: `
				function MyComponent() {
					useEffectEvent(() => {})();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`},
			},
		},

		// ---- Conditional expression assigned, both branches call hooks ----
		{
			Code: `
				function Component() {
					const x = a ? useFoo() : useBar();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
				{Message: `React Hook "useBar" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- 4-deep && short-circuit with hook on rightmost ----
		{
			Code: `
				function Component() {
					return a && b && c && useFoo();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- Switch with hook + return mixed across cases ----
		{
			Code: `
				function Component(x) {
					switch (x) {
						case 1: useFoo(); return;
						default: useBar();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
				{Message: `React Hook "useBar" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- Labeled continue: hook BEFORE the continue in loop body ----
		// useFoo runs every iteration's first half (continue or no);
		// `for (const x of arr) { useFoo(); if (a) continue outer; ... }` —
		// useFoo is in body → loopError.
		{
			Code: `
				function Component() {
					outer: for (const x of arr) {
						useFoo();
						if (a) continue outer;
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- Decorator at module scope calling a hook ----
		// `@useDecorator() class C {}` — useDecorator is at the call's
		// position; no enclosing function → topLevelError.
		{
			Code: `
				@useDecorator() class C {}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useDecorator" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},

		// ---- Try-finally without catch, hook AFTER throwable in try ----
		// (Potential divergence point: upstream's path counting reasoning
		// vs our "any prior throwable in try" heuristic.)
		{
			Code: `
				function Component() {
					try {
						foo();
						useHook();
					} finally {
						bar();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// ---- Edge: BigInt-territory checks ----
		// #1 N if-else with one branch returning each time, then hook.
		// Upstream's path counting: useState segment path = 1 / total = 2^N.
		// Our sibling-walk catches the first `if (...) return;` and reports
		// early-return; both implementations report.
		{
			Code: `
				function useFoo() {
					if (a1) {} else { return; }
					if (a2) {} else { return; }
					if (a3) {} else { return; }
					if (a4) {} else { return; }
					if (a5) {} else { return; }
					useState();
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`},
			},
		},

		// #7 try / catch / finally — hook in each. The catch branch is
		// conditional; try block first stmt and finally first stmt are
		// unconditional in upstream's path counting (and ours).
		{
			Code: `
				function Component() {
					try { useFoo(); } catch { useBar(); } finally { useBaz(); }
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useBar" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// #23 Nested useEffectEvent: inner `b` is itself a hook (matches
		// `use[A-Z]`) declared inside the outer useEffectEvent's callback,
		// not the component's top level → upstream and we both report
		// "called inside a callback".
		{
			Code: `
				function Comp() {
					const a = useEffectEvent(() => {
						const b = useEffectEvent(() => {});
						return b;
					});
					useEffect(() => a());
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useEffectEvent" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},

		// #24 useEffectEvent binding outside try, but the useEffect call
		// referencing it sits in a try block AFTER a throwable statement
		// → useEffect itself is reported as conditional.
		{
			Code: `
				function Comp() {
					const onClick = useEffectEvent(() => {});
					try {
						doSomething();
						useEffect(() => { onClick(); });
					} catch {}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useEffect" is called conditionally. React Hooks must be called in the exact same order in every component render.`},
			},
		},

		// #29 trailing $FlowFixMe on the SAME line as the hook does NOT
		// suppress (upstream requires line === hookLine - 1).
		{
			Code: `
				function notAComponent() {
					useState(); // $FlowFixMe[react-rule-hook]
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useState" is called in function "notAComponent" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`},
			},
		},

		// #36 ClassStaticBlock — hook directly inside `static {}` block.
		// Behavior: ClassStaticBlock is NOT a function-like container in
		// our model → findEnclosingFunction skips up to module → fn=nil →
		// topLevelError. Upstream's behavior on parser-hermes / standard
		// ESTree may differ; we lock in our observed behavior here.
		{
			Code: `
				class C {
					static {
						useFoo();
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: `React Hook "useFoo" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`},
			},
		},
}

func TestRulesOfHooksRule_Extras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &RulesOfHooksRule,
		rulesOfHooksExtrasValid,
		rulesOfHooksExtrasInvalid,
	)
}
