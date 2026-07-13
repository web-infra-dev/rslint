package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// ConfigWithGitignore prepends the .gitignore patterns that apply to a lint
// invocation. A nil targetFiles slice scans the config directory, as the CLI
// path does; a non-nil slice limits collection to the ancestor chains of those
// explicit files, as used by API and LSP requests. The input config is never
// mutated.
func ConfigWithGitignore(config RslintConfig, configDir string, fsys vfs.FS, targetFiles []string) RslintConfig {
	collectionFiles := targetFiles
	var isDirectoryBlocked func(string) bool
	if targetFiles == nil {
		configIgnores := extractConfigIgnores(config)
		if len(configIgnores) > 0 {
			isDirectoryBlocked = func(relativePath string) bool {
				return isDirAbsolutelyBlocked(relativePath, configIgnores)
			}
		}
	} else {
		if fsys != nil && len(targetFiles) > 0 {
			collectionFiles = make([]string, len(targetFiles))
			for i, file := range targetFiles {
				collectionFiles[i] = gitignoreCollectionFilePath(file, configDir, fsys)
			}
		}
	}
	globs := gitignore.Collect(configDir, fsys, collectionFiles, isDirectoryBlocked)
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
