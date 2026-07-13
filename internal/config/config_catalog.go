package config

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// AutoJSConfigFileNames is the automatic-discovery priority. Explicit config
// paths are not restricted to these names or extensions.
var AutoJSConfigFileNames = []string{
	"rslint.config.js",
	"rslint.config.mjs",
	"rslint.config.ts",
	"rslint.config.mts",
}

type ConfigDiscoveryMode uint8

const (
	ConfigDiscoveryAuto ConfigDiscoveryMode = iota
	ConfigDiscoveryExplicit
)

// DiscoveryFile is a file target that may participate in config discovery.
// CanonicalPath is consulted only when the complete lexical ancestry contains
// no config candidate.
type DiscoveryFile struct {
	Path          string
	CanonicalPath string
	// Explicit distinguishes a literal CLI/API/LSP target from a file found by
	// a directory or glob walk. Only literal files get the config-global-ignore
	// ownership exception.
	Explicit bool
}

// ConfigDiscoveryRequest describes one immutable config-catalog build.
// Directories are recursive discovery roots unless LimitDirectoryWalkToFiles
// bounds them to an already-expanded target set. When ImplicitCWD is true and
// no files or directories are supplied, CWD is used as the sole directory root.
type ConfigDiscoveryRequest struct {
	CWD                string
	Mode               ConfigDiscoveryMode
	ExplicitConfigPath string
	Files              []DiscoveryFile
	Directories        []string
	// LimitDirectoryWalkToFiles is used when a host has already expanded its
	// directory/glob inputs to an exact target set. Only target-ancestor branches
	// can govern that request. CLI/LSP directory roots leave this false and retain
	// the full recursive catalog walk, including mixed CLI file+directory input.
	LimitDirectoryWalkToFiles bool
	ImplicitCWD               bool
	Fresh                     bool
	SingleThreaded            bool
}

type ConfigLoadCandidate struct {
	ID              string `json:"id"`
	ConfigPath      string `json:"configPath"`
	ConfigDirectory string `json:"configDirectory"`
}

type ConfigLoadBatchRequest struct {
	ProtocolVersion int                   `json:"protocolVersion"`
	TransactionID   string                `json:"transactionId"`
	LoadMode        ConfigModuleLoadMode  `json:"loadMode"`
	SingleThreaded  bool                  `json:"singleThreaded,omitempty"`
	Candidates      []ConfigLoadCandidate `json:"candidates"`
}

const ConfigDiscoveryProtocolVersion = 1

type ConfigModuleLoadMode string

const (
	ConfigModuleLoadCached ConfigModuleLoadMode = "cached"
	ConfigModuleLoadFresh  ConfigModuleLoadMode = "fresh"
)

type ConfigModuleError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ConfigLoadResult struct {
	ID                string              `json:"id"`
	Status            string              `json:"status"`
	Entries           RslintConfig        `json:"entries,omitempty"`
	SourceFingerprint string              `json:"sourceFingerprint,omitempty"`
	EslintPlugins     []EslintPluginEntry `json:"eslintPlugins,omitempty"`
	HasPluginConfig   bool                `json:"hasPluginConfig,omitempty"`
	Error             *ConfigModuleError  `json:"error,omitempty"`
}

type ConfigLoadBatchResponse struct {
	TransactionID string             `json:"transactionId"`
	Results       []ConfigLoadResult `json:"results"`
}

// ConfigModuleLoader is the only JavaScript-aware boundary in config
// discovery. Implementations evaluate and normalize modules; Go validates the
// returned batch envelope and config entries before using them.
type ConfigModuleLoader interface {
	LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error)
}

type ConfigActivationRequest struct {
	ProtocolVersion    int      `json:"protocolVersion"`
	TransactionID      string   `json:"transactionId"`
	EffectiveConfigIDs []string `json:"effectiveConfigIds"`
}

type ConfigActivationResponse struct {
	TransactionID       string              `json:"transactionId"`
	EslintPluginEntries []EslintPluginEntry `json:"eslintPluginEntries"`
}

// ConfigModuleActivator is optional. Native adapters implement it by asking
// the Node host to summarize exactly the effective IDs, which also closes the
// source-fingerprint race before a plugin host is prepared.
type ConfigModuleActivator interface {
	ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error)
}

type ConfigSource struct {
	CandidateID     string
	Path            string
	Directory       string
	Fingerprint     string
	EslintPlugins   []EslintPluginEntry
	HasPluginConfig bool
	ExplicitOnly    bool
	ExplicitConfig  bool
}

type ConfigFailure struct {
	Path      string
	Directory string
	Kind      string
	Message   string
}

type ConfigDiscoveryStats struct {
	DirectoriesVisited int
	CandidatesFound    int
	ConfigsRequested   int
	ConfigsLoaded      int
	DirectoriesPruned  int
}

// ConfigCatalog is the deterministic snapshot produced by one build. Configs
// contains only effective, successfully loaded ownership boundaries. Requested
// failures remain in Failures/Stats; ignored or unreachable candidates are
// omitted because their modules are never requested.
type ConfigCatalog struct {
	Generation         uint64
	TransactionID      string
	Configs            map[string]RslintConfig
	Sources            map[string]ConfigSource
	EffectiveConfigIDs []string
	EslintPlugins      []EslintPluginEntry
	Scopes             map[string]LintDiscoveryScope
	Failures           []ConfigFailure
	Stats              ConfigDiscoveryStats
	Resolver           *ConfigOwnerResolver
}

func (catalog *ConfigCatalog) ConfigDirectories() []string {
	if catalog == nil {
		return nil
	}
	directories := make([]string, 0, len(catalog.Configs))
	for directory := range catalog.Configs {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	return directories
}

var ErrAllConfigsFailed = errors.New("all discovered JavaScript configs failed to load")

type ConfigDiscoveryProtocolError struct {
	Message string
}

func (err *ConfigDiscoveryProtocolError) Error() string {
	return "config loader protocol: " + err.Message
}

// ConfigDiscoverySession serializes builds over one VFS and loader. A CLI/API
// request typically performs one build; an LSP can retain the session and use
// monotonically increasing generations for reloads.
type ConfigDiscoverySession struct {
	fs     vfs.FS
	loader ConfigModuleLoader
	id     uint64

	mu         sync.Mutex
	generation uint64
}

var configDiscoverySessionSequence atomic.Uint64

func NewConfigDiscoverySession(fsys vfs.FS, loader ConfigModuleLoader) *ConfigDiscoverySession {
	return &ConfigDiscoverySession{
		fs:     fsys,
		loader: loader,
		id:     configDiscoverySessionSequence.Add(1),
	}
}

func (session *ConfigDiscoverySession) Build(ctx context.Context, request ConfigDiscoveryRequest) (*ConfigCatalog, error) {
	if session == nil || session.fs == nil {
		return nil, errors.New("config discovery requires a filesystem")
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	session.generation++

	builder := configCatalogBuilder{
		ctx:           ctx,
		fs:            session.fs,
		loader:        session.loader,
		request:       request,
		generation:    session.generation,
		transactionID: fmt.Sprintf("config-discovery-%d-%d", session.id, session.generation),
		loadStates:    make(map[string]*configLoadState),
		configs:       make(map[string]RslintConfig),
		sources:       make(map[string]ConfigSource),
		scopes:        make(map[string]LintDiscoveryScope),
		failureByPath: make(map[string]ConfigFailure),
	}
	return builder.build()
}

func normalizeDiscoveryPath(path string, cwd string) string {
	if tspath.PathIsAbsolute(path) {
		return tspath.NormalizePath(path)
	}
	return tspath.ResolvePath(cwd, path)
}

func configDiscoveryProtocolError(format string, args ...any) error {
	return &ConfigDiscoveryProtocolError{Message: fmt.Sprintf(format, args...)}
}
