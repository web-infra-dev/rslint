package linter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestPipelinePublishesTextDiagnosticSources(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined, pluginProgressiveAfterNative} {
		t.Run(fmt.Sprintf("mode-%d", mode), func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			path := tspath.ResolvePath(root, "source.ts")
			const text = "const value = 1;\r\n"
			generation := pipelineTestGeneration(t, root, path, text, []rule.ConfiguredRule{
				{
					Name: "native/check", Severity: rule.SeverityWarning,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						ctx.ReportRange(core.NewTextRange(6, 11), rule.RuleMessage{Id: "value", Description: "value"})
						return nil
					},
				},
				{Name: "plugin/check", IsEslintPluginRule: true, Severity: rule.SeverityError},
			}, &EslintPluginFileConfig{})
			original := generation.Native.Programs[0].SourceFiles()[0]
			var releases int
			provider := pipelineTestProvider(generation, func() { releases++ })
			dispatch := func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
				if mode != PluginConcurrentJoined && releases != 1 {
					t.Fatalf("detached dispatch preceded generation release: %d", releases)
				}
				return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
					FilePath:    request.Files[0].Path,
					Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "plugin", StartPos: 6, EndPos: 11}},
				}}}, nil
			}
			var result PipelineResult
			var err error
			var pluginDiagnostics []rule.RuleDiagnostic
			if mode == pluginProgressiveAfterNative {
				presentation := &pipelineProgressiveDiagnostics{}
				result, err = RunPipeline(context.Background(), NewProgressiveLintRequest(provider, ArtifactDemand{}, presentation))
				if err != nil || presentation.run == nil || len(presentation.baseline) != 1 {
					t.Fatalf("progressive observation = %+v, error = %v", presentation, err)
				}
				if _, retained := presentation.baseline[0].SourceFile.(*ast.SourceFile); retained {
					t.Fatal("progressive baseline retained a compiler AST")
				}
				outcome, runErr := presentation.run(context.Background(), dispatch)
				if runErr != nil || outcome.DispatchError != nil {
					t.Fatalf("progressive plugin errors = %v / %v", runErr, outcome.DispatchError)
				}
				pluginDiagnostics = outcome.Diagnostics
			} else {
				result, err = RunPipeline(context.Background(), NewLintRequest(provider, ObservationPolicy{
					Plugin: mode, Demand: ArtifactDemand{LintedFiles: true},
				}, dispatch))
				if err != nil {
					t.Fatal(err)
				}
				outcome, joined := result.Observation.JoinedPluginOutcome()
				if !joined {
					t.Fatal("joined plugin result is missing")
				}
				pluginDiagnostics = outcome.Diagnostics
				files := result.Observation.Native.Files
				if len(files) != 1 || files[0].SourceFile != original {
					t.Fatal("detaching diagnostics changed an explicitly requested AST artifact")
				}
			}
			if releases != 1 || len(result.Observation.Native.Diagnostics) != 1 || len(pluginDiagnostics) != 1 {
				t.Fatalf("release/native/plugin counts = %d / %d / %d", releases, len(result.Observation.Native.Diagnostics), len(pluginDiagnostics))
			}
			native := result.Observation.Native.Diagnostics[0]
			if native.FilePath != path || native.Severity != rule.SeverityWarning || native.Message.Id != "value" || native.Range != core.NewTextRange(6, 11) {
				t.Fatalf("native diagnostic metadata changed: %+v", native)
			}
			for _, diagnostic := range []rule.RuleDiagnostic{native, pluginDiagnostics[0]} {
				if _, retained := diagnostic.SourceFile.(*ast.SourceFile); retained {
					t.Fatal("published diagnostic retained a compiler AST")
				}
				if diagnostic.SourceFile.Text() != text {
					t.Fatalf("diagnostic text = %q", diagnostic.SourceFile.Text())
				}
			}
			if mode == PluginConcurrentJoined && native.SourceFile != pluginDiagnostics[0].SourceFile {
				t.Fatal("diagnostics from one source object did not share a text projection")
			}
		})
	}
}

func TestDiagnosticSourceProjectionPreservesIdentityAndPositions(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	path := tspath.ResolvePath(root, "source.ts")
	const text = "// 😀\r\nconst café = 1;\u2028debugger;\u2029"
	first := pipelineTestProgram(t, root, path, text).SourceFiles()[0]
	second := pipelineTestProgram(t, root, path, "const other = 2;").SourceFiles()[0]
	alreadyText := newTextSourceFile(text)
	fixes := []rule.RuleFix{{Range: core.NewTextRange(0, 0), Text: "// fix\n"}}
	suggestions := []rule.RuleSuggestion{{Message: rule.RuleMessage{Description: "suggestion"}, FixesArr: fixes}}
	original := rule.RuleDiagnostic{
		FilePath: path, SourceFile: first, RuleName: "native/check", Severity: rule.SeverityError,
		Message: rule.RuleMessage{Id: "message", Description: "message"}, Range: core.NewTextRange(0, len(text)),
		FixesPtr: &fixes, Suggestions: &suggestions, Origin: rule.DiagnosticOriginTypeScript, PreFormatted: true,
	}
	observation := ObservationResult{
		Native: NativeObservation{Diagnostics: []rule.RuleDiagnostic{
			original,
			{FilePath: path, SourceFile: second},
			{FilePath: path, SourceFile: alreadyText},
			{FilePath: path},
		}},
		pluginOutcome: EslintPluginDispatchOutcome{Diagnostics: []rule.RuleDiagnostic{{FilePath: path, SourceFile: first}}},
	}
	observation.detachDiagnosticSources()
	projected := observation.Native.Diagnostics[0]
	if _, retained := projected.SourceFile.(*ast.SourceFile); retained {
		t.Fatal("compiler source was not detached")
	}
	if projected.SourceFile != observation.pluginOutcome.Diagnostics[0].SourceFile || projected.SourceFile == observation.Native.Diagnostics[1].SourceFile {
		t.Fatal("source identity was replaced with path identity")
	}
	if observation.Native.Diagnostics[1].SourceFile.Text() != "const other = 2;" || observation.Native.Diagnostics[2].SourceFile != alreadyText || observation.Native.Diagnostics[3].SourceFile != nil {
		t.Fatal("projection changed a distinct, text-only, or missing source")
	}
	frame := projected.SourceFile
	projected.SourceFile = original.SourceFile
	if !reflect.DeepEqual(projected, original) {
		t.Fatal("projection changed diagnostic metadata, fixes, or suggestions")
	}
	observation.detachDiagnosticSources()
	if observation.Native.Diagnostics[0].SourceFile != frame {
		t.Fatal("repeated projection changed text source identity")
	}
	wantLines := first.ECMALineMap()
	var readers sync.WaitGroup
	for range 8 {
		readers.Go(func() {
			if frame.Text() != text || !slices.Equal(frame.ECMALineMap(), wantLines) {
				t.Error("source text or ECMAScript line map changed")
			}
			for pos := range text {
				line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(frame, pos)
				wantLine, wantColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(first, pos)
				if line != wantLine || column != wantColumn {
					t.Errorf("position %d = %d:%d, want %d:%d", pos, line, column, wantLine, wantColumn)
				}
			}
		})
	}
	readers.Wait()
}

func TestPipelineDetachesTypeCheckOnlyDiagnostics(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"source.ts": "const value: number = 'text';",
	})
	generation := Generation{Native: NativeGeneration{
		Programs: wrapTestPrograms(program), TypeCheck: true, SingleThreaded: true,
	}}
	result, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil), ObservationPolicy{}, nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := result.Observation.Native.Diagnostics
	if len(diagnostics) == 0 {
		t.Fatal("missing type-check-only diagnostics")
	}
	for _, diagnostic := range diagnostics {
		if _, retained := diagnostic.SourceFile.(*ast.SourceFile); retained {
			t.Fatal("type-check-only diagnostic retained a compiler AST")
		}
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
