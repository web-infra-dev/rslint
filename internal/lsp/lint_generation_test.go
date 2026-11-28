package lsp

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/jsonrpc"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil/estree"
	"github.com/web-infra-dev/rslint/internal/typeservice"
)

type pluginTypeRequestReader struct{ messages chan *lsproto.Message }

func (r pluginTypeRequestReader) Read() (*lsproto.Message, error) {
	message, ok := <-r.messages
	if !ok {
		return nil, io.EOF
	}
	return message, nil
}

func TestLSPTypedEnrichmentReadsItsLeasedGeneration(t *testing.T) {
	directory, file, raw, fsys := programOver(t, "export const value = 1;\n")
	generation := newLintGeneration(raw, sourceFileForPath(raw, file, fsys), lspConfigTarget(file, directory, fsys), directory, true,
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}, &linter.EslintPluginFileConfig{TypeServices: true}, nil)
	released := 0
	presentation := &capturedProgressiveDiagnostics{}
	_, err := linter.RunPipeline(t.Context(), linter.NewProgressiveLintRequest(linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
		return generation, func() { released++ }, nil
	}), linter.ArtifactDemand{}, presentation))
	if err != nil {
		t.Fatal(err)
	}
	if released != 0 || presentation.run == nil {
		t.Fatalf("typed enrichment lost lease: %d", released)
	}
	reader := pluginTypeRequestReader{messages: make(chan *lsproto.Message, 1)}
	server := &Server{outgoingQueue: make(chan *lsproto.Message, 1), requestQueue: make(chan *lsproto.RequestMessage, 1), initializeParams: &lsproto.InitializeParams{}, r: reader}
	readCtx, cancelRead := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancelRead()
	readDone := make(chan error, 1)
	go func() { readDone <- server.readLoop(readCtx) }()
	defer func() { close(reader.messages); <-readDone }()
	var request *lsproto.RequestMessage
	outcome, err := presentation.run.Run(t.Context(), linter.EslintPluginDispatchFunc(func(ctx context.Context, input linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if len(input.Files) == 0 {
			return &linter.EslintPluginLintResult{}, nil
		}
		service := input.Files[0].TypeServices
		if service == nil {
			t.Fatal("missing typed document binding")
		}
		request = &lsproto.RequestMessage{ID: jsonrpc.NewIDString("typed"), Method: methodTypeService, Params: linter.EslintPluginTypeRequest{Session: service.Session, Generation: service.Binding.Generation, Source: &file}}
		// No document dispatcher drains requestQueue: a blocking fix-all is
		// waiting for its reverse plugin request at this exact boundary.
		reader.messages <- request.Message()
		var response *lsproto.Message
		select {
		case response = <-server.outgoingQueue:
		case <-readCtx.Done():
			t.Fatal("nested typed read was blocked behind serialized document work")
		}
		if response.Kind != jsonrpc.MessageKindResponse {
			t.Fatal("type service did not answer on the document connection")
		}
		result, ok := response.AsResponse().Result.(*linter.EslintPluginTypeResult)
		if !ok || result.Source == nil || result.Source.FileName != file || len(result.Source.Tree) == 0 {
			t.Fatal("wrong source generation reply")
		}
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: input.Files[0].Path}}}, nil
	}))
	if err != nil || outcome.DispatchError != nil {
		t.Fatalf("typed enrichment: %v/%v", err, outcome.DispatchError)
	}
	if released != 1 {
		t.Fatalf("producer release count %d", released)
	}
	reader.messages <- request.Message()
	var reply *lsproto.Message
	select {
	case reply = <-server.outgoingQueue:
	case <-readCtx.Done():
		t.Fatal("retired session did not answer")
	}
	status := reply.AsResponse().Result.(*linter.EslintPluginTypeResult)
	if status.Status != "closed" {
		t.Fatal("retired document generation accepted a type query")
	}
}

func TestCancelTypedEnrichmentFinalizesBeforeNextSelection(t *testing.T) {
	directory, file, raw, fsys := programOver(t, "export const value = 1;")
	started := make(chan struct{})
	server := &Server{generationReleaseReady: make(chan struct{}, 1), inflightPluginDispatch: make(map[lsproto.DocumentUri]*pluginDispatchHandle), pluginResultCh: make(chan pluginLintResult, 1)}
	server.eslintPluginDispatch = func(ctx context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if len(request.Files) == 0 {
			return &linter.EslintPluginLintResult{}, nil
		}
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	generation := newLintGeneration(raw, sourceFileForPath(raw, file, fsys), lspConfigTarget(file, directory, fsys), directory, true,
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}, &linter.EslintPluginFileConfig{TypeServices: true}, nil)
	generation.Plugin.ReleaseOnOwner = server.queueGenerationRelease
	finalized := 0
	presentation := &capturedProgressiveDiagnostics{}
	_, err := linter.RunPipeline(t.Context(), linter.NewProgressiveLintRequest(linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
		return generation, func() { finalized++ }, nil
	}), linter.ArtifactDemand{}, presentation))
	if err != nil {
		t.Fatal(err)
	}
	uri := documentURIFromPath(file)
	server.startDiagnosticEnrichment(t.Context(), uri, 1, "config", presentation.run)
	<-started
	if finalized != 0 {
		t.Fatal("typed generation finalized while a worker could still query")
	}
	server.cancelInflightPluginDispatch(uri)
	if finalized != 1 {
		t.Fatal("cancel/join did not finalize old generation before the next selection")
	}
	// The next selection must not inherit a pending callback for its old owner.
	server.drainGenerationReleases()
	if finalized != 1 {
		t.Fatal("old generation finalized twice")
	}
}

func TestLSPSourceOnlyFallbackDoesNotAdmitPluginTypeServices(t *testing.T) {
	directory, file, raw, fsys := programOver(t, "export const value = 1;")
	config := &linter.EslintPluginFileConfig{TypeServices: true}
	generation := newLintGeneration(raw, sourceFileForPath(raw, file, fsys), lspConfigTarget(file, directory, fsys), directory, false,
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}, config, nil)
	if !config.TypeServices {
		t.Fatal("generation mutated the frozen effective config")
	}
	_, err := linter.RunPipeline(t.Context(), linter.NewLintRequest(linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
		return generation, nil, nil
	}), linter.ObservationPolicy{}, linter.EslintPluginDispatchFunc(func(ctx context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if len(request.Files) != 1 || request.Files[0].TypeServices == nil || !request.Files[0].TypeServices.SyntaxOnly {
			t.Fatal("source-only fallback borrowed a checker from its compiler container")
		}
		service := request.Files[0].TypeServices
		opened, err := linter.QueryEslintPluginTypes(ctx, linter.EslintPluginTypeRequest{
			Session:    service.Session,
			Generation: service.Binding.Generation,
			Open:       true,
			Projection: &service.Binding.SourceFileName,
		})
		if err != nil || opened.Descriptor == nil || !opened.Descriptor.SyntaxOnly {
			t.Fatalf("source-only fallback did not expose syntax-only descriptor: %v", err)
		}
		if _, err := linter.QueryEslintPluginTypes(ctx, linter.EslintPluginTypeRequest{
			Session:    service.Session,
			Generation: service.Binding.Generation,
			Queries:    []typeservice.Query{{Scope: service.Binding.Scope, Request: &typeservice.CheckerRequest{Op: "void"}}},
		}); err == nil {
			t.Fatal("source-only fallback allowed a semantic checker query")
		}
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: request.Files[0].Path}}}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
}

func runLSPGenerationForTest(
	ctx context.Context,
	generation linter.Generation,
	release linter.ReleaseFunc,
	demand linter.ArtifactDemand,
) (linter.PipelineResult, error) {
	return linter.RunPipeline(ctx, linter.NewLintRequest(
		linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
			return generation, release, nil
		}),
		linter.ObservationPolicy{
			Demand:        demand,
			Plugin:        linter.PluginAfterNativeJoined,
			PluginFailure: linter.PluginDiscardOnFailure,
		},
		linter.EslintPluginDispatchFunc(func(_ context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			results := make([]linter.EslintPluginFileResult, len(request.Files))
			for index, file := range request.Files {
				results[index].FilePath = file.Path
			}
			return &linter.EslintPluginLintResult{Results: results}, nil
		}),
	))
}

func TestLSPGenerationReleaseLifecycle(t *testing.T) {
	t.Run("normal execution", func(t *testing.T) {
		directory, file, program, fsys := programOver(t, "const value = 1;\n")
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		ruleRan := false
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name:     "test/release",
				Severity: rule.SeverityWarning,
				Run: func(rule.RuleContext) rule.RuleListeners {
					if releases != 0 {
						t.Fatal("generation finalized before rule execution completed")
					}
					ruleRan = true
					return nil
				},
			}},
			nil,
			nil,
		)
		_, err := runLSPGenerationForTest(
			context.Background(), generation, func() { releases++ }, linter.ArtifactDemand{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !ruleRan || releases != 1 {
			t.Fatalf("rule ran/release calls = %v/%d, want true/1", ruleRan, releases)
		}
	})

	t.Run("syntax error", func(t *testing.T) {
		directory, file, program, fsys := programOver(t, "const value = ;\n")
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		ruleRan := false
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name: "test/must-not-run",
				Run: func(rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					return nil
				},
			}},
			nil,
			nil,
		)
		result, err := runLSPGenerationForTest(
			context.Background(), generation, func() { releases++ }, linter.ArtifactDemand{Native: rule.EditDemandAll},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Observation.Native.HasTargetSyntaxErrors || ruleRan || releases != 1 {
			t.Fatalf("syntax/rule/release = %v/%v/%d, want true/false/1", result.Observation.Native.HasTargetSyntaxErrors, ruleRan, releases)
		}
	})

	t.Run("plugin work detached before release", func(t *testing.T) {
		const content = "const value = 1;\n"
		directory, file, program, fsys := programOver(t, content)
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name:               "test/plugin",
				Severity:           rule.SeverityWarning,
				IsEslintPluginRule: true,
			}},
			&linter.EslintPluginFileConfig{ConfigKey: "frozen-config"},
			nil,
		)
		presentation := &capturedProgressiveDiagnostics{}
		_, err := linter.RunPipeline(context.Background(), linter.NewProgressiveLintRequest(
			linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
				return generation, func() { releases++ }, nil
			}),
			linter.ArtifactDemand{Plugin: rule.EditDemandAll},
			presentation,
		))
		if err != nil {
			t.Fatal(err)
		}
		if releases != 1 || presentation.run == nil {
			t.Fatalf("release/plugin run = %d/%v, want 1/non-nil", releases, presentation.run != nil)
		}
		var request linter.EslintPluginLintRequest
		outcome, err := presentation.run.Run(context.Background(), linter.EslintPluginDispatchFunc(func(_ context.Context, got linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			request = got
			return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: got.Files[0].Path}}}, nil
		}))
		if err != nil || outcome.DispatchError != nil {
			t.Fatalf("plugin work errors = %v/%v", err, outcome.DispatchError)
		}
		if len(request.Files) != 1 || estree.SourceText(t, request.Files[0].Frame) != content {
			t.Fatalf("detached plugin request = %+v", request.Files)
		}
	})

	t.Run("canceled before acquisition", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		acquisitions := 0
		result, err := linter.RunPipeline(ctx, linter.NewLintRequest(
			linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
				acquisitions++
				return linter.Generation{}, nil, nil
			}),
			linter.ObservationPolicy{},
			nil,
		))
		_ = result
		if !errors.Is(err, context.Canceled) || acquisitions != 0 {
			t.Fatalf("error/acquisitions = %v/%d, want context.Canceled/0", err, acquisitions)
		}
	})
}

func TestDocumentGenerationProviderFinalizesBeforePublication(t *testing.T) {
	fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
	server := fixture.server
	server.backgroundCtx = context.Background()
	if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
		t.Fatal(err)
	}
	server.lintPrograms = fixture.store
	snapshot := documentLintSnapshotForTest(
		server,
		fixture.sourceURI,
		config.RslintConfig{},
		filepath.Dir(fixture.configPath),
		false,
		[]string{fixture.configPath},
	)
	releases := 0
	provider := &documentGenerationProvider{
		server:   server,
		uri:      fixture.sourceURI,
		snapshot: snapshot,
		requestPrograms: func(
			context.Context,
			lsproto.DocumentUri,
			target.File,
		) (lintProjectLoaders, linter.ReleaseFunc) {
			return lintProjectLoaders{}, func() { releases++ }
		},
		buildGeneration: func(
			*compiler.Program,
			*ast.SourceFile,
			target.File,
			string,
			bool,
			documentLintSnapshot,
		) linter.Generation {
			panic("generation assembly failed")
		},
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _, _ = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
	}()
	if recovered == nil {
		t.Fatal("expected generation assembly panic")
	}
	if releases != 1 {
		t.Fatalf("resident Program finalizer calls = %d, want 1", releases)
	}
}

func TestLSPProjectConfigErrorClassification(t *testing.T) {
	const content = "export const value = 1;\n"
	directory := tspath.NormalizePath(t.TempDir())
	sourcePath := tspath.ResolvePath(directory, "source.ts")
	configPath := tspath.ResolvePath(directory, "tsconfig.json")
	for path, text := range map[string]string{
		sourcePath: content,
		configPath: `{"files":["source.ts"]}`,
	} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	uri := documentURIFromPath(sourcePath)
	server.documents[uri] = content
	snapshot := documentLintSnapshotForTest(server, uri, nil, directory, false, []string{"./missing.json"})
	if snapshot.configResolved || snapshot.projectPolicyError != nil {
		t.Fatal("expected unresolved project settings")
	}
	for name, provider := range map[string]linter.GenerationProvider{
		"diagnostics": &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot},
		"fix all":     server.newSpeculativeGenerationProvider(uri, content, snapshot),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(provider, linter.ObservationPolicy{}, nil))
			var projectErr *documentProjectConfigError
			if !errors.As(err, &projectErr) || !strings.Contains(err.Error(), "missing.json") {
				t.Fatalf("lazy project error = %v, want document project configuration error", err)
			}
			if !errors.Is(err, projectErr.cause) {
				t.Fatal("project configuration wrapper must preserve its cause")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			plan, err := provider.PlanGeneration(ctx, linter.SourceSnapshot{})
			if err != nil {
				t.Fatal(err)
			}
			_, release, err := plan.Loads[0].Acquire(ctx)
			if !errors.Is(err, context.Canceled) || errors.As(err, &projectErr) || release != nil {
				t.Fatalf("canceled acquisition = %v, release = %v, want unclassified cancellation", err, release != nil)
			}
		})
	}

	t.Run("loader failure", func(t *testing.T) {
		loadErr := errors.New("project metadata loader failed")
		loads, releases := 0, 0
		provider := &documentGenerationProvider{
			server: server,
			uri:    uri,
			snapshot: documentLintSnapshotForTest(
				server, uri, nil, directory, false, []string{configPath},
			),
			requestPrograms: func(context.Context, lsproto.DocumentUri, target.File) (lintProjectLoaders, linter.ReleaseFunc) {
				return lintProjectLoaders{
					metadata: func(string) (*lintProjectMetadata, bool, error) {
						loads++
						return nil, false, loadErr
					},
				}, func() { releases++ }
			},
		}
		_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(provider, linter.ObservationPolicy{}, nil))
		var projectErr *documentProjectConfigError
		if !errors.Is(err, loadErr) || errors.As(err, &projectErr) || loads != 1 || releases != 1 {
			t.Fatalf("error/loads/releases = %v/%d/%d, want unclassified loader error/1/1", err, loads, releases)
		}
	})
}

func TestLSPTypeServiceJSONv2OmitsZeroFields(t *testing.T) {
	raw, err := json.Marshal(&linter.EslintPluginTypeResult{
		Descriptor: &typeservice.Descriptor{
			Generation:       1,
			CurrentDirectory: "/repo",
			Sources: []typeservice.Source{{
				FileName: "/repo/a.ts",
				Path:     "/repo/a.ts",
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(raw)
	if strings.Contains(encoded, "impliedNodeFormat") {
		t.Fatalf("JSON v2 marshaled zero ImpliedNodeFormat: %s", encoded)
	}
	if strings.Contains(encoded, "syntaxOnly") {
		t.Fatalf("JSON v2 marshaled false SyntaxOnly: %s", encoded)
	}
	if !isBlockingMethod(lsproto.MethodShutdown) {
		t.Fatal("shutdown must run serially on dispatchLoop")
	}
}
