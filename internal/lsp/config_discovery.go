package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

const (
	methodConfigRefresh   = lsproto.Method("rslint/configRefresh")
	methodLoadConfigs     = lsproto.Method("rslint/loadConfigs")
	methodActivateConfigs = lsproto.Method("rslint/activateConfigs")
	methodCommitConfigs   = lsproto.Method("rslint/commitConfigs")
	methodAbortConfigs    = lsproto.Method("rslint/abortConfigs")

	configTransactionControlTimeout = 5 * time.Second
)

type configRefreshRequest struct {
	ProtocolVersion int    `json:"protocolVersion"`
	Reason          string `json:"reason"`
}

type configRefreshResponse struct {
	TransactionID string `json:"transactionId"`
	Generation    string `json:"generation"`
	ConfigCount   int    `json:"configCount"`
}

type configActivationWireResponse struct {
	TransactionID       string                     `json:"transactionId"`
	Generation          string                     `json:"generation"`
	EslintPluginEntries []config.EslintPluginEntry `json:"eslintPluginEntries"`
	PluginHostReady     bool                       `json:"pluginHostReady"`
}

type configTransactionControlRequest struct {
	ProtocolVersion int    `json:"protocolVersion"`
	TransactionID   string `json:"transactionId"`
}

type configCommitWireResponse struct {
	TransactionID string `json:"transactionId"`
	Generation    string `json:"generation"`
	Committed     bool   `json:"committed"`
}

type configAbortWireResponse struct {
	TransactionID string `json:"transactionId"`
	Generation    string `json:"generation"`
	Aborted       bool   `json:"aborted"`
}

// lspConfigModuleLoader adapts the shared discovery coordinator's loader
// boundary to LSP reverse requests. One instance belongs to exactly one
// configRefresh transaction.
type lspConfigModuleLoader struct {
	server         *Server
	transactionID  string
	frontierSeen   bool
	activated      bool
	pluginDegraded bool
	candidates     []discovery.ConfigLoadCandidate
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
	loader.frontierSeen = true
	// Keep the exact Go-selected candidate boundaries for LSP's startup
	// recovery policy. The shared coordinator remains authoritative for
	// validating the response and deciding whether the build is ErrAll; this
	// copy is consulted only after that sentinel is returned.
	loader.candidates = append(loader.candidates, request.Candidates...)
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
	if response.Generation != request.TransactionID {
		return discovery.ConfigActivationResponse{}, fmt.Errorf(
			"activateConfigs generation mismatch: got %q, want %q",
			response.Generation,
			request.TransactionID,
		)
	}
	if !response.PluginHostReady {
		if !loader.server.configDiscoveryV2HasLastGood {
			// On first startup, a valid JS catalog must not take down the entire
			// language client merely because the optional community-plugin worker
			// could not initialize. PluginLintPool staged a retryable no-host
			// generation; commit the native/semantic config atomically with it and
			// expose no plugin metadata. Once any usable v2 catalog exists, the
			// normal last-good rule applies and a failed replacement aborts.
			loader.activated = true
			loader.pluginDegraded = true
			return discovery.ConfigActivationResponse{
				TransactionID: response.TransactionID,
			}, nil
		}
		return discovery.ConfigActivationResponse{}, errors.New("client could not prepare the config plugin host")
	}
	loader.activated = true
	return discovery.ConfigActivationResponse{
		TransactionID:       response.TransactionID,
		EslintPluginEntries: append([]config.EslintPluginEntry(nil), response.EslintPluginEntries...),
	}, nil
}

func (loader *lspConfigModuleLoader) observeTransaction(transactionID string) error {
	if transactionID == "" {
		return errors.New("config transaction ID is empty")
	}
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
	if loader.activated {
		return nil
	}
	if !loader.frontierSeen {
		// ConfigModuleHost creates its transaction session on loadConfigs. An
		// empty catalog has no natural frontier, so explicitly establish the
		// session before asking the host to activate zero effective IDs.
		response, err := loader.LoadConfigs(ctx, discovery.ConfigLoadBatchRequest{
			ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
			TransactionID:   catalog.TransactionID,
			LoadMode:        discovery.ConfigModuleLoadFresh,
			Candidates:      []discovery.ConfigLoadCandidate{},
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
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		TransactionID:   catalog.TransactionID,
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

func (loader *lspConfigModuleLoader) unavailableCatalog(fsys vfs.FS) (*discovery.ConfigCatalog, error) {
	if loader == nil || loader.transactionID == "" || len(loader.candidates) == 0 {
		return nil, errors.New("all-config-failed recovery has no loaded candidate frontier")
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	directoriesByID := make(map[string]string, len(loader.candidates))
	for _, candidate := range loader.candidates {
		directory := tspath.NormalizePath(candidate.ConfigDirectory)
		if directory == "" {
			continue
		}
		identity := lspLexicalPathID(directory, caseSensitive)
		if _, exists := directoriesByID[identity]; !exists {
			directoriesByID[identity] = directory
		}
	}
	if len(directoriesByID) == 0 {
		return nil, errors.New("all-config-failed recovery has no valid config directory")
	}
	configs := make(map[string]config.RslintConfig, len(directoriesByID))
	for _, directory := range directoriesByID {
		// An empty entry set establishes an ownership boundary without ever
		// applying JSON fallback beneath a config module that did not load.
		configs[directory] = config.RslintConfig{}
	}
	return &discovery.ConfigCatalog{
		TransactionID:      loader.transactionID,
		Configs:            configs,
		EffectiveConfigIDs: []string{},
	}, nil
}

func (loader *lspConfigModuleLoader) commit(ctx context.Context, transactionID string) error {
	if err := loader.observeTransaction(transactionID); err != nil {
		return err
	}
	request := configTransactionControlRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		TransactionID:   transactionID,
	}
	raw, err := loader.server.sendRequest(ctx, methodCommitConfigs, request)
	if err != nil {
		return fmt.Errorf("commit config transaction: %w", err)
	}
	var response configCommitWireResponse
	if err := decodeConfigTransactionResult(raw, &response, "commitConfigs"); err != nil {
		return err
	}
	if response.TransactionID != transactionID || response.Generation != transactionID || !response.Committed {
		return fmt.Errorf(
			"invalid commitConfigs response: transactionId=%q generation=%q committed=%t",
			response.TransactionID,
			response.Generation,
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
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		TransactionID:   transactionID,
	}
	raw, err := loader.server.sendRequest(ctx, methodAbortConfigs, request)
	if err != nil {
		return fmt.Errorf("abort config transaction: %w", err)
	}
	var response configAbortWireResponse
	if err := decodeConfigTransactionResult(raw, &response, "abortConfigs"); err != nil {
		return err
	}
	if response.TransactionID != transactionID || response.Generation != transactionID || !response.Aborted {
		return fmt.Errorf(
			"invalid abortConfigs response: transactionId=%q generation=%q aborted=%t",
			response.TransactionID,
			response.Generation,
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
	tsConfigPaths       map[string][]string
	ownerResolver       *config.ConfigOwnerResolver
	configKeyByPath     map[string]string
	unavailableConfigs  map[string]struct{}
	jsonConfig          config.RslintConfig
	jsonConfigPath      string
	jsonTsConfigPaths   []string
	generation          string
	eslintPluginEntries []config.EslintPluginEntry
	usableLastGood      bool
	configGenerationFS  vfs.FS
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
	if request.ProtocolVersion != discovery.ConfigDiscoveryProtocolVersion {
		return configRefreshResponse{}, fmt.Errorf(
			"unsupported config refresh protocol %d",
			request.ProtocolVersion,
		)
	}
	switch request.Reason {
	case "initial", "config-change", "gitignore-change", "dependency-change":
	default:
		return configRefreshResponse{}, fmt.Errorf("unsupported config refresh reason %q", request.Reason)
	}
	if s.fs == nil {
		return configRefreshResponse{}, errors.New("config refresh requires a filesystem")
	}
	// Set this before doing fallible discovery: a failed initial transaction is
	// still proof that the client installed the v2 reverse handlers, and a later
	// config or config-scoped .gitignore event must retry while keeping last-good.
	s.configDiscoveryV2Active = true

	// cachedvfs is intentionally scoped to one generation. A long-lived cache
	// would make file creation/deletion and .gitignore edits invisible to later
	// refreshes; retaining the committed resolver retains only the last-good
	// generation's path identity snapshot.
	generationFS := newConfigGenerationFS(bundled.WrapFS(cachedvfs.From(s.fs)))
	loader := &lspConfigModuleLoader{server: s}
	catalog, err := discovery.Build(ctx, generationFS, loader, discovery.ConfigDiscoveryRequest{
		CWD:         tspath.NormalizePath(s.cwd),
		Mode:        discovery.ConfigDiscoveryAuto,
		ImplicitCWD: true,
		Fresh:       true,
	})
	recoveredUnavailable := false
	var unavailableCause error
	if err != nil {
		if !errors.Is(err, discovery.ErrAllConfigsFailed) || s.configDiscoveryV2HasLastGood {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, loader.transactionID, err)
		}
		// A broken config on first startup must not tear down the language
		// client, and JSON fallback must not silently lint through that broken
		// boundary. Commit a real empty Node generation plus unavailable Go
		// ownership boundaries. This is intentionally not "last-good": a later
		// refresh may replace it with another recovery snapshot until one full
		// usable catalog has committed.
		unavailableCause = err
		catalog, err = loader.unavailableCatalog(generationFS)
		if err != nil {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(
				ctx,
				loader,
				loader.transactionID,
				errors.Join(unavailableCause, err),
			)
		}
		recoveredUnavailable = true
	} else if s.configDiscoveryV2HasLastGood {
		if failure, invalidates := s.failureAtCommittedConfigBoundary(generationFS, catalog); invalidates {
			err = fmt.Errorf(
				"config refresh failed at last-good boundary %q: %s",
				failure.Directory,
				failure.Message,
			)
			return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
		}
	}

	snapshot, err := s.prepareDiscoveredConfigSnapshot(generationFS, catalog)
	if err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, catalog.TransactionID, err)
	}
	// Activation/commit can overlap a later filesystem mutation. Freeze the
	// candidate's observed .gitignore state now; that mutation belongs to the
	// next transaction and must not alter this generation's target admission.
	generationFS.freeze()
	snapshot.configGenerationFS = generationFS
	if recoveredUnavailable {
		snapshot.usableLastGood = false
		for configDir := range snapshot.configs {
			snapshot.unavailableConfigs[configDir] = struct{}{}
		}
	}
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
	if loader.pluginDegraded {
		log.Printf("[rslint] Committed config catalog without the unavailable ESLint-plugin host; community plugin rules are disabled until a later successful refresh")
	}
	if recoveredUnavailable {
		log.Printf(
			"[rslint] JavaScript configs failed during initial discovery; committed %d unavailable boundaries: %v",
			len(snapshot.unavailableConfigs),
			unavailableCause,
		)
	}
	for _, failure := range catalog.Failures {
		log.Printf("[rslint] Skipped config %s: %s", failure.Path, failure.Message)
	}
	return configRefreshResponse{
		TransactionID: catalog.TransactionID,
		Generation:    catalog.TransactionID,
		ConfigCount:   len(catalog.Configs),
	}, nil
}

func (s *Server) failureAtCommittedConfigBoundary(
	fsys vfs.FS,
	catalog *discovery.ConfigCatalog,
) (discovery.ConfigFailure, bool) {
	if catalog == nil || len(catalog.Failures) == 0 || len(s.jsConfigs) == 0 {
		return discovery.ConfigFailure{}, false
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	committed := make(map[string]struct{}, len(s.jsConfigs))
	for configKey := range s.jsConfigs {
		// Synthetic unavailable boundaries are retryable tombstones, not a
		// last-good config value. Treating one as committed would reject every
		// later refresh while that source remains broken, including a refresh
		// that discovers a new usable child config below the boundary.
		if _, unavailable := s.jsUnavailableConfigs[configKey]; unavailable {
			continue
		}
		configDir := configRoutingKeyToPath(configKey)
		if configDir != "" {
			committed[lspLexicalPathID(configDir, caseSensitive)] = struct{}{}
		}
	}
	for _, failure := range catalog.Failures {
		if _, exists := committed[lspLexicalPathID(failure.Directory, caseSensitive)]; exists {
			return failure, true
		}
	}
	return discovery.ConfigFailure{}, false
}

func lspLexicalPathID(filePath string, caseSensitive bool) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", caseSensitive))
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
) (*lspDiscoveredConfigSnapshot, error) {
	if catalog == nil {
		return nil, errors.New("cannot prepare a nil config catalog")
	}
	snapshot := &lspDiscoveredConfigSnapshot{
		configs:            make(map[string]config.RslintConfig, len(catalog.Configs)),
		tsConfigPaths:      make(map[string][]string, len(catalog.Configs)),
		configKeyByPath:    make(map[string]string, len(catalog.Configs)),
		unavailableConfigs: make(map[string]struct{}),
		generation:         catalog.TransactionID,
		usableLastGood:     true,
	}
	seenConfigDirs := make(map[string]string, len(catalog.Configs))
	for _, configDir := range catalog.ConfigDirectories() {
		entries := catalog.Configs[configDir]
		if err := config.ValidateConfig(entries); err != nil {
			return nil, fmt.Errorf("invalid discovered config for %q: %w", configDir, err)
		}
		configID := lspFilesystemPathID(configDir, fsys)
		if previous, exists := seenConfigDirs[configID]; exists {
			return nil, fmt.Errorf(
				"discovered config contains duplicate directories %q and %q",
				previous,
				configDir,
			)
		}
		seenConfigDirs[configID] = configDir
		paths, err := resolveTsConfigPathsWithFS(entries, configDir, fsys)
		if err != nil {
			return nil, fmt.Errorf("resolve tsconfig paths for %q: %w", configDir, err)
		}
		snapshot.configs[configDir] = append(config.RslintConfig(nil), entries...)
		snapshot.tsConfigPaths[configDir] = paths
		normalizedDir := tspath.NormalizePath(configDir)
		snapshot.configKeyByPath[normalizedDir] = configDir
	}

	// A failed candidate still blocks the JSON fallback when no usable JS
	// ancestor can own its subtree. The shared coordinator omits failures from
	// its effective catalog (which is correct for CLI/API parent fallback), so
	// LSP materializes only the outermost such directories as retryable empty
	// boundaries. A usable descendant remains in the same map and wins normal
	// nearest-owner lookup below that boundary.
	for _, configDir := range unavailableConfigBoundaryDirectories(fsys, catalog) {
		if _, exists := snapshot.configs[configDir]; exists {
			continue
		}
		snapshot.configs[configDir] = config.RslintConfig{}
		snapshot.tsConfigPaths[configDir] = nil
		snapshot.configKeyByPath[tspath.NormalizePath(configDir)] = configDir
		snapshot.unavailableConfigs[configDir] = struct{}{}
	}
	// Build all ownership and unavailable boundaries before collecting any
	// .gitignore source. This prevents a parent config from reading into a child
	// boundary merely because that child happened to be processed later.
	snapshot.ownerResolver = config.NewConfigOwnerResolver(snapshot.configs, fsys)
	for configDir, entries := range snapshot.configs {
		snapshot.configs[configDir] = config.ConfigWithGitignoreWithBoundaries(
			entries,
			configDir,
			fsys,
			nil,
			snapshot.ownerResolver.ChildConfigDirs(configDir),
		)
	}

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
	jsonCWD := tspath.NormalizePath(s.cwd)
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
// directories that have no successfully loaded JS ancestor. Nested failures
// covered by one such boundary are redundant; failures with a usable ancestor
// use the coordinator's ordinary parent fallback instead.
func unavailableConfigBoundaryDirectories(fsys vfs.FS, catalog *discovery.ConfigCatalog) []string {
	if catalog == nil || len(catalog.Failures) == 0 {
		return nil
	}
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	byIdentity := make(map[string]string, len(catalog.Failures))
	for _, failure := range catalog.Failures {
		directory := tspath.NormalizePath(failure.Directory)
		if directory == "" {
			continue
		}
		// This check is deliberately lexical. A failed lexical candidate blocks
		// the coordinator's canonical fallback, so a successful config reached
		// only through realpath must not hide the unavailable boundary either.
		if hasUsableLexicalConfigAncestor(directory, catalog.Configs, caseSensitive) {
			continue
		}
		identity := string(tspath.ToPath(directory, "", caseSensitive))
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
			if pathStringsEqual(candidate, boundary, caseSensitive) ||
				tspath.StartsWithDirectory(candidate, boundary, caseSensitive) {
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

func hasUsableLexicalConfigAncestor(
	directory string,
	configs map[string]config.RslintConfig,
	caseSensitive bool,
) bool {
	for configDir := range configs {
		configDir = tspath.NormalizePath(configDir)
		if pathStringsEqual(directory, configDir, caseSensitive) ||
			tspath.StartsWithDirectory(directory, configDir, caseSensitive) {
			return true
		}
	}
	return false
}

func (s *Server) commitDiscoveredConfigSnapshot(ctx context.Context, snapshot *lspDiscoveredConfigSnapshot) {
	// No fallible work remains after the client commits its staged plugin host.
	// The serialized dispatch loop makes this map swap atomic to every document
	// and code-action handler.
	s.invalidateOpenDocumentDiagnostics()
	s.jsConfigs = snapshot.configs
	s.tsConfigPathsByConfig = snapshot.tsConfigPaths
	s.jsConfigOwnerResolver = snapshot.ownerResolver
	s.jsConfigKeyByPath = snapshot.configKeyByPath
	s.jsUnavailableConfigs = snapshot.unavailableConfigs
	s.jsonConfig = snapshot.jsonConfig
	s.rslintConfigPath = snapshot.jsonConfigPath
	s.tsConfigPaths = snapshot.jsonTsConfigPaths
	s.eslintPluginConfigGeneration = snapshot.generation
	s.eslintPluginRules = eslintPluginRuleSet(snapshot.eslintPluginEntries)
	s.configDiscoveryV2HasLastGood = snapshot.usableLastGood
	s.configGenerationFS = snapshot.configGenerationFS
	config.RegisterEslintPluginRules(snapshot.eslintPluginEntries)
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
