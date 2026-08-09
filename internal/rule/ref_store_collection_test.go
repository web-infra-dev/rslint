package rule

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

func TestRefStoreStreamingEarlyQueryNeverReturnsPartial(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/streaming-early.ts", core.ScriptKindTS,
		"export {}; let x = 0, y = 0; consume(x); consume(x); consume(y); consume(x);")
	x := identifiers(sourceFile.AsNode(), "x")
	if len(x) != 4 {
		t.Fatalf("x occurrences = %d, want declaration plus three references", len(x))
	}
	sym := x[0].Parent.Symbol()
	if sym == nil {
		t.Fatal("x declaration has no symbol")
	}

	refs.request(RefNeedReferences)
	collector := RefCollector{store: refs}
	observe := collector.Start()
	if !observe.Active() || refs.state != refCollectionStreaming {
		t.Fatalf("collector Start state = %v, want streaming", refs.state)
	}
	observe.Observe(x[0])
	observe.Observe(x[1])
	first := refs.References(sym)
	if len(first) != 3 || first[0] != x[1] || first[1] != x[2] || first[2] != x[3] {
		t.Fatalf("early References = %v, want all future references %v", first, x[1:])
	}
	if refs.state != refCollectionCompleteByPrepass {
		t.Fatalf("early query state = %v, want complete-by-prepass", refs.state)
	}

	for _, node := range identifiers(sourceFile.AsNode(), "consume") {
		observe.Observe(node)
	}
	for _, node := range x[2:] {
		observe.Observe(node)
	}
	collector.Complete()
	second := refs.References(sym)
	if len(second) != len(first) {
		t.Fatalf("post-traversal References = %v, want stable early snapshot %v", second, first)
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("post-traversal reference %d changed", index)
		}
	}
	y := identifiers(sourceFile.AsNode(), "y")
	if len(y) != 2 {
		t.Fatalf("y occurrences = %d, want declaration and reference", len(y))
	}
	if got := refs.References(y[0].Parent.Symbol()); len(got) != 1 || got[0] != y[1] {
		t.Fatalf("different-name References after early fallback = %v, want y reference", got)
	}
}

func TestRefStoreStreamingCompletesAllRequestedCollections(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/streaming-all.ts", core.ScriptKindTS, `
import { value as imported } from "m";
const local = imported;
consume(local);
`)
	refs.request(RefNeedsAll)
	collector := RefCollector{store: refs}
	observe := collector.Start()
	if !observe.Active() {
		t.Fatal("requested features did not start collection")
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			observe.Observe(node)
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)
	collector.Complete()
	collector.Complete()
	if refs.state != refCollectionCompleteByStream || refs.built != RefNeedsAll {
		t.Fatalf("completed state = %v, built = %v; want stream/all", refs.state, refs.built)
	}
	if got := refs.BindingDeclarations(); len(got) != 2 {
		t.Fatalf("binding declarations = %d, want imported and local", len(got))
	}
	if got := refs.ImportBindings(); len(got) != 1 || got[0].LocalName.Text() != "imported" {
		t.Fatalf("import bindings = %#v, want imported", got)
	}
	local := identifiers(sourceFile.AsNode(), "local")
	if got := refs.References(local[0].Parent.Symbol()); len(got) != 1 || got[0] != local[1] {
		t.Fatalf("local references = %v, want final consume reference", got)
	}
}

func TestRefStoreNoRequestKeepsCollectionsCold(t *testing.T) {
	_, refs := newBoundRefStore(t, "/cold.ts", core.ScriptKindTS, "const value = 1; consume(value);")
	if observe := (RefCollector{store: refs}).Start(); observe.Active() {
		t.Fatal("store without a dynamic request started collection")
	}
	if refs.state != refCollectionCold || refs.candidates != nil || refs.refs != nil || refs.merged != nil || refs.facts != nil {
		t.Fatalf("cold store materialized collection state: %#v", refs)
	}
}

func TestRefStoreLateRequestUsesCompletePrepass(t *testing.T) {
	_, refs := newBoundRefStore(t, "/late-request.ts", core.ScriptKindTS,
		"import { value } from 'm'; const local = value;")
	if observe := (RefCollector{store: refs}).Start(); observe.Active() {
		t.Fatal("store unexpectedly started before a request")
	}
	refs.request(RefNeedBindingDeclarations | RefNeedImportBindings)
	if refs.state != refCollectionCompleteByPrepass {
		t.Fatalf("late request state = %v, want complete-by-prepass", refs.state)
	}
	if len(refs.BindingDeclarations()) != 2 || len(refs.ImportBindings()) != 1 {
		t.Fatal("late request did not build complete declaration/import facts")
	}
}

func TestRefStoreBuildsMixedPrepassAndStreamingStages(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/mixed-stages.ts", core.ScriptKindTS, `
import { value as imported } from "m";
const local = imported;
consume(local);
`)
	declarations := refs.BindingDeclarations()
	if len(declarations) != 2 {
		t.Fatalf("prepass declarations = %d, want imported and local", len(declarations))
	}
	refs.request(RefNeedReferences | RefNeedImportBindings)
	collector := RefCollector{store: refs}
	observer := collector.Start()
	if !observer.Active() {
		t.Fatal("missing reference/import features did not start streaming")
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			observer.Observe(node)
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)
	collector.Complete()
	if len(refs.ImportBindings()) != 1 || len(refs.BindingDeclarations()) != len(declarations) {
		t.Fatal("later feature build changed an already-materialized declaration collection")
	}
	local := identifiers(sourceFile.AsNode(), "local")
	if got := refs.References(local[0].Parent.Symbol()); len(got) != 1 || got[0] != local[1] {
		t.Fatalf("mixed-stage references = %v, want local use", got)
	}
}

func TestRefStoreLateFeatureRequestDuringStreamingFallsBackAtomically(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/streaming-late-feature.ts", core.ScriptKindTS, `
import { value as imported } from "m";
const local = imported;
consume(local);
`)
	refs.request(RefNeedReferences)
	collector := RefCollector{store: refs}
	observer := collector.Start()
	if !observer.Active() {
		t.Fatal("reference collection did not start")
	}
	first := identifiers(sourceFile.AsNode(), "imported")[0]
	observer.Observe(first)
	refs.request(RefNeedImportBindings)
	if refs.state != refCollectionCompleteByPrepass || refs.built != RefNeedReferences|RefNeedImportBindings {
		t.Fatalf("late feature request state=%v built=%v, want atomic full prepass", refs.state, refs.built)
	}
	observer.Observe(identifiers(sourceFile.AsNode(), "local")[1])
	collector.Complete()
	if got := refs.ImportBindings(); len(got) != 1 || got[0].LocalName != first {
		t.Fatalf("late-request imports = %#v, want complete import collection", got)
	}
	local := identifiers(sourceFile.AsNode(), "local")
	if got := refs.References(local[0].Parent.Symbol()); len(got) != 1 || got[0] != local[1] {
		t.Fatalf("late-request references = %v, want complete local use", got)
	}
}

func TestRefStoreEarlyFactMaterializationNeverReturnsLiveSlices(t *testing.T) {
	for _, firstQuery := range []RefNeeds{RefNeedBindingDeclarations, RefNeedImportBindings} {
		name := "declarations"
		if firstQuery == RefNeedImportBindings {
			name = "imports"
		}
		t.Run(name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/early-facts.ts", core.ScriptKindTS, `
import { a as A, b as B } from "m";
const local = A;
consume(local, B);
`)
			refs.request(RefNeedBindingDeclarations | RefNeedImportBindings)
			collector := RefCollector{store: refs}
			observer := collector.Start()
			if !observer.Active() {
				t.Fatal("fact collection did not start")
			}
			localBindings := identifiers(sourceFile.AsNode(), "A")
			if len(localBindings) != 2 {
				t.Fatal("fixture did not expose the A binding and its use")
			}
			// Seed both streaming fact slices with a real local import binding;
			// the early query must discard that partial prefix before rebuilding.
			observer.Observe(localBindings[0])

			var firstDeclaration *BindingDeclarationFacts
			var firstImport *ImportBinding
			if firstQuery == RefNeedBindingDeclarations {
				got := refs.BindingDeclarations()
				if len(got) != 3 {
					t.Fatalf("early declarations = %d, want A, B, local", len(got))
				}
				firstDeclaration = &got[0]
			} else {
				got := refs.ImportBindings()
				if len(got) != 2 {
					t.Fatalf("early imports = %d, want A and B", len(got))
				}
				firstImport = &got[0]
			}
			if refs.state != refCollectionCompleteByPrepass {
				t.Fatalf("early fact query state = %v, want full-prepass completion", refs.state)
			}

			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindIdentifier {
					observer.Observe(node)
				}
				node.ForEachChild(visit)
				return false
			}
			sourceFile.AsNode().ForEachChild(visit)
			collector.Complete()
			if firstQuery == RefNeedBindingDeclarations {
				got := refs.BindingDeclarations()
				if len(got) != 3 || &got[0] != firstDeclaration {
					t.Fatal("post-traversal declaration slice changed after publication")
				}
				if len(refs.ImportBindings()) != 2 {
					t.Fatal("sibling import feature was not completed by atomic fallback")
				}
			} else {
				got := refs.ImportBindings()
				if len(got) != 2 || &got[0] != firstImport {
					t.Fatal("post-traversal import slice changed after publication")
				}
				if len(refs.BindingDeclarations()) != 3 {
					t.Fatal("sibling declaration feature was not completed by atomic fallback")
				}
			}
		})
	}
}
