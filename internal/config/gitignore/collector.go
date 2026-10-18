package gitignore

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
	ignore "github.com/web-infra-dev/rslint/internal/utils/gitignore"
)

// Pattern retains the path-node semantics of one repository-relative Git rule.
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

type gitignorePruneRule struct {
	patterns []Pattern
}

// Cursor is an immutable, filesystem-independent view of the .gitignore
// sources on one repository directory chain. It deliberately
// separates two questions:
//
//   - Enter reports whether the already-observed Git patterns ignore one
//     directory node. Config discovery may still reopen that node with a later
//     authored global-ignore negation.
//   - SourceReachable reports whether raw Git traversal can read the current
//     directory's .gitignore. Authored config never changes this state; only a
//     new repository boundary creates a new cursor/root.
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

// NewCursor starts a repository-scoped Git view. Its root is source-reachable.
func NewCursor(repositoryRoot string, useCaseSensitive bool) Cursor {
	repositoryRoot = normalizeGlobPath(repositoryRoot)
	return Cursor{
		rootDir:          repositoryRoot,
		useCaseSensitive: useCaseSensitive,
		sourceReachable:  repositoryRoot != "",
	}
}

// SourceReachable reports whether raw Git traversal permits reading the
// current directory's .gitignore.
func (cursor Cursor) SourceReachable() bool {
	return cursor.sourceReachable
}

// RootDirectory returns the path space in which AppendSourcePatterns projects
// its globs. It is immutable for the cursor's lifetime.
func (cursor Cursor) RootDirectory() string {
	return cursor.rootDir
}

// BlockSourceTraversal preserves already-observed matching rules while making
// .gitignore sources at the current directory and its descendants unreachable.
// A new repository boundary starts a fresh cursor.
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
// source projected into the repository root's path space. Callers keep the
// source IO path separate and pass the matching-space directory explicitly so
// lexical and canonical discovery routes cannot be mixed accidentally.
func (cursor Cursor) AppendSource(directory string, content string) (Cursor, []string) {
	next, patterns := cursor.AppendSourcePatterns(directory, content)
	return next, patternGlobs(patterns)
}

// AppendSourcePatterns preserves the directory-node matcher used by both the
// raw cursor and final file decisions.
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

	patterns := convertGitignoreToPatterns(content, relative)
	return cursor.appendPatterns(patterns), patterns
}

func (cursor Cursor) appendPatterns(patterns []Pattern) Cursor {
	next := cursor
	if len(patterns) > 0 {
		next.rules = append(
			append([]gitignorePruneRule(nil), cursor.rules...),
			gitignorePruneRule{patterns: patterns},
		)
	}
	return next
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

func convertGitignoreToGlobs(content string, baseDir string) []string {
	return patternGlobs(convertGitignoreToPatterns(content, baseDir))
}

func convertGitignoreToPatterns(content string, baseDir string) []Pattern {
	var patterns []Pattern
	for pattern := range ignore.ParseText(content, baseDir) {
		patterns = append(patterns, Pattern{
			Glob: pattern.Glob, NodeGlob: pattern.NodeGlob,
			Negated: pattern.Negated, DirectoryOnly: pattern.DirectoryOnly,
			ContentsOnly: pattern.ContentsOnly,
		})
	}
	return patterns
}
