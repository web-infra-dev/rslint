package target

import (
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
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
	files := freezeFilePathsForTest(allowFiles, fsys)
	explicitSet := newExplicitLintTargetSet(files, fsys)
	policy := rslintconfig.CollectGitPolicy(config, configDir, fsys, files, testDirectoryInputs(configDir, files, allowDirs, fsys))
	resetTargetWalkSpy(fsys)
	return discoverLintTargetsWithPreparedFiles(
		config,
		configDir,
		scanRoot,
		fsys,
		explicitSet.targetsForPaths(allowFiles),
		rslintconfig.ResolveDirectoryIdentities(allowDirs, fsys),
		nil,
		singleThreaded,
		pathSpaces,
		policy,
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
	files := freezeFilePathsForTest(allowFiles, fsys)
	explicitSet := newExplicitLintTargetSet(files, fsys)
	directories := allowDirs
	if files == nil && directories == nil {
		for directory := range configMap {
			directories = append(directories, directory)
		}
	}
	policy := rslintconfig.CollectGitPolicy(nil, "", fsys, files, testDirectoryInputs("", files, directories, fsys))
	resetTargetWalkSpy(fsys)
	return discoverLintTargetsMultiConfigWithPreparedFiles(
		configMap,
		scopes,
		fsys,
		files,
		rslintconfig.ResolveDirectoryIdentities(allowDirs, fsys),
		singleThreaded,
		ownerIndex,
		explicitSet,
		policy,
	)
}

// These suites measure target traversal after the independent Git observation.
func resetTargetWalkSpy(fsys vfs.FS) {
	if spy, ok := fsys.(interface{ resetAccessedDirs() }); ok {
		spy.resetAccessedDirs()
	}
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

func freezeFilePathsForTest(paths []string, fsys vfs.FS) []rslintconfig.PathIdentity {
	if paths == nil {
		return nil
	}
	files := make([]rslintconfig.PathIdentity, len(paths))
	for index, path := range paths {
		files[index] = rslintconfig.FreezeFileIdentity(path, fsys)
	}
	return files
}

func testDirectoryInputs(root string, files []rslintconfig.PathIdentity, directories []string, fsys vfs.FS) []rslintconfig.DirectoryIdentity {
	if files == nil && directories == nil && root != "" {
		directories = []string{root}
	}
	return rslintconfig.ResolveDirectoryIdentities(directories, fsys)
}

func withTestGitPolicy(request Request) Request {
	root := request.ScanRoot
	if root == "" {
		root = request.ConfigDirectory
	}
	if request.Files == nil && request.Directories == nil {
		request.Directories = rslintconfig.ResolveDirectoryIdentities([]string{root}, request.FS)
	}
	request.GitPolicy = rslintconfig.CollectGitPolicy(request.Config, request.ConfigDirectory, request.FS, request.Files, request.Directories)
	return request
}

func configWithGitForFileForTest(authored rslintconfig.RslintConfig, path, directory string) *rslintconfig.MergedConfig {
	fsys := osvfs.FS()
	identity := rslintconfig.FreezeFileIdentity(path, fsys)
	policy := rslintconfig.CollectGitPolicy(authored, directory, fsys, []rslintconfig.PathIdentity{identity}, nil)
	spaces := rslintconfig.NewPathSpaceSnapshot(map[string]rslintconfig.RslintConfig{directory: authored}, fsys)
	matcher, err := rslintconfig.NewTargetMatcherWithPathSpaces(authored, directory, fsys, spaces, policy)
	if err != nil {
		panic(err)
	}
	if matcher.MatchFile(identity).GloballyIgnored() {
		return nil
	}
	return authored.GetConfigForFile(path, directory)
}
