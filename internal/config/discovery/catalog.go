package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"path"
	"sort"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

// AutoJSConfigFileNames is the automatic-discovery priority. Explicit config
// paths are not restricted to these names, but they must use a supported
// JavaScript or TypeScript module extension.
var AutoJSConfigFileNames = []string{
	"rslint.config.js",
	"rslint.config.mjs",
	"rslint.config.ts",
	"rslint.config.mts",
}

// IsSupportedConfigModulePath reports whether an explicit path can be loaded
// as a runtime configuration module. Keep this boundary in discovery so every
// native entry point rejects legacy JSON/JSONC before invoking a host loader.
func IsSupportedConfigModulePath(filePath string) bool {
	switch path.Ext(tspath.NormalizePath(filePath)) {
	case ".js", ".mjs", ".cjs", ".ts", ".mts", ".cts":
		return true
	default:
		return false
	}
}

// DiscoveryFile keeps a frozen target identity and its config-ownership role.
// Only literal files get the authored-global-ignore ownership exception.
type DiscoveryFile struct {
	rslintconfig.PathIdentity
	Explicit bool
	// ScanDirectories lists the inputs that actually expanded this file. Nil
	// associates a non-literal file with all supplied containing roots.
	ScanDirectories []string
}

// ConfigDiscoveryRequest describes one immutable config-catalog build.
// Directories are recursive discovery roots unless LimitDirectoryWalkToFiles
// bounds them to an already-expanded target set. When ImplicitCWD is true and
// no files or directories are supplied, CWD is used as the sole directory root.
type ConfigDiscoveryRequest struct {
	CWD         string
	Files       []DiscoveryFile
	Directories []rslintconfig.DirectoryIdentity
	// FallbackConfig supplies input policy when discovery finds no config owner.
	// It is not a config owner or a load candidate.
	FallbackConfig rslintconfig.RslintConfig
	// LimitDirectoryWalkToFiles is used when a host has already expanded its
	// directory/glob inputs to an exact target set. Only target-ancestor branches
	// can govern that request. CLI/LSP directory roots leave this false and retain
	// the full recursive catalog walk, including mixed CLI file+directory input.
	LimitDirectoryWalkToFiles bool
	// RecursiveDirectories keep literal directory inputs unbounded when they
	// are combined with already-expanded glob candidates.
	RecursiveDirectories []rslintconfig.DirectoryIdentity
	ImplicitCWD          bool
	// AdditionalGitFiles add exact Git chains without participating in config discovery.
	AdditionalGitFiles []rslintconfig.PathIdentity
	Fresh              bool
	SingleThreaded     bool
}

// ExplicitConfigRequest supplies one invocation-wide config array, either
// from a JS/TS module or already resolved by the caller. It is a
// separate operation from automatic discovery; inline arrays do not invoke
// the module loader.
type ExplicitConfigRequest struct {
	CWD        string
	ConfigPath string
	// Config supplies an already-resolved inline array instead of a module.
	Config          rslintconfig.RslintConfig
	ConfigDirectory string
	// ProjectOnly omits target Git-source traversal when the loaded config's
	// project declarations do not depend on per-file service/root/reset options.
	ProjectOnly bool
	// TargetFiles supplies frozen exact targets. With no TargetDirectories, nil
	// means an implicit CWD directory input; a non-nil empty slice means no targets.
	TargetFiles []DiscoveryFile
	// TargetDirectories supplies independent frozen recursive directory inputs.
	// It may be combined with TargetFiles for mixed CLI invocations.
	TargetDirectories []rslintconfig.DirectoryIdentity
	// LimitDirectoryWalkToFiles bounds these roots to the supplied candidates.
	LimitDirectoryWalkToFiles bool
	RecursiveDirectories      []rslintconfig.DirectoryIdentity
	AdditionalGitFiles        []rslintconfig.PathIdentity
	Fresh                     bool
	SingleThreaded            bool
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

// ConfigCatalog freezes authored configs, their ownership routes, and Git
// policy independently for one discovery transaction. Failures are business
// results: even a catalog without successful modules has complete Git policy.
type ConfigCatalog struct {
	TransactionID      string
	GitPolicy          *rslintconfig.GitPolicy
	PathSpaces         *rslintconfig.PathSpaceSnapshot
	TargetFiles        []rslintconfig.PathIdentity
	TargetDirectories  []rslintconfig.DirectoryIdentity
	Configs            map[string]rslintconfig.RslintConfig
	EffectiveConfigIDs []string
	EslintPlugins      []rslintconfig.EslintPluginEntry
	Scopes             map[string]target.OwnerScope
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

// AllConfigsFailedError derives the adapter-facing failure without discarding
// the completed catalog. A request with no candidates is not a load failure.
func (catalog *ConfigCatalog) AllConfigsFailedError() error {
	if catalog == nil || len(catalog.Configs) > 0 || len(catalog.Failures) == 0 {
		return nil
	}
	return allConfigsFailedError(catalog.Failures)
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
	return buildConfigCatalog(ctx, fsys, loader, request, nil)
}

// LoadExplicitConfig loads one exact module or accepts an inline array. Git
// reachability never gates that source, and its location does not set Git roots.
func LoadExplicitConfig(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ExplicitConfigRequest) (*ConfigCatalog, error) {
	if request.Config == nil && request.ConfigPath == "" {
		return nil, errors.New("explicit config discovery requires a config path")
	}
	if request.Config == nil && !IsSupportedConfigModulePath(request.ConfigPath) {
		return nil, fmt.Errorf("explicit config path must name a JS/TS config module: %q", request.ConfigPath)
	}
	automatic := ConfigDiscoveryRequest{
		CWD:                       request.CWD,
		Files:                     request.TargetFiles,
		Directories:               request.TargetDirectories,
		LimitDirectoryWalkToFiles: request.LimitDirectoryWalkToFiles,
		RecursiveDirectories:      request.RecursiveDirectories,
		AdditionalGitFiles:        request.AdditionalGitFiles,
		ImplicitCWD:               request.TargetFiles == nil && request.TargetDirectories == nil,
		Fresh:                     request.Fresh,
		SingleThreaded:            request.SingleThreaded,
	}
	return buildConfigCatalog(ctx, fsys, loader, automatic, &request)
}

func buildConfigCatalog(
	ctx context.Context,
	fsys vfs.FS,
	loader ConfigModuleLoader,
	request ConfigDiscoveryRequest,
	explicit *ExplicitConfigRequest,
) (*ConfigCatalog, error) {
	if fsys == nil {
		return nil, errors.New("config discovery requires a filesystem")
	}
	transactionID := nextConfigDiscoveryTransactionID()

	coordinator := newDiscoveryCoordinator(
		ctx,
		fsys,
		loader,
		request,
		explicit,
		transactionID,
	)
	return coordinator.build()
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
