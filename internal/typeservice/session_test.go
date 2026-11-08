package typeservice

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
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
	if _, err := session.BindTarget(file); err != nil {
		t.Fatal(err)
	}
	table := encoder.GetNodeIndexTable(file)
	missing := uint32(0)
	for index, node := range table.Nodes {
		if index > 0 && node == nil {
			missing = uint32(index)
			break
		}
	}
	if missing == 0 {
		t.Fatal("fixture has no node-list slot")
	}
	if _, err := session.Query(t.Context(), Query{Scope: Scope{descriptor.Generation, descriptor.Checker}, Request: &ResolveNameRequest{Name: "value", Node: &NodeRef{name, missing}}}); err == nil {
		t.Fatal("node-list slot admitted as global resolveName location")
	}
	query := Query{Scope: Scope{Generation: descriptor.Generation, Checker: descriptor.Checker}, Request: &VoidRequest{}}
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
	var identifier uint32
	for index, node := range table.Nodes {
		if node != nil && ast.IsIdentifier(node) {
			identifier = uint32(index)
			break
		}
	}
	if identifier == 0 {
		t.Fatal("fixture identifier missing")
	}
	batch := []Query{query, {Scope: query.Scope, Request: &NodeRequest{Op: "type", Node: NodeRef{name, identifier}}}, {Scope: query.Scope, Request: &NodeRequest{Op: "nodeRecord", Node: NodeRef{name, identifier}}}}
	packets, err := session.QueryBatch(t.Context(), batch)
	if err != nil {
		t.Fatal(err)
	}
	for i, q := range batch {
		packet, err := session.Query(t.Context(), q)
		if err != nil || !bytes.Equal(packet, packets[i]) {
			t.Fatalf("batch query %d differs from scalar: %v", i, err)
		}
	}
	// A metadata-only batch must not borrow a checker held by a native rule.
	_, release := p.TypeCheckerForFileExclusive(core.WithCheckerLifetime(t.Context(), core.CheckerLifetimeAPI), file)
	metadataDone := make(chan error, 1)
	go func() { _, err := session.QueryBatch(t.Context(), batch[2:]); metadataDone <- err }()
	select {
	case err := <-metadataDone:
		release()
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		release()
		t.Fatal("syntax metadata borrowed a checker")
	}
	if _, err := session.QueryBatch(t.Context(), make([]Query, 257)); err == nil {
		t.Fatal("unbounded query batch admitted")
	}
	stale := query
	stale.Generation++
	if _, err := session.QueryBatch(t.Context(), []Query{query, stale}); err == nil {
		t.Fatal("stale batch admitted")
	}
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	canceledQuery := Query{Scope: query.Scope, Request: &NodeRequest{Op: "type", Node: NodeRef{"not-owned", 1}}}
	if _, err := session.Query(canceled, canceledQuery); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled query: %v", err)
	}
	if _, err := New(canceled, p, []*ast.SourceFile{file}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled construction: %v", err)
	}
}
