package target

import (
	"errors"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// OwnerScope records explicit-file provenance supplied by config discovery.
// ExplicitOnly keeps a config loaded solely for literal files out of automatic
// ownership, handoff, and directory-walk decisions.
type OwnerScope struct {
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
}

// NewOwnerIndex snapshots an already-loaded config catalog. It never discovers
// or parses config files.
func NewOwnerIndex(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
) *OwnerIndex {
	pathSpaces := rslintconfig.NewPathSpaceSnapshot(configMap, fsys)
	index, err := newOwnerIndexWithPathSpaces(configMap, fsys, pathSpaces)
	if err != nil {
		panic(err)
	}
	return index
}

// newOwnerIndexWithPathSpaces binds ownership to an existing frozen
// generation. Missing owner bases are invariant errors.
func newOwnerIndexWithPathSpaces(
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
	index, err := newOwnerIndexWithPathSpaces(configMap, fsys, pathSpaces)
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
) rslintconfig.TargetMatcher {
	matcher, err := rslintconfig.NewTargetMatcherWithPathSpaces(
		config,
		configDirectory,
		fsys,
		pathSpaces,
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
	return index.index.nearestConfigForIdentity(identity)
}

// ChildOwnerDirectories returns direct lexical handoff boundaries below owner.
func (index *OwnerIndex) ChildOwnerDirectories(owner string) []string {
	if index == nil || index.index == nil {
		return nil
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
