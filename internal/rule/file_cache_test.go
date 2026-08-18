package rule

import "testing"

type fileCacheTestKey struct{}

func TestCachedByFileReusesWithinOneFile(t *testing.T) {
	ctx := RuleContext{}.WithFileCache(NewFileCache())
	created := 0
	build := func() *int {
		created++
		return new(int)
	}

	first := CachedByFile(ctx, fileCacheTestKey{}, build)
	second := CachedByFile(ctx, fileCacheTestKey{}, build)
	if first != second {
		t.Fatal("same file cache did not reuse its value")
	}
	if created != 1 {
		t.Fatalf("builder called %d times, want 1", created)
	}
}

func TestCachedByFileSeparatesFiles(t *testing.T) {
	first := CachedByFile(
		RuleContext{}.WithFileCache(NewFileCache()),
		fileCacheTestKey{},
		func() *int { return new(int) },
	)
	second := CachedByFile(
		RuleContext{}.WithFileCache(NewFileCache()),
		fileCacheTestKey{},
		func() *int { return new(int) },
	)
	if first == second {
		t.Fatal("different file caches reused a value")
	}
}

func TestCachedByFileWithoutCacheBuildsEveryTime(t *testing.T) {
	created := 0
	build := func() *int {
		created++
		return new(int)
	}
	ctx := RuleContext{}
	first := CachedByFile(ctx, fileCacheTestKey{}, build)
	second := CachedByFile(ctx, fileCacheTestKey{}, build)
	if first == second {
		t.Fatal("context without a file cache unexpectedly reused a value")
	}
	if created != 2 {
		t.Fatalf("builder called %d times, want 2", created)
	}
}

func TestRuleContextProcessCurrentDirectoryUsesFileSharedState(t *testing.T) {
	ctx := RuleContext{}.WithFileCache(
		NewFileCacheWithProcessCurrentDirectory("/repo"),
	)
	if got := ctx.ProcessCurrentDirectory(); got != "/repo" {
		t.Fatalf("ProcessCurrentDirectory() = %q, want /repo", got)
	}
	if got := (&RuleContext{}).ProcessCurrentDirectory(); got != "" {
		t.Fatalf("empty ProcessCurrentDirectory() = %q, want empty", got)
	}
}
