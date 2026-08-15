package program_test

import (
	"testing"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

func TestNilProgramProducesEmptyModuleGraph(t *testing.T) {
	var sourceProgram *lintprogram.Program
	graph := sourceProgram.ModuleGraph()
	if graph.Files() != nil {
		t.Fatal("nil Program must produce an empty module graph")
	}
}

func TestModuleGraphRejectsForeignSourceGeneration(t *testing.T) {
	ownerRaw, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	foreignRaw, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	ownerFile := ownerRaw.GetSourceFile(specifierCacheImporter)
	foreignFile := foreignRaw.GetSourceFile(specifierCacheImporter)
	if ownerFile == nil || foreignFile == nil || ownerFile == foreignFile {
		t.Fatal("fixture did not create distinct source generations")
	}

	graph := lintprogram.NewFromCompiler(ownerRaw).ModuleGraph()
	if references := graph.References(foreignFile, lintprogram.ESModuleReferences); references != nil {
		t.Fatalf("foreign source produced module references: %+v", references)
	}
	references := graph.References(ownerFile, lintprogram.ESModuleReferences)
	if len(references) != 1 || references[0].Target == nil || references[0].Target.FileName() != specifierCacheTarget {
		t.Fatalf("owned source module references = %+v", references)
	}
}
