package config

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// PathIdentity keeps the caller-visible path together with the physical file
// and parent identities observed by the caller. Config matching consumes this
// value without owning target discovery or re-resolving an already frozen
// file.
type PathIdentity struct {
	Path                string
	CanonicalPath       string
	CanonicalParentPath string
}

// DirectoryIdentity is the lexical and physical identity of one directory.
// It is a low-level path value shared by lint-target and Git-ignore scope
// planning; neither consumer owns the other's policy.
type DirectoryIdentity struct {
	LexicalPath   string
	CanonicalPath string
}

// ResolveFileIdentities freezes requested file spellings without changing
// their order or the distinction between omitted and empty input lists.
func ResolveFileIdentities(paths []string, fsys vfs.FS) []PathIdentity {
	if paths == nil {
		return nil
	}
	identities := make([]PathIdentity, len(paths))
	for index, path := range paths {
		identities[index] = FreezeFileIdentity(path, fsys)
	}
	return identities
}

// ResolveDirectoryIdentities freezes requested directory spellings without
// changing their order or containment.
func ResolveDirectoryIdentities(directories []string, fsys vfs.FS) []DirectoryIdentity {
	if directories == nil {
		return nil
	}
	resolved := make([]DirectoryIdentity, 0, len(directories))
	for _, directory := range directories {
		lexicalPath := tspath.NormalizePath(directory)
		canonicalPath := lexicalPath
		if fsys != nil {
			if realPath := fsys.Realpath(lexicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		resolved = append(resolved, DirectoryIdentity{
			LexicalPath:   lexicalPath,
			CanonicalPath: canonicalPath,
		})
	}
	return resolved
}

// FreezeFileIdentity observes the caller-visible file and its physical parent
// once at the input boundary. A second parent observation retries a concurrent
// directory-alias move instead of combining two known generations.
func FreezeFileIdentity(filePath string, fsys vfs.FS) PathIdentity {
	filePath = tspath.NormalizePath(filePath)
	parent := tspath.GetDirectoryPath(filePath)
	identity := PathIdentity{Path: filePath, CanonicalPath: filePath, CanonicalParentPath: parent}
	if fsys == nil {
		return identity
	}
	for range 2 {
		before := canonicalPathOrSelf(parent, fsys)
		identity.CanonicalPath = canonicalPathOrSelf(filePath, fsys)
		identity.CanonicalParentPath = canonicalPathOrSelf(parent, fsys)
		if before == identity.CanonicalParentPath {
			break
		}
	}
	return identity
}

func canonicalPathOrSelf(path string, fsys vfs.FS) string {
	if fsys != nil {
		if canonical := fsys.Realpath(path); canonical != "" {
			return tspath.NormalizePath(canonical)
		}
	}
	return tspath.NormalizePath(path)
}
