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
	gitDirectory       string
	gitCursor          gitignore.Cursor
	gitActive          bool
	targets            *discoveryTargetTrie
}
type suspendedDiscoveryNode struct {
	node      discoveryWalkNode
	candidate configCandidate
}

type discoveryWalkResult struct {
	children             []discoveryWalkNode
	pending              *suspendedDiscoveryNode
	candidateToAdopt     *configLoadState
	gitignoreObservation *gitignoreObservation
	directoriesVisited   int
	directoriesPruned    int
	err                  error
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
			if result.candidateToAdopt != nil {
				if err := coordinator.adoptCandidate(result.candidateToAdopt, false); err != nil {
					return err
				}
			}
			if result.gitignoreObservation != nil {
				coordinator.git.recordObservation(*result.gitignoreObservation)
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
					item.node.ownerDir = state.candidate.directory
					item.node.ownerPath = state.candidate.path
					item.node.gitCursor = gitignore.NewCursor(
						state.candidate.directory,
						coordinator.fs.UseCaseSensitiveFileNames(),
					)
					item.node.gitDirectory = state.candidate.directory
					item.node.gitActive = true
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
	if coordinator.isGloballyIgnoredDirectory(node.ownerPath, node.directory, node.canonicalDirectory) {
		return discoveryWalkResult{directoriesPruned: 1}
	}
	result := discoveryWalkResult{}

	if coordinator.explicitConfigPath == "" {
		if candidate, found := coordinator.findCandidateForOwner(
			node.directory,
			node.ownerPath,
			node.canonicalDirectory,
		); found && candidate.directory != node.ownerDir {
			state := coordinator.modules.state(candidate.path)
			if state == nil {
				result.pending = &suspendedDiscoveryNode{node: node, candidate: candidate}
				return result
			}
			if state.failure == nil {
				result.candidateToAdopt = state
				node.ownerDir = state.candidate.directory
				node.ownerPath = state.candidate.path
				node.gitCursor = gitignore.NewCursor(
					state.candidate.directory,
					coordinator.fs.UseCaseSensitiveFileNames(),
				)
				node.gitDirectory = state.candidate.directory
				node.gitActive = true
			}
		}
	}
	result.directoriesVisited = 1
	if node.gitActive {
		nextCursor, observation := coordinator.git.readSource(
			node.ownerDir,
			node.directory,
			node.gitDirectory,
			node.gitCursor,
		)
		node.gitCursor = nextCursor
		result.gitignoreObservation = observation
	}
	walkTargets := node.targets
	if walkTargets != nil && walkTargets.recursive {
		walkTargets = nil
	}
	if walkTargets != nil && len(walkTargets.children) == 0 {
		return result
	}

	entries := coordinator.fs.GetAccessibleEntries(node.directory)
	directories := append([]string(nil), entries.Directories...)
	sort.Strings(directories)
	parentRealPath := ""
	if entries.Symlinks == nil && len(directories) > 0 {
		parentRealPath = coordinator.fs.Realpath(node.directory)
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
		if isSymlinkDirectoryChild(coordinator.fs, node.directory, parentRealPath, name, entries) {
			continue
		}
		child := tspath.CombinePaths(node.directory, name)
		canonicalChild := ""
		if node.canonicalDirectory != "" {
			canonicalChild = tspath.CombinePaths(node.canonicalDirectory, name)
		}
		if coordinator.isGloballyIgnoredDirectory(node.ownerPath, child, canonicalChild) {
			result.directoriesPruned++
			continue
		}
		childGitDirectory := ""
		childGitCursor := gitignore.Cursor{}
		childGitActive := false
		if node.gitActive {
			childGitDirectory = tspath.CombinePaths(node.gitDirectory, name)
			nextCursor, gitBlocked := node.gitCursor.Enter(childGitDirectory)
			if gitBlocked && !coordinator.reopensGitignoredDirectory(
				node.ownerPath,
				child,
				canonicalChild,
			) {
				result.directoriesPruned++
				continue
			}
			childGitCursor = nextCursor
			childGitActive = true
		}
		children = append(children, discoveryWalkNode{
			directory:          child,
			canonicalDirectory: canonicalChild,
			ownerDir:           node.ownerDir,
			ownerPath:          node.ownerPath,
			gitDirectory:       childGitDirectory,
			gitCursor:          childGitCursor,
			gitActive:          childGitActive,
			targets:            childTargets,
		})
	}
	result.children = children
	return result
}

func (coordinator *discoveryCoordinator) isGloballyIgnoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	matcher, ok := coordinator.modules.globalIgnoreMatcher(ownerPath)
	return ok && matcher.BlocksDirectory(directory, canonicalDirectory)
}

func (coordinator *discoveryCoordinator) reopensGitignoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	matcher, ok := coordinator.modules.globalIgnoreMatcher(ownerPath)
	return ok && matcher.ReopensDirectoryNode(directory, canonicalDirectory)
}

func (coordinator *discoveryCoordinator) isGloballyIgnoredCandidate(ownerPath string, candidatePath string, canonicalPath string) bool {
	matcher, ok := coordinator.modules.globalIgnoreMatcher(ownerPath)
	return ok && matcher.IgnoresPath(candidatePath, canonicalPath)
}
