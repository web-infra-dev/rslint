package config

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// ResolveConfigFilePathSpace returns the path pair used for files and ignores
// matching. User-authored paths keep their lexical relationship to the config
// directory; when both paths are aliases of one physical config tree, the pair
// is instead anchored on that physical tree. Canonical file identity is never
// allowed to rewrite an external lexical file symlink into a different relative
// selector path.
func ResolveConfigFilePathSpace(filePath string, configDir string, fsys vfs.FS) (string, string) {
	return ResolveConfigFilePathSpaceWithCanonical(filePath, "", configDir, fsys)
}

// ResolveConfigFilePathSpaceWithCanonical is ResolveConfigFilePathSpace with an
// optional physical file identity already established by target discovery.
func ResolveConfigFilePathSpaceWithCanonical(filePath string, canonicalPath string, configDir string, fsys vfs.FS) (string, string) {
	return ResolveConfigFilePathSpaceForTarget(PathIdentity{
		Path:          filePath,
		CanonicalPath: canonicalPath,
	}, configDir, fsys)
}

// ResolveConfigFilePathSpaceForTarget projects a frozen file and parent
// identity into one authored config base without resolving the target again.
func ResolveConfigFilePathSpaceForTarget(
	target PathIdentity,
	configDir string,
	fsys vfs.FS,
) (string, string) {
	filePath := target.Path
	filePath = tspath.NormalizePath(filePath)
	configDir = tspath.NormalizePath(configDir)
	if configDir == "" {
		return filePath, ""
	}
	physicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			physicalConfigDir = tspath.NormalizePath(realPath)
		}
	}

	return resolveConfigPathSpaceWithCanonicalParent(
		filePath,
		target.CanonicalPath,
		target.CanonicalParentPath,
		configDir,
		physicalConfigDir,
		verifiedConfigAliasAncestors(configDir, physicalConfigDir, fsys),
		fsys,
		true,
	)
}

// ResolveConfigDirectoryPathSpace is the directory-root counterpart to
// ResolveConfigFilePathSpace. A directory alias that resolves into the
// physical config tree belongs to that tree's matching space; unlike a file
// symlink, the directory root is itself a coordinate system and must not retain
// a competing lexical base.
func ResolveConfigDirectoryPathSpace(directory string, configDir string, fsys vfs.FS) (string, string) {
	directory = tspath.NormalizePath(directory)
	configDir = tspath.NormalizePath(configDir)
	if configDir == "" {
		return directory, ""
	}
	physicalConfigDir := configDir
	canonicalDirectory := directory
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			physicalConfigDir = tspath.NormalizePath(realPath)
		}
		if realPath := fsys.Realpath(directory); realPath != "" {
			canonicalDirectory = tspath.NormalizePath(realPath)
		}
	}
	// A requested directory that physically contains the config root selects
	// that complete config-owned tree. Retain the physical containment pair so
	// directory discovery and Git projection can recognize it even when the
	// config was reached through a symlink in an unrelated lexical parent.
	if _, containsConfigRoot := RelativePathWithinConfigRoot(
		physicalConfigDir,
		canonicalDirectory,
		true,
	); containsConfigRoot {
		return canonicalDirectory, physicalConfigDir
	}
	return resolveConfigPathSpace(
		directory,
		canonicalDirectory,
		configDir,
		physicalConfigDir,
		verifiedConfigAliasAncestors(configDir, physicalConfigDir, fsys),
		fsys,
		false,
	)
}

// verifiedConfigAliasAncestors returns physical ancestor roots that preserve
// the config directory's complete lexical suffix. They let sibling paths use
// one coordinate system when an ancestor (for example macOS /tmp) is aliased,
// without treating a symlink on the config directory alone as authority over
// unrelated physical siblings. Filesystem roots are deliberately excluded.
func verifiedConfigAliasAncestors(configDir string, physicalConfigDir string, fsys vfs.FS) []string {
	if fsys == nil {
		return nil
	}
	var ancestors []string
	for current := configDir; current != ""; {
		parent := tspath.GetDirectoryPath(current)
		if parent == current {
			break
		}
		physicalCurrent := ""
		if current == configDir {
			physicalCurrent = physicalConfigDir
		} else {
			physicalCurrent = fsys.Realpath(current)
		}
		if physicalCurrent != "" {
			physicalCurrent = tspath.NormalizePath(physicalCurrent)
			if !pathsEqual(current, physicalCurrent, true) {
				relativeConfig, within := RelativePathWithinConfigRoot(configDir, current, true)
				preservesSuffix := within && pathsEqual(
					tspath.ResolvePath(physicalCurrent, relativeConfig),
					physicalConfigDir,
					true,
				)
				if preservesSuffix {
					duplicate := false
					for _, ancestor := range ancestors {
						if pathsEqual(ancestor, physicalCurrent, true) {
							duplicate = true
							break
						}
					}
					if !duplicate {
						ancestors = append(ancestors, physicalCurrent)
					}
				}
			}
		}
		current = parent
	}
	return ancestors
}

func pathUsesVerifiedConfigAlias(path string, physicalAncestors []string) bool {
	for _, ancestor := range physicalAncestors {
		if _, within := RelativePathWithinConfigRoot(path, ancestor, true); within {
			return true
		}
	}
	return false
}

func resolveConfigPathSpace(
	filePath string,
	canonicalPath string,
	configDir string,
	physicalConfigDir string,
	physicalAliasAncestors []string,
	fsys vfs.FS,
	preserveDistinctFileSymlink bool,
) (string, string) {
	return resolveConfigPathSpaceWithCanonicalParent(
		filePath,
		canonicalPath,
		"",
		configDir,
		physicalConfigDir,
		physicalAliasAncestors,
		fsys,
		preserveDistinctFileSymlink,
	)
}

func resolveConfigPathSpaceWithCanonicalParent(
	filePath string,
	canonicalPath string,
	canonicalParentPath string,
	configDir string,
	physicalConfigDir string,
	physicalAliasAncestors []string,
	fsys vfs.FS,
	preserveDistinctFileSymlink bool,
) (string, string) {
	// The ordinary path is already in the config's exact lexical/physical
	// coordinate system. Returning it directly avoids decomposing and rebuilding
	// the same absolute path for every discovered file.
	if configDir == physicalConfigDir && isPathWithinNormalizedRoot(filePath, configDir) {
		return filePath, configDir
	}
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, true); ok {
		return tspath.ResolvePath(physicalConfigDir, relative), physicalConfigDir
	}
	// Match paths may already have been projected onto the physical config
	// directory by target binding. Preserve their relative spelling so applying
	// this resolver twice cannot erase a config-observable directory alias.
	if relative, ok := RelativePathWithinConfigRoot(filePath, physicalConfigDir, true); ok {
		return tspath.ResolvePath(physicalConfigDir, relative), physicalConfigDir
	}
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, false); ok && fsys != nil {
		if canonicalParentPath != "" {
			expectedPhysicalFile := expectedPhysicalFileFromLexicalPathWithCanonicalParent(
				filePath,
				canonicalParentPath,
				fsys,
			)
			if _, within := RelativePathWithinConfigRoot(
				expectedPhysicalFile,
				physicalConfigDir,
				true,
			); within {
				return tspath.ResolvePath(physicalConfigDir, relative), physicalConfigDir
			}
		} else {
			aliasRoot := filePath
			for remaining := relative; remaining != ""; remaining = tspath.GetDirectoryPath(remaining) {
				aliasRoot = tspath.GetDirectoryPath(aliasRoot)
			}
			if realRoot := fsys.Realpath(aliasRoot); realRoot != "" &&
				tspath.ComparePaths(
					tspath.NormalizePath(realRoot),
					physicalConfigDir,
					tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true},
				) == 0 {
				return tspath.ResolvePath(physicalConfigDir, relative), physicalConfigDir
			}
		}
	}

	physicalFilePath := ""
	if canonicalPath != "" {
		physicalFilePath = tspath.NormalizePath(canonicalPath)
	}
	if physicalFilePath == "" {
		physicalFilePath = filePath
		if fsys != nil {
			if realPath := fsys.Realpath(filePath); realPath != "" {
				physicalFilePath = tspath.NormalizePath(realPath)
			}
		}
	}
	if _, caseAlias := RelativePathWithinConfigRoot(filePath, configDir, false); caseAlias {
		// The case-insensitive spelling looked like the config tree, but native
		// realpath verification above proved it was a distinct physical root.
		// Preserve that distinction rather than manufacturing a mixed-case path.
		return physicalFilePath, physicalConfigDir
	}
	// Compare the physical file with its physical parent plus lexical basename.
	// A difference at that final component identifies a file symlink (or an
	// explicit canonical identity hint) whose user-visible selector path must be
	// retained. Differences introduced only by an ancestor directory alias,
	// native casing, or a platform path such as macOS /var -> /private/var still
	// share the physical config space.
	expectedPhysicalFile := physicalFilePath
	if preserveDistinctFileSymlink {
		expectedPhysicalFile = expectedPhysicalFileFromLexicalPathWithCanonicalParent(
			filePath,
			canonicalParentPath,
			fsys,
		)
	}
	targetHasDistinctLeafIdentity := preserveDistinctFileSymlink &&
		hasDistinctFileIdentityWithCanonicalParent(
			filePath,
			physicalFilePath,
			canonicalParentPath,
			fsys,
		)
	// A verified ancestor alias establishes the coordinate system independently
	// of the leaf's identity. Keep the lexical basename on the physical parent
	// so a file symlink still matches the path the user selected, while its
	// canonical destination remains only an identity for deduplication.
	if pathUsesVerifiedConfigAlias(expectedPhysicalFile, physicalAliasAncestors) {
		return expectedPhysicalFile, physicalConfigDir
	}
	if !targetHasDistinctLeafIdentity {
		if relative, ok := RelativePathWithinConfigRoot(physicalFilePath, physicalConfigDir, true); ok {
			return tspath.ResolvePath(physicalConfigDir, relative), physicalConfigDir
		}
	}

	// The target is outside both identities of the config tree. This is valid
	// for an invocation-wide external config whose files/ignores selectors use
	// ../ paths, and it includes an external lexical symlink whose physical file
	// happens to live inside the config tree. Keep the complete lexical pair:
	// projecting only the target onto physicalConfigDir would normalize away
	// meaningful ../ segments whenever the lexical and physical config roots
	// have different parent layouts.
	return filePath, configDir
}

func expectedPhysicalFileFromLexicalPathWithCanonicalParent(
	filePath string,
	canonicalParentPath string,
	fsys vfs.FS,
) string {
	expectedPhysicalFile := filePath
	physicalParent := canonicalParentPath
	if physicalParent == "" && fsys != nil {
		physicalParent = fsys.Realpath(tspath.GetDirectoryPath(filePath))
	}
	if physicalParent == "" {
		return expectedPhysicalFile
	}
	return tspath.ResolvePath(
		tspath.NormalizePath(physicalParent),
		tspath.GetBaseFileName(filePath),
	)
}

// hasDistinctFileIdentityWithCanonicalParent distinguishes a leaf symlink
// (whose caller-visible name remains a matcher coordinate) from a
// directory/root alias (whose canonical descendant path may be used as a
// fallback coordinate).
func hasDistinctFileIdentityWithCanonicalParent(
	filePath string,
	canonicalPath string,
	canonicalParentPath string,
	fsys vfs.FS,
) bool {
	return tspath.ComparePaths(
		expectedPhysicalFileFromLexicalPathWithCanonicalParent(
			filePath,
			canonicalParentPath,
			fsys,
		),
		canonicalPath,
		tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true},
	) != 0
}

// RelativePathWithinConfigRoot returns filePath relative to configDir when it
// is inside the config's lexical path space.
func RelativePathWithinConfigRoot(filePath string, configDir string, useCaseSensitive bool) (string, bool) {
	options := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	if tspath.ComparePaths(filePath, configDir, options) == 0 {
		return "", true
	}
	if !tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive) {
		return "", false
	}
	return tspath.GetRelativePathFromDirectory(configDir, filePath, options), true
}

// isPathWithinNormalizedRoot is the allocation-free containment check for
// already-normalized, case-sensitive paths used by frozen target identities.
func isPathWithinNormalizedRoot(filePath string, root string) bool {
	return filePath == root || tspath.StartsWithDirectory(filePath, root, true)
}
