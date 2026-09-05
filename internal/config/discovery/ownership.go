package discovery

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type discoverySeed struct {
	path               string
	searchDir          string
	canonicalSearchDir string
	canonicalWalkDir   string
	usingCanonical     bool
	lexicalCandidate   bool
	explicitFile       bool
	ownerDir           string
	ownerPath          string
	gitignoreRoot      string
	gitDirectory       string
	gitCursor          gitignore.Cursor
	gitActive          bool
	done               bool
}
type directorySeedResolution struct {
	seed            *discoverySeed
	candidates      []configCandidate
	next            int
	gitDirectory    string
	gitCursor       gitignore.Cursor
	gitActive       bool
	configReachable bool
}

// addCanonicalSeedFallback records one physical ancestry without replacing the
// authored path. The fallback is intentionally dormant until the complete
// lexical ancestry has produced no candidate at all.
func (coordinator *discoveryCoordinator) addCanonicalSeedFallback(seed *discoverySeed, canonicalPath string, isDirectory bool) {
	if seed == nil || canonicalPath == "" {
		return
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	canonicalDirectory := canonicalPath
	if !isDirectory {
		canonicalDirectory = tspath.GetDirectoryPath(canonicalPath)
	}
	if canonicalDirectory == "" || canonicalDirectory == seed.searchDir {
		return
	}
	seed.canonicalWalkDir = canonicalDirectory
	seed.canonicalSearchDir = canonicalDirectory
}

// resolveDirectorySeedOwners evaluates each directory's complete config
// ancestry from outermost to innermost. This ordering is what makes an
// ancestor config's global ignore a discovery boundary for a nested config.
// Only the final reachable successful owner is activated later; ancestors
// loaded solely to decide reachability never leak into the effective catalog.
func (coordinator *discoveryCoordinator) resolveDirectorySeedOwners(seeds []*discoverySeed) error {
	resolutions := make([]directorySeedResolution, 0, len(seeds))
	for _, seed := range seeds {
		candidates := coordinator.findCandidateChain(seed.searchDir)
		if len(candidates) == 0 && seed.canonicalSearchDir != "" {
			seed.usingCanonical = true
			candidates = coordinator.findCandidateChain(seed.canonicalSearchDir)
		}
		resolutions = append(resolutions, directorySeedResolution{
			seed:            seed,
			candidates:      candidates,
			configReachable: true,
		})
	}

	for {
		if err := coordinator.ctx.Err(); err != nil {
			return err
		}
		var candidates []configCandidate
		pending := make([]*directorySeedResolution, 0, len(resolutions))
		for index := range resolutions {
			resolution := &resolutions[index]
			for resolution.next < len(resolution.candidates) {
				candidateDirectory := resolution.candidates[resolution.next].directory
				if resolution.gitActive {
					rootDirectory := resolution.seed.searchDir
					if resolution.seed.usingCanonical {
						rootDirectory = resolution.seed.canonicalSearchDir
					}
					reachable, authoredBlocked := coordinator.advanceDirectorySeedGit(
						resolution,
						candidateDirectory,
						discoveryPathsEqual(
							candidateDirectory,
							rootDirectory,
							coordinator.fs.UseCaseSensitiveFileNames(),
						),
					)
					if authoredBlocked {
						resolution.gitActive = false
						resolution.next = len(resolution.candidates)
						break
					}
					if !reachable {
						// Only the supplied root itself may reopen an inherited
						// Git-inaccessible ancestry. Hidden intermediate configs are
						// never evaluated.
						resolution.next++
						continue
					}
				}
				candidate, found := coordinator.findCandidateForOwner(
					candidateDirectory,
					resolution.seed.ownerPath,
					candidateDirectory,
				)
				if !found {
					// File-cover ignores keep the directory traversable for later
					// negations, but a config candidate that remains ignored is not
					// evaluated. Continue at the next candidate-bearing directory.
					resolution.next++
					continue
				}
				// The highest-priority on-disk candidate may be hidden by the
				// current owner's authored ignore while a lower-priority filename
				// remains reachable. Persist the actual request candidate so the
				// post-batch ownership update reads the matching load state.
				resolution.candidates[resolution.next] = candidate
				candidates = append(candidates, candidate)
				pending = append(pending, resolution)
				break
			}
		}
		if len(candidates) == 0 {
			break
		}
		if err := coordinator.modules.loadCandidates(candidates); err != nil {
			return err
		}
		for _, resolution := range pending {
			candidate := resolution.candidates[resolution.next]
			state := coordinator.modules.state(candidate.path)
			if state != nil && state.failure == nil {
				resolution.seed.ownerDir = state.candidate.directory
				resolution.seed.ownerPath = state.candidate.path
				resolution.gitCursor = gitignore.NewCursor(
					state.candidate.directory,
					coordinator.fs.UseCaseSensitiveFileNames(),
				)
				resolution.gitDirectory = state.candidate.directory
				resolution.gitActive = true
				resolution.configReachable = true
			}
			resolution.next++
		}
	}

	for index := range resolutions {
		resolution := &resolutions[index]
		if !resolution.gitActive {
			continue
		}
		rootDirectory := resolution.seed.searchDir
		if resolution.seed.usingCanonical {
			rootDirectory = resolution.seed.canonicalSearchDir
		}
		_, authoredBlocked := coordinator.advanceDirectorySeedGit(resolution, rootDirectory, true)
		if authoredBlocked {
			continue
		}
		resolution.seed.gitDirectory = resolution.gitDirectory
		resolution.seed.gitCursor = resolution.gitCursor
		resolution.seed.gitActive = true
	}
	return nil
}

// advanceDirectorySeedGit advances one requested directory root through the
// current owner's Git path space without reading the destination's local
// .gitignore. A candidate in the destination is therefore evaluated first; a
// successful candidate can reset ownership before that source is observed.
func (coordinator *discoveryCoordinator) advanceDirectorySeedGit(
	resolution *directorySeedResolution,
	destination string,
	reopenDestination bool,
) (reachable bool, authoredBlocked bool) {
	if resolution == nil || !resolution.gitActive {
		return true, false
	}
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	destination = tspath.NormalizePath(destination)
	if discoveryPathsEqual(resolution.gitDirectory, destination, useCaseSensitive) {
		if reopenDestination {
			resolution.configReachable = true
		}
		return resolution.configReachable, false
	}
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		destination,
		resolution.gitDirectory,
		useCaseSensitive,
	)
	if !within {
		resolution.gitCursor, _ = resolution.gitCursor.Enter(destination)
		resolution.gitDirectory = destination
		resolution.configReachable = false
		return false, false
	}

	current := resolution.gitDirectory
	for _, component := range splitDiscoveryPath(relative) {
		resolution.gitCursor = coordinator.git.observeSource(
			resolution.seed.ownerDir,
			current,
			current,
			resolution.gitCursor,
		)

		nextDirectory := tspath.CombinePaths(current, component)
		nextCursor, gitBlocked := resolution.gitCursor.Enter(nextDirectory)
		if nextCursor.SourceReachable() {
			entries := coordinator.fs.GetAccessibleEntries(current)
			parentRealPath := ""
			if entries.Symlinks == nil {
				parentRealPath = coordinator.fs.Realpath(current)
			}
			if isSymlinkDirectoryChild(
				coordinator.fs,
				current,
				parentRealPath,
				component,
				entries,
			) {
				nextCursor = nextCursor.BlockSourceTraversal()
			}
		}
		resolution.gitCursor = nextCursor
		resolution.gitDirectory = nextDirectory

		if coordinator.isGloballyIgnoredDirectory(
			resolution.seed.ownerPath,
			nextDirectory,
			nextDirectory,
		) {
			resolution.configReachable = false
			return false, true
		}
		if gitBlocked && !coordinator.reopensGitignoredDirectory(
			resolution.seed.ownerPath,
			nextDirectory,
			nextDirectory,
		) {
			resolution.configReachable = false
		}
		if reopenDestination &&
			discoveryPathsEqual(nextDirectory, destination, useCaseSensitive) {
			resolution.configReachable = true
		}
		current = nextDirectory
	}
	return resolution.configReachable, false
}

// configDiscoveryParent returns the lexical filesystem parent without walking
// above a UNC share. tspath's generic root parser treats only the server as the
// root, which is appropriate for URLs but not for filesystem config discovery.
func configDiscoveryParent(directory string) string {
	directory = tspath.NormalizePath(directory)
	if strings.HasPrefix(directory, "//") {
		serverAndRest := strings.Trim(directory[2:], "/")
		serverEnd := strings.IndexByte(serverAndRest, '/')
		if serverEnd < 0 || !strings.Contains(serverAndRest[serverEnd+1:], "/") {
			return ""
		}
	}
	parent := tspath.GetDirectoryPath(directory)
	if parent == directory {
		return ""
	}
	return parent
}

func (coordinator *discoveryCoordinator) findCandidateChain(startDirectory string) []configCandidate {
	var reverse []configCandidate
	for directory := tspath.NormalizePath(startDirectory); directory != ""; {
		if candidate, ok := coordinator.findCandidate(directory); ok {
			reverse = append(reverse, candidate)
		}
		directory = configDiscoveryParent(directory)
	}
	candidates := make([]configCandidate, len(reverse))
	for index := range reverse {
		candidates[len(reverse)-1-index] = reverse[index]
	}
	return candidates
}

// resolveSeedOwners deliberately remains nearest-first for literal files.
// That is the only config-global-ignore ownership exception. Default
// exclusions are still enforced by findCandidate.
func (coordinator *discoveryCoordinator) resolveSeedOwners(seeds []*discoverySeed) error {
	for {
		if err := coordinator.ctx.Err(); err != nil {
			return err
		}
		var candidates []configCandidate
		candidateBySeed := make(map[*discoverySeed]configCandidate)
		for _, seed := range seeds {
			if seed.done {
				continue
			}
			candidate, found := coordinator.findCandidateUp(seed.searchDir)
			if !found {
				if !seed.usingCanonical && !seed.lexicalCandidate && seed.canonicalSearchDir != "" {
					seed.usingCanonical = true
					seed.searchDir = seed.canonicalSearchDir
					candidate, found = coordinator.findCandidateUp(seed.searchDir)
				}
			}
			if !found {
				seed.done = true
				continue
			}
			if !seed.usingCanonical {
				seed.lexicalCandidate = true
			}
			candidateBySeed[seed] = candidate
			candidates = append(candidates, candidate)
		}
		if len(candidates) == 0 {
			return nil
		}
		if err := coordinator.modules.loadCandidates(candidates); err != nil {
			return err
		}
		for _, seed := range seeds {
			candidate, ok := candidateBySeed[seed]
			if !ok {
				continue
			}
			state := coordinator.modules.state(candidate.path)
			if state != nil && state.failure == nil {
				seed.ownerDir = state.candidate.directory
				seed.ownerPath = state.candidate.path
				seed.done = true
				continue
			}
			seed.searchDir = configDiscoveryParent(candidate.directory)
			if seed.searchDir == "" {
				seed.done = true
			}
		}
	}
}

func (coordinator *discoveryCoordinator) findCandidateUp(startDirectory string) (configCandidate, bool) {
	for directory := tspath.NormalizePath(startDirectory); directory != ""; {
		if candidate, ok := coordinator.findCandidate(directory); ok {
			return candidate, true
		}
		directory = configDiscoveryParent(directory)
	}
	return configCandidate{}, false
}

func (coordinator *discoveryCoordinator) findCandidate(directory string) (configCandidate, bool) {
	return coordinator.findCandidateForOwner(directory, "", "")
}

func (coordinator *discoveryCoordinator) findCandidateForOwner(
	directory string,
	ownerPath string,
	canonicalDirectory string,
) (configCandidate, bool) {
	for _, name := range AutoJSConfigFileNames {
		candidatePath := tspath.CombinePaths(directory, name)
		if isDefaultDiscoveryExcluded(candidatePath, coordinator.request.CWD, coordinator.fs.UseCaseSensitiveFileNames()) ||
			!coordinator.fs.FileExists(candidatePath) {
			continue
		}
		canonicalCandidate := ""
		if canonicalDirectory != "" {
			canonicalCandidate = tspath.CombinePaths(canonicalDirectory, name)
		}
		if coordinator.isGloballyIgnoredCandidate(ownerPath, candidatePath, canonicalCandidate) {
			continue
		}
		return configCandidate{path: tspath.NormalizePath(candidatePath), directory: directory}, true
	}
	return configCandidate{}, false
}
