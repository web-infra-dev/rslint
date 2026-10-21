package discovery

import (
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type discoveryWalkNode struct {
	directory          string
	canonicalDirectory string
	ownerDir           string
	ownerPath          string
	ignoreMatcher      rslintconfig.GlobalIgnoreMatcher
	configReachable    bool
	gitCursor          gitignore.Cursor
	targets            *discoveryTargetTrie
}
type suspendedDiscoveryNode struct {
	node      discoveryWalkNode
	candidate configCandidate
}

type discoveryWalkResult struct {
	children           []discoveryWalkNode
	files              []string
	pending            *suspendedDiscoveryNode
	candidateToAdopt   *configLoadState
	directoriesVisited int
	directoriesPruned  int
	err                error
}
type discoveryWalkStats struct {
	directoriesVisited int
	directoriesPruned  int
}

func (coordinator *discoveryCoordinator) walkDirectories(roots []discoveryWalkNode) error {
	queue := append([]discoveryWalkNode(nil), roots...)
	for len(queue) > 0 {
		if err := coordinator.ctx.Err(); err != nil {
			return err
		}
		sort.Slice(queue, func(i, j int) bool { return queue[i].directory < queue[j].directory })
		for _, node := range queue {
			if _, exists := coordinator.physicalDirectories[node.directory]; !exists {
				coordinator.physicalDirectories[node.directory] = node.canonicalDirectory
			}
		}
		var next []discoveryWalkNode
		var suspended []suspendedDiscoveryNode
		var candidates []configCandidate
		results := coordinator.processWalkFrontier(queue)
		// Results are indexed by the already-sorted frontier. Merging on this
		// goroutine keeps catalog state, stats, and loader batches deterministic
		// regardless of worker completion order.
		for _, result := range results {
			if result.err != nil {
				return result.err
			}
			coordinator.walkStats.directoriesVisited += result.directoriesVisited
			coordinator.walkStats.directoriesPruned += result.directoriesPruned
			for _, file := range result.files {
				coordinator.reachableFiles[file] = true
			}
			if result.candidateToAdopt != nil {
				if err := coordinator.adoptCandidate(result.candidateToAdopt, false); err != nil {
					return err
				}
				candidate := result.candidateToAdopt.candidate
				coordinator.draft.addDirectory(candidate.directory, rslintconfig.DirectoryIdentity{
					LexicalPath: candidate.directory, CanonicalPath: candidate.physicalDirectory,
				})
			}
			next = append(next, result.children...)
			if result.pending != nil {
				suspended = append(suspended, *result.pending)
				candidates = append(candidates, result.pending.candidate)
			}
		}
		if len(candidates) > 0 {
			if err := coordinator.modules.loadCandidates(candidates); err != nil {
				return err
			}
			for _, item := range suspended {
				state := coordinator.modules.state(item.candidate.path)
				if state != nil && state.failure == nil {
					if err := coordinator.adoptCandidate(state, false); err != nil {
						return err
					}
					coordinator.draft.addDirectory(state.candidate.directory, rslintconfig.DirectoryIdentity{
						LexicalPath: state.candidate.directory, CanonicalPath: state.candidate.physicalDirectory,
					})
					item.node.ownerDir = state.candidate.directory
					item.node.ownerPath = state.candidate.path
					item.node.ignoreMatcher = state.ignoreMatcher
				}
				next = append(next, item.node)
			}
		}
		// Overlapping requested roots are intentionally independent routes. The
		// same lexical directory can carry inherited Git state on one route and
		// explicit-root reachability on another, so directory-only deduplication
		// would be incorrect.
		queue = next
	}
	return nil
}

func (coordinator *discoveryCoordinator) processWalkFrontier(nodes []discoveryWalkNode) []discoveryWalkResult {
	results := make([]discoveryWalkResult, len(nodes))
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if coordinator.request.SingleThreaded {
		workers = 1
	}
	if workers > len(nodes) {
		workers = len(nodes)
	}
	if workers <= 1 {
		for index, node := range nodes {
			results[index] = coordinator.processWalkNode(node)
		}
		return results
	}

	jobs := make(chan int)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers)
	for range workers {
		go func() {
			defer waitGroup.Done()
			for index := range jobs {
				results[index] = coordinator.processWalkNode(nodes[index])
			}
		}()
	}
	for index := range nodes {
		jobs <- index
	}
	close(jobs)
	waitGroup.Wait()
	return results
}

func (coordinator *discoveryCoordinator) processWalkNode(node discoveryWalkNode) discoveryWalkResult {
	if err := coordinator.ctx.Err(); err != nil {
		return discoveryWalkResult{err: err}
	}
	if node.ignoreMatcher.BlocksDirectory(node.directory, node.canonicalDirectory) {
		return discoveryWalkResult{directoriesPruned: 1}
	}
	result := discoveryWalkResult{}

	if coordinator.explicit == nil && node.configReachable {
		if candidate, found := coordinator.findCandidateForOwner(
			node.directory,
			node.ownerPath,
			node.canonicalDirectory,
		); found && rslintconfig.ExactPathID(candidate.directory) != rslintconfig.ExactPathID(node.ownerDir) {
			state := coordinator.modules.state(candidate.path)
			if state == nil {
				result.pending = &suspendedDiscoveryNode{node: node, candidate: candidate}
				return result
			}
			if state.failure == nil {
				result.candidateToAdopt = state
				node.ownerDir = state.candidate.directory
				node.ownerPath = state.candidate.path
				node.ignoreMatcher = state.ignoreMatcher
			}
		}
	}
	result.directoriesVisited = 1
	node.gitCursor = coordinator.git.Observe(node.gitCursor, node.canonicalDirectory)
	// A supplied directory may select its own config while remaining a raw Git
	// traversal boundary. That ownership exception does not authorize searching
	// its hidden children for another repository.
	if !node.gitCursor.SourceReachable() && !node.ignoreMatcher.MayReopenGitDescendants(node.directory, node.canonicalDirectory) {
		result.directoriesPruned = 1
		return result
	}
	walkTargets := node.targets
	if walkTargets != nil {
		result.files = walkTargets.files
	}
	if walkTargets != nil && len(walkTargets.children) == 0 {
		return result
	}

	entries := coordinator.fs.GetAccessibleEntries(node.canonicalDirectory)
	directories := append([]string(nil), entries.Directories...)
	sort.Strings(directories)
	parentRealPath := ""
	if entries.Symlinks == nil && len(directories) > 0 {
		parentRealPath = node.canonicalDirectory
	}
	children := make([]discoveryWalkNode, 0, len(directories))
	for _, name := range directories {
		var childTargets *discoveryTargetTrie
		if walkTargets != nil {
			childTargets = walkTargets.child(name, coordinator.fs.UseCaseSensitiveFileNames())
			if childTargets == nil {
				continue
			}
		}
		if rslintconfig.IsDefaultExcludedPath(name, "", coordinator.fs.UseCaseSensitiveFileNames()) {
			continue
		}
		if isSymlinkDirectoryChild(coordinator.fs, node.canonicalDirectory, parentRealPath, name, entries) {
			continue
		}
		child := tspath.CombinePaths(node.directory, name)
		canonicalChild := ""
		if node.canonicalDirectory != "" {
			canonicalChild = tspath.CombinePaths(node.canonicalDirectory, name)
		}
		if node.ignoreMatcher.BlocksDirectory(child, canonicalChild) {
			result.directoriesPruned++
			continue
		}
		nextCursor, entryBlocked := coordinator.git.Enter(node.gitCursor, canonicalChild)
		gitBlocked := entryBlocked || !nextCursor.SourceReachable()
		childConfigReachable := !gitBlocked || node.ignoreMatcher.ReopensDirectoryNode(child, canonicalChild)
		if gitBlocked && !node.ignoreMatcher.MayReopenGitDescendants(child, canonicalChild) {
			result.directoriesPruned++
			continue
		}
		children = append(children, discoveryWalkNode{
			directory:          child,
			canonicalDirectory: canonicalChild,
			ownerDir:           node.ownerDir,
			ownerPath:          node.ownerPath,
			ignoreMatcher:      node.ignoreMatcher,
			configReachable:    childConfigReachable,
			gitCursor:          nextCursor,
			targets:            childTargets,
		})
	}
	result.children = children
	return result
}

func (coordinator *discoveryCoordinator) isGloballyIgnoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	return coordinator.globalIgnoreMatcher(ownerPath).BlocksDirectory(directory, canonicalDirectory)
}

func (coordinator *discoveryCoordinator) reopensGitignoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	return coordinator.globalIgnoreMatcher(ownerPath).ReopensDirectoryNode(directory, canonicalDirectory)
}

func (coordinator *discoveryCoordinator) isGloballyIgnoredCandidate(ownerPath string, candidate rslintconfig.PathIdentity) bool {
	// Only an actual ancestor owner can hide a module filename; fallback
	// directory policy must not hide the module that establishes ownership.
	return ownerPath != "" && coordinator.globalIgnoreMatcher(ownerPath).IgnoresFile(candidate)
}
