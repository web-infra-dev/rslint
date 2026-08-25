package server

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type canonicalPathVFS struct {
	vfs.FS
	canonicalPaths map[string]string
}

func (fs *canonicalPathVFS) Realpath(filePath string) string {
	if canonicalPath := fs.canonicalPaths[rslintconfig.ExactPathID(filePath)]; canonicalPath != "" {
		return canonicalPath
	}
	return fs.FS.Realpath(filePath)
}

// SourceHasBOM forwards to the wrapped VFS. Embedding vfs.FS promotes only
// that interface's methods, so this layer has to pass [utils.BOMSource]
// through by hand or every overlay identity below it becomes invisible.
func (fs *canonicalPathVFS) SourceHasBOM(filePath string) bool {
	return utils.SourceHasBOM(fs.FS, filePath)
}

func addEquivalentFileContentPaths(fileContents map[string]string, configDirectory string, currentDirectory string, fs vfs.FS) {
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

	comparePathOptions := tspath.ComparePathsOptions{
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
		if fromDir == "" || toDir == "" || !tspath.ContainsPath(fromDir, filePath, comparePathOptions) {
			return
		}
		relativePath := tspath.GetRelativePathFromDirectory(fromDir, filePath, comparePathOptions)
		if relativePath == "" {
			return
		}
		addAlias(tspath.ResolvePath(toDir, relativePath), content)
	}

	realConfigDirectory := fs.Realpath(configDirectory)
	for _, entry := range entries {
		if realPath := fs.Realpath(entry.path); realPath != "" && realPath != entry.path {
			addAlias(realPath, entry.content)
		}
		if realConfigDirectory != "" && tspath.ComparePaths(configDirectory, realConfigDirectory, comparePathOptions) != 0 {
			addDirectoryAlias(configDirectory, realConfigDirectory, entry.path, entry.content)
			addDirectoryAlias(realConfigDirectory, configDirectory, entry.path, entry.content)
		}
	}
}

type equivalentFileContentUpdate struct {
	content string
	paths   []string
}

// setEquivalentFileContents overwrites every known spelling of the updated
// request-local sources. OverlayVFS checks an exact virtual key before resolving
// real paths, so leaving any pre-existing alias stale can make the next fix
// generation parse the previous text. All updates in one generation are batched
// so a large multi-file fix scans the overlay only once.
func setEquivalentFileContents(fileContents map[string]string, fs vfs.FS, updates []equivalentFileContentUpdate) {
	if fileContents == nil || len(updates) == 0 {
		return
	}

	updateByIdentity := make(map[string]int)
	addIdentity := func(updateIndex int, filePath string) {
		if filePath == "" {
			return
		}
		filePath = tspath.NormalizePath(filePath)
		identity := rslintconfig.ExactPathID(filePath)
		if _, exists := updateByIdentity[identity]; !exists {
			updateByIdentity[identity] = updateIndex
		}
		if fs != nil {
			if realPath := fs.Realpath(filePath); realPath != "" {
				realIdentity := rslintconfig.ExactPathID(tspath.NormalizePath(realPath))
				if _, exists := updateByIdentity[realIdentity]; !exists {
					updateByIdentity[realIdentity] = updateIndex
				}
			}
		}
	}
	for updateIndex, update := range updates {
		for _, filePath := range update.paths {
			addIdentity(updateIndex, filePath)
		}
	}

	for filePath := range fileContents {
		normalizedPath := tspath.NormalizePath(filePath)
		updateIndex, matches := updateByIdentity[rslintconfig.ExactPathID(normalizedPath)]
		if !matches && fs != nil {
			realPath := fs.Realpath(normalizedPath)
			if realPath != "" {
				updateIndex, matches = updateByIdentity[rslintconfig.ExactPathID(tspath.NormalizePath(realPath))]
			}
		}
		if matches {
			fileContents[filePath] = updates[updateIndex].content
		}
	}
	for _, update := range updates {
		for _, filePath := range update.paths {
			if filePath == "" {
				continue
			}
			normalizedPath := tspath.NormalizePath(filePath)
			fileContents[normalizedPath] = update.content
			if fs != nil {
				if realPath := fs.Realpath(normalizedPath); realPath != "" {
					fileContents[tspath.NormalizePath(realPath)] = update.content
				}
			}
		}
	}
}
