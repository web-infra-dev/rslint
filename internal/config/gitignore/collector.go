package gitignore

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Pattern preserves both views of one collected .gitignore rule:
//
//   - Glob is the compatibility projection exposed through the synthetic
//     config entry and used to derive conservative directory pruning.
//   - NodeGlob matches the path node named by the Git rule before a trailing
//     directory slash is expanded to /**/*. Keeping this view is essential for
//     ordered negations: !dist-path/ re-includes the dist-path node, not every
//     independently ignored node below it.
//
// DirectoryOnly retains Git's trailing-slash restriction. Negated is stored
// separately because an escaped leading exclamation mark is literal pattern
// text, not a negation. ContentsOnly distinguishes a genuine trailing "/**"
// from an unrooted bare "**" whose generated NodeGlob also ends in "/**".
type Pattern struct {
	Glob          string
	NodeGlob      string
	Negated       bool
	DirectoryOnly bool
	ContentsOnly  bool
}

func normalizeGlobPath(path string) string {
	return strings.ReplaceAll(tspath.NormalizePath(path), "\\", "/")
}

func matchGitignoreGlob(pattern string, path string, useCaseSensitive bool) bool {
	if !useCaseSensitive {
		pattern = strings.ToLower(pattern)
		path = strings.ToLower(path)
	}
	return utils.MatchGlob(pattern, path)
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
	return patternGlobs(CollectPatternsWithBoundaries(configDir, fsys, targetFiles, isDirectoryBlocked, stopDirs))
}

// CollectPatterns is the structured form of Collect. Callers that feed the
// final ignore matcher must use this form so Git directory-node semantics are
// not lost in the compatibility glob projection.
func CollectPatterns(configDir string, fsys vfs.FS, targetFiles []string, isDirectoryBlocked func(string) bool) []Pattern {
	return CollectPatternsWithBoundaries(configDir, fsys, targetFiles, isDirectoryBlocked, nil)
}

// CollectPatternsWithBoundaries is the structured form of
// CollectWithBoundaries.
func CollectPatternsWithBoundaries(configDir string, fsys vfs.FS, targetFiles []string, isDirectoryBlocked func(string) bool, stopDirs []string) []Pattern {
	if targetFiles == nil {
		return readGitignoreAsPatternsWithBoundaries(configDir, fsys, isDirectoryBlocked, stopDirs)
	}
	return readGitignoreAsPatternsForFilesWithBoundaries(configDir, fsys, targetFiles, stopDirs)
}

func patternGlobs(patterns []Pattern) []string {
	if len(patterns) == 0 {
		return nil
	}
	globs := make([]string, len(patterns))
	for index, pattern := range patterns {
		globs[index] = pattern.Glob
	}
	return globs
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
	return patternGlobs(readGitignoreAsPatternsWithBoundaries(configDir, fsys, isDirectoryBlocked, stopDirs))
}

func readGitignoreAsPatternsWithBoundaries(configDir string, fsys vfs.FS, isDirectoryBlocked func(string) bool, stopDirs []string) []Pattern {
	if fsys == nil {
		return nil
	}
	normalizedRoot := normalizeGlobPath(configDir)
	boundaries := normalizeCollectionBoundaries(normalizedRoot, stopDirs, fsys.UseCaseSensitiveFileNames())
	var allPatterns []Pattern

	collectGitignorePatterns(normalizedRoot, "", fsys, &allPatterns, isDirectoryBlocked, nil, boundaries)

	if len(allPatterns) == 0 {
		return nil
	}
	return allPatterns
}

// readGitignoreAsGlobsForFilesWithBoundaries reads only .gitignore files on
// each directory chain from configDir to an explicit target. This is used by
// API-style calls where the target set is already known; unlike the full
// collector, it does not scan every descendant of configDir.
func readGitignoreAsGlobsForFilesWithBoundaries(configDir string, fsys vfs.FS, files []string, stopDirs []string) []string {
	return patternGlobs(readGitignoreAsPatternsForFilesWithBoundaries(configDir, fsys, files, stopDirs))
}

func readGitignoreAsPatternsForFilesWithBoundaries(configDir string, fsys vfs.FS, files []string, stopDirs []string) []Pattern {
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

	var allPatterns []Pattern
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
		rel, _ := relativeDir(normalizedConfigDir, dir, useCaseSensitive)
		if isDirIgnoredByPruneRules(rel, pruneRules, useCaseSensitive) {
			prunedDirs = append(prunedDirs, dir)
			continue
		}

		content, ok := fsys.ReadFile(tspath.CombinePaths(dir, ".gitignore"))
		if !ok {
			continue
		}
		localPatterns := convertGitignoreToPatterns(content, rel)
		allPatterns = append(allPatterns, localPatterns...)
		if len(localPatterns) > 0 {
			pruneRules = append(pruneRules, gitignorePruneRule{patterns: localPatterns})
		}
	}
	if len(allPatterns) == 0 {
		return nil
	}
	return allPatterns
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
	patterns []Pattern
}

// Cursor is an immutable, filesystem-independent view of the .gitignore
// sources on one config owner's directory chain. It deliberately
// separates two questions:
//
//   - Enter reports whether the already-observed Git patterns ignore one
//     directory node. Config discovery may still reopen that node with a later
//     authored global-ignore negation.
//   - SourceReachable reports whether raw Git traversal can read the current
//     directory's .gitignore. Authored config never changes this state; only a
//     successfully loaded child config creates a new cursor/root.
//
// Values are safe to copy between discovery workers. Enter only changes scalar
// state; AppendSource clones the rule slice before extending it, so siblings
// never share a writable backing array.
type Cursor struct {
	rootDir          string
	useCaseSensitive bool
	sourceReachable  bool
	rules            []gitignorePruneRule
}

// NewCursor starts a config-scoped Git view. The owner directory itself is
// always source-reachable; parent .gitignore files are intentionally excluded.
func NewCursor(ownerRoot string, useCaseSensitive bool) Cursor {
	ownerRoot = normalizeGlobPath(ownerRoot)
	return Cursor{
		rootDir:          ownerRoot,
		useCaseSensitive: useCaseSensitive,
		sourceReachable:  ownerRoot != "",
	}
}

// SourceReachable reports whether raw Git traversal permits reading the
// current directory's .gitignore.
func (cursor Cursor) SourceReachable() bool {
	return cursor.sourceReachable
}

// BlockSourceTraversal preserves already-observed matching rules while making
// .gitignore sources at the current directory and its descendants unreachable.
// A successfully loaded child config starts a new cursor and therefore a new
// source boundary.
func (cursor Cursor) BlockSourceTraversal() Cursor {
	cursor.sourceReachable = false
	return cursor
}

// Enter advances to directory and reports whether the already-observed raw Git
// patterns ignore that exact directory node. The returned cursor keeps raw
// source reachability monotonic: a negation below an ignored parent cannot make
// nested .gitignore sources visible.
func (cursor Cursor) Enter(directory string) (Cursor, bool) {
	directory = normalizeGlobPath(directory)
	if _, ok := relativeDir(cursor.rootDir, directory, cursor.useCaseSensitive); !ok {
		next := cursor
		next.sourceReachable = false
		return next, true
	}
	blocked := cursor.blocksDirectoryNode(directory)
	next := cursor
	next.sourceReachable = cursor.sourceReachable && !blocked
	return next, blocked
}

// AppendSource extends the cursor with directory's .gitignore and returns the
// source projected into the config owner's target path space. Callers keep the
// source IO path separate and pass the matching-space directory explicitly so
// lexical and canonical discovery routes cannot be mixed accidentally.
func (cursor Cursor) AppendSource(directory string, content string) (Cursor, []string) {
	next, patterns := cursor.AppendSourcePatterns(directory, content)
	return next, patternGlobs(patterns)
}

// AppendSourcePatterns is the structured form of AppendSource. It keeps the
// directory-node matcher needed by final file decisions while AppendSource
// remains available to compatibility callers that only inspect projected
// globs.
func (cursor Cursor) AppendSourcePatterns(directory string, content string) (Cursor, []Pattern) {
	if !cursor.sourceReachable {
		return cursor, nil
	}
	directory = normalizeGlobPath(directory)
	relative, ok := relativeDir(cursor.rootDir, directory, cursor.useCaseSensitive)
	if !ok {
		next := cursor
		next.sourceReachable = false
		return next, nil
	}

	next := cursor
	patterns := convertGitignoreToPatterns(content, relative)
	if len(patterns) > 0 {
		next.rules = append(
			append([]gitignorePruneRule(nil), cursor.rules...),
			gitignorePruneRule{patterns: patterns},
		)
	}
	return next, patterns
}

func (cursor Cursor) blocksDirectoryNode(directory string) bool {
	if cursor.rootDir == "" || directory == "" {
		return false
	}
	relative, ok := relativeDir(cursor.rootDir, directory, cursor.useCaseSensitive)
	if !ok || relative == "" {
		return false
	}
	ignored := false
	for _, rule := range cursor.rules {
		for _, pattern := range rule.patterns {
			if gitignorePrunePatternMatchesDirectoryNode(pattern, relative, cursor.useCaseSensitive) {
				ignored = !pattern.Negated
			}
		}
	}
	return ignored
}

// gitignorePrunePatternMatchesDirectoryNode matches only the current directory
// node. Parent reachability is already enforced by the recursive walk, the
// explicit-chain prunedDirs set, or Cursor.sourceReachable. Re-evaluating an
// ancestor here would incorrectly let !parent/ re-include an independently
// ignored child directory.
func gitignorePrunePatternMatchesDirectoryNode(pattern Pattern, relative string, useCaseSensitive bool) bool {
	glob := pattern.NodeGlob
	if pattern.ContentsOnly {
		// Git's trailing /** matches only contents. Requiring one final
		// component keeps the prefix directory itself reachable.
		glob += "/*"
	}
	return matchGitignoreGlob(glob, relative, useCaseSensitive)
}

func isDirIgnoredByPruneRules(relative string, rules []gitignorePruneRule, useCaseSensitive bool) bool {
	if relative == "" {
		return false
	}
	ignored := false
	for _, rule := range rules {
		for _, pattern := range rule.patterns {
			if gitignorePrunePatternMatchesDirectoryNode(pattern, relative, useCaseSensitive) {
				ignored = !pattern.Negated
			}
		}
	}
	return ignored
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

// collectGitignorePatterns recursively scans for .gitignore files and converts
// their rules to structured patterns. Raw patterns from parent .gitignore
// files are used to prune directories during scanning.
//
// isDirectoryBlocked provides directory-level pruning from the caller's
// config policy. If it returns true, files below that directory cannot be
// linted, so nested .gitignore sources there are irrelevant.
func collectGitignorePatterns(absDir string, relDir string, fsys vfs.FS, result *[]Pattern, isDirectoryBlocked func(string) bool, pruneRules []gitignorePruneRule, boundaries []string) {
	collectGitignorePatternsRecursive(absDir, relDir, fsys, result, isDirectoryBlocked, pruneRules, boundaries, &gitignoreWalkState{
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

func collectGitignorePatternsRecursive(absDir string, relDir string, fsys vfs.FS, result *[]Pattern, isDirectoryBlocked func(string) bool, pruneRules []gitignorePruneRule, boundaries []string, state *gitignoreWalkState) {
	gitignorePath := tspath.CombinePaths(absDir, ".gitignore")
	if content, ok := fsys.ReadFile(gitignorePath); ok {
		localPatterns := convertGitignoreToPatterns(content, relDir)
		*result = append(*result, localPatterns...)
		if len(localPatterns) > 0 {
			pruneRules = append(pruneRules, gitignorePruneRule{patterns: localPatterns})
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
		if isDirIgnoredByPruneRules(childRel, pruneRules, fsys.UseCaseSensitiveFileNames()) {
			continue
		}

		if childRealPath != "" {
			if _, visited := state.visited[childRealPath]; visited {
				continue
			}
			state.visited[childRealPath] = struct{}{}
		}

		collectGitignorePatternsRecursive(childAbs, childRel, fsys, result, isDirectoryBlocked, pruneRules, boundaries, state)
	}
}

// convertGitignoreToGlobs converts .gitignore file content into rslint glob
// patterns. baseDir is the directory containing the .gitignore relative to
// the config root (empty for root .gitignore).
//
// The compatibility glob view uses these projections:
//
//	"dist/"     → "**/dist/**/*"      (unrooted dir, matches at any depth)
//	"/dist"     → "dist"              (root-anchored)
//	"*.log"     → "**/*.log"          (unrooted file pattern)
//	"src/dist"  → "src/dist"          (contains /, implicitly rooted)
//	"!dist/"    → "!**/dist/**/*"     (negation preserved)
//
// For nested .gitignore (baseDir != ""), patterns are prefixed:
//
//	baseDir="pkg/app", "tmp/" → "pkg/app/**/tmp/**/*"
func convertGitignoreToGlobs(content string, baseDir string) []string {
	return patternGlobs(convertGitignoreToPatterns(content, baseDir))
}

func convertGitignoreToPatterns(content string, baseDir string) []Pattern {
	var patterns []Pattern
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		pattern, ok := convertSinglePatternToPattern(line, baseDir)
		if ok {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

func normalizeGitignoreLine(line string) (string, bool) {
	// A CR from CRLF is a line terminator, not pattern text. Git otherwise
	// preserves leading whitespace and removes only unescaped trailing
	// whitespace. In particular, ` leading` is a real filename pattern and
	// `trailing\ ` retains its final space.
	line = strings.TrimSuffix(line, "\r")
	line = trimUnescapedTrailingGitignoreWhitespace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	return line, true
}

func trimUnescapedTrailingGitignoreWhitespace(line string) string {
	for len(line) > 0 {
		last := line[len(line)-1]
		if last != ' ' && last != '\t' {
			break
		}
		backslashes := 0
		for index := len(line) - 2; index >= 0 && line[index] == '\\'; index-- {
			backslashes++
		}
		if backslashes%2 == 1 {
			break
		}
		line = line[:len(line)-1]
	}
	return line
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
	pattern, ok := convertSinglePatternToPattern(line, baseDir)
	if !ok {
		return ""
	}
	return pattern.Glob
}

func convertSinglePatternToPattern(line string, baseDir string) (Pattern, bool) {
	negated := false
	if strings.HasPrefix(line, "!") {
		negated = true
		line = line[1:]
	}
	if containsParentPathComponent(line) {
		return Pattern{}, false
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
		return Pattern{}, false
	}
	contentsOnly := strings.HasSuffix(line, "/**")
	line = gitignorePatternForConfigGlob(line)
	baseGlob := gitignoreBaseDirForConfigGlob(baseDir)

	// Build the glob pattern
	var glob string
	if rooted {
		if baseGlob != "" {
			glob = baseGlob + "/" + line
		} else {
			glob = line
		}
	} else {
		// Unrooted: matches at any depth
		if baseGlob != "" {
			glob = baseGlob + "/**/" + line
		} else {
			glob = "**/" + line
		}
	}

	nodeGlob := glob

	// Append /**/* for directory patterns to match all contents.
	// Use /**/* (file-level) instead of /** (directory-level) because collected
	// rules participate in one ordered global-ignore sequence and later
	// negations must be able to re-include descendants.
	if dirOnly && !strings.HasSuffix(glob, "/**/*") {
		if contentsOnly {
			// A genuine trailing /** already supplies the recursive part; add
			// one required component so the directory before it is not matched.
			glob += "/*"
		} else {
			glob += "/**/*"
		}
	}

	if negated {
		glob = "!" + glob
	} else if strings.HasPrefix(glob, "!") {
		// A leading `\!` is a literal exclamation mark. Keep the raw config glob
		// from looking like an rslint negation; NormalizePath later removes `./`
		// without changing the matcher.
		glob = "./" + glob
	}
	return Pattern{
		Glob:          glob,
		NodeGlob:      nodeGlob,
		Negated:       negated,
		DirectoryOnly: dirOnly,
		ContentsOnly:  contentsOnly,
	}, true
}

// gitignoreBaseDirForConfigGlob quotes the real directory names contributed by
// the filesystem. Unlike the .gitignore rule itself, this prefix is never glob
// syntax: a repository directory literally named pkg[1] or pkg{a} must remain
// literal on Unix, Windows, and case-insensitive macOS volumes.
func gitignoreBaseDirForConfigGlob(baseDir string) string {
	if !strings.ContainsAny(baseDir, "*?[{}") {
		return baseDir
	}
	var result strings.Builder
	result.Grow(len(baseDir))
	for index := range len(baseDir) {
		character := baseDir[index]
		if character == '/' {
			result.WriteByte(character)
		} else {
			appendConfigGlobLiteral(&result, character)
		}
	}
	return result.String()
}

// gitignorePatternForConfigGlob removes Git's backslash quoting before a
// pattern enters rslint's generic glob pipeline. That pipeline normalizes
// backslashes as path separators, so escaped glob metacharacters must instead
// be represented as one-character classes. Doublestar supports brace
// alternation while Git does not, therefore unescaped braces are quoted too.
func gitignorePatternForConfigGlob(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern))
	inClass := false
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		if character == '\\' && index+1 < len(pattern) {
			index++
			appendConfigGlobLiteral(&result, pattern[index])
			continue
		}
		if character == '[' {
			inClass = true
			result.WriteByte(character)
			continue
		}
		if character == ']' && inClass {
			inClass = false
			result.WriteByte(character)
			continue
		}
		if !inClass && (character == '{' || character == '}') {
			appendConfigGlobLiteral(&result, character)
			continue
		}
		result.WriteByte(character)
	}
	return result.String()
}

func appendConfigGlobLiteral(result *strings.Builder, character byte) {
	switch character {
	case '*':
		result.WriteString("[*]")
	case '?':
		result.WriteString("[?]")
	case '[':
		result.WriteString("[[]")
	case '{':
		result.WriteString("[{]")
	case '}':
		result.WriteString("[}]")
	default:
		result.WriteByte(character)
	}
}

func containsParentPathComponent(pattern string) bool {
	// Git patterns always use `/` as the separator; backslash quotes the next
	// byte even on Windows and must not manufacture a parent component.
	for _, component := range strings.Split(pattern, "/") {
		if component == ".." {
			return true
		}
	}
	return false
}
