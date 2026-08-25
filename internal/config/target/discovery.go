package target

import (
	"path"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func discoverLintTargetsWithPreparedFiles(
	config rslintconfig.RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) []File {
	configDir = tspath.NormalizePath(configDir)
	scanRoot = tspath.NormalizePath(scanRoot)
	if scanRoot == "" {
		scanRoot = configDir
	}
	if allowDirs == nil {
		return discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, pathSpaces,
		)
	}

	targets := []File{}
	if explicitFiles != nil {
		targets = discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, pathSpaces,
		)
	}
	if len(allowDirs) == 0 {
		return targets
	}

	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	targetMatcher := mustTargetMatcher(config, configDir, fsys, pathSpaces)
	scanRootCanonical := scanRoot
	if fsys != nil {
		if realPath := fsys.Realpath(scanRoot); realPath != "" {
			scanRootCanonical = tspath.NormalizePath(realPath)
		}
	}
	for _, root := range rslintconfig.CoalesceDirectoryIdentities(allowDirs, fsys) {
		isDefaultRoot := pathsEqual(root.LexicalPath, scanRoot, true) ||
			pathsEqual(root.CanonicalPath, scanRootCanonical, true)
		if !isDefaultRoot && (rslintconfig.IsDefaultExcludedPath(root.LexicalPath, scanRoot, useCaseSensitive) ||
			targetMatcher.CanPruneDirectory(root)) {
			continue
		}
		targets = append(targets, discoverLintTargetsWithinRoot(
			config,
			configDir,
			root.LexicalPath,
			fsys,
			nil,
			[]string{root.LexicalPath},
			stopDirs,
			singleThreaded,
			pathSpaces,
		)...)
	}

	seen := make(map[string]struct{}, len(targets))
	deduplicated := targets[:0]
	for _, target := range targets {
		key := rslintconfig.ExactPathID(target.Path)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduplicated = append(deduplicated, target)
	}
	sort.Slice(deduplicated, func(i, j int) bool {
		return deduplicated[i].Path < deduplicated[j].Path
	})
	return deduplicated
}

func discoverLintTargetsWithinRoot(
	config rslintconfig.RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) []File {
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	comparisonKey := func(filePath string) string {
		return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
	}
	configDir = tspath.NormalizePath(configDir)
	scanRoot = tspath.NormalizePath(scanRoot)
	if scanRoot == "" {
		scanRoot = configDir
	}
	configPathForMatching := func(filePath string) string {
		matchPath, _, err := pathSpaces.ResolvePath(
			tspath.NormalizePath(filePath),
			configDir,
			fsys,
		)
		if err != nil {
			panic(err)
		}
		return matchPath
	}
	resolvedAllowDirs := resolveAllowedDirectories(allowDirs, fsys)
	directWalkRoot := resolvedAllowedDirectory{
		LexicalPath:   scanRoot,
		CanonicalPath: scanRoot,
	}
	directWalkProjection := allowDirs == nil
	if len(resolvedAllowDirs) == 1 &&
		rslintconfig.ExactPathID(resolvedAllowDirs[0].LexicalPath) == rslintconfig.ExactPathID(scanRoot) {
		directWalkRoot = resolvedAllowDirs[0]
		directWalkProjection = true
	}
	directProjectionForPath := func(fullPath string) (requestedPathProjection, bool) {
		projection := requestedPathProjection{Path: fullPath}
		if directWalkRoot.LexicalPath == directWalkRoot.CanonicalPath {
			return projection, true
		}
		relative, within := rslintconfig.RelativePathWithinConfigRoot(
			fullPath,
			directWalkRoot.LexicalPath,
			true,
		)
		if !within {
			return requestedPathProjection{}, false
		}
		projection.CanonicalPath = tspath.ResolvePath(
			directWalkRoot.CanonicalPath,
			relative,
		)
		return projection, true
	}

	targetMatcher := mustTargetMatcher(config, configDir, fsys, pathSpaces)

	var allowFileSet map[string]*explicitLintTarget
	if explicitFiles != nil {
		allowFileSet = make(map[string]*explicitLintTarget, len(explicitFiles))
		for _, explicitFile := range explicitFiles {
			if explicitFile == nil {
				continue
			}
			key := comparisonKey(explicitFile.target.Path)
			if _, exists := allowFileSet[key]; !exists {
				allowFileSet[key] = explicitFile
			}
		}
	}

	targetFiles := []File{}
	seenTargets := make(map[string]struct{})
	addTarget := func(
		filePath string,
		canonicalPath string,
		canonicalParentPath string,
	) {
		filePath = tspath.NormalizePath(filePath)
		if canonicalPath == "" {
			canonicalPath = filePath
		} else {
			canonicalPath = tspath.NormalizePath(canonicalPath)
		}
		if canonicalParentPath == "" {
			canonicalParentPath = tspath.GetDirectoryPath(canonicalPath)
		} else {
			canonicalParentPath = tspath.NormalizePath(canonicalParentPath)
		}
		key := comparisonKey(filePath)
		if _, seen := seenTargets[key]; seen {
			return
		}
		seenTargets[key] = struct{}{}
		targetFiles = append(targetFiles, File{
			PathIdentity: rslintconfig.PathIdentity{
				Path:                filePath,
				CanonicalPath:       canonicalPath,
				CanonicalParentPath: canonicalParentPath,
			},
			ConfigDirectory: configDir,
		})
	}
	includeDiscoveredFile := func(target File) bool {
		if !rslintconfig.IsSupportedLintFile(target.Path) {
			return false
		}
		decision := targetMatcher.MatchFile(target.Identity())
		return decision.Selected && !decision.GloballyIgnored
	}

	addExplicitTargets := func() {
		for _, explicitFile := range allowFileSet {
			if explicitFile.selectedBy(
				configDir,
				scanRoot,
				useCaseSensitive,
				&targetMatcher,
			) {
				addTarget(
					explicitFile.target.Path,
					explicitFile.target.CanonicalPath,
					explicitFile.target.CanonicalParentPath,
				)
			}
		}
	}

	// Literal targets establish their frozen identity before any overlapping
	// directory walk can discover the same lexical path with a later filesystem
	// observation. The file-only case then returns immediately (lint-staged).
	if allowFileSet != nil {
		addExplicitTargets()
		if allowDirs == nil {
			sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Path < targetFiles[j].Path })
			return targetFiles
		}
	}

	normalizedScanRoot := normalizeWalkPath(scanRoot)

	var (
		targetMu  sync.Mutex
		dirIgnore sync.Map // map[string]bool — pattern check cache
	)

	stopWalkDirs := normalizeStopWalkDirs(normalizedScanRoot, stopDirs, useCaseSensitive)

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	pathProjectionsForWalk := func(fullPath string) []requestedPathProjection {
		if directWalkProjection {
			projection, ok := directProjectionForPath(fullPath)
			if !ok {
				return nil
			}
			return []requestedPathProjection{projection}
		}
		return projectPathThroughRequestedDirectories(
			fullPath,
			configPathForMatching(fullPath),
			resolvedAllowDirs,
			useCaseSensitive,
		)
	}
	allProjectedDirectoriesMatch := func(
		fullPath string,
		predicate func(requestedPathProjection) bool,
	) bool {
		if directWalkProjection {
			projection, ok := directProjectionForPath(fullPath)
			if !ok {
				return false
			}
			return predicate(projection)
		}
		projections := pathProjectionsForWalk(fullPath)
		if len(projections) == 0 {
			// An unresolved projection is not proof that a physical subtree is
			// irrelevant. Keep walking conservatively.
			return false
		}
		for _, projection := range projections {
			if !predicate(projection) {
				return false
			}
		}
		return true
	}

	processFile := func(walkPath string, needsRealpath bool) {
		fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
		projections := pathProjectionsForWalk(fullPath)
		if len(projections) == 0 {
			if allowFileSet == nil {
				return
			}
			if _, ok := allowFileSet[comparisonKey(fullPath)]; !ok {
				return
			}
			projections = []requestedPathProjection{{Path: fullPath}}
		}
		resolvedCanonicalPath := ""
		resolvedCanonicalParentPath := ""
		if needsRealpath && fsys != nil {
			identity := FreezeFileIdentity(fullPath, fsys)
			resolvedCanonicalPath = identity.CanonicalPath
			resolvedCanonicalParentPath = identity.CanonicalParentPath
		}
		for _, projection := range projections {
			canonicalPath := projection.CanonicalPath
			if resolvedCanonicalPath != "" {
				canonicalPath = resolvedCanonicalPath
			}
			if canonicalPath == "" {
				canonicalPath = fullPath
			}
			canonicalParentPath := resolvedCanonicalParentPath
			if canonicalParentPath == "" {
				canonicalParentPath = tspath.GetDirectoryPath(canonicalPath)
			}
			if rslintconfig.IsDefaultExcludedPath(projection.Path, scanRoot, useCaseSensitive) {
				continue
			}
			target := File{
				PathIdentity: rslintconfig.PathIdentity{
					Path:                projection.Path,
					CanonicalPath:       canonicalPath,
					CanonicalParentPath: canonicalParentPath,
				},
				ConfigDirectory: configDir,
			}
			if !includeDiscoveredFile(target) {
				continue
			}

			targetMu.Lock()
			addTarget(projection.Path, canonicalPath, canonicalParentPath)
			targetMu.Unlock()
		}
	}

	work := func(walkPath string) []string {
		directory := normalizedScanRoot
		if walkPath != "." {
			directory = tspath.CombinePaths(normalizedScanRoot, walkPath)
		}
		entries := readLintDirectory(fsys, directory)

		var childDirs []string
		for _, e := range entries {
			name := e.name
			if e.directory {
				if rslintconfig.IsDefaultExcludedDirectoryName(name, useCaseSensitive) {
					continue
				}
				childPath := path.Join(walkPath, name)
				if isStoppedWalkPath(childPath, stopWalkDirs, useCaseSensitive) {
					continue
				}
				if cached, ok := dirIgnore.Load(childPath); ok {
					blocked, _ := cached.(bool)
					if blocked {
						continue
					}
				} else {
					childFullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, childPath))
					blocked := allProjectedDirectoriesMatch(
						childFullPath,
						func(projection requestedPathProjection) bool {
							return targetMatcher.CanPruneDirectory(rslintconfig.DirectoryIdentity{
								LexicalPath:   projection.Path,
								CanonicalPath: projection.CanonicalPath,
							})
						},
					)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				processFile(path.Join(walkPath, name), e.needsRealpath)
			}
		}
		return childDirs
	}

	pool := newWalkPool(workers)
	walkRoots := discoverWalkRoots(normalizedScanRoot, allowDirs, fsys)
	walkRoots = filterInitialWalkRoots(
		walkRoots,
		func(walkPath string) bool {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
			return allProjectedDirectoriesMatch(
				fullPath,
				func(projection requestedPathProjection) bool {
					return rslintconfig.IsDefaultExcludedPath(projection.Path, scanRoot, useCaseSensitive)
				},
			)
		},
		func(walkPath string) bool {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
			return allProjectedDirectoriesMatch(
				fullPath,
				func(projection requestedPathProjection) bool {
					return targetMatcher.CanPruneDirectory(rslintconfig.DirectoryIdentity{
						LexicalPath:   projection.Path,
						CanonicalPath: projection.CanonicalPath,
					})
				},
			)
		},
		stopWalkDirs,
		useCaseSensitive,
	)
	pool.submitMany(walkRoots)
	pool.run(work)

	sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Path < targetFiles[j].Path })
	return targetFiles
}
