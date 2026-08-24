package config

import (
	"errors"
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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
func NewPathSpaceSnapshot(configMap map[string]RslintConfig, fsys vfs.FS) *PathSpaceSnapshot {
	bases := make(map[string]configTargetBase, len(configMap))
	freezeBase := func(directory string) {
		directory = normalizeAuthoredBase(directory)
		baseID := authoredBaseID(directory)
		if _, exists := bases[baseID]; !exists {
			bases[baseID] = freezeConfigTargetBase(directory, fsys)
		}
	}
	for configDir, entries := range configMap {
		freezeBase(configDir)
		for _, entry := range entries {
			freezeBase(configEntryBaseDirectory(entry, configDir))
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
		directory = tspath.RemoveTrailingDirectorySeparators(directory)
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

// ExactPathID returns the case-sensitive normalized identity used for frozen
// paths and canonical-file deduplication.
func ExactPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

// PathsEqual compares two paths using TypeScript's cross-platform path rules.
func PathsEqual(left string, right string, useCaseSensitive bool) bool {
	return tspath.ComparePaths(
		left,
		right,
		tspath.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive},
	) == 0
}

func pathsEqual(left string, right string, useCaseSensitive bool) bool {
	return PathsEqual(left, right, useCaseSensitive)
}
