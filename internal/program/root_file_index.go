package program

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// RootFileIndex identifies files selected directly by a parsed TypeScript
// config. Dependencies admitted only through imports, libraries, types, or
// project resolution are deliberately absent.
type RootFileIndex struct {
	fileNames      []string
	resolver       *PathIdentityResolver
	batch          *rootFileIndexBatch
	exactRoots     map[string]struct{}
	canonicalOnce  sync.Once
	canonicalRoots map[string]struct{}
}

func NewRootFileIndex(fileNames []string, fsys vfs.FS) *RootFileIndex {
	var resolver *PathIdentityResolver
	if fsys != nil {
		resolver = NewPathIdentityResolver(fsys, true, nil)
	}
	return newRootFileIndex(fileNames, resolver, nil)
}

// NewRootFileIndexWithResolver shares one generation-local physical identity
// cache with other root or Program source indexes.
func NewRootFileIndexWithResolver(
	fileNames []string,
	resolver *PathIdentityResolver,
) *RootFileIndex {
	return newRootFileIndex(fileNames, resolver, nil)
}

// NewRootFileIndexes constructs root indexes that preserve exact-first lookup.
// Their first physical miss lazily resolves all supplied roots as one batch.
func NewRootFileIndexes(
	fileNamesByProject [][]string,
	resolver *PathIdentityResolver,
) []*RootFileIndex {
	rootPaths := make([]string, 0)
	for _, fileNames := range fileNamesByProject {
		rootPaths = append(rootPaths, fileNames...)
	}
	batch := &rootFileIndexBatch{resolver: resolver, rootPaths: rootPaths}
	indexes := make([]*RootFileIndex, len(fileNamesByProject))
	for project, fileNames := range fileNamesByProject {
		indexes[project] = newRootFileIndex(fileNames, resolver, batch)
	}
	return indexes
}

type rootFileIndexBatch struct {
	once      sync.Once
	resolver  *PathIdentityResolver
	rootPaths []string
}

func (batch *rootFileIndexBatch) resolve() {
	if batch == nil || batch.resolver == nil {
		return
	}
	batch.once.Do(func() {
		batch.resolver.CanonicalPathIDs(batch.rootPaths)
		batch.rootPaths = nil
	})
}

func newRootFileIndex(
	fileNames []string,
	resolver *PathIdentityResolver,
	batch *rootFileIndexBatch,
) *RootFileIndex {
	exactRoots := make(map[string]struct{}, len(fileNames))
	for _, rootFileName := range fileNames {
		exactRoots[pathIdentityID(rootFileName)] = struct{}{}
	}
	return &RootFileIndex{
		fileNames:  append([]string(nil), fileNames...),
		resolver:   resolver,
		batch:      batch,
		exactRoots: exactRoots,
	}
}

// Contains reports whether fileName is a direct config root. canonicalFileName
// may carry an already-resolved physical identity; when omitted it is resolved
// only after exact lexical lookup misses.
func (index *RootFileIndex) Contains(fileName string, canonicalFileName string) bool {
	if index == nil || fileName == "" {
		return false
	}
	exactID := pathIdentityID(fileName)
	if _, ok := index.exactRoots[exactID]; ok {
		return true
	}
	if index.resolver == nil {
		return false
	}
	index.canonicalOnce.Do(func() {
		index.batch.resolve()
		index.canonicalRoots = make(map[string]struct{}, len(index.fileNames))
		for _, canonicalID := range index.resolver.CanonicalPathIDs(index.fileNames) {
			if canonicalID != "" {
				index.canonicalRoots[canonicalID] = struct{}{}
			}
		}
	})
	canonicalID := ""
	if canonicalFileName != "" {
		// Discovery already paid for this physical identity. Do not resolve it
		// again for every project probed during broad target binding.
		canonicalID = pathIdentityID(canonicalFileName)
	} else {
		canonicalIDs := index.resolver.CanonicalPathIDs([]string{fileName})
		if len(canonicalIDs) > 0 {
			canonicalID = canonicalIDs[0]
		}
	}
	_, ok := index.canonicalRoots[canonicalID]
	return ok
}

// ContainsPathIDs is the pre-normalized form of Contains. Batch adapters use
// it when target discovery already supplied exact lexical and canonical IDs,
// avoiding the same path normalization for every candidate project probe.
func (index *RootFileIndex) ContainsPathIDs(exactID string, canonicalID string) bool {
	if index == nil || exactID == "" {
		return false
	}
	if _, ok := index.exactRoots[exactID]; ok {
		return true
	}
	if index.resolver == nil || canonicalID == "" {
		return false
	}
	index.canonicalOnce.Do(func() {
		index.batch.resolve()
		index.canonicalRoots = make(map[string]struct{}, len(index.fileNames))
		for _, resolvedID := range index.resolver.CanonicalPathIDs(index.fileNames) {
			if resolvedID != "" {
				index.canonicalRoots[resolvedID] = struct{}{}
			}
		}
	})
	_, ok := index.canonicalRoots[canonicalID]
	return ok
}
