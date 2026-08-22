package config

import (
	"io/fs"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DefaultLintFileExtensions are the file extensions rslint discovers when a
// config entry omits `files`. This intentionally extends ESLint's default
// .js/.mjs/.cjs set with JSX and TypeScript-family files.
var DefaultLintFileExtensions = []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".mts", ".cts"}

var defaultLintFileExtensionSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(DefaultLintFileExtensions))
	for _, ext := range DefaultLintFileExtensions {
		m[ext] = struct{}{}
	}
	return m
}()

// IsSupportedLintFile reports whether rslint can parse and lint this path.
func IsSupportedLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[strings.ToLower(path.Ext(filePath))]
	return ok
}

func isDefaultLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[path.Ext(filePath)]
	return ok
}

// isFileSelectedByConfig reports whether the config itself selects filePath.
// The implicit default baseline is always present. An explicit `files` entry
// extends it only for paths not excluded by that same entry's `ignores`.
// Another matching entry or the default baseline may still select the path.
func isFileSelectedByConfig(config RslintConfig, filePath string, configDir string) bool {
	if configNeedsTargetResolver(config) {
		decision := newConfigTargetResolver(config, configDir, nil).resolve(filePath, "")
		return decision.selected && !decision.globallyIgnored
	}
	if isDefaultLintFile(filePath) {
		return true
	}
	for _, entry := range config {
		if !isGlobalIgnoreEntry(entry) &&
			hasFileSelectors(entry) &&
			isFileMatchedByConfigEntry(filePath, entry, configDir) &&
			!isFileIgnored(filePath, ParseIgnorePatterns(entry.Ignores), configDir) {
			return true
		}
	}
	return false
}

// defaultExcludeDirs is a set of directory names always excluded from walking.
var defaultExcludeDirs = func() map[string]struct{} {
	m := make(map[string]struct{}, len(utils.DefaultExcludeDirNames))
	for _, name := range utils.DefaultExcludeDirNames {
		m[name] = struct{}{}
	}
	return m
}()

// DiscoverLintFiles resolves the lint target set for one config directory.
// Target selection is independent from TypeScript Program membership:
//
//   - CLI/API files and directories first constrain the search space.
//   - Rslint's default lintable extensions are always selected. Non-global
//     config entries contribute additional `files` patterns.
//   - Global ignores, including injected .gitignore entries, remove files.
//   - Entry-level ignores prevent that entry from selecting or configuring a
//     file. They do not remove a file selected by the default baseline or a
//     different entry.
//   - Explicit file targets are retained even when they do not match any
//     config entry's files patterns, matching ESLint's empty-result behavior.
//
// Returned paths are absolute, normalized, deduplicated, and sorted.
func DiscoverLintFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := discoverLintTargetsWithStopDirs(
		config, configDir, configDir, fsys, allowFiles, allowDirs, nil, singleThreaded, nil,
	)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

// DiscoverLintTargets is the identity-preserving form of DiscoverLintFiles.
// CanonicalPath is derived without a per-file realpath call for regular files
// reached by directory traversal. Explicit directory aliases are resolved once
// and projected onto their descendants; explicit files and file symlinks are
// resolved individually because their physical target cannot be inferred from
// the containing directory.
func DiscoverLintTargets(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	return discoverLintTargetsFromRoot(config, configDir, configDir, fsys, allowFiles, allowDirs, singleThreaded)
}

// discoverLintTargetsFromRoot is the single-config target walker with separate
// authored-config and filesystem-scan directories. Most callers use the same
// directory through DiscoverLintTargets. An explicit external config keeps its
// own ConfigDirectory on every result while searching only ScanRoot and the
// requested file/directory scope below it.
func discoverLintTargetsFromRoot(
	config RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	return discoverLintTargetsWithStopDirs(
		config,
		configDir,
		scanRoot,
		fsys,
		allowFiles,
		allowDirs,
		nil,
		singleThreaded,
		nil,
	)
}

func discoverLintTargetsWithStopDirs(
	config RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
	frozenBases map[string]configTargetBase,
) []DiscoveredLintTarget {
	explicitSet := newExplicitLintTargetSet(allowFiles, fsys)
	return discoverLintTargetsWithPreparedFiles(
		config,
		configDir,
		scanRoot,
		fsys,
		explicitSet.targetsForPaths(allowFiles),
		allowDirs,
		stopDirs,
		singleThreaded,
		frozenBases,
	)
}

func discoverLintTargetsWithPreparedFiles(
	config RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
	frozenBases map[string]configTargetBase,
) []DiscoveredLintTarget {
	configDir = tspath.NormalizePath(configDir)
	scanRoot = tspath.NormalizePath(scanRoot)
	if scanRoot == "" {
		scanRoot = configDir
	}
	if allowDirs == nil {
		return discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, frozenBases,
		)
	}

	targets := []DiscoveredLintTarget{}
	if explicitFiles != nil {
		targets = discoverLintTargetsWithinRoot(
			config, configDir, scanRoot, fsys, explicitFiles, nil, stopDirs, singleThreaded, frozenBases,
		)
	}
	if len(allowDirs) == 0 {
		return targets
	}

	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	targetResolver := newConfigTargetResolverWithBases(config, configDir, fsys, frozenBases)
	scanRootCanonical := scanRoot
	if fsys != nil {
		if realPath := fsys.Realpath(scanRoot); realPath != "" {
			scanRootCanonical = tspath.NormalizePath(realPath)
		}
	}
	for _, root := range coalesceRequestedDirectories(allowDirs, fsys) {
		isDefaultRoot := pathsEqual(root.lexicalPath, scanRoot, true) ||
			pathsEqual(root.canonicalPath, scanRootCanonical, true)
		if !isDefaultRoot && (IsDefaultExcludedPath(root.lexicalPath, scanRoot, useCaseSensitive) ||
			targetResolver.canPruneDirectory(root.lexicalPath, root.canonicalPath)) {
			continue
		}
		targets = append(targets, discoverLintTargetsWithinRoot(
			config,
			configDir,
			root.lexicalPath,
			fsys,
			nil,
			[]string{root.lexicalPath},
			stopDirs,
			singleThreaded,
			frozenBases,
		)...)
	}

	seen := make(map[string]struct{}, len(targets))
	deduplicated := targets[:0]
	for _, target := range targets {
		key := exactPathID(target.Path)
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
	config RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
	frozenBases map[string]configTargetBase,
) []DiscoveredLintTarget {
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
	configBase, frozen := frozenBases[exactPathID(configDir)]
	if !frozen {
		configBase = freezeConfigTargetBase(configDir, fsys)
	}
	configMatchDir := configBase.physicalDirectory
	configAliasAncestors := configBase.physicalAliasAncestors
	configPathForMatching := func(filePath string) string {
		matchPath, _ := resolveConfigPathSpace(
			tspath.NormalizePath(filePath),
			"",
			configDir,
			configMatchDir,
			configAliasAncestors,
			fsys,
			true,
		)
		return matchPath
	}
	resolvedAllowDirs := resolveAllowedDirectories(allowDirs, fsys)
	directWalkRoot := resolvedAllowedDirectory{
		lexicalPath:   scanRoot,
		canonicalPath: scanRoot,
	}
	directWalkProjection := allowDirs == nil
	if len(resolvedAllowDirs) == 1 &&
		exactPathID(resolvedAllowDirs[0].lexicalPath) == exactPathID(scanRoot) {
		directWalkRoot = resolvedAllowDirs[0]
		directWalkProjection = true
	}
	directProjectionForPath := func(fullPath string) (requestedPathProjection, bool) {
		projection := requestedPathProjection{path: fullPath}
		if directWalkRoot.lexicalPath == directWalkRoot.canonicalPath {
			return projection, true
		}
		relative, within := RelativePathWithinConfigRoot(
			fullPath,
			directWalkRoot.lexicalPath,
			true,
		)
		if !within {
			return requestedPathProjection{}, false
		}
		projection.canonicalPath = tspath.ResolvePath(
			directWalkRoot.canonicalPath,
			relative,
		)
		return projection, true
	}

	targetResolver := newConfigTargetResolverWithBases(config, configDir, fsys, frozenBases)

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

	targetFiles := []DiscoveredLintTarget{}
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
		targetFiles = append(targetFiles, DiscoveredLintTarget{
			Path:                filePath,
			CanonicalPath:       canonicalPath,
			CanonicalParentPath: canonicalParentPath,
			ConfigDirectory:     configDir,
		})
	}
	includeDiscoveredFile := func(target DiscoveredLintTarget) bool {
		if !IsSupportedLintFile(target.Path) {
			return false
		}
		decision := targetResolver.resolveTarget(target)
		return decision.selected && !decision.globallyIgnored
	}

	addExplicitTargets := func() {
		for _, explicitFile := range allowFileSet {
			if explicitFile.selectedBy(
				configDir,
				scanRoot,
				useCaseSensitive,
				targetResolver,
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

	normalizedScanRoot := normalizeGlobPath(scanRoot)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedScanRoot}

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
			projections = []requestedPathProjection{{path: fullPath}}
		}
		resolvedCanonicalPath := ""
		resolvedCanonicalParentPath := ""
		if needsRealpath && fsys != nil {
			identity := FreezeLintTargetIdentity(fullPath, fsys)
			resolvedCanonicalPath = identity.CanonicalPath
			resolvedCanonicalParentPath = identity.CanonicalParentPath
		}
		for _, projection := range projections {
			canonicalPath := projection.canonicalPath
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
			if IsDefaultExcludedPath(projection.path, scanRoot, useCaseSensitive) {
				continue
			}
			target := DiscoveredLintTarget{
				Path:                projection.path,
				CanonicalPath:       canonicalPath,
				CanonicalParentPath: canonicalParentPath,
				ConfigDirectory:     configDir,
			}
			if !includeDiscoveredFile(target) {
				continue
			}

			targetMu.Lock()
			addTarget(projection.path, canonicalPath, canonicalParentPath)
			targetMu.Unlock()
		}
	}

	work := func(walkPath string) []string {
		f, err := fsAdapter.Open(walkPath)
		if err != nil {
			return nil
		}
		rdf, ok := f.(fs.ReadDirFile)
		if !ok {
			f.Close()
			return nil
		}
		entries, _ := rdf.ReadDir(-1)
		f.Close()

		var childDirs []string
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				if isDefaultExcludedDirName(name, useCaseSensitive) {
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
							return targetResolver.canPruneDirectory(
								projection.path,
								projection.canonicalPath,
							)
						},
					)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				needsRealpath := e.Type()&fs.ModeSymlink != 0
				if entry, ok := e.(interface{ needsCanonicalRealpath() bool }); ok {
					needsRealpath = needsRealpath || entry.needsCanonicalRealpath()
				}
				processFile(path.Join(walkPath, name), needsRealpath)
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
					return IsDefaultExcludedPath(projection.path, scanRoot, useCaseSensitive)
				},
			)
		},
		func(walkPath string) bool {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedScanRoot, walkPath))
			return allProjectedDirectoriesMatch(
				fullPath,
				func(projection requestedPathProjection) bool {
					return targetResolver.canPruneDirectory(
						projection.path,
						projection.canonicalPath,
					)
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

// FreezeLintTargetIdentity captures the lexical file plus a physical
// file/parent pair. Reading the parent on both sides detects one concurrent
// directory-symlink move and retries, reducing the chance of combining two
// filesystem observations without claiming an atomic filesystem snapshot.
func FreezeLintTargetIdentity(filePath string, fsys vfs.FS) DiscoveredLintTarget {
	filePath = tspath.NormalizePath(filePath)
	parentPath := tspath.GetDirectoryPath(filePath)
	canonicalPath := filePath
	canonicalParentPath := parentPath
	if fsys == nil {
		return DiscoveredLintTarget{
			Path:                filePath,
			CanonicalPath:       canonicalPath,
			CanonicalParentPath: canonicalParentPath,
		}
	}
	for range 2 {
		parentBefore := canonicalPathOrSelf(parentPath, fsys)
		canonicalPath = canonicalPathOrSelf(filePath, fsys)
		parentAfter := canonicalPathOrSelf(parentPath, fsys)
		canonicalParentPath = parentAfter
		if parentBefore == parentAfter {
			break
		}
	}
	return DiscoveredLintTarget{
		Path:                filePath,
		CanonicalPath:       canonicalPath,
		CanonicalParentPath: canonicalParentPath,
	}
}

func canonicalPathOrSelf(filePath string, fsys vfs.FS) string {
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return tspath.NormalizePath(filePath)
}

func normalizeStopWalkDirs(configDir string, stopDirs []string, useCaseSensitive bool) []string {
	if len(stopDirs) == 0 {
		return nil
	}

	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	seen := make(map[string]struct{}, len(stopDirs))
	normalized := make([]string, 0, len(stopDirs))
	for _, rawDir := range stopDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, useCaseSensitive) ||
			!tspath.StartsWithDirectory(dir, configDir, useCaseSensitive) {
			continue
		}
		rel := tspath.NormalizePath(tspath.GetRelativePathFromDirectory(configDir, dir, compareOptions))
		if rel == "" || rel == "." {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		normalized = append(normalized, rel)
	}
	sort.Strings(normalized)
	return normalized
}

func filterInitialWalkRoots(
	roots []string,
	isDefaultExcluded func(string) bool,
	canPrune func(string) bool,
	stopWalkDirs []string,
	useCaseSensitive bool,
) []string {
	if len(roots) == 0 {
		return roots
	}

	filtered := make([]string, 0, len(roots))
	for _, root := range roots {
		root = tspath.NormalizePath(root)
		if root == "" {
			root = "."
		}
		// A root represented as "." may still be nested below a default-excluded
		// directory in the config's path space when ConfigDirectory and ScanRoot
		// differ (for example an external config invoked from node_modules/pkg).
		if isDefaultExcluded != nil && isDefaultExcluded(root) {
			continue
		}
		if root != "." {
			if isStoppedWalkPath(root, stopWalkDirs, useCaseSensitive) ||
				(canPrune != nil && canPrune(root)) {
				continue
			}
		}
		filtered = append(filtered, root)
	}
	return filtered
}

// IsDefaultExcludedPath reports whether filePath contains one of rslint's
// always-excluded directory names, interpreted relative to scanRoot when
// possible. Default excludes belong to the invocation's lexical scan scope,
// not to any config entry's authored path base.
func IsDefaultExcludedPath(filePath string, scanRoot string, useCaseSensitive bool) bool {
	filePath = tspath.NormalizePath(filePath)
	scanRoot = tspath.NormalizePath(scanRoot)
	if hasDefaultExcludedSegment(scanRoot, useCaseSensitive) {
		return true
	}
	if pathsEqual(filePath, scanRoot, useCaseSensitive) ||
		tspath.StartsWithDirectory(filePath, scanRoot, useCaseSensitive) {
		rel := tspath.GetRelativePathFromDirectory(scanRoot, filePath, tspath.ComparePathsOptions{
			CurrentDirectory:          scanRoot,
			UseCaseSensitiveFileNames: useCaseSensitive,
		})
		return hasDefaultExcludedSegment(rel, useCaseSensitive)
	}
	return hasDefaultExcludedSegment(filePath, useCaseSensitive)
}

func isDefaultExcludedDirName(name string, useCaseSensitive bool) bool {
	for excluded := range defaultExcludeDirs {
		if pathsEqual(name, excluded, useCaseSensitive) {
			return true
		}
	}
	return false
}

func hasDefaultExcludedSegment(walkPath string, useCaseSensitive bool) bool {
	for _, segment := range strings.Split(walkPath, "/") {
		if isDefaultExcludedDirName(segment, useCaseSensitive) {
			return true
		}
	}
	return false
}

func isStoppedWalkPath(walkPath string, stopWalkDirs []string, useCaseSensitive bool) bool {
	if len(stopWalkDirs) == 0 {
		return false
	}
	walkPath = tspath.NormalizePath(walkPath)
	if walkPath == "" || walkPath == "." {
		return false
	}
	for _, stopDir := range stopWalkDirs {
		if pathsEqual(walkPath, stopDir, useCaseSensitive) ||
			tspath.StartsWithDirectory(walkPath, stopDir, useCaseSensitive) {
			return true
		}
	}
	return false
}

func isFileInAllowedDirs(filePath string, allowDirs []string, useCaseSensitive bool) bool {
	for _, dir := range allowDirs {
		dir = tspath.NormalizePath(dir)
		if pathsEqual(filePath, dir, useCaseSensitive) ||
			tspath.StartsWithDirectory(filePath, dir, useCaseSensitive) {
			return true
		}
	}
	return false
}

func isFileInAllowedDirsWithFS(filePath string, allowDirs []string, fsys vfs.FS) bool {
	if isFileInAllowedDirs(filePath, allowDirs, true) {
		return true
	}
	if fsys == nil {
		return false
	}
	realFilePath := fsys.Realpath(filePath)
	if realFilePath == "" {
		return false
	}
	canonicalFile := tspath.NormalizePath(realFilePath)
	if canonicalFile == tspath.NormalizePath(filePath) {
		return false
	}
	for _, dir := range allowDirs {
		realDir := fsys.Realpath(tspath.NormalizePath(dir))
		if realDir == "" {
			continue
		}
		canonicalDir := tspath.NormalizePath(realDir)
		if pathsEqual(canonicalFile, canonicalDir, true) ||
			tspath.StartsWithDirectory(canonicalFile, canonicalDir, true) {
			return true
		}
	}
	return false
}

type resolvedAllowedDirectory struct {
	lexicalPath   string
	canonicalPath string
}

type requestedPathProjection struct {
	path          string
	canonicalPath string
}

// coalesceRequestedDirectories returns the smallest set of independent walk
// roots that covers the user's directory arguments. A nested root is dropped
// only when both lexical and canonical containment produce the same relative
// suffix, so a differently spelled symlink scope remains independent. Keeping
// roots independent also avoids invalid common-parent calculations across
// Windows drives and UNC shares.
func coalesceRequestedDirectories(directories []string, fsys vfs.FS) []resolvedAllowedDirectory {
	if len(directories) == 0 {
		return nil
	}
	roots := make([]resolvedAllowedDirectory, 0, len(directories))
	contains := func(parent resolvedAllowedDirectory, child resolvedAllowedDirectory) bool {
		lexicalRelative, lexicalWithin := RelativePathWithinConfigRoot(
			child.lexicalPath,
			parent.lexicalPath,
			true,
		)
		canonicalRelative, canonicalWithin := RelativePathWithinConfigRoot(
			child.canonicalPath,
			parent.canonicalPath,
			true,
		)
		return lexicalWithin && canonicalWithin &&
			exactPathID(lexicalRelative) == exactPathID(canonicalRelative)
	}
	for _, directory := range directories {
		lexicalPath := tspath.NormalizePath(directory)
		canonicalPath := lexicalPath
		if fsys != nil {
			if realPath := fsys.Realpath(lexicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		candidate := resolvedAllowedDirectory{
			lexicalPath:   lexicalPath,
			canonicalPath: canonicalPath,
		}
		covered := false
		for _, existing := range roots {
			if contains(existing, candidate) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		kept := roots[:0]
		for _, existing := range roots {
			if !contains(candidate, existing) {
				kept = append(kept, existing)
			}
		}
		roots = append(kept, candidate)
	}
	return roots
}

func resolveAllowedDirectories(allowDirs []string, fsys vfs.FS) []resolvedAllowedDirectory {
	if allowDirs == nil {
		return nil
	}
	resolved := make([]resolvedAllowedDirectory, 0, len(allowDirs))
	for _, dir := range allowDirs {
		lexicalPath := tspath.NormalizePath(dir)
		canonicalPath := lexicalPath
		if fsys != nil {
			if realPath := fsys.Realpath(lexicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		resolved = append(resolved, resolvedAllowedDirectory{
			lexicalPath:   lexicalPath,
			canonicalPath: canonicalPath,
		})
	}
	return resolved
}

// projectPathThroughRequestedDirectories returns every caller-visible spelling
// of one walked physical path. Multiple requested directory aliases may point
// at the same subtree, and each spelling remains a legitimate config target;
// callers perform selection for every projection and canonical-deduplicate
// only after selection.
func projectPathThroughRequestedDirectories(
	filePath string,
	matchPath string,
	allowDirs []resolvedAllowedDirectory,
	useCaseSensitive bool,
) []requestedPathProjection {
	projections := make([]requestedPathProjection, 0, len(allowDirs))
	seen := make(map[string]struct{}, len(allowDirs))
	appendProjection := func(path string, canonicalPath string) {
		path = tspath.NormalizePath(path)
		if canonicalPath != "" {
			canonicalPath = tspath.NormalizePath(canonicalPath)
		}
		key := exactPathID(path)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		projections = append(projections, requestedPathProjection{
			path:          path,
			canonicalPath: canonicalPath,
		})
	}
	for _, dir := range allowDirs {
		if _, within := RelativePathWithinConfigRoot(filePath, dir.lexicalPath, useCaseSensitive); within {
			relative, exactWithin := RelativePathWithinConfigRoot(
				filePath,
				dir.lexicalPath,
				true,
			)
			if !exactWithin {
				// A case-insensitive lexical resemblance is authoritative only
				// when the frozen physical spelling verifies the same root. This
				// keeps native casing aliases while rejecting a distinct tree that
				// differs only by case on a case-sensitive filesystem facade.
				physicalRelative, verified := RelativePathWithinConfigRoot(
					filePath,
					dir.canonicalPath,
					true,
				)
				if !verified && matchPath != filePath {
					physicalRelative, verified = RelativePathWithinConfigRoot(
						matchPath,
						dir.canonicalPath,
						true,
					)
				}
				if !verified {
					continue
				}
				relative = physicalRelative
			}
			canonicalPath := ""
			if dir.lexicalPath != dir.canonicalPath {
				canonicalPath = tspath.ResolvePath(dir.canonicalPath, relative)
			}
			appendProjection(tspath.ResolvePath(dir.lexicalPath, relative), canonicalPath)
		}
		if dir.lexicalPath == dir.canonicalPath {
			continue
		}
		relative, within := RelativePathWithinConfigRoot(filePath, dir.canonicalPath, true)
		if !within && matchPath != filePath {
			relative, within = RelativePathWithinConfigRoot(matchPath, dir.canonicalPath, true)
		}
		if within {
			appendProjection(
				tspath.ResolvePath(dir.lexicalPath, relative),
				tspath.ResolvePath(dir.canonicalPath, relative),
			)
		}
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].path < projections[j].path
	})
	return projections
}

func discoverWalkRoots(configDir string, allowDirs []string, fsys vfs.FS) []string {
	if allowDirs == nil {
		return []string{"."}
	}
	if len(allowDirs) == 0 {
		return nil
	}

	configDir = tspath.NormalizePath(configDir)
	canonicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			canonicalConfigDir = tspath.NormalizePath(realPath)
		}
	}
	seen := make(map[string]struct{}, len(allowDirs))
	roots := make([]string, 0, len(allowDirs))
	addRoot := func(root string) {
		if root == "" {
			root = "."
		}
		root = tspath.NormalizePath(root)
		if root == "." {
			roots = []string{"."}
			seen = map[string]struct{}{".": {}}
			return
		}
		for _, existing := range roots {
			if existing == "." {
				return
			}
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(root, existing, true) {
				return
			}
		}
		filtered := roots[:0]
		seen = make(map[string]struct{}, len(allowDirs))
		for _, existing := range roots {
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(existing, root, true) {
				continue
			}
			seen[existing] = struct{}{}
			filtered = append(filtered, existing)
		}
		roots = filtered
		if _, ok := seen[root]; ok {
			return
		}
		seen[root] = struct{}{}
		roots = append(roots, root)
	}

	for _, rawDir := range allowDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, true) ||
			tspath.StartsWithDirectory(configDir, dir, true) {
			return []string{"."}
		}
		if relative, within := RelativePathWithinConfigRoot(dir, configDir, true); within {
			addRoot(relative)
			continue
		}
		if fsys == nil {
			continue
		}
		realDir := fsys.Realpath(dir)
		if realDir == "" {
			continue
		}
		canonicalDir := tspath.NormalizePath(realDir)
		if pathsEqual(canonicalDir, canonicalConfigDir, true) ||
			tspath.StartsWithDirectory(canonicalConfigDir, canonicalDir, true) {
			return []string{"."}
		}
		if relative, within := RelativePathWithinConfigRoot(canonicalDir, canonicalConfigDir, true); within {
			addRoot(relative)
		}
	}

	sort.Strings(roots)
	return roots
}

func pathsEqual(a, b string, useCaseSensitive bool) bool {
	if useCaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

// DiscoveredLintTarget preserves the caller-visible path, physical file and
// parent identities, and config owner established at the target boundary.
// CanonicalParentPath distinguishes a leaf file symlink from a directory
// alias without asking the live filesystem again during config matching.
type DiscoveredLintTarget struct {
	Path                string
	CanonicalPath       string
	CanonicalParentPath string
	ConfigDirectory     string
}

// ExplicitFileOutcome records the target-selection facts needed by a caller
// that reports skipped literal files. Ignored and Exists are deliberately
// independent: the CLI historically reports an ignored missing path as
// ignored, so presentation can preserve that priority without re-reading the
// filesystem or re-running config ownership after the lint plan is frozen.
type ExplicitFileOutcome struct {
	Path    string
	Ignored bool
	Exists  bool
}

type explicitLintTarget struct {
	target    DiscoveredLintTarget
	supported bool
	exists    bool
	ignored   bool
	evaluated bool
}

type explicitLintTargetSet struct {
	byPath         map[tspath.Path]*explicitLintTarget
	requestedPaths []string
}

func (target *explicitLintTarget) selectedBy(
	configDirectory string,
	scanRoot string,
	useCaseSensitive bool,
	resolver *configTargetResolver,
) bool {
	if target == nil {
		return false
	}
	if !target.evaluated {
		target.target.ConfigDirectory = tspath.NormalizePath(configDirectory)
		target.ignored = IsDefaultExcludedPath(
			target.target.Path,
			scanRoot,
			useCaseSensitive,
		)
		if resolver != nil && resolver.resolveTarget(target.target).globallyIgnored {
			target.ignored = true
		}
		target.evaluated = true
	}
	return target.supported && target.exists && !target.ignored
}

func newExplicitLintTargetSet(
	filePaths []string,
	fsys vfs.FS,
) *explicitLintTargetSet {
	if filePaths == nil {
		return nil
	}
	set := &explicitLintTargetSet{
		byPath:         make(map[tspath.Path]*explicitLintTarget, len(filePaths)),
		requestedPaths: make([]string, 0, len(filePaths)),
	}
	set.add(filePaths, true, fsys)
	return set
}

func (set *explicitLintTargetSet) add(
	filePaths []string,
	requested bool,
	fsys vfs.FS,
) {
	if set == nil {
		return
	}
	for _, filePath := range filePaths {
		filePath = tspath.NormalizePath(filePath)
		if requested {
			set.requestedPaths = append(set.requestedPaths, filePath)
		}
		pathID := tspath.ToPath(filePath, "", true)
		if _, exists := set.byPath[pathID]; exists {
			continue
		}
		exists := true
		if fsys != nil {
			exists = fsys.FileExists(filePath)
		}
		set.byPath[pathID] = &explicitLintTarget{
			target:    FreezeLintTargetIdentity(filePath, fsys),
			supported: IsSupportedLintFile(filePath),
			exists:    exists,
		}
	}
}

func (set *explicitLintTargetSet) targetsForPaths(filePaths []string) []*explicitLintTarget {
	if filePaths == nil {
		return nil
	}
	targets := make([]*explicitLintTarget, 0, len(filePaths))
	seen := make(map[tspath.Path]struct{}, len(filePaths))
	for _, filePath := range filePaths {
		pathID := tspath.ToPath(tspath.NormalizePath(filePath), "", true)
		if _, exists := seen[pathID]; exists {
			continue
		}
		seen[pathID] = struct{}{}
		if target := set.byPath[pathID]; target != nil {
			targets = append(targets, target)
		}
	}
	return targets
}

func (set *explicitLintTargetSet) outcomes() []ExplicitFileOutcome {
	if set == nil || len(set.requestedPaths) == 0 {
		return nil
	}
	outcomes := make([]ExplicitFileOutcome, 0, len(set.requestedPaths))
	for _, filePath := range set.requestedPaths {
		target := set.byPath[tspath.ToPath(filePath, "", true)]
		if target == nil {
			continue
		}
		outcomes = append(outcomes, ExplicitFileOutcome{
			Path:    filePath,
			Ignored: target.ignored,
			Exists:  target.exists,
		})
	}
	return outcomes
}

// DiscoverLintTargetsMultiConfig resolves owned lint targets across a config
// map. Scope files are already assigned to their config by the discovery catalog.
func DiscoverLintTargetsMultiConfig(
	configMap map[string]RslintConfig,
	scopes map[string]LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	return discoverLintTargetsMultiConfigWithOwner(
		configMap,
		scopes,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
		NewConfigOwnerResolver(configMap, fsys),
	)
}

func discoverLintTargetsMultiConfigWithOwner(
	configMap map[string]RslintConfig,
	scopes map[string]LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
	ownerResolver *ConfigOwnerResolver,
) []DiscoveredLintTarget {
	explicitSet := newExplicitLintTargetSet(allowFiles, fsys)
	return discoverLintTargetsMultiConfigWithPreparedFiles(
		configMap,
		scopes,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
		ownerResolver,
		explicitSet,
	)
}

func discoverLintTargetsMultiConfigWithPreparedFiles(
	configMap map[string]RslintConfig,
	scopes map[string]LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
	ownerResolver *ConfigOwnerResolver,
	explicitSet *explicitLintTargetSet,
) []DiscoveredLintTarget {
	if len(configMap) == 0 {
		return nil
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	// A config found only because a literal file bypassed its parent's global
	// ignore is not a general ownership boundary. Build automatic ownership and
	// subtree handoff decisions from the normally reachable config set; the full
	// map is still processed below so catalog-scoped literal files use their
	// explicit-only config.
	automaticConfigMap := configMapForAutomaticTargets(configMap, scopes)
	automaticOwnerResolver := newConfigOwnerResolverWithBases(
		automaticConfigMap,
		fsys,
		ownerResolver.frozenBases,
	)

	// Explicit files are assigned to their nearest config once. Passing the
	// complete list to every config makes lint-staged-style invocations
	// O(configs*files), and also asks every config to evaluate ignores for files
	// it cannot own. A non-nil empty bucket is preserved below so the explicit
	// file-only fast path still suppresses directory walking for configs that own
	// no requested files.
	assignedExplicitOwners := make(map[tspath.Path]string)
	for _, configDir := range configDirs {
		scope, ok := scopes[configDir]
		if !ok || scope.Files == nil {
			continue
		}
		for _, filePath := range scope.Files {
			key := tspath.ToPath(tspath.NormalizePath(filePath), "", true)
			if _, exists := assignedExplicitOwners[key]; !exists {
				assignedExplicitOwners[key] = configDir
			}
		}
	}
	if explicitSet == nil {
		explicitSet = newExplicitLintTargetSet([]string{}, fsys)
	}
	for _, scope := range scopes {
		explicitSet.add(scope.Files, false, fsys)
	}
	filesByConfig := make(map[string][]*explicitLintTarget)
	for _, explicitFile := range explicitSet.targetsForPaths(allowFiles) {
		pathID := tspath.ToPath(explicitFile.target.Path, "", true)
		if _, assigned := assignedExplicitOwners[pathID]; assigned {
			continue
		}
		owner, _ := automaticOwnerResolver.ResolveTarget(explicitFile.target)
		if owner != "" {
			assignedExplicitOwners[pathID] = owner
			filesByConfig[owner] = appendUniqueExplicitLintTarget(
				filesByConfig[owner],
				explicitFile,
			)
		}
	}
	filesSpecifiedByConfig := make(map[string]bool, len(configDirs))
	if allowFiles != nil {
		for _, configDir := range configDirs {
			filesSpecifiedByConfig[configDir] = true
		}
	}
	for configDir, scope := range scopes {
		if scope.Files == nil {
			continue
		}
		// A directory can be reached both automatically and through a literal
		// target whose parent-ignore exception selected another candidate. Config
		// discovery resolves that collision to one automatic boundary; retain its
		// automatically assigned files when adding the catalog-owned literal scope.
		for _, explicitFile := range explicitSet.targetsForPaths(scope.Files) {
			filesByConfig[configDir] = appendUniqueExplicitLintTarget(
				filesByConfig[configDir],
				explicitFile,
			)
		}
		filesSpecifiedByConfig[configDir] = true
	}

	seen := make(map[tspath.Path]struct{})
	var allTargets []DiscoveredLintTarget
	for _, configDir := range configDirs {
		var configAllowFiles []*explicitLintTarget
		if filesSpecifiedByConfig[configDir] {
			configAllowFiles = filesByConfig[configDir]
			if configAllowFiles == nil {
				configAllowFiles = []*explicitLintTarget{}
			}
		}
		configAllowDirs := allowDirs
		if scopes[configDir].ExplicitOnly {
			configAllowDirs = []string{}
		}
		targets := discoverLintTargetsForConfigInMap(
			configMap,
			automaticOwnerResolver,
			assignedExplicitOwners,
			configDir,
			fsys,
			configAllowFiles,
			configAllowDirs,
			singleThreaded,
			ownerResolver.frozenBases,
		)
		for _, target := range targets {
			pathID := tspath.ToPath(tspath.NormalizePath(target.Path), "", true)
			if _, exists := seen[pathID]; !exists {
				seen[pathID] = struct{}{}
				allTargets = append(allTargets, target)
			}
		}
	}
	sort.Slice(allTargets, func(i, j int) bool {
		return allTargets[i].Path < allTargets[j].Path
	})
	return allTargets
}

func appendUniqueExplicitLintTarget(
	targets []*explicitLintTarget,
	target *explicitLintTarget,
) []*explicitLintTarget {
	if target == nil {
		return targets
	}
	pathID := tspath.ToPath(target.target.Path, "", true)
	for _, existing := range targets {
		if existing != nil && tspath.ToPath(existing.target.Path, "", true) == pathID {
			return targets
		}
	}
	return append(targets, target)
}

// DiscoverLintFilesMultiConfig resolves lint target paths across a config map.
func DiscoverLintFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := DiscoverLintTargetsMultiConfig(configMap, nil, fsys, allowFiles, allowDirs, singleThreaded)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

// DiscoverLintFilesForConfigInMap resolves lint targets owned by one config in
// a multi-config map. Descendant config directories are treated as handoff
// boundaries so parent configs don't walk subtrees that a nearer config owns.
func DiscoverLintFilesForConfigInMap(
	configMap map[string]RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	ownerResolver := NewConfigOwnerResolver(configMap, fsys)
	explicitSet := newExplicitLintTargetSet(allowFiles, fsys)
	targets := discoverLintTargetsForConfigInMap(
		configMap,
		ownerResolver,
		nil,
		configDir,
		fsys,
		explicitSet.targetsForPaths(allowFiles),
		allowDirs,
		singleThreaded,
		ownerResolver.frozenBases,
	)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

func discoverLintTargetsForConfigInMap(
	configMap map[string]RslintConfig,
	ownerResolver *ConfigOwnerResolver,
	assignedExplicitOwners map[tspath.Path]string,
	configDir string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	singleThreaded bool,
	frozenBases map[string]configTargetBase,
) []DiscoveredLintTarget {
	cfg, ok := configMap[configDir]
	if !ok {
		return nil
	}

	stopDirs := ownerResolver.ChildConfigDirs(configDir)
	targets := discoverLintTargetsWithinRoot(
		cfg,
		configDir,
		configDir,
		fsys,
		explicitFiles,
		allowDirs,
		stopDirs,
		singleThreaded,
		frozenBases,
	)
	if len(targets) == 0 {
		return targets
	}

	ownedTargets := make([]DiscoveredLintTarget, 0, len(targets))
	for _, target := range targets {
		targetID := tspath.ToPath(tspath.NormalizePath(target.Path), "", true)
		if assignedOwner, assigned := assignedExplicitOwners[targetID]; assigned {
			if assignedOwner == configDir {
				target.ConfigDirectory = configDir
				ownedTargets = append(ownedTargets, target)
			}
			continue
		}
		ownerDir, _ := ownerResolver.ResolveTarget(target)
		if ownerDir == configDir {
			target.ConfigDirectory = configDir
			ownedTargets = append(ownedTargets, target)
		}
	}
	return ownedTargets
}

type configDirectoryIndex struct {
	fsys                     vfs.FS
	configKeyByPath          map[tspath.Path]string
	caseFoldedConfigKeys     map[tspath.Path][]string
	canonicalConfigKeyByPath map[tspath.Path]string
	ambiguousCanonicalPaths  map[tspath.Path]struct{}
	normalizedByKey          map[string]string
	canonicalByKey           map[string]string
	childrenByKey            map[string][]string
}

func newConfigDirectoryIndex(configMap map[string]RslintConfig, fsys vfs.FS) *configDirectoryIndex {
	return newConfigDirectoryIndexWithBases(configMap, fsys, nil)
}

func newConfigDirectoryIndexWithBases(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	frozenBases map[string]configTargetBase,
) *configDirectoryIndex {
	index := &configDirectoryIndex{
		fsys:                     fsys,
		configKeyByPath:          make(map[tspath.Path]string, len(configMap)),
		caseFoldedConfigKeys:     make(map[tspath.Path][]string, len(configMap)),
		canonicalConfigKeyByPath: make(map[tspath.Path]string, len(configMap)),
		ambiguousCanonicalPaths:  make(map[tspath.Path]struct{}),
		normalizedByKey:          make(map[string]string, len(configMap)),
		canonicalByKey:           make(map[string]string, len(configMap)),
		childrenByKey:            make(map[string][]string, len(configMap)),
	}
	configKeys := make([]string, 0, len(configMap))
	for configKey := range configMap {
		configKeys = append(configKeys, configKey)
	}
	sort.Strings(configKeys)
	for _, configKey := range configKeys {
		normalized := tspath.NormalizePath(configKey)
		if len(normalized) > tspath.GetRootLength(normalized) {
			normalized = tspath.RemoveTrailingDirectorySeparators(normalized)
		}
		index.normalizedByKey[configKey] = normalized
		pathID := tspath.ToPath(normalized, "", true)
		if _, exists := index.configKeyByPath[pathID]; !exists {
			index.configKeyByPath[pathID] = configKey
		}
		foldedPathID := tspath.ToPath(normalized, "", false)
		index.caseFoldedConfigKeys[foldedPathID] = append(index.caseFoldedConfigKeys[foldedPathID], configKey)

		canonical := normalized
		if base, frozen := frozenBases[exactPathID(normalized)]; frozen {
			canonical = base.physicalDirectory
		} else if fsys != nil {
			if realPath := fsys.Realpath(normalized); realPath != "" {
				canonical = tspath.NormalizePath(realPath)
			}
		}
		index.canonicalByKey[configKey] = canonical
		canonicalID := tspath.ToPath(canonical, "", true)
		if _, ambiguous := index.ambiguousCanonicalPaths[canonicalID]; ambiguous {
			continue
		}
		if existing, exists := index.canonicalConfigKeyByPath[canonicalID]; !exists {
			index.canonicalConfigKeyByPath[canonicalID] = configKey
		} else if existing != configKey {
			// Lexical aliases remain independently addressable. A physical-path
			// fallback cannot choose between them, so leave it unresolved instead
			// of silently assigning the file to the first map entry.
			delete(index.canonicalConfigKeyByPath, canonicalID)
			index.ambiguousCanonicalPaths[canonicalID] = struct{}{}
		}
	}

	for _, configKey := range configKeys {
		normalized := index.normalizedByKey[configKey]
		if parentKey, ok := index.nearestLexicalConfigAncestor(normalized); ok {
			index.addChildBoundary(parentKey, normalized)
		}
	}
	for configKey := range index.childrenByKey {
		sort.Strings(index.childrenByKey[configKey])
	}
	return index
}

func (index *configDirectoryIndex) nearestLexicalConfigAncestor(configDir string) (string, bool) {
	current := tspath.GetDirectoryPath(configDir)
	for current != "" && current != configDir {
		if configKey, ok := index.configKeyForLexicalDirectory(current); ok {
			return configKey, true
		}
		next := tspath.GetDirectoryPath(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (index *configDirectoryIndex) addChildBoundary(configKey string, boundary string) {
	boundary = tspath.NormalizePath(boundary)
	for _, existing := range index.childrenByKey[configKey] {
		if existing == boundary {
			return
		}
	}
	index.childrenByKey[configKey] = append(index.childrenByKey[configKey], boundary)
}

func (index *configDirectoryIndex) childConfigDirs(configKey string) []string {
	if index == nil {
		return nil
	}
	return index.childrenByKey[configKey]
}

func (index *configDirectoryIndex) nearestConfig(filePath string) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath = tspath.NormalizePath(filePath)
	if configKey, ok := index.nearestLexicalConfig(filePath); ok {
		return configKey, true
	}
	if index.fsys == nil {
		return "", false
	}
	realPath := index.fsys.Realpath(filePath)
	if realPath == "" {
		return "", false
	}
	return index.nearestConfigInPathSpace(
		tspath.NormalizePath(realPath),
		index.canonicalConfigKeyByPath,
	)
}

func (index *configDirectoryIndex) nearestConfigWithCanonicalPath(
	filePath string,
	canonicalPath string,
) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath = tspath.NormalizePath(filePath)
	lexicalKey, lexicalFound := index.nearestLexicalConfig(filePath)
	if lexicalFound {
		return lexicalKey, true
	}
	if canonicalPath == "" {
		return "", false
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	return index.nearestConfigInPathSpace(
		canonicalPath,
		index.canonicalConfigKeyByPath,
	)
}

func (index *configDirectoryIndex) nearestLexicalConfig(filePath string) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath = tspath.NormalizePath(filePath)
	current := tspath.GetDirectoryPath(filePath)
	for current != "" {
		if configKey, ok := index.configKeyForLexicalDirectory(current); ok {
			return configKey, true
		}
		next := tspath.GetDirectoryPath(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (index *configDirectoryIndex) configKeyForLexicalDirectory(directory string) (string, bool) {
	if index == nil {
		return "", false
	}
	if configKey, ok := index.configKeyByPath[tspath.ToPath(directory, "", true)]; ok {
		return configKey, true
	}
	if index.fsys == nil {
		return "", false
	}
	candidates := index.caseFoldedConfigKeys[tspath.ToPath(directory, "", false)]
	if len(candidates) == 0 {
		return "", false
	}
	canonicalDirectory := index.fsys.Realpath(directory)
	if canonicalDirectory == "" {
		return "", false
	}
	canonicalDirectory = tspath.NormalizePath(canonicalDirectory)
	for _, configKey := range candidates {
		if pathsEqual(canonicalDirectory, index.canonicalByKey[configKey], true) {
			return configKey, true
		}
	}
	return "", false
}

func (index *configDirectoryIndex) nearestConfigInPathSpace(
	filePath string,
	configKeyByPath map[tspath.Path]string,
) (string, bool) {
	current := tspath.GetDirectoryPath(filePath)
	for current != "" {
		if configKey, ok := configKeyByPath[tspath.ToPath(current, "", true)]; ok {
			return configKey, true
		}
		next := tspath.GetDirectoryPath(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

// walkPool is a fixed-size worker pool with an unbounded internal LIFO queue,
// used by lint-target discovery to walk directory trees with a bounded number of
// live goroutines. Properties:
//
//   - At most `workers` goroutines exist concurrently.
//   - submitMany never blocks (queue grows as needed; total memory ~ peak
//     queue size, bounded by FS branching × depth).
//   - run() returns once all submitted work and all transitively submitted
//     work have completed. Detection: when every worker is simultaneously
//     idle and the queue is empty.
//
// workers=1 degenerates to an effectively serial DFS-style traversal (LIFO
// pops the most recently submitted dir); the only goroutine is the worker
// itself. This is what --singleThreaded relies on.
type walkPool struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []string
	idle    int
	workers int
	done    bool
}

func newWalkPool(workers int) *walkPool {
	if workers < 1 {
		workers = 1
	}
	p := &walkPool{workers: workers}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// submitMany appends dirs to the queue and wakes idle workers. Safe to call
// from any goroutine.
func (p *walkPool) submitMany(dirs []string) {
	if len(dirs) == 0 {
		return
	}
	p.mu.Lock()
	p.queue = append(p.queue, dirs...)
	p.mu.Unlock()
	if len(dirs) == 1 {
		p.cond.Signal()
	} else {
		p.cond.Broadcast()
	}
}

// take pops a job from the queue, blocking if empty. Returns ("", false)
// only when the queue is empty AND every worker is simultaneously idle —
// which means no more work will ever appear.
func (p *walkPool) take() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		if len(p.queue) > 0 {
			n := len(p.queue) - 1
			dir := p.queue[n]
			p.queue = p.queue[:n]
			return dir, true
		}
		p.idle++
		if p.idle == p.workers {
			p.done = true
			p.cond.Broadcast()
			return "", false
		}
		p.cond.Wait()
		if p.done {
			return "", false
		}
		p.idle--
	}
}

// run drives the worker pool. Each worker pulls jobs and calls work(dir),
// which returns the child directories to enqueue. Returns when all reachable
// work is processed.
//
// When workers == 1, runs the loop on the calling goroutine directly — no
// goroutines are spawned at all. This is what --singleThreaded relies on:
// callers can rely on the Go side spawning no extra goroutines.
func (p *walkPool) run(work func(string) []string) {
	if p.workers == 1 {
		for {
			dir, ok := p.take()
			if !ok {
				return
			}
			p.submitMany(work(dir))
		}
	}
	var wg sync.WaitGroup
	for range p.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				dir, ok := p.take()
				if !ok {
					return
				}
				p.submitMany(work(dir))
			}
		}()
	}
	wg.Wait()
}
