package config

import (
	"errors"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// PathSpaceSnapshot freezes every authored config base used by one evaluation
// generation. The internal map is never exposed or mutated after construction.
// An empty snapshot is valid; a nil snapshot or a missing requested base is an
// invariant error for snapshot-aware consumers.
type PathSpaceSnapshot struct {
	bases map[string]configTargetBase
}

// NewPathSpaceSnapshot freezes owning config directories and any distinct
// authored bases carried by composed entries.
func NewPathSpaceSnapshot(configMap map[string]RslintConfig, fsys vfs.FS, directories ...DirectoryIdentity) *PathSpaceSnapshot {
	return (*PathSpaceSnapshot)(nil).WithConfigBases(configMap, fsys, directories...)
}

// WithConfigBases extends a snapshot without re-observing an existing base.
// The supplied directory identities anchor newly loaded config owners and the
// request roots through which their authored bases were reached. These are
// directory identities; individual file hints never establish directory bases.
func (snapshot *PathSpaceSnapshot) WithConfigBases(configMap map[string]RslintConfig, fsys vfs.FS, directories ...DirectoryIdentity) *PathSpaceSnapshot {
	capacity := len(configMap) + len(directories)
	if snapshot != nil {
		capacity += len(snapshot.bases)
	}
	bases := make(map[string]configTargetBase, capacity)
	if snapshot != nil {
		for id, base := range snapshot.bases {
			bases[id] = base
		}
	}
	var seeded []string
	for _, identity := range directories {
		directory := normalizeAuthoredBase(identity.LexicalPath)
		id := authoredBaseID(directory)
		if _, exists := bases[id]; exists {
			continue
		}
		physical := identity.CanonicalPath
		if physical == "" {
			physical = directory
		}
		bases[id] = configTargetBase{directory: directory, physicalDirectory: normalizeAuthoredBase(physical)}
		seeded = append(seeded, directory)
	}
	physicalDirectory := func(directory string) string {
		directory = normalizeAuthoredBase(directory)
		for current := directory; current != ""; {
			if base, exists := bases[authoredBaseID(current)]; exists {
				if current == directory {
					return base.physicalDirectory
				}
				relative, within := RelativePathWithinConfigRoot(directory, current, true)
				if within {
					return canonicalDirectoryPathOrSelf(tspath.ResolvePath(base.physicalDirectory, relative), fsys)
				}
			}
			parent := tspath.GetDirectoryPath(current)
			if parent == current {
				break
			}
			current = parent
		}
		return canonicalDirectoryPathOrSelf(directory, fsys)
	}
	freezeAncestors := func(directory string) {
		id := authoredBaseID(directory)
		base := bases[id]
		base.physicalAliasAncestors = verifiedConfigAliasAncestorsWithLookup(directory, base.physicalDirectory, fsys, physicalDirectory)
		bases[id] = base
	}
	for _, directory := range seeded {
		freezeAncestors(directory)
	}
	freezeBase := func(directory string) {
		directory = normalizeAuthoredBase(directory)
		baseID := authoredBaseID(directory)
		if _, exists := bases[baseID]; !exists {
			bases[baseID] = configTargetBase{directory: directory, physicalDirectory: physicalDirectory(directory)}
			freezeAncestors(directory)
		}
	}
	for configDir, entries := range configMap {
		freezeBase(configDir)
		for _, entry := range entries {
			origin := configEntryPathOrigin(entry, configDir)
			freezeBase(origin.directory)
			if origin.basePathScoped {
				freezeBase(origin.configArrayBase)
			}
		}
	}
	return &PathSpaceSnapshot{bases: bases}
}

func (snapshot *PathSpaceSnapshot) base(directory string) (configTargetBase, bool) {
	if snapshot == nil {
		return configTargetBase{}, false
	}
	base, ok := snapshot.bases[authoredBaseID(directory)]
	return base, ok
}

func normalizeAuthoredBase(directory string) string {
	directory = tspath.NormalizePath(directory)
	if len(directory) > tspath.GetRootLength(directory) {
		directory = pathutil.RemoveTrailingDirectorySeparators(directory)
	}
	return directory
}

func authoredBaseID(directory string) string {
	return ExactPathID(normalizeAuthoredBase(directory))
}

func (snapshot *PathSpaceSnapshot) requireBase(directory string) (configTargetBase, error) {
	if snapshot == nil {
		return configTargetBase{}, errors.New("path-space snapshot is required")
	}
	base, ok := snapshot.base(directory)
	if !ok {
		return configTargetBase{}, fmt.Errorf("path-space snapshot does not contain authored base %q", tspath.NormalizePath(directory))
	}
	return base, nil
}

// PhysicalDirectory returns the frozen physical identity of an authored base.
func (snapshot *PathSpaceSnapshot) PhysicalDirectory(directory string) (string, bool) {
	base, ok := snapshot.base(directory)
	if !ok {
		return "", false
	}
	return base.physicalDirectory, true
}

// DirectoryIdentities returns copies of the frozen lexical and physical
// directory values without exposing matching or alias metadata.
func (snapshot *PathSpaceSnapshot) DirectoryIdentities() []DirectoryIdentity {
	if snapshot == nil {
		return nil
	}
	directories := make([]DirectoryIdentity, 0, len(snapshot.bases))
	for _, base := range snapshot.bases {
		directories = append(directories, DirectoryIdentity{LexicalPath: base.directory, CanonicalPath: base.physicalDirectory})
	}
	return directories
}

// SameDirectory reports whether two owner spellings are the same frozen
// config root. Case-insensitive lexical equality is accepted only after both
// spellings are verified to share one physical identity.
func (snapshot *PathSpaceSnapshot) SameDirectory(left string, right string, useCaseSensitive bool) bool {
	left = tspath.NormalizePath(left)
	right = tspath.NormalizePath(right)
	if ExactPathID(left) == ExactPathID(right) {
		return true
	}
	if useCaseSensitive || !PathsEqual(left, right, false) {
		return false
	}
	leftBase, leftFrozen := snapshot.base(left)
	rightBase, rightFrozen := snapshot.base(right)
	return leftFrozen && rightFrozen &&
		ExactPathID(leftBase.physicalDirectory) == ExactPathID(rightBase.physicalDirectory)
}

// ResolvePath projects one lexical file or directory path into an authored
// config base. The base is frozen by this snapshot; resolving the supplied path
// may still consult fsys when its caller has not already established a physical
// identity.
func (snapshot *PathSpaceSnapshot) ResolvePath(
	path string,
	configDir string,
	fsys vfs.FS,
) (string, string, error) {
	base, err := snapshot.requireBase(configDir)
	if err != nil {
		return "", "", err
	}
	path = tspath.NormalizePath(path)
	matchPath, matchDir := resolveConfigPathSpaceWithCanonicalParent(
		path,
		"",
		"",
		base.directory,
		base.physicalDirectory,
		base.physicalAliasAncestors,
		fsys,
		true,
	)
	return matchPath, matchDir, nil
}

// ExactPathID returns the normalized identity used for frozen paths and
// canonical-file deduplication. Windows drive letters are case-insensitive;
// directory and file names retain their exact casing.
func ExactPathID(filePath string) string {
	return utils.ExactPathID(filePath)
}

// PathsEqual compares two paths using TypeScript's cross-platform path rules.
func PathsEqual(left string, right string, useCaseSensitive bool) bool {
	return pathutil.ComparePaths(
		left,
		right,
		pathutil.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive},
	) == 0
}

func pathsEqual(left string, right string, useCaseSensitive bool) bool {
	return PathsEqual(left, right, useCaseSensitive)
}
