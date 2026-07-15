package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// pluginConfigResolver resolves the eslint-plugin wire configKey plus the
// cached merged config for a file. lintResolver uses the target binding's
// owning config when available. The low-level API may supply a distinct opaque
// plugin routing identity through pluginConfigDirByOwner; CLI discovery already
// uses its Go-owned config directory as that identity and leaves the map nil.
type pluginConfigResolver struct {
	lintResolvers          []*lintConfigResolver
	pluginConfigDirByOwner map[string]string
}

func (r pluginConfigResolver) resolveForView(viewIndex int, filePath string) (wireKey string, merged *rslintconfig.MergedConfig) {
	if viewIndex < 0 || viewIndex >= len(r.lintResolvers) {
		return "", nil
	}
	lintResolver := r.lintResolvers[viewIndex]
	if lintResolver == nil {
		return "", nil
	}
	cfgDir, ok := lintResolver.ownerConfigDirForFile(filePath)
	if !ok {
		return "", nil
	}
	wireKey = cfgDir
	if pluginConfigDir, ok := r.pluginConfigDirByOwner[cfgDir]; ok {
		wireKey = pluginConfigDir
	}
	// Routing still comes from the owning config directory, but config payloads
	// must use discovery's sole exact-path selection. Re-running the generic
	// resolver here would lose live files/ignores predicates and could send the
	// plugin worker different languageOptions/settings than native lint used.
	return wireKey, lintResolver.ConfigForFile(filePath)
}

// buildPluginFileInputs collects, from RunLinter's lint targets, the files that
// have eslint-plugin rules and assembles their dispatch inputs. It reuses
// linter.CollectLintTargets so the dispatched file set matches the native pass
// exactly, and reuses each target's already-loaded *ast.SourceFile as the
// rebuild frame so Go never re-reads or re-decodes the file.
func buildPluginFileInputs(runOpts linter.RunLinterOptions, resolver pluginConfigResolver) []linter.EslintPluginFileInput {
	targets := linter.CollectLintTargets(runOpts)
	if len(targets) == 0 {
		return nil
	}
	var inputs []linter.EslintPluginFileInput
	for _, t := range targets {
		// Pure-native files never need a plugin routing key or merged plugin maps.
		// Skip that lookup before consulting the per-run config resolver.
		if !hasEslintPluginRule(t.Rules) {
			continue
		}
		filePath := t.Path
		wireKey, merged := resolver.resolveForView(t.ViewIndex, t.File.FileName())
		languageOptions, settings := rslintconfig.PluginMergedMaps(merged)
		// text=nil → the worker reads disk; sourceFile=t.File → Go rebuilds
		// against the frame ts-go already loaded (no re-read/decode). Shared
		// filter/assembly with the LSP path (F1).
		input, ok := linter.BuildEslintPluginFileInput(filePath, wireKey, t.Rules, languageOptions, settings, nil, t.File)
		if !ok {
			continue
		}
		inputs = append(inputs, input)
	}
	return inputs
}

// hasEslintPluginRule reports whether any configured rule is dispatched to the
// Node plugin host (rather than run natively in Go).
func hasEslintPluginRule(rules []linter.ConfiguredRule) bool {
	for _, r := range rules {
		if r.IsEslintPluginRule {
			return true
		}
	}
	return false
}

// dispatchPluginLintAsync runs the eslint-plugin dispatch in a goroutine
// (in parallel with the native RunLinter pass) and returns a channel that
// yields the rebuilt diagnostics once dispatch completes. The caller awaits
// it just before merging diagnostics for output / --fix.
func dispatchPluginLintAsync(
	ctx context.Context,
	dispatch linter.EslintPluginDispatcher,
	inputs []linter.EslintPluginFileInput,
	fix bool,
	suggestionsMode string,
) <-chan []rule.RuleDiagnostic {
	ch := make(chan []rule.RuleDiagnostic, 1)
	go func() {
		// onDiagnostic is invoked serially (DispatchEslintPluginRules fans batches
		// out to goroutines but emits diagnostics single-threaded after Wait), so
		// the local slice needs no lock.
		var diags []rule.RuleDiagnostic
		if dispatch != nil && len(inputs) > 0 {
			err := linter.DispatchEslintPluginRules(ctx, dispatch, inputs, fix, suggestionsMode,
				func(d rule.RuleDiagnostic) { diags = append(diags, d) })
			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Fprintf(os.Stderr, "rslint: eslint-plugin lint error: %v\n", err)
				// A dispatch error means one or more batches failed to run their
				// plugin rules (concurrent batches no longer abort each other, so
				// the rest may have succeeded and already emitted). Surface it as
				// an error diagnostic so the exit code reflects the failure instead
				// of a stderr-only false green (per-file worker failures already
				// surface inside DispatchEslintPluginRules; this covers a batch
				// dispatch error).
				diags = append(diags, linter.NewEslintPluginErrorDiagnostic(
					dispatchFailurePath(inputs), "rslint/plugin-lint-error",
					"ESLint plugin lint dispatch failed: "+err.Error()))
			}
		}
		ch <- diags
	}()
	return ch
}

// dispatchFailurePath anchors a dispatch-failure diagnostic to a real file so it
// renders with a location. One representative input suffices: the diagnostic
// exists to surface the failure and non-zero the exit code, not to attribute it
// to a specific file.
func dispatchFailurePath(inputs []linter.EslintPluginFileInput) string {
	if len(inputs) > 0 {
		return inputs[0].Path
	}
	return ""
}

// pluginSuggestionsMode picks the suggestion-collection mode for the
// worker. Suggestions are only materialized when fixing (the CLI applies
// them like fixes); otherwise the worker records descriptors without
// running them.
func pluginSuggestionsMode(fix bool) string {
	if fix {
		return linter.SuggestionsModeEager
	}
	return linter.SuggestionsModeOff
}
