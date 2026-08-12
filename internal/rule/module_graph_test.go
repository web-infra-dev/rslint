package rule

import "testing"

func TestNilProgramProducesEmptyModuleGraph(t *testing.T) {
	graph := NewModuleGraph(nil)
	if graph.Files() != nil {
		t.Fatal("nil Program must produce an empty module graph")
	}
}
