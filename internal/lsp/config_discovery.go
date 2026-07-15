package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/hostpath"
)

const (
	methodConfigRefresh            = lsproto.Method("rslint/configRefresh")
	methodLoadConfigs              = lsproto.Method("rslint/loadConfigs")
	methodEvaluateConfigPredicates = lsproto.Method("rslint/evaluateConfigPredicates")
	methodActivateConfigs          = lsproto.Method("rslint/activateConfigs")
	methodCommitConfigs            = lsproto.Method("rslint/commitConfigs")
	methodAbortConfigs             = lsproto.Method("rslint/abortConfigs")

	configTransactionControlTimeout = 5 * time.Second
)

type configRefreshRequest struct {
	Reason string `json:"reason"`
}

type configRefreshResponse struct {
	TransactionID string `json:"transactionId"`
}

type configActivationWireResponse struct {
	TransactionID       string                     `json:"transactionId"`
	EslintPluginEntries []config.EslintPluginEntry `json:"eslintPluginEntries"`
	PluginHostReady     bool                       `json:"pluginHostReady"`
}

type configTransactionControlRequest struct {
	TransactionID string `json:"transactionId"`
}

type configCommitWireResponse struct {
	TransactionID string `json:"transactionId"`
	Committed     bool   `json:"committed"`
}

type configAbortWireResponse struct {
	TransactionID string `json:"transactionId"`
	Aborted       bool   `json:"aborted"`
}

// lspConfigModuleLoader adapts the shared discovery coordinator's loader
// boundary to LSP reverse requests. One instance belongs to exactly one
// configRefresh transaction.
type lspConfigModuleLoader struct {
	mu             sync.Mutex
	server         *Server
	transactionID  string
	frontierSeen   bool
	activated      bool
	pluginDegraded bool
}

func (loader *lspConfigModuleLoader) LoadConfigs(
	ctx context.Context,
	request discovery.ConfigLoadBatchRequest,
) (discovery.ConfigLoadBatchResponse, error) {
	if err := loader.observeTransaction(request.TransactionID); err != nil {
		return discovery.ConfigLoadBatchResponse{}, err
	}
	raw, err := loader.server.sendRequest(ctx, methodLoadConfigs, request)
	if err != nil {
		return discovery.ConfigLoadBatchResponse{}, fmt.Errorf("load config modules: %w", err)
	}
	var response discovery.ConfigLoadBatchResponse
	if err := decodeConfigTransactionResult(raw, &response, "loadConfigs"); err != nil {
		return discovery.ConfigLoadBatchResponse{}, err
	}
	if response.TransactionID != request.TransactionID {
		return discovery.ConfigLoadBatchResponse{}, fmt.Errorf(
			"loadConfigs transaction mismatch: got %q, want %q",
			response.TransactionID,
			request.TransactionID,
		)
	}
	loader.mu.Lock()
	loader.frontierSeen = true
	loader.mu.Unlock()
	return response, nil
}

func (loader *lspConfigModuleLoader) ActivateConfigs(
	ctx context.Context,
	request discovery.ConfigActivationRequest,
) (discovery.ConfigActivationResponse, error) {
	if err := loader.observeTransaction(request.TransactionID); err != nil {
		return discovery.ConfigActivationResponse{}, err
	}
	raw, err := loader.server.sendRequest(ctx, methodActivateConfigs, request)
	if err != nil {
		return discovery.ConfigActivationResponse{}, fmt.Errorf("activate config modules: %w", err)
	}
	var response configActivationWireResponse
	if err := decodeConfigTransactionResult(raw, &response, "activateConfigs"); err != nil {
		return discovery.ConfigActivationResponse{}, err
	}
	if response.TransactionID != request.TransactionID {
		return discovery.ConfigActivationResponse{}, fmt.Errorf(
			"activateConfigs transaction mismatch: got %q, want %q",
			response.TransactionID,
			request.TransactionID,
		)
	}
	if !response.PluginHostReady {
		if !loader.server.configDiscoveryHasLastGood {
			// On first startup, a valid JS catalog must not take down the entire
			// language client merely because the optional community-plugin worker
			// could not initialize. PluginLintPool staged a retryable no-host
			// generation; commit the native/semantic config atomically with it and
			// expose no plugin metadata. Once any usable catalog exists, the
			// normal last-good rule applies and a failed replacement aborts.
			loader.mu.Lock()
			loader.activated = true
			loader.pluginDegraded = true
			loader.mu.Unlock()
			return discovery.ConfigActivationResponse{
				TransactionID: response.TransactionID,
			}, nil
		}
		return discovery.ConfigActivationResponse{}, errors.New("client could not prepare the config plugin host")
	}
	loader.mu.Lock()
	loader.activated = true
	loader.mu.Unlock()
	return discovery.ConfigActivationResponse{
		TransactionID:       response.TransactionID,
		EslintPluginEntries: append([]config.EslintPluginEntry(nil), response.EslintPluginEntries...),
	}, nil
}

func (loader *lspConfigModuleLoader) EvaluateConfigPredicates(
	ctx context.Context,
	request discovery.ConfigPredicateBatchRequest,
) (discovery.ConfigPredicateBatchResponse, error) {
	if err := loader.observeTransaction(request.TransactionID); err != nil {
		return discovery.ConfigPredicateBatchResponse{}, err
	}
	raw, err := loader.server.sendRequest(ctx, methodEvaluateConfigPredicates, request)
	if err != nil {
		return discovery.ConfigPredicateBatchResponse{}, fmt.Errorf("evaluate config predicates: %w", err)
	}
	var response discovery.ConfigPredicateBatchResponse
	if err := decodeConfigTransactionResult(raw, &response, "evaluateConfigPredicates"); err != nil {
		return discovery.ConfigPredicateBatchResponse{}, err
	}
	if response.TransactionID != request.TransactionID {
		return discovery.ConfigPredicateBatchResponse{}, fmt.Errorf(
			"evaluateConfigPredicates transaction mismatch: got %q, want %q",
			response.TransactionID,
			request.TransactionID,
		)
	}
	return response, nil
}

func (loader *lspConfigModuleLoader) observeTransaction(transactionID string) error {
	if transactionID == "" {
		return errors.New("config transaction ID is empty")
	}
	loader.mu.Lock()
	defer loader.mu.Unlock()
	if loader.transactionID == "" {
		loader.transactionID = transactionID
		return nil
	}
	if loader.transactionID != transactionID {
		return fmt.Errorf(
			"config transaction changed from %q to %q",
			loader.transactionID,
			transactionID,
		)
	}
	return nil
}

func (loader *lspConfigModuleLoader) state() (transactionID string, frontierSeen bool, activated bool, pluginDegraded bool) {
	if loader == nil {
		return "", false, false, false
	}
	loader.mu.Lock()
	defer loader.mu.Unlock()
	return loader.transactionID, loader.frontierSeen, loader.activated, loader.pluginDegraded
}

func (loader *lspConfigModuleLoader) ensureActivated(
	ctx context.Context,
	catalog *discovery.ConfigCatalog,
) error {
	if catalog == nil {
		return errors.New("cannot activate a nil config catalog")
	}
	if err := loader.observeTransaction(catalog.TransactionID); err != nil {
		return err
	}
	_, frontierSeen, activated, _ := loader.state()
	if activated {
		return nil
	}
	if !frontierSeen {
		// ConfigModuleHost creates its transaction session on loadConfigs. An
		// empty catalog has no natural frontier, so explicitly establish the
		// session before asking the host to activate zero effective IDs.
		response, err := loader.LoadConfigs(ctx, discovery.ConfigLoadBatchRequest{
			TransactionID: catalog.TransactionID,
			LoadMode:      discovery.ConfigModuleLoadFresh,
			Candidates:    []discovery.ConfigLoadCandidate{},
		})
		if err != nil {
			return err
		}
		if response.Results == nil || len(response.Results) != 0 {
			return fmt.Errorf(
				"empty loadConfigs bootstrap returned %d results, want an explicit empty array",
				len(response.Results),
			)
		}
	}
	// The shared coordinator deliberately skips activation when automatic
	// discovery finds no module candidates. LSP still stages an empty plugin
	// generation so deleting the last JS config can be committed atomically.
	response, err := loader.ActivateConfigs(ctx, discovery.ConfigActivationRequest{
		TransactionID: catalog.TransactionID,
		// Preserve [] rather than nil on the wire. ConfigModuleHost requires an
		// array even when a transaction activates no effective modules.
		EffectiveConfigIDs: append([]string{}, catalog.EffectiveConfigIDs...),
	})
	if err != nil {
		return err
	}
	catalog.EslintPlugins = append([]config.EslintPluginEntry(nil), response.EslintPluginEntries...)
	return nil
}

func (loader *lspConfigModuleLoader) commit(ctx context.Context, transactionID string) error {
	if err := loader.observeTransaction(transactionID); err != nil {
		return err
	}
	request := configTransactionControlRequest{
		TransactionID: transactionID,
	}
	raw, err := loader.server.sendRequest(ctx, methodCommitConfigs, request)
	if err != nil {
		return fmt.Errorf("commit config transaction: %w", err)
	}
	var response configCommitWireResponse
	if err := decodeConfigTransactionResult(raw, &response, "commitConfigs"); err != nil {
		return err
	}
	if response.TransactionID != transactionID || !response.Committed {
		return fmt.Errorf(
			"invalid commitConfigs response: transactionId=%q committed=%t",
			response.TransactionID,
			response.Committed,
		)
	}
	return nil
}

func (loader *lspConfigModuleLoader) abort(ctx context.Context, transactionID string) error {
	if transactionID == "" {
		return nil
	}
	request := configTransactionControlRequest{
		TransactionID: transactionID,
	}
	raw, err := loader.server.sendRequest(ctx, methodAbortConfigs, request)
	if err != nil {
		return fmt.Errorf("abort config transaction: %w", err)
	}
	var response configAbortWireResponse
	if err := decodeConfigTransactionResult(raw, &response, "abortConfigs"); err != nil {
		return err
	}
	if response.TransactionID != transactionID || !response.Aborted {
		return fmt.Errorf(
			"invalid abortConfigs response: transactionId=%q aborted=%t",
			response.TransactionID,
			response.Aborted,
		)
	}
	return nil
}

func decodeConfigTransactionResult(raw any, target any, method string) error {
	data, err := stdjson.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal %s result: %w", method, err)
	}
	if err := stdjson.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s result: %w", method, err)
	}
	return nil
}

type lspDiscoveredConfigSnapshot struct {
	configs             map[string]config.RslintConfig
	evaluators          map[string]*config.ConfigEvaluator
	tsConfigPaths       map[string][]string
	ownerResolver       *config.ConfigOwnerResolver
	unavailableConfigs  map[string]struct{}
	jsonConfig          config.RslintConfig
	jsonConfigPath      string
	jsonTsConfigPaths   []string
	transactionID       string
	eslintPluginEntries []config.EslintPluginEntry
	usableLastGood      bool
	lookupDirs          map[string]struct{}
	catalog             *discovery.ConfigCatalog
}

func (s *Server) openDocumentConfigLookupPaths() []string {
	paths := make([]string, 0, len(s.documents))
	seen := make(map[string]struct{}, len(s.documents))
	caseSensitive := s.fs == nil || s.fs.UseCaseSensitiveFileNames()
	for uri := range s.documents {
		filePath := normalizeURIHostPath(uri)
		if filePath == "" {
			continue
		}
		id := lspLexicalPathID(filePath, caseSensitive)
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		paths = append(paths, filePath)
	}
	sort.Strings(paths)
	return paths
}

func (s *Server) handleConfigRefresh(ctx context.Context, params any) (configRefreshResponse, error) {
	var request configRefreshRequest
	data, err := stdjson.Marshal(params)
	if err != nil {
		return configRefreshResponse{}, fmt.Errorf("marshal config refresh params: %w", err)
	}
	if err := stdjson.Unmarshal(data, &request); err != nil {
		return configRefreshResponse{}, fmt.Errorf("parse config refresh params: %w", err)
	}
	switch request.Reason {
	case "initial", "config-change", "gitignore-change", "dependency-change", "open-document":
	default:
		return configRefreshResponse{}, fmt.Errorf("unsupported config refresh reason %q", request.Reason)
	}
	if s.fs == nil {
		return configRefreshResponse{}, errors.New("config refresh requires a filesystem")
	}
	// Set this before doing fallible discovery: a failed initial transaction is
	// still proof that the client installed the reverse handlers, and a later
	// config or config-scoped .gitignore event must retry while keeping last-good.
	s.configDiscoveryActive = true

	// cachedvfs is intentionally scoped to one generation. A long-lived cache
	// would make file creation/deletion and .gitignore edits invisible to later
	// refreshes; retaining the committed resolver retains only the last-good
	// generation's path identity snapshot.
	snapshotFS := newConfigSnapshotFS(bundled.WrapFS(cachedvfs.From(s.fs)))
	loader := &lspConfigModuleLoader{server: s}
	lookupPaths := s.openDocumentConfigLookupPaths()
	catalog, err := discovery.Build(ctx, snapshotFS, loader, discovery.ConfigDiscoveryRequest{
		CWD:                     hostpath.Normalize(s.cwd),
		Mode:                    discovery.ConfigDiscoveryAuto,
		Inputs:                  []string{"."},
		CollectTargets:          false,
		GlobInputPaths:          true,
		AllowMissingConfig:      true,
		RetainUnconfiguredAreas: true,
		// Initial discovery retains successful siblings when another config is
		// broken. Later generations must keep collecting while such a committed
		// unavailable boundary exists, otherwise that known failure would make
		// every unrelated config edit from ever committing.
		CollectConfigFailures: !s.configDiscoveryHasLastGood || len(s.jsUnavailableConfigs) > 0,
		ProbeRootConfig:       true,
		LookupPaths:           lookupPaths,
		Fresh:                 true,
	})
	catalogCommitted := false
	defer func() {
		if catalog != nil && !catalogCommitted {
			catalog.ClosePredicateEvaluation()
		}
	}()
	recoveredUnavailable := false
	var unavailableCause error
	var recoveryFailures []discovery.ConfigFailure
	if err != nil {
		transactionID, _, _, _ := loader.state()
		if !errors.Is(err, discovery.ErrAllConfigsFailed) || s.configDiscoveryHasLastGood {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, transactionID, err)
		}
		// A broken config on first startup must not tear down the language
		// client, and JSON fallback must not silently lint through that broken
		// boundary. Commit a real empty Node generation plus unavailable Go
		// ownership boundaries. This is intentionally not "last-good": a later
		// refresh may replace it with another recovery snapshot until one full
		// usable catalog has committed.
		unavailableCause = err
		var allFailed *discovery.AllConfigsFailedError
		if !errors.As(err, &allFailed) || len(allFailed.Failures) == 0 || transactionID == "" {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(
				ctx,
				loader,
				transactionID,
				errors.Join(unavailableCause, errors.New("all-config-failed recovery has no typed failure catalog")),
			)
		}
		catalog = &discovery.ConfigCatalog{
			TransactionID:      transactionID,
			Configs:            make(map[string]config.RslintConfig),
			EffectiveConfigIDs: []string{},
		}
		recoveryFailures = append([]discovery.ConfigFailure(nil), allFailed.Failures...)
		recoveredUnavailable = true
	} else {
		recoveryFailures = append([]discovery.ConfigFailure(nil), catalog.Failures...)
		if len(recoveryFailures) > 0 {
			unavailableCause = &discovery.AllConfigsFailedError{Failures: recoveryFailures}
			if s.configDiscoveryHasLastGood &&
				!configFailuresCoveredByCommittedUnavailable(
					snapshotFS,
					recoveryFailures,
					s.jsConfigs,
					s.jsUnavailableConfigs,
				) {
				return configRefreshResponse{}, s.abortFailedConfigRefresh(
					ctx,
					loader,
					catalog.TransactionID,
					unavailableCause,
				)
			}
			recoveredUnavailable = true
		}
	}

	snapshot, err := s.prepareDiscoveredConfigSnapshot(snapshotFS, catalog, recoveryFailures)
	if err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
	}
	if recoveredUnavailable && len(snapshot.unavailableConfigs) == 0 {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(
			ctx,
			loader,
			catalog.TransactionID,
			errors.Join(unavailableCause, errors.New("all-config-failed recovery has no valid failure boundary")),
		)
	}
	// Discovery has already evaluated every path needed to build this catalog.
	// Its retained evaluator and snapshot filesystem own any later exact-path
	// .gitignore reads; a watcher event starts a new transaction rather than
	// mutating this catalog or its live predicate session.
	if err := loader.ensureActivated(ctx, catalog); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
	}
	snapshot.eslintPluginEntries = append([]config.EslintPluginEntry(nil), catalog.EslintPlugins...)

	// Cancellation is honored throughout discovery and activation. Once commit
	// begins, finish the local two-phase control exchange under a short bounded
	// context so cancellation cannot leave Node committed while Go silently
	// abandons the matching catalog.
	if err := ctx.Err(); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
	}
	controlCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), configTransactionControlTimeout)
	defer cancel()
	if err := loader.commit(controlCtx, catalog.TransactionID); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
	}

	s.commitDiscoveredConfigSnapshot(ctx, snapshot)
	catalogCommitted = true
	_, _, _, pluginDegraded := loader.state()
	if pluginDegraded {
		log.Printf("[rslint] Committed config catalog without the unavailable ESLint-plugin host; community plugin rules are disabled until a later successful refresh")
	}
	if recoveredUnavailable {
		log.Printf(
			"[rslint] JavaScript configs failed during initial discovery; committed %d unavailable boundaries: %v",
			len(snapshot.unavailableConfigs),
			unavailableCause,
		)
	}
	return configRefreshResponse{
		TransactionID: catalog.TransactionID,
	}, nil
}

func lspLexicalPathID(filePath string, caseSensitive bool) string {
	filePath = normalizeRootAwareHostPath(filePath)
	return hostpath.Identity(filePath, hostpath.DirectoryForRoot(filePath, filePath), caseSensitive)
}

func (s *Server) abortFailedConfigRefresh(
	ctx context.Context,
	loader *lspConfigModuleLoader,
	transactionID string,
	cause error,
) error {
	if loader == nil || transactionID == "" {
		return cause
	}
	abortCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), configTransactionControlTimeout)
	defer cancel()
	if err := loader.abort(abortCtx, transactionID); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (s *Server) prepareDiscoveredConfigSnapshot(
	fsys vfs.FS,
	catalog *discovery.ConfigCatalog,
	recoveryFailures []discovery.ConfigFailure,
) (*lspDiscoveredConfigSnapshot, error) {
	if catalog == nil {
		return nil, errors.New("cannot prepare a nil config catalog")
	}
	snapshot := &lspDiscoveredConfigSnapshot{
		configs:            make(map[string]config.RslintConfig, len(catalog.Configs)),
		evaluators:         make(map[string]*config.ConfigEvaluator, len(catalog.Configs)),
		tsConfigPaths:      make(map[string][]string, len(catalog.Configs)),
		unavailableConfigs: make(map[string]struct{}),
		transactionID:      catalog.TransactionID,
		// An empty catalog is a successfully committed absence of JavaScript
		// config, not a usable JavaScript last-good generation. If a broken
		// config appears later, it must establish an unavailable boundary rather
		// than preserving JSON fallback beneath that new boundary.
		usableLastGood: len(catalog.Configs) > 0,
		lookupDirs:     make(map[string]struct{}),
		catalog:        catalog,
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	for _, filePath := range s.openDocumentConfigLookupPaths() {
		directory := hostpath.DirectoryForRoot(filePath, filePath)
		snapshot.lookupDirs[lspLexicalPathID(directory, caseSensitive)] = struct{}{}
	}
	for _, configDir := range catalog.ConfigDirectories() {
		entries := catalog.Configs[configDir]
		if err := config.ValidateConfig(entries); err != nil {
			return nil, fmt.Errorf("invalid discovered config for %q: %w", configDir, err)
		}
		if err := validateRuleOptionsForConfig(entries, configDir); err != nil {
			return nil, err
		}
		paths, err := resolveTsConfigPathsWithFS(entries, configDir, fsys)
		if err != nil {
			return nil, fmt.Errorf("resolve tsconfig paths for %q: %w", configDir, err)
		}
		snapshot.configs[configDir] = append(config.RslintConfig(nil), entries...)
		evaluator := catalog.ConfigEvaluatorForDirectory(configDir)
		if evaluator == nil {
			return nil, fmt.Errorf("config discovery invariant: discovered JavaScript config %q has no committed evaluator", configDir)
		}
		snapshot.evaluators[configDir] = evaluator
		snapshot.tsConfigPaths[configDir] = paths
	}

	// Recovery materializes only the outermost failed nearest-config directories
	// as retryable empty boundaries. A catalog may also contain successful
	// sibling owners, so one committed snapshot can intentionally be partial.
	for _, configDir := range unavailableConfigBoundaryDirectories(fsys, recoveryFailures) {
		if _, exists := snapshot.configs[configDir]; exists {
			continue
		}
		snapshot.configs[configDir] = config.RslintConfig{}
		snapshot.tsConfigPaths[configDir] = nil
		snapshot.unavailableConfigs[configDir] = struct{}{}
	}
	// JavaScript owners retain discovery's live evaluator. Its .gitignore layer
	// reads only exact target ancestry and is already backed by this generation's
	// snapshot filesystem, so no owner-subtree materialization is needed here.
	snapshot.ownerResolver = config.NewConfigOwnerResolver(snapshot.configs, fsys)

	jsonConfig, jsonPath, jsonTsConfigs, err := loadJSONConfigFallbackWithFS(s.cwd, fsys)
	if err != nil {
		return nil, err
	}
	// JSON fallback owns only portions of the workspace not handed to a JS/TS
	// config (including an unavailable one). Include a synthetic cwd owner solely
	// to calculate its direct source boundaries; JS ownership remains unchanged.
	jsonBoundaryConfigs := make(map[string]config.RslintConfig, len(snapshot.configs)+1)
	for configDir, entries := range snapshot.configs {
		jsonBoundaryConfigs[configDir] = entries
	}
	jsonCWD := hostpath.Normalize(s.cwd)
	jsonBoundaryConfigs[jsonCWD] = jsonConfig
	jsonBoundaryResolver := config.NewConfigOwnerResolver(jsonBoundaryConfigs, fsys)
	snapshot.jsonConfig = config.ConfigWithGitignoreWithBoundaries(
		jsonConfig,
		jsonCWD,
		fsys,
		nil,
		jsonBoundaryResolver.ChildConfigDirs(jsonCWD),
	)
	snapshot.jsonConfigPath = jsonPath
	snapshot.jsonTsConfigPaths = jsonTsConfigs
	return snapshot, nil
}

// unavailableConfigBoundaryDirectories returns the shallowest failed config
// directories. A successful ancestor does not erase a broken nearest-config
// boundary: files below that boundary must not silently inherit the ancestor
// or JSON fallback. Nested failures covered by one failed boundary are
// redundant.
func unavailableConfigBoundaryDirectories(
	fsys vfs.FS,
	failures []discovery.ConfigFailure,
) []string {
	if len(failures) == 0 {
		return nil
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	byIdentity := make(map[string]string, len(failures))
	for _, failure := range failures {
		directory := hostpath.Normalize(failure.Directory)
		if directory == "" {
			continue
		}
		identity := lspLexicalPathID(directory, caseSensitive)
		if _, exists := byIdentity[identity]; !exists {
			byIdentity[identity] = directory
		}
	}

	candidates := make([]string, 0, len(byIdentity))
	for _, directory := range byIdentity {
		candidates = append(candidates, directory)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if len(candidates[i]) != len(candidates[j]) {
			return len(candidates[i]) < len(candidates[j])
		}
		return candidates[i] < candidates[j]
	})
	boundaries := candidates[:0]
	for _, candidate := range candidates {
		covered := false
		for _, boundary := range boundaries {
			_, within := hostpath.RelativeWithin(candidate, boundary, caseSensitive)
			if within {
				covered = true
				break
			}
		}
		if !covered {
			boundaries = append(boundaries, candidate)
		}
	}
	return boundaries
}

// configFailuresCoveredByCommittedUnavailable permits a partial replacement
// only when every failed candidate remains inside a boundary that was already
// committed as unavailable. A formerly usable config becoming invalid (even
// below such a boundary) and a brand-new failed boundary both reject the whole
// transaction, preserving the complete last-good generation.
func configFailuresCoveredByCommittedUnavailable(
	fsys vfs.FS,
	failures []discovery.ConfigFailure,
	committedConfigs map[string]config.RslintConfig,
	committedUnavailable map[string]struct{},
) bool {
	if len(failures) == 0 || len(committedUnavailable) == 0 {
		return false
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	type unavailableBoundary struct {
		lexical  string
		physical string
	}
	unavailableByID := make(map[string]struct{}, len(committedUnavailable))
	unavailableBoundaries := make([]unavailableBoundary, 0, len(committedUnavailable))
	for directory := range committedUnavailable {
		directory = hostpath.Normalize(directory)
		unavailableByID[lspLexicalPathID(directory, caseSensitive)] = struct{}{}
		unavailableBoundaries = append(unavailableBoundaries, unavailableBoundary{
			lexical:  directory,
			physical: lspPhysicalDirectoryPath(directory, fsys),
		})
	}
	activeLexicalByID := make(map[string]struct{}, len(committedConfigs))
	activePhysicalByID := make(map[string]struct{}, len(committedConfigs))
	for directory := range committedConfigs {
		lexicalID := lspLexicalPathID(directory, caseSensitive)
		if _, unavailable := unavailableByID[lexicalID]; !unavailable {
			activeLexicalByID[lexicalID] = struct{}{}
			activePhysicalByID[lspLexicalPathID(lspPhysicalDirectoryPath(directory, fsys), caseSensitive)] = struct{}{}
		}
	}

	for _, failure := range failures {
		directory := hostpath.Normalize(failure.Directory)
		if directory == "" {
			return false
		}
		physicalDirectory := lspPhysicalDirectoryPath(directory, fsys)
		if _, wasActive := activeLexicalByID[lspLexicalPathID(directory, caseSensitive)]; wasActive {
			return false
		}
		if _, wasActive := activePhysicalByID[lspLexicalPathID(physicalDirectory, caseSensitive)]; wasActive {
			return false
		}
		covered := false
		for _, unavailable := range unavailableBoundaries {
			if _, within := hostpath.RelativeWithin(directory, unavailable.lexical, caseSensitive); within {
				covered = true
				break
			}
			if _, within := hostpath.RelativeWithin(physicalDirectory, unavailable.physical, caseSensitive); within {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

func lspPhysicalDirectoryPath(directory string, fsys vfs.FS) string {
	directory = hostpath.Normalize(directory)
	if fsys != nil {
		if realPath := fsys.Realpath(directory); realPath != "" {
			return hostpath.NormalizeForRoot(directory, realPath)
		}
	}
	return directory
}

func (s *Server) commitDiscoveredConfigSnapshot(ctx context.Context, snapshot *lspDiscoveredConfigSnapshot) {
	// No fallible work remains after the client commits its staged plugin host.
	// The serialized dispatch loop makes this map swap atomic to every document
	// and code-action handler.
	s.invalidateOpenDocumentDiagnostics()
	previousCatalog := s.activeConfigCatalog
	s.jsConfigs = snapshot.configs
	s.jsConfigEvaluators = snapshot.evaluators
	s.activeConfigCatalog = snapshot.catalog
	s.tsConfigPathsByConfig = snapshot.tsConfigPaths
	s.jsConfigOwnerResolver = snapshot.ownerResolver
	s.jsUnavailableConfigs = snapshot.unavailableConfigs
	s.configDiscoveryLookupDirs = snapshot.lookupDirs
	s.jsonConfig = snapshot.jsonConfig
	s.rslintConfigPath = snapshot.jsonConfigPath
	s.tsConfigPaths = snapshot.jsonTsConfigPaths
	s.eslintPluginConfigGeneration = snapshot.transactionID
	s.eslintPluginRules = eslintPluginRuleSet(snapshot.eslintPluginEntries)
	s.configDiscoveryHasLastGood = snapshot.usableLastGood
	s.configSnapshotIncludesGitignore = true
	config.RegisterEslintPluginRules(snapshot.eslintPluginEntries)
	if previousCatalog != nil && previousCatalog != snapshot.catalog {
		previousCatalog.ClosePredicateEvaluation()
	}
	log.Printf("[rslint] Committed Go-discovered JS/TS config catalog (%d config files)", len(snapshot.configs))
	if err := s.RefreshDiagnostics(ctx); err != nil {
		log.Printf("[rslint] Failed to refresh diagnostics after config refresh: %v", err)
	}
}

func loadJSONConfigFallbackWithFS(
	cwd string,
	fsys vfs.FS,
) (config.RslintConfig, string, []string, error) {
	configPath, found := findRslintConfig(fsys, cwd)
	if !found {
		return config.RslintConfig{}, "", nil, nil
	}
	loader := config.NewConfigLoader(fsys, cwd)
	rslintConfig, _, err := loader.LoadRslintConfig(configPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("load JSON fallback %q: %w", configPath, err)
	}
	paths, err := resolveTsConfigPathsWithFS(rslintConfig, cwd, fsys)
	if err != nil {
		return nil, "", nil, fmt.Errorf("resolve tsconfig paths for JSON fallback %q: %w", configPath, err)
	}
	return rslintConfig, configPath, paths, nil
}

func resolveTsConfigPathsWithFS(cfg config.RslintConfig, cwd string, fsys vfs.FS) ([]string, error) {
	paths, err := config.ResolveTsConfigPaths(cfg, cwd, fsys)
	if err != nil {
		return nil, err
	}
	for index, projectPath := range paths {
		if realPath := fsys.Realpath(projectPath); realPath != "" {
			projectPath = realPath
		}
		paths[index] = tspath.NormalizePath(projectPath)
	}
	return paths, nil
}
