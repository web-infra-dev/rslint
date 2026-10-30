package linter

import (
	"context"
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// maxFixRounds is the product-wide safety bound. A round is one non-empty
// in-memory change-set application; a final verification observation and the
// optional external commit are not rounds. The bound stays private so
// integrations cannot independently select a pass count and drift apart.
const maxFixRounds = 10

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

// GenerationProvider plans one immutable lint generation from the
// pipeline-owned in-memory snapshot. The provider may combine it with a
// read-only disk base, an overlay VFS, or editor state, but it never owns or
// advances fix rounds and must not persist the snapshot.
//
// Planning does not construct Programs. The pipeline acquires and releases
// each planned unit; a provider only decides how to construct or borrow it.
type GenerationProvider interface {
	PlanGeneration(ctx context.Context, snapshot SourceSnapshot) (GenerationPlan, error)
}

// GenerationProviderFunc adapts a function to GenerationProvider.
type GenerationProviderFunc func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error)

func (f GenerationProviderFunc) AcquireGeneration(ctx context.Context, snapshot SourceSnapshot) (Generation, ReleaseFunc, error) {
	if f == nil {
		return Generation{}, nil, errors.New("linter pipeline: generation provider function must not be nil")
	}
	return f(ctx, snapshot)
}

func (f GenerationProviderFunc) PlanGeneration(_ context.Context, snapshot SourceSnapshot) (GenerationPlan, error) {
	return GenerationPlan{Loads: []GenerationLoad{{Acquire: func(ctx context.Context) (Generation, ReleaseFunc, error) {
		return f.AcquireGeneration(ctx, snapshot)
	}}}}, nil
}

// GenerationPlan describes ordered, disjoint bindings of the same immutable
// source snapshot. Each load constructs complete Programs or borrows an owner's
// coherent generation. Ownership and any compatibility binding are resolved by
// the provider, before publishing work; the executor never retries rules.
// A single load uses the same execution and result boundary as multiple loads.
// Loads transfers its private slice to the executor. Each acquisition runs at
// most once. Fix providers publish one coherent source-identity binding.
type GenerationPlan struct {
	Loads          []GenerationLoad
	SingleThreaded bool
}

// GenerationLoad acquires one complete binding under the core's shard budget.
// Counts are hints only; they never change roots, dependencies or checker pools.
// Release ends this observation's ownership or borrow, not an owner's cache.
type GenerationLoad struct {
	TargetCount  int
	CheckerCount int
	Acquire      func(context.Context) (Generation, ReleaseFunc, error)
}

// GenerationLoadError identifies acquisition failures for presentation;
// entrypoints may preserve their existing error wording.
type GenerationLoadError struct{ Err error }

func (e *GenerationLoadError) Error() string { return e.Err.Error() }
func (e *GenerationLoadError) Unwrap() error { return e.Err }

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
// Detached observations freeze ESTree before releasing the generation. Joined
// observations project inside their plugin goroutine. ConfigForFile and WirePath
// must be deterministic and side-effect free for the generation's lifetime.
type PluginGeneration struct {
	ConfigForFile EslintPluginFileConfigResolver
	WirePath      func(sourcePath string) string
}

// ArtifactDemand states which optional observation artifacts producers must
// materialize. Planning changes is a separate request kind, so a final
// verification may still collect edits without scheduling another commit.
type ArtifactDemand struct {
	Native rule.EditDemand
	Plugin rule.EditDemand
	// LintedFiles returns selected target paths without retaining compiler ASTs.
	LintedFiles bool
	// SourceFiles additionally retains the selected ASTs, including their paths,
	// for consumers such as the API's optional encoded-source response.
	SourceFiles bool
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

// ProgramExecution controls the native/type-check computation boundary.
type ProgramExecution uint8

const (
	// ProgramPhases joins all native tasks before starting program diagnostics.
	// This is the compatibility default for API, autofix, and editor requests.
	ProgramPhases ProgramExecution = iota
	// ProgramComplete runs each Program's diagnostics after its native workers
	// join. Native progress retains capacity for the pending diagnostics until
	// the task completes. Other Programs may still be linting; diagnostics are
	// buffered in input order and reported only after all native reports. This
	// is an explicit opt-in for non-mutating batch requests whose rules permit
	// that interleaving.
	ProgramComplete
)

// ObservationPolicy contains only producer and scheduling semantics shared by
// integrations; it does not name CLI, API, or LSP entrypoints.
type ObservationPolicy struct {
	Demand           ArtifactDemand
	Plugin           PluginExecution
	PluginFailure    PluginFailurePolicy
	ProgramExecution ProgramExecution
}

// DeferredPluginRun is a single-use, generation-detached enrichment task. It
// retains frozen wire/config/text only; the executor injects its own transport
// under its own timeout and cancellation lifecycle.
type DeferredPluginRun func(
	ctx context.Context,
	dispatcher EslintPluginDispatcher,
) (EslintPluginDispatchOutcome, error)

// ProgressiveDiagnostics binds baseline presentation and asynchronous
// enrichment admission to one request-scoped identity. RunPipeline hands off
// the baseline synchronously only after releasing the generation, then
// conditionally submits enrichment. Submit takes ownership synchronously and
// returns promptly without running the task inline. Presentation may be staged
// until inputs are validated; obsolete, unstarted enrichment can be discarded
// because it owns only detached data. Admitted enrichment runs exactly once.
// Timeout, cancellation, transport injection, admission, and presentation belong
// to the implementation.
type ProgressiveDiagnostics interface {
	PublishBaseline(ctx context.Context, diagnostics []rule.RuleDiagnostic)
	Submit(parentCtx context.Context, run DeferredPluginRun)
}

// AutofixPolicy configures the observable result and syntax behavior of the
// product-owned bounded apply-and-observe operation.
type AutofixPolicy struct {
	VerifyAfterLastRound     bool
	VerificationDemand       ArtifactDemand
	StopOnTargetSyntaxErrors bool

	// maxRounds is fixed by the constructors for production requests. Keeping
	// it private prevents CLI, API, and LSP from acquiring separate fix limits;
	// package tests may lower it to exercise boundary behavior economically.
	maxRounds int
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
	if autofix.maxRounds == 0 {
		autofix.maxRounds = maxFixRounds
	}
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
	if r.policy.ProgramExecution > ProgramComplete {
		return errors.New("linter pipeline: program execution policy is invalid")
	}
	if r.policy.ProgramExecution == ProgramComplete && r.kind != pipelineRequestLint {
		return errors.New("linter pipeline: complete Program execution requires a non-mutating batch request")
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
		if r.autofix.maxRounds <= 0 {
			return errors.New("linter pipeline: autofix round limit must be positive")
		}
		if r.autofix.maxRounds > maxFixRounds {
			return errors.New("linter pipeline: autofix round limit exceeds the product safety bound")
		}
		if !r.autofix.VerificationDemand.valid() {
			return errors.New("linter pipeline: verification artifact demand is invalid")
		}
	default:
		return errors.New("linter pipeline: request kind is invalid")
	}
	return nil
}
