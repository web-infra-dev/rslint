package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
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
	ProtocolVersion int                       `json:"protocolVersion"`
	Reason          string                    `json:"reason"`
	ConfigPath      optionalConfigRefreshPath `json:"configPath"`
}

// optionalConfigRefreshPath preserves the distinction between an absent
// configPath (automatic discovery) and an invalid explicit null value.
type optionalConfigRefreshPath struct {
	Value   string
	Present bool
}

func (path *optionalConfigRefreshPath) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return errors.New("configPath must be a string when present")
	}
	var value string
	if err := stdjson.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("configPath must be a string: %w", err)
	}
	path.Value = value
	path.Present = true
	return nil
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
	ProtocolVersion int    `json:"protocolVersion"`
	TransactionID   string `json:"transactionId"`
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
	loader.frontierSeen = true
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

type preparedLSPConfigSnapshot struct {
	configs                  map[string]config.RslintConfig
	tsConfigPaths            map[string][]string
	ownerIndex               *target.OwnerIndex
	unavailableConfigs       map[string]struct{}
	fallbackConfig           config.RslintConfig
	fallbackConfigOwnerIndex *target.OwnerIndex
	fallbackConfigDirectory  string
	transactionID            string
	usableLastGood           bool
}

// lspDiscoveredConfigSnapshot is commit-ready: its configuration state and
// exact rule catalog always belong to the same Node transaction.
type lspDiscoveredConfigSnapshot struct {
	preparedLSPConfigSnapshot
	ruleCatalog                *rule.Catalog
	fileConfigResolvers        map[string]*config.FileConfigResolver
	fallbackFileConfigResolver *config.FileConfigResolver
	shadowedPluginRules        []string
}

func (s *Server) handleConfigRefresh(ctx context.Context, request configRefreshRequest) (configRefreshResponse, error) {
	if request.ProtocolVersion != discovery.ConfigDiscoveryProtocolVersion {
		return configRefreshResponse{}, fmt.Errorf(
			"%w: unsupported config refresh protocol %d; expected %d",
			lsproto.ErrorCodeInvalidParams,
			request.ProtocolVersion,
			discovery.ConfigDiscoveryProtocolVersion,
		)
	}
	switch request.Reason {
	case "initial", "config-change", "gitignore-change", "dependency-change":
	default:
		return configRefreshResponse{}, fmt.Errorf(
			"%w: unsupported config refresh reason %q",
			lsproto.ErrorCodeInvalidParams,
			request.Reason,
		)
	}
	if s.fs == nil {
		return configRefreshResponse{}, errors.New("config refresh requires a filesystem")
	}
	if err := ctx.Err(); err != nil {
		return configRefreshResponse{}, err
	}

	configPath := ""
	if request.ConfigPath.Present {
		var err error
		configPath, err = normalizeLSPExplicitConfigPath(request.ConfigPath.Value)
		if err != nil {
			return configRefreshResponse{}, fmt.Errorf("%w: %w", lsproto.ErrorCodeInvalidParams, err)
		}
	}
	if !s.configRefreshInitialized {
		if request.Reason != "initial" {
			return configRefreshResponse{}, fmt.Errorf(
				"%w: the first config refresh must use reason %q",
				lsproto.ErrorCodeInvalidParams,
				"initial",
			)
		}
		s.configRefreshInitialized = true
		s.configRefreshConfigPath = configPath
	} else {
		lockedExplicit := s.configRefreshConfigPath != ""
		if request.ConfigPath.Present != lockedExplicit ||
			(lockedExplicit && !lspConfigPathsEqual(configPath, s.configRefreshConfigPath, s.fs)) {
			return configRefreshResponse{}, fmt.Errorf(
				"%w: config path selection is fixed for this server process; restart the server to change it",
				lsproto.ErrorCodeInvalidParams,
			)
		}
	}
	// Set this before doing fallible discovery: a failed initial transaction is
	// still proof that the client installed the reverse handlers, and a later
	// config or config-scoped .gitignore event must retry while keeping last-good.
	s.configDiscoveryActive = true
	return s.refreshConfig(ctx)
}

func normalizeLSPExplicitConfigPath(configPath string) (string, error) {
	if configPath == "" {
		return "", errors.New("configPath must not be empty")
	}
	if strings.IndexByte(configPath, 0) >= 0 {
		return "", errors.New("configPath must not contain NUL")
	}
	if strings.HasPrefix(strings.ToLower(configPath), "file:") {
		return "", errors.New("configPath must be a native filesystem path, not a URI")
	}
	if !tspath.PathIsAbsolute(configPath) {
		return "", fmt.Errorf("configPath must be absolute: %q", configPath)
	}
	normalized := tspath.NormalizePath(configPath)
	if discovery.IsSupportedConfigModulePath(normalized) {
		return normalized, nil
	}
	return "", fmt.Errorf("configPath must name a JS/TS config module: %q", configPath)
}

func lspConfigPathsEqual(left string, right string, fsys vfs.FS) bool {
	caseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	return lspLexicalPathID(left, caseSensitive) == lspLexicalPathID(right, caseSensitive)
}

// refreshConfig rebuilds the config transaction selected by the client's
// initial refresh. Internal file watchers call this directly so they cannot
// reinterpret an omitted wire configPath as automatic discovery.
func (s *Server) refreshConfig(ctx context.Context) (configRefreshResponse, error) {
	if !s.configRefreshInitialized {
		return configRefreshResponse{}, errors.New("config refresh source is not initialized")
	}

	// cachedvfs is intentionally scoped to one generation. A long-lived cache
	// would make file creation/deletion and .gitignore edits invisible to later
	// refreshes; retaining the committed resolver retains only the last-good
	// generation's path identity snapshot.
	snapshotFS := newConfigSnapshotFS(bundled.WrapFS(cachedvfs.From(s.fs)))
	loader := &lspConfigModuleLoader{server: s}
	cwd := tspath.NormalizePath(s.cwd)
	var configCatalog *discovery.ConfigCatalog
	var err error
	if s.configRefreshConfigPath == "" {
		configCatalog, err = discovery.DiscoverAutomatic(ctx, snapshotFS, loader, discovery.ConfigDiscoveryRequest{
			CWD:         cwd,
			ImplicitCWD: true,
			Fresh:       true,
		})
	} else {
		configCatalog, err = discovery.LoadExplicitConfig(ctx, snapshotFS, loader, discovery.ExplicitConfigRequest{
			CWD:        cwd,
			ConfigPath: s.configRefreshConfigPath,
			Fresh:      true,
		})
	}
	recoveredUnavailable := false
	var unavailableCause error
	if err != nil {
		if !errors.Is(err, discovery.ErrAllConfigsFailed) || s.configDiscoveryHasLastGood {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, loader.transactionID, err)
		}
		// A broken config on first startup must not tear down the language
		// client, and the workspace fallback must not lint through that broken
		// boundary. Commit a real empty Node generation plus unavailable Go
		// ownership boundaries. This is intentionally not "last-good": a later
		// refresh may replace it with another recovery snapshot until one full
		// usable catalog has committed.
		unavailableCause = err
		var allFailed *discovery.AllConfigsFailedError
		if !errors.As(err, &allFailed) || len(allFailed.Failures) == 0 || loader.transactionID == "" {
			return configRefreshResponse{}, s.abortFailedConfigRefresh(
				ctx,
				loader,
				loader.transactionID,
				errors.Join(unavailableCause, errors.New("all-config-failed recovery has no typed failure catalog")),
			)
		}
		configCatalog = &discovery.ConfigCatalog{
			TransactionID:      loader.transactionID,
			Configs:            make(map[string]config.RslintConfig),
			EffectiveConfigIDs: []string{},
			Failures:           append([]discovery.ConfigFailure(nil), allFailed.Failures...),
			Explicit:           s.configRefreshConfigPath != "",
		}
		recoveredUnavailable = true
	} else if s.configDiscoveryHasLastGood {
		if failure, invalidates := s.failureAtCommittedConfigBoundary(snapshotFS, configCatalog); invalidates {
			err = fmt.Errorf(
				"config refresh failed at last-good boundary %q: %s",
				failure.Directory,
				failure.Message,
			)
			return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, configCatalog.TransactionID, err)
		}
	}

	preparedSnapshot, err := s.prepareDiscoveredConfigSnapshot(snapshotFS, configCatalog)
	if err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, configCatalog.TransactionID, err)
	}
	if recoveredUnavailable && len(preparedSnapshot.unavailableConfigs) == 0 {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(
			ctx,
			loader,
			configCatalog.TransactionID,
			errors.Join(unavailableCause, errors.New("all-config-failed recovery has no valid failure boundary")),
		)
	}
	// The candidate already contains materialized .gitignore patterns. A later
	// filesystem mutation belongs to the next transaction and cannot alter this
	// catalog while Node activation and commit complete.
	if err := loader.ensureActivated(ctx, configCatalog); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, configCatalog.TransactionID, err)
	}
	snapshot, err := completeDiscoveredConfigSnapshot(
		preparedSnapshot,
		configCatalog.EslintPlugins,
		snapshotFS,
	)
	if err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(
			ctx,
			loader,
			configCatalog.TransactionID,
			err,
		)
	}
	// Cancellation is honored throughout discovery and activation. Once commit
	// begins, finish the local two-phase control exchange under a short bounded
	// context so cancellation cannot leave Node committed while Go silently
	// abandons the matching catalog.
	if err := ctx.Err(); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, configCatalog.TransactionID, err)
	}
	controlCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), configTransactionControlTimeout)
	defer cancel()
	if err := loader.commit(controlCtx, configCatalog.TransactionID); err != nil {
		return configRefreshResponse{}, s.abortFailedConfigRefresh(ctx, loader, configCatalog.TransactionID, err)
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
	for _, failure := range configCatalog.Failures {
		log.Printf("[rslint] Skipped config %s: %s", failure.Path, failure.Message)
	}
	return configRefreshResponse{
		TransactionID: configCatalog.TransactionID,
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
		configDir := tspath.NormalizePath(configKey)
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
	configCatalog *discovery.ConfigCatalog,
) (*preparedLSPConfigSnapshot, error) {
	if configCatalog == nil {
		return nil, errors.New("cannot prepare a nil config catalog")
	}
	ruleCatalog := rules.All()
	snapshot := &preparedLSPConfigSnapshot{
		configs:            make(map[string]config.RslintConfig, len(configCatalog.Configs)),
		tsConfigPaths:      make(map[string][]string, len(configCatalog.Configs)),
		unavailableConfigs: make(map[string]struct{}),
		transactionID:      configCatalog.TransactionID,
		// An empty catalog is a successfully committed absence of a module
		// config, not a usable module-config last-good generation. If a broken
		// config appears later, it must establish an unavailable boundary rather
		// than exposing the workspace fallback beneath that new boundary.
		usableLastGood: len(configCatalog.Configs) > 0,
	}
	seenConfigDirs := make(map[string]string, len(configCatalog.Configs))
	for _, configDir := range configCatalog.ConfigDirectories() {
		entries := configCatalog.Configs[configDir]
		if err := config.ValidateConfig(entries); err != nil {
			return nil, fmt.Errorf("invalid discovered config for %q: %w", configDir, err)
		}
		normalizedEntries, err := validateRuleOptionsForConfig(entries, configDir, ruleCatalog)
		if err != nil {
			return nil, err
		}
		entries = normalizedEntries
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
	}

	// A failed candidate still blocks the workspace fallback when no usable
	// config ancestor can own its subtree. The shared coordinator omits failures from
	// its effective catalog (which is correct for CLI/API parent fallback), so
	// LSP materializes only the outermost such directories as retryable empty
	// boundaries. A usable descendant remains in the same map and wins normal
	// nearest-owner lookup below that boundary.
	for _, configDir := range unavailableConfigBoundaryDirectories(fsys, configCatalog) {
		if _, exists := snapshot.configs[configDir]; exists {
			continue
		}
		snapshot.configs[configDir] = config.RslintConfig{}
		snapshot.tsConfigPaths[configDir] = nil
		snapshot.unavailableConfigs[configDir] = struct{}{}
	}
	// Build the owner index only after all successful and unavailable
	// boundaries are present. Config values remain owned by this LSP snapshot;
	// the target package retains only routing and frozen path-space state.
	// Successful automatic catalog entries already contain the Git sources read
	// by their discovery generation; unavailable boundaries intentionally remain
	// empty and only stop workspace fallback ownership.
	snapshot.ownerIndex = target.NewOwnerIndex(snapshot.configs, fsys)
	if configCatalog.Explicit {
		// An explicit config is invocation-wide and owns every editor target, so
		// no workspace fallback participates in this mode.
		return snapshot, nil
	}

	// The empty workspace fallback owns only portions of the workspace not
	// handed to a discovered or unavailable config boundary. Include a synthetic
	// cwd owner solely to calculate its direct source boundaries; discovered
	// config ownership remains unchanged.
	fallbackBoundaryConfigs := make(map[string]config.RslintConfig, len(snapshot.configs)+1)
	for configDir, entries := range snapshot.configs {
		fallbackBoundaryConfigs[configDir] = entries
	}
	fallbackCWD := tspath.NormalizePath(s.cwd)
	fallbackBoundaryConfigs[fallbackCWD] = config.RslintConfig{}
	fallbackBoundaryResolver := target.NewOwnerIndex(fallbackBoundaryConfigs, fsys)
	snapshot.fallbackConfig = config.ConfigWithGitignoreWithBoundaries(
		config.RslintConfig{},
		fallbackCWD,
		fsys,
		nil,
		fallbackBoundaryResolver.ChildOwnerDirectories(fallbackCWD),
	)
	snapshot.fallbackConfigOwnerIndex = target.NewOwnerIndex(
		map[string]config.RslintConfig{fallbackCWD: snapshot.fallbackConfig},
		fsys,
	)
	snapshot.fallbackConfigDirectory = fallbackCWD
	return snapshot, nil
}

func completeDiscoveredConfigSnapshot(
	prepared *preparedLSPConfigSnapshot,
	plugins []config.EslintPluginEntry,
	fsys vfs.FS,
) (*lspDiscoveredConfigSnapshot, error) {
	if prepared == nil {
		panic("prepared LSP config snapshot is required")
	}
	ruleCatalog, shadowedPluginRules := deriveLSPRuleCatalog(rules.All(), plugins)
	fileConfigResolvers := make(map[string]*config.FileConfigResolver, len(prepared.configs))
	for _, configDirectory := range sortedConfigDirectories(prepared.configs) {
		resolver, err := config.NewFileConfigResolverWithPathSpaces(
			prepared.configs[configDirectory],
			configDirectory,
			fsys,
			prepared.ownerIndex.PathSpaces(),
			ruleCatalog,
		)
		if err != nil {
			return nil, fmt.Errorf("prepare file config resolver for %q: %w", configDirectory, err)
		}
		fileConfigResolvers[configDirectory] = resolver
	}

	var fallbackFileConfigResolver *config.FileConfigResolver
	if prepared.fallbackConfigOwnerIndex != nil {
		fallbackCWD := prepared.fallbackConfigDirectory
		resolver, err := config.NewFileConfigResolverWithPathSpaces(
			prepared.fallbackConfig,
			fallbackCWD,
			fsys,
			prepared.fallbackConfigOwnerIndex.PathSpaces(),
			rules.All(),
		)
		if err != nil {
			return nil, fmt.Errorf("prepare workspace fallback resolver for %q: %w", fallbackCWD, err)
		}
		fallbackFileConfigResolver = resolver
	}

	return &lspDiscoveredConfigSnapshot{
		preparedLSPConfigSnapshot:  *prepared,
		ruleCatalog:                ruleCatalog,
		fileConfigResolvers:        fileConfigResolvers,
		fallbackFileConfigResolver: fallbackFileConfigResolver,
		shadowedPluginRules:        shadowedPluginRules,
	}, nil
}

func sortedConfigDirectories(configs map[string]config.RslintConfig) []string {
	directories := make([]string, 0, len(configs))
	for directory := range configs {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	return directories
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
	s.invalidateLintProjectCaches()
	s.jsConfigs = snapshot.configs
	s.tsConfigPathsByConfig = snapshot.tsConfigPaths
	s.jsConfigOwnerIndex = snapshot.ownerIndex
	s.jsFileConfigResolvers = snapshot.fileConfigResolvers
	s.jsUnavailableConfigs = snapshot.unavailableConfigs
	s.fallbackConfig = snapshot.fallbackConfig
	s.fallbackConfigOwnerIndex = snapshot.fallbackConfigOwnerIndex
	s.fallbackFileConfigResolver = snapshot.fallbackFileConfigResolver
	s.eslintPluginConfigGeneration = snapshot.transactionID
	s.ruleCatalog = snapshot.ruleCatalog
	s.configDiscoveryHasLastGood = snapshot.usableLastGood
	s.configSnapshotIncludesGitignore = true
	for _, ruleName := range snapshot.shadowedPluginRules {
		fmt.Fprintf(
			os.Stderr,
			"rslint: plugin rule %q is shadowed by a built-in rule of the same name; using the built-in.\n",
			ruleName,
		)
	}
	log.Printf("[rslint] Committed Go-discovered JS/TS config catalog (%d config files)", len(snapshot.configs))
	if err := s.RefreshDiagnostics(ctx); err != nil {
		log.Printf("[rslint] Failed to refresh diagnostics after config refresh: %v", err)
	}
}

func resolveTsConfigPathsWithFS(cfg config.RslintConfig, cwd string, fsys vfs.FS) ([]string, error) {
	paths, err := config.ResolveTsConfigPaths(cfg, cwd, fsys)
	if err != nil {
		return nil, err
	}
	for index, projectPath := range paths {
		paths[index] = tspath.NormalizePath(projectPath)
	}
	return paths, nil
}
