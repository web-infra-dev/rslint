package loader

import (
	"runtime"
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

const (
	projectRootIdentityWorkersPerProcessor = 16
	projectRootIdentityWorkerLimit         = 256
)

// projectRootIdentityResolver freezes physical compiler-root identities for
// one operation. Target identities discovered upstream seed the resolver
// before it consults the filesystem; all other paths are resolved by one
// bounded per-file pass. It is never retained by a Session, ProjectSet, or
// Program.
type projectRootIdentityResolver struct {
	fsys                  vfs.FS
	singleThreaded        bool
	canonicalBySourcePath map[string]string
}

func newProjectRootIdentityResolver(
	targets []target.File,
	fsys vfs.FS,
	singleThreaded bool,
) *projectRootIdentityResolver {
	canonicalBySourcePath := make(map[string]string, len(targets)*2)
	targetCanonicalIDs := make(map[string]struct{}, len(targets))
	seedCanonicalTargetIdentityHints(targets, canonicalBySourcePath, targetCanonicalIDs)
	return &projectRootIdentityResolver{
		fsys:                  fsys,
		singleThreaded:        singleThreaded,
		canonicalBySourcePath: canonicalBySourcePath,
	}
}

func (resolver *projectRootIdentityResolver) knownCanonicalID(sourcePathID string) (string, bool) {
	if resolver == nil {
		return "", false
	}
	canonicalID, known := resolver.canonicalBySourcePath[sourcePathID]
	return canonicalID, known
}

func runProjectRootIdentityWork(singleThreaded bool, taskCount int, task func(int)) {
	if taskCount == 0 {
		return
	}
	// Filesystem identity work frequently blocks in VFS/cache layers. A higher
	// worker-to-processor ratio preserves lookup overlap, while the absolute
	// limit prevents one goroutine per source in very large projects.
	workerCount := min(
		runtime.GOMAXPROCS(0)*projectRootIdentityWorkersPerProcessor,
		projectRootIdentityWorkerLimit,
		taskCount,
	)
	if singleThreaded || workerCount == 1 {
		for index := range taskCount {
			task(index)
		}
		return
	}

	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				task(index)
			}
		}()
	}
	for index := range taskCount {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
}

func (resolver *projectRootIdentityResolver) canonicalRootPathIDs(sourcePaths []string) []string {
	canonicalIDs := make([]string, len(sourcePaths))
	if len(sourcePaths) == 0 {
		return canonicalIDs
	}
	if resolver == nil || resolver.fsys == nil {
		for index, sourcePath := range sourcePaths {
			canonicalIDs[index] = exactPathID(sourcePath)
		}
		return canonicalIDs
	}

	// Per-file realpath is the only identity operation guaranteed by vfs.FS.
	// Do not enumerate a directory or infer a file identity by joining its
	// parent's realpath: either can be incorrect for an overlay VFS, and the
	// directory work can be unbounded relative to the requested root set.
	runProjectRootIdentityWork(resolver.singleThreaded, len(sourcePaths), func(index int) {
		canonicalIDs[index] = canonicalPathID(sourcePaths[index], resolver.fsys)
	})
	return canonicalIDs
}
