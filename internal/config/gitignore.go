package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// ConfigWithGitignore prepends the .gitignore patterns that apply to a lint
// invocation. A nil targetFiles slice scans the config directory, as every CLI
// mode does; a non-nil slice limits collection to the ancestor chains of those
// explicit files, as used by API and LSP requests. The input config is never
// mutated.
func ConfigWithGitignore(config RslintConfig, configDir string, fsys vfs.FS, targetFiles []string) RslintConfig {
	var globs []string
	if targetFiles == nil {
		globs = ReadGitignoreAsGlobs(configDir, fsys, ExtractConfigIgnores(config))
	} else {
		collectionFiles := targetFiles
		if fsys != nil && len(targetFiles) > 0 {
			collectionFiles = make([]string, len(targetFiles))
			for i, file := range targetFiles {
				collectionFiles[i] = gitignoreCollectionFilePath(file, configDir, fsys)
			}
		}
		globs = ReadGitignoreAsGlobsForFiles(configDir, fsys, collectionFiles)
	}
	if len(globs) == 0 {
		return config
	}
	return append(RslintConfig{{Ignores: globs}}, config...)
}

func gitignoreCollectionFilePath(filePath string, configDir string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	matchFile, matchDir := ResolveConfigPathSpace(filePath, configDir, fsys)
	if relative, ok := relativeConfigPath(matchFile, matchDir, true); ok {
		return tspath.ResolvePath(configDir, relative)
	}
	return filePath
}

// ReadGitignoreAsGlobs collects ancestor and descendant .gitignore patterns
// for default directory discovery.
func ReadGitignoreAsGlobs(configDir string, fsys vfs.FS, configIgnores []IgnorePattern) []string {
	var isDirectoryBlocked gitignore.DirectoryBlocker
	if len(configIgnores) > 0 {
		isDirectoryBlocked = func(relativePath string) bool {
			return isDirAbsolutelyBlocked(relativePath, configIgnores)
		}
	}
	return gitignore.ReadGitignoreAsGlobs(configDir, fsys, isDirectoryBlocked)
}

// ReadGitignoreAsGlobsForFiles collects .gitignore patterns on explicit file
// ancestor chains without scanning unrelated descendants.
func ReadGitignoreAsGlobsForFiles(configDir string, fsys vfs.FS, files []string) []string {
	return gitignore.ReadGitignoreAsGlobsForFiles(configDir, fsys, files)
}

// ExtractConfigIgnores returns parsed global ignore entries for directory
// pruning during .gitignore discovery.
func ExtractConfigIgnores(config RslintConfig) []IgnorePattern {
	var ignores []string
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			ignores = append(ignores, entry.Ignores...)
		}
	}
	return ParseIgnorePatterns(ignores)
}
