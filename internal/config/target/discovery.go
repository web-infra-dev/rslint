package target

import (
	"path"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func discoverLintTargetsWithPreparedFiles(
	config rslintconfig.RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []rslintconfig.DirectoryIdentity,
	stopDirs []string,
	singleThreaded bool,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
	gitPolicy *rslintconfig.GitPolicy,
) []File {
	configDir = tspath.NormalizePath(configDir)
	scanRoot = tspath.NormalizePath(scanRoot)
	if scanRoot == "" {
		scanRoot = configDir
	}
	if allowDirs == nil {
		return discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, pathSpaces, gitPolicy,
		)
	}

	targets := []File{}
	if explicitFiles != nil {
		targets = discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, pathSpaces, gitPolicy,
		)
	}
	if len(allowDirs) == 0 {
		return targets
	}

	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	targetMatcher := mustTargetMatcher(config, configDir, fsys, pathSpaces, gitPolicy)
	for _, root := range allowDirs {
		if rslintconfig.IsDefaultExcludedPath(root.LexicalPath, scanRoot, useCaseSensitive) ||
			targetMatcher.CanPruneDirectory(root) {
			continue
		}
		targets = append(targets, discoverLintTargetsWithinRoot(
			config,
			configDir,
			root.LexicalPath,
			fsys,
			nil,
			[]rslintconfig.DirectoryIdentity{root},
			stopDirs,
			singleThreaded,
			pathSpaces,
			gitPolicy,
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
	allowDirs []rslintconfig.DirectoryIdentity,
	stopDirs []string,
	singleThreaded bool,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
	gitPolicy *rslintconfig.GitPolicy,
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

	resolvedAllowDirs := allowDirs
	directWalkRoot := frozenWalkRoot(scanRoot, allowDirs, pathSpaces)
	directWalkProjection := allowDirs == nil
	if len(resolvedAllowDirs) == 1 &&
		rslintconfig.ExactPathID(resolvedAllowDirs[0].LexicalPath) == rslintconfig.ExactPathID(scanRoot) {
		directWalkRoot = resolvedAllowDirs[0]
		directWalkProjection = true
	}

	targetMatcher := mustTargetMatcher(config, configDir, fsys, pathSpaces, gitPolicy)

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
	addTarget := func(target File) {
		target.Path = tspath.NormalizePath(target.Path)
		if target.CanonicalPath == "" {
			target.CanonicalPath = target.Path
		} else {
			target.CanonicalPath = tspath.NormalizePath(target.CanonicalPath)
		}
		if target.CanonicalParentPath == "" {
			target.CanonicalParentPath = tspath.GetDirectoryPath(target.CanonicalPath)
		} else {
			target.CanonicalParentPath = tspath.NormalizePath(target.CanonicalParentPath)
		}
		key := comparisonKey(target.Path)
		if _, seen := seenTargets[key]; seen {
			return
		}
		seenTargets[key] = struct{}{}
		target.ConfigDirectory = configDir
		targetFiles = append(targetFiles, target)
	}

	addExplicitTargets := func() {
		for _, explicitFile := range allowFileSet {
			if explicitFile.selectedBy(
				configDir,
				scanRoot,
				useCaseSensitive,
				&targetMatcher,
			) {
				addTarget(explicitFile.target)
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
	physicalWalkPath := func(walkPath string) string {
		fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
		for current := fullPath; current != ""; current = tspath.GetDirectoryPath(current) {
			if rslintconfig.ExactPathID(current) == rslintconfig.ExactPathID(directWalkRoot.LexicalPath) {
				break
			}
			if physical, frozen := pathSpaces.PhysicalDirectory(current); frozen {
				relative, within := rslintconfig.RelativePathWithinConfigRoot(fullPath, current, true)
				if within {
					return tspath.ResolvePath(physical, relative)
				}
			}
			if parent := tspath.GetDirectoryPath(current); parent == current {
				break
			}
		}
		return tspath.ResolvePath(directWalkRoot.CanonicalPath, walkPath)
	}

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

	pathProjectionsForWalk := func(fullPath, physicalPath string) []string {
		if directWalkProjection {
			return []string{fullPath}
		}
		return projectPathThroughRequestedDirectories(
			fullPath, physicalPath, resolvedAllowDirs, useCaseSensitive,
		)
	}
	allProjectedDirectoriesMatch := func(
		fullPath, physicalPath string,
		predicate func(rslintconfig.DirectoryIdentity) bool,
	) bool {
		if directWalkProjection {
			return predicate(rslintconfig.DirectoryIdentity{LexicalPath: fullPath, CanonicalPath: physicalPath})
		}
		projections := pathProjectionsForWalk(fullPath, physicalPath)
		if len(projections) == 0 {
			return false
		}
		for _, projection := range projections {
			if !predicate(rslintconfig.DirectoryIdentity{LexicalPath: projection, CanonicalPath: physicalPath}) {
				return false
			}
		}
		return true
	}

	processFile := func(walkPath, physicalPath string, needsRealpath bool) {
		fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
		projections := pathProjectionsForWalk(fullPath, physicalPath)
		if len(projections) == 0 {
			if allowFileSet == nil {
				return
			}
			if _, ok := allowFileSet[comparisonKey(fullPath)]; !ok {
				return
			}
			projections = []string{fullPath}
		}
		canonicalPath := physicalPath
		if needsRealpath && fsys != nil {
			if resolved := fsys.Realpath(physicalPath); resolved != "" {
				canonicalPath = resolved
			}
		}
		// A leaf symlink changes the file identity, never its containing directory.
		canonicalParentPath := tspath.GetDirectoryPath(physicalPath)
		for _, projection := range projections {
			if rslintconfig.IsDefaultExcludedPath(projection, scanRoot, useCaseSensitive) {
				continue
			}
			target := File{
				PathIdentity: rslintconfig.PathIdentity{
					Path:                projection,
					CanonicalPath:       canonicalPath,
					CanonicalParentPath: canonicalParentPath,
				},
				ConfigDirectory: configDir,
			}
			if !rslintconfig.IsSupportedLintFile(target.Path) {
				continue
			}
			match := targetMatcher.MatchFile(target.Identity())
			if !match.Selected() || match.GloballyIgnored() {
				continue
			}
			target.match = new(rslintconfig.TargetMatch)
			*target.match = match

			targetMu.Lock()
			addTarget(target)
			targetMu.Unlock()
		}
	}

	work := func(walkPath string) []string {
		physicalDirectory := physicalWalkPath(walkPath)
		entries := readLintDirectory(fsys, physicalDirectory)

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
						childFullPath, tspath.CombinePaths(physicalDirectory, name),
						targetMatcher.CanPruneDirectory,
					)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				processFile(path.Join(walkPath, name), tspath.CombinePaths(physicalDirectory, name), e.needsRealpath)
			}
		}
		return childDirs
	}

	pool := newWalkPool(workers)
	walkRoots := discoverWalkRoots(directWalkRoot, allowDirs)
	walkRoots = filterInitialWalkRoots(
		walkRoots,
		func(walkPath string) bool {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
			return allProjectedDirectoriesMatch(
				fullPath, physicalWalkPath(walkPath),
				func(directory rslintconfig.DirectoryIdentity) bool {
					return rslintconfig.IsDefaultExcludedPath(directory.LexicalPath, scanRoot, useCaseSensitive)
				},
			)
		},
		func(walkPath string) bool {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
			return allProjectedDirectoriesMatch(
				fullPath, physicalWalkPath(walkPath),
				targetMatcher.CanPruneDirectory,
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
