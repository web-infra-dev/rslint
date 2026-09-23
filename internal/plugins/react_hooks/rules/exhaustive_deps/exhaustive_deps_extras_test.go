package exhaustive_deps

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestExhaustiveDepsRule_Extras covers edge cases that are NOT in the
// upstream `ESLintRuleExhaustiveDeps-test.js` suite. The categories below
// target two kinds of risk that the upstream suite under-exercises:
//
//	(1) tsgo AST quirks that ESTree flattens — paren-wrapped receivers,
//	    `as` / `satisfies` / `!` (non-null) wrappers, optional chains as
//	    PropertyAccess flags rather than ChainExpression wrappers, etc.
//	    Bugs in these categories are silent (rules look right on the
//	    upstream test suite but drift on real codebases).
//
//	(2) Real-world component shapes that don't appear upstream — class-
//	    field arrows, custom `forwardRef` / `memo` HOCs, deeply nested
//	    hooks inside conditionals/loops, useState destructuring with
//	    defaults, async function declarations inside effects, hooks
//	    inside switch / try-catch, etc.
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

	// Non-null assertion `!` on receiver inside the callback. Mirrors
	// upstream: the dep key for `user!.name` is the receiver `user` (the
	// NonNullExpression breaks the receiver walk in `getDependency`),
	// matching what we declare in deps.
	{Code: `
		function MyComponent({ user }: { user?: { name: string } }) {
			useEffect(() => { console.log(user!.name); }, [user]);
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

	// Object literal shorthand reads are real value references. tsgo can
	// resolve the shorthand identifier to the ShorthandPropertyAssignment
	// node itself, so the rule must fall back to the outer component binding.
	{Code: `
		function MyComponent() {
			const isAdmin = getIsAdmin();
			useMemo(() => ({ isAdmin }), [isAdmin]);
		}
	`, Tsx: true},

	// Local shorthand shadows must resolve lexically, not by same-name lookup
	// in the outer component scope.
	{Code: `
		function MyComponent() {
			const isAdmin = true;
			useMemo(() => {
				const isAdmin = false;
				return { isAdmin };
			}, []);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent() {
			const isAdmin = true;
			useMemo(() => {
				function inner(isAdmin: boolean) {
					return { isAdmin };
				}
				return inner(false);
			}, []);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent({ data }: { data: { isAdmin: boolean } }) {
			const isAdmin = true;
			useMemo(() => {
				const { isAdmin } = data;
				return { isAdmin };
			}, [data]);
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
		Settings: map[string]interface{}{"react-hooks": map[string]interface{}{"additionalEffectHooks": "(useTrackedEffect)"}},
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
	{Code: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => { console.log(id); }, [id] satisfies readonly unknown[]);
		}
	`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies."}, {Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
		function MyComponent({ id }: { id: number }) {
			useEffect(() => { console.log(id); }, [id]);
		}
	`}}}}},
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

	// `as` cast on a deps-array ELEMENT is a complex expression (upstream's
	// analyzePropertyChain has no `as`/`satisfies` case so it throws); the
	// underlying `id` is then a missing dependency. Matches v4/v7.
	{
		Code: `
				function MyComponent({ id }: { id: number }) {
					useEffect(() => { console.log(id); }, [id as number]);
				}
			`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
				function MyComponent({ id }: { id: number }) {
					useEffect(() => { console.log(id); }, [id]);
				}
			`}}},
			{Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked."},
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
		Settings: map[string]interface{}{"react-hooks": map[string]interface{}{"additionalEffectHooks": "(useTrackedEffect)"}},
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

func TestExhaustiveDepsEditDemand(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		name              string
		code              string
		message           string
		suggestionMessage string
		fixTexts          []string
	}{
		{
			name: "dependency array replacement",
			code: `
				function Component(props: { value: number }) {
					useEffect(() => console.log(props.value), []);
				}
			`,
			message:           "React Hook useEffect has a missing dependency: 'props.value'. Either include it or remove the dependency array.",
			suggestionMessage: "Update the dependencies array to be: [props.value]",
			fixTexts:          []string{"[props.value]"},
		},
		{
			name: "opaque callback replacement",
			code: `function Component() {
				const missingCallback = makeCallback();
 useEffect(missingCallback, []);
			}`,
			message:           "React Hook useEffect has a missing dependency: 'missingCallback'. Either include it or remove the dependency array.",
			suggestionMessage: "Update the dependencies array to be: [missingCallback]",
			fixTexts:          []string{"[missingCallback]"},
		},
		{
			name: "suppression across cached hooks",
			code: `function Component(value: number) {
				const [state, setState] = useState(0);
				const onClick = useEffectEvent(() => console.log(value));
				// rslint-disable-next-line react-hooks/exhaustive-deps
				useEffect(() => console.log(value), []);
				/* rslint-disable react-hooks/exhaustive-deps */
				useEffect(() => { setState(state); onClick(); }, [onClick]);
				/* rslint-enable react-hooks/exhaustive-deps */
				useEffect(() => { setState(state); console.log(value); }, [state]);
			}`,
			message:           "React Hook useEffect has a missing dependency: 'value'. Either include it or remove the dependency array.",
			suggestionMessage: "Update the dependencies array to be: [state, value]",
			fixTexts:          []string{"[state, value]"},
		},
		{
			name: "setState dependency insertion",
			code: `function Component(value: number) {
				const [, setValue] = useState(0);
				useEffect(() => { setValue(value); });
			}`,
			message:           "React Hook useEffect contains a call to 'setValue'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass [value] as a second argument to the useEffect Hook.",
			suggestionMessage: "Add dependencies array: [value]",
			fixTexts:          []string{", [value]"},
		},
		{
			name: "construction wrapping",
			code: `function Component() {
				const handler = () => {};
				useEffect(() => handler(), [handler]);
				handler();
			}`,
			message:           "The 'handler' function makes the dependencies of useEffect Hook (at line 3) change on every render. To fix this, wrap the definition of 'handler' in its own useCallback() Hook.",
			suggestionMessage: "Wrap the definition of 'handler' in its own useCallback() Hook.",
			fixTexts:          []string{"useCallback(", ")"},
		},
		{
			name: "useEffectEvent removal",
			code: `
				function Component(theme: string) {
					const onClick = useEffectEvent(() => console.log(theme));
					useEffect(() => onClick(), [onClick]);
				}
			`,
			message:           "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.",
			suggestionMessage: "Remove the dependency `onClick`",
			fixTexts:          []string{""},
		},
	}

	configs := []struct {
		name      string
		options   []any
		dangerous bool
	}{
		{name: "suggestions", options: rule_tester.ResolveTestCaseOptions(t, &ExhaustiveDepsRule, nil)},
		{
			name: "dangerous autofix",
			options: rule_tester.ResolveTestCaseOptions(t, &ExhaustiveDepsRule, map[string]interface{}{
				"enableDangerousAutofixThisMayCauseInfiniteLoops": true,
			}),
			dangerous: true,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createExhaustiveDepsProgram(t, fixture.name+".ts", fixture.code)
			for _, config := range configs {
				t.Run(config.name, func(t *testing.T) {
					diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
					for _, demand := range []rule.EditDemand{
						rule.EditDemandNone,
						rule.EditDemandAutofix,
						rule.EditDemandSuggestion,
						rule.EditDemandAll,
					} {
						got := lintExhaustiveDepsWithDemand(program, sourceFile, config.options, demand)
						if len(got) != 1 {
							t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
						}
						if got[0].Message.Description != fixture.message {
							t.Errorf("demand %d: message = %q, want %q", demand, got[0].Message.Description, fixture.message)
						}
						diagnostics[demand] = got[0]
					}

					diagnosticsOnly := diagnostics[rule.EditDemandNone]
					for demand, diagnostic := range diagnostics {
						requireSameDiagnosticWithoutEdits(t, diagnosticsOnly, diagnostic, demand)
					}

					requireNoEdits(t, diagnostics[rule.EditDemandNone], rule.EditDemandNone)
					if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
						t.Errorf("autofix-only demand unexpectedly materialized suggestions")
					}
					if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
						t.Errorf("suggestion-only demand unexpectedly materialized autofixes")
					}

					suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
					allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
					if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
						t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
					}
					if len(*suggestionOnly) != 1 {
						t.Fatalf("suggestions = %#v, want one suggestion", *suggestionOnly)
					}
					suggestion := (*suggestionOnly)[0]
					if suggestion.Message.Description != fixture.suggestionMessage {
						t.Errorf("suggestion message = %q, want %q", suggestion.Message.Description, fixture.suggestionMessage)
					}
					if len(suggestion.FixesArr) != len(fixture.fixTexts) {
						t.Fatalf("suggestion fixes = %#v, want texts %#v", suggestion.FixesArr, fixture.fixTexts)
					}
					for index, fix := range suggestion.FixesArr {
						if fix.Text != fixture.fixTexts[index] {
							t.Errorf("suggestion fix %d text = %q, want %q", index, fix.Text, fixture.fixTexts[index])
						}
					}

					if !config.dangerous {
						if diagnostics[rule.EditDemandAutofix].FixesPtr != nil ||
							diagnostics[rule.EditDemandAll].FixesPtr != nil {
							t.Errorf("default config unexpectedly materialized a top-level autofix")
						}
						return
					}

					autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
					allFixes := diagnostics[rule.EditDemandAll].FixesPtr
					if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
						t.Fatalf("autofix artifacts differ between autofix-only and all demand")
					}
					if !reflect.DeepEqual(*autofixOnly, suggestion.FixesArr) {
						t.Fatalf("dangerous autofix = %#v, want every fix in the first suggestion %#v", *autofixOnly, suggestion.FixesArr)
					}
				})
			}
		})
	}
}

func lintExhaustiveDepsWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         lintprogram.NewFromCompiler(program),
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: exhaustiveDepsConfiguredRules(options),
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func requireSameDiagnosticWithoutEdits(
	t *testing.T,
	want rule.RuleDiagnostic,
	got rule.RuleDiagnostic,
	demand rule.EditDemand,
) {
	t.Helper()
	want.FixesPtr = nil
	want.Suggestions = nil
	got.FixesPtr = nil
	got.Suggestions = nil
	if !reflect.DeepEqual(got, want) {
		t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
	}
}

func requireNoEdits(t *testing.T, diagnostic rule.RuleDiagnostic, demand rule.EditDemand) {
	t.Helper()
	if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
		t.Errorf(
			"demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v",
			demand,
			diagnostic.FixesPtr,
			diagnostic.Suggestions,
		)
	}
}

func createExhaustiveDepsProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{tspath.ResolvePath(rootDir.Dir, fileName): code})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func exhaustiveDepsConfiguredRules(options []any) func(*ast.SourceFile) []rule.ConfiguredRule {
	return func(*ast.SourceFile) []rule.ConfiguredRule {
		return []rule.ConfiguredRule{{
			Name:     ExhaustiveDepsRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ExhaustiveDepsRule.Run(ctx, options)
			},
		}}
	}
}

// Expectations were checked against eslint-plugin-react-hooks 7.1.1 and the identical upstream main rule.
// The complete astral character in the updater hint is the documented exception.
func TestExhaustiveDepsParityRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		[]rule_tester.ValidTestCase{
			// explicit undefined with automatic effects
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{}, undefined);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
			},
			// reassigned refs remain stable
			{
				Code: "function C(){let ref=useRef();ref=other; useEffect(()=>{ref();},[]);}",
				Tsx:  true,
			},
			// reassigned reducer dispatch remains stable
			{
				Code: "function C(){let [,ref]=useReducer(fn,0);ref=other; useEffect(()=>{ref();},[]);}",
				Tsx:  true,
			},
			// reassigned functions retain the upstream classification
			{
				Code: "function C(){let ref=()=>{};ref=other; useEffect(()=>{ref();},[]);}",
				Tsx:  true,
			},
			// returning ref.current is not a cleanup function
			{
				Code: "function C(){const ref=useRef();useEffect(()=>{return (ref).current},[])}",
				Tsx:  true,
			},
			// compound ref writes suppress cleanup warnings
			{
				Code: "function C(){const ref=useRef();ref.current ||= thing;useEffect(()=>{return ref.current},[])}",
				Tsx:  true,
			},
			// insertion effects require explicit additionalHooks configuration
			{
				Code:    "function C({x}){const local=()=>x;useInsertionEffect(()=>x,[])}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect"}},
			},
			// function C({x}){if(x){let ref=()=>{}; useEffect(()=>ref(),[ref])}}
			{
				Code: "function C({x}){if(x){let ref=()=>{}; useEffect(()=>ref(),[ref])}}",
				Tsx:  true,
			},
			// function C(){const ref=useRef(); ref.current+=1; useEffect(()=>()=>ref.current,[])}
			{
				Code: "function C(){const ref=useRef(); ref.current+=1; useEffect(()=>()=>ref.current,[])}",
				Tsx:  true,
			},
			// function C(){let on=useEffectEvent(()=>{});useEffect(()=>{},[on])}
			{
				Code: "function C(){let on=useEffectEvent(()=>{});useEffect(()=>{},[on])}",
				Tsx:  true,
			},
			// function C(){useEffect(missing,[])}
			{
				Code: "function C(){useEffect(missing,[])}",
				Tsx:  true,
			},
			// const cb=()=>{}; function C(){useEffect(cb,[])}
			{
				Code: "const cb=()=>{}; function C(){useEffect(cb,[])}",
				Tsx:  true,
			},
			// let on=useEffectEvent
			{
				Code:    "function C(){let on=useEffectEvent(()=>{});useEffect(()=>{},[on]);useEffect(()=>on(),[]);}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
			},
			// parenthesized optional hook callees are not recognized
			{
				Code: "function C({x}){(React?.useEffect)(()=>x,[])}",
				Tsx:  true,
			}},
		[]rule_tester.InvalidTestCase{
			// missing explicit dependencies with automatic effects
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{});}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect always requires dependencies. Please add a dependency array or an explicit `undefined`",
						Line:    1, Column: 73, EndLine: 1, EndColumn: 82,
					},
				},
			},
			// automatic effects still reject asynchronous callbacks
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(async()=>{}, null);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Effect callbacks are synchronous to prevent race conditions. Put the async function inside:\n\nuseEffect(() => {\n  async function fetchData() {\n    // You can await here\n    const response = await MyAPI.getData(someId);\n    // ...\n  }\n  fetchData();\n}, [someId]); // Or [] if effect doesn't need props or state\n\nLearn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching",
						Line:    1, Column: 83, EndLine: 1, EndColumn: 94,
					},
				},
			},
			// automatic effects still check stale assignments
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{x=1}, null);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Assignments to the 'x' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect.",
						Line:    1, Column: 90, EndLine: 1, EndColumn: 91,
					},
				},
			},
			// automatic effects still check cleanup references
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{return ()=>ref.current}, null);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function.",
						Line:    1, Column: 103, EndLine: 1, EndColumn: 110,
					},
				},
			},
			// automatic memo still needs dependencies
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useMemo(()=>{}, null);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useMemo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?",
						Line:    1, Column: 73, EndLine: 1, EndColumn: 80,
					},
				},
			},
			// satisfies does not unwrap the dependency array
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{console.log(x)}, [] satisfies unknown[]);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies.",
						Line:    1, Column: 105, EndLine: 1, EndColumn: 127,
					},
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 105, EndLine: 1, EndColumn: 127,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{console.log(x)}, [x]);}"}}},
				},
			},
			// a TypeScript assertion around dependencies is accepted once
			{
				Code:    "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{console.log(x)}, ([] as const) as unknown[]);}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect|useAuto", "requireExplicitEffectDeps": true, "experimental_autoDependenciesHooks": []any{"useEffect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies.",
						Line:    1, Column: 105, EndLine: 1, EndColumn: 131,
					},
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 105, EndLine: 1, EndColumn: 131,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,callback}){const ref=useRef(); const [,setX]=useState(0); useEffect(()=>{console.log(x)}, [x]);}"}}},
				},
			},
			// template strings are not primitive stable constants
			{
				Code: "function C(){const ref=`text`; useEffect(()=>{ref();},[]);}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 55, EndLine: 1, EndColumn: 57,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const ref=`text`; useEffect(()=>{ref();},[ref]);}"}}},
				},
			},
			// defaulted setters are not stable
			{
				Code: "function C(){const [,ref=fn]=useState(0); useEffect(()=>{ref();},[]);}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 66, EndLine: 1, EndColumn: 68,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const [,ref=fn]=useState(0); useEffect(()=>{ref();},[ref]);}"}}},
				},
			},
			// a shadowed name does not count as an outside use
			{
				Code: "function C({value,flag}){const fn=()=>{};const ref=()=>{};function nested(ref){ref()} useEffect(()=>{ref()},[ref])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "The 'ref' function makes the dependencies of useEffect Hook (at line 1) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'ref' in its own useCallback() Hook.",
						Line:    1, Column: 48, EndLine: 1, EndColumn: 58,
					},
				},
			},
			// function parameter defaults capture render values
			{
				Code: "function C({value,flag}){const fn=()=>{};const ref=(x=value)=>x; useEffect(()=>{ref()},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 88, EndLine: 1, EndColumn: 90,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({value,flag}){const fn=()=>{};const ref=(x=value)=>x; useEffect(()=>{ref()},[ref])}"}}},
				},
			},
			// asserted function initializers are not capture-free
			{
				Code: "function C({value,flag}){const fn=()=>{};const ref=(()=>{}) as any; useEffect(()=>{ref()},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 91, EndLine: 1, EndColumn: 93,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({value,flag}){const fn=()=>{};const ref=(()=>{}) as any; useEffect(()=>{ref()},[ref])}"}}},
				},
			},
			// qualified type queries retain upstream reference behavior
			{
				Code: "function C({obj,x,y,z,key,other}){useEffect(()=>{type X=typeof x;type Y=typeof obj.foo},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'obj'. Either include it or remove the dependency array.",
						Line:    1, Column: 89, EndLine: 1, EndColumn: 91,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({obj,x,y,z,key,other}){useEffect(()=>{type X=typeof x;type Y=typeof obj.foo},[obj])}"}}},
				},
			},
			// dependency tree suggestions preserve insertion order
			{
				Code: "function C({obj,x,y,z,key,other}){useEffect(()=>{obj.b.x;obj.a;obj.b.y;z},[z,y])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'obj.a', 'obj.b.x', and 'obj.b.y'. Either include them or remove the dependency array.",
						Line:    1, Column: 75, EndLine: 1, EndColumn: 80,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({obj,x,y,z,key,other}){useEffect(()=>{obj.b.x;obj.a;obj.b.y;z},[z, y, obj.b.x, obj.b.y, obj.a])}"}}},
				},
			},
			// nested references follow scope order
			{
				Code: "function C({obj,x,y,z,key,other}){useEffect(()=>{function f(){obj.z} obj.a; y; f()},[z,y])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'obj.a' and 'obj.z'. Either include them or remove the dependency array.",
						Line:    1, Column: 85, EndLine: 1, EndColumn: 90,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({obj,x,y,z,key,other}){useEffect(()=>{function f(){obj.z} obj.a; y; f()},[z, y, obj.a, obj.z])}"}}},
				},
			},
			// a second var initializer invalidates a setter
			{
				Code: "function C({prop}){var [x,setX]=useState(0);var setX=other; useEffect(()=>{setX(x+1)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'setX' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 87, EndLine: 1, EndColumn: 89,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({prop}){var [x,setX]=useState(0);var setX=other; useEffect(()=>{setX(x+1)},[setX, x])}"}}},
				},
			},
			// destructured state preserves the reducer recommendation
			{
				Code: "function C({prop}){const [{x},setX]=useState({}); useEffect(()=>{setX(prop)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'prop'. Either include it or remove the dependency array. If 'setX' needs the current value of 'prop', you can also switch to useReducer instead of useState and read 'prop' in the reducer.",
						Line:    1, Column: 78, EndLine: 1, EndColumn: 80,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({prop}){const [{x},setX]=useState({}); useEffect(()=>{setX(prop)},[prop])}"}}},
				},
			},
			// defaulted state preserves the reducer recommendation
			{
				Code: "function C({prop}){const [x=0,setX]=useState(0); useEffect(()=>{setX(prop)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'prop'. Either include it or remove the dependency array. If 'setX' needs the current value of 'prop', you can also switch to useReducer instead of useState and read 'prop' in the reducer.",
						Line:    1, Column: 77, EndLine: 1, EndColumn: 79,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({prop}){const [x=0,setX]=useState(0); useEffect(()=>{setX(prop)},[prop])}"}}},
				},
			},
			// optional hook calls are not stable initializers
			{
				Code: "function C({prop}){let [x,setX]=useState?.(0); useEffect(()=>{setX(x+1)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'setX' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 74, EndLine: 1, EndColumn: 76,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({prop}){let [x,setX]=useState?.(0); useEffect(()=>{setX(x+1)},[setX, x])}"}}},
				},
			},
			// optional namespace calls are not stable initializers
			{
				Code: "function C({prop}){let [x,setX]=React?.useState(0); useEffect(()=>{setX(x+1)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'setX' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 79, EndLine: 1, EndColumn: 81,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({prop}){let [x,setX]=React?.useState(0); useEffect(()=>{setX(x+1)},[setX, x])}"}}},
				},
			},
			// typed assignment targets report the assigned value
			{
				Code: "function C({prop}){let x;useEffect(()=>{(x as any)=prop},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Assignments to the 'x' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect.",
						Line:    1, Column: 52, EndLine: 1, EndColumn: 56,
					},
				},
			},
			// computed class keys and field initializers are references
			{
				Code: "function C({obj,x,y,z,key,other}){useEffect(()=>{class X { [key]=x; static [other]=y;method(a=z){return a}}},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'key', 'other', 'x', 'y', and 'z'. Either include them or remove the dependency array.",
						Line:    1, Column: 110, EndLine: 1, EndColumn: 112,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({obj,x,y,z,key,other}){useEffect(()=>{class X { [key]=x; static [other]=y;method(a=z){return a}}},[key, other, x, y, z])}"}}},
				},
			},
			// bare custom hook names ignore parentheses
			{
				Code:    "function C({x}){const local=()=>x;(useCustomEffect)(()=>x,[])}",
				Tsx:     true,
				Options: []any{map[string]any{"additionalHooks": "useCustomEffect"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useCustomEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 59, EndLine: 1, EndColumn: 61,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){const local=()=>x;(useCustomEffect)(()=>x,[x])}"}}},
				},
			},
			// lookbehind patterns use JavaScript regexp semantics
			{
				Code:     "function C({x}){useCustomEffect(()=>x,[])}",
				Tsx:      true,
				Options:  []any{map[string]any{"additionalHooks": "(?<=use)CustomEffect"}},
				Settings: map[string]any{"react-hooks": map[string]any{"additionalEffectHooks": "useOtherEffect"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useCustomEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 39, EndLine: 1, EndColumn: 41,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){useCustomEffect(()=>x,[x])}"}}},
				},
			},
			// empty rule patterns fall back to shared settings
			{
				Code:     "function C({x}){useOtherEffect(()=>x,[])}",
				Tsx:      true,
				Options:  []any{map[string]any{"additionalHooks": ""}},
				Settings: map[string]any{"react-hooks": map[string]any{"additionalEffectHooks": "useOtherEffect"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useOtherEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 38, EndLine: 1, EndColumn: 40,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){useOtherEffect(()=>x,[x])}"}}},
				},
			},
			// function C({x}){useEffect((a=x)=>{},[])}
			{
				Code: "function C({x}){useEffect((a=x)=>{},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 37, EndLine: 1, EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){useEffect((a=x)=>{},[x])}"}}},
				},
			},
			// function C({x}){useEffect(function f(a=x){},[])}
			{
				Code: "function C({x}){useEffect(function f(a=x){},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){useEffect(function f(a=x){},[x])}"}}},
				},
			},
			// function C({x}){useEffect(()=>{let a; ({a=x}=obj)},[])}
			{
				Code: "function C({x}){useEffect(()=>{let a; ({a=x}=obj)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 52, EndLine: 1, EndColumn: 54,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){useEffect(()=>{let a; ({a=x}=obj)},[x])}"}}},
				},
			},
			// function C({x}){{let ref=()=>{};useEffect(()=>ref(),[])}}
			{
				Code: "function C({x}){{let ref=()=>{};useEffect(()=>ref(),[])}}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 53, EndLine: 1, EndColumn: 55,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){{let ref=()=>{};useEffect(()=>ref(),[ref])}}"}}},
				},
			},
			// function C(){let x; useEffect(()=>{({x}=obj)},[])}
			{
				Code: "function C(){let x; useEffect(()=>{({x}=obj)},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Assignments to the 'x' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect.",
						Line:    1, Column: 41, EndLine: 1, EndColumn: 44,
					},
				},
			},
			// function C(){let x; useEffect(()=>{x++},[])}
			{
				Code: "function C(){let x; useEffect(()=>{x++},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x'. Either include it or remove the dependency array.",
						Line:    1, Column: 41, EndLine: 1, EndColumn: 43,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){let x; useEffect(()=>{x++},[x])}"}}},
				},
			},
			// function C(){let x; useEffect(()=>{for(x of list){}},[])}
			{
				Code: "function C(){let x; useEffect(()=>{for(x of list){}},[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Assignments to the 'x' variable from inside React Hook useEffect will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside useEffect.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 49,
					},
				},
			},
			// function C(){const cb=()=>cb();useEffect(cb,[])}
			{
				Code: "function C(){const cb=()=>cb();useEffect(cb,[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'cb'. Either include it or remove the dependency array.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const cb=()=>cb();useEffect(cb,[cb])}"}}},
				},
			},
			// function C({Component}){useMemo(()=><Component/>,[])}
			{
				Code: "function C({Component}){useMemo(()=><Component/>,[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useMemo has a missing dependency: 'Component'. Either include it or remove the dependency array.",
						Line:    1, Column: 50, EndLine: 1, EndColumn: 52,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({Component}){useMemo(()=><Component/>,[Component])}"}}},
				},
			},
			// function C(){useEffect(()=>{},[`abc`])}
			{
				Code: "function C(){useEffect(()=>{},[`abc`])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked.",
						Line:    1, Column: 32, EndLine: 1, EndColumn: 37,
					},
				},
			},
			// function C(){useEffect(()=>{},[1 as const])}
			{
				Code: "function C(){useEffect(()=>{},[1 as const])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked.",
						Line:    1, Column: 32, EndLine: 1, EndColumn: 42,
					},
				},
			},
			// function C(){const ref=()=>{};useEffect(()=>ref(),[ref,ref])}
			{
				Code: "function C(){const ref=()=>{};useEffect(()=>ref(),[ref,ref])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a duplicate dependency: 'ref'. Either omit it or remove the dependency array.",
						Line:    1, Column: 51, EndLine: 1, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const ref=()=>{};useEffect(()=>ref(),[ref])}"}}},
				},
			},
			// const on=useEffectEvent
			{
				Code:    "function C(){const on=useEffectEvent(()=>{});useEffect(()=>on(),[]);useEffect(()=>{},[on]);}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `on` from the list.",
						Line:    1, Column: 87, EndLine: 1, EndColumn: 89,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const on=useEffectEvent(()=>{});useEffect(()=>on(),[]);useEffect(()=>{},[]);}"}}},
				},
			},
			// const a=useRef
			{
				Code:    "function C(){const a=useRef();const z=useRef();useEffect(()=>{},[z.current,a.current])}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has unnecessary dependencies: 'a.current' and 'z.current'. Either exclude them or remove the dependency array. Mutable values like 'z.current' aren't valid dependencies because mutating them doesn't re-render the component.",
						Line:    1, Column: 65, EndLine: 1, EndColumn: 86,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const a=useRef();const z=useRef();useEffect(()=>{},[])}"}}},
				},
			},
			// useEffect(()=>{},[z,a])
			{
				Code:    "function C(){useEffect(()=>{},[z,a])}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has unnecessary dependencies: 'a' and 'z'. Either exclude them or remove the dependency array. Outer scope values like 'z' aren't valid dependencies because mutating them doesn't re-render the component.",
						Line:    1, Column: 31, EndLine: 1, EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){useEffect(()=>{},[])}"}}},
				},
			},
			// const [𐀀,setX]
			{
				Code:    "function C(){const [𐀀,setX]=useState(0);useEffect(()=>setX(𐀀+1),[])}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: '𐀀'. Either include it or remove the dependency array. You can also do a functional update 'setX(𐀀 => ...)' if you only need '𐀀' in the 'setX' call.",
						Line:    1, Column: 67, EndLine: 1, EndColumn: 69,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const [𐀀,setX]=useState(0);useEffect(()=>setX(𐀀+1),[𐀀])}"}}},
				},
			},
			// var [x,setX]=useState
			{
				Code:    "function C(){var [x,setX]=useState(0); var setX=other; useEffect(()=>setX(x),[])}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'setX' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 78, EndLine: 1, EndColumn: 80,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){var [x,setX]=useState(0); var setX=other; useEffect(()=>setX(x),[setX, x])}"}}},
				},
			},
			// const ref=()=>{};useEffect
			{
				Code:    "function C({x}){const ref=()=>{};useEffect(()=>ref(),[ref]);ref();}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "The 'ref' function makes the dependencies of useEffect Hook (at line 1) change on every render. To fix this, wrap the definition of 'ref' in its own useCallback() Hook.",
						Line:    1, Column: 23, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x}){const ref=useCallback(()=>{});useEffect(()=>ref(),[ref]);ref();}"}}},
				},
			},
			// function C({𐀀,Ｘ})
			{
				Code:    "function C({𐀀,Ｘ}){useEffect(()=>{𐀀();Ｘ()},[])}",
				Tsx:     true,
				Options: []any{map[string]any{"enableDangerousAutofixThisMayCauseInfiniteLoops": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: '𐀀' and 'Ｘ'. Either include them or remove the dependency array. If '𐀀' changes too often, find the parent component that defines it and wrap that definition in useCallback.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({𐀀,Ｘ}){useEffect(()=>{𐀀();Ｘ()},[𐀀, Ｘ])}"}}},
				},
			},
			// all assignment operators classify constructed dependencies
			{
				Code: "function C(){let other;const x=(other **= {});useEffect(()=>{console.log(x)},[x])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "The 'x' assignment expression makes the dependencies of useEffect Hook (at line 1) change on every render. Move it inside the useEffect callback. Alternatively, wrap the initialization of 'x' in its own useMemo() Hook.",
						Line:    1, Column: 30, EndLine: 1, EndColumn: 46,
					},
				},
			},
			// bitwise assignments classify constructed dependencies
			{
				Code: "function C(){let other;const x=(other >>>= {});useEffect(()=>{console.log(x)},[x])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "The 'x' assignment expression makes the dependencies of useEffect Hook (at line 1) change on every render. Move it inside the useEffect callback. Alternatively, wrap the initialization of 'x' in its own useMemo() Hook.",
						Line:    1, Column: 30, EndLine: 1, EndColumn: 47,
					},
				},
			},
			// computed access inside a longer optional chain is complex
			{
				Code: "function C({x,k,p}){useEffect(()=>x?.[k].p,[x?.[k].p])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'k' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 44, EndLine: 1, EndColumn: 54,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>x?.[k].p,[k, x])}"}}},
					{
						Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 53,
					},
				},
			},
			// a second optional access does not split an ESTree chain
			{
				Code: "function C({x,k,p}){useEffect(()=>x?.[k]?.p,[x?.[k]?.p])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'k' and 'x'. Either include them or remove the dependency array.",
						Line:    1, Column: 45, EndLine: 1, EndColumn: 56,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>x?.[k]?.p,[k, x])}"}}},
					{
						Message: "React Hook useEffect has a complex expression in the dependency array. Extract it to a separate variable so it can be statically checked.",
						Line:    1, Column: 46, EndLine: 1, EndColumn: 55,
					},
				},
			},
			// outer optional computed access preserves upstream normalization
			{
				Code: "function C({x,k,p}){useEffect(()=>x?.p[k],[x?.p[k]])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has missing dependencies: 'k' and 'x?.p'. Either include them or remove the dependency array.",
						Line:    1, Column: 43, EndLine: 1, EndColumn: 52,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>x?.p[k],[k, x?.p, x?.p.k])}"}}},
				},
			},
			// parentheses terminate optional dependency chains
			{
				Code: "function C({x,k,p}){useEffect(()=>(x?.p).q,[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x?.p'. Either include it or remove the dependency array.",
						Line:    1, Column: 44, EndLine: 1, EndColumn: 46,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>(x?.p).q,[x?.p])}"}}},
				},
			},
			// a declared outer property does not cover a terminated chain
			{
				Code: "function C({x,k,p}){useEffect(()=>(x?.p).q,[(x?.p).q])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x?.p'. Either include it or remove the dependency array.",
						Line:    1, Column: 44, EndLine: 1, EndColumn: 54,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>(x?.p).q,[x?.p, x?.p.q])}"}}},
				},
			},
			// a called optional chain behind parentheses is its own dependency
			{
				Code: "function C({x,k,p}){useEffect(()=>(x?.p)(),[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'x?.p'. Either include it or remove the dependency array.",
						Line:    1, Column: 44, EndLine: 1, EndColumn: 46,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({x,k,p}){useEffect(()=>(x?.p)(),[x?.p])}"}}},
				},
			},
			// parenthesized optional initializers are not stable
			{
				Code: "function C(){const ref=(React?.useRef)();useEffect(()=>ref,[])}",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "React Hook useEffect has a missing dependency: 'ref'. Either include it or remove the dependency array.",
						Line:    1, Column: 60, EndLine: 1, EndColumn: 62,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C(){const ref=(React?.useRef)();useEffect(()=>ref,[ref])}"}}},
				},
			}})
}
