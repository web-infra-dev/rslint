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

func defaultLintFilePatterns() []string {
	patterns := make([]string, 0, len(DefaultLintFileExtensions))
	for _, ext := range DefaultLintFileExtensions {
		patterns = append(patterns, "**/*"+ext)
	}
	return patterns
}

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
// The implicit default baseline is always present; explicit `files` patterns
// extend it. Entry-level ignores do not remove a target from this selector.
func isFileSelectedByConfig(config RslintConfig, filePath string, configDir string) bool {
	if isDefaultLintFile(filePath) {
		return true
	}
	for _, entry := range config {
		if !isGlobalIgnoreEntry(entry) && hasFileSelectors(entry) && isFileMatchedByConfigEntry(filePath, entry, configDir) {
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
//   - Entry-level ignores affect config/rule selection, not target selection.
//     A selected file excluded by every entry is still parsed with zero rules.
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
	return discoverLintFilesWithStopDirs(config, configDir, fsys, allowFiles, allowDirs, nil, singleThreaded)
}

func discoverLintFilesWithStopDirs(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
) []string {
	globalIgnores := ExtractConfigIgnores(config)
	globalIgnores = append(ParseIgnorePatterns(utils.DefaultIgnoreDirGlobs()), globalIgnores...)
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	comparisonKey := func(filePath string) string {
		return string(tspath.ToPath(tspath.NormalizePath(filePath), "", useCaseSensitive))
	}
	configDir = tspath.NormalizePath(configDir)
	configMatchDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			configMatchDir = tspath.NormalizePath(realPath)
		}
	}
	configPathForMatching := func(filePath string) string {
		filePath = tspath.NormalizePath(filePath)
		if relative, ok := relativePathWithinConfigRoot(filePath, configDir, useCaseSensitive); ok {
			return tspath.ResolvePath(configMatchDir, relative)
		}
		if fsys != nil {
			if realPath := fsys.Realpath(filePath); realPath != "" {
				canonical := tspath.NormalizePath(realPath)
				if relative, ok := relativePathWithinConfigRoot(canonical, configMatchDir, useCaseSensitive); ok {
					return tspath.ResolvePath(configMatchDir, relative)
				}
			}
		}
		return filePath
	}

	filesPatterns := collectLintFilePatterns(config)
	filesMatcher := buildFilesMatcher(filesPatterns)

	var allowFileSet map[string]string
	if allowFiles != nil {
		allowFileSet = make(map[string]string, len(allowFiles))
		for _, f := range allowFiles {
			normalized := tspath.NormalizePath(f)
			key := comparisonKey(normalized)
			if _, exists := allowFileSet[key]; !exists {
				allowFileSet[key] = normalized
			}
		}
	}

	targetFiles := []string{}
	seenTargets := make(map[string]struct{})
	addTarget := func(filePath string) {
		key := comparisonKey(filePath)
		if _, seen := seenTargets[key]; seen {
			return
		}
		seenTargets[key] = struct{}{}
		targetFiles = append(targetFiles, filePath)
	}
	isGloballyIgnored := func(filePath string) bool {
		matchPath := configPathForMatching(filePath)
		return isDirBlockedByIgnores(matchPath, globalIgnores, configMatchDir) ||
			isFileIgnored(matchPath, globalIgnores, configMatchDir)
	}

	includeExplicitFile := func(filePath string) bool {
		if !IsSupportedLintFile(filePath) {
			return false
		}
		if fsys != nil && !fsys.FileExists(filePath) {
			return false
		}
		if IsDefaultExcludedPath(configPathForMatching(filePath), configMatchDir, useCaseSensitive) {
			return false
		}
		return !isGloballyIgnored(filePath)
	}

	includeDiscoveredFile := func(filePath string) bool {
		if !IsSupportedLintFile(filePath) {
			return false
		}
		matchPath := configPathForMatching(filePath)
		if len(filesPatterns) == 0 || !filesMatcher(matchPath, configMatchDir) {
			return false
		}
		// Candidate patterns for an AND group intentionally form a superset.
		// Apply the complete selector here so negated and additional group
		// members cannot leak candidates into the target set.
		if !isFileSelectedByConfig(config, matchPath, configMatchDir) {
			return false
		}
		if isGloballyIgnored(filePath) {
			return false
		}
		return true
	}

	addExplicitTargets := func() {
		for _, f := range allowFileSet {
			if includeExplicitFile(f) {
				addTarget(f)
			}
		}
	}

	// Fast path for explicit file-only invocations, e.g. lint-staged.
	if allowFileSet != nil && allowDirs == nil {
		addExplicitTargets()
		sort.Strings(targetFiles)
		return targetFiles
	}

	normalizedConfigDir := normalizeGlobPath(configDir)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedConfigDir}

	var (
		targetMu  sync.Mutex
		dirIgnore sync.Map // map[string]bool — pattern check cache
	)

	neg := buildNegReach(globalIgnores)
	stopWalkDirs := normalizeStopWalkDirs(normalizedConfigDir, stopDirs, useCaseSensitive)

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	processFile := func(walkPath string) {
		fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedConfigDir, walkPath))

		if allowFileSet != nil || allowDirs != nil {
			inScope := false
			if allowFileSet != nil {
				if _, ok := allowFileSet[comparisonKey(fullPath)]; ok {
					inScope = true
				}
			}
			if !inScope && allowDirs != nil {
				if isFileInAllowedDirs(fullPath, allowDirs, useCaseSensitive) {
					inScope = true
				}
			}
			if !inScope {
				return
			}
		}

		if !includeDiscoveredFile(fullPath) {
			return
		}

		targetMu.Lock()
		addTarget(fullPath)
		targetMu.Unlock()
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
					blocked := canPruneDir(childPath, globalIgnores, neg)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				processFile(path.Join(walkPath, name))
			}
		}
		return childDirs
	}

	pool := newWalkPool(workers)
	walkRoots := discoverWalkRoots(normalizedConfigDir, allowDirs, fsys)
	walkRoots = filterInitialWalkRoots(walkRoots, globalIgnores, neg, stopWalkDirs, useCaseSensitive)
	pool.submitMany(walkRoots)
	pool.run(work)

	if allowFileSet != nil {
		addExplicitTargets()
	}

	sort.Strings(targetFiles)
	return targetFiles
}

func relativePathWithinConfigRoot(filePath string, configDir string, useCaseSensitive bool) (string, bool) {
	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	if tspath.ComparePaths(filePath, configDir, compareOptions) == 0 {
		return "", true
	}
	if !tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive) {
		return "", false
	}
	return tspath.GetRelativePathFromDirectory(configDir, filePath, compareOptions), true
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
	globalIgnores []IgnorePattern,
	neg negReach,
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
		if root != "." {
			if hasDefaultExcludedSegment(root, useCaseSensitive) ||
				isStoppedWalkPath(root, stopWalkDirs, useCaseSensitive) ||
				canPruneDir(root, globalIgnores, neg) {
				continue
			}
		}
		filtered = append(filtered, root)
	}
	return filtered
}

// IsDefaultExcludedPath reports whether filePath contains one of rslint's
// always-excluded directory names, interpreted relative to configDir when
// possible.
func IsDefaultExcludedPath(filePath string, configDir string, useCaseSensitive bool) bool {
	filePath = tspath.NormalizePath(filePath)
	configDir = tspath.NormalizePath(configDir)
	if pathsEqual(filePath, configDir, useCaseSensitive) ||
		tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive) {
		rel := tspath.GetRelativePathFromDirectory(configDir, filePath, tspath.ComparePathsOptions{
			CurrentDirectory:          configDir,
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

func discoverWalkRoots(configDir string, allowDirs []string, fsys vfs.FS) []string {
	if allowDirs == nil {
		return []string{"."}
	}
	if len(allowDirs) == 0 {
		return nil
	}

	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
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
			if pathsEqual(existing, root, useCaseSensitive) ||
				tspath.StartsWithDirectory(root, existing, useCaseSensitive) {
				return
			}
		}
		filtered := roots[:0]
		seen = make(map[string]struct{}, len(allowDirs))
		for _, existing := range roots {
			if pathsEqual(existing, root, useCaseSensitive) ||
				tspath.StartsWithDirectory(existing, root, useCaseSensitive) {
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
		if pathsEqual(dir, configDir, useCaseSensitive) ||
			tspath.StartsWithDirectory(configDir, dir, useCaseSensitive) {
			return []string{"."}
		}
		if tspath.StartsWithDirectory(dir, configDir, useCaseSensitive) {
			addRoot(tspath.GetRelativePathFromDirectory(configDir, dir, compareOptions))
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

func collectLintFilePatterns(config RslintConfig) []string {
	patterns := defaultLintFilePatterns()
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		for _, pattern := range entry.Files {
			if candidate, ok := positiveFilesCandidate(pattern); ok {
				patterns = append(patterns, candidate)
			} else {
				patterns = append(patterns, "**/*")
			}
		}
		for _, group := range entry.FilePatternGroups {
			hasPositiveCandidate := false
			for _, pattern := range group {
				if candidate, ok := positiveFilesCandidate(pattern); ok {
					patterns = append(patterns, candidate)
					hasPositiveCandidate = true
				}
			}
			if !hasPositiveCandidate {
				patterns = append(patterns, "**/*")
			}
		}
	}
	return deduplicate(patterns)
}

func positiveFilesCandidate(pattern string) (string, bool) {
	negated := false
	for strings.HasPrefix(pattern, "!") {
		negated = !negated
		pattern = strings.TrimPrefix(pattern, "!")
	}
	return pattern, !negated && pattern != ""
}

func buildFilesMatcher(patterns []string) func(filePath string, configDir string) bool {
	hasDefaultBaseline := patternsIncludeAllDefaultExtensions(patterns)
	if !hasDefaultBaseline {
		return func(filePath string, configDir string) bool {
			return isFileMatched(filePath, patterns, configDir)
		}
	}

	defaultPatterns := make(map[string]struct{}, len(DefaultLintFileExtensions))
	for _, pattern := range defaultLintFilePatterns() {
		defaultPatterns[tspath.NormalizePath(pattern)] = struct{}{}
	}
	additionalPatterns := make([]string, 0, len(patterns)-len(defaultPatterns))
	for _, pattern := range patterns {
		if _, isDefault := defaultPatterns[tspath.NormalizePath(pattern)]; !isDefault {
			additionalPatterns = append(additionalPatterns, pattern)
		}
	}

	return func(filePath string, configDir string) bool {
		// ESLint's default extension globs are case-sensitive even on a
		// case-insensitive filesystem. Keep this fast path exact-case; explicit
		// user patterns are evaluated normally below.
		if isDefaultLintFile(filePath) {
			return true
		}
		return len(additionalPatterns) > 0 && isFileMatched(filePath, additionalPatterns, configDir)
	}
}

func patternsIncludeAllDefaultExtensions(patterns []string) bool {
	if len(patterns) < len(DefaultLintFileExtensions) {
		return false
	}
	seen := make(map[string]struct{}, len(patterns))
	for _, pattern := range patterns {
		seen[tspath.NormalizePath(pattern)] = struct{}{}
	}
	for _, pattern := range defaultLintFilePatterns() {
		if _, ok := seen[tspath.NormalizePath(pattern)]; !ok {
			return false
		}
	}
	return true
}

// LintDiscoveryScope overrides the CLI/API scope for one config.
type LintDiscoveryScope struct {
	Files []string
	Dirs  []string
}

// DiscoveredLintTarget preserves the config owner established during the
// directory walk so later stages do not need to infer ownership from paths.
type DiscoveredLintTarget struct {
	Path            string
	ConfigDirectory string
}

// DiscoverLintTargetsMultiConfig resolves owned lint targets across a config
// map. Per-config scopes override the default scope when present.
func DiscoverLintTargetsMultiConfig(
	configMap map[string]RslintConfig,
	scopes map[string]LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	if len(configMap) == 0 {
		return nil
	}

	index := newConfigDirectoryIndex(configMap, fsys)
	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	// Explicit files are assigned to their nearest config once. Passing the
	// complete list to every config makes lint-staged-style invocations
	// O(configs*files), and also asks every config to evaluate ignores for files
	// it cannot own. A non-nil empty bucket is preserved below so the explicit
	// file-only fast path still suppresses directory walking for configs that own
	// no requested files.
	filesByConfig := index.assignExplicitFiles(allowFiles)
	filesSpecifiedByConfig := make(map[string]bool, len(configDirs))
	if allowFiles != nil {
		for _, configDir := range configDirs {
			filesSpecifiedByConfig[configDir] = true
		}
	}
	for configDir, scope := range scopes {
		delete(filesByConfig, configDir)
		filesSpecifiedByConfig[configDir] = scope.Files != nil
	}
	for _, scope := range scopes {
		if scope.Files == nil {
			continue
		}
		for owner, files := range index.assignExplicitFiles(scope.Files) {
			filesByConfig[owner] = append(filesByConfig[owner], files...)
			filesSpecifiedByConfig[owner] = true
		}
	}

	seen := make(map[tspath.Path]struct{})
	var allTargets []DiscoveredLintTarget
	for _, configDir := range configDirs {
		var configAllowFiles []string
		if filesSpecifiedByConfig[configDir] {
			configAllowFiles = filesByConfig[configDir]
			if configAllowFiles == nil {
				configAllowFiles = []string{}
			}
		}
		configAllowDirs := allowDirs
		if scope, ok := scopes[configDir]; ok {
			configAllowDirs = scope.Dirs
		}
		targets := discoverLintFilesForConfigInMap(configMap, index, configDir, fsys, configAllowFiles, configAllowDirs, singleThreaded)
		for _, f := range targets {
			pathID := tspath.ToPath(tspath.NormalizePath(f), "", index.useCaseSensitive)
			if _, exists := seen[pathID]; !exists {
				seen[pathID] = struct{}{}
				allTargets = append(allTargets, DiscoveredLintTarget{
					Path:            f,
					ConfigDirectory: configDir,
				})
			}
		}
	}
	sort.Slice(allTargets, func(i, j int) bool {
		return allTargets[i].Path < allTargets[j].Path
	})
	return allTargets
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
	return discoverLintFilesForConfigInMap(
		configMap,
		newConfigDirectoryIndex(configMap, fsys),
		configDir,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
	)
}

func discoverLintFilesForConfigInMap(
	configMap map[string]RslintConfig,
	index *configDirectoryIndex,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	cfg, ok := configMap[configDir]
	if !ok {
		return nil
	}

	stopDirs := index.childConfigDirs(configDir)
	targets := discoverLintFilesWithStopDirs(cfg, configDir, fsys, allowFiles, allowDirs, stopDirs, singleThreaded)
	if len(targets) == 0 {
		return targets
	}

	ownedTargets := make([]string, 0, len(targets))
	for _, f := range targets {
		ownerDir, _ := index.nearestConfig(f)
		if ownerDir == configDir {
			ownedTargets = append(ownedTargets, f)
		}
	}
	return ownedTargets
}

type configDirectoryIndex struct {
	fsys                     vfs.FS
	useCaseSensitive         bool
	configKeyByPath          map[tspath.Path]string
	canonicalConfigKeyByPath map[tspath.Path]string
	ambiguousCanonicalPaths  map[tspath.Path]struct{}
	normalizedByKey          map[string]string
	childrenByKey            map[string][]string
}

func newConfigDirectoryIndex(configMap map[string]RslintConfig, fsys vfs.FS) *configDirectoryIndex {
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}

	index := &configDirectoryIndex{
		fsys:                     fsys,
		useCaseSensitive:         useCaseSensitive,
		configKeyByPath:          make(map[tspath.Path]string, len(configMap)),
		canonicalConfigKeyByPath: make(map[tspath.Path]string, len(configMap)),
		ambiguousCanonicalPaths:  make(map[tspath.Path]struct{}),
		normalizedByKey:          make(map[string]string, len(configMap)),
		childrenByKey:            make(map[string][]string, len(configMap)),
	}
	configKeys := make([]string, 0, len(configMap))
	for configKey := range configMap {
		configKeys = append(configKeys, configKey)
	}
	sort.Strings(configKeys)
	for _, configKey := range configKeys {
		normalized := tspath.NormalizePath(configKey)
		index.normalizedByKey[configKey] = normalized
		pathID := tspath.ToPath(normalized, "", useCaseSensitive)
		if _, exists := index.configKeyByPath[pathID]; !exists {
			index.configKeyByPath[pathID] = configKey
		}

		canonical := normalized
		if fsys != nil {
			if realPath := fsys.Realpath(normalized); realPath != "" {
				canonical = tspath.NormalizePath(realPath)
			}
		}
		canonicalID := tspath.ToPath(canonical, "", useCaseSensitive)
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
		for parent := tspath.GetDirectoryPath(normalized); parent != "" && parent != normalized; {
			if parentKey, ok := index.configKeyByPath[tspath.ToPath(parent, "", useCaseSensitive)]; ok {
				index.childrenByKey[parentKey] = append(index.childrenByKey[parentKey], normalized)
				break
			}
			next := tspath.GetDirectoryPath(parent)
			if next == parent {
				break
			}
			parent = next
		}
	}
	for configKey := range index.childrenByKey {
		sort.Strings(index.childrenByKey[configKey])
	}
	return index
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
	if configKey, ok := index.nearestConfigInPathSpace(filePath, index.configKeyByPath); ok {
		return configKey, true
	}
	if index.fsys == nil {
		return "", false
	}
	realPath := index.fsys.Realpath(filePath)
	if realPath == "" {
		return "", false
	}
	return index.nearestConfigInPathSpace(tspath.NormalizePath(realPath), index.canonicalConfigKeyByPath)
}

func (index *configDirectoryIndex) nearestConfigInPathSpace(filePath string, configKeyByPath map[tspath.Path]string) (string, bool) {
	current := tspath.GetDirectoryPath(filePath)
	for current != "" {
		if configKey, ok := configKeyByPath[tspath.ToPath(current, "", index.useCaseSensitive)]; ok {
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

func (index *configDirectoryIndex) assignExplicitFiles(files []string) map[string][]string {
	if index == nil || files == nil {
		return nil
	}
	filesByConfig := make(map[string][]string)
	for _, filePath := range files {
		if owner, ok := index.nearestConfig(filePath); ok {
			filesByConfig[owner] = append(filesByConfig[owner], filePath)
		}
	}
	return filesByConfig
}

// DiscoverGapFiles returns resolved lint targets that are absent from existing
// Programs. The filesystem walk and config/default-files matching are owned by
// DiscoverLintFiles; this helper only subtracts programFiles for callers that
// need an AST-only fallback Program.
//
// Files pass through these filters in DiscoverLintFiles:
//  1. Inside CLI/API file or directory scope
//  2. Selected by config `files` patterns or default lintable extensions;
//     explicit file targets may pass this step to produce an empty result
//  3. Not in global ignores or .gitignore-injected ignores
//  4. Not already in programFiles
//
// Walking model:
//   - A bounded worker pool (see walkPool) traverses the directory tree.
//     Total live goroutines at any moment is at most `workers`.
//   - Default workers = max(2, GOMAXPROCS); singleThreaded forces workers=1
//     for fully serial traversal (a knob users rely on for debugging,
//     reproducibility, and constrained environments).
//   - The vfsAdapter is constructed with followSymlinks=false, so symlinked
//     subdirectories are skipped entirely. This matches ESLint v10's
//     flat-config file walker, which uses @humanfs/node and recurses only
//     when Dirent.isDirectory() is true (Node returns false for symlinks).
//     It also avoids the otherwise scheduling-dependent "first writer wins"
//     non-determinism a parallel walker would introduce.
//
// When allowFiles/allowDirs are provided (CLI args), only files within scope.
//
// Returns:
//   - []: no gaps found
//   - [...]: gap files to create a fallback Program for (sorted lexically)
func DiscoverGapFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	gapFiles := []string{}
	targetFiles := DiscoverLintFiles(config, configDir, fsys, allowFiles, allowDirs, singleThreaded)
	for _, fullPath := range targetFiles {
		if _, exists := programFiles[fullPath]; exists {
			continue
		}
		gapFiles = append(gapFiles, fullPath)
	}
	return gapFiles
}

// DiscoverGapFilesMultiConfig runs DiscoverGapFiles for each config in a
// monorepo config map and returns the union of all gap files.
//
// Configs are processed serially. Each DiscoverGapFiles invocation already
// uses an internal worker pool, so the total live goroutine count is bounded
// by the worker pool size, not by `len(configMap) × workers`. Cross-config
// parallelism can be added later if benchmarks justify it.
func DiscoverGapFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	index := newConfigDirectoryIndex(configMap, fsys)
	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	seen := make(map[string]struct{})
	var allGapFiles []string
	for _, configDir := range configDirs {
		targetFiles := discoverLintFilesForConfigInMap(configMap, index, configDir, fsys, allowFiles, allowDirs, singleThreaded)
		for _, f := range targetFiles {
			if _, exists := programFiles[f]; exists {
				continue
			}
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allGapFiles = append(allGapFiles, f)
			}
		}
	}

	sort.Strings(allGapFiles)
	return allGapFiles
}

// walkPool is a fixed-size worker pool with an unbounded internal LIFO queue,
// used by DiscoverGapFiles to walk directory trees with a bounded number of
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

// deduplicate returns a copy of the input slice with duplicates removed, preserving order.
func deduplicate(items []string) []string {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
