package utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
}

func newBlockingRealpathQuery() *blockingRealpathQuery {
	return &blockingRealpathQuery{
		started: make(chan struct{}, 64),
		release: make(chan struct{}),
	}
}

func (q *blockingRealpathQuery) run(path string) string {
	q.calls.Add(1)
	q.started <- struct{}{}
	<-q.release
	return path
}

func TestParallelProgramFSCoalescesConcurrentRealpathMisses(t *testing.T) {
	t.Parallel()

	gate := newBlockingRealpathQuery()
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

	select {
	case <-gate.started:
	case <-time.After(2 * time.Second):
		t.Fatal("underlying Realpath did not start")
	}
	// Keep the owner blocked long enough for every ready goroutine to reach
	// the cache. A cache stampede would be visible in calls.
	time.Sleep(50 * time.Millisecond)
	if got := gate.calls.Load(); got != 1 {
		t.Fatalf("underlying calls while cold query is blocked = %d, want 1", got)
	}
	close(gate.release)

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
}

func TestParallelProgramFSDoesNotSerializeDifferentPaths(t *testing.T) {
	t.Parallel()

	gate := newBlockingRealpathQuery()
	cached := newParallelProgramFS(&parallelProgramTestFS{realpath: gate.run})
	results := make(chan string, 2)
	go func() { results <- cached.Realpath("/first") }()
	go func() { results <- cached.Realpath("/second") }()

	for range 2 {
		select {
		case <-gate.started:
		case <-time.After(2 * time.Second):
			t.Fatal("different paths were serialized behind one query")
		}
	}
	close(gate.release)
	for range 2 {
		<-results
	}
}

func TestParallelProgramFSPanicReleasesWaitersAndRetries(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	firstStarted := make(chan struct{})
	panicNow := make(chan struct{})
	base := &parallelProgramTestFS{
		realpath: func(path string) string {
			if calls.Add(1) == 1 {
				close(firstStarted)
				<-panicNow
				panic("injected filesystem panic")
			}
			return path
		},
	}
	cached := newParallelProgramFS(base)
	ownerPanicked := make(chan bool, 1)
	go func() {
		panicked := false
		func() {
			defer func() { panicked = recover() != nil }()
			cached.Realpath("/panic")
		}()
		ownerPanicked <- panicked
	}()
	<-firstStarted

	waiterResult := make(chan string, 1)
	go func() { waiterResult <- cached.Realpath("/panic") }()
	time.Sleep(25 * time.Millisecond)
	close(panicNow)

	if !<-ownerPanicked {
		t.Fatal("owner panic was not preserved")
	}
	select {
	case result := <-waiterResult:
		if result != "/panic" {
			t.Fatalf("retry after panic = %q", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("panic stranded a concurrent waiter")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("underlying calls after retry = %d, want 2", got)
	}
	if cached.Realpath("/panic") != "/panic" || calls.Load() != 2 {
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
