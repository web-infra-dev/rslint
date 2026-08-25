package main

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// groupDiagsByFile groups a flat slice of diagnostics by their source file name.
func groupDiagsByFile(diags []rule.RuleDiagnostic) map[string][]rule.RuleDiagnostic {
	m := make(map[string][]rule.RuleDiagnostic)
	for _, d := range diags {
		f := d.FilePath
		m[f] = append(m[f], d)
	}
	return m
}

// remapDiagnosticTargetPaths keeps diagnostics in the caller's target path
// space when a TypeScript Program represents that target by another lexical or
// canonical source-file path. SourceFile remains unchanged because ranges and
// fixes are defined against its text; FilePath controls display and disk writes.
func remapDiagnosticTargetPaths(
	diags []rule.RuleDiagnostic,
	lintTargetBySourcePath map[string]target.File,
	filesystems ...vfs.FS,
) {
	if len(lintTargetBySourcePath) == 0 {
		return
	}
	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}
	for i := range diags {
		if lintTarget, ok := target.LookupSourceTarget(lintTargetBySourcePath, diags[i].FilePath, fsys); ok {
			diags[i].FilePath = lintTarget.Path
		}
	}
}

type typeScriptDiagnosticDedupeKey struct {
	path     string
	ruleName string
	pos      int
	end      int
	message  string
}

// deduplicateTypeScriptDiagnostics joins the lint-target syntax path and the
// program-wide type-check path. A file governed by a config without a tsconfig
// can be parsed into a source-only rslint Program while also belonging to another
// config's ts-go Program; --type-check legitimately visits both, but the same TypeScript
// diagnostic must be reported once.
func deduplicateTypeScriptDiagnostics(
	diags []rule.RuleDiagnostic,
	fsys vfs.FS,
	preferredCallerTargets ...map[string]string,
) []rule.RuleDiagnostic {
	if len(diags) == 0 {
		return diags
	}
	var callerTargetByCanonicalPath map[string]string
	if len(preferredCallerTargets) > 0 {
		callerTargetByCanonicalPath = preferredCallerTargets[0]
	}
	if len(diags) == 1 {
		if diags[0].Origin == rule.DiagnosticOriginTypeScript {
			canonicalID := canonicalFilesystemPathID(diags[0].FilePath, fsys)
			if callerTarget := callerTargetByCanonicalPath[canonicalID]; callerTarget != "" {
				diags[0].FilePath = callerTarget
			}
		}
		return diags
	}
	bestIndex := make(map[typeScriptDiagnosticDedupeKey]int)
	keys := make([]typeScriptDiagnosticDedupeKey, len(diags))
	for i, diagnostic := range diags {
		if diagnostic.Origin != rule.DiagnosticOriginTypeScript {
			continue
		}
		key := typeScriptDiagnosticDedupeKey{
			path:     canonicalFilesystemPathID(diagnostic.FilePath, fsys),
			ruleName: diagnostic.RuleName,
			pos:      diagnostic.Range.Pos(),
			end:      diagnostic.Range.End(),
			message:  diagnostic.Message.Description,
		}
		keys[i] = key
		current, exists := bestIndex[key]
		if !exists || preferTypeScriptDiagnostic(diagnostic, diags[current], callerTargetByCanonicalPath[key.path], fsys) {
			bestIndex[key] = i
		}
	}

	result := diags[:0]
	for i, diagnostic := range diags {
		if diagnostic.Origin != rule.DiagnosticOriginTypeScript {
			result = append(result, diagnostic)
			continue
		}
		key := keys[i]
		if bestIndex[key] != i {
			continue
		}
		if callerTarget := callerTargetByCanonicalPath[key.path]; callerTarget != "" {
			diagnostic.FilePath = callerTarget
		}
		result = append(result, diagnostic)
	}
	return result
}

func preferTypeScriptDiagnostic(candidate rule.RuleDiagnostic, current rule.RuleDiagnostic, callerTarget string, fsys vfs.FS) bool {
	if callerTarget != "" {
		candidateIsCaller := rslintconfig.ExactPathID(candidate.FilePath) == rslintconfig.ExactPathID(callerTarget)
		currentIsCaller := rslintconfig.ExactPathID(current.FilePath) == rslintconfig.ExactPathID(callerTarget)
		if candidateIsCaller != currentIsCaller {
			return candidateIsCaller
		}
	}
	candidatePath := tspath.NormalizePath(candidate.FilePath)
	currentPath := tspath.NormalizePath(current.FilePath)
	if candidatePath != currentPath {
		return candidatePath < currentPath
	}
	return false
}

// allowFileWarningKind categorizes why a CLI-specified file won't have lint
// rules applied. These are Phase-1 (lint) concepts; Phase 2 (type-check) is
// not affected by either case (it runs program-wide regardless of CLI scope
// and rslint ignores).
type allowFileWarningKind int

const (
	allowFileNotFound allowFileWarningKind = iota
	allowFileIgnored
)

// allowFileWarning is the structured form of a "this CLI-specified file
// won't be linted" diagnostic. Returned by collectAllowFileWarnings so the
// emission policy (skip in --type-check-only, format with relative paths,
// etc.) can be unit-tested without capturing stderr.
type allowFileWarning struct {
	Path string // absolute, normalized — matches what was put in allowFiles
	Kind allowFileWarningKind
}

// formatAllowFileWarning renders an allowFileWarning for stderr emission,
// resolving the absolute path against comparePathOptions. Returns "" for
// unknown kinds so callers can safely ignore the empty case.
func formatAllowFileWarning(w allowFileWarning, opts tspath.ComparePathsOptions) string {
	rel := tspath.ConvertToRelativePath(w.Path, opts)
	switch w.Kind {
	case allowFileNotFound:
		return fmt.Sprintf("warning: %s was not found, skipping", rel)
	case allowFileIgnored:
		return fmt.Sprintf("warning: %s is ignored because of a matching ignore pattern", rel)
	}
	return ""
}

// collectAllowFileWarnings maps target-plan explicit-file outcomes to CLI
// messages. It performs no filesystem or config lookup: target selection has
// already frozen the file identity, owner, ignore decision, and existence.
// Program membership is deliberately not consulted because a file outside
// every tsconfig can still be linted.
//
// This is a Phase-1 concern only. In --type-check-only mode the lint phase
// is skipped, so callers MUST gate emission on `!typeCheckOnly` — otherwise
// users see misleading messages like "is ignored because of a matching
// ignore pattern" while Phase 2 happily reports type errors for that same
// file. See website/docs/en/guide/type-checking.md ("type-check is not
// constrained by `files`/`ignores`").
func collectAllowFileWarnings(
	outcomes []target.ExplicitFileOutcome,
) []allowFileWarning {
	if len(outcomes) == 0 {
		return nil
	}

	var out []allowFileWarning
	for _, outcome := range outcomes {
		if outcome.Ignored {
			out = append(out, allowFileWarning{Path: outcome.Path, Kind: allowFileIgnored})
			continue
		}
		if !outcome.Exists {
			out = append(out, allowFileWarning{Path: outcome.Path, Kind: allowFileNotFound})
			continue
		}
	}
	return out
}

// shouldShortCircuitOutput returns true when rslint should bail early
// without printing diagnostics or a summary. The short-circuit exists so
// that e.g. `rslint nonexistent-file.ts` returns 0 with no spurious output
// when Phase 1 visited zero files.
//
// Any type-check mode (`--type-check` or `--type-check-only`) must NOT take
// the short-circuit: Phase 2 runs program-wide and is not gated by the CLI
// Scope/PerProgramFilter that drives lintedFileCount, so lintedFileCount==0
// is a normal state in which Phase 2 may still have produced diagnostics.
// Short-circuiting there would silently drop type errors that the user
// explicitly asked for — see website/docs/en/guide/type-checking.md.
func shouldShortCircuitOutput(typeCheckOnly, typeCheck, scopeRestricted bool, lintedFileCount int32) bool {
	if typeCheckOnly || typeCheck {
		return false
	}
	return scopeRestricted && lintedFileCount == 0
}
