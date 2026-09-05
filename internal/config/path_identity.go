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

// CoalesceDirectoryIdentities returns the smallest independent set covering
// the requested roots. A nested root is removed only when lexical and physical
// containment produce the same suffix, preserving distinct symlink scopes and
// avoiding invalid comparisons across Windows drives or UNC shares.
func CoalesceDirectoryIdentities(directories []string, fsys vfs.FS) []DirectoryIdentity {
	resolved := ResolveDirectoryIdentities(directories, fsys)
	if len(resolved) == 0 {
		return nil
	}
	roots := make([]DirectoryIdentity, 0, len(resolved))
	contains := func(parent DirectoryIdentity, child DirectoryIdentity) bool {
		lexicalRelative, lexicalWithin := RelativePathWithinConfigRoot(
			child.LexicalPath,
			parent.LexicalPath,
			true,
		)
		canonicalRelative, canonicalWithin := RelativePathWithinConfigRoot(
			child.CanonicalPath,
			parent.CanonicalPath,
			true,
		)
		return lexicalWithin && canonicalWithin &&
			ExactPathID(lexicalRelative) == ExactPathID(canonicalRelative)
	}
	for _, candidate := range resolved {
		covered := false
		for _, existing := range roots {
			if contains(existing, candidate) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		kept := roots[:0]
		for _, existing := range roots {
			if !contains(candidate, existing) {
				kept = append(kept, existing)
			}
		}
		roots = append(kept, candidate)
	}
	return roots
}
