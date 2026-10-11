package linter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestPipelineReclaimsCompletedProgramWhileOtherConsumersRun(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined} {
		for _, singleThreaded := range []bool{false, true} {
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAll} {
				t.Run(fmt.Sprintf("mode-%d-serial-%t-demand-%d", mode, singleThreaded, demand), func(t *testing.T) {
					var completedProgram weak.Pointer[compiler.Program]
					var completedSource weak.Pointer[ast.SourceFile]
					otherStarted := make(chan struct{})
					pluginStarted := make(chan struct{})
					resume := make(chan struct{})
					var resumeOnce sync.Once
					unblock := func() { resumeOnce.Do(func() { close(resume) }) }
					defer unblock()
					var releases atomic.Int32
					provider := GenerationProviderFunc(func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error) {
						completed, completedPaths := createTestProgramWithFilesAndCompilerOptions(t,
							map[string]string{"completed.ts": "const value = 1;"}, `{"noLib":true}`)
						other, otherPaths := createTestProgramWithFilesAndCompilerOptions(t,
							map[string]string{"other.ts": "const other = 1;"}, `{"noLib":true}`)
						completedProgram = weak.Make(completed)
						completedSource = weak.Make(completed.GetSourceFile(completedPaths["completed.ts"]))
						// A plugin rule may carry Go-only fields in the prepared plan.
						// Its wire projection must never keep this native closure.
						pluginRun := func(rule.RuleContext) rule.RuleListeners {
							runtime.KeepAlive(completed)
							return nil
						}
						return Generation{
							Native: NativeGeneration{
								// Serial work groups execute in reverse queue order.
								Programs:         []*program.Program{program.NewFromCompiler(other), program.NewFromCompiler(completed)},
								TargetsByProgram: [][]string{{otherPaths["other.ts"]}, {completedPaths["completed.ts"]}},
								SingleThreaded:   singleThreaded,
								RulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
									if strings.HasSuffix(file.FileName(), "/other.ts") {
										return []rule.ConfiguredRule{{Name: "native/wait", Run: func(rule.RuleContext) rule.RuleListeners {
											close(otherStarted)
											<-resume
											return nil
										}}}
									}
									return []rule.ConfiguredRule{
										{Name: "native/check", Run: func(ctx rule.RuleContext) rule.RuleListeners {
											span := core.NewTextRange(6, 11)
											ctx.ReportRangeWithDeferredFixes(span, rule.RuleMessage{Description: "native"}, func() []rule.RuleFix {
												return []rule.RuleFix{rule.RuleFixReplaceRange(span, "replacement")}
											})
											return nil
										}},
										{Name: "plugin/check", IsEslintPluginRule: true, Run: pluginRun},
									}
								},
							},
							Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
								// Even an adapter reader that captures the compiler is
								// preparation-only when no source changes are planned.
								runtime.KeepAlive(completed)
								return source.Text(), nil
							}},
							Plugin: &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig { return EslintPluginFileConfig{} }},
						}, func() { releases.Add(1) }, nil
					})
					type outcome struct {
						result PipelineResult
						err    error
					}
					done := make(chan outcome, 1)
					t.Cleanup(func() {
						unblock()
						<-done
					})
					go func() {
						defer close(done)
						result, err := RunPipeline(context.Background(), NewLintRequest(provider, ObservationPolicy{
							Plugin: mode,
							Demand: ArtifactDemand{Native: demand, Plugin: demand},
						}, func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
							close(pluginStarted)
							<-resume
							return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
								FilePath:    request.Files[0].Path,
								Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "plugin", StartPos: 6, EndPos: 11}},
							}}}, nil
						}))
						done <- outcome{result, err}
					}()
					startedConsumers := []<-chan struct{}{otherStarted}
					if mode == PluginConcurrentJoined {
						startedConsumers = append(startedConsumers, pluginStarted)
					}
					for _, started := range startedConsumers {
						select {
						case <-started:
						case got := <-done:
							t.Fatalf("pipeline ended before the blocking consumers started: %v", got.err)
						}
					}
					deadline := time.Now().Add(5 * time.Second)
					for completedProgram.Value() != nil || completedSource.Value() != nil {
						if time.Now().After(deadline) {
							t.Fatal("completed project remains reachable while another project and the plugin are blocked")
						}
						runtime.GC()
						runtime.Gosched()
					}
					if releases.Load() != 0 {
						t.Fatal("producer finalizer ran before all generation consumers joined")
					}
					unblock()
					got := <-done
					if got.err != nil || releases.Load() != 1 {
						t.Fatalf("pipeline error/releases = %v/%d", got.err, releases.Load())
					}
					if got.result.Observation.Native.Lint.LintedFileCount != 2 || len(got.result.Observation.Native.Diagnostics) != 1 {
						t.Fatalf("native result = %+v", got.result.Observation.Native)
					}
					plugin, joined := got.result.Observation.JoinedPluginOutcome()
					if !joined || len(plugin.Diagnostics) != 1 ||
						(mode == PluginConcurrentJoined && plugin.Diagnostics[0].SourceFile != got.result.Observation.Native.Diagnostics[0].SourceFile) {
						t.Fatal("native/plugin diagnostics lost their exact shared source frame")
					}
				})
			}
		}
	}
}

func TestConcurrentPipelineJoinsPluginBatchesBeforeReleaseOnPanic(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	panicPath := tspath.ResolvePath(root, "panic.ts")
	waitPath := tspath.ResolvePath(root, "wait.ts")
	waitStarted := make(chan struct{})
	waitStopped := make(chan struct{})
	panicTriggered := make(chan struct{})
	resume := make(chan struct{})
	var resumeOnce sync.Once
	unblock := func() { resumeOnce.Do(func() { close(resume) }) }
	generation := Generation{
		Native: NativeGeneration{
			Programs:         []*program.Program{pipelineTestProgram(t, root, panicPath, "a"), pipelineTestProgram(t, root, waitPath, "b")},
			TargetsByProgram: [][]string{{panicPath}, {waitPath}},
			SingleThreaded:   true,
			RulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
				name := "plugin/panic"
				if file.FileName() == waitPath {
					name = "plugin/wait"
				}
				return []rule.ConfiguredRule{{Name: name, IsEslintPluginRule: true}}
			},
		},
		Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) { return source.Text(), nil }},
		Plugin: &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig { return EslintPluginFileConfig{} }},
	}
	var releases atomic.Int32
	const marker = "plugin failed"
	type outcome struct {
		result PipelineResult
		err    error
	}
	done := make(chan outcome, 1)
	t.Cleanup(func() { unblock(); <-done })
	go func() {
		defer close(done)
		result, err := RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, func() {
				select {
				case <-waitStopped:
				default:
					t.Error("released generation before sibling plugin batch joined")
				}
				releases.Add(1)
			}),
			ObservationPolicy{PluginFailure: PluginKeepPartialWithSynthetic},
			func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
				if _, waiting := request.Rules["plugin/wait"]; waiting {
					close(waitStarted)
					<-resume
					close(waitStopped)
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
						FilePath:    request.Files[0].Path,
						Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/wait", Message: "surviving batch"}},
					}}}, nil
				}
				<-waitStarted
				close(panicTriggered)
				panic(marker)
			},
		))
		done <- outcome{result, err}
	}()
	select {
	case <-panicTriggered:
	case got := <-done:
		t.Fatalf("pipeline ended before the panic batch started: %v", got.err)
	}
	select {
	case got := <-done:
		t.Fatalf("pipeline returned before its surviving plugin batch joined: %v", got.err)
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	got := <-done
	plugin, joined := got.result.Observation.JoinedPluginOutcome()
	if got.err != nil || !joined || plugin.DispatchError == nil || !strings.Contains(plugin.DispatchError.Error(), marker) ||
		len(plugin.Diagnostics) != 2 || releases.Load() != 1 {
		t.Fatalf("pipeline/plugin/releases = %v/%+v/%d, want recovered plugin error, surviving diagnostics, release once", got.err, plugin, releases.Load())
	}
}

func TestConcurrentPipelineJoinsNativeShardsBeforeReleaseOnPanic(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)
	waitStarted := make(chan struct{})
	waitStopped := make(chan struct{})
	panicTriggered := make(chan struct{})
	resume := make(chan struct{})
	var resumeOnce sync.Once
	unblock := func() { resumeOnce.Do(func() { close(resume) }) }
	const marker = "native shard failed"
	run := func(ctx rule.RuleContext) rule.RuleListeners {
		switch {
		case strings.HasSuffix(ctx.SourceFile.FileName(), "/file-000.ts"):
			<-waitStarted
			close(panicTriggered)
			panic(marker)
		case strings.HasSuffix(ctx.SourceFile.FileName(), "/file-128.ts"):
			close(waitStarted)
			<-resume
			if !ctx.Program().IsValid() || ctx.SourceFile.Text() != "const value = 1;" {
				t.Error("sibling shard lost its source generation")
			}
			ctx.ReportRange(core.NewTextRange(0, 5), rule.RuleMessage{Description: "sibling still reporting"})
			close(waitStopped)
		}
		return nil
	}
	options := checkerFreeExecutionTestOptions(t, false, run)
	plan := options.LintPlan.programs[0]
	paths := make([]string, len(plan.files))
	for index, file := range plan.files {
		paths[index] = file.file.FileName()
	}
	generation := Generation{Native: NativeGeneration{
		Programs:         []*program.Program{plan.program},
		TargetsByProgram: [][]string{paths},
		RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: "native/panic", Run: run}}
		},
	}}
	var releases atomic.Int32
	done := make(chan any, 1)
	t.Cleanup(func() { unblock(); <-done })
	go func() {
		defer close(done)
		defer func() { done <- recover() }()
		_, _ = RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, func() {
			select {
			case <-waitStopped:
			default:
				t.Error("released generation while a native sibling shard still borrowed it")
			}
			releases.Add(1)
		}), ObservationPolicy{}, nil))
	}()
	select {
	case <-panicTriggered:
	case value := <-done:
		t.Fatalf("pipeline ended before the panic shard started: %v", value)
	}
	select {
	case value := <-done:
		t.Fatalf("pipeline propagated panic before its native sibling joined: %v", value)
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	if value := <-done; value != marker || releases.Load() != 1 {
		t.Fatalf("panic/releases = %v/%d, want original panic/1", value, releases.Load())
	}
}

func TestPipelineSharedCompilerRemainsValidUntilLastProjectCompletes(t *testing.T) {
	var firstFacade weak.Pointer[program.Program]
	var sharedCompiler weak.Pointer[compiler.Program]
	blocked := make(chan struct{})
	resume := make(chan struct{})
	var resumeOnce sync.Once
	unblock := func() { resumeOnce.Do(func() { close(resume) }) }
	defer unblock()
	provider := GenerationProviderFunc(func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error) {
		raw, paths := createTestProgramWithFilesAndCompilerOptions(t,
			map[string]string{"first.ts": "const first = 1;", "second.ts": "const second = 2;"}, `{"noLib":true}`)
		first := program.NewFromCompiler(raw)
		second := program.NewFromCompiler(raw)
		firstFacade = weak.Make(first)
		sharedCompiler = weak.Make(raw)
		return Generation{Native: NativeGeneration{
			Programs:         []*program.Program{second, first},
			TargetsByProgram: [][]string{{paths["second.ts"]}, {paths["first.ts"]}},
			SingleThreaded:   true,
			RulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "native/shared", Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if strings.HasSuffix(ctx.SourceFile.FileName(), "/second.ts") {
						close(blocked)
						<-resume
						if !ctx.Program().IsValid() || len(ctx.Program().SourceFiles()) != 2 || ctx.TypeChecker == nil {
							t.Error("ending the first task invalidated the shared compiler generation")
						}
					}
					ctx.ReportRange(core.NewTextRange(0, 5), rule.RuleMessage{Description: "valid source"})
					return nil
				}}}
			},
		}}, nil, nil
	})
	done := make(chan error, 1)
	t.Cleanup(func() { unblock(); <-done })
	go func() {
		defer close(done)
		_, err := RunPipeline(context.Background(), NewLintRequest(provider, ObservationPolicy{}, nil))
		done <- err
	}()
	select {
	case <-blocked:
	case err := <-done:
		t.Fatalf("pipeline finished before the second project blocked: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for firstFacade.Value() != nil {
		if time.Now().After(deadline) {
			t.Fatal("completed facade remains reachable")
		}
		runtime.GC()
		runtime.Gosched()
	}
	if sharedCompiler.Value() == nil {
		t.Fatal("shared compiler was collected while the second project still borrowed it")
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

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
