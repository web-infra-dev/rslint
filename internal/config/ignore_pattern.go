package config

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// GlobalIgnoreMatcher owns the authored path-space and global-ignore policy
// shared by config-candidate and lint-target discovery. Callers supply lexical
// paths plus an optional canonical fallback; matcher internals stay private so
// both discovery flows cannot accidentally diverge on ignore semantics.
type GlobalIgnoreMatcher struct {
	configDir string
	fs        vfs.FS
	patterns  []IgnorePattern
}

func NewGlobalIgnoreMatcher(config RslintConfig, configDir string, fsys vfs.FS) GlobalIgnoreMatcher {
	return GlobalIgnoreMatcher{
		configDir: tspath.NormalizePath(configDir),
		fs:        fsys,
		patterns:  extractConfigIgnores(config),
	}
}

// BlocksDirectory reports whether global ignores form an absolute traversal
// boundary at directory.
func (matcher GlobalIgnoreMatcher) BlocksDirectory(directory string, canonicalDirectory string) bool {
	relative, ok := matcher.relativePath(directory, canonicalDirectory)
	return ok && len(matcher.patterns) > 0 && isDirAbsolutelyBlocked(relative, matcher.patterns)
}

// ReopensDirectoryNode reports whether the ordered authored global-ignore
// patterns leave directory itself re-included. A pattern must match the current
// node: `!dir`, `!dir/`, and `!dir/**` reopen dir, while `!dir/**/*` and
// `!dir/file.ts` do not. Descendant patterns can still reopen a matching child.
//
// Positive authored directory blocks remain absolute under rslint's existing
// semantics and are checked by BlocksDirectory before callers consult this
// method.
func (matcher GlobalIgnoreMatcher) ReopensDirectoryNode(directory string, canonicalDirectory string) bool {
	relative, ok := matcher.relativePath(directory, canonicalDirectory)
	if !ok || len(matcher.patterns) == 0 {
		return false
	}
	relative = strings.TrimSuffix(relative, "/")
	reopened := false
	for _, pattern := range matcher.patterns {
		if ignorePatternMatches(pattern, relative) ||
			ignorePatternMatches(pattern, relative+"/") {
			reopened = pattern.Negated
		}
	}
	return reopened
}

// IgnoresPath reports whether global ignores exclude a config candidate.
func (matcher GlobalIgnoreMatcher) IgnoresPath(filePath string, canonicalPath string) bool {
	relative, ok := matcher.relativePath(filePath, canonicalPath)
	if !ok || len(matcher.patterns) == 0 {
		return false
	}
	return isDirBlockedByIgnores(relative, matcher.patterns, "") ||
		isFileIgnored(relative, matcher.patterns, "")
}

func (matcher GlobalIgnoreMatcher) relativePath(targetPath string, canonicalPath string) (string, bool) {
	if matcher.configDir == "" {
		return "", false
	}
	caseSensitive := matcher.fs == nil || matcher.fs.UseCaseSensitiveFileNames()
	relative, ok := RelativePathWithinConfigRoot(targetPath, matcher.configDir, caseSensitive)
	if !ok && canonicalPath != "" {
		matchPath, matchConfigDir := ResolveConfigPathSpaceWithCanonical(
			targetPath,
			canonicalPath,
			matcher.configDir,
			matcher.fs,
		)
		relative, ok = RelativePathWithinConfigRoot(matchPath, matchConfigDir, true)
	}
	if !ok || relative == "" {
		return "", false
	}
	return strings.ReplaceAll(tspath.NormalizePath(relative), "\\", "/"), true
}

// dirKind classifies an ignore pattern by how it bears on DIRECTORY decisions.
// It is derived once at parse time, replacing the per-call suffix-sniffing the
// pre-refactor linter did via isFileLevelPattern (and isDirPathBlocked, which
// called it) on every directory check.
type dirKind uint8

const (
	// dirNone: the pattern does not authorize any directory-level handling.
	// Pure file patterns (`*.log`) and single-level patterns (`dir/*`).
	dirNone dirKind = iota
	// dirAbsoluteBlock: an ESLint directory-level block (`dir/**`, bare names
	// like `**/build`). Blocks both the lint decision (GetConfigForFile) and
	// the walk; `!` negation can NEVER re-include inside it.
	dirAbsoluteBlock
	// dirFileLevelCover: a gitignore directory pattern in file-level form
	// (`dir/**/*`). It covers the whole subtree for WALK pruning, but `!` can
	// still re-include individual files (so pruning must be negation-aware).
	dirFileLevelCover
)

// IgnorePattern is a parsed ignore pattern that carries the structural metadata
// a raw glob string would otherwise force every consumer to re-derive:
//   - Negated: the gitignore/ESLint `!` re-include flag.
//   - Kind: the directory role (see dirKind), classified once here instead of
//     by suffix inspection at each call site.
//
// Glob is the normalized, `!`-stripped matcher fed to doublestar. For collected
// Git rules, its prefix ending at GitNodeGlobEnd is the original path-node
// matcher and the complete Glob is its conservative subtree projection. Keeping
// both views in one string avoids doubling the hot IgnorePattern's string
// headers. GitDirectoryOnly and GitContentsOnly retain the two Git-only roles.
type IgnorePattern struct {
	Glob             string
	GitNodeGlobEnd   int
	Negated          bool
	Kind             dirKind
	CaseInsensitive  bool
	GitPattern       bool
	GitDirectoryOnly bool
	GitContentsOnly  bool
}

// ParseIgnorePattern parses one raw ignore string (user config or a
// gitignore-converted glob) into the structured form. This is the SINGLE place
// that derives directory role / negation from the string; downstream code reads
// fields.
//
// PRINCIPLE: the directory role (Kind) is classified from the suffix of the RAW
// body (after stripping `!`); Glob is separately normalized as the matcher. Kind
// keys on the raw body — never the normalized Glob — because that is what the
// pre-refactor isFileLevelPattern sniffed; classifying the normalized form would
// drift for any pattern whose suffix class changes under normalization. Suffix
// → role (empty raw body → none):
//
//	`X/**/*`              → fileLevelCover (gitignore dir, prunable, `!`-aware)
//	`X/*` (not `X/**`)    → none           (single level, not subtree coverage)
//	`X/**`, bare, `*.log` → absoluteBlock   (ESLint dir-level, `!`-proof)
//
// Two normalize-sensitive cases fall out of "classify the raw suffix", both
// required for byte-equivalence with the pre-refactor linter:
//   - `./*` (Glob normalizes to bare `*`): raw ends `/*` → none. Classifying the
//     Glob would make it absoluteBlock, blocking every top-level dir while
//     isFileIgnored of `*` stays false for nested files — silently dropping them.
//   - `foo/..` (Glob normalizes to ""): raw is non-empty with no dir suffix →
//     absoluteBlock with an empty Glob. Pre-refactor isDirPathBlocked also kept
//     it absolute, its empty glob matching the empty leading segment of an
//     absolute dir path. (A literal empty / `!`-only pattern is the none case,
//     so the empty guard keys on body == "", not Glob == "".)
func ParseIgnorePattern(raw string) IgnorePattern {
	p := IgnorePattern{}
	body := raw
	if strings.HasPrefix(body, "!") {
		p.Negated = true
		body = body[1:]
	}
	p.Glob = normalizePattern(body)
	switch {
	case body == "":
		p.Kind = dirNone
	case strings.HasSuffix(body, "/**/*"):
		p.Kind = dirFileLevelCover
	case strings.HasSuffix(body, "/*") && !strings.HasSuffix(body, "/**"):
		p.Kind = dirNone
	default:
		p.Kind = dirAbsoluteBlock
	}
	return p
}

// ParseIgnorePatterns parses a list of raw ignore strings. Prefer parsing once
// per ignore set (it is fixed for a config / walk) and reusing the result — the
// directory walks (DiscoverGapFiles, the gitignore scan) hoist it out of their
// loops. The per-file GetConfigForFile path re-derives it, which is no costlier
// than the pre-refactor code (that normalized every pattern inside each matcher
// per file); centralizing the normalization here folds that work into one pass.
func ParseIgnorePatterns(raw []string) []IgnorePattern {
	if len(raw) == 0 {
		return nil
	}
	out := make([]IgnorePattern, len(raw))
	for i, s := range raw {
		out[i] = ParseIgnorePattern(s)
	}
	return out
}

// isFileIgnored evaluates patterns sequentially (later overrides earlier; `!`
// re-includes), aligned with ESLint v10. Matches only the cwd-relative path —
// never the absolute path — so `**/`-prefixed patterns can't hit system dirs.
func isFileIgnored(filePath string, patterns []IgnorePattern, cwd string) bool {
	if cwd == "" {
		return isFileIgnoredSimple(filePath, patterns)
	}
	normalizedPath := normalizePath(filePath, cwd)
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
	if pathEscapesCwd(unixPath) && hasCaseInsensitivePattern(patterns) {
		// On Windows and case-insensitive macOS volumes, callers may supply a
		// canonical path whose drive/share or directory casing differs from
		// cwd. A case-sensitive relative conversion then manufactures ../ even
		// though both paths name the same tree. Re-resolve only on that uncommon
		// fallback so the normal hot path pays no extra pattern scan.
		normalizedPath = normalizePathWithCaseSensitivity(filePath, cwd, false)
		unixPath = strings.ReplaceAll(normalizedPath, "\\", "/")
	}
	return isFileIgnoredNormalized(normalizedPath, unixPath, patterns)
}

func pathEscapesCwd(path string) bool {
	return path == ".." ||
		strings.HasPrefix(path, "../") ||
		tspath.PathIsAbsolute(path)
}

func hasCaseInsensitivePattern(patterns []IgnorePattern) bool {
	for _, pattern := range patterns {
		if pattern.CaseInsensitive {
			return true
		}
	}
	return false
}

// ignorePatternMatches reports whether path matches pattern.Glob, applying the
// case fold for case-insensitive patterns.
func ignorePatternMatches(pattern IgnorePattern, path string) bool {
	glob := pattern.Glob
	if pattern.CaseInsensitive {
		glob = strings.ToLower(glob)
		path = strings.ToLower(path)
	}
	return utils.MatchGlob(glob, path)
}

// isFileIgnoredSimple is the cwd-unavailable fallback (matches the raw path).
func isFileIgnoredSimple(filePath string, patterns []IgnorePattern) bool {
	return isFileIgnoredNormalized(filePath, strings.ReplaceAll(filePath, "\\", "/"), patterns)
}

func isFileIgnoredNormalized(path string, unixPath string, patterns []IgnorePattern) bool {
	ignored := false
	for index := 0; index < len(patterns); {
		p := patterns[index]
		if p.GitPattern {
			end := index + 1
			for end < len(patterns) && patterns[end].GitPattern {
				end++
			}
			gitState := newGitIgnorePathState(unixPath)
			groupIgnored := gitState.evaluate(patterns[index:end])
			if gitState.matched {
				ignored = groupIgnored
			}
			index = end
			continue
		}
		matched := ignorePatternMatches(p, path)
		if !matched && unixPath != path {
			matched = ignorePatternMatches(p, unixPath)
		}
		if matched {
			ignored = !p.Negated
		}
		index++
	}
	return ignored
}

// gitIgnorePathState evaluates a contiguous group of collected Git rules
// against every concrete node on one target path. Rules are visited in reverse:
// the first match for a node is its final Git decision. A negation resolves only
// the node it actually matches; an unresolved parent can therefore still be
// ignored by an earlier positive rule. This is the Git distinction erased by
// the historical !dir/**/* projection.
//
// Authored config patterns are evaluated after the complete Git group and may
// still intentionally override its final decision as a separate config layer.
type gitIgnorePathState struct {
	path                   string
	lowerPath              string
	fileDepth              int
	unresolved             int
	resolvedDepths         uint64
	resolvedDepthsOverflow []bool
	matched                bool
}

func newGitIgnorePathState(path string) gitIgnorePathState {
	path = strings.Trim(path, "/")
	if path == "." {
		path = ""
	} else if strings.HasPrefix(path, "./") ||
		strings.Contains(path, "//") ||
		strings.Contains(path, "/./") ||
		strings.HasSuffix(path, "/.") {
		// The normal path enters through normalizePath and never needs this
		// branch. Keep the cwd-less fallback equivalent to the former
		// component builder for redundant separators and "." components.
		parts := strings.Split(path, "/")
		filtered := parts[:0]
		for _, part := range parts {
			if part != "" && part != "." {
				filtered = append(filtered, part)
			}
		}
		path = strings.Join(filtered, "/")
	}
	depth := -1
	if path != "" {
		depth = strings.Count(path, "/")
	}
	return gitIgnorePathState{
		path:       path,
		fileDepth:  depth,
		unresolved: depth + 1,
	}
}

func (state *gitIgnorePathState) evaluate(patterns []IgnorePattern) bool {
	for index := len(patterns) - 1; index >= 0 && state.unresolved > 0; index-- {
		if state.apply(patterns[index]) {
			return true
		}
	}
	return false
}

// apply reports whether pattern is the final positive decision for any still
// unresolved path node. That is enough to ignore the target immediately:
// reverse traversal guarantees no earlier rule can override that node.
func (state *gitIgnorePathState) apply(pattern IgnorePattern) bool {
	if state.fileDepth < 0 {
		return false
	}
	path := state.path
	if pattern.CaseInsensitive {
		if state.lowerPath == "" {
			state.lowerPath = strings.ToLower(path)
		}
		path = state.lowerPath
	}

	if basenameGlob, ok := gitIgnoreRootBasenameGlob(pattern); ok {
		return state.applyRootBasename(pattern, path, basenameGlob)
	}

	// A non-directory-only rule may name the target file itself. Its subtree
	// projection separately tells us whether any ancestor can match.
	if !pattern.GitDirectoryOnly &&
		!state.depthResolved(state.fileDepth) &&
		gitIgnorePatternMatchesNode(pattern, path) {
		if state.applyDepth(state.fileDepth, pattern.Negated) {
			return true
		}
		if state.unresolved == 0 {
			return false
		}
	}

	// Glob is the node pattern's conservative subtree projection. One match
	// against the complete target cheaply rejects the overwhelmingly common
	// case where this rule cannot name any ancestor.
	if !utils.MatchGlob(pattern.Glob, path) {
		return false
	}

	end := strings.LastIndexByte(path, '/')
	for depth := state.fileDepth - 1; depth >= 0 && end >= 0; depth-- {
		node := path[:end]
		if !state.depthResolved(depth) &&
			gitIgnorePatternMatchesNode(pattern, node) {
			if state.applyDepth(depth, pattern.Negated) {
				return true
			}
			if state.unresolved == 0 {
				return false
			}
		}
		end = strings.LastIndexByte(node, '/')
	}
	return false
}

// gitIgnoreRootBasenameGlob recognizes the dominant .gitignore shape:
// an unrooted, single-component rule from the config root. Matching its small
// basename glob against path components avoids repeatedly running a **/ glob
// over growing full-path prefixes. A rooted "/**/name" rule has the same
// config-root semantics and is safe to take through this path too.
func gitIgnoreRootBasenameGlob(pattern IgnorePattern) (string, bool) {
	const prefix = "**/"
	nodeGlob := gitIgnoreNodeGlob(pattern)
	if !strings.HasPrefix(nodeGlob, prefix) ||
		strings.HasSuffix(nodeGlob, "/**") {
		return "", false
	}
	glob := strings.TrimPrefix(nodeGlob, prefix)
	return glob, glob != "" && !strings.Contains(glob, "/")
}

func (state *gitIgnorePathState) applyRootBasename(pattern IgnorePattern, path string, glob string) bool {
	var end int
	depth := state.fileDepth
	if pattern.GitDirectoryOnly {
		end = strings.LastIndexByte(path, '/')
		depth--
	} else {
		start := strings.LastIndexByte(path, '/') + 1
		if !state.depthResolved(depth) &&
			utils.MatchGlob(glob, path[start:]) {
			if state.applyDepth(depth, pattern.Negated) {
				return true
			}
			if state.unresolved == 0 {
				return false
			}
		}
		end = start - 1
		depth--
	}
	if end < 0 {
		return false
	}

	// A literal component is common enough to avoid both doublestar and glob
	// matching entirely. Wildcard components first use the subtree projection
	// to reject non-matches, then inspect only the candidate's ancestors.
	literal := !strings.ContainsAny(glob, "*?[{")
	if !literal && !utils.MatchGlob(pattern.Glob, path) {
		return false
	}
	for depth >= 0 && state.unresolved > 0 {
		start := strings.LastIndexByte(path[:end], '/') + 1
		component := path[start:end]
		if !state.depthResolved(depth) &&
			((literal && component == glob) || (!literal && utils.MatchGlob(glob, component))) {
			if state.applyDepth(depth, pattern.Negated) {
				return true
			}
		}
		if start == 0 {
			break
		}
		end = start - 1
		depth--
	}
	return false
}

func (state *gitIgnorePathState) applyDepth(depth int, negated bool) bool {
	state.matched = true
	if !negated {
		return true
	}
	state.resolveDepth(depth)
	return false
}

func (state *gitIgnorePathState) depthResolved(depth int) bool {
	if depth < 64 {
		return state.resolvedDepths&(uint64(1)<<uint(depth)) != 0
	}
	index := depth - 64
	return index < len(state.resolvedDepthsOverflow) &&
		state.resolvedDepthsOverflow[index]
}

func (state *gitIgnorePathState) resolveDepth(depth int) {
	if state.depthResolved(depth) {
		return
	}
	state.unresolved--
	if depth < 64 {
		state.resolvedDepths |= uint64(1) << uint(depth)
		return
	}
	index := depth - 64
	if index >= len(state.resolvedDepthsOverflow) {
		state.resolvedDepthsOverflow = append(
			state.resolvedDepthsOverflow,
			make([]bool, index-len(state.resolvedDepthsOverflow)+1)...,
		)
	}
	state.resolvedDepthsOverflow[index] = true
}

func gitIgnorePatternMatchesNode(pattern IgnorePattern, path string) bool {
	if pattern.GitContentsOnly {
		// Glob is NodeGlob with a required final component (/**/*), which
		// expresses Git's true trailing-/** semantics without ambiguity when
		// an earlier doublestar can match the same complete path.
		return utils.MatchGlob(pattern.Glob, path)
	}
	return utils.MatchGlob(gitIgnoreNodeGlob(pattern), path)
}

func gitIgnoreNodeGlob(pattern IgnorePattern) string {
	if pattern.GitNodeGlobEnd <= 0 || pattern.GitNodeGlobEnd > len(pattern.Glob) {
		return ""
	}
	return pattern.Glob[:pattern.GitNodeGlobEnd]
}

// isDirAbsolutelyBlocked reports whether dirPath (or an ancestor segment) is
// matched by a positive directory-level pattern (`dir/**`, bare names). This is
// the ESLint "directory blocking is absolute and cannot be negated" semantics —
// `!` and file-level patterns are excluded by Kind. Shared by GetConfigForFile
// (lint) and canPruneDir (walk).
func isDirAbsolutelyBlocked(dirPath string, patterns []IgnorePattern) bool {
	for i := range patterns {
		p := patterns[i]
		if p.Negated || p.Kind != dirAbsoluteBlock {
			continue
		}
		if ignorePatternMatches(p, dirPath) || ignorePatternMatches(p, dirPath+"/x") {
			return true
		}
		segments := strings.Split(dirPath, "/")
		for j := 1; j < len(segments); j++ {
			partial := strings.Join(segments[:j], "/")
			if ignorePatternMatches(p, partial) || ignorePatternMatches(p, partial+"/x") {
				return true
			}
		}
	}
	return false
}

// canPruneDir reports whether a directory walk may skip dirPath entirely. It is
// the single, negation-aware directory-prune predicate for the gap-file walk.
// The pre-refactor walk pruned only absolutely-blocked directories
// (isDirPathBlocked); canPruneDir keeps that and ADDS pruning of gitignore
// file-level covers (dir/**/*). Sound: prunes only when every descendant file
// would be rejected by the linter's own ignore decision (directory block OR
// isFileIgnored), so the gap-file set is unchanged.
//
//	absolute block (dir/**) → prune, ignoring `!` (ESLint absolute semantics)
//	file-level cover (dir/**/*) → prune ONLY if no `!` reaches into the subtree
func canPruneDir(dirPath string, patterns []IgnorePattern, neg negReach) bool {
	if isDirAbsolutelyBlocked(dirPath, patterns) {
		return true
	}
	if neg.overlaps(dirPath) {
		return false
	}
	for i := range patterns {
		p := patterns[i]
		if p.Negated || p.Kind != dirFileLevelCover {
			continue
		}
		// File-level `X/**/*` never matches the bare directory; probe `dir/x`.
		if ignorePatternMatches(p, dirPath+"/x") {
			return true
		}
	}
	return false
}

// negReach describes how far the `!` negations in an ignore set can re-include,
// so canPruneDir never prunes a directory whose subtree a `!` could re-include.
// Built once per ignore set via buildNegReach.
type negReach struct {
	prefixes []negPrefix
}

// negPrefix is one negation's reach: unrooted means "any depth" (conservatively
// blocks all file-level pruning); literal is the leading wildcard-free path of a
// rooted negation, used for subtree-overlap checks. caseInsensitive keeps that
// prefix anchored while applying the same case fold as final gitignore matching.
type negPrefix struct {
	unrooted        bool
	literal         string
	caseInsensitive bool
}

// buildNegReach extracts negation reaches from a parsed ignore set. Patterns
// already carry Negated + a normalized Glob, so no string re-parsing.
func buildNegReach(patterns []IgnorePattern) negReach {
	var out []negPrefix
	for i := range patterns {
		p := patterns[i]
		if !p.Negated {
			continue
		}
		// A negation with no concrete leading segment (`!**/keep`, `!*.log`,
		// empty) can re-include at any depth — literalSegmentPrefix returns ""
		// for those, and we mark them unrooted to conservatively disable
		// file-level pruning.
		if lit := literalSegmentPrefix(p.Glob); lit != "" {
			out = append(out, negPrefix{
				literal:         lit,
				caseInsensitive: p.CaseInsensitive,
			})
		} else {
			out = append(out, negPrefix{unrooted: true})
		}
	}
	return negReach{prefixes: out}
}

// overlaps reports whether any negation could re-include something in dirPath's
// subtree (so the directory must not be pruned).
func (n negReach) overlaps(dirPath string) bool {
	for _, np := range n.prefixes {
		literal := np.literal
		candidate := dirPath
		if np.caseInsensitive {
			literal = strings.ToLower(literal)
			candidate = strings.ToLower(candidate)
		}
		if np.unrooted || segPrefixEither(literal, candidate) {
			return true
		}
	}
	return false
}

// literalSegmentPrefix returns the leading path segments of pattern before the
// first glob metacharacter (`*`, `?`, `[`, `{`). For example `tests/e2e/**/*`
// → `tests/e2e`, `a/b*c/d` → `a`. Returns "" when the first segment already
// contains a metacharacter (the pattern is not anchored to a concrete path).
func literalSegmentPrefix(pattern string) string {
	i := strings.IndexAny(pattern, "*?[{")
	if i < 0 {
		return strings.TrimSuffix(pattern, "/")
	}
	prefix := pattern[:i]
	if idx := strings.LastIndex(prefix, "/"); idx >= 0 {
		return prefix[:idx]
	}
	return ""
}

// segPrefixEither reports whether a is a path-segment prefix of b or vice versa
// (e.g. "target" vs "target/keep", or "tests/e2e" vs "tests") — the negation
// target sits inside the directory, or the directory sits inside the negation's
// covered range.
func segPrefixEither(a, b string) bool {
	return segPrefix(a, b) || segPrefix(b, a)
}

// segPrefix reports whether prefix equals path or path lies under prefix/
// (segment-anchored, so "a/b" is a prefix of "a/b/c" but not of "a/bc").
func segPrefix(prefix, path string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
