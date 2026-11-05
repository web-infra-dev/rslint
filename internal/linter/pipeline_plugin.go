package linter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ErrDeferredPluginRunAlreadyInvoked reports an executor contract violation.
var ErrDeferredPluginRunAlreadyInvoked = errors.New("linter pipeline: deferred plugin run already invoked")

type pluginTask struct {
	resources        *pluginDispatchResources
	inputs           []EslintPluginFileInput
	targetPathByWire map[string]string
	fixCandidates    []fixSourceCandidate
	collectFixes     bool
	suggestionsMode  string
	timing           *TimingCollector
	failure          PluginFailurePolicy
}

type fixSourceCandidate struct {
	path   string
	source ast.SourceFileLike
}

func newDeferredPluginRun(task pluginTask) (DeferredPluginRun, error) {
	if len(task.fixCandidates) != 0 {
		return nil, errors.New("linter pipeline: deferred plugin run retained fix source candidates")
	}
	for _, input := range task.inputs {
		if input.SourceFile != nil || input.DiagnosticSource != nil {
			return nil, errors.New("linter pipeline: deferred plugin run retained a source frame")
		}
		if input.Frame == nil && input.FrameError == "" {
			return nil, errors.New("linter pipeline: deferred plugin run requires frozen ESTree")
		}
	}
	var state struct {
		sync.Mutex
		ran bool
	}
	return func(
		ctx context.Context,
		dispatcher EslintPluginDispatcher,
	) (EslintPluginDispatchOutcome, error) {
		if ctx == nil {
			return EslintPluginDispatchOutcome{}, errors.New("linter pipeline: plugin context must not be nil")
		}
		if dispatcher == nil {
			return EslintPluginDispatchOutcome{}, errors.New("linter pipeline: plugin dispatcher must not be nil")
		}
		state.Lock()
		if state.ran {
			state.Unlock()
			return EslintPluginDispatchOutcome{}, ErrDeferredPluginRunAlreadyInvoked
		}
		state.ran = true
		state.Unlock()
		return task.run(ctx, dispatcher), nil
	}, nil
}

func (t pluginTask) run(ctx context.Context, dispatcher EslintPluginDispatcher) EslintPluginDispatchOutcome {
	return applyPluginFailure(t.runRaw(ctx, dispatcher), t.failure)
}

// runRaw consumes one binding's inputs. Failure policy belongs to the complete
// observation, so a failed unit cannot discard only its own partial results or
// manufacture one synthetic error per Program.
func (t pluginTask) runRaw(ctx context.Context, dispatcher EslintPluginDispatcher) EslintPluginDispatchOutcome {
	failurePath := eslintPluginDispatchFailurePath(t.inputs)
	if path, ok := t.targetPathByWire[failurePath]; ok {
		failurePath = path
	}
	var diagnostics []rule.RuleDiagnostic
	var batches []pluginBatchRange
	notices, err := t.resources.dispatch(ctx, dispatcher, t.inputs, t.collectFixes, t.suggestionsMode, t.timing,
		func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		func(batch pluginBatchRange) { batches = append(batches, batch) })
	for index := range diagnostics {
		if targetPath, ok := t.targetPathByWire[diagnostics[index].FilePath]; ok {
			diagnostics[index].FilePath = targetPath
		}
	}
	return EslintPluginDispatchOutcome{Diagnostics: diagnostics, Notices: notices, DispatchError: err, failurePath: failurePath, batches: batches}
}

func applyPluginFailure(outcome EslintPluginDispatchOutcome, failure PluginFailurePolicy) EslintPluginDispatchOutcome {
	outcome.batches = nil
	if outcome.DispatchError == nil {
		return outcome
	}
	if failure == PluginDiscardOnFailure {
		outcome.Diagnostics = nil
	} else if !errors.Is(outcome.DispatchError, context.Canceled) {
		outcome.Diagnostics = append(outcome.Diagnostics, NewEslintPluginErrorDiagnostic(outcome.failurePath,
			"rslint/plugin-lint-error", "ESLint plugin lint dispatch failed: "+outcome.DispatchError.Error()))
	}
	return outcome
}

func materializePluginTask(
	ctx context.Context,
	plan *LintPlan,
	generation Generation,
	snapshot SourceSnapshot,
	policy ObservationPolicy,
	detached bool,
	planChanges bool,
	sourceReads pluginSourceReadSlots,
) (pluginTask, error) {
	if generation.Plugin == nil {
		return pluginTask{failure: policy.PluginFailure}, nil
	}
	inputs := BuildEslintPluginFileInputs(plan, generation.Plugin.ConfigForFile)
	targetPathByWire := make(map[string]string, len(inputs))
	var fixCandidates []fixSourceCandidate
	collectFixes := policy.Demand.Plugin&rule.EditDemandAutofix != 0
	needsFixSources := detached && planChanges && collectFixes
	if needsFixSources {
		fixCandidates = make([]fixSourceCandidate, 0, len(inputs))
	}
	missing := make([]int, 0, len(inputs))
	// Validate every path before calling the generation reader. Snapshot hits
	// already contain complete text and never enter the read queue.
	for index := range inputs {
		input := &inputs[index]
		sourcePath := input.Path
		targetPath := projectTargetPath(generation.Target.Path, sourcePath)
		wirePath := sourcePath
		if generation.Plugin.WirePath != nil {
			wirePath = generation.Plugin.WirePath(sourcePath)
		}
		if wirePath == "" {
			return pluginTask{}, fmt.Errorf("linter pipeline: plugin wire path for %q must not be empty", sourcePath)
		}
		if _, duplicate := targetPathByWire[wirePath]; duplicate {
			return pluginTask{}, fmt.Errorf("linter pipeline: duplicate plugin wire path %q", wirePath)
		}
		targetPathByWire[wirePath] = targetPath
		if text, ok := snapshot.Text(targetPath); ok {
			input.Text = new(text)
		} else {
			missing = append(missing, index)
		}
		if needsFixSources {
			fixCandidates = append(fixCandidates, fixSourceCandidate{
				path:   targetPath,
				source: input.SourceFile,
			})
		}
		input.Path = wirePath
	}
	readText := generation.Target.ReadText
	read := func(i int) error {
		input := &inputs[i]
		text, err := readGenerationText(readText, targetPathByWire[input.Path], input.SourceFile)
		if err != nil {
			return fmt.Errorf("linter pipeline: freeze plugin source %q: %w", input.Path, err)
		}
		input.Text = &text
		return nil
	}
	ready, readErr := sourceReads.prefetch(ctx, missing, len(inputs), generation.Native.SingleThreaded, read)
	var builder estree.Builder
	for index := range inputs[:ready] {
		input := &inputs[index]
		if input.Text == nil {
			if err := sourceReads.withSlot(ctx, func() error { return read(index) }); err != nil {
				return pluginTask{}, err
			}
		}
		if detached {
			frame, projectErr := projectPluginFrame(ctx, &builder, *input, nil)
			if projectErr != nil {
				if ctx.Err() != nil {
					return pluginTask{}, ctx.Err()
				}
				input.FrameError = projectErr.Error()
			} else {
				input.Frame = frame
			}
			frozen, err := detachPluginInput(*input)
			if err != nil {
				return pluginTask{}, fmt.Errorf("linter pipeline: freeze plugin input %q: %w", input.Path, err)
			}
			inputs[index] = frozen
		} else {
			// Joined wire tasks also need only rule identity, options and
			// severity. Never retain native Run closures or rule environments.
			rules := make([]rule.ConfiguredRule, len(input.Rules))
			for ruleIndex, configured := range input.Rules {
				rules[ruleIndex] = rule.ConfiguredRule{
					Name:               configured.Name,
					Severity:           configured.Severity,
					IsEslintPluginRule: true,
					Options:            configured.Options,
				}
			}
			input.Rules = rules
		}
	}
	if readErr != nil {
		return pluginTask{}, readErr
	}
	suggestionsMode := SuggestionsModeOff
	if policy.Demand.Plugin&rule.EditDemandSuggestion != 0 {
		suggestionsMode = SuggestionsModeEager
	}
	timing := generation.Native.Timing
	if detached {
		// Detached work must not retain mutable generation observers. LSP does not
		// collect plugin timing, and joined CLI/API work remains on the leased path.
		timing = nil
	}
	return pluginTask{
		inputs:           inputs,
		targetPathByWire: targetPathByWire,
		fixCandidates:    fixCandidates,
		collectFixes:     collectFixes,
		suggestionsMode:  suggestionsMode,
		timing:           timing,
		failure:          policy.PluginFailure,
	}, nil
}

// All bindings in an observation share these slots. Only source reads hold
// a slot; frame projection and JavaScript dispatch have their own lifetimes.
type pluginSourceReadSlots chan struct{}

func (slots pluginSourceReadSlots) withSlot(ctx context.Context, run func() error) error {
	select {
	case slots <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-slots }()
	if err := ctx.Err(); err != nil {
		return err
	}
	return run()
}

// Claim missing inputs in order, after acquiring a slot. When a reader fails,
// every earlier input is already claimed. Join those reads so the earliest
// error wins, and leave later unclaimed inputs alone.
func (slots pluginSourceReadSlots) prefetch(ctx context.Context, missing []int, count int, singleThreaded bool, read func(int) error) (int, error) {
	workers := min(cap(slots), len(missing))
	if singleThreaded || workers < 2 {
		return count, nil
	}
	// Stop waiting for slots as soon as the last input is claimed or a read
	// fails. Reads already in flight still finish under the generation lease.
	admission, stop := context.WithCancel(ctx)
	defer stop()
	var mu sync.Mutex
	ready := count
	var firstErr error
	readNext := func() error {
		mu.Lock()
		if len(missing) == 0 {
			mu.Unlock()
			return nil
		}
		index := missing[0]
		missing = missing[1:]
		if len(missing) == 0 {
			stop()
		}
		mu.Unlock()
		if err := read(index); err != nil {
			mu.Lock()
			missing = nil
			if index < ready {
				ready, firstErr = index, err
			}
			mu.Unlock()
			return err
		}
		return nil
	}
	work := newLintWorkGroup(false)
	for range workers {
		work.Queue(func() {
			defer stop()
			for {
				if err := slots.withSlot(admission, readNext); err != nil {
					return
				}
			}
		})
	}
	work.RunAndWait()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return ready, firstErr
}

func detachPluginInput(input EslintPluginFileInput) (EslintPluginFileInput, error) {
	if input.Text == nil || (input.Frame == nil && input.FrameError == "") {
		return EslintPluginFileInput{}, errors.New("frozen ESTree and diagnostic text are required")
	}
	text := *input.Text
	languageOptions, err := clonePluginJSON(input.LanguageOptions)
	if err != nil {
		return EslintPluginFileInput{}, fmt.Errorf("clone language options: %w", err)
	}
	settings, err := clonePluginJSON(input.Settings)
	if err != nil {
		return EslintPluginFileInput{}, fmt.Errorf("clone settings: %w", err)
	}
	rules := make([]rule.ConfiguredRule, len(input.Rules))
	for index, configured := range input.Rules {
		options, cloneErr := clonePluginJSON(configured.Options)
		if cloneErr != nil {
			return EslintPluginFileInput{}, fmt.Errorf("clone rule %q options: %w", configured.Name, cloneErr)
		}
		// Only these fields participate in plugin request construction and result
		// rebuilding. Dropping Environment and Run prevents the continuation from
		// retaining the generation's native rule/config graph.
		rules[index] = rule.ConfiguredRule{
			Name:               configured.Name,
			Severity:           configured.Severity,
			IsEslintPluginRule: true,
			Options:            options,
		}
	}
	return EslintPluginFileInput{
		Path:            input.Path,
		Text:            &text,
		SourceFile:      nil,
		Frame:           input.Frame,
		FrameError:      input.FrameError,
		ConfigKey:       input.ConfigKey,
		LanguageOptions: languageOptions,
		Settings:        settings,
		Rules:           rules,
	}, nil
}

func clonePluginJSON[T any](value T) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Batch ranges retain ordering metadata, never a second copy of diagnostics or
// AST owners. Units may finish in any order; publication keeps the established
// config-group, file and error precedence across the entire observation.
type pluginBatchRange struct {
	key                            string
	diagnosticStart, diagnosticEnd int
	noticeStart, noticeEnd         int
	err                            error
}

func mergePluginOutcomes(parts []EslintPluginDispatchOutcome) EslintPluginDispatchOutcome {
	var order []string
	groups := make(map[string]EslintPluginDispatchOutcome)
	var result EslintPluginDispatchOutcome
	for _, part := range parts {
		if result.failurePath == "" {
			result.failurePath = part.failurePath
		}
		for _, batch := range part.batches {
			group, exists := groups[batch.key]
			if !exists {
				order = append(order, batch.key)
			}
			group.Diagnostics = append(group.Diagnostics, part.Diagnostics[batch.diagnosticStart:batch.diagnosticEnd]...)
			group.Notices = append(group.Notices, part.Notices[batch.noticeStart:batch.noticeEnd]...)
			group.DispatchError = preferPluginError(group.DispatchError, batch.err)
			groups[batch.key] = group
		}
	}
	for _, key := range order {
		group := groups[key]
		result.Diagnostics = append(result.Diagnostics, group.Diagnostics...)
		result.Notices = append(result.Notices, group.Notices...)
		result.DispatchError = preferPluginError(result.DispatchError, group.DispatchError)
	}
	return result
}

func preferPluginError(first, next error) error {
	if next != nil && (first == nil || (errors.Is(first, context.Canceled) && !errors.Is(next, context.Canceled))) {
		return next
	}
	return first
}
