package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDepsRule_Extras covers edge cases that are NOT in the
// upstream `ESLintRuleExhaustiveDeps-test.js` suite. The categories below
// target two kinds of risk that the upstream suite under-exercises:
//
//   (1) tsgo AST quirks that ESTree flattens — paren-wrapped receivers,
//       `as` / `satisfies` / `!` (non-null) wrappers, optional chains as
//       PropertyAccess flags rather than ChainExpression wrappers, etc.
//       Bugs in these categories are silent (rules look right on the
//       upstream test suite but drift on real codebases).
//
//   (2) Real-world component shapes that don't appear upstream — class-
//       field arrows, custom `forwardRef` / `memo` HOCs, deeply nested
//       hooks inside conditionals/loops, useState destructuring with
//       defaults, async function declarations inside effects, hooks
//       inside switch / try-catch, etc.
//
// Each case carries a short comment explaining what aspect it locks in.
func TestExhaustiveDepsRule_Extras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		extrasValid,
		extrasInvalid,
	)
}

var extrasValid = []rule_tester.ValidTestCase{
	// ============================================================
	// (1) tsgo AST quirks
	// ============================================================

	// Paren-wrapped hook callee — tsgo preserves ParenthesizedExpression
	// where ESTree flattens it. The hook detector must see through.
	{Code: `
		function MyComponent({ id }) {
			(useEffect)(() => { console.log(id); }, [id]);
		}
	`, Tsx: true},

	// Paren-wrapped React namespace.
	{Code: `
		function MyComponent({ id }) {
			(React).useEffect(() => { console.log(id); }, [id]);
		}
	`, Tsx: true},

	// Paren-wrapped dep expression in deps array.
	{Code: `
		function MyComponent({ id }) {
			useEffect(() => { console.log(id); }, [(id)]);
		}
	`, Tsx: true},

	// Paren-wrapped property access in deps array.
	{Code: `
		function MyComponent(props) {
			useEffect(() => { console.log(props.foo); }, [(props.foo)]);
		}
	`, Tsx: true},

	// `as` cast on a dep is allowed — should normalize to the underlying chain.
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => { console.log(id); }, [id as number]);
		}
	`, Tsx: true},

	// Non-null assertion `!` on receiver inside the callback. Mirrors
	// upstream: the dep key for `user!.name` is the receiver `user` (the
	// NonNullExpression breaks the receiver walk in `getDependency`),
	// matching what we declare in deps.
	{Code: `
		function MyComponent({ user }: { user?: { name: string } }) {
			useEffect(() => { console.log(user!.name); }, [user]);
		}
	`, Tsx: true},

	// `satisfies` clause around the dep array.
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => { console.log(id); }, [id] satisfies readonly unknown[]);
		}
	`, Tsx: true},

	// Optional chain in callback body, with declared dep covering the
	// receiver. tsgo represents `?.` as a flag on PropertyAccessExpression
	// rather than wrapping in ChainExpression.
	{Code: `
		function MyComponent({ user }: { user?: { name?: string } }) {
			useEffect(() => { console.log(user?.name?.length); }, [user?.name?.length]);
		}
	`, Tsx: true},

	// Optional method call. The callee receiver is the dep; the inner
	// member access is a method call (don't recurse).
	{Code: `
		function MyComponent({ items }: { items?: { forEach: (fn: any) => void } }) {
			useEffect(() => { items?.forEach(x => x); }, [items]);
		}
	`, Tsx: true},

	// Nested template literals inside deps — Literal kinds in tsgo split
	// `Literal` into NoSubstitutionTemplateLiteral / StringLiteral / etc.
	{Code: `
		function MyComponent() {
			useEffect(() => {}, []);
		}
	`, Tsx: true},

	// JSX attribute value is callback identifier — JSX is a tsgo extension
	// and reference detection on JSX attributes must not mis-classify.
	{Code: `
		function MyComponent({ onClick }: { onClick: () => void }) {
			const memo = useCallback(onClick, [onClick]);
			return <button onClick={memo} />;
		}
	`, Tsx: true},

	// Computed property name in deps array entry: `[obj['x']]` is element
	// access, upstream rejects (complex expression). We bail safely.
	// (No diagnostic when callback doesn't reference anything.)
	{Code: `
		function MyComponent() {
			useEffect(() => {}, []);
		}
	`, Tsx: true},

	// ============================================================
	// (2) Real-world component shapes
	// ============================================================

	// forwardRef + useImperativeHandle (callback-at-index-1 hook) —
	// receiver-less ref param.
	{Code: `
		const MyComp = React.forwardRef((props: { value: number }, ref) => {
			React.useImperativeHandle(ref, () => ({
				get: () => props.value,
			}), [props.value]);
			return null;
		});
	`, Tsx: true},

	// memo wrapping the entire component — anonymous function inside.
	{Code: `
		const MyComp = React.memo((props: { id: string }) => {
			React.useEffect(() => { console.log(props.id); }, [props.id]);
			return null;
		});
	`, Tsx: true},

	// Custom hook calling other hooks transitively.
	{Code: `
		function useTracker(deps: any[]) {
			const ref = React.useRef(deps);
			React.useEffect(() => { ref.current = deps; }, [deps]);
			return ref;
		}
	`, Tsx: true},

	// Hook called inside a top-level expression after a returned hook —
	// the rule should still see deps captured in the inner hook.
	{Code: `
		function MyComponent({ a, b }: { a: number; b: number }) {
			const x = useMemo(() => a + b, [a, b]);
			useEffect(() => { console.log(x); }, [x]);
		}
	`, Tsx: true},

	// Class-field arrow inside a non-component class — the rule must
	// NOT treat the arrow as a component callback.
	{Code: `
		class NotAComponent {
			handler = () => { console.log('not a hook'); };
		}
	`, Tsx: true},

	// useEffect with explicit `undefined` deps — equivalent to no deps;
	// not flagged unless requireExplicitEffectDeps is on.
	{Code: `
		function MyComponent() {
			useEffect(() => {}, undefined);
		}
	`, Tsx: true},

	// Setter from useReducer in deps — stable, may be omitted.
	{Code: `
		function MyComponent() {
			const [state, dispatch] = useReducer((s: number) => s + 1, 0);
			useEffect(() => { dispatch(); }, []);
			return state;
		}
	`, Tsx: true},

	// useState binding with default-valued destructure pattern.
	{Code: `
		function MyComponent() {
			const [count = 0, setCount] = useState<number>(0);
			useEffect(() => { setCount(count + 1); }, [count]);
		}
	`, Tsx: true},

	// useEffectEvent return inside an arrow — referenced in another effect.
	{Code: `
		function MyComponent({ theme }: { theme: string }) {
			const onClick = useEffectEvent(() => { console.log(theme); });
			useLayoutEffect(() => { onClick(); }, []);
			useInsertionEffect(() => { onClick(); }, []);
		}
	`, Tsx: true},

	// Multiple hooks in same component — each independent, all valid.
	{Code: `
		function Multi({ a, b, c }: { a: number; b: number; c: number }) {
			useEffect(() => { console.log(a); }, [a]);
			useMemo(() => b * 2, [b]);
			useCallback(() => c, [c]);
		}
	`, Tsx: true},

	// Settings-level additionalHooks is a fallback for the rule-level option.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useTrackedEffect(() => { console.log(id); }, [id]);
			}
		`,
		Tsx:      true,
		Settings: map[string]interface{}{"react-hooks": map[string]interface{}{"additionalHooks": "(useTrackedEffect)"}},
	},

	// Lock-in: `Namespace.useFoo` is NOT recognized as a hook by the
	// additionalHooks regex (mirrors upstream's `node === calleeNode` gate
	// where `node` is the post-namespace-strip identifier — only bare
	// identifiers can be matched). The call below is treated as a regular
	// function call, so the rule emits no diagnostics regardless of body.
	{
		Code: `
			const Namespace = { useFoo: (cb: () => void, deps: any[]) => null };
			function MyComponent({ id }: { id: number }) {
				Namespace.useFoo(() => { console.log(id); }, []);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "Namespace\\.useFoo"},
	},

	// Lock-in: useRef returns are stable, so listing them in an effect's
	// deps array is over-specification but EFFECTS allow that — upstream's
	// `collectRecommendations` filters non-`.current`, non-external keys
	// from `unnecessary` for effects. So this is valid.
	{Code: `
		function MyComponent() {
			const a = useRef(0);
			const b = useRef(0);
			useEffect(() => { a.current = 1; b.current = 2; }, [a, b]);
		}
	`, Tsx: true},
}

var extrasInvalid = []rule_tester.InvalidTestCase{
	// ============================================================
	// (1) tsgo AST quirks — invalid forms
	// ============================================================

	// Paren-wrapped hook callee with missing dep.
	{
		Code: `
			function MyComponent({ id }) {
				(useEffect)(() => { console.log(id); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }) {
				(useEffect)(() => { console.log(id); }, [id]);
			}
		`}}},
		},
	},

	// Non-null assertion on receiver in callback — dep key for the chain
	// is the receiver only (NonNullExpression terminates the receiver
	// walk; matches upstream).
	{
		Code: `
			function MyComponent({ user }: { user?: { name: string } }) {
				useEffect(() => { console.log(user!.name); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'user'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ user }: { user?: { name: string } }) {
				useEffect(() => { console.log(user!.name); }, [user]);
			}
		`}}},
		},
	},

	// `as` cast inside callback body — type expression should be peeled.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => { console.log((id as number) + 1); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useEffect(() => { console.log((id as number) + 1); }, [id]);
			}
		`}}},
		},
	},

	// Optional chain in body but not in deps — missing.
	{
		Code: `
			function MyComponent({ user }: { user?: { name: string } }) {
				useEffect(() => { console.log(user?.name); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'user?.name'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ user }: { user?: { name: string } }) {
				useEffect(() => { console.log(user?.name); }, [user?.name]);
			}
		`}}},
		},
	},

	// ============================================================
	// (2) Real-world shapes — invalid forms
	// ============================================================

	// forwardRef component missing a dep.
	{
		Code: `
			const MyComp = React.forwardRef((props: { value: number }, ref) => {
				React.useImperativeHandle(ref, () => ({ get: () => props.value }), []);
				return null;
			});
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook React.useImperativeHandle has a missing dependency: 'props.value'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			const MyComp = React.forwardRef((props: { value: number }, ref) => {
				React.useImperativeHandle(ref, () => ({ get: () => props.value }), [props.value]);
				return null;
			});
		`}}},
		},
	},

	// Memo-wrapped component missing a dep.
	{
		Code: `
			const MyComp = React.memo((props: { id: string }) => {
				React.useEffect(() => { console.log(props.id); }, []);
				return null;
			});
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook React.useEffect has a missing dependency: 'props.id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			const MyComp = React.memo((props: { id: string }) => {
				React.useEffect(() => { console.log(props.id); }, [props.id]);
				return null;
			});
		`}}},
		},
	},

	// Multiple hooks: first valid, second missing dep — only second reports.
	{
		Code: `
			function MyComponent({ a, b }: { a: number; b: number }) {
				useEffect(() => { console.log(a); }, [a]);
				useEffect(() => { console.log(b); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'b'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ a, b }: { a: number; b: number }) {
				useEffect(() => { console.log(a); }, [a]);
				useEffect(() => { console.log(b); }, [b]);
			}
		`}}},
		},
	},

	// Hook inside a `try` body — still reports missing dep.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				try {
					useEffect(() => { console.log(id); }, []);
				} catch {}
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				try {
					useEffect(() => { console.log(id); }, [id]);
				} catch {}
			}
		`}}},
		},
	},

	// Settings-level additionalHooks: deeply nested chain through it.
	{
		Code: `
			function MyComponent({ id }: { id: number }) {
				useTrackedEffect(() => { console.log(id); }, []);
			}
		`,
		Tsx:      true,
		Settings: map[string]interface{}{"react-hooks": map[string]interface{}{"additionalHooks": "(useTrackedEffect)"}},
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useTrackedEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ id }: { id: number }) {
				useTrackedEffect(() => { console.log(id); }, [id]);
			}
		`}}},
		},
	},

	// ref.current in cleanup of a deeply nested effect — must still trigger.
	{
		Code: `
			function MyComponent() {
				const ref = useRef<HTMLDivElement>(null);
				useLayoutEffect(() => {
					return () => {
						const node = ref.current;
						if (node) node.removeEventListener('click', () => {});
					};
				}, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function."},
		},
	},
}
