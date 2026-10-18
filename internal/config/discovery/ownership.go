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
	physicalDirectory  string
	canonicalSearchDir string
	usingCanonical     bool
	lexicalCandidate   bool
	explicitFile       bool
	ownerDir           string
	ownerPath          string
	gitCursor          gitignore.Cursor
	done               bool
}

type directorySeedResolution struct {
	seed             *discoverySeed
	candidates       []configCandidate
	next             int
	currentDirectory string
	gitRoot          string
}

// addCanonicalSeedFallback preserves lexical config ancestry as the first
// choice. Physical ancestry is searched only when no lexical candidate exists.
func (coordinator *discoveryCoordinator) addCanonicalSeedFallback(seed *discoverySeed, canonicalPath string, isDirectory bool) {
	if seed == nil || canonicalPath == "" {
		return
	}
	canonicalDirectory := tspath.NormalizePath(canonicalPath)
	if !isDirectory {
		canonicalDirectory = tspath.GetDirectoryPath(canonicalDirectory)
	}
	if canonicalDirectory == "" || canonicalDirectory == seed.searchDir {
		return
	}
	seed.canonicalSearchDir = canonicalDirectory
}

// resolveDirectorySeedOwners evaluates authored config ancestry from outer to
// inner. Git source state is target-scoped and never resets when the owner
// changes. A supplied root retains its existing config-ownership exception.
func (coordinator *discoveryCoordinator) resolveDirectorySeedOwners(seeds []*discoverySeed) error {
	resolutions := make([]directorySeedResolution, 0, len(seeds))
	for _, seed := range seeds {
		candidates := coordinator.findCandidateChain(seed.searchDir)
		if len(candidates) == 0 && seed.canonicalSearchDir != "" {
			seed.usingCanonical = true
			candidates = coordinator.findCandidateChain(seed.canonicalSearchDir)
		}
		resolutions = append(resolutions, directorySeedResolution{
			seed:       seed,
			candidates: candidates,
			gitRoot:    coordinator.git.RootDirectory(seed.physicalDirectory),
		})
	}

	for {
		if err := coordinator.ctx.Err(); err != nil {
			return err
		}
		var candidates []configCandidate
		var pending []*directorySeedResolution
		for index := range resolutions {
			resolution := &resolutions[index]
			for resolution.next < len(resolution.candidates) {
				candidateDirectory := resolution.candidates[resolution.next].directory
				rootDirectory := resolution.seed.searchDir
				if resolution.seed.usingCanonical {
					rootDirectory = resolution.seed.canonicalSearchDir
				}
				reachable, authoredBlocked := coordinator.advanceDirectorySeedGit(
					resolution,
					candidateDirectory,
					discoveryPathsEqual(candidateDirectory, rootDirectory, coordinator.fs.UseCaseSensitiveFileNames()),
				)
				if authoredBlocked {
					resolution.next = len(resolution.candidates)
					break
				}
				if !reachable {
					resolution.next++
					continue
				}
				candidate, found := coordinator.findCandidateForOwner(
					candidateDirectory,
					resolution.seed.ownerPath,
					coordinator.physicalDirectory(candidateDirectory),
				)
				if !found {
					resolution.next++
					continue
				}
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
			state := coordinator.modules.state(resolution.candidates[resolution.next].path)
			if state != nil && state.failure == nil {
				resolution.seed.ownerDir = state.candidate.directory
				resolution.seed.ownerPath = state.candidate.path
			}
			resolution.next++
		}
	}

	for index := range resolutions {
		resolution := &resolutions[index]
		rootDirectory := resolution.seed.searchDir
		if resolution.seed.usingCanonical {
			rootDirectory = resolution.seed.canonicalSearchDir
		}
		_, authoredBlocked := coordinator.advanceDirectorySeedGit(resolution, rootDirectory, true)
		if authoredBlocked {
			continue
		}
		resolution.seed.gitCursor = coordinator.git.Ancestors(resolution.seed.physicalDirectory)
	}
	return nil
}

// advanceDirectorySeedGit checks the requested ancestry before executing a
// candidate. Git gates only the physical chain below the target's Git root;
// config search may continue above that boundary or along another lexical
// ancestry. Sources in an authored-pruned subtree are never observed here.
func (coordinator *discoveryCoordinator) advanceDirectorySeedGit(
	resolution *directorySeedResolution,
	destination string,
	reopenDestination bool,
) (reachable bool, authoredBlocked bool) {
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	destination = tspath.NormalizePath(destination)
	var reverse []string
	for directory := destination; directory != ""; directory = configDiscoveryParent(directory) {
		reverse = append(reverse, directory)
		if resolution.currentDirectory == "" || discoveryPathsEqual(directory, resolution.currentDirectory, useCaseSensitive) {
			break
		}
	}
	for index := len(reverse) - 1; index >= 0; index-- {
		directory := reverse[index]
		physical := coordinator.physicalDirectory(directory)
		if coordinator.isGloballyIgnoredDirectory(resolution.seed.ownerPath, directory, physical) {
			return false, true
		}
		reachable = true
		_, belowGitRoot := rslintconfig.RelativePathWithinConfigRoot(physical, resolution.gitRoot, useCaseSensitive)
		_, targetDescends := rslintconfig.RelativePathWithinConfigRoot(resolution.seed.physicalDirectory, physical, useCaseSensitive)
		if belowGitRoot && targetDescends {
			cursor := coordinator.git.Ancestors(physical)
			if !cursor.SourceReachable() && !coordinator.reopensGitignoredDirectory(resolution.seed.ownerPath, directory, physical) {
				reachable = false
			}
		}
		resolution.currentDirectory = directory
	}
	if reopenDestination {
		reachable = true
	}
	return reachable, false
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
	if canonicalDirectory == "" {
		canonicalDirectory = coordinator.physicalDirectory(directory)
	}
	for _, name := range AutoJSConfigFileNames {
		candidatePath := tspath.CombinePaths(directory, name)
		canonicalCandidate := tspath.CombinePaths(canonicalDirectory, name)
		if isDefaultDiscoveryExcluded(candidatePath, coordinator.request.CWD, coordinator.fs.UseCaseSensitiveFileNames()) ||
			!coordinator.fs.FileExists(canonicalCandidate) {
			continue
		}
		if coordinator.isGloballyIgnoredCandidate(ownerPath, rslintconfig.PathIdentity{
			Path: candidatePath, CanonicalPath: canonicalCandidate, CanonicalParentPath: canonicalDirectory,
		}) {
			continue
		}
		return configCandidate{path: tspath.NormalizePath(candidatePath), directory: directory, physicalDirectory: canonicalDirectory}, true
	}
	return configCandidate{}, false
}
