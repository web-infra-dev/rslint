package discovery

import (
	"context"
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type discoveryCoordinator struct {
	ctx                context.Context
	fs                 vfs.FS
	request            ConfigDiscoveryRequest
	explicitConfigPath string
	projectOnly        bool
	transactionID      string

	modules             moduleLoadCoordinator
	git                 *gitignore.Reader
	draft               catalogDraft
	walkStats           discoveryWalkStats
	physicalDirectories map[string]string
	targetFiles         []rslintconfig.PathIdentity
	targetDirectories   []rslintconfig.DirectoryIdentity
}

func newDiscoveryCoordinator(
	ctx context.Context,
	fsys vfs.FS,
	loader ConfigModuleLoader,
	request ConfigDiscoveryRequest,
	explicitConfigPath string,
	transactionID string,
) discoveryCoordinator {
	return discoveryCoordinator{
		ctx:                ctx,
		fs:                 fsys,
		request:            request,
		explicitConfigPath: explicitConfigPath,
		transactionID:      transactionID,
		modules: newModuleLoadCoordinator(
			ctx,
			fsys,
			loader,
			transactionID,
			request.Fresh,
			request.SingleThreaded,
		),
		git:                 gitignore.NewReader(fsys),
		draft:               newCatalogDraft(),
		physicalDirectories: make(map[string]string),
	}
}

func (coordinator *discoveryCoordinator) build() (*ConfigCatalog, error) {
	if err := coordinator.ctx.Err(); err != nil {
		return nil, err
	}
	cwd := coordinator.request.CWD
	if cwd == "" {
		return nil, errors.New("config discovery requires a working directory")
	}
	cwd = tspath.NormalizePath(cwd)
	coordinator.request.CWD = cwd
	files := coordinator.normalizedFiles()
	directoryRoots := coordinator.normalizedDirectoryRoots()
	coordinator.request.Files = files
	coordinator.request.Directories = directoryRoots
	coordinator.targetDirectories = directoryRoots
	if files != nil {
		coordinator.targetFiles = make([]rslintconfig.PathIdentity, 0, len(files))
	}
	for _, file := range files {
		coordinator.targetFiles = append(coordinator.targetFiles, file.PathIdentity)
		parent := tspath.GetDirectoryPath(file.Path)
		physicalParent := file.CanonicalParentPath
		if physicalParent == "" {
			physicalParent = parent
		}
		coordinator.physicalDirectories[parent] = physicalParent
	}
	for _, directory := range directoryRoots {
		coordinator.physicalDirectories[directory.LexicalPath] = directory.CanonicalPath
	}
	anchors := make([]rslintconfig.DirectoryIdentity, 0, len(coordinator.physicalDirectories))
	for lexical, physical := range coordinator.physicalDirectories {
		anchors = append(anchors, rslintconfig.DirectoryIdentity{LexicalPath: lexical, CanonicalPath: physical})
	}
	coordinator.modules.pathSpaces = coordinator.modules.pathSpaces.WithConfigBases(nil, coordinator.fs, anchors...)

	if coordinator.explicitConfigPath != "" {
		configPath := normalizeDiscoveryPath(coordinator.explicitConfigPath, cwd)
		candidate := configCandidate{
			path:              configPath,
			directory:         tspath.GetDirectoryPath(configPath),
			physicalDirectory: coordinator.physicalDirectory(tspath.GetDirectoryPath(configPath)),
			configArrayBase:   cwd,
		}
		if err := coordinator.modules.loadCandidates([]configCandidate{candidate}); err != nil {
			return nil, err
		}
		state := coordinator.modules.state(candidate.path)
		if err := coordinator.adoptCandidate(state, false); err != nil {
			return nil, err
		}
		if !coordinator.projectOnly || rslintconfig.HasProjectOptions(state.entries) {
			if err := coordinator.collectExplicitGitSources(state); err != nil {
				return nil, err
			}
		}
		return coordinator.catalog()
	}

	var targetsByRoot map[string]*discoveryTargetTrie
	boundedDirectoryWalk := coordinator.request.LimitDirectoryWalkToFiles && len(directoryRoots) > 0 && len(files) > 0
	if boundedDirectoryWalk {
		targetsByRoot = coordinator.targetAncestorTries(directoryRoots, files)
	}
	seeds := make([]*discoverySeed, 0, len(directoryRoots)+len(coordinator.request.Files))
	directorySeeds := make([]*discoverySeed, 0, len(directoryRoots))
	directorySeedByPath := make(map[string]*discoverySeed, len(directoryRoots))
	for _, identity := range directoryRoots {
		directory := identity.LexicalPath
		if boundedDirectoryWalk && targetsByRoot[directory] == nil {
			continue
		}
		useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
		defaultExcluded := isDefaultDiscoveryExcluded(directory, cwd, useCaseSensitive)
		ancestryOnly := defaultExcluded
		searchDirectory := directory
		if ancestryOnly {
			coordinator.walkStats.directoriesPruned++
			// A default-excluded directory is a downward traversal boundary, not a
			// reason to discard still-reachable configuration outside it. Skip the
			// root and every default-excluded ancestor, then resolve normally.
			searchDirectory = configDiscoveryParent(directory)
			for searchDirectory != "" && isDefaultDiscoveryExcluded(searchDirectory, cwd, useCaseSensitive) {
				searchDirectory = configDiscoveryParent(searchDirectory)
			}
			if searchDirectory == "" {
				continue
			}
		}
		seed := &discoverySeed{
			path:              directory,
			searchDir:         searchDirectory,
			physicalDirectory: identity.CanonicalPath,
		}
		if !ancestryOnly {
			coordinator.addCanonicalSeedFallback(seed, identity.CanonicalPath, true)
			directorySeedByPath[directory] = seed
		}
		seeds = append(seeds, seed)
		directorySeeds = append(directorySeeds, seed)
	}

	explicitSeeds := make([]*discoverySeed, 0, len(coordinator.request.Files))
	fileSeeds := make([]*discoverySeed, 0, len(coordinator.request.Files))
	for _, file := range files {
		if !file.Explicit {
			// Files produced by a glob/directory walk inherit the staged target
			// trie. Resolving each file independently would reopen parent-global-
			// ignored subtrees and turn a bounded walk back into O(files * ancestry).
			continue
		}
		fileDirectory := tspath.GetDirectoryPath(file.Path)
		defaultExcluded := isDefaultDiscoveryExcluded(file.Path, cwd, coordinator.fs.UseCaseSensitiveFileNames())
		seed := &discoverySeed{
			path:         file.Path,
			searchDir:    fileDirectory,
			explicitFile: true,
		}
		if !defaultExcluded {
			// A default-excluded literal may search only its reachable lexical
			// ancestry to explain the ignored result. A realpath fallback could
			// escape the ignored subtree and execute an unrelated physical config
			// even though the target can never enter the lint scope.
			coordinator.addCanonicalSeedFallback(seed, file.CanonicalPath, false)
		}
		seeds = append(seeds, seed)
		fileSeeds = append(fileSeeds, seed)
		if !defaultExcluded {
			explicitSeeds = append(explicitSeeds, seed)
		}
	}

	if err := coordinator.resolveDirectorySeedOwners(directorySeeds); err != nil {
		return nil, err
	}
	if err := coordinator.resolveSeedOwners(fileSeeds); err != nil {
		return nil, err
	}
	for _, seed := range seeds {
		if seed.ownerDir == "" {
			continue
		}
		state := coordinator.modules.state(seed.ownerPath)
		if state == nil {
			continue
		}
		if err := coordinator.adoptCandidate(state, seed.explicitFile); err != nil {
			return nil, err
		}
	}
	for _, seed := range explicitSeeds {
		if seed.ownerDir == "" {
			continue
		}
		coordinator.draft.addExplicitFile(seed.ownerDir, seed.path)
	}
	coordinator.collectExactGitFiles()

	walkRoots := make([]discoveryWalkNode, 0, len(directorySeedByPath))
	for _, identity := range directoryRoots {
		directory := identity.LexicalPath
		seed := directorySeedByPath[directory]
		if seed == nil {
			continue
		}
		walkRoots = append(walkRoots, discoveryWalkNode{
			directory:          directory,
			canonicalDirectory: seed.physicalDirectory,
			ownerDir:           seed.ownerDir,
			ownerPath:          seed.ownerPath,
			configReachable:    true,
			gitCursor:          seed.gitCursor,
			targets:            targetsByRoot[directory],
		})
	}
	if err := coordinator.walkDirectories(walkRoots); err != nil {
		return nil, err
	}

	return coordinator.catalog()
}

// adoptCandidate changes authored ownership without changing Git state.
func (coordinator *discoveryCoordinator) adoptCandidate(state *configLoadState, explicitOnly bool) error {
	if err := coordinator.draft.adoptCandidate(state, explicitOnly); err != nil {
		return err
	}
	return nil
}

func (coordinator *discoveryCoordinator) catalog() (*ConfigCatalog, error) {
	configs, scopes, effectiveIDs, err := coordinator.draft.finalize(coordinator.physicalDirectory, coordinator.fs.UseCaseSensitiveFileNames())
	if err != nil {
		return nil, err
	}
	eslintPlugins, err := coordinator.modules.activateEffectiveConfigs(effectiveIDs)
	if err != nil {
		return nil, err
	}
	return &ConfigCatalog{
		TransactionID:      coordinator.transactionID,
		GitPolicy:          rslintconfig.NewGitPolicy(coordinator.git.Snapshot()),
		PathSpaces:         coordinator.modules.pathSpaces,
		TargetFiles:        coordinator.targetFiles,
		TargetDirectories:  coordinator.targetDirectories,
		Configs:            configs,
		EffectiveConfigIDs: effectiveIDs,
		EslintPlugins:      eslintPlugins,
		Scopes:             scopes,
		Failures:           coordinator.modules.failures(),
		Stats: ConfigDiscoveryStats{
			DirectoriesVisited: coordinator.walkStats.directoriesVisited,
			ConfigsRequested:   coordinator.modules.configsRequested,
			ConfigsLoaded:      coordinator.modules.configsLoaded,
			DirectoriesPruned:  coordinator.walkStats.directoriesPruned,
		},
		Explicit: coordinator.explicitConfigPath != "",
	}, nil
}
