package discovery

import (
	"fmt"
	"sort"

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

func (draft *catalogDraft) addDirectory(ownerDirectory string, directory rslintconfig.DirectoryIdentity, recursive bool) {
	scope := draft.scopes[ownerDirectory]
	scope.Directories = appendUniqueSortedDirectory(scope.Directories, directory)
	if recursive {
		scope.RecursiveDirectories = appendUniqueSortedDirectory(scope.RecursiveDirectories, directory)
	}
	draft.scopes[ownerDirectory] = scope
}

func appendUniqueSortedDirectory(directories []rslintconfig.DirectoryIdentity, directory rslintconfig.DirectoryIdentity) []rslintconfig.DirectoryIdentity {
	for _, existing := range directories {
		if rslintconfig.ExactPathID(existing.LexicalPath) == rslintconfig.ExactPathID(directory.LexicalPath) {
			return directories
		}
	}
	directories = append(directories, directory)
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].LexicalPath < directories[j].LexicalPath
	})
	return directories
}

// finalize validates physical owner identity, publishes explicit-only
// provenance, and returns stable
// effective IDs. Node activation happens only after this method succeeds.
func (draft *catalogDraft) finalize(
	physicalDirectory func(string) string,
	fsys vfs.FS,
) (
	map[string]rslintconfig.RslintConfig,
	map[string]target.OwnerScope,
	[]string,
	error,
) {
	if err := draft.validateDirectoryIdentities(physicalDirectory, fsys); err != nil {
		return nil, nil, nil, err
	}
	for directory, source := range draft.selectedSources {
		scope := draft.scopes[directory]
		scope.ExplicitOnly = source.explicitOnly
		if scope.Directories == nil {
			scope.Directories = []rslintconfig.DirectoryIdentity{}
		}
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
// activates the effective candidate set. Native spellings already share an owner.
func (draft *catalogDraft) validateDirectoryIdentities(physicalDirectory func(string) string, fsys vfs.FS) error {
	if len(draft.configs) < 2 {
		return nil
	}
	directories := make([]string, 0, len(draft.configs))
	for directory := range draft.configs {
		directories = append(directories, tspath.NormalizePath(directory))
	}
	sort.Strings(directories)

	ownersByPhysical := make(map[string][]rslintconfig.DirectoryIdentity, len(directories))
	for _, directory := range directories {
		physicalPath := physicalDirectory(directory)
		key := configSpellingKey(physicalPath)
		for _, existing := range ownersByPhysical[key] {
			if samePhysicalPath(fsys, existing.CanonicalPath, physicalPath) {
				return fmt.Errorf( //nolint:staticcheck // Preserve the established user-facing JS API error contract.
					"Config directories %q and %q resolve to the same filesystem location %q",
					existing.LexicalPath, directory, physicalPath,
				)
			}
		}
		ownersByPhysical[key] = append(ownersByPhysical[key], rslintconfig.DirectoryIdentity{
			LexicalPath: directory, CanonicalPath: physicalPath,
		})
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
