package target

import (
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func discoverLintFiles(
	config rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := discoverLintTargetsFromRoot(
		config, configDir, configDir, fsys, allowFiles, allowDirs, singleThreaded,
	)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

func discoverLintTargets(
	config rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []File {
	return discoverLintTargetsFromRoot(
		config, configDir, configDir, fsys, allowFiles, allowDirs, singleThreaded,
	)
}

func discoverLintTargetsFromRoot(
	config rslintconfig.RslintConfig,
	configDir string,
	scanRoot string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []File {
	pathSpaces := rslintconfig.NewPathSpaceSnapshot(
		map[string]rslintconfig.RslintConfig{configDir: config},
		fsys,
	)
	explicitSet := newExplicitLintTargetSet(allowFiles, fsys)
	return discoverLintTargetsWithPreparedFiles(
		config,
		configDir,
		scanRoot,
		fsys,
		explicitSet.targetsForPaths(allowFiles),
		allowDirs,
		nil,
		singleThreaded,
		pathSpaces,
	)
}

func discoverLintTargetsMultiConfig(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]OwnerScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []File {
	ownerIndex := NewOwnerIndex(configMap, fsys)
	explicitSet := newExplicitLintTargetSet(allowFiles, fsys)
	return discoverLintTargetsMultiConfigWithPreparedFiles(
		configMap,
		scopes,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
		ownerIndex,
		explicitSet,
	)
}

func discoverLintFilesMultiConfig(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := discoverLintTargetsMultiConfig(
		configMap, nil, fsys, allowFiles, allowDirs, singleThreaded,
	)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

func newConfigDirectoryIndex(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
) *configDirectoryIndex {
	return newConfigDirectoryIndexWithPathSpaces(
		configMap,
		fsys,
		rslintconfig.NewPathSpaceSnapshot(configMap, fsys),
	)
}

func newOwnerIndexForAutomaticTargets(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]OwnerScope,
	fsys vfs.FS,
) *OwnerIndex {
	return NewOwnerIndex(configMapForAutomaticTargets(configMap, scopes), fsys)
}
