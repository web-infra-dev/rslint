package exhaustive_deps

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestExhaustiveDepsRule covers a curated subset of upstream
// `ESLintRuleExhaustiveDeps-test.js`. Cases are grouped by category to
// keep the file scrollable; comments above each case point at the
// corresponding upstream block when applicable.
//
// Upstream coverage that is NOT exercised here (and the reason):
//   - Flow-syntax tests (`component`, `hook`) — rslint's parser doesn't
//     parse Flow.
//   - Some scope-manager-dependent edge cases that require ESLint's
//     write-tracking semantics — we mark these as `Skip: true` with a
//     reason, so the file still documents the gap.
func TestExhaustiveDepsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		exhaustiveDepsValid,
		exhaustiveDepsInvalid,
	)
}

var exhaustiveDepsValid = []rule_tester.ValidTestCase{
	// ---- Empty deps with no captures ----
	{Code: `
		function MyComponent(props) {
			useEffect(() => {});
		}
	`, Tsx: true},
	{Code: `
		function MyComponent(props) {
			useEffect(() => { console.log('hi'); }, []);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent(props) {
			useEffect(() => {
				const local = {};
				console.log(local);
			}, []);
		}
	`, Tsx: true},

	// ---- Stable hook values: setState / useRef.current / dispatch ----
	{Code: `
		function MyComponent() {
			const [, setX] = useState(0);
			useEffect(() => { setX(1); }, []);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent() {
			const ref = useRef(null);
			useEffect(() => { ref.current = 1; }, []);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent() {
			const [, dispatch] = useReducer(reducer, 0);
			useEffect(() => { dispatch({ type: 'a' }); }, []);
		}
	`, Tsx: true},

	// ---- Effect callback referring to its own deps ----
	{Code: `
		function MyComponent({ id }) {
			useEffect(() => { console.log(id); }, [id]);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent({ items }) {
			useMemo(() => items.length, [items]);
		}
	`, Tsx: true},
	{Code: `
		function MyComponent({ a, b }) {
			useCallback(() => a + b, [a, b]);
		}
	`, Tsx: true},

	// ---- useEffectEvent — return value is stable, no need to list ----
	{Code: `
		function MyComponent({ theme }) {
			const onClick = useEffectEvent(() => { console.log(theme); });
			useEffect(() => { onClick(); }, []);
		}
	`, Tsx: true},

	// ---- useImperativeHandle (callbackIndex = 1) ----
	{Code: `
		function MyComponent({ value }, ref) {
			useImperativeHandle(ref, () => ({ get: () => value }), [value]);
		}
	`, Tsx: true},

	// ---- Optional chain in deps ----
	{Code: `
		function MyComponent({ user }) {
			useEffect(() => { console.log(user?.name); }, [user?.name]);
		}
	`, Tsx: true},

	// ---- Property chain: declaring 'props.foo' covers 'props.foo.bar' ----
	{Code: `
		function MyComponent(props) {
			useEffect(() => { console.log(props.foo.bar); }, [props.foo]);
		}
	`, Tsx: true},

	// ---- Reference to an external (module-scope) value — not a dep ----
	{Code: `
		const CONSTANT = 1;
		function MyComponent() {
			useEffect(() => { console.log(CONSTANT); }, []);
		}
	`, Tsx: true},

	// ---- additionalHooks option matching ----
	{
		Code: `
			function MyComponent({ id }) {
				useMyEffect(() => { console.log(id); }, [id]);
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"additionalHooks": "(useMyEffect)"},
	},

	// ---- nested function inside callback captures the value indirectly ----
	{Code: `
		function MyComponent({ id }) {
			useEffect(() => {
				function inner() { console.log(id); }
				inner();
			}, [id]);
		}
	`, Tsx: true},

	// ---- TS as expression around deps array ----
	{Code: `
		function MyComponent({ id }) {
			useEffect(() => { console.log(id); }, ([id] as const));
		}
	`, Tsx: true},

	// ---- useLayoutEffect / useInsertionEffect treated as effects ----
	{Code: `
		function MyComponent({ id }) {
			useLayoutEffect(() => { console.log(id); }, [id]);
			useInsertionEffect(() => { console.log(id); }, [id]);
		}
	`, Tsx: true},

	// ---- React.useEffect via namespace ----
	{Code: `
		function MyComponent({ id }) {
			React.useEffect(() => { console.log(id); }, [id]);
		}
	`, Tsx: true},

	// ---- Top-level call (not inside a component) — no diagnostic ----
	{Code: `
		useEffect(() => { /* no enclosing fn */ }, []);
	`, Tsx: true},
}

var exhaustiveDepsInvalid = []rule_tester.InvalidTestCase{
	// ---- Missing dep: simple ----
	{
		Code: `
			function MyComponent({ id }) {
				useEffect(() => { console.log(id); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{Output: `
			function MyComponent({ id }) {
				useEffect(() => { console.log(id); }, [id]);
			}
		`},
				},
			},
		},
	},

	// ---- Missing dep with property chain ----
	{
		Code: `
			function MyComponent(props) {
				useEffect(() => { console.log(props.id); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'props.id'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{Output: `
			function MyComponent(props) {
				useEffect(() => { console.log(props.id); }, [props.id]);
			}
		`},
				},
			},
		},
	},

	// ---- Unnecessary dep on useCallback ----
	{
		Code: `
			function MyComponent({ a, b }) {
				useCallback(() => a, [a, b]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has an unnecessary dependency: 'b'. Either exclude it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{Output: `
			function MyComponent({ a, b }) {
				useCallback(() => a, [a]);
			}
		`},
				},
			},
		},
	},

	// ---- Duplicate dep ----
	{
		Code: `
			function MyComponent({ a }) {
				useCallback(() => a, [a, a]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has a duplicate dependency: 'a'. Either omit it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{Output: `
			function MyComponent({ a }) {
				useCallback(() => a, [a]);
			}
		`},
				},
			},
		},
	},

	// ---- Async useEffect ----
	{
		Code: `
			function MyComponent() {
				useEffect(async () => { await foo(); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: ""},
		},
	},

	// ---- useMemo with no deps array ----
	{
		Code: `
			function MyComponent({ a }) {
				useMemo(() => a);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
		},
	},

	// ---- useCallback with no deps array ----
	{
		Code: `
			function MyComponent({ a }) {
				useCallback(() => a);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useCallback does nothing when called with only one argument. Did you forget to pass an array of dependencies?"},
		},
	},

	// ---- Spread element in deps ----
	{
		Code: `
			function MyComponent({ list }) {
				useEffect(() => {}, [...list]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies."},
		},
	},

	// ---- Deps array is not an array literal ----
	{
		Code: `
			function MyComponent({ a }) {
				useEffect(() => {}, a);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies."},
		},
	},

	// ---- String literal in deps ----
	{
		Code: `
			function MyComponent() {
				useEffect(() => {}, ['foo']);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The 'foo' literal is not a valid dependency because it never changes. You can safely remove it."},
		},
	},

	// ---- ref.current in cleanup ----
	{
		Code: `
			function MyComponent() {
				const ref = useRef(null);
				useEffect(() => {
					return () => { console.log(ref.current); };
				}, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function."},
		},
	},

	// ---- Async useEffect callback ----
	{
		Code: `
			function MyComponent() {
				useEffect(async () => { await foo(); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "Effect callbacks are synchronous to prevent race conditions. " +
				"Put the async function inside:\n\n" +
				"useEffect(() => {\n" +
				"  async function fetchData() {\n" +
				"    // You can await here\n" +
				"    const response = await MyAPI.getData(someId);\n" +
				"    // ...\n" +
				"  }\n" +
				"  fetchData();\n" +
				"}, [someId]); // Or [] if effect doesn't need props or state\n\n" +
				"Learn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching"},
		},
	},

	// ---- useEffect missing callback argument ----
	{
		Code: `
			function MyComponent() {
				useEffect();
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect requires an effect callback. Did you forget to pass a callback to the hook?"},
		},
	},

	// ---- Numeric literal in deps ----
	{
		Code: `
			function MyComponent() {
				useEffect(() => {}, [42]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The 42 literal is not a valid dependency because it never changes. You can safely remove it."},
		},
	},

	// ---- Functions returned from useEffectEvent must not be in deps ----
	{
		Code: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, [onClick]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{Output: `
			function MyComponent({ theme }) {
				const onClick = useEffectEvent(() => { console.log(theme); });
				useEffect(() => { onClick(); }, []);
			}
		`},
				}},
		},
	},

	// ---- requireExplicitEffectDeps: missing deps array on useEffect ----
	{
		Code: `
			function MyComponent() {
				useEffect(() => {});
			}
		`,
		Tsx:     true,
		Options: map[string]interface{}{"requireExplicitEffectDeps": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect always requires dependencies. Please add a dependency array or an explicit `undefined`"},
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
			name: "unresolved callback replacement",
			code: `function Component() {
				useEffect(missingCallback, []);
			}`,
			message:           "React Hook useEffect has a missing dependency: 'missingCallback'. Either include it or remove the dependency array.",
			suggestionMessage: "Update the dependencies array to be: [missingCallback]",
			fixTexts:          []string{"[missingCallback]"},
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
		{name: "suggestions", options: rule.NormalizeOptions(nil)},
		{
			name: "dangerous autofix",
			options: rule.NormalizeOptions(map[string]interface{}{
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
					if len(*autofixOnly) != 1 || !reflect.DeepEqual((*autofixOnly)[0], suggestion.FixesArr[0]) {
						t.Fatalf("dangerous autofix = %#v, want the first suggestion fix %#v", *autofixOnly, suggestion.FixesArr[0])
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
		Program:         program,
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: exhaustiveDepsConfiguredRules(options),
		ExcludePaths:    []string{},
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
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func exhaustiveDepsConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     ExhaustiveDepsRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ExhaustiveDepsRule.Run(ctx, options)
			},
		}}
	}
}
