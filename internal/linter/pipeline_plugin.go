package linter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ErrDeferredPluginRunAlreadyInvoked reports an executor contract violation.
var ErrDeferredPluginRunAlreadyInvoked = errors.New("linter pipeline: deferred plugin run already invoked")

type pluginTask struct {
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
		if input.SourceFile != nil {
			return nil, errors.New("linter pipeline: deferred plugin run retained a source frame")
		}
		if input.Text == nil {
			return nil, errors.New("linter pipeline: deferred plugin run requires inline text")
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
	var diagnostics []rule.RuleDiagnostic
	notices, err := dispatchEslintPluginRules(
		ctx,
		dispatcher,
		t.inputs,
		t.collectFixes,
		t.suggestionsMode,
		t.timing,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	if err != nil {
		switch t.failure {
		case PluginDiscardOnFailure:
			diagnostics = nil
		case PluginKeepPartialWithSynthetic:
			if !errors.Is(err, context.Canceled) {
				diagnostics = append(diagnostics, NewEslintPluginErrorDiagnostic(
					eslintPluginDispatchFailurePath(t.inputs),
					"rslint/plugin-lint-error",
					"ESLint plugin lint dispatch failed: "+err.Error(),
				))
			}
		}
	}
	for index := range diagnostics {
		if targetPath, ok := t.targetPathByWire[diagnostics[index].FilePath]; ok {
			diagnostics[index].FilePath = targetPath
		}
	}
	return EslintPluginDispatchOutcome{
		Diagnostics:   diagnostics,
		Notices:       notices,
		DispatchError: err,
	}
}

func materializePluginTask(
	plan *LintPlan,
	generation Generation,
	snapshot SourceSnapshot,
	policy ObservationPolicy,
	detached bool,
) (pluginTask, error) {
	if generation.Plugin == nil {
		return pluginTask{failure: policy.PluginFailure}, nil
	}
	inputs := BuildEslintPluginFileInputs(plan, generation.Plugin.ConfigForFile)
	inlineSources := detached || !generation.Plugin.HostReadsInitialText || !snapshot.Empty()
	targetPathByWire := make(map[string]string, len(inputs))
	fixCandidates := make([]fixSourceCandidate, 0, len(inputs))
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
		fixCandidates = append(fixCandidates, fixSourceCandidate{
			path:   targetPath,
			source: input.SourceFile,
		})
		input.Path = wirePath
	}
	for index := range inputs {
		input := &inputs[index]
		if inlineSources {
			text, err := readGenerationText(generation, snapshot, targetPaths[index], input.SourceFile)
			if err != nil {
				return pluginTask{}, fmt.Errorf("linter pipeline: freeze plugin source %q: %w", input.Path, err)
			}
			input.Text = &text
		}
		if detached {
			if input.Text == nil {
				return pluginTask{}, fmt.Errorf("linter pipeline: detached plugin source %q must be inline", input.Path)
			}
			frozen, err := detachPluginInput(*input)
			if err != nil {
				return pluginTask{}, fmt.Errorf("linter pipeline: freeze plugin input %q: %w", input.Path, err)
			}
			inputs[index] = frozen
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
		collectFixes:     policy.Demand.Plugin&rule.EditDemandAutofix != 0,
		suggestionsMode:  suggestionsMode,
		timing:           timing,
		failure:          policy.PluginFailure,
	}, nil
}

func detachPluginInput(input EslintPluginFileInput) (EslintPluginFileInput, error) {
	if input.Text == nil {
		return EslintPluginFileInput{}, errors.New("inline text is required")
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
