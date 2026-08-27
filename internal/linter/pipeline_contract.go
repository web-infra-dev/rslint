package linter

import (
	"context"
	"errors"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// MaxFixRounds is the product-wide safety bound. A round is one non-empty
// in-memory change-set application; a final verification observation and the
// optional external commit are not rounds.
const MaxFixRounds = 10

// ReleaseFunc ends the producer lease for one acquired generation. The
// pipeline wraps it with exact-once semantics immediately after acquisition;
// release must not invalidate immutable diagnostics or requested file
// projections already published in an observation result.
type ReleaseFunc func()

// SourceFileSnapshot is one pipeline-owned in-memory override in stable target
// path space. Text is the complete current file representation, including a
// byte order mark when the source medium carries one.
type SourceFileSnapshot struct {
	Path string
	Text string
}

// SourceSnapshot is the immutable in-memory state supplied before generation
// materialization. It contains only files whose current text differs from the
// initial source generation. Its zero value represents the initial state.
// Accessors always return detached data.
type SourceSnapshot struct {
	files []SourceFileSnapshot
	texts map[string]string
}

func (s SourceSnapshot) Files() []SourceFileSnapshot {
	return append([]SourceFileSnapshot(nil), s.files...)
}

func (s SourceSnapshot) Text(path string) (string, bool) {
	if s.texts != nil {
		text, ok := s.texts[path]
		return text, ok
	}
	for _, file := range s.files {
		if file.Path == path {
			return file.Text, true
		}
	}
	return "", false
}

func (s SourceSnapshot) Empty() bool { return len(s.files) == 0 }

// GenerationProvider materializes one immutable lint generation from the
// pipeline-owned in-memory snapshot. The provider may combine it with a
// read-only disk base, an overlay VFS, or editor state, but it never owns or
// advances fix rounds and must not persist the snapshot.
//
// AcquireGeneration and the returned ReleaseFunc are called on RunPipeline's
// goroutine. On an acquisition error, the provider retains responsibility for
// cleaning up resources that it did not successfully publish.
type GenerationProvider interface {
	AcquireGeneration(ctx context.Context, snapshot SourceSnapshot) (Generation, ReleaseFunc, error)
}

// GenerationProviderFunc adapts a function to GenerationProvider.
type GenerationProviderFunc func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error)

func (f GenerationProviderFunc) AcquireGeneration(ctx context.Context, snapshot SourceSnapshot) (Generation, ReleaseFunc, error) {
	if f == nil {
		return Generation{}, nil, errors.New("linter pipeline: generation provider function must not be nil")
	}
	return f(ctx, snapshot)
}

// FinalChangeCommitter projects a non-empty final in-memory delta onto an
// external medium at most once. It never participates in fix rounds. A medium
// that may have partially mutated a path returns an error without confirming
// that path.
type FinalChangeCommitter interface {
	CommitFinalChanges(ctx context.Context, changes []FileChange) (CommitResult, error)
}

// Generation is one immutable, internally coherent lint snapshot. Config
// discovery, Program construction, storage, and protocol state stay in the
// integration that implements GenerationProvider. Once published, the
// provider must not mutate any reachable generation state; the pipeline reads
// it directly rather than pretending a shallow slice copy could make Programs,
// callbacks, or rule environments immutable.
type Generation struct {
	Native NativeGeneration
	Target TargetProjection
	Plugin *PluginGeneration
}

// NativeGeneration is the input understood by the native lint engine.
type NativeGeneration struct {
	Programs         []*program.Program
	TargetsByProgram [][]string
	RulesForFile     RuleHandler
	Cwd              string
	TypeCheck        bool
	SingleThreaded   bool
	Timing           *TimingCollector
}

// TargetProjection binds Program-facing paths to the stable target identity
// shared by diagnostics and FileChange. ReadText exposes the exact complete
// representation in this generation, including a byte order mark when its
// source medium carries one. The pipeline uses that one view for validation,
// fix text snapshots, and inline plugin input. Path must be deterministic and
// side-effect free; ReadText may be called more than once and must always
// return the same generation text.
type TargetProjection struct {
	Path     func(sourcePath string) string
	ReadText func(targetPath string, source ast.SourceFileLike) (string, error)
}

// PluginGeneration is the generation-local projection needed to materialize
// third-party plugin work. WirePath is independent from TargetProjection.Path:
// nil preserves the Program source path used by disk-backed plugin hosts,
// while an overlay-backed host may project it to its public document identity.
// HostReadsInitialText asserts that the external host can read the exact text
// of this initial generation by wire path, allowing a joined observation to
// omit inline text. Detached work and every non-initial in-memory generation
// ignore this optimization and freeze every plugin target from the
// generation's unified text view, so one observation never mixes post-fix
// memory with stale disk. ConfigForFile and WirePath must be deterministic and
// side-effect free for the lifetime of the generation.
type PluginGeneration struct {
	ConfigForFile        EslintPluginFileConfigResolver
	WirePath             func(sourcePath string) string
	HostReadsInitialText bool
}

// ArtifactDemand states which optional observation artifacts producers must
// materialize. Planning changes is a separate request kind, so a final
// verification may still collect edits without scheduling another commit.
type ArtifactDemand struct {
	Native rule.EditDemand
	Plugin rule.EditDemand
	// LintedFiles retains the complete target/source projection for consumers
	// such as the API that return selected files or encode their ASTs.
	LintedFiles bool
}

func (d ArtifactDemand) valid() bool {
	return d.Native.IsValid() && d.Plugin.IsValid()
}

func (d ArtifactDemand) plansAutofixes() bool {
	return d.Native&rule.EditDemandAutofix != 0 || d.Plugin&rule.EditDemandAutofix != 0
}

// PluginExecution defines the relative scheduling of one observation's native
// and third-party work.
type PluginExecution uint8

const (
	// PluginConcurrentJoined starts plugin work before native lint and joins it
	// before releasing the generation. CLI and API use this mode.
	PluginConcurrentJoined PluginExecution = iota
	// PluginAfterNativeJoined detaches plugin work, runs native lint, releases
	// the generation, and then joins plugin work. LSP fix-all uses this mode.
	PluginAfterNativeJoined
	// pluginProgressiveAfterNative is selected only by
	// NewProgressiveLintRequest. Integrations cannot opt an ordinary lint/fix
	// request into a partially executed pipeline.
	pluginProgressiveAfterNative
)

func (m PluginExecution) valid() bool {
	return m <= pluginProgressiveAfterNative
}

// PluginFailurePolicy controls diagnostic completeness after a transport or
// reconstruction failure. Logging and protocol presentation stay with the
// integration through the returned structured outcome.
type PluginFailurePolicy uint8

const (
	// PluginKeepPartialWithSynthetic keeps completed plugin diagnostics and adds
	// a visible error diagnostic for non-cancellation failures.
	PluginKeepPartialWithSynthetic PluginFailurePolicy = iota
	// PluginDiscardOnFailure drops the observation's plugin diagnostics and
	// leaves it native-only. LSP uses this degradation policy.
	PluginDiscardOnFailure
)

func (p PluginFailurePolicy) valid() bool {
	return p <= PluginDiscardOnFailure
}

// ObservationPolicy contains only producer and scheduling semantics shared by
// integrations; it does not name CLI, API, or LSP entrypoints.
type ObservationPolicy struct {
	Demand        ArtifactDemand
	Plugin        PluginExecution
	PluginFailure PluginFailurePolicy
}

// DeferredPluginRun is a single-use, generation-detached enrichment task. It
// retains frozen wire/config/text only; the executor injects its own transport
// under its own timeout and cancellation lifecycle.
type DeferredPluginRun func(
	ctx context.Context,
	dispatcher EslintPluginDispatcher,
) (EslintPluginDispatchOutcome, error)

// ProgressiveDiagnostics binds baseline presentation and asynchronous
// enrichment admission to one request-scoped identity. RunPipeline publishes
// the baseline synchronously only after releasing the generation, then
// conditionally submits enrichment. Submit must register and take lifecycle
// ownership synchronously, return promptly without running the task inline, and
// invoke run exactly once, including cleanup when parentCtx is already canceled.
// Timeout, supersession, transport injection, admission, and presentation belong
// to the implementation.
type ProgressiveDiagnostics interface {
	PublishBaseline(ctx context.Context, diagnostics []rule.RuleDiagnostic)
	Submit(parentCtx context.Context, run DeferredPluginRun)
}

// AutofixPolicy configures bounded apply-and-observe behavior.
type AutofixPolicy struct {
	MaxRounds                int
	VerifyAfterLastRound     bool
	VerificationDemand       ArtifactDemand
	StopOnTargetSyntaxErrors bool
}

type pipelineRequestKind uint8

const (
	pipelineRequestLint pipelineRequestKind = iota
	pipelineRequestProgressiveLint
	pipelineRequestAutofix
)

// PipelineRequest is sealed so integrations select a complete operation
// without coordinating its internal stages or constructing deferred-fix
// states themselves.
type PipelineRequest struct {
	kind       pipelineRequestKind
	provider   GenerationProvider
	commit     *finalChangeCommit
	policy     ObservationPolicy
	autofix    AutofixPolicy
	dispatcher EslintPluginDispatcher
	presenter  ProgressiveDiagnostics
}

type finalChangeCommit struct {
	committer FinalChangeCommitter
}

// NewLintRequest constructs one complete non-mutating lint request.
func NewLintRequest(
	provider GenerationProvider,
	policy ObservationPolicy,
	dispatcher EslintPluginDispatcher,
) PipelineRequest {
	return PipelineRequest{
		kind:       pipelineRequestLint,
		provider:   provider,
		policy:     policy,
		dispatcher: dispatcher,
	}
}

// NewProgressiveLintRequest constructs a non-mutating lint request whose
// complete baseline is synchronously presented before optional plugin
// enrichment is submitted. The constructor seals the release/order/syntax and
// failure semantics so the integration implements ports without coordinating
// pipeline stages.
func NewProgressiveLintRequest(
	provider GenerationProvider,
	demand ArtifactDemand,
	presenter ProgressiveDiagnostics,
) PipelineRequest {
	return PipelineRequest{
		kind:     pipelineRequestProgressiveLint,
		provider: provider,
		policy: ObservationPolicy{
			Demand:        demand,
			Plugin:        pluginProgressiveAfterNative,
			PluginFailure: PluginDiscardOnFailure,
		},
		presenter: presenter,
	}
}

// NewAutofixRequest constructs a bounded in-memory autofix request. RunPipeline
// alone advances the SourceSnapshot between observations; no external medium
// is mutated.
func NewAutofixRequest(
	provider GenerationProvider,
	policy ObservationPolicy,
	autofix AutofixPolicy,
	dispatcher EslintPluginDispatcher,
) PipelineRequest {
	return PipelineRequest{
		kind:       pipelineRequestAutofix,
		provider:   provider,
		policy:     policy,
		autofix:    autofix,
		dispatcher: dispatcher,
	}
}

// NewAutofixRequestWithCommitter adds one terminal projection of the final
// in-memory delta to an external medium. The committer is never invoked between
// observations or after an unsuccessful pipeline.
func NewAutofixRequestWithCommitter(
	provider GenerationProvider,
	committer FinalChangeCommitter,
	policy ObservationPolicy,
	autofix AutofixPolicy,
	dispatcher EslintPluginDispatcher,
) PipelineRequest {
	request := NewAutofixRequest(provider, policy, autofix, dispatcher)
	request.commit = &finalChangeCommit{committer: committer}
	return request
}

func (r PipelineRequest) validate() error {
	if r.provider == nil {
		return errors.New("linter pipeline: generation provider must not be nil")
	}
	if !r.policy.Demand.valid() {
		return errors.New("linter pipeline: artifact demand is invalid")
	}
	if !r.policy.Plugin.valid() {
		return errors.New("linter pipeline: plugin execution policy is invalid")
	}
	if !r.policy.PluginFailure.valid() {
		return errors.New("linter pipeline: plugin failure policy is invalid")
	}
	if r.policy.Plugin == pluginProgressiveAfterNative && r.kind != pipelineRequestProgressiveLint {
		return errors.New("linter pipeline: progressive execution requires a progressive request")
	}
	switch r.kind {
	case pipelineRequestLint:
		if r.commit != nil {
			return errors.New("linter pipeline: lint request must not carry a final change committer")
		}
		if r.presenter != nil {
			return errors.New("linter pipeline: lint request must not carry progressive presentation ports")
		}
	case pipelineRequestProgressiveLint:
		if r.commit != nil || r.dispatcher != nil {
			return errors.New("linter pipeline: progressive request must inject transport through its executor")
		}
		if r.presenter == nil {
			return errors.New("linter pipeline: progressive diagnostics must not be nil")
		}
		if r.policy.Plugin != pluginProgressiveAfterNative {
			return errors.New("linter pipeline: progressive request execution policy is invalid")
		}
	case pipelineRequestAutofix:
		if r.commit != nil && r.commit.committer == nil {
			return errors.New("linter pipeline: final change committer must not be nil")
		}
		if r.policy.Plugin == pluginProgressiveAfterNative {
			return errors.New("linter pipeline: autofix requires joined plugin work")
		}
		if !r.policy.Demand.plansAutofixes() {
			return errors.New("linter pipeline: autofix observations must request autofix artifacts")
		}
		if r.autofix.MaxRounds <= 0 {
			return errors.New("linter pipeline: autofix MaxRounds must be positive")
		}
		if r.autofix.MaxRounds > MaxFixRounds {
			return errors.New("linter pipeline: autofix MaxRounds exceeds the product safety bound")
		}
		if !r.autofix.VerificationDemand.valid() {
			return errors.New("linter pipeline: verification artifact demand is invalid")
		}
	default:
		return errors.New("linter pipeline: request kind is invalid")
	}
	return nil
}
