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
	targetPaths := make([]string, len(inputs))
	// Validate the complete path projection before reading or freezing any
	// source. A malformed later input must not leave observable read side
	// effects from an earlier one or mask the structural contract error.
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
		targetPaths[index] = targetPath
		if needsFixSources {
			fixCandidates = append(fixCandidates, fixSourceCandidate{
				path:   targetPath,
				source: input.SourceFile,
			})
		}
		input.Path = wirePath
	}
	var builder estree.Builder
	for index := range inputs {
		input := &inputs[index]
		text, err := readTargetText(generation.Target.ReadText, snapshot, targetPaths[index], input.SourceFile)
		if err != nil {
			return pluginTask{}, fmt.Errorf("linter pipeline: freeze plugin source %q: %w", input.Path, err)
		}
		input.Text = &text
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
