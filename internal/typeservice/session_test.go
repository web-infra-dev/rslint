package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildTestProgram(t *testing.T, files map[string]string, roots []string) (*program.Program, *compiler.Program, string) {
	t.Helper()
	const root = "/type-service-test"
	overlay := make(map[string]string, len(files))
	for name, content := range files {
		overlay[tspath.ResolvePath(root, name)] = content
	}
	resolvedRoots := make([]string, len(roots))
	for i, name := range roots {
		resolvedRoots[i] = tspath.ResolvePath(root, name)
	}
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), overlay)
	raw, err := utils.CreateProgramFromOptions(
		false,
		&core.CompilerOptions{
			NoLib:   core.TSTrue,
			Strict:  core.TSTrue,
			Module:  core.ModuleKindESNext,
			AllowJs: core.TSTrue,
		},
		resolvedRoots,
		utils.CreateCompilerHost(root, fs),
	)
	if err != nil {
		t.Fatal(err)
	}
	return program.NewFromCompiler(raw), raw, root
}

// A native shard must remain available between JS queries. A session that
// retains an exclusive lease would deadlock this independently owned borrower.
func TestSessionReleasesCheckerLeaseAndPreservesEligibility(t *testing.T) {
	p, raw, root := buildTestProgram(t, map[string]string{"input.ts": "export const value = 1;"}, []string{"input.ts"})
	name := tspath.ResolvePath(root, "input.ts")
	file := p.GetSourceFile(name)
	if file == nil {
		t.Fatal("fixture source missing")
	}
	if _, err := New(t.Context(), p, nil); err == nil {
		t.Fatal("checker capability admitted a source without eligibility")
	}
	sourceOnly, err := program.NewFromBoundSources(raw, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New(t.Context(), sourceOnly, []*ast.SourceFile{file}); err == nil {
		t.Fatal("source-only Program admitted typed rules")
	}
	for _, owner := range []*program.Program{p, sourceOnly} {
		syntax, err := NewSyntaxOnly(t.Context(), owner, []*ast.SourceFile{file})
		if err != nil {
			t.Fatal(err)
		}
		binding, err := syntax.BindTarget(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := syntax.Query(t.Context(), Query{Scope: binding.Scope, Request: &NodeRequest{Op: "nodeRecord", Node: NodeRef{File: name, Index: 1}}}); err != nil {
			t.Fatal(err)
		}
		if _, err := syntax.Query(t.Context(), Query{Scope: binding.Scope, Request: &CheckerRequest{Op: "void"}}); err == nil {
			t.Fatal("syntax-only session acquired a checker")
		}
		syntax.Close()
	}
	session, err := New(t.Context(), p, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	binding, err := session.BindTarget(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Query(t.Context(), Query{Scope: binding.Scope, Request: &CheckerRequest{Op: "void"}}); err != nil {
		t.Fatal(err)
	}
	acquired := make(chan struct{})
	go func() {
		_, release := p.TypeCheckerForFileExclusive(core.WithCheckerLifetime(t.Context(), core.CheckerLifetimeAPI), file)
		release()
		close(acquired)
	}()
	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("type service retained the checker after the query")
	}
}

// Verify Session itself synchronizes concurrent QueryBatch, Descriptor, Source,
// and Close calls cleanly under the Go race detector without external locking.
func TestSessionConcurrentQueriesAndClose(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{
		"a.ts": "export const alpha: number = 1; export function add(x: number, y: number) { return x + y; }",
		"b.ts": "import { alpha } from './a'; export const beta = alpha + 2;",
	}, []string{"a.ts", "b.ts"})
	fileA := p.GetSourceFile(tspath.ResolvePath(root, "a.ts"))
	fileB := p.GetSourceFile(tspath.ResolvePath(root, "b.ts"))
	session, err := New(t.Context(), p, []*ast.SourceFile{fileA, fileB})
	if err != nil {
		t.Fatal(err)
	}
	binding, err := session.BindTarget(fileA)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				_, qErr := session.QueryBatch(t.Context(), []Query{
					{Scope: binding.Scope, Request: &NodeRequest{Op: "nodeRecord", Node: NodeRef{File: binding.SourceFileName, Index: 1}}},
					{Scope: binding.Scope, Request: &CheckerRequest{Op: "stringType"}},
					{Scope: binding.Scope, Request: &ProgramRequest{Op: "rootFileNames"}},
				})
				if qErr != nil && !errors.Is(qErr, ErrClosed) {
					t.Errorf("unexpected query error: %v", qErr)
					return
				}
				_, sErr := session.Source(binding.Generation, binding.SourceFileName)
				if sErr != nil && !errors.Is(sErr, ErrClosed) {
					t.Errorf("unexpected source error: %v", sErr)
					return
				}
			}
		}()
	}

	// Close mid-flight while worker goroutines are querying concurrently.
	time.Sleep(5 * time.Millisecond)
	session.Close()
	wg.Wait()

	if _, err := session.Descriptor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed after Close, got %v", err)
	}
	if _, err := session.BindTarget(fileA); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed from BindTarget, got %v", err)
	}
	if _, err := session.Query(t.Context(), Query{Scope: binding.Scope, Request: &CheckerRequest{Op: "void"}}); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed from Query, got %v", err)
	}
}

func TestSessionValidationAndProtocolErrors(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{"input.ts": "export const x = 1;"}, []string{"input.ts"})
	name := tspath.ResolvePath(root, "input.ts")
	file := p.GetSourceFile(name)
	session, err := New(t.Context(), p, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	binding, err := session.BindTarget(file)
	if err != nil {
		t.Fatal(err)
	}

	// Oversized batch rejection.
	tooMany := make([]Query, 257)
	for i := range tooMany {
		tooMany[i] = Query{Scope: binding.Scope, Request: &CheckerRequest{Op: "void"}}
	}
	if _, err := session.QueryBatch(t.Context(), tooMany); err == nil {
		t.Fatal("expected error for >256 query batch")
	}

	// Foreign generation rejection.
	if _, err := session.Query(t.Context(), Query{Scope: Scope{Generation: binding.Generation + 99}, Request: &CheckerRequest{Op: "void"}}); err == nil {
		t.Fatal("expected error for foreign generation")
	}

	// Cancelled context rejection.
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := session.Query(cancelled, Query{Scope: binding.Scope, Request: &CheckerRequest{Op: "void"}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Unknown JSON operation rejection (including removed dead ops like apparentProperties).
	var q Query
	if err := json.Unmarshal([]byte(`{"generation":1,"op":"apparentProperties","args":{"id":1}}`), &q); err == nil {
		t.Fatal("expected unsupported operation error for removed apparentProperties op")
	}
}

func TestSanitizeTruncatedUTF8(t *testing.T) {
	// Valid strings pass through unchanged.
	if got := sanitizeTruncatedUTF8("例例..."); got != "例例..." {
		t.Fatalf("unexpected change to valid string: %q", got)
	}
	// Truncation mid-rune trims incomplete trailing UTF-8 bytes before "...".
	broken := "例例"[:5] + "..."
	if utf8.ValidString(broken) {
		t.Fatal("test setup expected invalid UTF-8")
	}
	if got := sanitizeTruncatedUTF8(broken); got != "例..." || !utf8.ValidString(got) {
		t.Fatalf("expected valid trimmed string %q, got %q", "例...", got)
	}
}

func TestSyntaxOnlyDescriptorPreservesCompilerOptions(t *testing.T) {
	const root = "/type-service-test"
	fileName := tspath.ResolvePath(root, "input.ts")
	configName := tspath.ResolvePath(root, "tsconfig.json")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		configName: `{"compilerOptions":{"noLib":true,"experimentalDecorators":true,"emitDecoratorMetadata":true,"isolatedDeclarations":true}}`,
		fileName:   "export class A {}",
	})
	raw, err := utils.CreateProgram(false, fs, root, configName, utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatal(err)
	}
	p := program.NewFromCompiler(raw)
	file := p.GetSourceFile(fileName)
	session, err := NewSyntaxOnly(t.Context(), p, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	for _, desc := range []struct {
		name string
		get  func() (Descriptor, error)
	}{
		{"Descriptor", session.Descriptor},
		{"SyntaxDescriptor", func() (Descriptor, error) { return session.SyntaxDescriptor(file) }},
	} {
		d, err := desc.get()
		if err != nil {
			t.Fatalf("%s: %v", desc.name, err)
		}
		if !d.SyntaxOnly {
			t.Fatalf("%s: expected SyntaxOnly=true", desc.name)
		}
		var opts map[string]any
		if err := json.Unmarshal(d.CompilerOptions, &opts); err != nil {
			t.Fatalf("%s: invalid CompilerOptions JSON: %v", desc.name, err)
		}
		for _, key := range []string{"experimentalDecorators", "emitDecoratorMetadata", "isolatedDeclarations"} {
			if opts[key] != true {
				t.Fatalf("%s: expected %s=true in syntax-only descriptor, got %v", desc.name, key, opts[key])
			}
		}
	}
}

func TestConfigDiagnosticSourceAvailableWithoutLeakingIntoProgramSources(t *testing.T) {
	const root = "/type-service-config-diag"
	fileName := tspath.ResolvePath(root, "input.ts")
	configName := tspath.ResolvePath(root, "tsconfig.json")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		configName: "{\n  \"compilerOptions\": {\n    \"noLib\": true,\n    \"inlineSources\": true\n  },\n  \"include\": [\"input.ts\"]\n}\n",
		fileName:   "export const x = 1;\n",
	})
	raw, err := utils.CreateProgram(false, fs, root, configName, utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatal(err)
	}
	p := program.NewFromCompiler(raw)
	file := p.GetSourceFile(fileName)
	session, err := New(t.Context(), p, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	desc, err := session.Descriptor()
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range desc.Sources {
		if src.FileName == configName {
			t.Fatalf("tsconfig %q should not be listed in Program Descriptor.Sources", configName)
		}
	}

	cfgSrc, err := session.Source(desc.Generation, configName)
	if err != nil {
		t.Fatalf("Source(tsconfig): %v", err)
	}
	if cfgSrc == nil || cfgSrc.FileName() != configName {
		t.Fatalf("expected tsconfig SourceFile %q, got %v", configName, cfgSrc)
	}
	if len(desc.DiagnosticSources) != 1 || desc.DiagnosticSources[0].FileName != configName || desc.DiagnosticSources[0].Path != string(cfgSrc.Path()) {
		t.Fatalf("expected DiagnosticSources to contain tsconfig with compiler path %q, got %+v", cfgSrc.Path(), desc.DiagnosticSources)
	}

	// Cold nodeRecord queries against a config diagnostic source must succeed.
	packet, err := session.Query(t.Context(), Query{
		Scope:   Scope{Generation: desc.Generation},
		Request: &NodeRequest{Op: "nodeRecord", Node: NodeRef{File: configName, Index: 1}},
	})
	if err != nil || len(packet) == 0 {
		t.Fatalf("cold nodeRecord query on tsconfig failed: len=%d err=%v", len(packet), err)
	}
}

type caseSensitiveFS struct {
	vfs.FS
}

func (caseSensitiveFS) UseCaseSensitiveFileNames() bool { return true }

func TestConfigDiagnosticSourcePreservesCaseSensitivePath(t *testing.T) {
	const root = "/type-service-case-sensitive"
	upperCfg := tspath.ResolvePath(root, "PKG/tsconfig.json")
	upperFile := tspath.ResolvePath(root, "PKG/input.ts")
	lowerCfg := tspath.ResolvePath(root, "pkg/tsconfig.json")
	lowerFile := tspath.ResolvePath(root, "pkg/input.ts")
	cfgBody := "{\n  \"compilerOptions\": {\n    \"noLib\": true,\n    \"inlineSources\": true\n  },\n  \"include\": [\"input.ts\"]\n}\n"
	fs := caseSensitiveFS{utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		upperCfg:  cfgBody,
		upperFile: "export const upper = 1;\n",
		lowerCfg:  cfgBody,
		lowerFile: "export const lower = 2;\n",
	})}

	openDiagPath := func(cfg, file string) string {
		raw, err := utils.CreateProgram(false, fs, root, cfg, utils.CreateCompilerHost(root, fs))
		if err != nil {
			t.Fatal(err)
		}
		p := program.NewFromCompiler(raw)
		session, err := New(t.Context(), p, []*ast.SourceFile{p.GetSourceFile(file)})
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()
		desc, err := session.Descriptor()
		if err != nil {
			t.Fatal(err)
		}
		if len(desc.DiagnosticSources) != 1 {
			t.Fatalf("expected 1 diagnostic source for %q, got %+v", cfg, desc.DiagnosticSources)
		}
		return desc.DiagnosticSources[0].Path
	}

	upperPath := openDiagPath(upperCfg, upperFile)
	lowerPath := openDiagPath(lowerCfg, lowerFile)
	if upperPath != upperCfg || lowerPath != lowerCfg || upperPath == lowerPath {
		t.Fatalf("expected distinct case-sensitive config paths %q and %q, got %q and %q", upperCfg, lowerCfg, upperPath, lowerPath)
	}
}

func TestResolveNodeTokenValidationDoesNotPolluteSourceFileCache(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{
		"input.ts": "export type Nested = Array<Array<number>>;\n",
	}, []string{"input.ts"})
	fileName := tspath.ResolvePath(root, "input.ts")
	f := p.GetSourceFile(fileName)
	table := encoder.GetNodeIndexTable(f)
	if len(table.Nodes) < 2 {
		t.Fatal("expected populated node index table")
	}

	// Find the outer TypeReference node (`Array<Array<number>>`) and its two `>` tokens.
	var outerTypeRef uint32
	for idx, n := range table.Nodes {
		if n != nil && n.Kind == ast.KindTypeReference && strings.Contains(f.Text()[n.Pos():n.End()], "Array<Array<number>>") {
			outerTypeRef = uint32(idx)
			break
		}
	}
	if outerTypeRef == 0 {
		t.Fatal("failed to locate outer TypeReference")
	}

	// Mismatched client-supplied token kind MUST be rejected and MUST NOT
	// write into SourceFile.tokenCache.
	if _, err := resolveNode(f, NodeRef{
		File:  fileName,
		Index: 1,
		Token: &SyntaxToken{Kind: ast.KindPlusToken, Pos: 0, End: 6},
	}); err == nil {
		t.Fatal("expected mismatched token kind to be rejected")
	}
	// Calling SourceFile.GetOrCreateToken with the real kind afterwards must not panic.
	if tok := f.GetOrCreateToken(ast.KindExportKeyword, 0, 6, table.Nodes[1], ast.TokenFlagsNone); tok == nil || tok.Kind != ast.KindExportKeyword {
		t.Fatalf("expected clean SourceFile token cache after rejected token, got %v", tok)
	}

	// Valid re-scanned token (`>` inside `>>`) must succeed as an ephemeral token.
	gtPos := strings.Index(f.Text(), ">>")
	if gtPos < 0 {
		t.Fatal("expected >> in fixture")
	}
	resolved, err := resolveNode(f, NodeRef{
		File:  fileName,
		Index: outerTypeRef,
		Token: &SyntaxToken{Kind: ast.KindGreaterThanToken, Pos: gtPos, End: gtPos + 1},
	})
	if err != nil || resolved.Kind != ast.KindGreaterThanToken {
		t.Fatalf("expected re-scanned > token, got %v err=%v", resolved, err)
	}
}

func TestInvalidHandlesAndNodeKindsReturnErrorsWithoutPoisoningSession(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{
		"input.ts": "export const value = 1;\n",
	}, []string{"input.ts"})
	fileName := tspath.ResolvePath(root, "input.ts")
	f := p.GetSourceFile(fileName)
	session, err := New(t.Context(), p, []*ast.SourceFile{f})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	desc, err := session.Descriptor()
	if err != nil {
		t.Fatal(err)
	}

	// Invalid handles must return regular errors without poisoning the session.
	for _, req := range []Request{
		&TypeRequest{Op: "typeRecord", ID: 999999},
		&SymbolRequest{Op: "symbolRecord", ID: 999999},
		&SignatureRequest{Op: "signatureRecord", ID: 999999},
	} {
		if _, err := session.Query(t.Context(), Query{
			Scope:   Scope{Generation: desc.Generation},
			Request: req,
		}); err == nil || errors.Is(err, ErrPoisoned) {
			t.Fatalf("expected non-poisoning handle error for %q, got %v", req.operation(), err)
		}
	}

	// Wrong node kinds for checker ops that otherwise panic must be handled safely.
	srcRef := NodeRef{File: fileName, Index: 1}
	for _, req := range []Request{
		&NodeRequest{Op: "signature", Node: srcRef},
		&NodeRequest{Op: "propertySymbolOfDestructuringAssignment", Node: srcRef},
	} {
		if _, err := session.Query(t.Context(), Query{
			Scope:   Scope{Generation: desc.Generation},
			Request: req,
		}); err != nil {
			t.Fatalf("expected nil-handle response for %q on SourceFile, got err=%v", req.operation(), err)
		}
	}
	for _, req := range []Request{
		&NodeRequest{Op: "exportSpecifierLocalTargetSymbol", Node: srcRef},
		&NamedNodeRequest{Op: "validPropertyAccess", Node: srcRef, Name: "x"},
	} {
		if _, err := session.Query(t.Context(), Query{
			Scope:   Scope{Generation: desc.Generation},
			Request: req,
		}); err == nil || errors.Is(err, ErrPoisoned) {
			t.Fatalf("expected non-poisoning validation error for %q on SourceFile, got %v", req.operation(), err)
		}
	}

	// Session must still serve valid queries after all rejected inputs.
	if _, err := session.Query(t.Context(), Query{
		Scope:   Scope{Generation: desc.Generation},
		Request: &CheckerRequest{Op: "stringType"},
	}); err != nil {
		t.Fatalf("session unexpectedly unusable after validation errors: %v", err)
	}

	// A recovered internal panic must poison the session against further use.
	session.poisoned.Store(true)
	if _, err := session.Query(t.Context(), Query{
		Scope:   Scope{Generation: desc.Generation},
		Request: &CheckerRequest{Op: "stringType"},
	}); !errors.Is(err, ErrPoisoned) {
		t.Fatalf("expected ErrPoisoned after panic, got %v", err)
	}
}

type hookContext struct {
	context.Context
	calls atomic.Int32
	onErr func(call int32) error
}

func (c *hookContext) Err() error {
	call := c.calls.Add(1)
	if c.onErr != nil {
		if err := c.onErr(call); err != nil {
			return err
		}
	}
	return c.Context.Err()
}

type gateRequest struct {
	op   Operation
	onOp func()
}

func (r *gateRequest) operation() Operation {
	if r.onOp != nil {
		r.onOp()
	}
	return r.op
}

func TestSessionConcurrentBatchRejectedAfterPoison(t *testing.T) {
	t.Run("paused mixed batch", func(t *testing.T) {
		p, _, _ := buildTestProgram(t, map[string]string{
			"input.ts": "export const value = 1;",
		}, []string{"input.ts"})
		session, err := New(t.Context(), p, p.SourceFiles())
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()
		desc, err := session.Descriptor()
		if err != nil {
			t.Fatal(err)
		}

		paused := make(chan struct{})
		resume := make(chan struct{})
		var gateOnce sync.Once
		batchDone := make(chan error, 1)
		go func() {
			_, err := session.QueryBatch(t.Context(), []Query{
				{Scope: Scope{Generation: desc.Generation}, Request: &ProgramRequest{Op: "rootFileNames"}},
				{Scope: Scope{Generation: desc.Generation}, Request: &CheckerRequest{Op: "stringType"}},
				{Scope: Scope{Generation: desc.Generation}, Request: &gateRequest{
					op: "nodeRecord",
					onOp: func() {
						gateOnce.Do(func() {
							close(paused)
							<-resume
						})
					},
				}},
			})
			batchDone <- err
		}()

		<-paused

		// Trigger a controlled panic inside checkBatch while the mixed batch is
		// already past QueryBatch's initial poison check.
		panicCtx := &hookContext{
			Context: t.Context(),
			onErr: func(call int32) error {
				if call == 2 {
					panic("controlled internal failure")
				}
				return nil
			},
		}
		if _, err := session.Query(panicCtx, Query{
			Scope:   Scope{Generation: desc.Generation},
			Request: &CheckerRequest{Op: "stringType"},
		}); err == nil || !strings.Contains(err.Error(), "controlled internal failure") {
			t.Fatalf("expected recovered internal failure, got %v", err)
		}
		if _, err := session.Query(t.Context(), Query{
			Scope:   Scope{Generation: desc.Generation},
			Request: &CheckerRequest{Op: "stringType"},
		}); !errors.Is(err, ErrPoisoned) {
			t.Fatalf("expected ErrPoisoned on fresh query, got %v", err)
		}

		close(resume)
		if err := <-batchDone; !errors.Is(err, ErrPoisoned) {
			t.Fatalf("expected paused mixed batch to return ErrPoisoned, got %v", err)
		}
		if session.checker != nil {
			t.Fatal("resumed mixed batch unexpectedly acquired the semantic checker after poison")
		}
		if len(session.types) != 0 {
			t.Fatalf("resumed mixed batch created %d type handles after poison", len(session.types))
		}
	})

	t.Run("waiter queued on checkMu", func(t *testing.T) {
		p, _, _ := buildTestProgram(t, map[string]string{
			"input.ts": "export const value = 1;",
		}, []string{"input.ts"})
		session, err := New(t.Context(), p, p.SourceFiles())
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()
		desc, err := session.Descriptor()
		if err != nil {
			t.Fatal(err)
		}

		holdingCheckMu := make(chan struct{})
		triggerPanic := make(chan struct{})
		panicDone := make(chan error, 1)
		panicCtx := &hookContext{
			Context: t.Context(),
			onErr: func(call int32) error {
				// Call 4 is inside checkBatch's query loop right after the
				// checker lease is acquired and before execute() runs.
				if call == 4 {
					close(holdingCheckMu)
					<-triggerPanic
					panic("controlled checkMu panic")
				}
				return nil
			},
		}
		go func() {
			_, err := session.Query(panicCtx, Query{
				Scope:   Scope{Generation: desc.Generation},
				Request: &CheckerRequest{Op: "stringType"},
			})
			panicDone <- err
		}()

		<-holdingCheckMu

		waiterEntered := make(chan struct{})
		waiterDone := make(chan error, 1)
		waiterCtx := &hookContext{
			Context: t.Context(),
			onErr: func(call int32) error {
				if call == 1 {
					close(waiterEntered)
				}
				return nil
			},
		}
		go func() {
			_, err := session.Query(waiterCtx, Query{
				Scope:   Scope{Generation: desc.Generation},
				Request: &CheckerRequest{Op: "stringType"},
			})
			waiterDone <- err
		}()

		<-waiterEntered
		time.Sleep(10 * time.Millisecond)
		close(triggerPanic)

		if err := <-panicDone; err == nil || !strings.Contains(err.Error(), "controlled checkMu panic") {
			t.Fatalf("expected recovered checkMu panic, got %v", err)
		}
		if err := <-waiterDone; !errors.Is(err, ErrPoisoned) {
			t.Fatalf("expected waiter on checkMu to observe ErrPoisoned, got %v", err)
		}
		if len(session.types) != 0 {
			t.Fatalf("waiter on checkMu executed semantic query and created %d type handles after poison", len(session.types))
		}
	})
}

func TestProtocolOperationsMatchTypeScriptTables(t *testing.T) {
	protocolPath := filepath.Join("..", "..", "packages", "rslint", "src", "languages", "js", "type-service-protocol.ts")
	raw, err := os.ReadFile(protocolPath)
	if err != nil {
		t.Fatalf("read TS protocol file: %v", err)
	}
	const virtualPath = "/type-service-protocol.ts"
	sf := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: virtualPath,
		Path:     tspath.Path(virtualPath),
	}, string(raw), core.ScriptKindTS)

	var scalarInit, directInit *ast.Node
	var findDecls ast.Visitor
	findDecls = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			vd := node.AsVariableDeclaration()
			if vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier && vd.Initializer != nil {
				switch vd.Name().Text() {
				case "SCALAR_KINDS":
					scalarInit = ast.SkipOuterExpressions(vd.Initializer, ast.OEKAll)
				case "DIRECT_RECORD_OPS":
					directInit = ast.SkipOuterExpressions(vd.Initializer, ast.OEKAll)
				}
			}
		}
		node.ForEachChild(findDecls)
		return false
	}
	sf.ForEachChild(findDecls)

	if scalarInit == nil || scalarInit.Kind != ast.KindObjectLiteralExpression {
		t.Fatal("SCALAR_KINDS object literal not found in TS AST")
	}
	tsOps := map[Operation]bool{
		"typeRecord":      true,
		"symbolRecord":    true,
		"signatureRecord": true,
		"nodeRecord":      true,
	}
	for _, prop := range scalarInit.AsObjectLiteralExpression().Properties.Nodes {
		if prop.Kind != ast.KindPropertyAssignment {
			continue
		}
		name := prop.AsPropertyAssignment().Name()
		if name == nil {
			continue
		}
		switch name.Kind {
		case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			tsOps[Operation(name.Text())] = true
		}
	}
	if len(tsOps) < 50 {
		t.Fatalf("unexpectedly few TS operations parsed: %d", len(tsOps))
	}
	for op := range tsOps {
		var q Query
		payload, err := json.Marshal(struct {
			Generation uint32         `json:"generation"`
			Op         Operation      `json:"op"`
			Args       map[string]int `json:"args"`
		}{Generation: 1, Op: op, Args: map[string]int{}})
		if err != nil {
			t.Fatalf("marshal %q: %v", op, err)
		}
		if err := json.Unmarshal(payload, &q); err != nil {
			t.Errorf("TS operation %q failed to decode in Go Query.UnmarshalJSON: %v", op, err)
		}
	}

	if directInit == nil || directInit.Kind != ast.KindNewExpression {
		t.Fatal("DIRECT_RECORD_OPS new Set(...) expression not found in TS AST")
	}
	directArgs := directInit.AsNewExpression().Arguments
	if directArgs == nil || len(directArgs.Nodes) == 0 {
		t.Fatal("DIRECT_RECORD_OPS Set argument missing")
	}
	directArray := ast.SkipOuterExpressions(directArgs.Nodes[0], ast.OEKAll)
	if directArray == nil || directArray.Kind != ast.KindArrayLiteralExpression {
		t.Fatal("DIRECT_RECORD_OPS array literal argument not found in TS AST")
	}
	tsDirect := make(map[Operation]bool)
	for _, elem := range directArray.AsArrayLiteralExpression().Elements.Nodes {
		if elem.Kind == ast.KindStringLiteral || elem.Kind == ast.KindNoSubstitutionTemplateLiteral {
			tsDirect[Operation(elem.Text())] = true
		}
	}
	for op := range tsOps {
		inGo := directRecordKind(op) != ""
		inTS := tsDirect[op]
		if inGo != inTS {
			t.Errorf("direct-record mismatch for %q: Go=%v TS=%v", op, inGo, inTS)
		}
	}
	for op := range tsDirect {
		if !tsOps[op] {
			t.Errorf("DIRECT_RECORD_OPS contains unknown op %q", op)
		}
	}
}

func TestJSCheckFlagsBitAlignment(t *testing.T) {
	// Locks jsCheckFlags() against future typescript-go CheckFlags bit shifts.
	// Expected bit positions match TypeScript 6.0.3's ts.CheckFlags.
	cases := []struct {
		name string
		goF  ast.CheckFlags
		want uint32
	}{
		{"None", ast.CheckFlagsNone, 0},
		{"Instantiated", ast.CheckFlagsInstantiated, 1 << 0},
		{"SyntheticProperty", ast.CheckFlagsSyntheticProperty, 1 << 1},
		{"SyntheticMethod", ast.CheckFlagsSyntheticMethod, 1 << 2},
		{"Readonly", ast.CheckFlagsReadonly, 1 << 3},
		{"ReadPartial", ast.CheckFlagsReadPartial, 1 << 4},
		{"WritePartial", ast.CheckFlagsWritePartial, 1 << 5},
		{"HasNonUniformType", ast.CheckFlagsHasNonUniformType, 1 << 6},
		{"HasLiteralType", ast.CheckFlagsHasLiteralType, 1 << 7},
		{"ContainsPublic", ast.CheckFlagsContainsPublic, 1 << 8},
		{"ContainsProtected", ast.CheckFlagsContainsProtected, 1 << 9},
		{"ContainsPrivate", ast.CheckFlagsContainsPrivate, 1 << 10},
		{"ContainsStatic", ast.CheckFlagsContainsStatic, 1 << 11},
		{"Late", ast.CheckFlagsLate, 1 << 12},
		{"ReverseMapped", ast.CheckFlagsReverseMapped, 1 << 13},
		{"OptionalParameter", ast.CheckFlagsOptionalParameter, 1 << 14},
		{"RestParameter", ast.CheckFlagsRestParameter, 1 << 15},
		{"DeferredType", ast.CheckFlagsDeferredType, 1 << 16},
		{"HasNeverType", ast.CheckFlagsHasNeverType, 1 << 17},
		{"Mapped", ast.CheckFlagsMapped, 1 << 18},
		{"StripOptional", ast.CheckFlagsStripOptional, 1 << 19},
		{"Unresolved", ast.CheckFlagsUnresolved, 1 << 20},
		// Go-only private bits must be stripped cleanly.
		{"ContainsWritePublic", ast.CheckFlagsContainsWritePublic, 0},
		{"ContainsWriteProtected", ast.CheckFlagsContainsWriteProtected, 0},
		{"ContainsWritePrivate", ast.CheckFlagsContainsWritePrivate, 0},
		{"IsDiscriminant", ast.CheckFlagsIsDiscriminant, 0},
		{"IsDiscriminantComputed", ast.CheckFlagsIsDiscriminantComputed, 0},
		{"IndexSymbol", ast.CheckFlagsIndexSymbol, 0},
	}
	for _, tc := range cases {
		if got := jsCheckFlags(tc.goF); got != tc.want {
			t.Errorf("jsCheckFlags(%s = %#x) = %#x, want %#x", tc.name, uint32(tc.goF), got, tc.want)
		}
	}
}
