package exhaustive_deps

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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
