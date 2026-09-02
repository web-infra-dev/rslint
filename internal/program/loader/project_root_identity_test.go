package loader

import (
	"runtime"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

func TestFreezeProgramRootIdentitiesUsesTargetHintsAndCanonicalUnion(t *testing.T) {
	const (
		aliasA    = "/repo/a.ts"
		physicalA = "/physical/a.ts"
		aliasB    = "/repo/b.ts"
		physicalB = "/physical/b.ts"
		missing   = "/repo/missing.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{aliasA, physicalA, aliasB},
		map[string]string{aliasB: physicalB},
	)
	first := createBindingIndexTestProgram(t, fsys, aliasA, aliasB)
	second := createBindingIndexTestProgram(t, fsys, physicalA, missing)
	fsys.resetCalls()

	identities, err := NewSession(fsys).FreezeProgramRootIdentities(
		[]*lintprogram.Program{
			lintprogram.NewFromCompiler(first),
			lintprogram.NewFromCompiler(second),
		},
		[]target.File{{PathIdentity: rslintconfig.PathIdentity{
			Path:          aliasA,
			CanonicalPath: physicalA,
		}}},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		exactPathID(physicalA),
		exactPathID(physicalB),
		exactPathID(missing),
	}
	if !slices.Equal(identities, want) {
		t.Fatalf("identities = %q, want %q", identities, want)
	}
	if calls := fsys.callCount(aliasA); calls != 0 {
		t.Fatalf("frozen target identity was re-resolved %d time(s)", calls)
	}
	if calls := fsys.callCount(aliasB); calls != 1 {
		t.Fatalf("root without a target hint used Realpath %d time(s), want 1", calls)
	}
}

func TestFreezeProgramRootIdentitiesSkipsSourceOnlyPrograms(t *testing.T) {
	const (
		typeCheckedRoot = "/repo/type-checked.ts"
		sourceOnlyRoot  = "/repo/source-only.ts"
	)
	fsys := newBindingIndexTestFS([]string{typeCheckedRoot, sourceOnlyRoot}, nil)
	typeChecked := createBindingIndexTestProgram(t, fsys, typeCheckedRoot)
	sourceOnlyCompiler := createBindingIndexTestProgram(t, fsys, sourceOnlyRoot)
	sourceOnly, err := lintprogram.NewFromBoundSources(
		sourceOnlyCompiler,
		sourceOnlyCompiler.SourceFiles(),
	)
	if err != nil {
		t.Fatal(err)
	}
	zeroValue := &lintprogram.Program{}
	fsys.resetCalls()

	identities, err := NewSession(fsys).FreezeProgramRootIdentities(
		[]*lintprogram.Program{
			nil,
			zeroValue,
			sourceOnly,
			lintprogram.NewFromCompiler(typeChecked),
		},
		nil,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{exactPathID(typeCheckedRoot)}
	if !slices.Equal(identities, want) {
		t.Fatalf("identities = %q, want only type-capable Program root %q", identities, want)
	}
	if calls := fsys.callCount(sourceOnlyRoot); calls != 0 {
		t.Fatalf("source-only root was resolved %d time(s), want 0", calls)
	}
}

type projectRootDirectoryAccessFS struct {
	*bindingIndexTestFS

	mu    sync.Mutex
	calls int
}

func (fsys *projectRootDirectoryAccessFS) GetAccessibleEntries(directoryPath string) vfs.Entries {
	fsys.mu.Lock()
	fsys.calls++
	fsys.mu.Unlock()
	return fsys.bindingIndexTestFS.GetAccessibleEntries(directoryPath)
}

func (fsys *projectRootDirectoryAccessFS) accessibleEntriesCalls() int {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	return fsys.calls
}

func TestProjectRootIdentityResolverNeverEnumeratesDirectories(t *testing.T) {
	const rootCount = 2048
	rootPaths := make([]string, rootCount)
	realPaths := make(map[string]string, rootCount)
	fileNames := make([]string, 10_000)
	for index := range fileNames {
		fileNames[index] = "unrelated-" + strconv.Itoa(index) + ".ts"
	}
	for index := range rootPaths {
		fileName := "root-" + strconv.Itoa(index) + ".ts"
		rootPaths[index] = "/repo/" + fileName
		realPaths[rootPaths[index]] = "/physical/" + fileName
	}
	base := newBindingIndexTestFS(rootPaths, realPaths)
	base.entries["/repo"] = vfs.Entries{
		Files:    fileNames,
		Symlinks: map[string]struct{}{},
	}
	fsys := &projectRootDirectoryAccessFS{bindingIndexTestFS: base}
	fsys.resetCalls()

	identities := newProjectRootIdentityResolver(nil, fsys, false).canonicalRootPathIDs(rootPaths)
	if len(identities) != rootCount ||
		identities[0] != exactPathID("/physical/root-0.ts") ||
		identities[rootCount-1] != exactPathID("/physical/root-2047.ts") {
		t.Fatalf("root identities were not projected correctly")
	}
	if calls := fsys.accessibleEntriesCalls(); calls != 0 {
		t.Fatalf("root projection enumerated directories %d time(s), want 0", calls)
	}
	for _, index := range []int{0, rootCount / 2, rootCount - 1} {
		if calls := fsys.callCount(rootPaths[index]); calls != 1 {
			t.Fatalf("root %q used Realpath %d time(s), want 1", rootPaths[index], calls)
		}
	}
}

func TestProjectRootIdentityResolverUsesPerFileVFSIdentity(t *testing.T) {
	tests := []struct {
		name      string
		rootPaths []string
		realPaths map[string]string
		want      []string
	}{
		{
			name:      "file symlink and regular file",
			rootPaths: []string{"/repo/alias.ts", "/repo/regular.ts"},
			realPaths: map[string]string{
				"/repo/alias.ts":   "/elsewhere/alias.ts",
				"/repo/regular.ts": "/physical/regular.ts",
			},
			want: []string{exactPathID("/elsewhere/alias.ts"), exactPathID("/physical/regular.ts")},
		},
		{
			name:      "missing roots retain lexical identity",
			rootPaths: []string{"/repo/a.ts", "/repo/b.ts"},
			want:      []string{exactPathID("/repo/a.ts"), exactPathID("/repo/b.ts")},
		},
		{
			name:      "VFS-defined mapping is authoritative",
			rootPaths: []string{"/virtual/a.ts", "/virtual/b.ts"},
			realPaths: map[string]string{
				"/virtual/a.ts": "/overlay/a.ts",
				"/virtual/b.ts": "/pinned/b.ts",
			},
			want: []string{exactPathID("/overlay/a.ts"), exactPathID("/pinned/b.ts")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := newBindingIndexTestFS(test.rootPaths, test.realPaths)
			fsys.resetCalls()

			identities := newProjectRootIdentityResolver(nil, fsys, true).canonicalRootPathIDs(test.rootPaths)
			if !slices.Equal(identities, test.want) {
				t.Fatalf("identities = %q, want %q", identities, test.want)
			}
			for _, rootPath := range test.rootPaths {
				if calls := fsys.callCount(rootPath); calls != 1 {
					t.Fatalf("root %q resolved %d time(s), want 1", rootPath, calls)
				}
			}
		})
	}
}

type concurrentProjectRootIdentityFS struct {
	*bindingIndexTestFS

	concurrencyMu sync.Mutex
	active        int
	maxActive     int
	targetActive  int
	allActive     chan struct{}
	allActiveOnce sync.Once
	release       chan struct{}
}

func (fsys *concurrentProjectRootIdentityFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	fsys.concurrencyMu.Lock()
	fsys.active++
	fsys.maxActive = max(fsys.maxActive, fsys.active)
	if fsys.active >= fsys.targetActive {
		fsys.allActiveOnce.Do(func() { close(fsys.allActive) })
	}
	fsys.concurrencyMu.Unlock()

	<-fsys.release

	fsys.concurrencyMu.Lock()
	fsys.active--
	fsys.concurrencyMu.Unlock()
	return "/physical/" + tspath.GetBaseFileName(filePath)
}

func TestProjectRootIdentityResolverUsesBoundedWorkerPool(t *testing.T) {
	rootCount := projectRootIdentityWorkerLimit + 32
	rootPaths := make([]string, rootCount)
	for index := range rootPaths {
		rootPaths[index] = "/repo/root-" + strconv.Itoa(index) + ".ts"
	}
	wantWorkers := min(
		runtime.GOMAXPROCS(0)*projectRootIdentityWorkersPerProcessor,
		projectRootIdentityWorkerLimit,
		rootCount,
	)
	fsys := &concurrentProjectRootIdentityFS{
		bindingIndexTestFS: newBindingIndexTestFS(rootPaths, nil),
		targetActive:       wantWorkers,
		allActive:          make(chan struct{}),
		release:            make(chan struct{}),
	}
	done := make(chan []string, 1)
	go func() {
		done <- newProjectRootIdentityResolver(nil, fsys, false).canonicalRootPathIDs(rootPaths)
	}()

	reachedBound := false
	select {
	case <-fsys.allActive:
		reachedBound = true
	case <-time.After(2 * time.Second):
	}
	close(fsys.release)
	identities := <-done
	fsys.concurrencyMu.Lock()
	maxActive := fsys.maxActive
	fsys.concurrencyMu.Unlock()
	if !reachedBound || maxActive != wantWorkers {
		t.Fatalf("maximum concurrent roots = %d, want %d", maxActive, wantWorkers)
	}
	if len(identities) != rootCount || identities[0] != exactPathID("/physical/root-0.ts") {
		t.Fatalf("root identities were not resolved correctly")
	}
}

func TestProjectRootIdentityWorkHonorsSingleThreaded(t *testing.T) {
	var active atomic.Int64
	var maxActive atomic.Int64
	runProjectRootIdentityWork(true, 32, func(_ int) {
		current := active.Add(1)
		for observed := maxActive.Load(); current > observed && !maxActive.CompareAndSwap(observed, current); {
			observed = maxActive.Load()
		}
		time.Sleep(100 * time.Microsecond)
		active.Add(-1)
	})
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("single-threaded maximum concurrency = %d, want 1", got)
	}
}

func BenchmarkProjectRootIdentityResolver(b *testing.B) {
	const rootCount = 10_000
	rootPaths := make([]string, rootCount)
	for index := range rootPaths {
		rootPaths[index] = "/repo/root-" + strconv.Itoa(index) + ".ts"
	}
	fsys := newBindingIndexTestFS(rootPaths, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		newProjectRootIdentityResolver(nil, fsys, false).canonicalRootPathIDs(rootPaths)
	}
}

func TestFreezeProgramRootIdentitiesProtectsCanonicalTargetIdentity(t *testing.T) {
	const (
		rootPath       = "/repo/root.ts"
		otherCanonical = "/physical/other.ts"
	)
	fsys := newBindingIndexTestFS([]string{rootPath}, nil)
	program := createBindingIndexTestProgram(t, fsys, rootPath)

	identities, err := NewSession(fsys).FreezeProgramRootIdentities(
		[]*lintprogram.Program{lintprogram.NewFromCompiler(program)},
		[]target.File{
			{PathIdentity: rslintconfig.PathIdentity{Path: rootPath, CanonicalPath: otherCanonical}},
			{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/other.ts", CanonicalPath: rootPath}},
		},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{exactPathID(rootPath)}
	if !slices.Equal(identities, want) {
		t.Fatalf("identities = %q, want canonical target self-identity %q", identities, want)
	}
}

func TestFreezeProgramRootIdentitiesResolvesWithoutTargetPlan(t *testing.T) {
	const (
		aliasPath    = "/repo/root.ts"
		physicalPath = "/physical/root.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{aliasPath},
		map[string]string{aliasPath: physicalPath},
	)
	program := createBindingIndexTestProgram(t, fsys, aliasPath)
	fsys.resetCalls()

	identities, err := NewSession(fsys).FreezeProgramRootIdentities(
		[]*lintprogram.Program{lintprogram.NewFromCompiler(program)},
		nil,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{exactPathID(physicalPath)}
	if !slices.Equal(identities, want) {
		t.Fatalf("identities = %q, want %q", identities, want)
	}
	if calls := fsys.callCount(aliasPath); calls != 1 {
		t.Fatalf("root without a target hint resolved %d times, want 1", calls)
	}
}
