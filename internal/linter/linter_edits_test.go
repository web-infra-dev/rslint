package linter

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRunLinterDiagnosticConsumerEditDemand(t *testing.T) {
	const source = "const value = 1;\n"
	program, paths := createTestProgramWithFiles(t, map[string]string{"edits.ts": source})
	sourceFile := program.GetSourceFile(paths["edits.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 1 {
		t.Fatal("edit-demand fixture did not parse into one statement")
	}

	reportNode := sourceFile.Statements.Nodes[0]
	reportRange := utils.TrimNodeTextRange(sourceFile, reportNode)
	fix := rule.RuleFixReplaceRange(reportRange, "const replacement = 1;")
	fixMessage := rule.RuleMessage{
		Id:          "fix-diagnostic",
		Description: "always report the fix diagnostic",
		Data:        map[string]string{"kind": "invariant"},
	}
	suggestionMessage := rule.RuleMessage{
		Id:          "suggestion-diagnostic",
		Description: "always report the suggestion diagnostic",
	}
	suggestion := rule.RuleSuggestion{
		Message:  rule.RuleMessage{Id: "suggest", Description: "use replacement"},
		FixesArr: []rule.RuleFix{fix},
	}

	tests := []struct {
		name                string
		demand              rule.EditDemand
		wantFixBuilder      int
		wantSuggestBuilder  int
		wantFixAttachment   bool
		wantSuggestArtifact bool
	}{
		{name: "diagnostics", demand: 0},
		{
			name:              "autofixes",
			demand:            rule.EditDemandAutofix,
			wantFixBuilder:    1,
			wantFixAttachment: true,
		},
		{
			name:                "suggestions",
			demand:              rule.EditDemandSuggestion,
			wantSuggestBuilder:  1,
			wantSuggestArtifact: true,
		},
		{
			name:                "all edits",
			demand:              rule.EditDemandAll,
			wantFixBuilder:      1,
			wantSuggestBuilder:  1,
			wantFixAttachment:   true,
			wantSuggestArtifact: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixBuilderCalls := 0
			suggestionBuilderCalls := 0
			var got []rule.RuleDiagnostic
			timing := NewTimingCollector()

			_, err := RunLinter(RunLinterOptions{
				Programs:       []*compiler.Program{program},
				SingleThreaded: true,
				TargetFiles:    [][]string{{paths["edits.ts"]}},
				Timing:         timing,
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					return []ConfiguredRule{{
						Name:     "edit-demand",
						Severity: rule.SeverityWarning,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return rule.RuleListeners{
								ast.KindVariableStatement: func(node *ast.Node) {
									ctx.ReportNodeWithDeferredFixes(node, fixMessage, func() []rule.RuleFix {
										fixBuilderCalls++
										return []rule.RuleFix{fix}
									})
									ctx.ReportNodeWithDeferredSuggestions(node, suggestionMessage, func() []rule.RuleSuggestion {
										suggestionBuilderCalls++
										return []rule.RuleSuggestion{suggestion}
									})
								},
							}
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{
					Demand: testCase.demand,
					Report: func(d rule.RuleDiagnostic) { got = append(got, d) },
				},
			})
			if err != nil {
				t.Fatalf("RunLinter error: %v", err)
			}
			if len(got) != 2 {
				t.Fatalf("diagnostics = %d, want 2", len(got))
			}
			if fixBuilderCalls != testCase.wantFixBuilder {
				t.Errorf("fix builder calls = %d, want %d", fixBuilderCalls, testCase.wantFixBuilder)
			}
			if suggestionBuilderCalls != testCase.wantSuggestBuilder {
				t.Errorf("suggestion builder calls = %d, want %d", suggestionBuilderCalls, testCase.wantSuggestBuilder)
			}
			ruleTiming, ok := timing.Timings()["edit-demand"]
			if !ok || ruleTiming.Kind != RuleKindNative || ruleTiming.Files != 1 {
				t.Errorf("timed listener should preserve demand-driven reporting, got %#v", ruleTiming)
			}

			for index, diagnostic := range got {
				wantMessage := fixMessage
				if index == 1 {
					wantMessage = suggestionMessage
				}
				if diagnostic.RuleName != "edit-demand" ||
					!reflect.DeepEqual(diagnostic.Message, wantMessage) ||
					diagnostic.Range != reportRange ||
					diagnostic.SourceFile != sourceFile ||
					diagnostic.FilePath != paths["edits.ts"] ||
					diagnostic.Severity != rule.SeverityWarning ||
					diagnostic.Origin != rule.DiagnosticOriginLint ||
					diagnostic.PreFormatted {
					t.Errorf("diagnostic %d metadata changed: %#v", index, diagnostic)
				}
			}

			if (got[0].FixesPtr != nil) != testCase.wantFixAttachment {
				t.Errorf("fix presence = %v, want %v", got[0].FixesPtr != nil, testCase.wantFixAttachment)
			} else if testCase.wantFixAttachment && !reflect.DeepEqual(*got[0].FixesPtr, []rule.RuleFix{fix}) {
				t.Errorf("fixes = %#v, want %#v", *got[0].FixesPtr, []rule.RuleFix{fix})
			}
			if got[0].Suggestions != nil {
				t.Errorf("fix diagnostic unexpectedly has suggestions: %#v", *got[0].Suggestions)
			}
			if (got[1].Suggestions != nil) != testCase.wantSuggestArtifact {
				t.Errorf("suggestion presence = %v, want %v", got[1].Suggestions != nil, testCase.wantSuggestArtifact)
			} else if testCase.wantSuggestArtifact && !reflect.DeepEqual(*got[1].Suggestions, []rule.RuleSuggestion{suggestion}) {
				t.Errorf("suggestions = %#v, want %#v", *got[1].Suggestions, []rule.RuleSuggestion{suggestion})
			}
			if got[1].FixesPtr != nil {
				t.Errorf("suggestion diagnostic unexpectedly has autofixes: %#v", *got[1].FixesPtr)
			}
		})
	}
}

func TestRunLinterDeferredFixesSkipSuppressedDiagnostic(t *testing.T) {
	const source = "const visible = 1;\n// rslint-disable-next-line deferred-rule\nconst blocked = 2;\n"
	program, paths := createTestProgramWithFiles(t, map[string]string{"suppressed.ts": source})
	sourceFile := program.GetSourceFile(paths["suppressed.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 2 {
		t.Fatal("suppression fixture did not parse into two statements")
	}

	builderCalls := 0
	var got []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["suppressed.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "deferred-rule",
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					for _, node := range sourceFile.Statements.Nodes {
						ctx.ReportNodeWithDeferredFixes(
							node,
							rule.RuleMessage{Id: "report", Description: "report"},
							func() []rule.RuleFix {
								builderCalls++
								return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "replacement")}
							},
						)
					}
					return nil
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(d rule.RuleDiagnostic) { got = append(got, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("diagnostics = %d, want only the unsuppressed diagnostic", len(got))
	}
	if got[0].Range != utils.TrimNodeTextRange(sourceFile, sourceFile.Statements.Nodes[0]) {
		t.Fatalf("reported range = %v, want the unsuppressed statement", got[0].Range)
	}
	if builderCalls != 1 {
		t.Fatalf("builder calls = %d, want only the unsuppressed diagnostic", builderCalls)
	}
}

func TestRunLinterDeferredBuilderMayDeclineArtifact(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{"decline.ts": "const value = 1;\n"})
	sourceFile := program.GetSourceFile(paths["decline.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 1 {
		t.Fatal("decline fixture did not parse into one statement")
	}

	builderCalls := 0
	var got []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["decline.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "decline-fix",
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ctx.ReportNodeWithDeferredFixes(
						sourceFile.Statements.Nodes[0],
						rule.RuleMessage{Description: "still report"},
						func() []rule.RuleFix {
							builderCalls++
							return nil
						},
					)
					return nil
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(d rule.RuleDiagnostic) { got = append(got, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if builderCalls != 1 {
		t.Fatalf("builder calls = %d, want 1", builderCalls)
	}
	if len(got) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(got))
	}
	if got[0].FixesPtr != nil {
		t.Fatalf("declined fix should not create a fix attachment: %#v", *got[0].FixesPtr)
	}
}

func TestRunLinterLegacyReportsRespectEditDemand(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{"legacy.ts": "const value = 1;\n"})
	sourceFile := program.GetSourceFile(paths["legacy.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 1 {
		t.Fatal("legacy fixture did not parse into one statement")
	}
	node := sourceFile.Statements.Nodes[0]
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	fix := rule.RuleFixReplaceRange(textRange, "replacement")
	suggestion := rule.RuleSuggestion{
		Message:  rule.RuleMessage{Id: "suggest", Description: "suggest"},
		FixesArr: []rule.RuleFix{fix},
	}

	tests := []struct {
		name            string
		demand          rule.EditDemand
		wantFixes       bool
		wantSuggestions bool
	}{
		{name: "diagnostics"},
		{name: "autofixes", demand: rule.EditDemandAutofix, wantFixes: true},
		{name: "suggestions", demand: rule.EditDemandSuggestion, wantSuggestions: true},
		{name: "all edits", demand: rule.EditDemandAll, wantFixes: true, wantSuggestions: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var got []rule.RuleDiagnostic
			_, err := RunLinter(RunLinterOptions{
				Programs:       []*compiler.Program{program},
				SingleThreaded: true,
				TargetFiles:    [][]string{{paths["legacy.ts"]}},
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					return []ConfiguredRule{{
						Name:     "legacy-report",
						Severity: rule.SeverityWarning,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							ctx.ReportNodeWithFixesAndSuggestions(
								node,
								rule.RuleMessage{Id: "legacy", Description: "legacy"},
								[]rule.RuleFix{fix},
								[]rule.RuleSuggestion{suggestion},
							)
							return nil
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{
					Demand: testCase.demand,
					Report: func(d rule.RuleDiagnostic) { got = append(got, d) },
				},
			})
			if err != nil {
				t.Fatalf("RunLinter error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(got))
			}
			if (got[0].FixesPtr != nil) != testCase.wantFixes {
				t.Errorf("fix presence = %v, want %v", got[0].FixesPtr != nil, testCase.wantFixes)
			}
			if (got[0].Suggestions != nil) != testCase.wantSuggestions {
				t.Errorf("suggestion presence = %v, want %v", got[0].Suggestions != nil, testCase.wantSuggestions)
			}
		})
	}
}

func TestRunLinterDiscardingConsumerSkipsDeferredEdits(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{"discarded.ts": "const value = 1;\n"})
	sourceFile := program.GetSourceFile(paths["discarded.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 1 {
		t.Fatal("discard fixture did not parse into one statement")
	}

	builderCalls := 0
	result, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["discarded.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "discarded",
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ctx.ReportNodeWithDeferredFixes(
						sourceFile.Statements.Nodes[0],
						rule.RuleMessage{Description: "discarded"},
						func() []rule.RuleFix {
							builderCalls++
							return []rule.RuleFix{{}}
						},
					)
					return nil
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			// A nil report callback means there is no consumer. Even a
			// contradictory demand must not make discarded edits materialize.
			Demand: rule.EditDemandAll,
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Fatalf("linted files = %d, want 1", result.LintedFileCount)
	}
	if builderCalls != 0 {
		t.Fatalf("builder calls = %d, want 0 for a discarded consumer", builderCalls)
	}
}

func TestRunLinterRejectsInvalidEditDemand(t *testing.T) {
	_, err := RunLinter(RunLinterOptions{
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemand(1 << 7),
			Report: func(rule.RuleDiagnostic) {},
		},
	})
	if err == nil || err.Error() != "linter: invalid native edit demand" {
		t.Fatalf("error = %v, want invalid edit-demand failure", err)
	}
}

func TestLintSingleFileRejectsInvalidEditDemand(t *testing.T) {
	defer func() {
		got := recover()
		if got != "linter: invalid native edit demand" {
			t.Fatalf("panic = %v, want invalid edit-demand failure", got)
		}
	}()

	LintSingleFile(LintSingleFileOptions{
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemand(1 << 7),
			Report: func(rule.RuleDiagnostic) {},
		},
	})
}
