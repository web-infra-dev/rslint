package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// AutoJSConfigFileNames is the automatic-discovery priority. Explicit config
// paths are not restricted to these names or extensions.
var AutoJSConfigFileNames = []string{
	"rslint.config.js",
	"rslint.config.mjs",
	"rslint.config.ts",
	"rslint.config.mts",
}

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
	CWD         string
	Files       []DiscoveryFile
	Directories []string
	// LimitDirectoryWalkToFiles is used when a host has already expanded its
	// directory/glob inputs to an exact target set. Only target-ancestor branches
	// can govern that request. CLI/LSP directory roots leave this false and retain
	// the full recursive catalog walk, including mixed CLI file+directory input.
	LimitDirectoryWalkToFiles bool
	ImplicitCWD               bool
	Fresh                     bool
	SingleThreaded            bool
}

// ExplicitConfigRequest loads one invocation-wide JS/TS config. It is a
// separate operation from automatic discovery: absence of this request means
// automatic discovery, while absence of config discovery at the adapter means
// the low-level/JSON path.
type ExplicitConfigRequest struct {
	CWD            string
	ConfigPath     string
	Fresh          bool
	SingleThreaded bool
}

type configSource struct {
	CandidateID   string
	CandidatePath string
	ExplicitOnly  bool
}

type ConfigFailure struct {
	Path      string
	Directory string
	Kind      string
	Message   string
}

type ConfigDiscoveryStats struct {
	DirectoriesVisited int
	ConfigsRequested   int
	ConfigsLoaded      int
	DirectoriesPruned  int
}

// ConfigCatalog is the deterministic snapshot produced by one build. Configs
// contains only effective, successfully loaded ownership boundaries. Requested
// failures remain in Failures/Stats; ignored or unreachable candidates are
// omitted because their modules are never requested. Automatic catalogs freeze
// each owner's observed Git projection into Configs before publication;
// explicit catalogs leave invocation-scoped Git collection to their adapter.
type ConfigCatalog struct {
	TransactionID      string
	Configs            map[string]rslintconfig.RslintConfig
	EffectiveConfigIDs []string
	EslintPlugins      []rslintconfig.EslintPluginEntry
	Scopes             map[string]rslintconfig.LintDiscoveryScope
	Failures           []ConfigFailure
	Stats              ConfigDiscoveryStats
	// Explicit reports that the catalog came from one explicitly selected
	// invocation-wide flat config rather than hierarchical auto discovery.
	Explicit bool
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

// AllConfigsFailedError preserves every failed candidate while remaining
// compatible with errors.Is(err, ErrAllConfigsFailed). Adapters can therefore
// render their own diagnostics without parsing the user-facing summary.
type AllConfigsFailedError struct {
	Failures []ConfigFailure
	message  string
}

func (err *AllConfigsFailedError) Error() string {
	if err == nil || err.message == "" {
		return ErrAllConfigsFailed.Error()
	}
	return err.message
}

func (err *AllConfigsFailedError) Unwrap() error {
	return ErrAllConfigsFailed
}

type ConfigDiscoveryProtocolError struct {
	Message string
}

func (err *ConfigDiscoveryProtocolError) Error() string {
	return "config loader protocol: " + err.Message
}

var (
	// A LanguageClient may restart the native LSP process while its extension-side
	// ConfigModuleHost survives. A process nonce prevents the new process from
	// colliding with an uncollected session left by the old one; the sequence keeps
	// concurrent builds within this process deterministic and lock-free.
	configDiscoveryProcessNonce = rand.Text()
	configDiscoverySequence     atomic.Uint64
)

func nextConfigDiscoveryTransactionID() string {
	return fmt.Sprintf(
		"config-discovery-%s-%d",
		configDiscoveryProcessNonce,
		configDiscoverySequence.Add(1),
	)
}

// DiscoverAutomatic discovers and loads one immutable hierarchical config
// catalog. Transaction IDs are unique across concurrent builds and native
// process restarts so adapters can stage load/activate/commit as one operation.
func DiscoverAutomatic(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ConfigDiscoveryRequest) (*ConfigCatalog, error) {
	return buildConfigCatalog(ctx, fsys, loader, request, "")
}

// LoadExplicitConfig loads one exact invocation-wide config path. Automatic
// candidate discovery and Git reachability never participate in this operation.
func LoadExplicitConfig(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ExplicitConfigRequest) (*ConfigCatalog, error) {
	if request.ConfigPath == "" {
		return nil, errors.New("explicit config discovery requires a config path")
	}
	automatic := ConfigDiscoveryRequest{
		CWD:            request.CWD,
		Fresh:          request.Fresh,
		SingleThreaded: request.SingleThreaded,
	}
	return buildConfigCatalog(ctx, fsys, loader, automatic, request.ConfigPath)
}

func buildConfigCatalog(
	ctx context.Context,
	fsys vfs.FS,
	loader ConfigModuleLoader,
	request ConfigDiscoveryRequest,
	explicitConfigPath string,
) (*ConfigCatalog, error) {
	if fsys == nil {
		return nil, errors.New("config discovery requires a filesystem")
	}
	transactionID := nextConfigDiscoveryTransactionID()

	builder := configCatalogBuilder{
		ctx:                 ctx,
		fs:                  fsys,
		loader:              loader,
		request:             request,
		explicitConfigPath:  explicitConfigPath,
		transactionID:       transactionID,
		loadStates:          make(map[string]*configLoadState),
		loadStateByIdentity: make(map[tspath.Path]*configLoadState),
		configs:             make(map[string]rslintconfig.RslintConfig),
		sources:             make(map[string]configSource),
		scopes:              make(map[string]rslintconfig.LintDiscoveryScope),
		failureByPath:       make(map[string]ConfigFailure),
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
