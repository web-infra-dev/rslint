package typeservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// A native shard must remain available between JS queries. A session that
// retains an exclusive lease would deadlock this independently owned borrower.
func TestSessionReleasesCheckerLeaseAndPreservesEligibility(t *testing.T) {
	const root = "/type-service-test"
	name := tspath.ResolvePath(root, "input.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{name: "export const value = 1;"})
	raw, err := utils.CreateProgramFromOptions(false, &core.CompilerOptions{NoLib: core.TSTrue, Module: core.ModuleKindESNext}, []string{name}, utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatal(err)
	}
	p := program.NewFromCompiler(raw)
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
	session, err := New(t.Context(), p, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	descriptor, err := session.Descriptor()
	if err != nil {
		t.Fatal(err)
	}
	query := Query{Generation: descriptor.Generation, Checker: descriptor.Checker, Op: "void"}
	for range 2 {
		if _, err := session.Query(t.Context(), query); err != nil {
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
			t.Fatal("type service retained checker lease between queries")
		}
	}
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := session.Query(canceled, query); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled query: %v", err)
	}
	if _, err := New(canceled, p, []*ast.SourceFile{file}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled construction: %v", err)
	}
}
