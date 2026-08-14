package rule

import (
	"testing"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

func TestNilProgramProducesEmptyModuleGraph(t *testing.T) {
	graph := ModuleGraphFor(nil)
	if graph.Files() != nil {
		t.Fatal("nil Program must produce an empty module graph")
	}
}

func TestModuleGraphRejectsForeignSourceGeneration(t *testing.T) {
	ownerRaw := programCacheTestProgram(t)
	foreignRaw := programCacheTestProgram(t)
	ownerFile := ownerRaw.GetSourceFile("/program-cache-fixture/file.ts")
	foreignFile := foreignRaw.GetSourceFile("/program-cache-fixture/file.ts")
	if ownerFile == nil || foreignFile == nil || ownerFile == foreignFile {
		t.Fatal("fixture did not create distinct source generations")
	}

	graph := ModuleGraphFor(lintprogram.NewFromCompiler(ownerRaw))
	syntax := ModuleSyntax{ESModule: true}
	if edges := graph.Edges(foreignFile, syntax); edges != nil {
		t.Fatalf("foreign source produced module edges: %+v", edges)
	}
	edges := graph.Edges(ownerFile, syntax)
	if len(edges) != 1 || edges[0].Target == nil || edges[0].Target.FileName() != "/program-cache-fixture/dependency.ts" {
		t.Fatalf("owned source module edges = %+v", edges)
	}
}
