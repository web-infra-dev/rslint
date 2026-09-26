package linter

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestDeferredPipelineReleasesCompletedASTsBeforeNextFile(t *testing.T) {
	var first weak.Pointer[ast.SourceFile]
	var calls atomic.Int32
	configured := []rule.ConfiguredRule{{Name: "native/check", Run: func(ctx rule.RuleContext) rule.RuleListeners {
		if len(ctx.Program().SourceFiles()) != 1 {
			t.Error("independent execution retained other root ASTs")
		}
		if calls.Add(1) == 1 {
			first = weak.Make(ctx.SourceFile)
		} else {
			// Collection is only a test probe. Production never forces a GC.
			runtime.GC()
			runtime.GC()
			if first.Value() != nil {
				t.Error("completed file remains reachable during the next file")
			}
		}
		return rule.RuleListeners{ast.KindIdentifier: func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{Description: "identifier"})
		}}
	}}}
	generation := pipelineDeferredTestGeneration(t, map[string]string{
		"first.ts": "const first = '😀';\r\nfirst;", "second.ts": "const second = 2; second;",
	}, func(string) []rule.ConfiguredRule { return configured })
	var releases int
	result, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, func() { releases++ }), ObservationPolicy{}, nil,
	))
	if err != nil || calls.Load() != 2 || releases != 1 {
		t.Fatalf("error/calls/releases = %v/%d/%d", err, calls.Load(), releases)
	}
	native := result.Observation.Native
	if native.Lint.LintedFileCount != 2 || len(native.Diagnostics) != 4 || len(native.Files) != 0 {
		t.Fatalf("unexpected observation: %+v", native)
	}
	for _, diagnostic := range native.Diagnostics {
		if _, astRetained := diagnostic.SourceFile.(*ast.SourceFile); astRetained {
			t.Fatal("deferred diagnostic retained its AST")
		}
		if diagnostic.SourceFile.Text() == "" || len(diagnostic.SourceFile.ECMALineMap()) == 0 {
			t.Fatal("deferred diagnostic lost its source presentation")
		}
	}
}

func TestDeferredPipelinePreservesSyntaxAndZeroRuleCounts(t *testing.T) {
	var calls atomic.Int32
	generation := pipelineDeferredTestGeneration(t, map[string]string{
		"invalid.ts": "const invalid = ;", "empty.ts": "", "valid.ts": "const value = 1;",
	}, func(path string) []rule.ConfiguredRule {
		if strings.HasSuffix(path, "empty.ts") {
			return nil
		}
		return []rule.ConfiguredRule{
			{Name: "native/check", Run: func(ctx rule.RuleContext) rule.RuleListeners {
				calls.Add(1)
				ctx.ReportRange(core.NewTextRange(0, 1), rule.RuleMessage{Description: "check"})
				return nil
			}},
			{Name: "typed/check", RequiresTypeInfo: true, RequiresProgram: true, Run: func(rule.RuleContext) rule.RuleListeners {
				panic("source-only execution enabled a type-aware rule")
			}},
		}
	})
	result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
	if err != nil {
		t.Fatal(err)
	}
	native := result.Observation.Native
	if calls.Load() != 1 || native.Lint.LintedFileCount != 3 || !native.HasTargetSyntaxErrors || len(native.Diagnostics) < 2 {
		t.Fatalf("calls=%d observation=%+v", calls.Load(), native)
	}
	if len(result.ExecutedRules()) != 1 {
		t.Fatalf("executed rules = %v", result.ExecutedRules())
	}
}

func TestDeferredPipelinePreservesWholeProgramConsumers(t *testing.T) {
	for _, reason := range []string{"rule", "files", "edits", "type-check", "plugin", "no-path-resolver"} {
		t.Run(reason, func(t *testing.T) {
			var calls atomic.Int32
			configured := []rule.ConfiguredRule{{
				Name: "native/check", RequiresProgram: reason == "rule",
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					calls.Add(1)
					if len(ctx.Program().SourceFiles()) != 2 {
						t.Error("whole-Program consumer lost the zero-rule dependency")
					}
					ctx.ReportNode(ctx.SourceFile.AsNode(), rule.RuleMessage{Description: "check"})
					return nil
				},
			}}
			generation := pipelineDeferredTestGeneration(t, map[string]string{
				"first.ts": "import './second';", "second.ts": "export const value = 1;",
			}, func(path string) []rule.ConfiguredRule {
				if strings.HasSuffix(path, "second.ts") {
					return nil
				}
				return configured
			})
			policy := ObservationPolicy{}
			switch reason {
			case "files":
				policy.Demand.LintedFiles = true
			case "edits":
				policy.Demand.Native = rule.EditDemandAll
			case "type-check":
				generation.Native.TypeCheck = true
			case "plugin":
				generation.Plugin = &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig { return EslintPluginFileConfig{} }}
			case "no-path-resolver":
				generation.Native.RulesForPath = nil
			}
			result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), policy, nil))
			if err != nil || calls.Load() != 1 {
				t.Fatalf("error/calls = %v/%d", err, calls.Load())
			}
			if result.Observation.Native.Lint.LintedFileCount != 2 {
				t.Fatal("zero-rule dependency was not counted")
			}
			if reason == "files" && len(result.Observation.Native.Files) != 2 {
				t.Fatal("requested file artifacts were not retained")
			}
		})
	}
}

func TestDeferredPipelineJoinsWorkersBeforeReleaseOnFailure(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)
	for _, failure := range []string{"cancel", "panic", "goexit"} {
		t.Run(failure, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bothStarted := make(chan struct{})
			finishOther := make(chan struct{})
			var active, entered, released atomic.Int32
			configured := []rule.ConfiguredRule{{Name: "native/check", Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
				active.Add(1)
				defer active.Add(-1)
				if entered.Add(1) == 2 {
					close(bothStarted)
				}
				select {
				case <-bothStarted:
				case <-ctx.Done():
					return nil
				}
				if strings.HasSuffix(ruleCtx.SourceFile.FileName(), "first.ts") {
					defer close(finishOther)
					switch failure {
					case "cancel":
						cancel()
					case "panic":
						panic("deferred failure")
					case "goexit":
						runtime.Goexit()
					}
				} else {
					select {
					case <-finishOther:
					case <-ctx.Done():
					}
				}
				return nil
			}}}
			generation := pipelineDeferredTestGeneration(t, map[string]string{
				"first.ts": "export {};", "second.ts": "export {};",
			}, func(string) []rule.ConfiguredRule { return configured })
			generation.Native.SingleThreaded = false
			type outcome struct {
				err        error
				panicValue any
			}
			done := make(chan outcome, 1)
			go func() {
				result := outcome{}
				defer func() { result.panicValue = recover(); done <- result }()
				_, result.err = RunPipeline(ctx, NewLintRequest(pipelineTestProvider(generation, func() {
					if active.Load() != 0 {
						t.Error("generation released while a worker is active")
					}
					released.Add(1)
				}), ObservationPolicy{}, nil))
			}()
			select {
			case result := <-done:
				if released.Load() != 1 || (failure == "cancel" && !errors.Is(result.err, context.Canceled)) ||
					(failure != "cancel" && result.panicValue == nil) {
					t.Fatalf("result=%+v releases=%d", result, released.Load())
				}
			case <-time.After(10 * time.Second):
				cancel()
				t.Fatal("deferred workers did not finish")
			}
		})
	}
}

func TestPipelineDiagnosticSourcesDetachAcrossPublicationModes(t *testing.T) {
	for _, mode := range []struct {
		name string
		kind PluginExecution
	}{
		{"concurrent", PluginConcurrentJoined},
		{"after-native", PluginAfterNativeJoined},
		{"progressive", pluginProgressiveAfterNative},
	} {
		t.Run(mode.name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			path := tspath.ResolvePath(root, "source.ts")
			text := "const face = '😀';\r\nlet value = face;\u2028value;\u2029value;"
			start := strings.LastIndex(text, "value")
			textRange := core.NewTextRange(start, start+len("value"))
			fixes := []rule.RuleFix{{Range: textRange, Text: "face"}}
			suggestions := []rule.RuleSuggestion{{
				Message:  rule.RuleMessage{Id: "suggest", Description: "use face", Data: map[string]string{"name": "face"}},
				FixesArr: fixes,
			}}
			generation := pipelineTestGeneration(t, root, path, text, []rule.ConfiguredRule{
				{
					Name: "native/check", Severity: rule.SeverityWarning,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						ctx.ReportRangeWithFixesAndSuggestions(textRange, rule.RuleMessage{Id: "check", Description: "check value"}, fixes, suggestions)
						ctx.ReportRange(textRange, rule.RuleMessage{Description: "second diagnostic"})
						return nil
					},
				},
				{Name: "plugin/check", IsEslintPluginRule: true, Severity: rule.SeverityError},
			}, &EslintPluginFileConfig{})
			original := generation.Native.Programs[0].SourceFiles()[0]
			var releases int
			provider := pipelineTestProvider(generation, func() { releases++ })
			demand := ArtifactDemand{Native: rule.EditDemandAll, Plugin: rule.EditDemandAll, LintedFiles: true}
			presentation := &pipelineProgressiveDiagnostics{}
			var result PipelineResult
			var err error
			if mode.kind == pluginProgressiveAfterNative {
				result, err = RunPipeline(context.Background(), NewProgressiveLintRequest(provider, demand, presentation))
			} else {
				result, err = RunPipeline(context.Background(), NewLintRequest(provider, ObservationPolicy{
					Demand: demand, Plugin: mode.kind,
				}, func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
						FilePath:    request.Files[0].Path,
						Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "plugin", StartPos: 0, EndPos: 1}},
					}}}, nil
				}))
			}
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := result.Observation.Native.Diagnostics
			if releases != 1 || len(diagnostics) != 2 {
				t.Fatalf("releases/diagnostics = %d/%d", releases, len(diagnostics))
			}
			source := diagnostics[0].SourceFile
			if _, retainedAST := source.(*ast.SourceFile); retainedAST {
				t.Fatal("published diagnostic retained its AST")
			}
			if source != diagnostics[1].SourceFile || source.Text() != text || !slices.Equal(source.ECMALineMap(), original.ECMALineMap()) {
				t.Fatal("diagnostics lost shared source text or line metadata")
			}
			for offset := range text {
				wantLine, wantColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(original, offset)
				line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(source, offset)
				if line != wantLine || column != wantColumn {
					t.Fatalf("position at byte %d = %d:%d, want %d:%d", offset, line, column, wantLine, wantColumn)
				}
			}
			if diagnostics[0].Range != textRange || diagnostics[0].FilePath != path ||
				diagnostics[0].Severity != rule.SeverityWarning || diagnostics[0].Message.Id != "check" ||
				!reflect.DeepEqual(diagnostics[0].Fixes(), fixes) || !reflect.DeepEqual(*diagnostics[0].Suggestions, suggestions) {
				t.Fatalf("diagnostic payload changed: %+v", diagnostics[0])
			}
			if len(result.Observation.Native.Files) != 1 || result.Observation.Native.Files[0].SourceFile != original {
				t.Fatal("explicit source-file demand lost its AST")
			}
			if mode.kind == pluginProgressiveAfterNative {
				if len(presentation.baseline) != 2 || presentation.baseline[0].SourceFile != source || presentation.run == nil {
					t.Fatal("progressive presentation did not receive the detached baseline")
				}
			} else {
				joined, ok := result.Observation.JoinedPluginOutcome()
				if !ok || len(joined.Diagnostics) != 1 {
					t.Fatalf("joined plugin outcome = %+v", joined)
				}
				if _, retainedAST := joined.Diagnostics[0].SourceFile.(*ast.SourceFile); retainedAST {
					t.Fatal("joined plugin diagnostic retained its AST")
				}
				if mode.kind == PluginConcurrentJoined && joined.Diagnostics[0].SourceFile != source {
					t.Fatal("concurrent native and plugin diagnostics lost shared source identity")
				}
			}
		})
	}
}

func TestObservationDiagnosticSourcesPreserveDistinctGenerations(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	path := tspath.ResolvePath(root, "source.ts")
	first := pipelineTestProgram(t, root, path, "a").SourceFiles()[0]
	second := pipelineTestProgram(t, root, path, "a").SourceFiles()[0]
	textOnly := newTextSourceFile("plugin")
	observation := ObservationResult{Native: NativeObservation{Diagnostics: []rule.RuleDiagnostic{
		{FilePath: path, SourceFile: first},
		{FilePath: path, SourceFile: second},
		{FilePath: path, SourceFile: first},
		{FilePath: path, SourceFile: textOnly},
		{FilePath: path},
	}}}
	observation.detachDiagnosticSources()
	diagnostics := observation.Native.Diagnostics
	if diagnostics[0].SourceFile == diagnostics[1].SourceFile || diagnostics[0].SourceFile != diagnostics[2].SourceFile {
		t.Fatal("source identity was replaced by path or text equality")
	}
	if diagnostics[3].SourceFile != textOnly || diagnostics[4].SourceFile != nil {
		t.Fatal("non-AST diagnostic source was replaced")
	}
	firstProjection := diagnostics[0].SourceFile
	observation.detachDiagnosticSources()
	if diagnostics[0].SourceFile != firstProjection {
		t.Fatal("detaching an existing projection changed its identity")
	}
}

func TestProgressivePipelineChecksCancellationAfterRelease(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(t, root, fileName, "a", nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	presentation := &pipelineProgressiveDiagnostics{}
	result, err := RunPipeline(ctx, NewProgressiveLintRequest(
		pipelineTestProvider(generation, ReleaseFunc(cancel)),
		ArtifactDemand{},
		presentation,
	))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pipeline error = %v, want cancellation raised by release", err)
	}
	if result.Observation.Native.Lint != nil {
		t.Fatal("post-release canceled observation was published")
	}
	if presentation.baseline != nil || presentation.run != nil {
		t.Fatal("canceled progressive result reached presentation ports")
	}
}

func TestProgressivePipelineOwnsReleasePresentationGateAndSubmissionOrder(t *testing.T) {
	t.Run("eligible enrichment", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		fileName := tspath.ResolvePath(root, "source.ts")
		generation := pipelineTestGeneration(
			t,
			root,
			fileName,
			"const value = 1;",
			[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}},
			&EslintPluginFileConfig{},
		)
		released := false
		presented := false
		presentation := &pipelineProgressiveDiagnostics{
			onPublish: func() {
				if !released {
					t.Fatal("baseline was published before generation release")
				}
				presented = true
			},
			onSubmit: func() {
				if !presented {
					t.Fatal("enrichment was submitted before baseline publication")
				}
			},
		}
		result, err := RunPipeline(context.Background(), NewProgressiveLintRequest(
			pipelineTestProvider(generation, func() { released = true }),
			ArtifactDemand{Plugin: rule.EditDemandAll},
			presentation,
		))
		if err != nil {
			t.Fatal(err)
		}
		if presentation.run == nil || !released || !presented {
			t.Fatalf("run/released/presented = %v/%v/%v", presentation.run != nil, released, presented)
		}
		if _, complete := result.Observation.CompleteDiagnostics(); complete {
			t.Fatal("progressive result reported complete before enrichment")
		}
	})

	for _, test := range []struct {
		name         string
		text         string
		pluginConfig *EslintPluginFileConfig
		rules        []rule.ConfiguredRule
	}{
		{
			name:         "target syntax error",
			text:         "const value = ;",
			pluginConfig: &EslintPluginFileConfig{},
			rules:        []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}},
		},
		{name: "no plugin work", text: "const value = 1;"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			fileName := tspath.ResolvePath(root, "source.ts")
			generation := pipelineTestGeneration(t, root, fileName, test.text, test.rules, test.pluginConfig)
			presentation := &pipelineProgressiveDiagnostics{}
			result, err := RunPipeline(context.Background(), NewProgressiveLintRequest(
				pipelineTestProvider(generation, nil),
				ArtifactDemand{Plugin: rule.EditDemandAll},
				presentation,
			))
			if err != nil {
				t.Fatal(err)
			}
			if presentation.run != nil {
				t.Fatal("ineligible enrichment was submitted")
			}
			if _, complete := result.Observation.CompleteDiagnostics(); !complete {
				t.Fatal("baseline without enrichment was reported incomplete")
			}
		})
	}
}

func TestConcurrentPipelineChecksCancellationAfterRelease(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(t, root, fileName, "a", nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	result, err := RunPipeline(ctx, NewLintRequest(
		pipelineTestProvider(generation, ReleaseFunc(cancel)),
		ObservationPolicy{Plugin: PluginConcurrentJoined},
		nil,
	))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pipeline error = %v, want cancellation raised by release", err)
	}
	if result.Observation.Native.Lint != nil {
		t.Fatal("post-release canceled observation was published")
	}
}

func TestPipelineAfterNativeReleasesBeforePluginDispatch(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"a",
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}},
		&EslintPluginFileConfig{},
	)
	var releases atomic.Int32
	_, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, func() { releases.Add(1) }),
		ObservationPolicy{
			Demand:        ArtifactDemand{Plugin: rule.EditDemandAll},
			Plugin:        PluginAfterNativeJoined,
			PluginFailure: PluginDiscardOnFailure,
		},
		func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			if releases.Load() != 1 {
				t.Fatalf("release count at plugin dispatch = %d, want 1", releases.Load())
			}
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: request.Files[0].Path}}}, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	if releases.Load() != 1 {
		t.Fatalf("release calls = %d, want 1", releases.Load())
	}
}

func TestProgressivePluginRunIsFrozenAndSingleUse(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	settings := map[string]any{"value": "frozen"}
	options := []any{map[string]any{"choice": "frozen"}}
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"const value = 1;",
		[]rule.ConfiguredRule{{
			Name:               "plugin/check",
			IsEslintPluginRule: true,
			Options:            options,
		}},
		&EslintPluginFileConfig{Settings: settings},
	)
	var releases atomic.Int32
	presentation := &pipelineProgressiveDiagnostics{}
	_, err := RunPipeline(context.Background(), NewProgressiveLintRequest(
		pipelineTestProvider(generation, func() { releases.Add(1) }),
		ArtifactDemand{Plugin: rule.EditDemandAll},
		presentation,
	))
	if err != nil {
		t.Fatal(err)
	}
	if presentation.run == nil || releases.Load() != 1 {
		t.Fatalf("enrichment/release = %v/%d, want non-nil/1", presentation.run != nil, releases.Load())
	}
	settings["value"] = "mutated"
	options[0].(map[string]any)["choice"] = "mutated"
	var request EslintPluginLintRequest
	outcome, err := presentation.run(context.Background(), func(_ context.Context, got EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		request = got
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: got.Files[0].Path}}}, nil
	})
	if err != nil || outcome.DispatchError != nil {
		t.Fatalf("work errors = %v/%v", err, outcome.DispatchError)
	}
	frozenOptions := request.Rules["plugin/check"].Options
	if request.Files[0].Settings["value"] != "frozen" ||
		len(frozenOptions) != 1 || frozenOptions[0].(map[string]any)["choice"] != "frozen" {
		t.Fatalf("deferred request retained mutable config: %+v", request)
	}
	if _, err := presentation.run(context.Background(), func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{}, nil
	}); !errors.Is(err, ErrDeferredPluginRunAlreadyInvoked) {
		t.Fatalf("second run error = %v, want ErrDeferredPluginRunAlreadyInvoked", err)
	}
}

func TestConcurrentPipelineCancelsAndJoinsPluginBeforeReleaseOnNativePanic(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	pluginStarted := make(chan struct{})
	pluginStopped := make(chan struct{})
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"a",
		[]rule.ConfiguredRule{
			{
				Name: "native/panic",
				Run: func(rule.RuleContext) rule.RuleListeners {
					<-pluginStarted
					panic("native failed")
				},
			},
			{Name: "plugin/check", IsEslintPluginRule: true},
		},
		&EslintPluginFileConfig{},
	)
	var releases atomic.Int32
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _ = RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, func() {
				select {
				case <-pluginStopped:
				default:
					t.Fatal("generation released before plugin dispatch joined")
				}
				releases.Add(1)
			}),
			ObservationPolicy{Plugin: PluginConcurrentJoined},
			func(pluginCtx context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
				close(pluginStarted)
				<-pluginCtx.Done()
				close(pluginStopped)
				return nil, pluginCtx.Err()
			},
		))
	}()
	if recovered == nil || releases.Load() != 1 {
		t.Fatalf("panic/releases = %v/%d, want panic/1", recovered, releases.Load())
	}
}
