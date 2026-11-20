package discovery

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type configCandidate struct {
	path      string
	directory string
	// physicalDirectory anchors module IO while path and directory preserve
	// the caller's lexical candidate and config-owner identities.
	physicalDirectory string
	// configArrayBase is set only when ConfigArray A differs from the config
	// module directory, as for an explicit config. Empty means directory.
	configArrayBase string
}

func (candidate configCandidate) modulePath() string {
	return tspath.CombinePaths(candidate.physicalDirectory, tspath.GetBaseFileName(candidate.path))
}

// configLoadState binds one candidate to immutable authored entries and their
// ignore matcher. Git policy remains independent from these entries.
type configLoadState struct {
	id            string
	candidate     configCandidate
	entries       rslintconfig.RslintConfig
	ignoreMatcher rslintconfig.GlobalIgnoreMatcher
	failure       *ConfigModuleError
}

// moduleLoadCoordinator owns the Node-facing half of one discovery
// transaction: candidate identity, ordered batch IDs, immutable load results,
// failures, and final effective-config activation.
type moduleLoadCoordinator struct {
	ctx            context.Context
	fs             vfs.FS
	loader         ConfigModuleLoader
	transactionID  string
	fresh          bool
	singleThreaded bool
	pathSpaces     *rslintconfig.PathSpaceSnapshot

	statesByPath     map[string]*configLoadState
	ownerByDirectory map[rslintconfig.DirectoryIdentity]string
	ownersByPhysical map[string][]rslintconfig.DirectoryIdentity
	failureByPath    map[string]ConfigFailure
	nextCandidateID  int
	configsRequested int
	configsLoaded    int
}

func newModuleLoadCoordinator(
	ctx context.Context,
	fsys vfs.FS,
	loader ConfigModuleLoader,
	transactionID string,
	fresh bool,
	singleThreaded bool,
) moduleLoadCoordinator {
	return moduleLoadCoordinator{
		ctx:              ctx,
		fs:               fsys,
		loader:           loader,
		transactionID:    transactionID,
		fresh:            fresh,
		singleThreaded:   singleThreaded,
		statesByPath:     make(map[string]*configLoadState),
		ownerByDirectory: make(map[rslintconfig.DirectoryIdentity]string),
		ownersByPhysical: make(map[string][]rslintconfig.DirectoryIdentity),
		failureByPath:    make(map[string]ConfigFailure),
	}
}

func (coordinator *moduleLoadCoordinator) state(candidatePath string) *configLoadState {
	return coordinator.statesByPath[candidatePath]
}

func (coordinator *moduleLoadCoordinator) loadCandidates(rawCandidates []configCandidate) error {
	type candidateGroup struct {
		candidate configCandidate
		identity  string
		aliases   map[string]struct{}
	}
	// The first spelling in a frontier supplies the representative owner and module.
	sort.Slice(rawCandidates, func(i, j int) bool { return rawCandidates[i].path < rawCandidates[j].path })
	var groups []*candidateGroup
	pending := make(map[string]*candidateGroup, len(rawCandidates))
	for _, candidate := range rawCandidates {
		candidate.path = tspath.NormalizePath(candidate.path)
		candidate.directory = tspath.NormalizePath(candidate.directory)
		if candidate.physicalDirectory == "" {
			candidate.physicalDirectory = candidate.directory
			if physical := coordinator.fs.Realpath(candidate.directory); physical != "" {
				candidate.physicalDirectory = tspath.NormalizePath(physical)
			}
		}
		candidate.directory = coordinator.configOwner(candidate)
		identity := rslintconfig.ExactPathID(tspath.CombinePaths(candidate.directory, tspath.GetBaseFileName(candidate.path)))
		if state := coordinator.statesByPath[identity]; state != nil {
			if err := coordinator.validateCandidate(state.candidate, candidate); err != nil {
				return err
			}
			coordinator.statesByPath[candidate.path] = state
			continue
		}
		group := pending[identity]
		if group == nil {
			group = &candidateGroup{candidate: candidate, identity: identity, aliases: make(map[string]struct{})}
			pending[identity] = group
			groups = append(groups, group)
		} else if err := coordinator.validateCandidate(group.candidate, candidate); err != nil {
			return err
		}
		group.aliases[candidate.path] = struct{}{}
	}
	if len(groups) == 0 {
		return nil
	}
	if coordinator.loader == nil {
		return errors.New("javascript config candidates require a module loader")
	}

	loadMode := ConfigModuleLoadCached
	if coordinator.fresh {
		loadMode = ConfigModuleLoadFresh
	}
	request := ConfigLoadBatchRequest{
		ProtocolVersion: ConfigDiscoveryProtocolVersion,
		TransactionID:   coordinator.transactionID,
		LoadMode:        loadMode,
		SingleThreaded:  coordinator.singleThreaded,
	}
	request.Candidates = make([]ConfigLoadCandidate, 0, len(groups))
	for _, group := range groups {
		candidate := group.candidate
		coordinator.nextCandidateID++
		id := fmt.Sprintf("config-%06d", coordinator.nextCandidateID)
		request.Candidates = append(request.Candidates, ConfigLoadCandidate{
			ID:              id,
			ConfigPath:      candidate.modulePath(),
			ConfigDirectory: candidate.directory,
		})
	}
	coordinator.configsRequested += len(request.Candidates)
	response, err := coordinator.loader.LoadConfigs(coordinator.ctx, request)
	if err != nil {
		return err
	}
	results, err := validateConfigLoadBatch(request, response)
	if err != nil {
		return err
	}

	configsForBases := make(map[string]rslintconfig.RslintConfig, len(request.Candidates))
	directoryIdentities := make([]rslintconfig.DirectoryIdentity, 0, len(request.Candidates))
	loadedStates := make([]*configLoadState, 0, len(request.Candidates))
	for index, wireCandidate := range request.Candidates {
		group := groups[index]
		candidate := group.candidate
		result := results[wireCandidate.ID]
		state := &configLoadState{
			id:        wireCandidate.ID,
			candidate: candidate,
		}
		// Failed candidates remain known owner boundaries for transactional LSP
		// fallback, so their directory identity is frozen as well.
		directoryIdentities = append(directoryIdentities, rslintconfig.DirectoryIdentity{
			LexicalPath: candidate.directory, CanonicalPath: candidate.physicalDirectory,
		})
		if result.Status == "failed" {
			failure := *result.Error
			state.failure = &failure
		} else if err := rslintconfig.ValidateConfig(result.Entries); err != nil {
			state.failure = &ConfigModuleError{Code: "invalid", Message: err.Error()}
		} else {
			entries := append(rslintconfig.RslintConfig(nil), result.Entries...)
			configArrayBase := candidate.configArrayBase
			if configArrayBase == "" {
				configArrayBase = candidate.directory
			}
			state.entries = rslintconfig.ConfigWithResolvedBasePaths(
				entries,
				configArrayBase,
			)
			configsForBases[candidate.directory] = append(configsForBases[candidate.directory], state.entries...)
			loadedStates = append(loadedStates, state)
			coordinator.configsLoaded++
		}
		coordinator.statesByPath[group.identity] = state
		for aliasPath := range group.aliases {
			coordinator.statesByPath[aliasPath] = state
		}
		if state.failure != nil {
			coordinator.failureByPath[candidate.path] = ConfigFailure{
				Path:      candidate.path,
				Directory: candidate.directory,
				Kind:      state.failure.Code,
				Message:   state.failure.Message,
			}
		}
	}
	coordinator.pathSpaces = coordinator.pathSpaces.WithConfigBases(configsForBases, coordinator.fs, directoryIdentities...)
	for _, state := range loadedStates {
		matcher, err := rslintconfig.NewGlobalIgnoreMatcherWithPathSpaces(state.entries, state.candidate.directory, coordinator.fs, coordinator.pathSpaces)
		if err != nil {
			return err
		}
		state.ignoreMatcher = matcher
	}
	return nil
}

// Full Unicode folding supplies a lookup bucket, including canonically equivalent
// spellings. The filesystem decides whether two names in that bucket are one entry.
func configSpellingKey(path string) string {
	for i := range len(path) {
		if path[i] >= utf8.RuneSelf {
			return cases.Fold().String(norm.NFD.String(path))
		}
	}
	return strings.ToLower(path)
}

func (coordinator *moduleLoadCoordinator) configOwner(candidate configCandidate) string {
	directory := rslintconfig.DirectoryIdentity{LexicalPath: candidate.directory, CanonicalPath: candidate.physicalDirectory}
	if owner, ok := coordinator.ownerByDirectory[directory]; ok {
		return owner
	}
	key := configSpellingKey(directory.CanonicalPath)
	for _, owner := range coordinator.ownersByPhysical[key] {
		if samePhysicalPath(coordinator.fs, owner.CanonicalPath, directory.CanonicalPath) &&
			sameDirectoryEntries(coordinator.fs, owner.LexicalPath, directory.LexicalPath) {
			coordinator.ownerByDirectory[directory] = owner.LexicalPath
			return owner.LexicalPath
		}
	}
	coordinator.ownersByPhysical[key] = append(coordinator.ownersByPhysical[key], directory)
	coordinator.ownerByDirectory[directory] = directory.LexicalPath
	return directory.LexicalPath
}

// Two distinct symlinks can lead to the same directory, including on a volume
// with different case sensitivity from the host default. Verify the stored name
// of each differing component, not only the final filesystem object.
func sameDirectoryEntries(fsys vfs.FS, left, right string) bool {
	left, right = rslintconfig.ExactPathID(left), rslintconfig.ExactPathID(right)
	native, _ := fsys.(interface{ DirectoryEntryName(path string) string })
	for left != right {
		parent, otherParent := configDiscoveryParent(left), configDiscoveryParent(right)
		if parent == "" || otherParent == "" {
			// A UNC share is a filesystem root. Its server and share names are
			// case-insensitive and need no enumeration above the share.
			return parent == otherParent && rslintconfig.PathsEqual(left, right, false)
		}
		name, other := tspath.GetBaseFileName(left), tspath.GetBaseFileName(right)
		if name != other {
			if native != nil {
				name, other = native.DirectoryEntryName(left), native.DirectoryEntryName(right)
			} else {
				directories := fsys.GetAccessibleEntries(parent).Directories
				name, other = storedDirectoryName(directories, name), storedDirectoryName(directories, other)
			}
			if name == "" || name != other {
				return false
			}
		}
		left, right = parent, otherParent
	}
	return true
}

func storedDirectoryName(names []string, requested string) string {
	if slices.Contains(names, requested) {
		return requested
	}
	// A case-sensitive volume may preserve both Café and café while treating
	// the composed and decomposed spelling of each as the same entry.
	for _, key := range []func(string) string{norm.NFD.String, configSpellingKey} {
		want, match := key(requested), ""
		for _, name := range names {
			if key(name) == want {
				if match != "" {
					return ""
				}
				match = name
			}
		}
		if match != "" {
			return match
		}
	}
	return ""
}

func samePhysicalPath(fsys vfs.FS, left, right string) bool {
	if rslintconfig.ExactPathID(left) == rslintconfig.ExactPathID(right) {
		return true
	}
	leftInfo, rightInfo := fsys.Stat(left), fsys.Stat(right)
	return leftInfo != nil && rightInfo != nil && os.SameFile(leftInfo, rightInfo)
}

func (coordinator *moduleLoadCoordinator) validateCandidate(left, right configCandidate) error {
	if left.path == right.path {
		if rslintconfig.ExactPathID(left.physicalDirectory) != rslintconfig.ExactPathID(right.physicalDirectory) {
			return fmt.Errorf("config candidate %q changed physical directory from %q to %q", left.path, left.physicalDirectory, right.physicalDirectory)
		}
		return nil
	}
	if leftPath, rightPath := left.modulePath(), right.modulePath(); leftPath != rightPath {
		leftPath = cmp.Or(coordinator.fs.Realpath(leftPath), leftPath)
		rightPath = cmp.Or(coordinator.fs.Realpath(rightPath), rightPath)
		if !samePhysicalPath(coordinator.fs, leftPath, rightPath) {
			return fmt.Errorf("config candidates %q and %q resolve to distinct filesystem paths", left.path, right.path)
		}
	}
	return nil
}

func (coordinator *moduleLoadCoordinator) globalIgnoreMatcher(ownerPath string) (rslintconfig.GlobalIgnoreMatcher, bool) {
	if ownerPath == "" {
		return rslintconfig.GlobalIgnoreMatcher{}, false
	}
	state := coordinator.state(ownerPath)
	if state == nil || state.failure != nil {
		return rslintconfig.GlobalIgnoreMatcher{}, false
	}
	return state.ignoreMatcher, true
}

func (coordinator *moduleLoadCoordinator) failures() []ConfigFailure {
	failures := make([]ConfigFailure, 0, len(coordinator.failureByPath))
	for _, failure := range coordinator.failureByPath {
		failures = append(failures, failure)
	}
	sort.Slice(failures, func(i, j int) bool { return failures[i].Path < failures[j].Path })
	return failures
}

func allConfigsFailedError(failures []ConfigFailure) error {
	if len(failures) == 0 {
		return ErrAllConfigsFailed
	}
	failure := failures[0]
	displayPath := filepath.FromSlash(failure.Path)
	message := ""
	switch failure.Kind {
	case "invalid":
		message = fmt.Sprintf("%s: invalid config in %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
	default:
		message = fmt.Sprintf("%s: failed to load config %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
	}
	return &AllConfigsFailedError{Failures: failures, message: message}
}

func (coordinator *moduleLoadCoordinator) activateEffectiveConfigs(effectiveConfigIDs []string) ([]rslintconfig.EslintPluginEntry, error) {
	if len(effectiveConfigIDs) == 0 {
		return nil, nil
	}
	response, err := coordinator.loader.ActivateConfigs(coordinator.ctx, ConfigActivationRequest{
		ProtocolVersion:    ConfigDiscoveryProtocolVersion,
		TransactionID:      coordinator.transactionID,
		EffectiveConfigIDs: effectiveConfigIDs,
	})
	if err != nil {
		return nil, err
	}
	if response.TransactionID != coordinator.transactionID {
		return nil, configDiscoveryProtocolError(
			"activation transaction mismatch: got %q, want %q",
			response.TransactionID,
			coordinator.transactionID,
		)
	}
	return cloneEslintPluginEntries(response.EslintPluginEntries), nil
}

func validateConfigLoadBatch(request ConfigLoadBatchRequest, response ConfigLoadBatchResponse) (map[string]ConfigLoadResult, error) {
	if request.ProtocolVersion != ConfigDiscoveryProtocolVersion {
		return nil, configDiscoveryProtocolError("unsupported request protocol version %d", request.ProtocolVersion)
	}
	if request.TransactionID == "" {
		return nil, configDiscoveryProtocolError("request transactionId is empty")
	}
	if response.TransactionID != request.TransactionID {
		return nil, configDiscoveryProtocolError("transaction mismatch: got %q, want %q", response.TransactionID, request.TransactionID)
	}
	if len(response.Results) != len(request.Candidates) {
		return nil, configDiscoveryProtocolError("result count mismatch: got %d, want %d", len(response.Results), len(request.Candidates))
	}
	requestByID := make(map[string]ConfigLoadCandidate, len(request.Candidates))
	for _, candidate := range request.Candidates {
		if candidate.ID == "" {
			return nil, configDiscoveryProtocolError("request contains an empty candidate id")
		}
		if _, duplicate := requestByID[candidate.ID]; duplicate {
			return nil, configDiscoveryProtocolError("request contains duplicate candidate id %q", candidate.ID)
		}
		requestByID[candidate.ID] = candidate
	}
	results := make(map[string]ConfigLoadResult, len(response.Results))
	for index, result := range response.Results {
		_, exists := requestByID[result.ID]
		if !exists {
			return nil, configDiscoveryProtocolError("response contains unknown candidate id %q", result.ID)
		}
		if _, duplicate := results[result.ID]; duplicate {
			return nil, configDiscoveryProtocolError("response contains duplicate candidate id %q", result.ID)
		}
		if request.Candidates[index].ID != result.ID {
			return nil, configDiscoveryProtocolError("result order mismatch at index %d: got id %q, want %q", index, result.ID, request.Candidates[index].ID)
		}
		if result.Status != "loaded" && result.Status != "failed" {
			return nil, configDiscoveryProtocolError("candidate %q has invalid status %q", result.ID, result.Status)
		}
		if result.Status == "loaded" && result.Error != nil {
			return nil, configDiscoveryProtocolError("loaded candidate %q contains an error", result.ID)
		}
		if result.Status == "failed" && (result.Error == nil || result.Error.Message == "") {
			return nil, configDiscoveryProtocolError("failed candidate %q has no error message", result.ID)
		}
		results[result.ID] = result
	}
	return results, nil
}

func cloneEslintPluginEntries(entries []rslintconfig.EslintPluginEntry) []rslintconfig.EslintPluginEntry {
	if len(entries) == 0 {
		return nil
	}
	cloned := make([]rslintconfig.EslintPluginEntry, len(entries))
	for index, entry := range entries {
		cloned[index] = rslintconfig.EslintPluginEntry{
			Prefix:    entry.Prefix,
			RuleNames: append([]string(nil), entry.RuleNames...),
		}
	}
	return cloned
}
