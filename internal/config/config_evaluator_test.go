package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

type evaluatorPredicateCall struct {
	id        string
	path      string
	directory bool
}

type evaluatorPredicateResolver struct {
	mu        sync.Mutex
	functions map[string]func(string, bool) (bool, error)
	calls     []evaluatorPredicateCall
}

func (resolver *evaluatorPredicateResolver) ResolveConfigPredicate(
	_ context.Context,
	predicateID string,
	filePath string,
	directory bool,
) (bool, error) {
	resolver.mu.Lock()
	resolver.calls = append(resolver.calls, evaluatorPredicateCall{predicateID, filePath, directory})
	function := resolver.functions[predicateID]
	resolver.mu.Unlock()
	if function == nil {
		return false, errors.New("unknown predicate " + predicateID)
	}
	return function(filePath, directory)
}

func (resolver *evaluatorPredicateResolver) callIDs() []string {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	ids := make([]string, len(resolver.calls))
	for index, call := range resolver.calls {
		ids[index] = call.id
	}
	return ids
}

func mustDecodeEvaluatorConfig(t *testing.T, source string) RslintConfig {
	t.Helper()
	config, err := DecodeModuleConfig([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func TestConfigEvaluatorFilesShortCircuitAndUniversalOrder(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	filePath := tspath.CombinePaths(root, "x.ts")
	tests := []struct {
		name       string
		configJSON string
		values     map[string]bool
		wantCalls  []string
	}{
		{
			name: "top-level OR",
			configJSON: `[{"files":[
				{"$rslintPredicate":"a"},
				{"$rslintPredicate":"b"},
				{"$rslintPredicate":"unreached"}
			]}]`,
			values:    map[string]bool{"a": false, "b": true},
			wantCalls: []string{"a", "b"},
		},
		{
			name: "nested AND then next OR branch",
			configJSON: `[{"files":[[
				{"$rslintPredicate":"a"},
				{"$rslintPredicate":"b"},
				{"$rslintPredicate":"unreached"}
			],{"$rslintPredicate":"d"}]}]`,
			values:    map[string]bool{"a": true, "b": false, "d": true},
			wantCalls: []string{"a", "b", "d"},
		},
		{
			name: "specific predicate precedes earlier universal selector",
			configJSON: `[{"files":[
				"**/*",
				{"$rslintPredicate":"specific"}
			]}]`,
			values:    map[string]bool{"specific": true},
			wantCalls: []string{"specific"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := &evaluatorPredicateResolver{functions: make(map[string]func(string, bool) (bool, error))}
			for id, value := range test.values {
				resolver.functions[id] = func(string, bool) (bool, error) { return value, nil }
			}
			resolver.functions["unreached"] = func(string, bool) (bool, error) {
				return false, errors.New("unreached predicate executed")
			}
			evaluator := NewConfigEvaluator(mustDecodeEvaluatorConfig(t, test.configJSON), root, nil, resolver)
			resolution, err := evaluator.GetConfigForFile(context.Background(), filePath)
			if err != nil {
				t.Fatal(err)
			}
			if resolution.Status != ConfigFileConfigured || resolution.Config == nil {
				t.Fatalf("resolution = %+v, want configured", resolution)
			}
			if got := resolver.callIDs(); !reflect.DeepEqual(got, test.wantCalls) {
				t.Fatalf("predicate calls = %v, want %v", got, test.wantCalls)
			}
		})
	}
}

func TestConfigEvaluatorUsesMinimatchUTF16UnitsForFilesAndIgnores(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	filePath := tspath.CombinePaths(root, "😀")
	tests := []struct {
		name       string
		configJSON string
		want       ConfigFileStatus
	}{
		{
			name:       "one-unit files selector misses",
			configJSON: `[{"files":["?"]}]`,
			want:       ConfigFileUnconfigured,
		},
		{
			name:       "two-unit files selector matches",
			configJSON: `[{"files":["??"]}]`,
			want:       ConfigFileConfigured,
		},
		{
			name:       "one-unit global ignore misses",
			configJSON: `[{"ignores":["?"]},{"files":["??"]}]`,
			want:       ConfigFileConfigured,
		},
		{
			name:       "two-unit global ignore matches",
			configJSON: `[{"ignores":["??"]},{"files":["??"]}]`,
			want:       ConfigFileIgnored,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolution, err := NewConfigEvaluator(
				mustDecodeEvaluatorConfig(t, test.configJSON),
				root,
				nil,
				nil,
			).GetConfigForFile(context.Background(), filePath)
			if err != nil {
				t.Fatal(err)
			}
			if resolution.Status != test.want {
				t.Fatalf("status = %q, want %q", resolution.Status, test.want)
			}
		})
	}
}

func TestConfigEvaluatorIgnoreStateMachineAndUniversalFallback(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	filePath := tspath.CombinePaths(root, "x.ts")
	t.Run("global function ignore can be reopened then ignored again", func(t *testing.T) {
		config := mustDecodeEvaluatorConfig(t, `[
			{"ignores":[{"$rslintPredicate":"first"},"!x.ts",{"$rslintPredicate":"second"}]},
			{"files":["**/*.ts"]}
		]`)
		resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"first":  func(string, bool) (bool, error) { return true, nil },
			"second": func(string, bool) (bool, error) { return true, nil },
		}}
		resolution, err := NewConfigEvaluator(config, root, nil, resolver).GetConfigForFile(context.Background(), filePath)
		if err != nil {
			t.Fatal(err)
		}
		if resolution.Status != ConfigFileIgnored {
			t.Fatalf("status = %q, want ignored", resolution.Status)
		}
		if got, want := resolver.callIDs(), []string{"first", "second"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("predicate calls = %v, want %v", got, want)
		}
	})

	t.Run("positive string skips function until negation reopens", func(t *testing.T) {
		config := mustDecodeEvaluatorConfig(t, `[
			{"ignores":["x.ts",{"$rslintPredicate":"skipped"},"!x.ts"]},
			{"files":["**/*.ts"]}
		]`)
		resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"skipped": func(string, bool) (bool, error) { return false, errors.New("ignored function executed") },
		}}
		resolution, err := NewConfigEvaluator(config, root, nil, resolver).GetConfigForFile(context.Background(), filePath)
		if err != nil {
			t.Fatal(err)
		}
		if resolution.Status != ConfigFileConfigured {
			t.Fatalf("status = %q, want configured", resolution.Status)
		}
		if got := resolver.callIDs(); len(got) != 0 {
			t.Fatalf("predicate calls = %v, want none", got)
		}
	})

	t.Run("local ignore is evaluated again for universal fallback", func(t *testing.T) {
		config := mustDecodeEvaluatorConfig(t, `[
			{"files":["**/*.ts"],"name":"baseline"},
			{"files":["**/*.ts","**/*"],"ignores":[{"$rslintPredicate":"local"}]}
		]`)
		count := 0
		resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"local": func(string, bool) (bool, error) {
				count++
				return count == 1, nil
			},
		}}
		resolution, err := NewConfigEvaluator(config, root, nil, resolver).GetConfigForFile(context.Background(), filePath)
		if err != nil {
			t.Fatal(err)
		}
		if resolution.Status != ConfigFileConfigured {
			t.Fatalf("status = %q, want configured", resolution.Status)
		}
		if got, want := resolver.callIDs(), []string{"local", "local"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("predicate calls = %v, want %v", got, want)
		}
	})
}

func TestConfigEvaluatorBasePathCacheAndErrors(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	inside := tspath.CombinePaths(root, "sub", "x.custom")
	outside := tspath.CombinePaths(root, "other", "x.custom")
	config := mustDecodeEvaluatorConfig(t, `[{"basePath":"sub","files":[{"$rslintPredicate":"file"}]}]`)
	resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
		"file": func(string, bool) (bool, error) { return true, nil },
	}}
	evaluator := NewConfigEvaluator(config, root, nil, resolver)
	if resolution, err := evaluator.GetConfigForFile(context.Background(), outside); err != nil || resolution.Status != ConfigFileUnconfigured {
		t.Fatalf("outside resolution = %+v, err=%v", resolution, err)
	}
	if got := resolver.callIDs(); len(got) != 0 {
		t.Fatalf("outside basePath called predicates: %v", got)
	}
	for range 2 {
		resolution, err := evaluator.GetConfigForFile(context.Background(), inside)
		if err != nil || resolution.Status != ConfigFileConfigured {
			t.Fatalf("inside resolution = %+v, err=%v", resolution, err)
		}
	}
	resolver.mu.Lock()
	calls := append([]evaluatorPredicateCall(nil), resolver.calls...)
	resolver.mu.Unlock()
	if len(calls) != 1 || calls[0].path != inside || calls[0].directory {
		t.Fatalf("calls = %+v, want one absolute file call", calls)
	}

	t.Run("throwing query is not cached", func(t *testing.T) {
		failure := errors.New("predicate boom")
		throwing := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"file": func(string, bool) (bool, error) { return false, failure },
		}}
		throwingEvaluator := NewConfigEvaluator(config, root, nil, throwing)
		for range 2 {
			if _, err := throwingEvaluator.GetConfigForFile(context.Background(), inside); !errors.Is(err, failure) {
				t.Fatalf("error = %v, want predicate boom", err)
			}
		}
		if got, want := throwing.callIDs(), []string{"file", "file"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("predicate calls = %v, want %v", got, want)
		}
	})

	t.Run("concurrent exact queries coalesce", func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})
		var once sync.Once
		coalesced := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"file": func(string, bool) (bool, error) {
				once.Do(func() { close(started) })
				<-release
				return true, nil
			},
		}}
		coalescedEvaluator := NewConfigEvaluator(config, root, nil, coalesced)
		results := make(chan error, 2)
		for range 2 {
			go func() {
				resolution, err := coalescedEvaluator.GetConfigForFile(context.Background(), inside)
				if err == nil && resolution.Status != ConfigFileConfigured {
					err = errors.New("query was not configured")
				}
				results <- err
			}()
		}
		<-started
		close(release)
		for range 2 {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
		if got, want := coalesced.callIDs(), []string{"file"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("predicate calls = %v, want %v", got, want)
		}
	})
}

func TestConfigEvaluatorConcurrentWaitersRetryThrowingQueries(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())

	t.Run("file", func(t *testing.T) {
		filePath := tspath.CombinePaths(root, "x.ts")
		config := mustDecodeEvaluatorConfig(t, `[{"files":[{"$rslintPredicate":"file"}]}]`)
		started := make(chan struct{})
		release := make(chan struct{})
		var calls atomic.Int32
		resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"file": func(string, bool) (bool, error) {
				if calls.Add(1) == 1 {
					close(started)
					<-release
					return false, errors.New("first file call failed")
				}
				return true, nil
			},
		}}
		evaluator := NewConfigEvaluator(config, root, nil, resolver)
		results := make(chan error, 32)
		go func() {
			_, err := evaluator.GetConfigForFile(context.Background(), filePath)
			results <- err
		}()
		<-started
		launched := make(chan struct{}, 31)
		for range 31 {
			go func() {
				launched <- struct{}{}
				resolution, err := evaluator.GetConfigForFile(context.Background(), filePath)
				if err == nil && resolution.Status != ConfigFileConfigured {
					err = errors.New("retried file was not configured")
				}
				results <- err
			}()
		}
		for range 31 {
			<-launched
		}
		for range 100 {
			runtime.Gosched()
		}
		close(release)
		errorsSeen := 0
		for range 32 {
			if err := <-results; err != nil {
				errorsSeen++
			}
		}
		if errorsSeen != 1 || calls.Load() != 2 {
			t.Fatalf("errors=%d calls=%d, want one throw followed by one shared success", errorsSeen, calls.Load())
		}
	})

	t.Run("directory", func(t *testing.T) {
		directory := tspath.CombinePaths(root, "nested")
		config := mustDecodeEvaluatorConfig(t, `[{"ignores":[{"$rslintPredicate":"directory"}]}]`)
		started := make(chan struct{})
		release := make(chan struct{})
		var calls atomic.Int32
		resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
			"directory": func(string, bool) (bool, error) {
				if calls.Add(1) == 1 {
					close(started)
					<-release
					return false, errors.New("first directory call failed")
				}
				return false, nil
			},
		}}
		evaluator := NewConfigEvaluator(config, root, nil, resolver)
		results := make(chan error, 16)
		go func() {
			_, err := evaluator.IsDirectoryIgnored(context.Background(), directory)
			results <- err
		}()
		<-started
		launched := make(chan struct{}, 15)
		for range 15 {
			go func() {
				launched <- struct{}{}
				ignored, err := evaluator.IsDirectoryIgnored(context.Background(), directory)
				if err == nil && ignored {
					err = errors.New("retried directory was ignored")
				}
				results <- err
			}()
		}
		for range 15 {
			<-launched
		}
		for range 100 {
			runtime.Gosched()
		}
		close(release)
		errorsSeen := 0
		for range 16 {
			if err := <-results; err != nil {
				errorsSeen++
			}
		}
		if errorsSeen != 1 || calls.Load() != 2 {
			t.Fatalf("errors=%d calls=%d, want one throw followed by one shared success", errorsSeen, calls.Load())
		}
	})
}

func TestConfigEvaluatorCachesMergedConfigByMatchingIndices(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"base": "error"}},
		{Files: []string{"special/**"}, Rules: Rules{"special": "error"}},
	}
	evaluator := NewConfigEvaluator(config, root, nil, nil)
	resolve := func(relative string) *MergedConfig {
		t.Helper()
		resolution, err := evaluator.GetConfigForFile(context.Background(), tspath.CombinePaths(root, relative))
		if err != nil || resolution.Status != ConfigFileConfigured || resolution.Config == nil {
			t.Fatalf("resolution for %q = %+v, %v", relative, resolution, err)
		}
		return resolution.Config
	}
	first := resolve("a.ts")
	second := resolve("nested/b.ts")
	different := resolve("special/c.ts")
	if first != second {
		t.Fatal("identical matching-index sets did not share their immutable merged config")
	}
	if first == different {
		t.Fatal("different matching-index sets shared a merged config")
	}
}

type blockingGitignoreFS struct {
	vfs.FS
	root    string
	started chan string
	release <-chan struct{}
	mu      sync.Mutex
	reads   map[string]int
}

func (fsys *blockingGitignoreFS) ReadFile(path string) (string, bool) {
	path = tspath.NormalizePath(path)
	if tspath.GetBaseFileName(path) == ".gitignore" {
		fsys.mu.Lock()
		fsys.reads[path]++
		fsys.mu.Unlock()
		if tspath.GetDirectoryPath(path) != fsys.root && fsys.started != nil {
			fsys.started <- path
			<-fsys.release
		}
	}
	return fsys.FS.ReadFile(path)
}

func TestConfigEvaluatorGitignoreSourcesLoadConcurrentlyAndCoalesce(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	for _, directory := range []string{"a", "b"} {
		path := tspath.CombinePaths(root, directory)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, ".gitignore"), []byte("ignored.ts\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	release := make(chan struct{})
	fsys := &blockingGitignoreFS{
		FS: osvfs.FS(), root: root, started: make(chan string, 2), release: release,
		reads: make(map[string]int),
	}
	evaluator := NewConfigEvaluatorWithGitignore(RslintConfig{{Files: []string{"**/*.ts"}}}, root, fsys, nil)
	results := make(chan error, 2)
	for _, relative := range []string{"a/one.ts", "b/two.ts"} {
		go func() {
			_, err := evaluator.GetConfigForFile(context.Background(), tspath.CombinePaths(root, relative))
			results <- err
		}()
	}
	for range 2 {
		select {
		case <-fsys.started:
		case <-time.After(2 * time.Second):
			t.Fatal("different .gitignore source directories were serialized")
		}
	}
	close(release)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	fsys.mu.Lock()
	rootReads := fsys.reads[tspath.CombinePaths(root, ".gitignore")]
	fsys.mu.Unlock()
	if rootReads != 1 {
		t.Fatalf("shared root .gitignore reads = %d, want 1", rootReads)
	}

	// A fresh evaluator proves two distinct files in one source directory share
	// one in-flight/source result rather than reading the source twice.
	sharedRelease := make(chan struct{})
	sharedFS := &blockingGitignoreFS{
		FS: osvfs.FS(), root: root, started: make(chan string, 2), release: sharedRelease,
		reads: make(map[string]int),
	}
	shared := NewConfigEvaluatorWithGitignore(RslintConfig{{Files: []string{"**/*.ts"}}}, root, sharedFS, nil)
	sharedResults := make(chan error, 2)
	for _, relative := range []string{"a/one.ts", "a/two.ts"} {
		go func() {
			_, err := shared.GetConfigForFile(context.Background(), tspath.CombinePaths(root, relative))
			sharedResults <- err
		}()
	}
	<-sharedFS.started
	close(sharedRelease)
	for range 2 {
		if err := <-sharedResults; err != nil {
			t.Fatal(err)
		}
	}
	sharedFS.mu.Lock()
	reads := sharedFS.reads[tspath.CombinePaths(root, "a/.gitignore")]
	sharedFS.mu.Unlock()
	if reads != 1 {
		t.Fatalf("shared .gitignore reads = %d, want 1", reads)
	}
}

func TestConfigEvaluatorGitignoreAncestryReadsEachSourceOnce(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	for _, relative := range []string{"shared/a", "shared/b"} {
		if err := os.MkdirAll(tspath.CombinePaths(root, relative), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, relative := range []string{".gitignore", "shared/.gitignore", "shared/a/.gitignore", "shared/b/.gitignore"} {
		if err := os.WriteFile(tspath.CombinePaths(root, relative), []byte("ignored.ts\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fsys := &blockingGitignoreFS{FS: osvfs.FS(), root: root, reads: make(map[string]int)}
	evaluator := NewConfigEvaluatorWithGitignore(RslintConfig{{Files: []string{"**/*.ts"}}}, root, fsys, nil)
	var waitGroup sync.WaitGroup
	for _, relative := range []string{"shared/a/one.ts", "shared/a/two.ts", "shared/b/three.ts"} {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if _, err := evaluator.GetConfigForFile(context.Background(), tspath.CombinePaths(root, relative)); err != nil {
				t.Errorf("GetConfigForFile(%q): %v", relative, err)
			}
		}()
	}
	waitGroup.Wait()

	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	for _, relative := range []string{".gitignore", "shared/.gitignore", "shared/a/.gitignore", "shared/b/.gitignore"} {
		path := tspath.CombinePaths(root, relative)
		if reads := fsys.reads[path]; reads != 1 {
			t.Fatalf("%s reads = %d, want 1", relative, reads)
		}
	}
}

func TestConfigEvaluatorDirectoryCacheMatchesConfigArray(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	config := mustDecodeEvaluatorConfig(t, `[
		{"ignores":[{"$rslintPredicate":"directory"}]},
		{"files":["**/*.ts"]}
	]`)
	resolver := &evaluatorPredicateResolver{functions: map[string]func(string, bool) (bool, error){
		"directory": func(_ string, directory bool) (bool, error) {
			if !directory {
				return false, errors.New("unexpected file call")
			}
			return false, nil
		},
	}}
	evaluator := NewConfigEvaluator(config, root, nil, resolver)
	a := tspath.CombinePaths(root, "a")
	b := tspath.CombinePaths(a, "b")
	c := tspath.CombinePaths(a, "c")
	for _, directory := range []string{b, c, b} {
		ignored, err := evaluator.IsDirectoryIgnored(context.Background(), directory)
		if err != nil || ignored {
			t.Fatalf("IsDirectoryIgnored(%q) = %t, %v", directory, ignored, err)
		}
	}
	resolver.mu.Lock()
	calls := append([]evaluatorPredicateCall(nil), resolver.calls...)
	resolver.mu.Unlock()
	wantPaths := []string{a, b, a, c}
	gotPaths := make([]string, len(calls))
	for index, call := range calls {
		if !call.directory {
			t.Fatalf("call %d was not marked as a directory: %+v", index, call)
		}
		gotPaths[index] = call.path
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("directory calls = %v, want %v", gotPaths, wantPaths)
	}
}

func TestConfigEvaluatorWithGitignoreUsesExactAncestryAndOrderedReinclude(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "src", "keep.ts")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("src/*.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root = tspath.NormalizePath(root)
	target = tspath.NormalizePath(target)
	baseline := RslintConfig{{Files: []string{"**/*.ts"}}}
	resolution, err := NewConfigEvaluatorWithGitignore(baseline, root, osvfs.FS(), nil).
		GetConfigForFile(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != ConfigFileIgnored {
		t.Fatalf("gitignored status = %q, want ignored", resolution.Status)
	}

	reopened := append(append(RslintConfig(nil), baseline...), ConfigEntry{Ignores: []string{"!src/keep.ts"}})
	resolution, err = NewConfigEvaluatorWithGitignore(reopened, root, osvfs.FS(), nil).
		GetConfigForFile(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != ConfigFileConfigured {
		t.Fatalf("authored negation status = %q, want configured", resolution.Status)
	}
}
