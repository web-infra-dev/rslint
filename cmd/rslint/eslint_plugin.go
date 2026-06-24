package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// pluginConfigResolver bundles everything needed to resolve, per file, the
// eslint-plugin wire configKey + merged config. configMap/originalConfigDir are
// the multi-config maps (normalized dir → entries, and normalized dir → the raw
// dir the JS host sent); rslintConfig/currentDirectory are the single-config
// fallback (which never mounts plugins). Bundling them keeps the normalized-vs-
// raw two-map duality private to resolve().
type pluginConfigResolver struct {
	configMap         map[string]rslintconfig.RslintConfig
	originalConfigDir map[string]string
	rslintConfig      rslintconfig.RslintConfig
	currentDirectory  string
}

// resolve returns the worker wire configKey + merged config for filePath. Go
// matches the file against its NORMALIZED owning-config key (FindNearestConfig),
// then echoes the RAW configDirectory the JS host sent as the wire configKey —
// that is what the Node worker keys its plugin map on. POSIX / single-config
// fall back to the normalized key, where raw == normalized.
func (r pluginConfigResolver) resolve(filePath string) (wireKey string, merged *rslintconfig.MergedConfig) {
	if r.configMap != nil {
		cfgDir, cfg := rslintconfig.FindNearestConfig(filePath, r.configMap)
		if cfg == nil {
			return "", nil
		}
		wireKey = cfgDir
		if raw, ok := r.originalConfigDir[cfgDir]; ok {
			wireKey = raw
		}
		return wireKey, cfg.GetConfigForFile(filePath, cfgDir)
	}
	return r.currentDirectory, r.rslintConfig.GetConfigForFile(filePath, r.currentDirectory)
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
	// Phase 1 (serial): assemble each plugin file's dispatch input WITHOUT its
	// type snapshot yet, recording the program + file so phase 2 can build
	// snapshots in parallel.
	type pending struct {
		input   linter.EslintPluginFileInput
		program *compiler.Program
		file    *ast.SourceFile
	}
	var pendings []pending
	for _, t := range targets {
		// Short-circuit pure-native files before resolving their config (resolve
		// re-runs GetConfigForFile), avoiding that cost across the whole repo.
		if !hasEslintPluginRule(t.Rules) {
			continue
		}
		filePath := t.File.FileName()
		wireKey, merged := resolver.resolve(filePath)
		languageOptions, settings := rslintconfig.PluginMergedMaps(merged)
		// text=nil → the worker reads disk; sourceFile=t.File → Go rebuilds
		// against the frame ts-go already loaded (no re-read/decode). Shared
		// filter/assembly with the LSP path (F1).
		input, ok := linter.BuildEslintPluginFileInput(filePath, wireKey, t.Rules, languageOptions, settings, nil, t.File)
		if !ok {
			continue
		}
		pendings = append(pendings, pending{input: input, program: t.Program, file: t.File})
	}
	if len(pendings) == 0 {
		return nil
	}

	// Phase 2 (parallel): build each file's type snapshot, sharded by owning
	// checker exactly like the native pass (linter.go's runLintRulesInProgram,
	// F1). A type snapshot is built for every file that HAS a program (project-
	// gate: parserOptions.project makes the program available — NOT gated on any
	// requiresTypeChecking meta). A file with no program gets no snapshot, and a
	// rule calling getParserServices there throws in the worker, like real ESLint.
	// Snapshots are self-contained per file (type-ids consistent within whichever
	// checker built them), so different files may use different pool checkers.
	// Build is serial WITHIN one checker (per-checker mutex) but PARALLEL across
	// checkers, and runs BEFORE RunLinter so it never races the native pass on the
	// same checker (this func completes before RunLinter is called in cmd.go).
	ctx := context.Background()
	type checkerKey struct {
		program *compiler.Program
		checker *checker.Checker
	}
	groups := map[checkerKey][]int{}
	for i := range pendings {
		p := pendings[i].program
		if p == nil {
			continue // no program → no snapshot (getParserServices throws in the worker)
		}
		chk, release := p.GetTypeCheckerForFile(ctx, pendings[i].file)
		release()
		key := checkerKey{program: p, checker: chk}
		groups[key] = append(groups[key], i)
	}
	wg := core.NewWorkGroup(runOpts.SingleThreaded)
	for key, idxs := range groups {
		key, idxs := key, idxs
		wg.Queue(func() {
			chk, done := key.program.GetTypeCheckerForFileExclusive(ctx, pendings[idxs[0]].file)
			defer done()
			for _, i := range idxs {
				linter.AttachTypeSnapshot(&pendings[i].input, chk, pendings[i].file)
			}
		})
	}
	wg.RunAndWait()

	inputs := make([]linter.EslintPluginFileInput, len(pendings))
	for i := range pendings {
		inputs[i] = pendings[i].input
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
