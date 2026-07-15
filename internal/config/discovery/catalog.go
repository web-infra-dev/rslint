package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// autoJSConfigFileNames is the hard-coded automatic-discovery priority.
// Explicit config paths are not restricted to these names or extensions.
var autoJSConfigFileNames = [...]string{
	"rslint.config.js",
	"rslint.config.mjs",
	"rslint.config.ts",
	"rslint.config.mts",
}

// AutoJSConfigFileNames returns a copy so callers cannot mutate discovery's
// process-wide filename policy.
func AutoJSConfigFileNames() []string {
	return append([]string(nil), autoJSConfigFileNames[:]...)
}

type ConfigDiscoveryMode uint8

const (
	ConfigDiscoveryAuto ConfigDiscoveryMode = iota
	ConfigDiscoveryExplicit
	// ConfigDiscoveryInline uses OverrideConfig as one invocation-wide flat
	// config without asking the JavaScript module host to load a file. The high-
	// level API uses it for overrideConfigFile:true while Go still owns input
	// planning and target discovery.
	ConfigDiscoveryInline
)

type DiscoveredTarget struct {
	// Path is the lexical spelling reached by the search. Config lookup and
	// final result deduplication intentionally do not replace it with Realpath.
	Path            string
	ConfigDirectory string
	Explicit        bool
	// MergedConfig is the immutable result of the owner's sole ConfigArray
	// selection, including the product-level .gitignore layer. Downstream
	// native/plugin lint and fix passes reuse it and never execute live matchers
	// again.
	MergedConfig *rslintconfig.MergedConfig `json:"-"`
}

type ExplicitInputStatus string

const (
	ExplicitInputConfigured   ExplicitInputStatus = "configured"
	ExplicitInputIgnored      ExplicitInputStatus = "ignored"
	ExplicitInputUnconfigured ExplicitInputStatus = "unconfigured"
	ExplicitInputExternal     ExplicitInputStatus = "external"
)

// ExplicitInputResult retains direct-file provenance even when the file is
// ignored or unconfigured. CLI/API adapters use it for ESLint-compatible
// per-file warnings without widening the actual target plan.
type ExplicitInputResult struct {
	Path            string
	ConfigDirectory string
	Status          ExplicitInputStatus
	Order           int
}

// ConfigDiscoveryRequest describes one immutable config-catalog build. Inputs
// are the caller's raw lint patterns; adapters pass "." explicitly for a
// no-argument invocation so discovery never guesses an implicit root.
type ConfigDiscoveryRequest struct {
	CWD                     string
	Mode                    ConfigDiscoveryMode
	ExplicitConfigPath      string
	Inputs                  []string
	OverrideConfig          rslintconfig.RslintConfig
	CollectTargets          bool
	GlobInputPaths          bool
	ErrorOnUnmatchedPattern bool
	// AllowMissingConfig is reserved for the CLI/LSP JSON/JSONC product
	// fallback. Missing JS config remains fatal whenever at least one JS config
	// is reached in the same build.
	AllowMissingConfig bool
	// RetainUnconfiguredAreas lets catalog-only consumers keep successfully
	// discovered JS owners while leaving directories with no JS owner available
	// to the retained JSON/JSONC resolver. LSP uses this for mixed workspaces;
	// CLI/API lintFiles keep ESLint's fatal mixed-owner behavior.
	RetainUnconfiguredAreas bool
	// CollectConfigFailures turns nearest-config load failures into unavailable
	// catalog boundaries and continues independent sibling searches. LSP uses
	// this to publish a partial first-generation snapshot; CLI/API remain
	// fail-fast like ESLint lintFiles.
	CollectConfigFailures bool
	// ProbeRootConfig asks catalog-only consumers such as LSP to resolve the
	// config governing CWD even when the workspace has no entries. Lint target
	// discovery leaves this false to preserve ESLint findFiles semantics.
	ProbeRootConfig bool
	// LookupPaths asks a catalog-only consumer to resolve nearest configs for
	// exact paths without turning them into lint inputs. LSP uses open documents
	// here so a nested config hidden behind a parent/default ignore is still
	// selected exactly as ESLint lintText({filePath}) would select it. Paths need
	// not exist; only their lexical parent directories participate.
	LookupPaths    []string
	Fresh          bool
	SingleThreaded bool
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

// ConfigCatalog is the deterministic snapshot produced by one successful
// build. Configs contains only effective, successfully loaded ownership
// boundaries. Failures is populated only when CollectConfigFailures asks a
// catalog consumer to retain independent successful owners while representing
// failed nearest-config boundaries as unavailable. Fail-fast callers receive a
// typed AllConfigsFailedError instead.
type ConfigCatalog struct {
	TransactionID      string
	Configs            map[string]rslintconfig.RslintConfig
	EffectiveConfigIDs []string
	EslintPlugins      []rslintconfig.EslintPluginEntry
	Targets            []DiscoveredTarget
	ExplicitInputs     []ExplicitInputResult
	Stats              ConfigDiscoveryStats
	Failures           []ConfigFailure
	// Explicit reports that the catalog came from one explicitly selected
	// invocation-wide flat config rather than hierarchical auto discovery.
	Explicit bool

	predicateCoordinator *predicateCoordinator
	configEvaluators     map[string]*rslintconfig.ConfigEvaluator
}

// ConfigEvaluatorForDirectory returns the committed evaluator created while
// discovering an owner. It carries ConfigArray's exact file/directory caches,
// including live matcher side effects already reached by the search walk.
func (catalog *ConfigCatalog) ConfigEvaluatorForDirectory(basePath string) *rslintconfig.ConfigEvaluator {
	if catalog == nil {
		return nil
	}
	return catalog.configEvaluators[rslintconfig.NormalizeHostPath(basePath)]
}

// ClosePredicateEvaluation cancels the Go-side predicate worker and resolves
// its pending requests. It does not delete the Node transaction session;
// transport owners keep or discard that session according to CLI/API/LSP
// lifecycle rules.
func (catalog *ConfigCatalog) ClosePredicateEvaluation() {
	if catalog == nil || catalog.predicateCoordinator == nil {
		return
	}
	catalog.predicateCoordinator.Close()
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

// AllConfigsFailedError describes the nearest config failure that aborted a
// fail-fast CLI/API build. Catalog consumers that need sibling aggregation use
// ConfigDiscoveryRequest.CollectConfigFailures and ConfigCatalog.Failures.
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

// Build discovers and loads one immutable config catalog. Transaction IDs are
// unique across concurrent builds and native-process restarts so adapters can
// stage load/activate/commit as one operation. The operation itself is
// intentionally stateless; long-lived lifecycle and rollback belong to the
// CLI/API/LSP adapters that own the transport.
func Build(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ConfigDiscoveryRequest) (*ConfigCatalog, error) {
	if fsys == nil {
		return nil, errors.New("config discovery requires a filesystem")
	}
	return buildSearchCatalog(ctx, fsys, loader, request, nextConfigDiscoveryTransactionID())
}

func normalizeDiscoveryPath(path string, cwd string) string {
	if rslintconfig.HostPathIsAbsolute(path) {
		return rslintconfig.NormalizeHostPath(path)
	}
	return rslintconfig.ResolveHostPath(cwd, path)
}

func configDiscoveryProtocolError(format string, args ...any) error {
	return &ConfigDiscoveryProtocolError{Message: fmt.Sprintf(format, args...)}
}
