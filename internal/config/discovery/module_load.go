package discovery

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type configCandidate struct {
	path      string
	directory string
	// configArrayBase is set only when ConfigArray A differs from the config
	// module directory, as for an explicit config. Empty means directory.
	configArrayBase string
}

// configLoadState binds one candidate to the raw config and authored-ignore
// matcher returned by Node. The effective catalog always clones these entries
// before adding Git-derived config.
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

	statesByPath     map[string]*configLoadState
	stateByIdentity  map[tspath.Path]*configLoadState
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
		ctx:             ctx,
		fs:              fsys,
		loader:          loader,
		transactionID:   transactionID,
		fresh:           fresh,
		singleThreaded:  singleThreaded,
		statesByPath:    make(map[string]*configLoadState),
		stateByIdentity: make(map[tspath.Path]*configLoadState),
		failureByPath:   make(map[string]ConfigFailure),
	}
}

func (coordinator *moduleLoadCoordinator) state(candidatePath string) *configLoadState {
	return coordinator.statesByPath[candidatePath]
}

func (coordinator *moduleLoadCoordinator) hasCandidateResults() bool {
	return len(coordinator.statesByPath) > 0
}

func (coordinator *moduleLoadCoordinator) loadCandidates(rawCandidates []configCandidate) error {
	type candidateGroup struct {
		candidate configCandidate
		aliases   map[string]configCandidate
	}
	groups := make(map[tspath.Path]*candidateGroup, len(rawCandidates))
	for _, candidate := range rawCandidates {
		candidate.path = tspath.NormalizePath(candidate.path)
		candidate.directory = tspath.NormalizePath(candidate.directory)
		identity := tspath.ToPath(candidate.path, "", coordinator.fs.UseCaseSensitiveFileNames())
		if state := coordinator.stateByIdentity[identity]; state != nil {
			if err := coordinator.validateNativeCaseAlias(state.candidate, candidate); err != nil {
				return err
			}
			coordinator.statesByPath[candidate.path] = state
			continue
		}
		group := groups[identity]
		if group == nil {
			group = &candidateGroup{
				candidate: candidate,
				aliases:   make(map[string]configCandidate),
			}
			groups[identity] = group
		} else {
			if err := coordinator.validateNativeCaseAlias(group.candidate, candidate); err != nil {
				return err
			}
			if candidate.path < group.candidate.path {
				group.candidate = candidate
			}
		}
		group.aliases[candidate.path] = candidate
	}
	if len(groups) == 0 {
		return nil
	}
	if coordinator.loader == nil {
		return errors.New("javascript config candidates require a module loader")
	}

	paths := make([]string, 0, len(groups))
	groupByPath := make(map[string]*candidateGroup, len(groups))
	for _, group := range groups {
		paths = append(paths, group.candidate.path)
		groupByPath[group.candidate.path] = group
	}
	sort.Strings(paths)
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
	request.Candidates = make([]ConfigLoadCandidate, 0, len(paths))
	for _, path := range paths {
		candidate := groupByPath[path].candidate
		coordinator.nextCandidateID++
		id := fmt.Sprintf("config-%06d", coordinator.nextCandidateID)
		request.Candidates = append(request.Candidates, ConfigLoadCandidate{
			ID:              id,
			ConfigPath:      candidate.path,
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

	for _, wireCandidate := range request.Candidates {
		group := groupByPath[wireCandidate.ConfigPath]
		candidate := group.candidate
		result := results[wireCandidate.ID]
		state := &configLoadState{
			id:        wireCandidate.ID,
			candidate: candidate,
		}
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
			state.ignoreMatcher = rslintconfig.NewGlobalIgnoreMatcher(
				state.entries,
				candidate.directory,
				coordinator.fs,
			)
			coordinator.configsLoaded++
		}
		identity := tspath.ToPath(candidate.path, "", coordinator.fs.UseCaseSensitiveFileNames())
		coordinator.stateByIdentity[identity] = state
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
	return nil
}

// validateNativeCaseAlias protects candidate coalescing with the same
// physical-directory invariant used by the final catalog.
func (coordinator *moduleLoadCoordinator) validateNativeCaseAlias(left configCandidate, right configCandidate) error {
	if left.path == right.path {
		return nil
	}
	if coordinator.fs.UseCaseSensitiveFileNames() || !strings.EqualFold(left.path, right.path) {
		return fmt.Errorf("config candidate identity collision between %q and %q", left.path, right.path)
	}
	physicalPath := func(path string) string {
		if realPath := coordinator.fs.Realpath(path); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
		return tspath.NormalizePath(path)
	}
	leftPhysicalDirectory := physicalPath(left.directory)
	rightPhysicalDirectory := physicalPath(right.directory)
	leftPhysicalPath := physicalPath(left.path)
	rightPhysicalPath := physicalPath(right.path)
	if tspath.ToPath(leftPhysicalDirectory, "", true) != tspath.ToPath(rightPhysicalDirectory, "", true) ||
		tspath.ToPath(leftPhysicalPath, "", true) != tspath.ToPath(rightPhysicalPath, "", true) {
		return fmt.Errorf(
			"config candidates %q and %q differ only by case but resolve to distinct filesystem paths",
			left.path,
			right.path,
		)
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

func (coordinator *moduleLoadCoordinator) allConfigsFailedError() error {
	failures := coordinator.failures()
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
	if !coordinator.hasCandidateResults() {
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
