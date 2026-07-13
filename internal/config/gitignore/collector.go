package gitignore

import (
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func normalizeGlobPath(path string) string {
	return strings.ReplaceAll(tspath.NormalizePath(path), "\\", "/")
}

func matchGlob(pattern string, path string) bool {
	matched, err := doublestar.Match(pattern, path)
	return err == nil && matched
}

func isDefaultExcludedDirName(name string, useCaseSensitive bool) bool {
	for _, excluded := range utils.DefaultExcludeDirNames {
		if name == excluded || (!useCaseSensitive && strings.EqualFold(name, excluded)) {
			return true
		}
	}
	return false
}

// Collect reads the .gitignore files relevant to one lint invocation. A nil
// targetFiles slice scans descendants of configDir; a non-nil slice reads only
// the directory chains between configDir and the explicit files. In full-scan
// mode, isDirectoryBlocked may prune descendants; its argument is a
// slash-separated directory path relative to configDir. The callback is not
// used for explicit target files.
func Collect(configDir string, fsys vfs.FS, targetFiles []string, isDirectoryBlocked func(string) bool) []string {
	return CollectWithBoundaries(configDir, fsys, targetFiles, isDirectoryBlocked, nil)
}

// CollectWithBoundaries is Collect with additional descendant handoff
// boundaries. A boundary directory and everything below it belongs to another
// config, so none of its .gitignore files participate in this collection.
func CollectWithBoundaries(configDir string, fsys vfs.FS, targetFiles []string, isDirectoryBlocked func(string) bool, stopDirs []string) []string {
	if targetFiles == nil {
		return readGitignoreAsGlobsWithBoundaries(configDir, fsys, isDirectoryBlocked, stopDirs)
	}
	return readGitignoreAsGlobsForFilesWithBoundaries(configDir, fsys, targetFiles, stopDirs)
}

// readGitignoreAsGlobs reads .gitignore files relevant to configDir and
// converts their patterns to rslint glob format suitable for use as global
// ignore patterns.
//
// It walks DOWN from configDir, collecting its .gitignore and nested
// .gitignore files with directory-scoped prefixes. configDir is a hard upper
// boundary: .gitignore files in its parents are never read.
//
// isDirectoryBlocked applies the caller's global config ignores. Descendant
// .gitignore files below a blocked directory are irrelevant because files in
// that directory cannot be linted.
//
// Returns nil if no .gitignore files are found.
func readGitignoreAsGlobs(configDir string, fsys vfs.FS, isDirectoryBlocked func(string) bool) []string {
	return readGitignoreAsGlobsWithBoundaries(configDir, fsys, isDirectoryBlocked, nil)
}

func readGitignoreAsGlobsWithBoundaries(configDir string, fsys vfs.FS, isDirectoryBlocked func(string) bool, stopDirs []string) []string {
	if fsys == nil {
		return nil
	}
	normalizedRoot := normalizeGlobPath(configDir)
	boundaries := normalizeCollectionBoundaries(normalizedRoot, stopDirs, fsys.UseCaseSensitiveFileNames())
	var allGlobs []string

	collectGitignoreGlobs(normalizedRoot, "", fsys, &allGlobs, isDirectoryBlocked, nil, boundaries)

	if len(allGlobs) == 0 {
		return nil
	}
	return allGlobs
}

// readGitignoreAsGlobsForFilesWithBoundaries reads only .gitignore files on
// each directory chain from configDir to an explicit target. This is used by
// API-style calls where the target set is already known; unlike the full
// collector, it does not scan every descendant of configDir.
func readGitignoreAsGlobsForFilesWithBoundaries(configDir string, fsys vfs.FS, files []string, stopDirs []string) []string {
	if fsys == nil || len(files) == 0 {
		return nil
	}

	normalizedConfigDir := normalizeGlobPath(configDir)
	useCaseSensitive := fsys.UseCaseSensitiveFileNames()
	boundaries := normalizeCollectionBoundaries(normalizedConfigDir, stopDirs, useCaseSensitive)
	dirSet := make(map[string]struct{})
	for _, file := range files {
		targetDir := dirOfPath(normalizeGlobPath(file))
		rel, ok := relativeDir(normalizedConfigDir, targetDir, useCaseSensitive)
		if !ok {
			continue
		}

		current := normalizedConfigDir
		dirSet[current] = struct{}{}
		for _, component := range splitPathComponents(rel) {
			current = tspath.CombinePaths(current, component)
			if isCollectionBoundary(current, boundaries, useCaseSensitive) {
				break
			}
			dirSet[current] = struct{}{}
		}
	}

	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	sortByPathDepth(dirs)

	var allGlobs []string
	var pruneRules []gitignorePruneRule
	var prunedDirs []string
	for _, dir := range dirs {
		if isUnderPrunedDir(dir, prunedDirs, useCaseSensitive) {
			continue
		}
		if isDescendantSymlinkDir(normalizedConfigDir, dir, fsys) {
			prunedDirs = append(prunedDirs, dir)
			continue
		}
		if isDirIgnoredByPruneRules(dir, pruneRules, useCaseSensitive) {
			prunedDirs = append(prunedDirs, dir)
			continue
		}

		content, ok := fsys.ReadFile(tspath.CombinePaths(dir, ".gitignore"))
		if !ok {
			continue
		}
		rel, _ := relativeDir(normalizedConfigDir, dir, useCaseSensitive)
		allGlobs = append(allGlobs, convertGitignoreToGlobs(content, rel)...)
		if rule, ok := newGitignorePruneRule(dir, content); ok {
			pruneRules = append(pruneRules, rule)
		}
	}
	if len(allGlobs) == 0 {
		return nil
	}
	return allGlobs
}

func normalizeCollectionBoundaries(configDir string, stopDirs []string, useCaseSensitive bool) []string {
	boundaries := make([]string, 0, len(stopDirs))
	for _, stopDir := range stopDirs {
		stopDir = normalizeGlobPath(stopDir)
		rel, ok := relativeDir(configDir, stopDir, useCaseSensitive)
		if !ok || rel == "" {
			continue
		}
		boundaries = append(boundaries, stopDir)
	}
	return boundaries
}

func isCollectionBoundary(dir string, boundaries []string, useCaseSensitive bool) bool {
	for _, boundary := range boundaries {
		if tspath.ComparePaths(dir, boundary, tspath.ComparePathsOptions{
			UseCaseSensitiveFileNames: useCaseSensitive,
		}) == 0 {
			return true
		}
	}
	return false
}

func isUnderPrunedDir(dir string, prunedDirs []string, useCaseSensitive bool) bool {
	for _, prunedDir := range prunedDirs {
		if _, ok := relativeDir(prunedDir, dir, useCaseSensitive); ok {
			return true
		}
	}
	return false
}

func isDescendantSymlinkDir(configDir string, dir string, fsys vfs.FS) bool {
	rel, ok := relativeDir(configDir, dir, fsys.UseCaseSensitiveFileNames())
	if !ok || rel == "" {
		return false
	}

	parent := parentDir(dir)
	name := tspath.GetBaseFileName(dir)
	entries := fsys.GetAccessibleEntries(parent)
	if entries.Symlinks != nil {
		for symlink := range entries.Symlinks {
			if symlink == name || (!fsys.UseCaseSensitiveFileNames() && strings.EqualFold(symlink, name)) {
				return true
			}
		}
		return false
	}

	parentRealPath := fsys.Realpath(parent)
	dirRealPath := fsys.Realpath(dir)
	if parentRealPath == "" || dirRealPath == "" {
		return false
	}
	expectedRealPath := tspath.CombinePaths(parentRealPath, name)
	return tspath.ComparePaths(dirRealPath, expectedRealPath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: fsys.UseCaseSensitiveFileNames(),
	}) != 0
}

type gitignorePruneRule struct {
	baseDir  string
	patterns []gitignorePrunePattern
}

type gitignorePrunePattern struct {
	negated bool
	rooted  bool
	pattern string
}

func newGitignorePruneRule(baseDir string, content string) (gitignorePruneRule, bool) {
	patterns := parseGitignorePrunePatterns(content)
	if len(patterns) == 0 {
		return gitignorePruneRule{}, false
	}
	return gitignorePruneRule{baseDir: baseDir, patterns: patterns}, true
}

func parseGitignorePrunePatterns(content string) []gitignorePrunePattern {
	var patterns []gitignorePrunePattern
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = line[1:]
		}
		if containsParentPathComponent(line) {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		rooted := false
		if strings.HasPrefix(line, "/") {
			rooted = true
			line = strings.TrimPrefix(line, "/")
		} else if strings.Contains(line, "/") {
			rooted = true
		}
		if line == "" {
			continue
		}
		patterns = append(patterns, gitignorePrunePattern{
			negated: negated,
			rooted:  rooted,
			pattern: line,
		})
	}
	return patterns
}

func isDirIgnoredByPruneRules(dir string, rules []gitignorePruneRule, useCaseSensitive bool) bool {
	ignored := false
	for _, rule := range rules {
		rel, ok := relativeDir(rule.baseDir, dir, useCaseSensitive)
		if !ok || rel == "" {
			continue
		}
		for _, pattern := range rule.patterns {
			if gitignorePrunePatternCoversDir(pattern, rel) {
				ignored = !pattern.negated
			}
		}
	}
	return ignored
}

func gitignorePrunePatternCoversDir(pattern gitignorePrunePattern, relDir string) bool {
	for current := relDir; current != ""; current = parentDir(current) {
		if gitignorePrunePatternMatchesPath(pattern, current) {
			return true
		}
	}
	return false
}

func gitignorePrunePatternMatchesPath(pattern gitignorePrunePattern, relPath string) bool {
	if pattern.rooted {
		return matchGlob(pattern.pattern, relPath)
	}
	for _, part := range splitPathComponents(relPath) {
		if matchGlob(pattern.pattern, part) {
			return true
		}
	}
	return false
}

type filesystemPath struct {
	root            string
	rest            string
	caseInsensitive bool
}

// splitFilesystemPath treats a UNC server/share pair as the volume root.
// tspath's generic root parser stops at the server, which is appropriate for
// URLs but would let filesystem traversal escape above a Windows share.
func splitFilesystemPath(path string) filesystemPath {
	path = tspath.NormalizePath(path)
	if strings.HasPrefix(path, "//") {
		serverAndRest := path[2:]
		serverEnd := strings.Index(serverAndRest, "/")
		if serverEnd < 0 {
			return filesystemPath{root: path, caseInsensitive: true}
		}

		shareStart := 2 + serverEnd + 1
		shareAndRest := path[shareStart:]
		shareEnd := strings.Index(shareAndRest, "/")
		if shareAndRest == "" {
			return filesystemPath{root: path, caseInsensitive: true}
		}

		rootEnd := len(path)
		if shareEnd >= 0 {
			rootEnd = shareStart + shareEnd
		}
		root := strings.TrimSuffix(path[:rootEnd], "/") + "/"
		rest := strings.Trim(path[rootEnd:], "/")
		return filesystemPath{root: root, rest: rest, caseInsensitive: true}
	}

	rootLength := tspath.GetRootLength(path)
	if rootLength == 0 {
		return filesystemPath{rest: strings.Trim(path, "/")}
	}

	root := path[:rootLength]
	caseInsensitive := len(root) >= 2 && root[1] == ':'
	if caseInsensitive && strings.HasSuffix(root, ":") {
		root += "/"
	}
	return filesystemPath{
		root:            root,
		rest:            strings.Trim(path[rootLength:], "/"),
		caseInsensitive: caseInsensitive,
	}
}

func joinFilesystemPath(path filesystemPath, rest string) string {
	if rest == "" {
		return path.root
	}
	if path.root == "" {
		return rest
	}
	return tspath.CombinePaths(path.root, rest)
}

func sameFilesystemRoot(left filesystemPath, right filesystemPath, useCaseSensitive bool) bool {
	if left.caseInsensitive != right.caseInsensitive {
		return false
	}
	if left.caseInsensitive || !useCaseSensitive {
		return strings.EqualFold(left.root, right.root)
	}
	return left.root == right.root
}

func equalFilesystemPath(left string, right string, caseInsensitive bool) bool {
	if caseInsensitive {
		return strings.EqualFold(left, right)
	}
	return left == right
}

// parentDir returns the parent directory of dir. Filesystem roots are returned
// as the parent of their direct children and "" as their own parent.
func parentDir(dir string) string {
	path := splitFilesystemPath(dir)
	if path.rest == "" {
		return ""
	}

	idx := strings.LastIndex(path.rest, "/")
	if idx < 0 {
		return path.root
	}
	return joinFilesystemPath(path, path.rest[:idx])
}

func dirOfPath(filePath string) string {
	path := splitFilesystemPath(filePath)
	if path.rest == "" {
		return path.root
	}

	idx := strings.LastIndex(path.rest, "/")
	if idx < 0 {
		return path.root
	}
	return joinFilesystemPath(path, path.rest[:idx])
}

func relativeDir(root string, dir string, useCaseSensitive bool) (string, bool) {
	rootPath := splitFilesystemPath(root)
	dirPath := splitFilesystemPath(dir)
	if !sameFilesystemRoot(rootPath, dirPath, useCaseSensitive) {
		return "", false
	}
	caseInsensitive := rootPath.caseInsensitive || !useCaseSensitive
	if equalFilesystemPath(rootPath.rest, dirPath.rest, caseInsensitive) {
		return "", true
	}
	if rootPath.rest == "" {
		return dirPath.rest, true
	}

	prefix := rootPath.rest + "/"
	if len(dirPath.rest) > len(prefix) && equalFilesystemPath(prefix, dirPath.rest[:len(prefix)], caseInsensitive) {
		return dirPath.rest[len(prefix):], true
	}
	return "", false
}

func sortByPathDepth(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		depthI := filesystemPathDepth(paths[i])
		depthJ := filesystemPathDepth(paths[j])
		if depthI != depthJ {
			return depthI < depthJ
		}
		return paths[i] < paths[j]
	})
}

func filesystemPathDepth(path string) int {
	rest := splitFilesystemPath(path).rest
	if rest == "" {
		return 0
	}
	return strings.Count(rest, "/") + 1
}

// collectGitignoreGlobs recursively scans for .gitignore files and converts
// their patterns to globs. Raw patterns from parent .gitignore files are used
// to prune directories during scanning.
//
// isDirectoryBlocked provides directory-level pruning from the caller's
// config policy. If it returns true, files below that directory cannot be
// linted, so nested .gitignore sources there are irrelevant.
func collectGitignoreGlobs(absDir string, relDir string, fsys vfs.FS, result *[]string, isDirectoryBlocked func(string) bool, pruneRules []gitignorePruneRule, boundaries []string) {
	collectGitignoreGlobsRecursive(absDir, relDir, fsys, result, isDirectoryBlocked, pruneRules, boundaries, &gitignoreWalkState{
		resolvedPaths: make(map[string]string),
		visited:       make(map[string]struct{}),
	})
}

type gitignoreWalkState struct {
	resolvedPaths map[string]string
	visited       map[string]struct{}
}

func (s *gitignoreWalkState) realpath(path string, fsys vfs.FS) string {
	if realpath, ok := s.resolvedPaths[path]; ok {
		return realpath
	}
	realpath := fsys.Realpath(path)
	s.resolvedPaths[path] = realpath
	return realpath
}

func collectGitignoreGlobsRecursive(absDir string, relDir string, fsys vfs.FS, result *[]string, isDirectoryBlocked func(string) bool, pruneRules []gitignorePruneRule, boundaries []string, state *gitignoreWalkState) {
	gitignorePath := tspath.CombinePaths(absDir, ".gitignore")
	if content, ok := fsys.ReadFile(gitignorePath); ok {
		localGlobs := convertGitignoreToGlobs(content, relDir)
		*result = append(*result, localGlobs...)
		if rule, ok := newGitignorePruneRule(absDir, content); ok {
			pruneRules = append(pruneRules, rule)
		}
	}

	entries := fsys.GetAccessibleEntries(absDir)
	parentRealPath := ""
	if entries.Symlinks == nil && len(entries.Directories) > 0 {
		parentRealPath = state.realpath(absDir, fsys)
		state.visited[parentRealPath] = struct{}{}
	}
	for _, dir := range entries.Directories {
		if isDefaultExcludedDirName(dir, fsys.UseCaseSensitiveFileNames()) {
			continue
		}

		childAbs := tspath.CombinePaths(absDir, dir)
		if isCollectionBoundary(childAbs, boundaries, fsys.UseCaseSensitiveFileNames()) {
			continue
		}
		childRealPath := ""
		if entries.Symlinks != nil {
			if _, isSymlink := entries.Symlinks[dir]; isSymlink {
				continue
			}
		} else {
			childRealPath = state.realpath(childAbs, fsys)
			expectedRealPath := tspath.CombinePaths(parentRealPath, dir)
			if tspath.ComparePaths(childRealPath, expectedRealPath, tspath.ComparePathsOptions{
				UseCaseSensitiveFileNames: fsys.UseCaseSensitiveFileNames(),
			}) != 0 {
				continue
			}
		}

		childRel := dir
		if relDir != "" {
			childRel = relDir + "/" + dir
		}

		if isDirectoryBlocked != nil && isDirectoryBlocked(childRel) {
			continue
		}

		// Prune directories already ignored by collected .gitignore patterns.
		if isDirIgnoredByPruneRules(childAbs, pruneRules, fsys.UseCaseSensitiveFileNames()) {
			continue
		}

		if childRealPath != "" {
			if _, visited := state.visited[childRealPath]; visited {
				continue
			}
			state.visited[childRealPath] = struct{}{}
		}

		collectGitignoreGlobsRecursive(childAbs, childRel, fsys, result, isDirectoryBlocked, pruneRules, boundaries, state)
	}
}

// convertGitignoreToGlobs converts .gitignore file content into rslint glob
// patterns. baseDir is the directory containing the .gitignore relative to
// the config root (empty for root .gitignore).
//
// Conversion rules (aligned with git spec):
//
//	"dist/"     → "**/dist/**"        (unrooted dir, matches at any depth)
//	"/dist"     → "dist/**"           (root-anchored)
//	"*.log"     → "**/*.log"          (unrooted file pattern)
//	"src/dist"  → "src/dist/**"       (contains /, implicitly rooted)
//	"!dist/"    → "!**/dist/**"       (negation preserved)
//
// For nested .gitignore (baseDir != ""), patterns are prefixed:
//
//	baseDir="pkg/app", "tmp/" → "pkg/app/**/tmp/**"
func convertGitignoreToGlobs(content string, baseDir string) []string {
	var globs []string
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		glob := convertSinglePattern(line, baseDir)
		if glob != "" {
			globs = append(globs, glob)
		}
	}
	return globs
}

func normalizeGitignoreLine(line string) (string, bool) {
	// Strip trailing whitespace (git spec)
	line = strings.TrimRight(line, " \t\r")
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	// Strip leading whitespace
	line = strings.TrimLeft(line, " \t")
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	return line, true
}

func splitPathComponents(p string) []string {
	var parts []string
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}
	return parts
}

// convertSinglePattern converts one gitignore pattern line to a glob.
func convertSinglePattern(line string, baseDir string) string {
	negated := false
	if strings.HasPrefix(line, "!") {
		negated = true
		line = line[1:]
	}
	if containsParentPathComponent(line) {
		return ""
	}

	// Trailing / means directory-only; for glob purposes, we match dir/**
	dirOnly := false
	if strings.HasSuffix(line, "/") {
		dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Determine if rooted
	rooted := false
	if strings.HasPrefix(line, "/") {
		rooted = true
		line = strings.TrimPrefix(line, "/")
	} else if strings.Contains(line, "/") {
		// Contains / in middle → implicitly rooted relative to .gitignore dir
		rooted = true
	}

	if line == "" {
		return ""
	}

	// Build the glob pattern
	var glob string
	if rooted {
		if baseDir != "" {
			glob = baseDir + "/" + line
		} else {
			glob = line
		}
	} else {
		// Unrooted: matches at any depth
		if baseDir != "" {
			glob = baseDir + "/**/" + line
		} else {
			glob = "**/" + line
		}
	}

	// Append /**/* for directory patterns to match all contents.
	// Use /**/* (file-level) instead of /** (directory-level) because gitignore
	// directories must stay `!`-reversible. The config matcher treats
	// `dir/**/*` as a file-level cover while `dir/**` is an absolute directory
	// block, so `.gitignore` directory patterns require the former.
	if dirOnly {
		if !strings.HasSuffix(glob, "/**/*") {
			glob = glob + "/**/*"
		}
	}

	if negated {
		glob = "!" + glob
	}
	return glob
}

func containsParentPathComponent(pattern string) bool {
	for _, component := range strings.Split(strings.ReplaceAll(pattern, "\\", "/"), "/") {
		if component == ".." {
			return true
		}
	}
	return false
}
