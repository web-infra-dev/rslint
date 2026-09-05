package discovery

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

type selectedConfigSource struct {
	candidateID   string
	candidatePath string
	explicitOnly  bool
}

// catalogDraft is the sole writer for effective configs, selected sources,
// and target-owner scopes during one discovery transaction.
type catalogDraft struct {
	configs         map[string]rslintconfig.RslintConfig
	selectedSources map[string]selectedConfigSource
	scopes          map[string]target.OwnerScope
}

func newCatalogDraft() catalogDraft {
	return catalogDraft{
		configs:         make(map[string]rslintconfig.RslintConfig),
		selectedSources: make(map[string]selectedConfigSource),
		scopes:          make(map[string]target.OwnerScope),
	}
}

func (draft *catalogDraft) hasConfigs() bool {
	return len(draft.configs) > 0
}

// adoptCandidate selects one successfully loaded config for an owner. Literal
// explicit-only routes may be replaced by an automatic route, but never the
// reverse.
func (draft *catalogDraft) adoptCandidate(state *configLoadState, explicitOnly bool) error {
	if state == nil || state.failure != nil {
		return nil
	}
	directory := state.candidate.directory
	source, exists := draft.selectedSources[directory]
	if !exists {
		draft.installCandidate(state, explicitOnly)
		return nil
	}
	if source.candidateID == state.id {
		source.explicitOnly = source.explicitOnly && explicitOnly
		draft.selectedSources[directory] = source
		return nil
	}
	if source.explicitOnly && !explicitOnly {
		// A literal target may bypass its parent's authored ignore and discover a
		// different filename in this directory. Once an automatic route reaches
		// the directory, that route defines its shared config boundary.
		draft.installCandidate(state, false)
		return nil
	}
	if !source.explicitOnly && explicitOnly {
		return nil
	}
	return fmt.Errorf(
		"ambiguous config candidates %q and %q for directory %q",
		source.candidatePath,
		state.candidate.path,
		directory,
	)
}

func (draft *catalogDraft) installCandidate(state *configLoadState, explicitOnly bool) {
	directory := state.candidate.directory
	draft.configs[directory] = append(rslintconfig.RslintConfig(nil), state.entries...)
	draft.selectedSources[directory] = selectedConfigSource{
		candidateID:   state.id,
		candidatePath: state.candidate.path,
		explicitOnly:  explicitOnly,
	}
}

func (draft *catalogDraft) addExplicitFile(ownerDirectory string, filePath string) {
	scope := draft.scopes[ownerDirectory]
	scope.ExplicitFiles = appendUniqueSortedPath(scope.ExplicitFiles, filePath)
	draft.scopes[ownerDirectory] = scope
}

// finalize validates physical owner identity, writes each Git-materialized
// config exactly once, publishes explicit-only provenance, and returns stable
// effective IDs. Node activation happens only after this method succeeds.
func (draft *catalogDraft) finalize(
	fsys vfs.FS,
	projection *gitProjection,
) (
	map[string]rslintconfig.RslintConfig,
	map[string]target.OwnerScope,
	[]string,
	error,
) {
	if err := draft.validateDirectoryIdentities(fsys); err != nil {
		return nil, nil, nil, err
	}
	for ownerDirectory, entries := range draft.configs {
		draft.configs[ownerDirectory] = projection.materialize(ownerDirectory, entries)
	}
	for directory, source := range draft.selectedSources {
		scope := draft.scopes[directory]
		scope.ExplicitOnly = source.explicitOnly
		draft.scopes[directory] = scope
	}
	effectiveIDs := make([]string, 0, len(draft.selectedSources))
	for _, source := range draft.selectedSources {
		effectiveIDs = append(effectiveIDs, source.candidateID)
	}
	sort.Strings(effectiveIDs)
	return draft.configs, draft.scopes, effectiveIDs, nil
}

// validateDirectoryIdentities rejects ambiguous ownership before Node
// activates the effective candidate set. Alternate native casing is safe only
// after Realpath verifies one exact physical directory.
func (draft *catalogDraft) validateDirectoryIdentities(fsys vfs.FS) error {
	directories := make([]string, 0, len(draft.configs))
	for directory := range draft.configs {
		directories = append(directories, tspath.NormalizePath(directory))
	}
	sort.Strings(directories)

	lexicalByPhysical := make(map[tspath.Path]string, len(directories))
	for _, directory := range directories {
		physicalDirectory := directory
		if realPath := fsys.Realpath(directory); realPath != "" {
			physicalDirectory = tspath.NormalizePath(realPath)
		}
		physicalID := tspath.ToPath(physicalDirectory, "", true)
		existing, collision := lexicalByPhysical[physicalID]
		if !collision {
			lexicalByPhysical[physicalID] = directory
			continue
		}
		if !fsys.UseCaseSensitiveFileNames() && strings.EqualFold(existing, directory) {
			continue
		}
		return fmt.Errorf( //nolint:staticcheck // Preserve the established user-facing JS API error contract.
			"Config directories %q and %q resolve to the same filesystem location %q",
			existing,
			directory,
			physicalDirectory,
		)
	}
	return nil
}

func appendUniqueSortedPath(paths []string, path string) []string {
	for _, existing := range paths {
		if existing == path {
			return paths
		}
	}
	paths = append(paths, path)
	sort.Strings(paths)
	return paths
}
