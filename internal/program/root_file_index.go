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
	fileNames      []string
	fsys           vfs.FS
	exactRoots     map[string]struct{}
	canonicalOnce  sync.Once
	canonicalRoots map[string]struct{}
}

func NewRootFileIndex(fileNames []string, fsys vfs.FS) *RootFileIndex {
	exactRoots := make(map[string]struct{}, len(fileNames))
	for _, rootFileName := range fileNames {
		exactRoots[rootFilePathID(rootFileName)] = struct{}{}
	}
	return &RootFileIndex{
		fileNames:  append([]string(nil), fileNames...),
		fsys:       fsys,
		exactRoots: exactRoots,
	}
}

func rootFilePathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func rootFileCanonicalPathID(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return rootFilePathID(filePath)
}

// Contains reports whether fileName is a direct config root. canonicalFileName
// may carry an already-resolved physical identity; when omitted it is resolved
// only after exact lexical lookup misses.
func (index *RootFileIndex) Contains(fileName string, canonicalFileName string) bool {
	if index == nil || fileName == "" {
		return false
	}
	if _, ok := index.exactRoots[rootFilePathID(fileName)]; ok {
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
			)] = struct{}{}
		}
	})
	canonicalID := ""
	if canonicalFileName != "" {
		// Discovery already paid for this physical identity. Do not resolve it
		// again for every project probed during broad target binding.
		canonicalID = rootFilePathID(canonicalFileName)
	} else {
		canonicalID = rootFileCanonicalPathID(fileName, index.fsys)
	}
	_, ok := index.canonicalRoots[canonicalID]
	return ok
}
