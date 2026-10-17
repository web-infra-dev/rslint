package linter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"golang.org/x/sync/errgroup"
)

// ─── wire types — mirror packages/rslint/src/eslint-plugin/plugin/plugin-lint-protocol.ts ───
//
// Offsets (startPos/endPos, fix range) are UTF-8 BYTE offsets: the Node
// worker converts from UTF-16 before shipping, and Go's scanner / TextRange
// consume bytes. The request `files[]` element keys its path as `path`
// (the Node builder maps it to the worker's filePath); the result keys it
// as `filePath`.

type EslintPluginRuleConfig struct {
	// Options is ESLint's post-severity options array (context.options).
	// Severity is NOT sent — Go reattaches it per-file after results return.
	Options []any `json:"options"`
}

type EslintPluginLintFile struct {
	Path            string         `json:"path"`
	Frame           []byte         `json:"frame,omitempty"`
	FrameAttachment *uint32        `json:"frameAttachment,omitempty"`
	ConfigKey       string         `json:"configKey"`
	LanguageOptions map[string]any `json:"languageOptions,omitempty"`
	Settings        map[string]any `json:"settings,omitempty"`
}

type EslintPluginLintRequest struct {
	Generation      string                            `json:"generation,omitempty"`
	Files           []EslintPluginLintFile            `json:"files"`
	Rules           map[string]EslintPluginRuleConfig `json:"rules"`
	CollectFixes    bool                              `json:"collectFixes"`
	SuggestionsMode string                            `json:"suggestionsMode"`
	CollectTiming   bool                              `json:"collectTiming,omitempty"`
}

// WithFrameAttachments moves the immutable binary frames into the existing IPC
// attachment channel. LSP serializes the same frames inline as base64 instead.
// It preserves the request and does not own storage or transport lifetimes.
func (req EslintPluginLintRequest) WithFrameAttachments() (EslintPluginLintRequest, [][]byte) {
	files := make([]EslintPluginLintFile, len(req.Files))
	frames := make([][]byte, 0, len(req.Files))
	for i, file := range req.Files {
		files[i] = file
		index := uint32(len(frames))
		frames = append(frames, file.Frame)
		files[i].Frame = nil
		files[i].FrameAttachment = &index
	}
	req.Files = files
	return req, frames
}

// SuggestionsMode values for EslintPluginLintRequest.SuggestionsMode — the wire
// contract the Node worker interprets. "eager" materializes each suggestion's
// fix for consumers such as Node API and LSP code actions; suggestions are never
// applied by the autofix pipeline. "off" records only their descriptors without
// running their fixers.
const (
	SuggestionsModeOff   = "off"
	SuggestionsModeEager = "eager"
)

type EslintPluginFix struct {
	Range [2]int `json:"range"`
	Text  string `json:"text"`
}

type EslintPluginSuggestion struct {
	MessageId string            `json:"messageId,omitempty"`
	Desc      string            `json:"desc,omitempty"`
	Fixes     []EslintPluginFix `json:"fixes"`
}

type EslintPluginDiagnostic struct {
	RuleName    string                   `json:"ruleName"`
	MessageId   string                   `json:"messageId,omitempty"`
	Message     string                   `json:"message"`
	StartPos    int                      `json:"startPos"`
	EndPos      int                      `json:"endPos"`
	Fixes       []EslintPluginFix        `json:"fixes,omitempty"`
	Suggestions []EslintPluginSuggestion `json:"suggestions,omitempty"`
}

type EslintPluginRuleError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type EslintPluginFileResult struct {
	FilePath    string                   `json:"filePath"`
	Diagnostics []EslintPluginDiagnostic `json:"diagnostics"`
	ParseError  string                   `json:"parseError,omitempty"`
	Cancelled   bool                     `json:"cancelled"`
	RuleErrors  []EslintPluginRuleError  `json:"ruleErrors,omitempty"`
	RuleTimes   map[string]float64       `json:"ruleTimes,omitempty"`
}

type EslintPluginLintResult struct {
	Results []EslintPluginFileResult `json:"results"`
}

// EslintPluginDispatcher opens an output/dispatch scope for each request to the Node host
// and returns its result. It must honor ctx and return after cancellation.
// Concurrent pipeline work joins dispatch before ending its producer lease;
// detached inputs need no generation borrow. The CLI uses the generic IPC
// channel; the LSP server over an `rslint/pluginLint` request.
// Dispatch borrows owned frames until return, or consumes allocator-owned frames
// during publication. Callers do not inspect frames after entering Dispatch.
// A dispatcher retaining owned bytes after return must copy them. Already frozen
// input frames keep their independent ownership.
type EslintPluginDispatcher interface {
	OpenBatch(ctx context.Context) (*EslintPluginBatch, error)
}

// EslintPluginBatch owns the output borrow for one projection window. A nil
// Allocate uses the scheduler's owned output cache (for inline/LSP transport).
// Discard rolls back a failed encoding's last allocation. Release runs exactly
// once on every path; a provider may end producer access earlier at publication.
// Storage providers know bytes and requests, never native AST or ESTree schema.
type EslintPluginBatch struct {
	Allocate func(int) []byte
	Discard  func([]byte)
	Dispatch EslintPluginDispatchFunc
	Release  func()
}

// EslintPluginDispatchFunc is the owned-frame adapter used by inline/LSP hosts.
type EslintPluginDispatchFunc func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error)

func (f EslintPluginDispatchFunc) OpenBatch(ctx context.Context) (*EslintPluginBatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errors.New("plugin dispatcher is nil")
	}
	return &EslintPluginBatch{Dispatch: f}, nil
}

var errPluginWindowFull = errors.New("plugin ESTree window is full")

// EslintPluginFileInput is the linter's generation-local representation of one
// plugin-lint input before wire batching.
type EslintPluginFileInput struct {
	Path string
	// Text is the complete generation snapshot, including its original BOM.
	// It stays Go-local for projection validation and detached diagnostics.
	Text *string
	// SourceFile is the exact Program syntax root. Projection consumes this
	// borrow before dispatch; detached work freezes Frame and drops it earlier.
	SourceFile *ast.SourceFile
	// DiagnosticSource carries the existing native/plugin diagnostic identity.
	// Read-only observations use the source's independent text projection;
	// fix planning keeps the original frame until text validation completes.
	DiagnosticSource ast.SourceFileLike
	// Frame owns detached syntax bytes; FrameError freezes a projection failure.
	Frame           []byte
	FrameError      string
	ConfigKey       string
	LanguageOptions map[string]any
	Settings        map[string]any
	// Rules are the plugin rules (IsEslintPluginRule) configured for this
	// file, carrying Name / Options / Severity.
	Rules []rule.ConfiguredRule
}

// buildEslintPluginFileInput adds file-specific defaults and routing to an
// already prepared plugin projection. It never selects or copies rules.
func buildEslintPluginFileInput(filePath, configKey string, rules *preparedRuleSet, languageOptions, settings map[string]any, text *string, sourceFile *ast.SourceFile) EslintPluginFileInput {
	_, _, normalizedLanguageOptions := rule.ResolveLanguageDefaults(filePath, rules.pluginLanguageOptions)
	effectiveLanguageOptions := make(map[string]any, len(languageOptions)+2)
	for name, value := range languageOptions {
		effectiveLanguageOptions[name] = value
	}
	effectiveLanguageOptions["ecmaVersion"] = normalizedLanguageOptions.EffectiveECMAVersion()
	effectiveLanguageOptions["sourceType"] = normalizedLanguageOptions.EffectiveSourceType()
	return EslintPluginFileInput{
		Path:            filePath,
		Text:            text,
		SourceFile:      sourceFile,
		ConfigKey:       configKey,
		LanguageOptions: effectiveLanguageOptions,
		Settings:        settings,
		Rules:           rules.plugins,
	}
}

// EslintPluginFileConfig is the config projection required to dispatch one
// prepared lint target to a third-party plugin host. Integrations resolve it
// from their own config ownership model, keeping the linter independent from
// config package types and transport-specific routing identities.
type EslintPluginFileConfig struct {
	ConfigKey       string
	LanguageOptions map[string]any
	Settings        map[string]any
}

// EslintPluginFileConfigResolver resolves the plugin-facing config projection
// for one Program source path.
type EslintPluginFileConfigResolver func(filePath string) EslintPluginFileConfig

// BuildEslintPluginFileInputs projects third-party plugin inputs from the same
// prepared file/rule plan consumed by native linting. Pure-native targets do
// not invoke resolveConfig.
func BuildEslintPluginFileInputs(
	plan *LintPlan,
	resolveConfig EslintPluginFileConfigResolver,
) []EslintPluginFileInput {
	if plan == nil {
		return nil
	}
	var inputs []EslintPluginFileInput
	for _, programPlan := range plan.programs {
		for _, filePlan := range programPlan.files {
			if filePlan.rules == nil || len(filePlan.rules.plugins) == 0 {
				continue
			}
			filePath := filePlan.file.FileName()
			var fileConfig EslintPluginFileConfig
			if resolveConfig != nil {
				fileConfig = resolveConfig(filePath)
			}
			inputs = append(inputs, buildEslintPluginFileInput(
				filePath,
				fileConfig.ConfigKey,
				filePlan.rules,
				fileConfig.LanguageOptions,
				fileConfig.Settings,
				nil,
				filePlan.file,
			))
		}
	}
	return inputs
}

// eslintPluginShutdownSentinel is the sole benign worker failure during a
// graceful pool teardown. Input/decoder failures are never syntax recovery and
// must remain visible through the existing plugin failure policy.
const eslintPluginShutdownSentinel = "shutdown"

const (
	pluginWindowFiles = 32
	pluginWindowBytes = 64 << 20
)

// DispatchEslintPluginRules groups files by their rule-config signature and
// dispatches each homogeneous batch to the Node host as an independent reverse
// IPC request, feeding the rebuilt diagnostics through onDiagnostic — the SAME
// sink as native rules, so they merge into the unified output / --fix pipeline.
// Intended to be called from a goroutine running in parallel with native
// linting.
//
// Batches are dispatched CONCURRENTLY (bounded by dispatchConcurrency). A serial
// loop leaves the Node worker pool idle between batches whenever the file set
// fragments into many small batches (monorepo multi-config / per-rule option
// splits) — it only saturates the pool within a single batch. Dispatching
// batches concurrently lets every batch's files share the one pool, keeping it
// saturated across batches. `dispatch` is therefore invoked from multiple
// goroutines and MUST be safe for concurrent use (the CLI IPC channel and the
// LSP reverse-request sender both are; LSP also only ever has one batch).
// Per-batch diagnostics are collected into index-keyed slots and emitted in
// batch order AFTER all batches finish, so onDiagnostic stays single-threaded
// (callers keep their lock-free collectors) and output ordering is
// deterministic. A batch failure no longer aborts the others; the first error
// in batch order is returned, preserving the previous serial error semantics
// for callers.
func DispatchEslintPluginRules(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	files []EslintPluginFileInput,
	collectFixes bool,
	suggestionsMode string,
	timing *TimingCollector,
	onDiagnostic DiagnosticHandler,
) error {
	_, err := dispatchEslintPluginRules(
		ctx,
		dispatch,
		slices.Clone(files),
		collectFixes,
		suggestionsMode,
		timing,
		onDiagnostic,
	)
	return err
}

func dispatchEslintPluginRules(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	files []EslintPluginFileInput,
	collectFixes bool,
	suggestionsMode string,
	timing *TimingCollector,
	onDiagnostic DiagnosticHandler,
) ([]EslintPluginProtocolNotice, error) {
	// Internal pipeline tasks transfer ownership of these private input slots.
	// Exported dispatch helpers clone the caller's slice before entering here.
	defer clear(files)
	if len(files) == 0 || dispatch == nil {
		return nil, nil
	}

	// Split only the lightweight input slices before entering the existing
	// concurrency limit. A large homogeneous group must share that capacity,
	// rather than serially alternating projection and worker execution.
	var batches [][]EslintPluginFileInput
	for _, group := range groupEslintPluginFiles(files) {
		for start := 0; start < len(group); start += pluginWindowFiles {
			batches = append(batches, group[start:min(start+pluginWindowFiles, len(group))])
		}
	}
	clear(files)
	batchDiags := make([][]rule.RuleDiagnostic, len(batches))
	batchNotices := make([][]EslintPluginProtocolNotice, len(batches))
	batchErrs := make([]error, len(batches))

	var g errgroup.Group
	limit := dispatchConcurrency()
	// This invocation owns a finite cache of projection and output buffers.
	// A window keeps exclusive ownership until its dispatcher has returned.
	builders := make(chan *pluginProjectionBuffers, limit)
	g.SetLimit(limit)
	for i, batch := range batches {
		g.Go(func() error {
			// Each batch writes only its own slot, so concurrent batches share
			// no state; diagnostics are emitted serially after Wait. The timing
			// collector is the one mutex-guarded exception, merged per file.
			batchErrs[i] = dispatchOneBatch(
				ctx,
				dispatch,
				batch,
				builders,
				collectFixes,
				suggestionsMode,
				timing,
				func(d rule.RuleDiagnostic) { batchDiags[i] = append(batchDiags[i], d) },
				func(notice EslintPluginProtocolNotice) {
					batchNotices[i] = append(batchNotices[i], notice)
				},
			)
			return nil
		})
	}
	_ = g.Wait()

	for _, diags := range batchDiags {
		for _, d := range diags {
			onDiagnostic(d)
		}
	}
	var notices []EslintPluginProtocolNotice
	for _, batch := range batchNotices {
		notices = append(notices, batch...)
	}
	// A real (non-cancellation) error outranks a cooperative cancel: with batches
	// running concurrently, a later batch's genuine transport failure must not be
	// masked by an earlier batch's context.Canceled (a superseded file). Fall
	// back to a Canceled only when every failing batch was canceled.
	var canceledErr error
	for _, batchErr := range batchErrs {
		if batchErr == nil {
			continue
		}
		if !errors.Is(batchErr, context.Canceled) {
			return notices, batchErr
		}
		if canceledErr == nil {
			canceledErr = batchErr
		}
	}
	return notices, canceledErr
}

// EslintPluginProtocolNoticeKind identifies a recoverable worker-protocol
// anomaly. Integrations decide how to present these structured notices; the
// linter never writes process output.
type EslintPluginProtocolNoticeKind uint8

const (
	EslintPluginMissingFileResult EslintPluginProtocolNoticeKind = iota
	EslintPluginUnconfiguredDiagnostic
)

// EslintPluginProtocolNotice describes a recoverable plugin response anomaly.
// The diagnostic stream remains usable, so notices are deliberately separate
// from DispatchError and its failure policy.
type EslintPluginProtocolNotice struct {
	Kind     EslintPluginProtocolNoticeKind
	FilePath string
	RuleName string
}

// EslintPluginDispatchOutcome is the result of one plugin dispatch.
// DispatchError is retained for surface logging. A non-cancellation failure
// also contributes a synthetic error diagnostic so it cannot produce a false
// successful lint result.
type EslintPluginDispatchOutcome struct {
	Diagnostics   []rule.RuleDiagnostic
	Notices       []EslintPluginProtocolNotice
	DispatchError error
}

// DispatchEslintPluginRulesAsync is the lower-level asynchronous dispatch
// primitive. RunPipeline owns its scheduling relative to native lint in
// product entrypoints.
func DispatchEslintPluginRulesAsync(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	inputs []EslintPluginFileInput,
	collectFixes bool,
	suggestionsMode string,
	timing *TimingCollector,
) <-chan EslintPluginDispatchOutcome {
	ch := make(chan EslintPluginDispatchOutcome, 1)
	go func() {
		outcome := DispatchEslintPluginRulesWithOutcome(
			ctx,
			dispatch,
			inputs,
			collectFixes,
			suggestionsMode,
			timing,
		)
		ch <- outcome
	}()
	return ch
}

// DispatchEslintPluginRulesWithOutcome dispatches plugin rules and collects
// their diagnostics into a structured result. RunPipeline applies product
// scheduling and failure policy around this lower-level operation.
func DispatchEslintPluginRulesWithOutcome(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	inputs []EslintPluginFileInput,
	collectFixes bool,
	suggestionsMode string,
	timing *TimingCollector,
) EslintPluginDispatchOutcome {
	var diagnostics []rule.RuleDiagnostic
	notices, err := dispatchEslintPluginRules(
		ctx,
		dispatch,
		slices.Clone(inputs),
		collectFixes,
		suggestionsMode,
		timing,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		diagnostics = append(diagnostics, NewEslintPluginErrorDiagnostic(
			eslintPluginDispatchFailurePath(inputs),
			"rslint/plugin-lint-error",
			"ESLint plugin lint dispatch failed: "+err.Error(),
		))
	}
	return EslintPluginDispatchOutcome{
		Diagnostics:   diagnostics,
		Notices:       notices,
		DispatchError: err,
	}
}

func eslintPluginDispatchFailurePath(inputs []EslintPluginFileInput) string {
	if len(inputs) > 0 {
		return inputs[0].Path
	}
	return ""
}

// The scheduler owns output reuse; the language encoder owns only scratch.
// Both caches are bounded separately and contain no Program references.
type pluginProjectionBuffers struct {
	syntax         estree.Builder
	output         [][]byte
	outputCapacity int
}

func (b *pluginProjectionBuffers) reset() {
	b.syntax.Reset()
	for index := range b.output {
		b.output[index] = b.output[index][:0]
	}
}

// Pack frames in stable blocks with a bounded total capacity. Growing storage
// never copies completed frames or invalidates their views. Overflow is owned.
func (b *pluginProjectionBuffers) allocate(size int) []byte {
	const limit = 4 << 20
	previous := 0
	for index, block := range b.output {
		start := len(block)
		if size <= cap(block)-start {
			b.output[index] = block[:start+size]
			return b.output[index][start : start+size : start+size]
		}
		previous = max(previous, cap(block))
	}
	if size > limit-b.outputCapacity {
		return make([]byte, size)
	}
	capacity := min(limit-b.outputCapacity, max(size, 2*previous, 64<<10))
	block := make([]byte, size, capacity)
	b.output = append(b.output, block)
	b.outputCapacity += capacity
	return block[:size:size]
}

// dispatchOneBatch sends one batch's reverse request and rebuilds its
// diagnostics through onDiagnostic. A panic is trapped and returned as an error
// so a worker-side failure can't crash the driver process (each batch runs in
// its own goroutine, so a panic must be trapped here rather than escaping).
func dispatchOneBatch(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	batch []EslintPluginFileInput,
	builders chan *pluginProjectionBuffers,
	collectFixes bool,
	suggestionsMode string,
	timing *TimingCollector,
	onDiagnostic DiagnosticHandler,
	onNotice func(EslintPluginProtocolNotice),
) (err error) {
	defer clear(batch)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("eslint-plugin dispatch panicked: %v", r)
		}
	}()
	// Output buffers stay exclusive until dispatch returns, including errors
	// and cancellation. No publication/reader lifetime enters the encoder.
	var builder *pluginProjectionBuffers
	releaseBuilder := func() {
		if builder != nil {
			builder.reset()
			select {
			case builders <- builder:
			default:
			}
			builder = nil
		}
	}
	defer releaseBuilder()
	var output *EslintPluginBatch
	releaseOutput := func() {
		current := output
		output = nil
		if current != nil && current.Release != nil {
			current.Release()
		}
	}
	defer releaseOutput()
	for offset := 0; offset < len(batch); {
		if err := ctx.Err(); err != nil {
			return err
		}
		var outputErr error
		output, outputErr = dispatch.OpenBatch(ctx)
		if outputErr != nil {
			return outputErr
		}
		if output == nil || output.Dispatch == nil {
			return errors.New("plugin dispatcher returned an invalid batch")
		}
		window := make([]EslintPluginFileInput, 0, min(pluginWindowFiles, len(batch)-offset))
		bytes := 0
		for offset < len(batch) && len(window) < pluginWindowFiles {
			input := batch[offset]
			if input.Frame != nil && len(window) > 0 && bytes+len(input.Frame) > pluginWindowBytes {
				break
			}
			if input.Frame != nil && output.Allocate != nil {
				data := output.Allocate(len(input.Frame))
				copy(data, input.Frame)
				input.Frame = data
			}
			if input.DiagnosticSource == nil && input.SourceFile != nil {
				input.DiagnosticSource = input.SourceFile
			}
			if input.Frame == nil && input.FrameError == "" {
				if builder == nil {
					select {
					case builder = <-builders:
					default:
						builder = new(pluginProjectionBuffers)
					}
				}
				var allocated []byte
				frame, projectErr := projectPluginFrame(ctx, &builder.syntax, input, func(size int) []byte {
					if len(window) > 0 && bytes+size > pluginWindowBytes {
						panic(errPluginWindowFull)
					}
					if output.Allocate != nil {
						allocated = output.Allocate(size)
						return allocated
					}
					return builder.allocate(size)
				})
				if projectErr != nil {
					if allocated != nil && output.Discard != nil {
						output.Discard(allocated)
					}
					// The frame was sized before allocation. Leave this input untouched
					// and rebuild it in the next window; no transport lease crosses windows.
					if errors.Is(projectErr, errPluginWindowFull) {
						break
					}
					if ctx.Err() != nil {
						return ctx.Err()
					}
					input.FrameError = projectErr.Error()
				} else {
					input.Frame = frame
				}
			}
			// Projection has consumed the syntax borrow. Metadata keeps only its
			// diagnostic frame; no original input slot pins completed ASTs.
			input.SourceFile = nil
			batch[offset] = EslintPluginFileInput{}
			if input.FrameError != "" {
				window = append(window, input)
				offset++
				continue
			}
			window = append(window, input)
			bytes += len(input.Frame)
			offset++
		}
		if len(window) == 0 {
			releaseOutput()
			continue
		}
		ready := make([]EslintPluginFileInput, 0, len(window))
		failed := make([]EslintPluginFileInput, 0)
		failures := &EslintPluginLintResult{}
		for _, input := range window {
			if input.FrameError != "" {
				failed = append(failed, input)
				failures.Results = append(failures.Results, EslintPluginFileResult{FilePath: input.Path, ParseError: input.FrameError})
			} else {
				ready = append(ready, input)
			}
		}
		if len(ready) == 0 {
			releaseOutput()
			if err := applyEslintPluginResults(window, failures, timing, onDiagnostic, onNotice); err != nil {
				return err
			}
			continue
		}
		req := buildEslintPluginRequest(ready, collectFixes, suggestionsMode, timing != nil)
		// Inline base64 plus metadata must also fit the generic 256 MiB message
		// ceiling. Include each file's base64 padding and the enclosing RPC envelope.
		metadata := req
		metadata.Files = append([]EslintPluginLintFile(nil), req.Files...)
		for index := range metadata.Files {
			metadata.Files[index].Frame = nil
		}
		encoded, encodeErr := json.Marshal(metadata)
		if encodeErr != nil {
			return encodeErr
		}
		inlineBytes := 0
		for _, file := range req.Files {
			inlineBytes += (len(file.Frame)+2)/3*4 + len(`,"frame":""`)
		}
		if len(encoded)+inlineBytes > 256<<20-4096 {
			return errors.New("plugin frame batch exceeds IPC message limit")
		}
		// Results use source/rule metadata. Drop their frame references and
		// oversized scratch buffers before dispatch. Owned output remains borrowed
		// until return; shared publication consumes the producer's byte access.
		for index := range window {
			window[index].Frame = nil
		}
		for index := range ready {
			ready[index].Frame = nil
		}
		if builder != nil {
			builder.reset()
		}
		res, dispatchErr := output.Dispatch(ctx, req)
		releaseOutput()
		releaseBuilder()
		if dispatchErr != nil || res == nil {
			if err := applyEslintPluginResults(failed, failures, timing, onDiagnostic, onNotice); err != nil {
				return err
			}
			if dispatchErr != nil {
				return dispatchErr
			}
			continue
		}
		failures.Results = append(failures.Results, res.Results...)
		if err := applyEslintPluginResults(window, failures, timing, onDiagnostic, onNotice); err != nil {
			return err
		}
	}
	return nil
}

// dispatchConcurrency bounds how many plugin-lint batches are dispatched
// concurrently. The Node worker pool has up to 8 workers (capped at the core
// count; see worker-pool.ts), so to keep it saturated the in-flight batch count
// must at least match that worker count; a small buffer on top lets a freed
// worker pick up the next batch's tasks from the shared queue without idling
// during the Go↔Node round trip between batches. Capping it keeps a heavily
// fragmented file set from dispatching every batch's reverse request at once —
// Each bounded file window is one batch. Byte-heavy windows may make smaller
// sequential requests while retaining this same concurrency slot.
func dispatchConcurrency() int {
	const poolWorkers = 8 // mirrors the Node worker pool cap (worker-pool.ts)
	const roundTripBuffer = 2
	n := runtime.NumCPU()
	if n > poolWorkers {
		n = poolWorkers
	}
	return n + roundTripBuffer
}

// groupEslintPluginFiles buckets files by (configKey + rule-config
// signature). A batch shares a single `rules` map on the wire, so every file
// in it must agree on the configured rules' names AND options. Severity is
// deliberately NOT part of the key: it never travels on the wire (only options
// do) and is reattached per-file in applyEslintPluginResults, so two files
// configuring the same rule at different severities produce a byte-identical
// request and SHOULD share a batch rather than needlessly fragment the pool.
func groupEslintPluginFiles(files []EslintPluginFileInput) [][]EslintPluginFileInput {
	order := []string{}
	groups := map[string][]EslintPluginFileInput{}
	for _, f := range files {
		key := eslintPluginBatchKey(f)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], f)
	}
	out := make([][]EslintPluginFileInput, 0, len(order))
	for _, k := range order {
		out = append(out, groups[k])
	}
	return out
}

func eslintPluginBatchKey(f EslintPluginFileInput) string {
	type sigRule struct {
		Name    string `json:"name"`
		Options []any  `json:"options"`
	}
	sig := make([]sigRule, 0, len(f.Rules))
	for _, r := range f.Rules {
		sig = append(sig, sigRule{Name: r.Name, Options: r.Options})
	}
	sort.Slice(sig, func(i, j int) bool { return sig[i].Name < sig[j].Name })
	b, err := json.Marshal(struct {
		ConfigKey string    `json:"configKey"`
		Rules     []sigRule `json:"rules"`
	}{f.ConfigKey, sig})
	if err != nil {
		// A non-serializable option value would break
		// the batch key. Fall back to a per-file key so the file lints in its own
		// batch rather than silently mis-grouping with others.
		return f.Path
	}
	return string(b)
}

func buildEslintPluginRequest(batch []EslintPluginFileInput, collectFixes bool, suggestionsMode string, collectTiming bool) EslintPluginLintRequest {
	rules := map[string]EslintPluginRuleConfig{}
	for _, r := range batch[0].Rules {
		rules[r.Name] = EslintPluginRuleConfig{Options: r.Options}
	}
	wireFiles := make([]EslintPluginLintFile, 0, len(batch))
	for _, f := range batch {
		wireFiles = append(wireFiles, EslintPluginLintFile{
			Path:            f.Path,
			Frame:           f.Frame[:len(f.Frame):len(f.Frame)],
			ConfigKey:       f.ConfigKey,
			LanguageOptions: f.LanguageOptions,
			Settings:        f.Settings,
		})
	}
	return EslintPluginLintRequest{
		Files:           wireFiles,
		Rules:           rules,
		CollectFixes:    collectFixes,
		SuggestionsMode: suggestionsMode,
		CollectTiming:   collectTiming,
	}
}

func applyEslintPluginResults(
	batch []EslintPluginFileInput,
	res *EslintPluginLintResult,
	timing *TimingCollector,
	onDiagnostic DiagnosticHandler,
	onNotice func(EslintPluginProtocolNotice),
) error {
	byPath := map[string]*EslintPluginFileResult{}
	for i := range res.Results {
		byPath[res.Results[i].FilePath] = &res.Results[i]
	}
	for _, f := range batch {
		fr, ok := byPath[f.Path]
		if !ok {
			if onNotice != nil {
				onNotice(EslintPluginProtocolNotice{
					Kind:     EslintPluginMissingFileResult,
					FilePath: f.Path,
				})
			}
			continue
		}
		if fr.Cancelled {
			return context.Canceled
		}
		if timing != nil && len(fr.RuleTimes) > 0 {
			timing.addFileRuleTimesMS(fr.FilePath, fr.RuleTimes)
		}

		sevByRule := map[string]rule.DiagnosticSeverity{}
		for _, cr := range f.Rules {
			sevByRule[cr.Name] = cr.Severity
		}
		tsf, ok := eslintPluginSourceFile(f)
		if !ok {
			// No frame to rebuild against (neither a native SourceFile nor an
			// overlay was provided) — shouldn't happen; surface rather than
			// collapse every diagnostic to offset 0.
			onDiagnostic(makeEslintPluginErrorDiagnostic(f.Path, newTextSourceFile(""),
				"rslint/plugin-lint-error", "ESLint plugin lint failed: no source frame for "+f.Path))
			continue
		}
		text := tsf.Text()
		// The worker computes every offset against the BOM-stripped source (it
		// slices a leading BOM before parsing), which is the one coordinate
		// space rslint uses as well — see eslintPluginSourceFile. So worker
		// offsets land in this frame as they stand.
		clamp := func(p int) int {
			if p < 0 {
				return 0
			}
			if p > len(text) {
				return len(text)
			}
			return p
		}

		// Program owns syntax errors. Projection/decoder and worker failures
		// use this existing error field and must remain visible. The sole benign
		// value is the "shutdown" sentinel (graceful pool teardown racing a task).
		if fr.ParseError != "" {
			if fr.ParseError != eslintPluginShutdownSentinel {
				onDiagnostic(makeEslintPluginErrorDiagnostic(f.Path, tsf,
					"rslint/plugin-lint-error", "ESLint plugin lint failed: "+fr.ParseError))
			}
			continue
		}
		// Rule throws are authoritative — the editor never renders stderr,
		// so surface each as a visible diagnostic.
		for _, re := range fr.RuleErrors {
			onDiagnostic(makeEslintPluginErrorDiagnostic(f.Path, tsf, re.Rule,
				"rule error: "+re.Message))
		}

		for _, d := range fr.Diagnostics {
			sev, known := sevByRule[d.RuleName]
			if !known {
				sev = rule.SeverityError
				if onNotice != nil {
					onNotice(EslintPluginProtocolNotice{
						Kind:     EslintPluginUnconfiguredDiagnostic,
						FilePath: f.Path,
						RuleName: d.RuleName,
					})
				}
			}
			start := clamp(d.StartPos)
			end := clamp(d.EndPos)
			if end < start {
				end = start
			}
			rd := rule.RuleDiagnostic{
				RuleName:    d.RuleName,
				Range:       core.NewTextRange(start, end),
				Message:     rule.RuleMessage{Id: d.MessageId, Description: d.Message},
				SourceFile:  tsf,
				FilePath:    f.Path,
				Severity:    sev,
				FixesPtr:    rebuildEslintPluginFixes(d.Fixes, clamp),
				Suggestions: rebuildEslintPluginSuggestions(d.Suggestions, clamp),
			}
			onDiagnostic(rd)
		}
	}
	return nil
}

// Diagnostics use the same BOM-stripped source that was frozen in the frame.
// Joined work can use its leased SourceFile; detached work owns only the text.
func eslintPluginSourceFile(f EslintPluginFileInput) (ast.SourceFileLike, bool) {
	if f.DiagnosticSource != nil {
		return f.DiagnosticSource, true
	}
	if f.SourceFile != nil {
		return f.SourceFile, true
	}
	if f.Text != nil {
		return newTextSourceFile(strings.TrimPrefix(*f.Text, utils.BOM)), true
	}
	return nil, false
}

func makeEslintPluginErrorDiagnostic(path string, sf ast.SourceFileLike, ruleName, msg string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		RuleName:   ruleName,
		Range:      core.NewTextRange(0, 0),
		Message:    rule.RuleMessage{Description: msg},
		SourceFile: sf,
		FilePath:   path,
		Severity:   rule.SeverityError,
	}
}

// NewEslintPluginErrorDiagnostic builds a synthetic error diagnostic for an
// ESLint-plugin lint failure that has no per-file source location — e.g. a
// total dispatch failure where the whole batch never ran. The CLI appends it
// to the diagnostic set so the failure affects the exit code instead of being
// a stderr-only false green; path anchors it to a file for rendering.
func NewEslintPluginErrorDiagnostic(path, ruleName, msg string) rule.RuleDiagnostic {
	return makeEslintPluginErrorDiagnostic(path, newTextSourceFile(""), ruleName, msg)
}

func rebuildEslintPluginFixes(fixes []EslintPluginFix, clamp func(int) int) *[]rule.RuleFix {
	if len(fixes) == 0 {
		return nil
	}
	out := make([]rule.RuleFix, 0, len(fixes))
	for _, fx := range fixes {
		start := clamp(fx.Range[0])
		end := clamp(fx.Range[1])
		if end < start {
			end = start
		}
		out = append(out, rule.RuleFix{Text: fx.Text, Range: core.NewTextRange(start, end)})
	}
	return &out
}

func rebuildEslintPluginSuggestions(suggestions []EslintPluginSuggestion, clamp func(int) int) *[]rule.RuleSuggestion {
	if len(suggestions) == 0 {
		return nil
	}
	out := make([]rule.RuleSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		fixes := []rule.RuleFix{}
		for _, fx := range s.Fixes {
			start := clamp(fx.Range[0])
			end := clamp(fx.Range[1])
			if end < start {
				end = start
			}
			fixes = append(fixes, rule.RuleFix{Text: fx.Text, Range: core.NewTextRange(start, end)})
		}
		out = append(out, rule.RuleSuggestion{
			Message:  rule.RuleMessage{Id: s.MessageId, Description: s.Desc},
			FixesArr: fixes,
		})
	}
	return &out
}

func projectPluginFrame(ctx context.Context, builder *estree.Builder, input EslintPluginFileInput, allocate func(int) []byte) ([]byte, error) {
	if input.SourceFile == nil {
		return nil, errors.New("plugin ESTree projection requires the Program SourceFile")
	}
	if input.Text == nil {
		return nil, errors.New("plugin ESTree projection requires the source snapshot")
	}
	text := *input.Text
	hasBOM := strings.HasPrefix(text, "\ufeff")
	if strings.TrimPrefix(text, "\ufeff") != input.SourceFile.Text() {
		return nil, errors.New("plugin source snapshot does not match Program syntax")
	}
	sourceType, _ := input.LanguageOptions["sourceType"].(string)
	if sourceType == "" {
		sourceType = "module"
	}
	return builder.BuildInto(ctx, input.SourceFile, estree.Options{SourceType: sourceType, HasBOM: hasBOM}, allocate)
}
