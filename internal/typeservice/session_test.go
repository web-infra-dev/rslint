package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
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

func TestProtocolOperationsMatchTypeScriptTables(t *testing.T) {
	protocolPath := filepath.Join("..", "..", "packages", "rslint", "src", "languages", "js", "type-service-protocol.ts")
	raw, err := os.ReadFile(protocolPath)
	if err != nil {
		t.Fatalf("read TS protocol file: %v", err)
	}
	src := string(raw)

	// Extract SCALAR_KINDS keys (+ the 4 record ops) and verify every TS op
	// decodes in Go Query.UnmarshalJSON.
	scalarStart := strings.Index(src, "export const SCALAR_KINDS")
	if scalarStart < 0 {
		t.Fatal("SCALAR_KINDS not found")
	}
	scalarEnd := strings.Index(src[scalarStart:], "};")
	if scalarEnd < 0 {
		t.Fatal("SCALAR_KINDS closing brace not found")
	}
	scalarBlock := src[scalarStart : scalarStart+scalarEnd]
	keyRe := regexp.MustCompile(`(?m)^\s+([a-zA-Z0-9_]+):\s+\d+,`)
	tsOps := map[Operation]bool{
		"typeRecord":      true,
		"symbolRecord":    true,
		"signatureRecord": true,
		"nodeRecord":      true,
	}
	for _, m := range keyRe.FindAllStringSubmatch(scalarBlock, -1) {
		tsOps[Operation(m[1])] = true
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

	// Extract DIRECT_RECORD_OPS from TS and assert exact 1:1 match with Go directRecordKind.
	directStart := strings.Index(src, "export const DIRECT_RECORD_OPS")
	if directStart < 0 {
		t.Fatal("DIRECT_RECORD_OPS not found")
	}
	directEnd := strings.Index(src[directStart:], "]);")
	if directEnd < 0 {
		t.Fatal("DIRECT_RECORD_OPS closing bracket not found")
	}
	directBlock := src[directStart : directStart+directEnd]
	strRe := regexp.MustCompile(`'([a-zA-Z0-9_]+)'`)
	tsDirect := make(map[Operation]bool)
	for _, m := range strRe.FindAllStringSubmatch(directBlock, -1) {
		tsDirect[Operation(m[1])] = true
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
