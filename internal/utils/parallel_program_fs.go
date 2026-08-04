package utils

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// parallelProgramFS coalesces Realpath queries that concurrent TypeScript
// Program builds issue for the same cold path. This view retains completed
// Realpath results for its request lifetime; every other operation remains the
// responsibility of the request's existing VFS/cache stack.
//
// Keys are deliberately left in the caller's path space. In particular, this
// layer never cleans, case-folds, or resolves a path before caching it, so
// lexical, casing, separator, UNC, symlink, and overlay identities remain the
// responsibility of the wrapped VFS.
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

func (f *parallelProgramFS) Realpath(path string) string {
	for {
		if cached, ok := f.realpathCache.Load(path); ok {
			if value, completed := completedRealpath(cached); completed {
				return value
			}
			flight := mustRealpathFlight(cached)
			flight.ready.Wait()
			if flight.completed {
				return flight.value
			}
			// A panicking owner removes its entry and releases waiters without
			// publishing a value. Retry so one waiter can become the next owner.
			continue
		}

		candidate := &realpathFlight{}
		candidate.ready.Add(1)
		actual, loaded := f.realpathCache.LoadOrStore(path, candidate)
		if !loaded {
			return f.queryAndPublishRealpath(path, candidate)
		}
		if value, completed := completedRealpath(actual); completed {
			return value
		}
		flight := mustRealpathFlight(actual)
		flight.ready.Wait()
		if flight.completed {
			return flight.value
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

	value = f.FS.Realpath(path)
	flight.value = value
	flight.completed = true
	// Replace the transient flight with the plain value so warm reads have
	// the same one-lookup shape and retained allocation profile as cachedvfs.
	f.realpathCache.Store(path, value)
	completed = true
	flight.ready.Done()
	return value
}
