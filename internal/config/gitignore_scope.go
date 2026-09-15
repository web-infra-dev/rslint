package config

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// GitignoreCollectionScope is one independent lexical root used to collect Git
// ignore sources for an invocation-wide config. Targets representable in the
// invocation cwd share that root. A requested directory outside cwd starts its
// own root. Exact files outside every supplied directory scope do not invent a
// new Git boundary; they retain the established cwd-bounded policy. This keeps
// multiple sibling directory targets independent without changing the authored
// base of the governing rslint config.
type GitignoreCollectionScope struct {
	Root        string
	Files       []string
	Directories []string
}

// PlanGitignoreCollectionScopes converts one invocation cwd plus its target
// union into deterministic Git collection roots. The cwd remains the implicit
// full-scan root. Descendant and verified alias targets stay in that scope;
// sibling directory targets get independent roots.
func PlanGitignoreCollectionScopes(
	defaultRoot string,
	fsys vfs.FS,
	targetFiles []string,
	targetDirectories []string,
) []GitignoreCollectionScope {
	defaultRoot = tspath.NormalizePath(defaultRoot)
	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	pathID := func(path string) tspath.Path {
		return tspath.ToPath(tspath.NormalizePath(path), "", useCaseSensitive)
	}

	var defaultScope *GitignoreCollectionScope
	ensureDefaultScope := func() *GitignoreCollectionScope {
		if defaultScope == nil {
			defaultScope = &GitignoreCollectionScope{Root: defaultRoot}
		}
		return defaultScope
	}
	if targetFiles == nil && targetDirectories == nil {
		ensureDefaultScope()
	}

	scopeByRoot := make(map[tspath.Path]*GitignoreCollectionScope)
	outsideScopes := make([]*GitignoreCollectionScope, 0, len(targetDirectories))
	ensureOutsideScope := func(root string) *GitignoreCollectionScope {
		root = tspath.NormalizePath(root)
		identity := pathID(root)
		if scope := scopeByRoot[identity]; scope != nil {
			return scope
		}
		scope := &GitignoreCollectionScope{Root: root}
		scopeByRoot[identity] = scope
		outsideScopes = append(outsideScopes, scope)
		return scope
	}

	for _, directory := range CoalesceDirectoryIdentities(targetDirectories, fsys) {
		if directoryWithinGitignoreRoot(
			directory.LexicalPath,
			directory.CanonicalPath,
			defaultRoot,
			fsys,
			useCaseSensitive,
		) {
			scope := ensureDefaultScope()
			scope.Directories = appendUniquePath(scope.Directories, directory.LexicalPath, pathID)
			continue
		}
		scope := ensureOutsideScope(directory.LexicalPath)
		scope.Directories = appendUniquePath(scope.Directories, directory.LexicalPath, pathID)
	}

	sort.Slice(outsideScopes, func(i, j int) bool {
		return outsideScopes[i].Root < outsideScopes[j].Root
	})
	for _, file := range targetFiles {
		file = tspath.NormalizePath(file)
		if collectionPathWithinGitignoreRoot(file, defaultRoot, fsys, useCaseSensitive) {
			scope := ensureDefaultScope()
			scope.Files = appendUniquePath(scope.Files, file, pathID)
			continue
		}

		var selected *GitignoreCollectionScope
		for _, scope := range outsideScopes {
			if _, within := RelativePathWithinConfigRoot(file, scope.Root, useCaseSensitive); within {
				selected = scope
				break
			}
		}
		if selected == nil {
			for _, scope := range outsideScopes {
				if collectionPathWithinGitignoreRoot(file, scope.Root, fsys, useCaseSensitive) {
					selected = scope
					break
				}
			}
		}
		if selected == nil {
			continue
		}
		selected.Files = appendUniquePath(selected.Files, file, pathID)
	}

	scopes := make([]GitignoreCollectionScope, 0, len(outsideScopes)+1)
	if defaultScope != nil {
		scopes = append(scopes, *defaultScope)
	}
	for _, scope := range outsideScopes {
		if defaultScope != nil && pathID(scope.Root) == pathID(defaultScope.Root) {
			continue
		}
		scopes = append(scopes, *scope)
	}
	return scopes
}

func directoryWithinGitignoreRoot(
	lexicalDirectory string,
	canonicalDirectory string,
	root string,
	fsys vfs.FS,
	useCaseSensitive bool,
) bool {
	if _, within := RelativePathWithinConfigRoot(
		lexicalDirectory,
		root,
		useCaseSensitive,
	); within {
		return true
	}
	if fsys == nil {
		return false
	}
	physicalRoot := fsys.Realpath(root)
	if physicalRoot == "" {
		return false
	}
	physicalRoot = tspath.NormalizePath(physicalRoot)
	if canonicalDirectory == "" {
		canonicalDirectory = fsys.Realpath(lexicalDirectory)
	}
	if canonicalDirectory == "" {
		return false
	}
	_, within := RelativePathWithinConfigRoot(
		tspath.NormalizePath(canonicalDirectory),
		physicalRoot,
		useCaseSensitive,
	)
	return within
}

func collectionPathWithinGitignoreRoot(
	filePath string,
	root string,
	fsys vfs.FS,
	useCaseSensitive bool,
) bool {
	collectionPath := ResolveGitignoreCollectionPath(filePath, "", root, fsys)
	_, within := RelativePathWithinConfigRoot(collectionPath, root, useCaseSensitive)
	return within
}

func appendUniquePath(
	paths []string,
	path string,
	pathID func(string) tspath.Path,
) []string {
	identity := pathID(path)
	for _, existing := range paths {
		if pathID(existing) == identity {
			return paths
		}
	}
	return append(paths, tspath.NormalizePath(path))
}

// CollectedGitignorePatternsForRoot binds one collected pattern group to the
// exact lexical, physical, and config-projected identities of its Git root.
// Keeping all three explicit lets final matching, pruning, and reopening use
// the same target/root pair without rebasing the config entry itself.
func CollectedGitignorePatternsForRoot(
	patterns []gitignore.Pattern,
	configDirectory string,
	gitignoreRoot string,
	fsys vfs.FS,
) []gitignore.Pattern {
	if len(patterns) == 0 {
		return nil
	}
	root := resolveCollectedGitignoreScope(configDirectory, gitignoreRoot, fsys, false)
	bound := append([]gitignore.Pattern(nil), patterns...)
	for index := range bound {
		bound[index].MatchDirectory = root.matchDirectory
		bound[index].PhysicalMatchDirectory = root.physicalDirectory
		bound[index].LexicalMatchDirectory = root.lexicalDirectory
	}
	return bound
}

func resolveCollectedGitignoreScope(
	configDirectory string,
	gitignoreRoot string,
	fsys vfs.FS,
	caseInsensitive bool,
) collectedGitignoreScope {
	configDirectory = tspath.NormalizePath(configDirectory)
	gitignoreRoot = tspath.NormalizePath(gitignoreRoot)
	matchDirectory, _ := ResolveConfigDirectoryPathSpace(gitignoreRoot, configDirectory, fsys)
	physicalDirectory := gitignoreRoot
	if fsys != nil {
		if realPath := fsys.Realpath(gitignoreRoot); realPath != "" {
			physicalDirectory = tspath.NormalizePath(realPath)
		}
	}
	return collectedGitignoreScope{
		matchDirectory:    matchDirectory,
		physicalDirectory: physicalDirectory,
		lexicalDirectory:  gitignoreRoot,
		caseInsensitive:   caseInsensitive,
	}
}
