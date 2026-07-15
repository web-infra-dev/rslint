package config

import (
	"strings"

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
	config = WithDefaultGlobalIgnores(config)
	collectionFiles := targetFiles
	var isDirectoryBlocked func(string) bool
	if targetFiles == nil {
		isDirectoryBlocked = gitignoreDirectoryBlocker(config, configDir, fsys)
	} else if fsys != nil && len(targetFiles) > 0 {
		collectionFiles = make([]string, len(targetFiles))
		for i, file := range targetFiles {
			collectionFiles[i] = gitignoreCollectionFilePath(file, configDir, fsys)
		}
	}
	globs := gitignore.CollectWithBoundaries(configDir, fsys, collectionFiles, isDirectoryBlocked, stopDirs)
	return configWithCollectedGitignore(config, globs, fsys)
}

func gitignoreDirectoryBlocker(config RslintConfig, configDir string, fsys vfs.FS) func(string) bool {
	layers := compileConfigIgnoreLayers(config, configDir, fsys)
	if len(layers) == 0 {
		return nil
	}
	return func(relativePath string) bool {
		directory := resolvePathForRoot(configDir, configDir, relativePath)
		return isDirectoryIgnoredByConfigLayers(directory, "", layers, fsys)
	}
}

func configWithCollectedGitignore(config RslintConfig, globs []string, fsys vfs.FS) RslintConfig {
	if len(globs) == 0 {
		return config
	}
	caseInsensitive := fsys != nil && !fsys.UseCaseSensitiveFileNames()
	gitignoreEntry := ConfigEntry{
		Ignores:                  globs,
		gitignoreSemantics:       true,
		gitignoreCaseInsensitive: caseInsensitive,
	}
	// Keep one ordered policy: .gitignore, product defaults, authored config.
	// The product defaults remain the discovery baseline even when .gitignore
	// negates node_modules or .git; an authored config negation can deliberately
	// reopen either earlier source. This also keeps exact/LSP admission identical
	// to CLI/API discovery, whose .gitignore phase cannot expand the catalog.
	effective := make(RslintConfig, 0, len(config)+1)
	for _, entry := range config {
		if entry.defaultIgnores {
			effective = append(effective, gitignoreEntry, entry)
			continue
		}
		effective = append(effective, entry)
	}
	return effective
}

// parseCollectedGitignorePatterns projects collected Git patterns onto the
// flat-config matcher without turning them into irreversible ESLint directory
// blocks. The synthetic patterns still participate in the same ordered list as
// authored config ignores, so a later config negation can re-include a target.
func parseCollectedGitignorePatterns(globs []string, caseInsensitive bool) []IgnorePattern {
	patterns := make([]IgnorePattern, 0, len(globs)*2)
	parse := func(raw string) IgnorePattern {
		pattern := parseGitignorePattern(raw)
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
		patterns = append(patterns, direct)
		patterns = append(patterns, parse(prefix+body+"/**/*"))
	}
	return patterns
}

func gitignoreCollectionFilePath(filePath string, configDir string, fsys vfs.FS) string {
	filePath = normalizePathForRoot(configDir, filePath)
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	// The normal discovery path is already in the governing config's lexical
	// space. Preserve it without repeatedly resolving the same config root through realpath for
	// every directory and file query; physical projection is only needed for a
	// compiler-reported path or another alias outside that lexical root.
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, caseSensitive); ok {
		return resolvePathForRoot(configDir, configDir, relative)
	}
	matchFile, matchDir := ResolveConfigPathSpace(filePath, configDir, fsys)
	if relative, ok := RelativePathWithinConfigRoot(matchFile, matchDir, true); ok {
		return resolvePathForRoot(configDir, configDir, relative)
	}
	return filePath
}
