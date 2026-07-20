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
		configIgnores := extractConfigIgnores(config)
		if len(configIgnores) > 0 {
			isDirectoryBlocked = func(relativePath string) bool {
				return isDirAbsolutelyBlocked(relativePath, configIgnores)
			}
		}
	} else if fsys != nil && len(targetFiles) > 0 {
		collectionFiles = make([]string, len(targetFiles))
		for i, file := range targetFiles {
			collectionFiles[i] = ResolveGitignoreCollectionPath(file, "", configDir, fsys)
		}
	}
	globs := gitignore.CollectWithBoundaries(configDir, fsys, collectionFiles, isDirectoryBlocked, stopDirs)
	caseInsensitive := fsys != nil && !fsys.UseCaseSensitiveFileNames()
	return ConfigWithCollectedGitignore(config, globs, caseInsensitive)
}

// ConfigWithCollectedGitignore prepends one already-collected Git projection
// without retaining a filesystem. Both the standalone collector path and
// staged config discovery use this constructor so private Git matching metadata
// cannot diverge between them.
func ConfigWithCollectedGitignore(config RslintConfig, globs []string, caseInsensitive bool) RslintConfig {
	if len(globs) == 0 {
		return config
	}
	gitignoreEntry := ConfigEntry{
		Ignores:                  append([]string(nil), globs...),
		gitignoreSemantics:       true,
		gitignoreCaseInsensitive: caseInsensitive,
	}
	effective := make(RslintConfig, 0, len(config)+1)
	effective = append(effective, gitignoreEntry)
	effective = append(effective, config...)
	return effective
}

// parseCollectedGitignorePatterns projects collected Git patterns onto the
// flat-config matcher without turning them into irreversible ESLint directory
// blocks. The synthetic patterns still participate in the same ordered list as
// authored config ignores, so a later config negation can re-include a target.
func parseCollectedGitignorePatterns(globs []string, caseInsensitive bool) []IgnorePattern {
	patterns := make([]IgnorePattern, 0, len(globs)*2)
	parse := func(raw string) IgnorePattern {
		pattern := ParseIgnorePattern(raw)
		pattern.CaseInsensitive = caseInsensitive
		return pattern
	}
	for _, raw := range globs {
		negated := strings.HasPrefix(raw, "!")
		body := strings.TrimPrefix(raw, "!")
		if body == "" {
			continue
		}

		prefix := ""
		if negated {
			prefix = "!"
		}
		if strings.HasSuffix(body, "/**") && !strings.HasSuffix(body, "/**/*") {
			patterns = append(patterns, parse(prefix+body+"/*"))
			continue
		}
		if strings.HasSuffix(body, "/**/*") {
			patterns = append(patterns, parse(raw))
			continue
		}

		direct := parse(raw)
		direct.Kind = dirNone
		patterns = append(patterns, direct)
		patterns = append(patterns, parse(prefix+body+"/**/*"))
	}
	return patterns
}

// ResolveGitignoreCollectionPath maps one exact target into the config root's
// lexical path space. This keeps Git source lookup stable when the config root
// and target use different symlink, casing, or canonical spellings.
func ResolveGitignoreCollectionPath(filePath string, canonicalPath string, configDir string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	matchFile, matchDir := ResolveConfigPathSpaceWithCanonical(filePath, canonicalPath, configDir, fsys)
	if relative, ok := RelativePathWithinConfigRoot(matchFile, matchDir, true); ok {
		return tspath.ResolvePath(configDir, relative)
	}
	return filePath
}
