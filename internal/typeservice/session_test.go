package typeservice

import (
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
	for _, owner := range []*program.Program{p, sourceOnly} {
		syntax, err := NewSyntaxOnly(t.Context(), owner, []*ast.SourceFile{file})
		if err != nil {
			t.Fatal(err)
		}
		binding, err := syntax.BindTarget(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := syntax.Query(t.Context(), Query{binding.Scope, &NodeRequest{Op: "nodeRecord", Node: NodeRef{name, 1}}}); err != nil {
			t.Fatal(err)
		}
		if _, err := syntax.Query(t.Context(), Query{binding.Scope, &VoidRequest{}}); err == nil {
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
	if _, err := session.Query(t.Context(), Query{Scope: binding.Scope, Request: &VoidRequest{}}); err != nil {
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
