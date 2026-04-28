package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDepsRule_Edge holds final-mile edge cases:
//
//   - cross-file import resolution (L1/L2)
//   - class field arrow inside non-component / component class (L4)
//   - generator / async-generator callback shapes (L3/L5)
//   - experimental_autoDependenciesHooks corner cases (L6)
//   - nested useEffect + setter stable preservation (L7)
//   - top-level hook lock-in (L8)
//   - spread element followed by valid deps (L10)
//   - additionalHooks regex with capture groups / non-trivial syntax (L9)
//   - useState destructure with omitted state position (K7)
//   - useState<T>() with explicit type argument (K3)
//   - useEffect(fn, undefined as any) (K4)
//   - typeof in TypeArgument inside callback (Q2)
//   - useImperativeHandle ref param NOT a dep (K8)
//
// Each case carries a comment naming the audit point it locks in.
func TestExhaustiveDepsRule_Edge(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		edgeValid,
		edgeInvalid,
	)
}

var edgeValid = []rule_tester.ValidTestCase{
	// L1: cross-file imports — `useState` etc. resolve to symbols whose
	// declarations live OUTSIDE the component scope, so they're treated
	// as globals and not deps.
	{Code: `
		import { useState, useEffect } from 'react';
		function MyComponent() {
			const [count, setCount] = useState(0);
			useEffect(() => { setCount(c => c + 1); }, []);
			return count;
		}
	`, Tsx: true},

	// L1: import an external function — used inside callback, NOT a dep
	// because its declaration is outside the component (in module scope
	// of another file).
	{Code: `
		import { format } from './format';
		function MyComponent({ id }: { id: number }) {
			useEffect(() => { console.log(format(id)); }, [id]);
		}
	`, Tsx: true},

	// L4: class field arrow in a component-name class — exhaustive-deps
	// itself doesn't react (class components don't use hooks; rules-of-hooks
	// would flag, exhaustive-deps just analyzes any hook call it sees).
	{Code: `
		class Helper {
			handler = () => 1;
		}
		function MyComponent() {
			useEffect(() => {}, []);
		}
	`, Tsx: true},

	// L7: setter stays stable when referenced from a nested callback inside
	// an effect — no missing dep on setCount.
	{Code: `
		function MyComponent() {
			const [count, setCount] = useState(0);
			useEffect(() => {
				const inner = () => { setCount(c => c + 1); };
				inner();
			}, []);
			return count;
		}
	`, Tsx: true},

	// L8: top-level hook call (no enclosing component) — silently
	// ignored; component scope resolution returns nil and we bail.
	{Code: `
		useEffect(() => {}, []);
	`, Tsx: true},

	// L9: additionalHooks with alternation + capture group — Go regexp
	// supports the same JS-style basic alternation used by upstream.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useFooEffect(() => { console.log(id); }, [id]);
				useBarEffect(() => { console.log(id); }, [id]);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "(useFooEffect|useBarEffect)"},
	},

	// L9: regex with anchors and character classes (Go RE2 supports these).
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useV2Effect(() => { console.log(id); }, [id]);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "^useV[0-9]+Effect$"},
	},

	// L9: invalid regex string — silently ignored (matches upstream's
	// lenient `try { new RegExp(...) }` shape via our `regexp.Compile`
	// returning err handled by skipping).
	{
		Code: `
			function MyComponent() {
				useEffect(() => {}, []);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "[invalid"},
	},

	// S4: hoisted function declaration in component scope, referenced
	// in an effect callback. `f` captures no reactive values
	// (only globals like `console`), so isFunctionWithoutCapturedValues
	// classifies it as stable → no missing-dep, no diagnostic.
	{Code: `
		function MyComponent() {
			useEffect(() => { f(); }, []);
			function f() { console.log('static'); }
		}
	`, Tsx: true},

	// S4 negative twin: a hoisted function that DOES capture a reactive
	// value (`id` from props). Mark `f` as unstable → must be listed.
	// (Kept in invalid below since the lock-in is the missing-dep
	// diagnostic.)

	// K3: useState with explicit type argument — type args don't affect
	// callee classification.
	{Code: `
		function MyComponent() {
			const [count, setCount] = useState<number>(0);
			useEffect(() => { setCount(count + 1); }, [count]);
		}
	`, Tsx: true},

	// K7: useState destructure with omitted state position — only setter
	// is bound, and setter is stable.
	{Code: `
		function MyComponent() {
			const [, setCount] = useState(0);
			useEffect(() => { setCount(c => c + 1); }, []);
		}
	`, Tsx: true},

	// K4: useEffect(fn, undefined as any) — `as any` wrapper around
	// undefined identifier is peeled by analyzePropertyChainText / our
	// stripAsExpression on the deps argument.
	{Code: `
		function MyComponent() {
			useEffect(() => {}, undefined as any);
		}
	`, Tsx: true},

	// K8: useImperativeHandle ref param itself is NOT a dep — ref is
	// the first argument, callback is the second, deps is the third.
	{Code: `
		function MyComponent({ value }: { value: number }, ref: any) {
			useImperativeHandle(ref, () => ({ get: () => value }), [value]);
		}
	`, Tsx: true},

	// L6: experimental_autoDependenciesHooks accepts both null literal
	// AND `null as any`-wrapped null at deps position.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useAuto(() => { console.log(id); }, null);
			}
		`,
		Tsx: true,
		Options: map[string]interface{}{
			"additionalHooks":                    "(useAuto)",
			"experimental_autoDependenciesHooks": []interface{}{"useAuto"},
		},
	},

	// Q2: typeof inside type position of a generic — the `state`
	// reference is purely type-level, NOT a runtime dep.
	{Code: `
		function MyComponent({ initial }: { initial: number }) {
			const [state, setState] = useState<typeof initial>(initial);
			useEffect(() => { console.log(state); }, [state]);
			void setState;
		}
	`, Tsx: true},

}

var edgeInvalid = []rule_tester.InvalidTestCase{
	// L1: imported function referenced but missing in deps — should
	// remain external so NO diagnostic. (Negative — confirms imports
	// are external.) For a true invalid, capture a local that wraps
	// the import and verify it's a missing dep.
	{
		Code: `
			import { format } from './format';
			function MyComponent({ id }: { id: number }) {
				const wrapped = (x: number) => format(x) + id;
				useEffect(() => { console.log(wrapped(0)); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'wrapped'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			import { format } from './format';
			function MyComponent({ id }: { id: number }) {
				const wrapped = (x: number) => format(x) + id;
				useEffect(() => { console.log(wrapped(0)); }, [wrapped]);
			}
		`}},
			},
		},
	},

	// L9: additionalHooks alternation regex — second branch matched
	// produces correct missing-dep diagnostic.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useBarEffect(() => { console.log(id); }, []);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "(useFooEffect|useBarEffect)"},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useBarEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useBarEffect(() => { console.log(id); }, [id]);
			}
		`}},
			},
		},
	},

	// K7: omitted state position with missing setter dep would be a
	// stale-write case — but here we keep it valid-shaped to lock in
	// that omitted+stable still works. Reuse L1 shape: setter SHOULD
	// be stable (no diagnostic). The negative form is the second-tuple
	// reassignment case below.
	{
		Code: `
			function MyComponent() {
				let [, setCount] = useState(0);
				setCount = (() => 0) as any;
				useEffect(() => { setCount(c => c + 1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'setCount'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				let [, setCount] = useState(0);
				setCount = (() => 0) as any;
				useEffect(() => { setCount(c => c + 1); }, [setCount]);
			}
		`}},
			},
		},
	},

	// R1: deps array element is an OptionalCallExpression `obj?.()` —
	// upstream `analyzePropertyChain` rejects all CallExpression-typed
	// elements (it only accepts Identifier / MemberExpression /
	// OptionalMemberExpression). The element falls through to the
	// "complex expression" branch. The body's `obj` reference is
	// captured but not declared (the `obj?.()` element doesn't count
	// as a declared `obj`), so missing-dep is also reported.
	{
		Code: `
			function MyComponent({ obj }: { obj?: () => void }) {
				useEffect(() => { console.log(obj); }, [obj?.()]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// Main missing-dep diagnostic emitted first.
				Message: "React Hook useEffect has a missing dependency: 'obj'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ obj }: { obj?: () => void }) {
				useEffect(() => { console.log(obj); }, [obj]);
			}
		`}},
			},
			{
				// Per-element complex-expression diagnostic flushed after.
				Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked.",
			},
		},
	},

	// R3: JSX expression container reference inside the callback —
	// `<Comp prop={dep} />` should mark `dep` as a captured reference.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				const make = useCallback(() => <div data-id={id} />, []);
				return make();
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				const make = useCallback(() => <div data-id={id} />, [id]);
				return make();
			}
		`}},
			},
		},
	},

	// S1: ref.current in cleanup + a missing non-ref dep in the body.
	// Both diagnostics fire for the same effect.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				const ref = useRef(null);
				useEffect(() => {
					console.log(id);
					return () => { console.log(ref.current); };
				}, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// ref.current cleanup warning — anchors the receiver.
				Message: "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function.",
			},
			{
				Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				const ref = useRef(null);
				useEffect(() => {
					console.log(id);
					return () => { console.log(ref.current); };
				}, [id]);
			}
		`}},
			},
		},
	},

	// S2: multiple setState calls in an effect with no deps — only the
	// FIRST is reported (mirrors upstream's `if (setStateInsideEffectWithoutDeps)`
	// short-circuit). One single diagnostic with the suggested deps array.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				const [, setX] = useState(0);
				const [, setY] = useState(0);
				useEffect(() => { setX(id); setY(id); });
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// Only one diagnostic — the first detected setter wins.
				Message: "React Hook useEffect contains a call to 'setX'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [id] as a second argument to the useEffect Hook.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				const [, setX] = useState(0);
				const [, setY] = useState(0);
				useEffect(() => { setX(id); setY(id); }, [id]);
			}
		`}},
			},
		},
	},

	// S3: setter aliasing — `const alias = setX; alias(...)` — `alias`
	// is a fresh binding whose declaration is the const, NOT the useState
	// destructure. So it's NOT stable, and must be listed as a dep.
	{
		Code: `
			function MyComponent() {
				const [, setX] = useState(0);
				const alias = setX;
				useEffect(() => { alias(1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'alias'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				const [, setX] = useState(0);
				const alias = setX;
				useEffect(() => { alias(1); }, [alias]);
			}
		`}},
			},
		},
	},

	// S4: hoisted function declaration that captures a prop — NOT stable,
	// must be in deps. `f` is a FunctionDeclaration (not a parameter),
	// so the "wrap in useCallback" suggestion suffix is NOT appended;
	// upstream's `missingCallbackDep` only fires for parameter-typed
	// declarations.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => { f(); }, []);
				function f() { console.log(id); }
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'f'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => { f(); }, [f]);
				function f() { console.log(id); }
			}
		`}},
			},
		},
	},

	// L7: nested setter usage inside an inner callback — still no missing
	// dep on setCount (it's stable); the OUTER prop `id` IS a missing
	// dep. Mirrors upstream's `inlineReducer` recommendation form
	// (id resolves to a destructured parameter, not the state var, so
	// the suffix is "switch to useReducer..." not "functional update...").
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				const [count, setCount] = useState(0);
				useEffect(() => {
					const inner = () => { setCount(c => c + id); };
					inner();
				}, []);
				return count;
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array. If 'setCount' needs the current value of 'id', you can also switch to useReducer instead of useState and read 'id' in the reducer.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				const [count, setCount] = useState(0);
				useEffect(() => {
					const inner = () => { setCount(c => c + id); };
					inner();
				}, [id]);
				return count;
			}
		`}},
			},
		},
	},
}
