package utils

// cspell:ignore synctest

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/microsoft/typescript-go/shim/vfs"
)

type parallelProgramTestFS struct {
	vfs.FS
	directoryExists func(string) bool
	fileExists      func(string) bool
	readFile        func(string) (string, bool)
	realpath        func(string) string
	writeFile       func(string, string) error
}

func (f *parallelProgramTestFS) DirectoryExists(path string) bool {
	return f.directoryExists(path)
}

func (f *parallelProgramTestFS) FileExists(path string) bool {
	return f.fileExists(path)
}

func (f *parallelProgramTestFS) ReadFile(path string) (string, bool) {
	return f.readFile(path)
}

func (f *parallelProgramTestFS) Realpath(path string) string {
	return f.realpath(path)
}

func (f *parallelProgramTestFS) WriteFile(path string, content string) error {
	return f.writeFile(path, content)
}

type blockingRealpathQuery struct {
	calls       atomic.Int32
	release     chan struct{}
	releaseOnce sync.Once
}

func newBlockingRealpathQuery() *blockingRealpathQuery {
	return &blockingRealpathQuery{
		release: make(chan struct{}),
	}
}

func (q *blockingRealpathQuery) run(path string) string {
	q.calls.Add(1)
	<-q.release
	return path
}

func (q *blockingRealpathQuery) releaseAll() {
	q.releaseOnce.Do(func() { close(q.release) })
}

func TestParallelProgramFSCoalescesConcurrentRealpathMisses(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		gate := newBlockingRealpathQuery()
		defer gate.releaseAll()
		cached := newParallelProgramFS(&parallelProgramTestFS{realpath: gate.run})

		const workers = 32
		start := make(chan struct{})
		var ready sync.WaitGroup
		ready.Add(workers)
		results := make(chan string, workers)
		for range workers {
			go func() {
				ready.Done()
				<-start
				results <- cached.Realpath("/same/exact/path")
			}()
		}
		ready.Wait()
		close(start)

		// Wait until the owner is blocked in the underlying filesystem and all
		// other goroutines are blocked on its flight. This observes the same
		// state as a sleep, without depending on host scheduling speed.
		synctest.Wait()
		if got := gate.calls.Load(); got != 1 {
			t.Fatalf("underlying calls while cold query is blocked = %d, want 1", got)
		}
		gate.releaseAll()

		for range workers {
			if got := <-results; got != "/same/exact/path" {
				t.Fatalf("Realpath result = %q", got)
			}
		}
		if got := cached.Realpath("/same/exact/path"); got != "/same/exact/path" {
			t.Fatalf("warm Realpath result = %q", got)
		}
		if got := gate.calls.Load(); got != 1 {
			t.Fatalf("underlying calls after warm query = %d, want 1", got)
		}
	})
}

func TestParallelProgramFSDoesNotSerializeDifferentPaths(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		gate := newBlockingRealpathQuery()
		defer gate.releaseAll()
		cached := newParallelProgramFS(&parallelProgramTestFS{realpath: gate.run})
		results := make(chan string, 2)
		go func() { results <- cached.Realpath("/first") }()
		go func() { results <- cached.Realpath("/second") }()

		// Both goroutines must independently reach the underlying filesystem
		// before either query is released. synctest makes that observation
		// independent of runner load while retaining the concurrency assertion.
		synctest.Wait()
		if got := gate.calls.Load(); got != 2 {
			t.Fatalf("concurrent underlying calls for different paths = %d, want 2", got)
		}
		gate.releaseAll()

		seen := make(map[string]bool, 2)
		for range 2 {
			seen[<-results] = true
		}
		if !seen["/first"] || !seen["/second"] || len(seen) != 2 {
			t.Fatalf("Realpath results for different paths = %v", seen)
		}
	})
}

func TestParallelProgramFSPanicReleasesWaitersAndRetries(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		testParallelProgramFSPanicRemovesFailedFlight(t)
		testParallelProgramFSPanicReleasesWaiterAndRetries(t)
	})
}

func testParallelProgramFSPanicRemovesFailedFlight(t *testing.T) {
	const (
		path          = "/panic-probe"
		injectedPanic = "injected filesystem panic"
	)

	cached := newParallelProgramFS(&parallelProgramTestFS{
		realpath: func(string) string { panic(injectedPanic) },
	})
	var ownerPanic any
	func() {
		defer func() { ownerPanic = recover() }()
		cached.Realpath(path)
	}()

	if ownerPanic != injectedPanic {
		t.Errorf("owner panic = %v, want %q", ownerPanic, injectedPanic)
	}
	if entry, loaded := cached.realpathCache.Load(path); loaded {
		t.Errorf("panicking owner retained failed cache entry %T", entry)
		cached.realpathCache.Delete(path)
	}
}

func testParallelProgramFSPanicReleasesWaiterAndRetries(t *testing.T) {
	const (
		path          = "/panic-waiter"
		injectedPanic = "injected filesystem panic"
	)

	var calls atomic.Int32
	firstStarted := make(chan struct{})
	panicNow := make(chan struct{})
	ownerReleased := false
	defer func() {
		if !ownerReleased {
			close(panicNow)
		}
	}()
	base := &parallelProgramTestFS{
		realpath: func(path string) string {
			if calls.Add(1) == 1 {
				close(firstStarted)
				<-panicNow
				panic(injectedPanic)
			}
			return path
		},
	}
	cached := newParallelProgramFS(base)
	ownerPanic := make(chan any, 1)
	go func() {
		defer func() { ownerPanic <- recover() }()
		cached.Realpath(path)
	}()
	<-firstStarted

	entry, loaded := cached.realpathCache.Load(path)
	if !loaded {
		t.Fatal("owner flight is missing from the cache")
	}
	ownerFlight, ok := entry.(*realpathFlight)
	if !ok {
		t.Fatalf("owner cache entry = %T, want *realpathFlight", entry)
	}

	waiterResult := make(chan string, 1)
	go func() { waiterResult <- cached.Realpath(path) }()
	// Once the bubble is idle, an empty result, one underlying call, and the
	// original cache entry prove that the waiter is blocked on ownerFlight.
	synctest.Wait()

	waiterReturnedEarly := false
	var result string
	select {
	case result = <-waiterResult:
		waiterReturnedEarly = true
		t.Errorf("waiter returned before owner panic: %q", result)
	default:
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("underlying calls before owner panic = %d, want 1", got)
	}
	if actual, stillLoaded := cached.realpathCache.Load(path); !stillLoaded {
		t.Error("owner flight disappeared before owner panic")
	} else if actual != ownerFlight {
		t.Errorf("cache entry before owner panic = %T, want original owner flight", actual)
	}

	// The probe above verifies that production panic cleanup deletes a failed
	// flight. Delete this phase's flight first so this waiter isolates and tests
	// the cleanup's release-and-retry behavior without a broken delete making it spin forever.
	if !cached.realpathCache.CompareAndDelete(path, ownerFlight) {
		t.Error("failed to remove owner flight before panic")
		cached.realpathCache.Delete(path)
	}
	close(panicNow)
	ownerReleased = true
	if got := <-ownerPanic; got != injectedPanic {
		t.Errorf("owner panic = %v, want %q", got, injectedPanic)
	}

	if !waiterReturnedEarly {
		synctest.Wait()
		select {
		case result = <-waiterResult:
		default:
			t.Fatal("panic stranded a concurrent waiter")
		}
	}
	if result != path {
		t.Fatalf("retry after panic = %q", result)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("underlying calls after retry = %d, want 2", got)
	}
	cachedValue, loaded := cached.realpathCache.Load(path)
	if !loaded {
		t.Fatal("successful retry was not cached")
	}
	if value, ok := cachedValue.(string); !ok || value != path {
		t.Fatalf("cache entry after retry = %#v (%T), want plain string %q", cachedValue, cachedValue, path)
	}
	if cached.Realpath(path) != path || calls.Load() != 2 {
		t.Fatal("successful retry was not cached")
	}
}

func TestParallelProgramFSUsesExactCrossPlatformPathKeys(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	calls := make(map[string]int)
	cached := newParallelProgramFS(&parallelProgramTestFS{
		realpath: func(path string) string {
			mu.Lock()
			calls[path]++
			mu.Unlock()
			return path
		},
	})
	paths := []string{
		"/workspace/A.ts",
		"/workspace/a.ts",
		"/workspace/./a.ts",
		"C:/workspace/a.ts",
		`C:\workspace\a.ts`,
		"//server/share/a.ts",
		`\\server\share\a.ts`,
	}
	for _, path := range paths {
		if got := cached.Realpath(path); got != path {
			t.Fatalf("cold Realpath(%q) = %q", path, got)
		}
		if got := cached.Realpath(path); got != path {
			t.Fatalf("warm Realpath(%q) = %q", path, got)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	for _, path := range paths {
		if got := calls[path]; got != 1 {
			t.Errorf("underlying calls for exact key %q = %d, want 1", path, got)
		}
	}
}

func TestParallelProgramFSCachesEmptyRealpath(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	cached := newParallelProgramFS(&parallelProgramTestFS{
		realpath: func(string) string {
			calls.Add(1)
			return ""
		},
	})
	if got := cached.Realpath("/missing"); got != "" {
		t.Fatalf("cold Realpath = %q, want empty", got)
	}
	if got := cached.Realpath("/missing"); got != "" {
		t.Fatalf("warm Realpath = %q, want empty", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("underlying calls = %d, want 1", got)
	}
}

func TestParallelProgramFSPreservesOtherVFSMethods(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	exists := false
	fileExistsCalls := 0
	readCalls := 0
	base := &parallelProgramTestFS{
		fileExists: func(string) bool {
			mu.Lock()
			defer mu.Unlock()
			fileExistsCalls++
			return exists
		},
		readFile: func(string) (string, bool) {
			mu.Lock()
			defer mu.Unlock()
			readCalls++
			return "content", true
		},
		writeFile: func(string, string) error {
			mu.Lock()
			exists = true
			mu.Unlock()
			return nil
		},
	}
	cached := newParallelProgramFS(base)

	if cached.FileExists("/mutable") {
		t.Fatal("initial FileExists returned true")
	}
	if err := cached.WriteFile("/mutable", "content"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !cached.FileExists("/mutable") {
		t.Fatal("FileExists was not delegated after WriteFile")
	}
	cached.ReadFile("/mutable")
	cached.ReadFile("/mutable")

	mu.Lock()
	defer mu.Unlock()
	if fileExistsCalls != 2 || readCalls != 2 {
		t.Fatalf("underlying calls: FileExists=%d ReadFile=%d", fileExistsCalls, readCalls)
	}
}
