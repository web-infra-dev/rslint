package rule

import "testing"

type generationScopeTestKey struct{ name string }

func TestCachedByGenerationRequiresScope(t *testing.T) {
	built := false
	_, ok := CachedByGeneration(RuleContext{}.WithFileCache(NewFileCache()), generationScopeTestKey{}, func(*GenerationScope) int {
		built = true
		return 1
	})
	if ok || built {
		t.Fatalf("ok = %v, built = %v without a generation scope", ok, built)
	}
}

func TestCachedByGenerationSharesAcrossFilesByKey(t *testing.T) {
	scope := NewGenerationScope(nil, true)
	builds := 0
	build := func(got *GenerationScope) *int {
		if got != scope {
			t.Fatal("builder received a different scope")
		}
		builds++
		return new(int)
	}
	first, _ := CachedByGeneration(RuleContext{}.WithFileCache(NewLintFileCache("/a", scope, nil)), generationScopeTestKey{"x"}, build)
	second, _ := CachedByGeneration(RuleContext{}.WithFileCache(NewLintFileCache("/b", scope, nil)), generationScopeTestKey{"x"}, build)
	other, _ := CachedByGeneration(RuleContext{}.WithFileCache(NewLintFileCache("/a", scope, nil)), generationScopeTestKey{"y"}, build)
	if first != second || first == other || builds != 2 {
		t.Fatalf("first == second: %v, first == other: %v, builds = %d", first == second, first == other, builds)
	}
	if !scope.SingleThreaded() {
		t.Fatal("scope lost its single-threaded request")
	}
}
