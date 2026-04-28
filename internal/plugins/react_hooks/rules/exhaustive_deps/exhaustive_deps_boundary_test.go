package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDepsRule_Boundary covers boundary cases that the
// upstream suite under-exercises: deeply nested effects/callbacks,
// IIFE async wrappers inside synchronous effects, hooks inside switch /
// try-catch / for-of bodies, multi-component files, big-int / regex
// literal deps, useEffectEvent diagnostic singularity, and other
// real-world shapes that have caused regressions in similar ports.
func TestExhaustiveDepsRule_Boundary(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		boundaryValid,
		boundaryInvalid,
	)
}

var boundaryValid = []rule_tester.ValidTestCase{
	// IIFE async inside sync effect — the outer effect is sync, so the
	// async-effect diagnostic must NOT fire. The inner async function's
	// captures still count as effect deps.
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => {
				(async () => {
					await fetch('/api/' + id);
				})();
			}, [id]);
		}
	`, Tsx: true},

	// async function declaration inside sync effect — same rationale.
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => {
				async function load() {
					await fetch('/api/' + id);
				}
				load();
			}, [id]);
		}
	`, Tsx: true},

	// setTimeout with async callback — the async is the timer cb, not
	// the effect itself.
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => {
				const t = setTimeout(async () => {
					await fetch('/api/' + id);
				}, 1000);
				return () => clearTimeout(t);
			}, [id]);
		}
	`, Tsx: true},

	// Multiple components in one file — each should be analyzed
	// independently, with no cross-contamination of state symbols.
	{Code: `
		function A({ id }: { id: number }) {
			useEffect(() => { console.log(id); }, [id]);
		}
		function B({ name }: { name: string }) {
			useEffect(() => { console.log(name); }, [name]);
		}
	`, Tsx: true},

	// Hook inside a for-of loop — rules-of-hooks would flag this, but
	// exhaustive-deps still analyzes it correctly (and the deps are OK).
	{Code: `
		function MyComponent({ items }: { items: number[] }) {
			for (const item of items) {
				useEffect(() => { console.log(item); }, [item]);
			}
		}
	`, Tsx: true},

	// useImperativeHandle: ref param itself should NOT be considered
	// a dep (callback is at index 1).
	{Code: `
		function MyComponent({ value }: { value: number }, ref: any) {
			useImperativeHandle(ref, () => ({ get: () => value }), [value]);
		}
	`, Tsx: true},
}

var boundaryInvalid = []rule_tester.InvalidTestCase{
	// IIFE async inside sync effect — captured `id` must still be a dep.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => {
					(async () => { await fetch('/api/' + id); })();
				}, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => {
					(async () => { await fetch('/api/' + id); })();
				}, [id]);
			}
		`}},
			},
		},
	},

	// useEffectEvent in deps array — verifies the dedicated diagnostic
	// fires AND no spurious "unnecessary"/"missing" diagnostic fires
	// alongside it (single diagnostic about the use-effect-event
	// inclusion only).
	{
		Code: `
			function MyComponent({ theme }: { theme: string }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, [onClick]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ theme }: { theme: string }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, []);
			}
		`}},
			},
		},
	},

	// BigInt literal in deps.
	{
		Code: `
			function MyComponent() {
				useEffect(() => {}, [123n]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The 123n literal is not a valid dependency because it never changes. You can safely remove it."},
		},
	},

	// Regex literal in deps.
	{
		Code: `
			function MyComponent() {
				useEffect(() => {}, [/foo/]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The /foo/ literal is not a valid dependency because it never changes. You can safely remove it."},
		},
	},

	// Two components in one file, second has missing dep — must NOT
	// confuse with first component's locals.
	{
		Code: `
			function A({ id }: { id: number }) {
				useEffect(() => { console.log(id); }, [id]);
			}
			function B({ name }: { name: string }) {
				useEffect(() => { console.log(name); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'name'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function A({ id }: { id: number }) {
				useEffect(() => { console.log(id); }, [id]);
			}
			function B({ name }: { name: string }) {
				useEffect(() => { console.log(name); }, [name]);
			}
		`}},
			},
		},
	},

	// Empty additionalHooks (rule-level falsy) falls back to settings.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useTracked(() => { console.log(id); }, []);
			}
		`,
		Tsx:      true,
		Options:  map[string]interface{}{"additionalHooks": ""},
		Settings: map[string]interface{}{"react-hooks": map[string]interface{}{"additionalHooks": "(useTracked)"}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useTracked has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useTracked(() => { console.log(id); }, [id]);
			}
		`}},
			},
		},
	},

	// useState updater message uses the right initial character for a
	// non-ASCII state name. Exercises the rune-iteration fix in
	// setStateRecommendation.
	{
		Code: `
			function MyComponent() {
				const [状态, set状态] = useState(0);
				useEffect(() => { set状态(状态 + 1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// Initial-char of the missing dep "状态" must be its first
				// rune ("状"), not a sliced byte that would corrupt the UTF-8.
				Message: "React Hook useEffect has a missing dependency: '状态'. Either include it or remove the dependency array. You can also do a functional update 'set状态(状 => ...)' if you only need '状态' in the 'set状态' call.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				const [状态, set状态] = useState(0);
				useEffect(() => { set状态(状态 + 1); }, [状态]);
			}
		`}},
			},
		},
	},

	// useReducer dispatch is registered in setStateCallSites so that
	// `setState-without-deps` detection fires for `dispatch()` inside
	// an effect with no deps array. Lock-in for F1.
	{
		Code: `
			function MyComponent() {
				const [state, dispatch] = useReducer((s: number) => s + 1, 0);
				useEffect(() => { dispatch(); });
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect contains a call to 'dispatch'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [] as a second argument to the useEffect Hook.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				const [state, dispatch] = useReducer((s: number) => s + 1, 0);
				useEffect(() => { dispatch(); }, []);
			}
		`}},
			},
		},
	},

	// Compound assignment to setter (F7 — `setX += 1`) — flags it as
	// extra write, so setter is no longer stable, missing dep emitted.
	{
		Code: `
			function MyComponent() {
				let [count, setCount] = useState(0);
				setCount += 1 as any;
				useEffect(() => { setCount(c => c + 1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'setCount'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				let [count, setCount] = useState(0);
				setCount += 1 as any;
				useEffect(() => { setCount(c => c + 1); }, [setCount]);
			}
		`}},
			},
		},
	},

	// Postfix increment (F7 — `setX++`) — same as above.
	{
		Code: `
			function MyComponent() {
				let [count, setCount] = useState(0);
				(setCount as any)++;
				useEffect(() => { setCount(c => c + 1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'setCount'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent() {
				let [count, setCount] = useState(0);
				(setCount as any)++;
				useEffect(() => { setCount(c => c + 1); }, [setCount]);
			}
		`}},
			},
		},
	},

	// Function reassignment (F8) — handler is reassigned, so its
	// `isFunctionWithoutCapturedValues` stability is invalidated and
	// it must be listed as a dep.
	{
		Code: `
			function MyComponent({ flag }: { flag: boolean }) {
				let handler = () => 1;
				if (flag) handler = () => 2;
				useEffect(() => { handler(); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'handler'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ flag }: { flag: boolean }) {
				let handler = () => 1;
				if (flag) handler = () => 2;
				useEffect(() => { handler(); }, [handler]);
			}
		`}},
			},
		},
	},

	// ElementAccess assignment LHS (F3) — `obj['x'] = ...` should yield
	// dep key `obj` (the receiver), matching `.x = ...`.
	{
		Code: `
			function MyComponent({ obj }: { obj: Record<string, number> }) {
				useEffect(() => { obj['x'] = 1; }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'obj'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ obj }: { obj: Record<string, number> }) {
				useEffect(() => { obj['x'] = 1; }, [obj]);
			}
		`}},
			},
		},
	},

	// useEffectEvent rejection + enableDangerousAutofix (F6) — promote
	// suggestion to top-level autofix. Output field captures the fixed
	// code; suggestion array is also kept.
	{
		Code: `
			function MyComponent({ theme }: { theme: string }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, [onClick]);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"enableDangerousAutofixThisMayCauseInfiniteLoops": true},
		Output: []string{`
			function MyComponent({ theme }: { theme: string }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, []);
			}
		`},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ theme }: { theme: string }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, []);
			}
		`}},
			},
		},
	},

	// Nested useEffect inside useEffect callback — mirrors upstream's
	// `gatherDependenciesRecursively` walk, which descends into every
	// child scope of the callback (including nested function bodies).
	// The outer hook collects BOTH `inner` (captured transitively via
	// the nested callback) AND `outer` as missing deps, producing one
	// combined "missing dependencies: 'inner' and 'outer'" report.
	// The inner hook is analyzed independently in its own listener pass.
	{
		Code: `
			function MyComponent({ outer, inner }: { outer: number; inner: number }) {
				useEffect(() => {
					console.log(outer);
					useEffect(() => { console.log(inner); }, []);
				}, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// Outer effect's combined missing deps. The visitor fires
				// the outer CallExpression listener before descending
				// into the callback body, so this report comes first.
				Message: "React Hook useEffect has missing dependencies: 'inner' and 'outer'. Either include them or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ outer, inner }: { outer: number; inner: number }) {
				useEffect(() => {
					console.log(outer);
					useEffect(() => { console.log(inner); }, []);
				}, [inner, outer]);
			}
		`}},
			},
			{
				// Inner effect's missing dep, fired during the descent.
				Message: "React Hook useEffect has a missing dependency: 'inner'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ outer, inner }: { outer: number; inner: number }) {
				useEffect(() => {
					console.log(outer);
					useEffect(() => { console.log(inner); }, [inner]);
				}, []);
			}
		`}},
			},
		},
	},
}
