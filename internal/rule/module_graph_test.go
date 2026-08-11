package rule

import "testing"

func TestStandaloneModuleGraphNormalizesTypedNilRuntime(t *testing.T) {
	var runtime *programSourceRuntime
	graph := NewStandaloneModuleGraph(nil, runtime)
	if graph.sourceRuntime() != nil || graph.Files() != nil {
		t.Fatal("typed-nil standalone runtime must produce an empty module graph")
	}
}
