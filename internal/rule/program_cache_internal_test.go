package rule

import (
	"runtime"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

type programCacheTestKey struct{}

// TestCachedByProgramReusesWithinOneProgram locks in the point of the cache:
// two callers asking about the same Program under the same key get the value
// the first of them built.
func TestCachedByProgramReusesWithinOneProgram(t *testing.T) {
	program := programCacheTestProgram(t)

	builds := 0
	build := func() *int {
		builds++
		value := builds
		return &value
	}

	first := CachedByProgram(program, programCacheTestKey{}, build)
	second := CachedByProgram(program, programCacheTestKey{}, build)

	if builds != 1 {
		t.Fatalf("build ran %d times, want 1", builds)
	}
	if first != second {
		t.Fatalf("second call returned a different value")
	}
	runtime.KeepAlive(program)
}

// TestCachedByProgramReleasesCollectedPrograms locks in that an entry never
// outlives the Program it belongs to. The cache is reachable for the life of
// the process, so an entry that kept its Program alive would hold every source
// file of every Program the process ever linted.
func TestCachedByProgramReleasesCollectedPrograms(t *testing.T) {
	var key weak.Pointer[compiler.Program]

	// The Program is confined to this call so that nothing in the test frame
	// keeps it reachable afterwards.
	func() {
		program := programCacheTestProgram(t)
		key = weak.Make(program)
		CachedByProgram(program, programCacheTestKey{}, func() *int {
			value := 1
			return &value
		})
		if _, ok := programCaches.Load(key); !ok {
			t.Fatal("CachedByProgram stored no entry for the program")
		}
	}()

	// Cleanups run on a collection and then on their own goroutine, so the
	// entry disappears shortly after the Program becomes unreachable rather
	// than at a point this test can name.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := programCaches.Load(key); !ok {
			return
		}
		runtime.GC()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the cache entry outlived its program")
}

func programCacheTestProgram(t *testing.T) *compiler.Program {
	t.Helper()

	files := map[string]string{"/program-cache-fixture/file.ts": "export const value = 1;\n"}
	fs := rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := rslint_utils.CreateCompilerHost("/", fs)
	program, err := rslint_utils.CreateProgramFromOptions(true, &core.CompilerOptions{}, []string{"/program-cache-fixture/file.ts"}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}
	return program
}
