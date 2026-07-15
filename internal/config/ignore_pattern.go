package config

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/minimatch"
)

// GlobalIgnoreMatcher owns the authored path-space and global-ignore policy
// shared by config-candidate and lint-target discovery. Callers supply lexical
// paths plus an optional canonical fallback; matcher internals stay private so
// both discovery flows cannot accidentally diverge on ignore semantics.
type GlobalIgnoreMatcher struct {
	fs     vfs.FS
	layers []configIgnoreLayer
}

// configIgnoreLayer preserves the ConfigArray order and path root of one
// global-ignore entry. basePath is resolved once from the config directory;
// patterns never need to be reparsed while walking or resolving files.
type configIgnoreLayer struct {
	basePath          string
	canonicalBasePath string
	patterns          []IgnorePattern
	matchers          []compiledIgnoreMatcher
}

type compiledIgnoreMatcher struct {
	pattern     *IgnorePattern
	predicateID string
}

// WithDefaultGlobalIgnores installs the product-level default ignore entry in
// the same ordered config array as authored ignores. Later negations can reopen
// it exactly like ESLint's defaultConfig, including root-only .git/ and
// any-depth node_modules/ behavior.
func WithDefaultGlobalIgnores(config RslintConfig) RslintConfig {
	for _, entry := range config {
		if entry.defaultIgnores {
			return config
		}
	}
	effective := make(RslintConfig, 0, len(config)+1)
	effective = append(effective, ConfigEntry{
		Name:           "rslint/default-ignores",
		Ignores:        []string{"**/node_modules/", ".git/"},
		defaultIgnores: true,
	})
	effective = append(effective, config...)
	return effective
}

func NewGlobalIgnoreMatcher(config RslintConfig, configDir string, fsys vfs.FS) GlobalIgnoreMatcher {
	return GlobalIgnoreMatcher{
		fs:     fsys,
		layers: compileConfigIgnoreLayers(config, configDir, fsys),
	}
}

// BlocksDirectory reports whether directory or one of its ancestors is ignored
// by the ordered global-ignore view.
func (matcher GlobalIgnoreMatcher) BlocksDirectory(directory string, canonicalDirectory string) bool {
	return isDirectoryIgnoredByConfigLayers(
		directory,
		canonicalDirectory,
		matcher.layers,
		matcher.fs,
	)
}

// IgnoresPath reports whether global ignores exclude a config candidate.
func (matcher GlobalIgnoreMatcher) IgnoresPath(filePath string, canonicalPath string) bool {
	return isFileIgnoredByConfigLayers(
		filePath,
		canonicalPath,
		matcher.layers,
		matcher.fs,
	)
}

func compileConfigIgnoreLayers(config RslintConfig, configDir string, fsys vfs.FS) []configIgnoreLayer {
	layers := make([]configIgnoreLayer, 0, len(config))
	for _, entry := range config {
		if !isGlobalIgnoreEntry(entry) {
			continue
		}
		patterns := ParseIgnorePatterns(entry.Ignores)
		ordered := make([]compiledIgnoreMatcher, 0, len(ignoreMatchers(entry)))
		if entry.gitignoreSemantics {
			patterns = parseCollectedGitignorePatterns(entry.Ignores, entry.gitignoreCaseInsensitive)
			for index := range patterns {
				ordered = append(ordered, compiledIgnoreMatcher{pattern: &patterns[index]})
			}
		} else {
			patterns = patterns[:0]
			for _, matcher := range ignoreMatchers(entry) {
				if matcher.isPredicate() {
					ordered = append(ordered, compiledIgnoreMatcher{predicateID: matcher.predicateID})
					continue
				}
				pattern := ParseIgnorePattern(matcher.pattern)
				patterns = append(patterns, pattern)
				ordered = append(ordered, compiledIgnoreMatcher{pattern: &patterns[len(patterns)-1]})
			}
		}
		if len(ordered) == 0 {
			continue
		}
		basePath := resolveConfigEntryBasePath(configDir, entry.BasePath)
		canonicalBasePath := ""
		if fsys != nil && basePath != "" {
			if realPath := fsys.Realpath(basePath); realPath != "" {
				canonicalBasePath = NormalizeHostPath(realPath)
			}
		}
		layers = append(layers, configIgnoreLayer{
			basePath:          basePath,
			canonicalBasePath: canonicalBasePath,
			patterns:          patterns,
			matchers:          ordered,
		})
	}
	return layers
}

func resolveConfigEntryBasePath(configDir string, entryBasePath string) string {
	if configDir != "" {
		configDir = normalizePathForRoot(configDir, configDir)
	}
	if entryBasePath != "" {
		entryBasePath = normalizePathForRoot(configDir, entryBasePath)
	}
	if entryBasePath == "." {
		entryBasePath = ""
	}
	if entryBasePath == "" {
		return configDir
	}
	if pathIsAbsoluteForRoot(configDir, entryBasePath) || configDir == "" {
		return entryBasePath
	}
	return normalizePathForRoot(configDir, resolvePathForRoot(configDir, configDir, entryBasePath))
}

func (layer configIgnoreLayer) relativePath(targetPath string, canonicalPath string, fsys vfs.FS) (string, bool) {
	targetPath = normalizePathForRoot(layer.basePath, targetPath)
	if layer.basePath == "" {
		if targetPath == "" || targetPath == "." {
			return "", false
		}
		return strings.ReplaceAll(targetPath, "\\", "/"), true
	}
	caseSensitive := selectorScopeCaseSensitive(layer.basePath)
	relative, ok := RelativePathWithinConfigRoot(targetPath, layer.basePath, caseSensitive)
	if !ok && layer.canonicalBasePath != "" {
		if canonicalPath == "" && fsys != nil {
			canonicalPath = fsys.Realpath(targetPath)
		}
		if canonicalPath != "" {
			relative, ok = RelativePathWithinConfigRoot(
				normalizePathForRoot(layer.canonicalBasePath, canonicalPath),
				layer.canonicalBasePath,
				selectorScopeCaseSensitive(layer.canonicalBasePath),
			)
		}
	}
	if !ok || relative == "" || relative == "." {
		// ConfigArray deliberately skips an ignore entry at its own basePath.
		return "", false
	}
	relative = normalizePathForRoot(layer.basePath, relative)
	if !pathUsesNativePOSIXSemantics(layer.basePath) {
		relative = strings.ReplaceAll(relative, "\\", "/")
	}
	return strings.TrimPrefix(relative, "./"), true
}

func isDirectoryIgnoredByConfigLayers(
	directory string,
	canonicalDirectory string,
	layers []configIgnoreLayer,
	fsys vfs.FS,
) bool {
	if len(layers) == 0 {
		return false
	}
	directory = NormalizeHostPath(directory)
	canonicalDirectory = NormalizeHostPath(canonicalDirectory)
	ancestors := pathAncestors(directory)
	canonicalAncestors := pathAncestors(canonicalDirectory)
	for index, ancestor := range ancestors {
		canonicalAncestor := ""
		if len(canonicalAncestors) == len(ancestors) {
			canonicalAncestor = canonicalAncestors[index]
		}
		if isDirectoryIgnoredDirectlyByConfigLayers(ancestor, canonicalAncestor, layers, fsys) {
			return true
		}
	}
	return false
}

func isDirectoryIgnoredDirectlyByConfigLayers(
	directory string,
	canonicalDirectory string,
	layers []configIgnoreLayer,
	fsys vfs.FS,
) bool {
	ignored := false
	for _, layer := range layers {
		relative, within := layer.relativePath(directory, canonicalDirectory, fsys)
		if !within {
			continue
		}
		for _, pattern := range layer.patterns {
			if searchDirectoryIgnorePatternMatches(pattern, relative) {
				ignored = !pattern.Negated
			}
		}
	}
	return ignored
}

func searchDirectoryIgnorePatternMatches(pattern IgnorePattern, relativeDirectory string) bool {
	// ConfigArray checks a directory using a trailing separator. The matcher
	// also checks the bare spelling to model Minimatch's optional terminal
	// empty segment exactly.
	return ignorePatternMatches(pattern, strings.TrimSuffix(relativeDirectory, "/")+"/") ||
		ignorePatternMatches(pattern, strings.TrimSuffix(relativeDirectory, "/"))
}

func isFileIgnoredByConfigLayers(
	filePath string,
	canonicalPath string,
	layers []configIgnoreLayer,
	fsys vfs.FS,
) bool {
	if len(layers) == 0 {
		return false
	}
	canonicalDirectory := ""
	if canonicalPath != "" {
		canonicalDirectory = HostDirectoryPath(canonicalPath)
	}
	if isDirectoryIgnoredByConfigLayers(
		HostDirectoryPath(filePath),
		canonicalDirectory,
		layers,
		fsys,
	) {
		return true
	}
	ignored := false
	for _, layer := range layers {
		relative, within := layer.relativePath(filePath, canonicalPath, fsys)
		if !within {
			continue
		}
		for _, pattern := range layer.patterns {
			if ignorePatternMatches(pattern, relative) {
				ignored = !pattern.Negated
			}
		}
	}
	return ignored
}

func pathAncestors(value string) []string {
	if value == "" || value == "." {
		return nil
	}
	var reversed []string
	for current := value; current != "" && current != "."; {
		reversed = append(reversed, current)
		parent := directoryPathForRoot(value, current)
		if parent == current {
			break
		}
		current = parent
	}
	ancestors := make([]string, len(reversed))
	for index := range reversed {
		ancestors[len(reversed)-1-index] = reversed[index]
	}
	return ancestors
}

// IgnorePattern is one precompiled, ordered ConfigArray ignore pattern. Glob is
// the normalized positive matcher body; Negated records are re-inclusions.
type IgnorePattern struct {
	Glob            string
	Negated         bool
	CaseInsensitive bool
	gitignore       bool
	matcher         *minimatch.SearchPattern
}

// ParseIgnorePattern parses one raw ignore string (user config or a
// gitignore-converted glob) into the structured form.
//
// ConfigArray normalizes exactly one initial "./", including the instance
// immediately after exactly one leading "!". Any number of leading exclamation
// marks then makes this an ordered re-inclusion, while the positive matcher is
// compiled after stripping all of them (Minimatch flipNegate semantics).
func ParseIgnorePattern(raw string) IgnorePattern {
	p := IgnorePattern{}
	normalized := normalizeConfigPattern(raw)
	p.Negated = strings.HasPrefix(normalized, "!")
	body := strings.TrimLeft(normalized, "!")
	p.Glob = body
	if compiled, err := minimatch.CompileRelativePattern(p.Glob); err == nil {
		p.matcher = &compiled
	}
	return p
}

func parseGitignorePattern(raw string) IgnorePattern {
	p := IgnorePattern{gitignore: true}
	p.Negated = strings.HasPrefix(raw, "!")
	body := strings.TrimPrefix(raw, "!")
	// The collector uses ./ solely to protect a rooted, escaped leading ! from
	// ConfigArray syntax. It is not a path segment in the Git pattern.
	body = strings.TrimPrefix(body, "./")
	p.Glob = body
	if compiled, err := minimatch.CompileGitignorePattern(body); err == nil {
		p.matcher = &compiled
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
	if pattern.CaseInsensitive {
		compile := minimatch.CompileRelativePattern
		if pattern.gitignore {
			compile = minimatch.CompileGitignorePattern
		}
		compiled, err := compile(strings.ToLower(pattern.Glob))
		return err == nil && compiled.Match(strings.ToLower(path))
	}
	if pattern.matcher == nil {
		return false
	}
	return pattern.matcher.Match(path)
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
