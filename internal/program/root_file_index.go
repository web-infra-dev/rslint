package program

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// RootFileIndex identifies files selected directly by a parsed TypeScript
// config. Dependencies admitted only through imports, libraries, types, or
// project resolution are deliberately absent.
type RootFileIndex struct {
	fileNames        []string
	fsys             vfs.FS
	useCaseSensitive bool
	exactRoots       map[string]struct{}
	canonicalOnce    sync.Once
	canonicalRoots   map[string]struct{}
}

func NewRootFileIndex(fileNames []string, fsys vfs.FS) *RootFileIndex {
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	exactRoots := make(map[string]struct{}, len(fileNames))
	for _, rootFileName := range fileNames {
		exactRoots[rootFilePathID(rootFileName, useCaseSensitive)] = struct{}{}
	}
	return &RootFileIndex{
		fileNames:        append([]string(nil), fileNames...),
		fsys:             fsys,
		useCaseSensitive: useCaseSensitive,
		exactRoots:       exactRoots,
	}
}

func rootFilePathID(filePath string, useCaseSensitive bool) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", useCaseSensitive))
}

func rootFileCanonicalPathID(filePath string, fsys vfs.FS, useCaseSensitive bool) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return rootFilePathID(filePath, useCaseSensitive)
}

// Contains reports whether fileName is a direct config root. canonicalFileName
// may carry an already-resolved physical identity; when omitted it is resolved
// only after exact lexical lookup misses.
func (index *RootFileIndex) Contains(fileName string, canonicalFileName string) bool {
	if index == nil || fileName == "" {
		return false
	}
	if _, ok := index.exactRoots[rootFilePathID(fileName, index.useCaseSensitive)]; ok {
		return true
	}
	if index.fsys == nil {
		return false
	}
	index.canonicalOnce.Do(func() {
		index.canonicalRoots = make(map[string]struct{}, len(index.fileNames))
		for _, rootFileName := range index.fileNames {
			index.canonicalRoots[rootFileCanonicalPathID(
				rootFileName,
				index.fsys,
				index.useCaseSensitive,
			)] = struct{}{}
		}
	})
	canonicalID := ""
	if canonicalFileName != "" {
		// Discovery already paid for this physical identity. Do not resolve it
		// again for every project probed during broad target binding.
		canonicalID = rootFilePathID(canonicalFileName, index.useCaseSensitive)
	} else {
		canonicalID = rootFileCanonicalPathID(fileName, index.fsys, index.useCaseSensitive)
	}
	_, ok := index.canonicalRoots[canonicalID]
	return ok
}
