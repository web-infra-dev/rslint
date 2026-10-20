package target

import (
	"errors"
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// OwnerScope records admitted directory entries and explicit-file provenance
// supplied by config discovery. Directories is nil when no scope was supplied;
// a non-nil empty slice means that only the scoped files may be selected.
// ExplicitOnly keeps a config loaded solely for literal files out of automatic
// ownership, handoff, and directory-walk decisions.
type OwnerScope struct {
	Directories   []rslintconfig.DirectoryIdentity
	ExplicitFiles []string
	ExplicitOnly  bool
}

func configMapForAutomaticTargets(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]OwnerScope,
) map[string]rslintconfig.RslintConfig {
	for configDir := range configMap {
		if !scopes[configDir].ExplicitOnly {
			continue
		}
		automaticConfigMap := make(map[string]rslintconfig.RslintConfig, len(configMap)-1)
		for candidateDir, candidateConfig := range configMap {
			if !scopes[candidateDir].ExplicitOnly {
				automaticConfigMap[candidateDir] = candidateConfig
			}
		}
		return automaticConfigMap
	}
	return configMap
}

// OwnerIndex is a read-only config-owner catalog for one path-space
// generation. It resolves ownership only; config matching and rule expansion
// remain config responsibilities.
type OwnerIndex struct {
	index      *configDirectoryIndex
	pathSpaces *rslintconfig.PathSpaceSnapshot

	ownersByEntry   map[string]string
	childrenByOwner map[string][]string
}

// newAutomaticOwnerIndex uses the same admitted entries for ownership and walking.
// An ancestor config loaded for a requested subdirectory governs that entry,
// not files that another route still assigns to the enclosing owner.
func newAutomaticOwnerIndex(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]OwnerScope,
	fsys vfs.FS,
	spaces *rslintconfig.PathSpaceSnapshot,
) *OwnerIndex {
	entries := make(map[string]rslintconfig.RslintConfig)
	owners := make(map[string]string)
	for owner, config := range configMapForAutomaticTargets(configMap, scopes) {
		directories := scopes[owner].Directories
		if directories == nil {
			entries[owner], owners[owner] = config, owner
			continue
		}
		for _, directory := range directories {
			entries[directory.LexicalPath], owners[directory.LexicalPath] = config, owner
		}
	}
	index := mustOwnerIndexWithPathSpaces(entries, fsys, spaces)
	index.ownersByEntry = owners
	index.childrenByOwner = make(map[string][]string)
	for entry, children := range index.index.childrenByKey {
		owner := owners[entry]
		for _, child := range children {
			if owners[child] != owner {
				index.childrenByOwner[owner] = append(index.childrenByOwner[owner], child)
			}
		}
	}
	for owner := range index.childrenByOwner {
		sort.Strings(index.childrenByOwner[owner])
	}
	return index
}

// NewOwnerIndex snapshots an already-loaded config catalog. It never discovers
// or parses config files.
func NewOwnerIndex(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
) *OwnerIndex {
	pathSpaces := rslintconfig.NewPathSpaceSnapshot(configMap, fsys)
	index, err := NewOwnerIndexWithPathSpaces(configMap, fsys, pathSpaces)
	if err != nil {
		panic(err)
	}
	return index
}

// NewOwnerIndexWithPathSpaces binds ownership to an existing frozen
// generation. Missing owner bases are invariant errors.
func NewOwnerIndexWithPathSpaces(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) (*OwnerIndex, error) {
	if pathSpaces == nil {
		return nil, errors.New("path-space snapshot is required")
	}
	for configDir := range configMap {
		if _, ok := pathSpaces.PhysicalDirectory(configDir); !ok {
			return nil, fmt.Errorf("path-space snapshot does not contain config owner %q", configDir)
		}
	}
	return &OwnerIndex{
		index:      newConfigDirectoryIndexWithPathSpaces(configMap, fsys, pathSpaces),
		pathSpaces: pathSpaces,
	}, nil
}

func mustOwnerIndexWithPathSpaces(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) *OwnerIndex {
	index, err := NewOwnerIndexWithPathSpaces(configMap, fsys, pathSpaces)
	if err != nil {
		panic(err)
	}
	return index
}

func mustTargetMatcher(
	config rslintconfig.RslintConfig,
	configDirectory string,
	fsys vfs.FS,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
	gitPolicy *rslintconfig.GitPolicy,
) rslintconfig.TargetMatcher {
	matcher, err := rslintconfig.NewTargetMatcherWithPathSpaces(
		config,
		configDirectory,
		fsys,
		pathSpaces,
		gitPolicy,
	)
	if err != nil {
		panic(err)
	}
	return matcher
}

// Resolve resolves the nearest owner from an already-frozen identity without
// re-resolving the file itself.
func (index *OwnerIndex) Resolve(identity rslintconfig.PathIdentity) (string, bool) {
	if index == nil || index.index == nil {
		return "", false
	}
	directory, found := index.index.nearestConfigForIdentity(identity)
	if found && index.ownersByEntry != nil {
		return index.ownersByEntry[directory], true
	}
	return directory, found
}

// ChildOwnerDirectories returns direct lexical handoff boundaries below owner.
func (index *OwnerIndex) ChildOwnerDirectories(owner string) []string {
	if index == nil || index.index == nil {
		return nil
	}
	if index.ownersByEntry != nil {
		return append([]string(nil), index.childrenByOwner[owner]...)
	}
	return append([]string(nil), index.index.childConfigDirs(owner)...)
}

// PathSpaces returns the read-only path-space generation shared by this index.
func (index *OwnerIndex) PathSpaces() *rslintconfig.PathSpaceSnapshot {
	if index == nil {
		return nil
	}
	return index.pathSpaces
}
