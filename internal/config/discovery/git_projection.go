package discovery

import (
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type gitignoreObservation struct {
	ownerDirectory  string
	scopeRoot       string
	sourceDirectory string
	patterns        []gitignore.Pattern
}

type gitignoreReadResult struct {
	content string
	exists  bool
}

// gitProjection owns one transaction-wide view of Git ignore sources. Reads
// are frozen once per lexical source even when concurrent or overlapping
// discovery routes observe it.
type gitProjection struct {
	fs vfs.FS

	sources     map[string]map[tspath.Path]gitignoreObservation
	scopeRoots  map[string][]string
	readMu      sync.Mutex
	readCache   map[tspath.Path]gitignoreReadResult
	readPending map[tspath.Path]chan struct{}
}

func newGitProjection(fsys vfs.FS) gitProjection {
	return gitProjection{fs: fsys}
}

func (projection *gitProjection) readSource(
	ownerDirectory string,
	sourceDirectory string,
	matchDirectory string,
	cursor gitignore.Cursor,
) (gitignore.Cursor, *gitignoreObservation) {
	if ownerDirectory == "" || sourceDirectory == "" || matchDirectory == "" || !cursor.SourceReachable() {
		return cursor, nil
	}
	content := projection.readFile(sourceDirectory)
	if !content.exists {
		return cursor, nil
	}
	next, patterns := cursor.AppendSourcePatterns(matchDirectory, content.content)
	if len(patterns) == 0 {
		return next, nil
	}
	matchRoot := cursor.RootDirectory()
	patterns = rslintconfig.CollectedGitignorePatternsForRoot(
		patterns,
		ownerDirectory,
		matchRoot,
		projection.fs,
	)
	return next, &gitignoreObservation{
		ownerDirectory:  ownerDirectory,
		scopeRoot:       matchRoot,
		sourceDirectory: matchDirectory,
		patterns:        patterns,
	}
}

// readFile freezes each lexical source for one catalog generation. Different
// source paths may read concurrently; duplicate in-flight reads share the
// first result, including a cached miss.
func (projection *gitProjection) readFile(sourceDirectory string) gitignoreReadResult {
	path := tspath.CombinePaths(sourceDirectory, ".gitignore")
	identity := tspath.ToPath(
		tspath.NormalizePath(path),
		"",
		projection.fs.UseCaseSensitiveFileNames(),
	)
	for {
		projection.readMu.Lock()
		if projection.readCache == nil {
			projection.readCache = make(map[tspath.Path]gitignoreReadResult)
		}
		if projection.readPending == nil {
			projection.readPending = make(map[tspath.Path]chan struct{})
		}
		if result, exists := projection.readCache[identity]; exists {
			projection.readMu.Unlock()
			return result
		}
		if wait, pending := projection.readPending[identity]; pending {
			if wait == nil {
				wait = make(chan struct{})
				projection.readPending[identity] = wait
			}
			projection.readMu.Unlock()
			<-wait
			continue
		}
		projection.readPending[identity] = nil
		projection.readMu.Unlock()
		break
	}

	content, exists := projection.fs.ReadFile(path)
	result := gitignoreReadResult{content: content, exists: exists}
	projection.readMu.Lock()
	projection.readCache[identity] = result
	wait := projection.readPending[identity]
	delete(projection.readPending, identity)
	if wait != nil {
		close(wait)
	}
	projection.readMu.Unlock()
	return result
}

func (projection *gitProjection) observeSource(
	ownerDirectory string,
	sourceDirectory string,
	matchDirectory string,
	cursor gitignore.Cursor,
) gitignore.Cursor {
	next, observation := projection.readSource(
		ownerDirectory,
		sourceDirectory,
		matchDirectory,
		cursor,
	)
	if observation != nil {
		projection.recordObservation(*observation)
	}
	return next
}

func (projection *gitProjection) recordObservation(observation gitignoreObservation) {
	if observation.ownerDirectory == "" || observation.scopeRoot == "" || observation.sourceDirectory == "" || len(observation.patterns) == 0 {
		return
	}
	projection.recordScope(observation.ownerDirectory, observation.scopeRoot)
	if projection.sources == nil {
		projection.sources = make(map[string]map[tspath.Path]gitignoreObservation)
	}
	sources := projection.sources[observation.ownerDirectory]
	if sources == nil {
		sources = make(map[tspath.Path]gitignoreObservation)
		projection.sources[observation.ownerDirectory] = sources
	}
	scopeIdentity := tspath.ToPath(
		tspath.NormalizePath(observation.scopeRoot),
		"",
		projection.fs.UseCaseSensitiveFileNames(),
	)
	sourceIdentity := tspath.ToPath(
		tspath.NormalizePath(observation.sourceDirectory),
		"",
		projection.fs.UseCaseSensitiveFileNames(),
	)
	identity := tspath.Path(string(scopeIdentity) + "\x00" + string(sourceIdentity))
	if _, exists := sources[identity]; exists {
		return
	}
	observation.patterns = append([]gitignore.Pattern(nil), observation.patterns...)
	sources[identity] = observation
}

func (projection *gitProjection) recordScope(ownerDirectory string, scopeRoot string) {
	if ownerDirectory == "" || scopeRoot == "" {
		return
	}
	if projection.scopeRoots == nil {
		projection.scopeRoots = make(map[string][]string)
	}
	scopeRoot = tspath.NormalizePath(scopeRoot)
	identity := tspath.ToPath(scopeRoot, "", projection.fs.UseCaseSensitiveFileNames())
	for _, existing := range projection.scopeRoots[ownerDirectory] {
		if tspath.ToPath(existing, "", projection.fs.UseCaseSensitiveFileNames()) == identity {
			return
		}
	}
	projection.scopeRoots[ownerDirectory] = append(projection.scopeRoots[ownerDirectory], scopeRoot)
}

// collectExactTargets records source chains for targets whose owner is already
// known. Observations remain source-scoped so exact chains and directory walks
// can merge without duplicating parent patterns.
func (projection *gitProjection) collectExactTargets(seeds []*discoverySeed) {
	useCaseSensitive := projection.fs.UseCaseSensitiveFileNames()
	cursorByOwnerAndDirectory := make(map[string]map[tspath.Path]gitignore.Cursor)
	barrierByOwnerAndDirectory := make(map[string]map[tspath.Path]struct{})
	for _, seed := range seeds {
		if seed == nil || seed.ownerDir == "" {
			continue
		}
		ownerCursors := cursorByOwnerAndDirectory[seed.ownerDir]
		if ownerCursors == nil {
			ownerCursors = make(map[tspath.Path]gitignore.Cursor)
			cursorByOwnerAndDirectory[seed.ownerDir] = ownerCursors
		}
		ownerBarriers := barrierByOwnerAndDirectory[seed.ownerDir]
		if ownerBarriers == nil {
			ownerBarriers = make(map[tspath.Path]struct{})
			barrierByOwnerAndDirectory[seed.ownerDir] = ownerBarriers
		}
		projection.collectExactTargetChain(seed, ownerCursors, ownerBarriers, useCaseSensitive)
	}
}

func (projection *gitProjection) collectExactTargetChain(
	seed *discoverySeed,
	cursorByDirectory map[tspath.Path]gitignore.Cursor,
	barriers map[tspath.Path]struct{},
	useCaseSensitive bool,
) {
	gitignoreRoot := seed.gitignoreRoot
	if gitignoreRoot == "" {
		gitignoreRoot = seed.ownerDir
	}
	targetDirectory := tspath.GetDirectoryPath(seed.path)
	matchDirectory := targetDirectory
	if _, within := rslintconfig.RelativePathWithinConfigRoot(
		matchDirectory,
		gitignoreRoot,
		useCaseSensitive,
	); !within {
		matchDirectory = seed.canonicalSearchDir
		if matchDirectory == "" {
			return
		}
		if _, within = rslintconfig.RelativePathWithinConfigRoot(
			matchDirectory,
			gitignoreRoot,
			useCaseSensitive,
		); !within {
			return
		}
	}

	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		matchDirectory,
		gitignoreRoot,
		useCaseSensitive,
	)
	if !within {
		return
	}
	cursor := gitignore.NewCursor(gitignoreRoot, useCaseSensitive)
	currentMatch := gitignoreRoot
	currentSource := gitignoreRoot
	components := splitDiscoveryPath(relative)
	for index := 0; ; index++ {
		identity := tspath.ToPath(currentMatch, "", useCaseSensitive)
		if cached, exists := cursorByDirectory[identity]; exists {
			cursor = cached
		} else {
			cursor = projection.observeSource(
				seed.ownerDir,
				currentSource,
				currentMatch,
				cursor,
			)
			cursorByDirectory[identity] = cursor
		}
		if index == len(components) || !cursor.SourceReachable() {
			return
		}
		component := components[index]
		nextSource := tspath.CombinePaths(currentSource, component)
		nextMatch := tspath.CombinePaths(currentMatch, component)
		nextIdentity := tspath.ToPath(nextMatch, "", useCaseSensitive)
		if _, blocked := barriers[nextIdentity]; blocked {
			return
		}
		if cached, exists := cursorByDirectory[nextIdentity]; exists {
			currentSource = nextSource
			currentMatch = nextMatch
			cursor = cached
			continue
		}
		entries := projection.fs.GetAccessibleEntries(currentSource)
		parentRealPath := ""
		if entries.Symlinks == nil {
			parentRealPath = projection.fs.Realpath(currentSource)
		}
		if isSymlinkDirectoryChild(projection.fs, currentSource, parentRealPath, component, entries) {
			barriers[nextIdentity] = struct{}{}
			return
		}
		currentSource = nextSource
		currentMatch = nextMatch
		next, blocked := cursor.Enter(currentMatch)
		cursor = next
		if blocked {
			barriers[nextIdentity] = struct{}{}
			return
		}
	}
}

// materialize returns an effective config with the transaction-frozen Git
// projection. It never mutates or retains the catalog draft.
func (projection *gitProjection) materialize(
	ownerDirectory string,
	entries rslintconfig.RslintConfig,
) rslintconfig.RslintConfig {
	sources := projection.sources[ownerDirectory]
	if len(sources) == 0 {
		return entries
	}
	observations := make([]gitignoreObservation, 0, len(sources))
	for _, observation := range sources {
		observations = append(observations, observation)
	}
	scopeRoots := projection.scopeRoots[ownerDirectory]
	scopeOrder := make(map[tspath.Path]int, len(scopeRoots))
	for index, root := range scopeRoots {
		scopeOrder[tspath.ToPath(root, "", projection.fs.UseCaseSensitiveFileNames())] = index
	}
	caseInsensitive := !projection.fs.UseCaseSensitiveFileNames()
	sort.Slice(observations, func(i, j int) bool {
		leftScope := scopeOrder[tspath.ToPath(observations[i].scopeRoot, "", projection.fs.UseCaseSensitiveFileNames())]
		rightScope := scopeOrder[tspath.ToPath(observations[j].scopeRoot, "", projection.fs.UseCaseSensitiveFileNames())]
		if leftScope != rightScope {
			return leftScope < rightScope
		}
		leftDepth := projection.sourceDepth(observations[i].scopeRoot, observations[i].sourceDirectory)
		rightDepth := projection.sourceDepth(observations[j].scopeRoot, observations[j].sourceDirectory)
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		left := observations[i].sourceDirectory
		right := observations[j].sourceDirectory
		if caseInsensitive {
			leftFolded := strings.ToLower(left)
			rightFolded := strings.ToLower(right)
			if leftFolded != rightFolded {
				return leftFolded < rightFolded
			}
		}
		return left < right
	})
	var patterns []gitignore.Pattern
	for _, observation := range observations {
		patterns = append(patterns, observation.patterns...)
	}
	return rslintconfig.ConfigWithCollectedGitignoreScopes(
		entries,
		patterns,
		scopeRoots,
		ownerDirectory,
		projection.fs,
		caseInsensitive,
	)
}

func (projection *gitProjection) sourceDepth(scopeRoot string, sourceDirectory string) int {
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		sourceDirectory,
		scopeRoot,
		projection.fs.UseCaseSensitiveFileNames(),
	)
	if !within || relative == "" {
		return 0
	}
	return len(splitDiscoveryPath(relative))
}

func isSymlinkDirectoryChild(
	fsys vfs.FS,
	parentDirectory string,
	parentRealPath string,
	name string,
	entries vfs.Entries,
) bool {
	if entries.Symlinks != nil {
		for symlink := range entries.Symlinks {
			if symlink == name || (!fsys.UseCaseSensitiveFileNames() && strings.EqualFold(symlink, name)) {
				return true
			}
		}
		return false
	}
	childDirectory := tspath.CombinePaths(parentDirectory, name)
	childRealPath := fsys.Realpath(childDirectory)
	if parentRealPath == "" || childRealPath == "" {
		return false
	}
	expectedRealPath := tspath.CombinePaths(parentRealPath, name)
	return tspath.ComparePaths(childRealPath, expectedRealPath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: fsys.UseCaseSensitiveFileNames(),
	}) != 0
}
