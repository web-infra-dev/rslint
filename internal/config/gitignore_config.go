package config

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// ConfigWithGitignore prepends the .gitignore patterns that apply to a lint
// invocation. A nil targetFiles slice scans the config-owned subtree, as used
// by LSP and directory-based CLI discovery. A non-nil slice limits collection
// to the directory chains between configDir and exact targets, as used by API,
// file-only CLI, and explicit-only scopes in mixed CLI requests. The input
// config is never mutated.
func ConfigWithGitignore(config RslintConfig, configDir string, fsys vfs.FS, targetFiles []string) RslintConfig {
	return ConfigWithGitignoreWithBoundaries(config, configDir, fsys, targetFiles, nil)
}

// ConfigWithGitignoreWithBoundaries applies the shared .gitignore policy while
// excluding caller-supplied descendant ownership boundaries. A boundary and
// its subtree are handed off without reading that subtree's .gitignore files.
func ConfigWithGitignoreWithBoundaries(config RslintConfig, configDir string, fsys vfs.FS, targetFiles []string, stopDirs []string) RslintConfig {
	collectionFiles := targetFiles
	var isDirectoryBlocked func(string) bool
	if targetFiles == nil {
		if hasGlobalIgnoreEntries(config) {
			matcher := NewGlobalIgnoreMatcher(config, configDir, fsys)
			isDirectoryBlocked = func(relativePath string) bool {
				return matcher.BlocksDirectory(
					tspath.ResolvePath(configDir, relativePath),
					"",
				)
			}
		}
	} else if fsys != nil && len(targetFiles) > 0 {
		collectionFiles = make([]string, len(targetFiles))
		for i, file := range targetFiles {
			collectionFiles[i] = ResolveGitignoreCollectionPath(file, "", configDir, fsys)
		}
	}
	return configWithGitignoreCollectionFiles(
		config,
		configDir,
		fsys,
		collectionFiles,
		isDirectoryBlocked,
		stopDirs,
	)
}

// ConfigWithGitignoreForExactTarget collects the .gitignore chain for one
// target whose lexical and canonical identities were already frozen by the
// caller. It never resolves the target through the filesystem again.
func ConfigWithGitignoreForExactTarget(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	target PathIdentity,
) RslintConfig {
	collectionFile := resolveGitignoreCollectionTargetPath(target, configDir, fsys)
	return configWithGitignoreCollectionFiles(
		config,
		configDir,
		fsys,
		[]string{collectionFile},
		nil,
		nil,
	)
}

func configWithGitignoreCollectionFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	collectionFiles []string,
	isDirectoryBlocked func(string) bool,
	stopDirs []string,
) RslintConfig {
	patterns := gitignore.CollectPatternsWithBoundaries(configDir, fsys, collectionFiles, isDirectoryBlocked, stopDirs)
	caseInsensitive := fsys != nil && !fsys.UseCaseSensitiveFileNames()
	patterns = CollectedGitignorePatternsForRoot(patterns, configDir, configDir, fsys)
	return ConfigWithCollectedGitignoreScopes(
		config,
		patterns,
		[]string{configDir},
		configDir,
		fsys,
		caseInsensitive,
	)
}

// ConfigWithGitignoreForTargets prepends only the .gitignore patterns that can
// affect the supplied exact files and recursive directory targets. When both
// target slices are nil, it retains the full-tree behavior used by no-argument
// CLI and LSP calls.
func ConfigWithGitignoreForTargets(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	targetFiles []string,
	targetDirectories []string,
) RslintConfig {
	return ConfigWithGitignoreForTargetsFromRoot(
		config,
		configDir,
		configDir,
		fsys,
		targetFiles,
		targetDirectories,
	)
}

// ConfigWithGitignoreForTargetsFromRoot keeps the authored config path space
// separate from the invocation's Git-ignore root. This matters for an explicit
// config outside cwd: files and ignores remain relative to configDir, while
// .gitignore sources still come from the files and directories the user asked
// Rslint to scan.
func ConfigWithGitignoreForTargetsFromRoot(
	config RslintConfig,
	configDir string,
	gitignoreRoot string,
	fsys vfs.FS,
	targetFiles []string,
	targetDirectories []string,
) RslintConfig {
	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	configDir = tspath.NormalizePath(configDir)
	matcher := NewGlobalIgnoreMatcher(config, configDir, fsys)
	hasGlobalIgnores := hasGlobalIgnoreEntries(config)
	var patterns []gitignore.Pattern
	scopes := PlanGitignoreCollectionScopes(
		gitignoreRoot,
		fsys,
		targetFiles,
		targetDirectories,
	)
	roots := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		roots = append(roots, scope.Root)
		var collectionFiles []string
		if scope.Files != nil {
			collectionFiles = make([]string, len(scope.Files))
			for index, file := range scope.Files {
				collectionFiles[index] = ResolveGitignoreCollectionPath(file, "", scope.Root, fsys)
			}
		}
		var collectionDirectories []string
		if scope.Directories != nil {
			collectionDirectories = make([]string, len(scope.Directories))
			for index, directory := range scope.Directories {
				collectionDirectories[index] = ResolveGitignoreCollectionDirectory(directory, scope.Root, fsys)
			}
		}
		var isDirectoryBlocked func(string) bool
		if hasGlobalIgnores {
			root := scope.Root
			isDirectoryBlocked = func(relativePath string) bool {
				absolutePath := tspath.ResolvePath(root, relativePath)
				return matcher.BlocksDirectory(absolutePath, "")
			}
		}
		collected := gitignore.CollectPatternsForTargets(
			scope.Root,
			fsys,
			collectionFiles,
			collectionDirectories,
			isDirectoryBlocked,
		)
		patterns = append(patterns, CollectedGitignorePatternsForRoot(
			collected,
			configDir,
			scope.Root,
			fsys,
		)...)
	}
	caseInsensitive := !useCaseSensitive
	return ConfigWithCollectedGitignoreScopes(
		config,
		patterns,
		roots,
		configDir,
		fsys,
		caseInsensitive,
	)
}

func hasGlobalIgnoreEntries(config RslintConfig) bool {
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			return true
		}
	}
	return false
}

// ConfigWithCollectedGitignore prepends an already-collected Git projection
// whose patterns carry their own root metadata. Filesystem-aware collection
// paths use ConfigWithCollectedGitignoreScopes so empty active scopes are also
// retained.
func ConfigWithCollectedGitignore(config RslintConfig, patterns []gitignore.Pattern, caseInsensitive bool) RslintConfig {
	if len(patterns) == 0 {
		return config
	}
	parsed := parseCollectedGitignorePatterns(patterns, caseInsensitive)
	return configWithCollectedGitignoreMetadata(
		config,
		patterns,
		parsed,
		collectedGitignoreScopesFromPatterns(parsed),
	)
}

// ConfigWithCollectedGitignoreScopes prepends one collected Git projection and
// retains every active collection root, including roots that produced no
// patterns. Empty roots are still semantically relevant when another root has
// patterns: they prevent canonical aliases from borrowing that other root's
// Git policy.
func ConfigWithCollectedGitignoreScopes(
	config RslintConfig,
	patterns []gitignore.Pattern,
	scopeRoots []string,
	configDirectory string,
	fsys vfs.FS,
	caseInsensitive bool,
) RslintConfig {
	if len(patterns) == 0 {
		return config
	}
	parsed := parseCollectedGitignorePatterns(patterns, caseInsensitive)
	scopes := make([]collectedGitignoreScope, 0, len(scopeRoots))
	for _, root := range scopeRoots {
		resolved := resolveCollectedGitignoreScope(
			configDirectory,
			root,
			fsys,
			caseInsensitive,
		)
		duplicate := false
		for _, existing := range scopes {
			if existing == resolved {
				duplicate = true
				break
			}
		}
		if !duplicate {
			scopes = append(scopes, resolved)
		}
	}
	return configWithCollectedGitignoreMetadata(config, patterns, parsed, scopes)
}

func configWithCollectedGitignoreMetadata(
	config RslintConfig,
	patterns []gitignore.Pattern,
	parsed []IgnorePattern,
	scopes []collectedGitignoreScope,
) RslintConfig {
	bindCollectedGitignorePatterns(parsed, scopes)
	gitignoreEntry := ConfigEntry{
		Ignores: make([]string, len(patterns)),
		collectedGitignore: &collectedGitignoreMetadata{
			ignores: parsed,
			scopes:  scopes,
		},
	}
	for index, pattern := range patterns {
		gitignoreEntry.Ignores[index] = pattern.Glob
	}
	effective := make(RslintConfig, 0, len(config)+1)
	effective = append(effective, gitignoreEntry)
	effective = append(effective, config...)
	return effective
}

func collectedGitignoreScopesFromPatterns(patterns []IgnorePattern) []collectedGitignoreScope {
	var scopes []collectedGitignoreScope
	for _, pattern := range patterns {
		if !pattern.GitPattern || !patternHasMatchDirectory(pattern) {
			continue
		}
		scope := collectedGitignoreScope{
			matchDirectory:    pattern.MatchDirectory,
			physicalDirectory: pattern.PhysicalMatchDirectory,
			lexicalDirectory:  pattern.LexicalMatchDirectory,
			caseInsensitive:   pattern.CaseInsensitive,
		}
		if scope.lexicalDirectory == "" {
			scope.lexicalDirectory = scope.matchDirectory
		}
		if scope.physicalDirectory == "" {
			scope.physicalDirectory = scope.matchDirectory
		}
		duplicate := false
		for _, existing := range scopes {
			if existing == scope {
				duplicate = true
				break
			}
		}
		if !duplicate {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func bindCollectedGitignorePatterns(patterns []IgnorePattern, scopes []collectedGitignoreScope) {
	for patternIndex := range patterns {
		pattern := &patterns[patternIndex]
		if !pattern.GitPattern || !patternHasMatchDirectory(*pattern) {
			continue
		}
		pattern.gitScope = unmatchedGitScope
		scope := collectedGitignoreScope{
			matchDirectory:    pattern.MatchDirectory,
			physicalDirectory: pattern.PhysicalMatchDirectory,
			lexicalDirectory:  pattern.LexicalMatchDirectory,
			caseInsensitive:   pattern.CaseInsensitive,
		}
		if scope.lexicalDirectory == "" {
			scope.lexicalDirectory = scope.matchDirectory
		}
		if scope.physicalDirectory == "" {
			scope.physicalDirectory = scope.matchDirectory
		}
		for scopeIndex, candidate := range scopes {
			if candidate == scope {
				pattern.gitScope = uint32(scopeIndex + 1)
				break
			}
		}
	}
}

// parseCollectedGitignorePatterns projects collected Git patterns onto the
// flat-config matcher without turning them into irreversible ESLint directory
// blocks. The synthetic patterns still participate in the same ordered list as
// authored config ignores, so a later config negation can re-include a target.
func parseCollectedGitignorePatterns(collected []gitignore.Pattern, caseInsensitive bool) []IgnorePattern {
	patterns := make([]IgnorePattern, 0, len(collected))
	parse := func(raw string) IgnorePattern {
		pattern := ParseIgnorePattern(raw)
		pattern.CaseInsensitive = caseInsensitive
		return pattern
	}
	for _, source := range collected {
		body := source.Glob
		if source.Negated {
			body = strings.TrimPrefix(body, "!")
		}
		nodeGlob := normalizePattern(source.NodeGlob)
		if body == "" || nodeGlob == "" {
			continue
		}

		// Every Git rule can match a directory node, including a rule without a
		// trailing slash (for example "build"). Keep a subtree projection for
		// sound walk pruning. The original node matcher remains a prefix of that
		// projection and GitNodeGlobEnd records its boundary.
		projection := body
		if strings.HasSuffix(projection, "/**") && !strings.HasSuffix(projection, "/**/*") {
			projection += "/*"
		} else if !strings.HasSuffix(projection, "/**/*") {
			projection += "/**/*"
		}
		if source.Negated {
			projection = "!" + projection
		}
		pattern := parse(projection)
		pattern.GitPattern = true
		pattern.GitDirectoryOnly = source.DirectoryOnly
		pattern.GitContentsOnly = source.ContentsOnly
		pattern.MatchDirectory = source.MatchDirectory
		pattern.PhysicalMatchDirectory = source.PhysicalMatchDirectory
		pattern.LexicalMatchDirectory = source.LexicalMatchDirectory
		if caseInsensitive {
			// Git patterns are immutable after parsing. Fold them once here;
			// the per-file matcher then only needs to fold the target path once,
			// rather than allocating lower-case copies for every pattern/node.
			pattern.Glob = strings.ToLower(pattern.Glob)
			nodeGlob = strings.ToLower(nodeGlob)
		}
		if !strings.HasPrefix(pattern.Glob, nodeGlob) {
			// Collected patterns always satisfy this compact-representation
			// invariant. Reject an inconsistent manually constructed value
			// instead of slicing an unrelated projection at match time.
			continue
		}
		pattern.GitNodeGlobEnd = len(nodeGlob)
		patterns = append(patterns, pattern)
	}
	return patterns
}

// ResolveGitignoreCollectionPath maps one exact target into the config root's
// lexical path space. This keeps Git source lookup stable when the config root
// and target use different symlink, casing, or canonical spellings.
func ResolveGitignoreCollectionPath(filePath string, canonicalPath string, configDir string, fsys vfs.FS) string {
	return resolveGitignoreCollectionTargetPath(PathIdentity{
		Path:          filePath,
		CanonicalPath: canonicalPath,
	}, configDir, fsys)
}

func resolveGitignoreCollectionTargetPath(
	target PathIdentity,
	configDir string,
	fsys vfs.FS,
) string {
	target.Path = tspath.NormalizePath(target.Path)
	matchFile, matchDir := ResolveConfigFilePathSpaceForTarget(target, configDir, fsys)
	if relative, ok := RelativePathWithinConfigRoot(matchFile, matchDir, true); ok {
		return tspath.ResolvePath(configDir, relative)
	}
	return target.Path
}

// ResolveGitignoreCollectionDirectory is the directory-target form of
// ResolveGitignoreCollectionPath. A directory that physically contains the
// config root maps to the root because its relevant projection is the complete
// config-owned tree.
func ResolveGitignoreCollectionDirectory(directory string, configDir string, fsys vfs.FS) string {
	directory = tspath.NormalizePath(directory)
	matchDirectory, matchConfigDir := ResolveConfigDirectoryPathSpace(directory, configDir, fsys)
	if relative, ok := RelativePathWithinConfigRoot(matchDirectory, matchConfigDir, true); ok {
		return tspath.ResolvePath(configDir, relative)
	}
	if _, containsConfigRoot := RelativePathWithinConfigRoot(matchConfigDir, matchDirectory, true); containsConfigRoot {
		return tspath.NormalizePath(configDir)
	}
	return directory
}
