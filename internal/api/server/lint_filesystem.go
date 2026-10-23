package server

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

type canonicalPathVFS struct {
	vfs.FS
	canonicalPaths map[string]string
}

func (fs *canonicalPathVFS) Realpath(filePath tspath.RootedPath) tspath.RootedPath {
	if canonicalPath := fs.canonicalPaths[rslintconfig.ExactPathID(filePath.AsString())]; canonicalPath != "" {
		return pathutil.Path(canonicalPath)
	}
	return fs.FS.Realpath(filePath)
}

// SourceHasBOM forwards to the wrapped VFS. Embedding vfs.FS promotes only
// that interface's methods, so this layer has to pass [utils.BOMSource]
// through by hand or every overlay identity below it becomes invisible.
func (fs *canonicalPathVFS) SourceHasBOM(filePath string) bool {
	return utils.SourceHasBOM(fs.FS, filePath)
}

func addEquivalentFileContentPaths(fileContents map[string]string, configDirectory string, currentDirectory string, fs vfs.FS, canonicalPaths map[string]string) {
	if len(fileContents) == 0 || fs == nil {
		return
	}

	type fileContentEntry struct {
		path    string
		content string
	}
	entries := make([]fileContentEntry, 0, len(fileContents))
	for filePath, content := range fileContents {
		entries = append(entries, fileContentEntry{path: filePath, content: content})
	}

	comparePathOptions := pathutil.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: true,
	}
	addAlias := func(alias string, content string) {
		if alias == "" {
			return
		}
		if _, exists := fileContents[alias]; exists {
			return
		}
		fileContents[alias] = content
	}
	addDirectoryAlias := func(fromDir string, toDir string, filePath string, content string) {
		if fromDir == "" || toDir == "" || !pathutil.ContainsPath(fromDir, filePath, comparePathOptions) {
			return
		}
		relativePath := pathutil.GetRelativePathFromDirectory(fromDir, filePath, comparePathOptions)
		if relativePath == "" {
			return
		}
		addAlias(tspath.ResolvePath(toDir, relativePath), content)
	}

	realConfigDirectory := fs.Realpath(pathutil.Path(configDirectory))
	for _, entry := range entries {
		// Overlay sources and lint targets share physical ancestry even when
		// several parent directories have not been created.
		identity := rslintconfig.FreezeFileIdentity(entry.path, fs)
		addAlias(tspath.CombinePaths(identity.CanonicalParentPath, tspath.GetBaseFileName(entry.path)), entry.content)
		addAlias(identity.CanonicalPath, entry.content)
		addAlias(canonicalPaths[rslintconfig.ExactPathID(entry.path)], entry.content)
		if realConfigDirectory != "" && pathutil.ComparePaths(configDirectory, realConfigDirectory.AsString(), comparePathOptions) != 0 {
			addDirectoryAlias(configDirectory, realConfigDirectory.AsString(), entry.path, entry.content)
			addDirectoryAlias(realConfigDirectory.AsString(), configDirectory, entry.path, entry.content)
		}
	}
}
