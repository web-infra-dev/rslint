package linter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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

func TestDeferredRootsBoundConstructionThroughLintCompletion(t *testing.T) {
	for _, singleThreaded := range []bool{true, false} {
		t.Run(strconv.FormatBool(singleThreaded), func(t *testing.T) {
			limit := runtime.GOMAXPROCS(0)
			if singleThreaded {
				limit = 1
			}
			generation := pipelineDeferredGeneration(t, limit+3)
			generation.Native.SingleThreaded = singleThreaded
			build := generation.Native.DeferredRoots.Build
			var active, maximum, built atomic.Int32
			generation.Native.DeferredRoots.Build = func(ctx context.Context, name string) (*program.Program, error) {
				current := active.Add(1)
				built.Add(1)
				for previous := maximum.Load(); current > previous; previous = maximum.Load() {
					if maximum.CompareAndSwap(previous, current) {
						break
					}
				}
				return build(ctx, name)
			}
			entered := make(chan struct{}, limit+3)
			unblock := make(chan struct{})
			var unblockOnce sync.Once
			defer unblockOnce.Do(func() { close(unblock) })
			generation.Native.RulesForFile = func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "test", SupportsFileIsolation: true, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					defer active.Add(-1)
					entered <- struct{}{}
					<-unblock
					ctx.ReportRange(core.NewTextRange(0, 1), rule.RuleMessage{Description: "first"})
					ctx.ReportRange(core.NewTextRange(0, 1), rule.RuleMessage{Description: "second"})
					return nil
				}}}
			}
			done := make(chan error, 1)
			go func() {
				result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
				if err == nil {
					diagnostics := result.Observation.Native.Diagnostics
					if result.Observation.Native.Lint.LintedFileCount != int32(limit+3) || len(diagnostics) != 2*(limit+3) {
						err = fmt.Errorf("incomplete observation: %+v", result.Observation.Native)
					} else {
						for i, diagnostic := range diagnostics {
							if _, isAST := diagnostic.SourceFile.(*ast.SourceFile); isAST {
								err = errors.New("diagnostic retained AST")
							}
							want := "first"
							if i%2 != 0 {
								want = "second"
							}
							if diagnostic.Message.Description != want {
								err = errors.New("same-position diagnostic order changed")
							}
						}
					}
				}
				done <- err
			}()
			for range limit {
				select {
				case <-entered:
				case <-time.After(10 * time.Second):
					t.Fatal("workers did not reach lint")
				}
			}
			if got := built.Load(); got != int32(limit) {
				t.Fatalf("built %d roots while %d lint slots were occupied", got, limit)
			}
			unblockOnce.Do(func() { close(unblock) })
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if active.Load() != 0 || maximum.Load() > int32(limit) {
				t.Fatalf("active/max = %d/%d", active.Load(), maximum.Load())
			}
		})
	}
}

func TestDeferredRootsReleaseCompletedASTBeforeNextBuild(t *testing.T) {
	generation := pipelineDeferredGeneration(t, 2)
	generation.Native.SingleThreaded = true
	build := generation.Native.DeferredRoots.Build
	var previous weak.Pointer[ast.SourceFile]
	generation.Native.DeferredRoots.Build = func(ctx context.Context, name string) (*program.Program, error) {
		if name == generation.Native.DeferredRoots.FileNames[1] {
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				runtime.GC()
				if previous.Value() == nil {
					break
				}
				time.Sleep(time.Millisecond)
			}
			if previous.Value() != nil {
				return nil, errors.New("completed source AST retained at next build")
			}
		}
		p, err := build(ctx, name)
		if err == nil {
			previous = weak.Make(p.SourceFiles()[0])
		}
		return p, err
	}
	generation.Native.RulesForFile = func(*ast.SourceFile) []rule.ConfiguredRule {
		return []rule.ConfiguredRule{{Name: "test", SupportsFileIsolation: true, Run: func(ctx rule.RuleContext) rule.RuleListeners {
			// Listener closures and diagnostics both formerly retained the file.
			return rule.RuleListeners{ast.KindVariableDeclaration: func(node *ast.Node) { ctx.ReportNode(node, rule.RuleMessage{Description: "variable"}) }}
		}}}
	}
	result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Observation.Native.Diagnostics) != 2 {
		t.Fatal("diagnostics lost after releasing ASTs")
	}
}

func TestDeferredRootsJoinWorkersBeforeReleaseOnFailure(t *testing.T) {
	for _, failure := range []string{"error", "panic", "goexit", "cancel"} {
		t.Run(failure, func(t *testing.T) {
			limit := runtime.GOMAXPROCS(0)
			generation := pipelineDeferredGeneration(t, limit+2)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var active atomic.Int32
			var arrivals atomic.Int32
			allStarted := make(chan struct{})
			generation.Native.DeferredRoots.Build = func(workerCtx context.Context, path string) (*program.Program, error) {
				active.Add(1)
				defer active.Add(-1)
				if arrivals.Add(1) == int32(limit) {
					close(allStarted)
				}
				if path == generation.Native.DeferredRoots.FileNames[0] {
					select {
					case <-allStarted:
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					switch failure {
					case "panic":
						panic("source failure")
					case "goexit":
						runtime.Goexit()
					case "cancel":
						cancel()
						return nil, ctx.Err()
					default:
						return nil, errors.New("source failure")
					}
				}
				<-workerCtx.Done()
				return nil, workerCtx.Err()
			}
			var releases int
			var recovered any
			var result PipelineResult
			var err error
			func() {
				defer func() { recovered = recover() }()
				result, err = RunPipeline(ctx, NewLintRequest(pipelineTestProvider(generation, func() {
					releases++
					if active.Load() != 0 {
						t.Error("generation released with active builders")
					}
				}), ObservationPolicy{}, nil))
			}()
			if releases != 1 || active.Load() != 0 {
				t.Fatalf("releases/active = %d/%d", releases, active.Load())
			}
			if failure == "panic" || failure == "goexit" {
				if recovered == nil {
					t.Fatal("worker abnormal exit was lost")
				}
			} else if err == nil {
				t.Fatal("worker failure was lost")
			}
			if result.Observation.Native.Lint != nil {
				t.Fatal("published a partial successful observation")
			}
			if got := arrivals.Load(); got != int32(limit) {
				t.Fatalf("claimed additional files after failure: %d", got)
			}
		})
	}
}

func TestDeferredRootsAbortDuringRuleResolutionAndTraversal(t *testing.T) {
	for _, stage := range []string{"resolve", "initialize", "visit"} {
		for _, failure := range []string{"panic", "goexit", "cancel"} {
			t.Run(stage+"/"+failure, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				generation := pipelineDeferredGeneration(t, 3)
				generation.Native.SingleThreaded = true
				build := generation.Native.DeferredRoots.Build
				var builds, releases int
				generation.Native.DeferredRoots.Build = func(ctx context.Context, name string) (*program.Program, error) {
					builds++
					return build(ctx, name)
				}
				fail := func() {
					switch failure {
					case "panic":
						panic("rule failure")
					case "goexit":
						runtime.Goexit()
					case "cancel":
						cancel()
					}
				}
				generation.Native.RulesForFile = func(*ast.SourceFile) []rule.ConfiguredRule {
					if stage == "resolve" {
						fail()
					}
					return []rule.ConfiguredRule{{Name: "failure", SupportsFileIsolation: true, Run: func(rule.RuleContext) rule.RuleListeners {
						if stage == "initialize" {
							fail()
						}
						return rule.RuleListeners{ast.KindVariableDeclaration: func(*ast.Node) {
							if stage == "visit" {
								fail()
							}
						}}
					}}}
				}
				var recovered any
				var result PipelineResult
				var err error
				func() {
					defer func() { recovered = recover() }()
					result, err = RunPipeline(ctx, NewLintRequest(pipelineTestProvider(generation, func() { releases++ }), ObservationPolicy{}, nil))
				}()
				if builds != 1 || releases != 1 || result.Observation.Native.Lint != nil {
					t.Fatalf("builds=%d releases=%d result=%+v", builds, releases, result)
				}
				if failure == "cancel" {
					if !errors.Is(err, context.Canceled) {
						t.Fatalf("cancellation lost: %v", err)
					}
				} else if recovered == nil {
					t.Fatal("rule abnormal exit was lost")
				}
			})
		}
	}
}
