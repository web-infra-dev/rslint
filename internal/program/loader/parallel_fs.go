package loader

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// parallelProgramFS coalesces Realpath queries that concurrent TypeScript
// Program builds issue for the same cold path. This view retains completed
// Realpath results for its request lifetime; every other operation remains the
// responsibility of the request's existing VFS/cache stack.
//
// Keys retain the caller's normalized rooted spelling. This layer never
// case-folds or resolves a path before caching it, so casing, UNC, symlink,
// and overlay identities remain the responsibility of the wrapped VFS.
type parallelProgramFS struct {
	vfs.FS
	realpathCache sync.Map // exact path string -> string or *realpathFlight
}

type realpathFlight struct {
	ready     sync.WaitGroup
	value     string
	completed bool
}

var _ vfs.FS = (*parallelProgramFS)(nil)

func newParallelProgramFS(fs vfs.FS) *parallelProgramFS {
	return &parallelProgramFS{FS: fs}
}

// SourceHasBOM forwards to the wrapped VFS. Embedding vfs.FS promotes only
// that interface's methods, so this layer has to pass [BOMSource] through by
// hand or every overlay identity below it becomes invisible.
func (f *parallelProgramFS) SourceHasBOM(path string) bool {
	return utils.SourceHasBOM(f.FS, path)
}

func (f *parallelProgramFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	for {
		if cached, ok := f.realpathCache.Load(path.AsString()); ok {
			if value, completed := completedRealpath(cached); completed {
				return pathutil.Path(value)
			}
			flight := mustRealpathFlight(cached)
			flight.ready.Wait()
			if flight.completed {
				return pathutil.Path(flight.value)
			}
			// A panicking owner removes its entry and releases waiters without
			// publishing a value. Retry so one waiter can become the next owner.
			continue
		}

		candidate := &realpathFlight{}
		candidate.ready.Add(1)
		actual, loaded := f.realpathCache.LoadOrStore(path.AsString(), candidate)
		if !loaded {
			return pathutil.Path(f.queryAndPublishRealpath(path.AsString(), candidate))
		}
		if value, completed := completedRealpath(actual); completed {
			return pathutil.Path(value)
		}
		flight := mustRealpathFlight(actual)
		flight.ready.Wait()
		if flight.completed {
			return pathutil.Path(flight.value)
		}
	}
}

func completedRealpath(value any) (string, bool) {
	result, ok := value.(string)
	return result, ok
}

func mustRealpathFlight(value any) *realpathFlight {
	flight, ok := value.(*realpathFlight)
	if !ok {
		panic("parallel Program realpath cache contains an unexpected value")
	}
	return flight
}

func (f *parallelProgramFS) queryAndPublishRealpath(path string, flight *realpathFlight) (value string) {
	completed := false
	defer func() {
		if !completed {
			// Preserve the panic while making the path immediately retryable
			// and ensuring concurrent waiters cannot be stranded.
			f.realpathCache.CompareAndDelete(path, flight)
			flight.ready.Done()
		}
	}()

	value = f.FS.Realpath(pathutil.Path(path)).AsString()
	flight.value = value
	flight.completed = true
	// Replace the transient flight with the plain value so warm reads have
	// the same one-lookup shape and retained allocation profile as cachedvfs.
	f.realpathCache.Store(path, value)
	completed = true
	flight.ready.Done()
	return value
}
