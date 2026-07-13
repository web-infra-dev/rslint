package config

import (
	"strings"
)

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
// Glob is the normalized, `!`-stripped matcher fed to doublestar — byte-for-byte
// the same string the old []string pipeline matched against, so file-level
// matching (isFileIgnored) is unchanged.
type IgnorePattern struct {
	Glob            string
	Negated         bool
	Kind            dirKind
	CaseInsensitive bool
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

	ignored := false
	for _, p := range patterns {
		matched := ignorePatternMatches(p, normalizedPath)
		if !matched && unixPath != normalizedPath {
			matched = ignorePatternMatches(p, unixPath)
		}
		if matched {
			ignored = !p.Negated
		}
	}
	return ignored
}

func ignorePatternMatches(pattern IgnorePattern, path string) bool {
	glob := pattern.Glob
	if pattern.CaseInsensitive {
		glob = strings.ToLower(glob)
		path = strings.ToLower(path)
	}
	return matchGlob(glob, path)
}

// isFileIgnoredSimple is the cwd-unavailable fallback (matches the raw path).
func isFileIgnoredSimple(filePath string, patterns []IgnorePattern) bool {
	ignored := false
	for _, p := range patterns {
		if ignorePatternMatches(p, filePath) {
			ignored = !p.Negated
		}
	}
	return ignored
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
// rooted negation, used for subtree-overlap checks.
type negPrefix struct {
	unrooted bool
	literal  string
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
		if p.CaseInsensitive {
			// Path-prefix reachability has no case mode. Conservatively disable
			// file-level pruning; final matching still uses case-folded patterns.
			out = append(out, negPrefix{unrooted: true})
			continue
		}
		// A negation with no concrete leading segment (`!**/keep`, `!*.log`,
		// empty) can re-include at any depth — literalSegmentPrefix returns ""
		// for those, and we mark them unrooted to conservatively disable
		// file-level pruning.
		if lit := literalSegmentPrefix(p.Glob); lit != "" {
			out = append(out, negPrefix{literal: lit})
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
		if np.unrooted || segPrefixEither(np.literal, dirPath) {
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
