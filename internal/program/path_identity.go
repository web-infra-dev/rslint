package program

import (
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// PathIdentity is a lexical path paired with an authoritative physical path
// already established by target discovery.
type PathIdentity struct {
	Path          string
	CanonicalPath string
}

type pathDirectoryIdentity struct {
	canonicalPath string
	entries       vfs.Entries
}

// PathIdentityResolver is a generation-local, concurrency-safe cache for
// exact physical file identities. It is deliberately unaware of projects,
// config ownership, and source membership.
type PathIdentityResolver struct {
	fsys           vfs.FS
	singleThreaded bool

	mu                     sync.Mutex
	canonicalByLexicalPath map[string]string
	directoryIdentities    map[string]pathDirectoryIdentity
}

func NewPathIdentityResolver(
	fsys vfs.FS,
	singleThreaded bool,
	known []PathIdentity,
) *PathIdentityResolver {
	resolver := &PathIdentityResolver{
		fsys:                   fsys,
		singleThreaded:         singleThreaded,
		canonicalByLexicalPath: make(map[string]string, len(known)*2),
	}
	canonicalIDs := make(map[string]struct{}, len(known))
	for _, identity := range known {
		if identity.CanonicalPath == "" {
			continue
		}
		canonicalID := pathIdentityID(identity.CanonicalPath)
		canonicalIDs[canonicalID] = struct{}{}
		resolver.canonicalByLexicalPath[canonicalID] = canonicalID
	}
	for _, identity := range known {
		if identity.Path == "" || identity.CanonicalPath == "" {
			continue
		}
		lexicalID := pathIdentityID(identity.Path)
		if _, isCanonicalIdentity := canonicalIDs[lexicalID]; !isCanonicalIdentity {
			resolver.canonicalByLexicalPath[lexicalID] = pathIdentityID(identity.CanonicalPath)
		}
	}
	return resolver
}

// CanonicalPathIDs resolves normalized lexical paths to exact canonical IDs.
// Known discovery identities win over filesystem inference so case-distinct
// overlay files cannot be collapsed by a global filesystem case flag.
func (resolver *PathIdentityResolver) CanonicalPathIDs(paths []string) []string {
	canonicalIDs := make([]string, len(paths))
	if resolver == nil || len(paths) == 0 {
		return canonicalIDs
	}
	resolver.mu.Lock()
	defer resolver.mu.Unlock()

	unknownIndexByLexicalID := make(map[string]int)
	var unknownPaths []string
	positionsByUnknownIndex := make([][]int, 0)
	for position, path := range paths {
		lexicalID := pathIdentityID(path)
		if canonicalID, known := resolver.canonicalByLexicalPath[lexicalID]; known {
			canonicalIDs[position] = canonicalID
			continue
		}
		unknownIndex, exists := unknownIndexByLexicalID[lexicalID]
		if !exists {
			unknownIndex = len(unknownPaths)
			unknownIndexByLexicalID[lexicalID] = unknownIndex
			unknownPaths = append(unknownPaths, tspath.NormalizePath(path))
			positionsByUnknownIndex = append(positionsByUnknownIndex, nil)
		}
		positionsByUnknownIndex[unknownIndex] = append(positionsByUnknownIndex[unknownIndex], position)
	}

	resolvedIDs := resolver.resolveUnknownPaths(unknownPaths)
	for unknownIndex, canonicalID := range resolvedIDs {
		lexicalID := pathIdentityID(unknownPaths[unknownIndex])
		if existing, known := resolver.canonicalByLexicalPath[lexicalID]; known {
			canonicalID = existing
		} else {
			resolver.canonicalByLexicalPath[lexicalID] = canonicalID
		}
		for _, position := range positionsByUnknownIndex[unknownIndex] {
			canonicalIDs[position] = canonicalID
		}
	}
	return canonicalIDs
}

func (resolver *PathIdentityResolver) resolveUnknownPaths(paths []string) []string {
	canonicalIDs := make([]string, len(paths))
	if len(paths) == 0 {
		return canonicalIDs
	}
	if resolver.fsys == nil {
		for index, path := range paths {
			canonicalIDs[index] = pathIdentityID(path)
		}
		return canonicalIDs
	}
	if resolver.directoryIdentities == nil {
		resolver.directoryIdentities = make(map[string]pathDirectoryIdentity)
	}

	pathIndexesByDirectoryID := make(map[string][]int)
	directoryPathByID := make(map[string]string)
	for index, path := range paths {
		directoryPath := tspath.GetDirectoryPath(path)
		directoryID := pathIdentityID(directoryPath)
		directoryPathByID[directoryID] = directoryPath
		pathIndexesByDirectoryID[directoryID] = append(pathIndexesByDirectoryID[directoryID], index)
	}
	directoryIDs := make([]string, 0, len(pathIndexesByDirectoryID))
	for directoryID := range pathIndexesByDirectoryID {
		directoryIDs = append(directoryIDs, directoryID)
	}
	sort.Strings(directoryIDs)

	pendingIdentities := make([]pathDirectoryIdentity, len(directoryIDs))
	hasPendingIdentity := make([]bool, len(directoryIDs))
	useCaseSensitive := resolver.fsys.UseCaseSensitiveFileNames()
	work := core.NewWorkGroup(resolver.singleThreaded)
	queuePath := func(pathIndex int, directory pathDirectoryIdentity) {
		work.Queue(func() {
			path := paths[pathIndex]
			fileName, regular := regularFileNameFromEntries(
				directory.entries,
				tspath.GetBaseFileName(path),
				useCaseSensitive,
			)
			if regular {
				canonicalIDs[pathIndex] = pathIdentityID(
					tspath.CombinePaths(directory.canonicalPath, fileName),
				)
			} else {
				canonicalIDs[pathIndex] = pathIdentityID(authoritativePath(path, resolver.fsys))
			}
		})
	}
	for directoryIndex, directoryID := range directoryIDs {
		pathIndexes := pathIndexesByDirectoryID[directoryID]
		directory, cached := resolver.directoryIdentities[directoryID]
		if len(pathIndexes) == 1 && !cached {
			pathIndex := pathIndexes[0]
			work.Queue(func() {
				canonicalIDs[pathIndex] = pathIdentityID(authoritativePath(paths[pathIndex], resolver.fsys))
			})
			continue
		}
		if cached {
			for _, pathIndex := range pathIndexes {
				queuePath(pathIndex, directory)
			}
			continue
		}
		work.Queue(func() {
			directoryPath := directoryPathByID[directoryID]
			directory = pathDirectoryIdentity{
				canonicalPath: authoritativePath(directoryPath, resolver.fsys),
				entries:       resolver.fsys.GetAccessibleEntries(directoryPath),
			}
			pendingIdentities[directoryIndex] = directory
			hasPendingIdentity[directoryIndex] = true
			for _, pathIndex := range pathIndexes {
				queuePath(pathIndex, directory)
			}
		})
	}
	work.RunAndWait()
	for index, resolved := range hasPendingIdentity {
		if resolved {
			resolver.directoryIdentities[directoryIDs[index]] = pendingIdentities[index]
		}
	}
	return canonicalIDs
}

func regularFileNameFromEntries(entries vfs.Entries, fileName string, useCaseSensitive bool) (string, bool) {
	if entries.Symlinks == nil {
		return "", false
	}
	canonicalFileName := tspath.GetCanonicalFileName(fileName, useCaseSensitive)
	match := ""
	for _, entryName := range entries.Files {
		if tspath.GetCanonicalFileName(entryName, useCaseSensitive) != canonicalFileName {
			continue
		}
		if match != "" && match != entryName {
			return "", false
		}
		match = entryName
	}
	if match == "" {
		return "", false
	}
	if _, isSymlink := entries.Symlinks[match]; isSymlink {
		return "", false
	}
	return match, true
}

func authoritativePath(path string, fsys vfs.FS) string {
	path = tspath.NormalizePath(path)
	if fsys != nil {
		if realPath := fsys.Realpath(path); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return path
}

func pathIdentityID(path string) string {
	return string(tspath.ToPath(tspath.NormalizePath(path), "", true))
}
