package linter

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// LintedFile is one complete selected target projection. It includes
// syntax-error and zero-rule files so API callers do not have to rediscover
// selection through diagnostic side effects.
type LintedFile struct {
	Path       string
	SourceFile *ast.SourceFile
}

// NativeObservation is the complete native result for one generation and its
// requested ArtifactDemand.
type NativeObservation struct {
	Diagnostics []rule.RuleDiagnostic
	Lint        *LintResult
	// Files is populated only when ArtifactDemand.LintedFiles is requested.
	Files                 []LintedFile
	HasTargetSyntaxErrors bool
}

type pluginObservationKind uint8

const (
	pluginObservationNone pluginObservationKind = iota
	pluginObservationJoined
	pluginObservationProgressive
)

// ObservationResult distinguishes complete joined observations from a
// progressively presented baseline. Asynchronous enrichment never mutates the
// returned value.
type ObservationResult struct {
	Index  int
	Native NativeObservation

	pluginKind    pluginObservationKind
	pluginOutcome EslintPluginDispatchOutcome
}

// CompleteDiagnostics returns a fresh combined slice when plugin production is
// complete. The boolean is false for a progressively presented baseline.
func (r ObservationResult) CompleteDiagnostics() ([]rule.RuleDiagnostic, bool) {
	if r.pluginKind == pluginObservationProgressive {
		return nil, false
	}
	diagnostics := append([]rule.RuleDiagnostic(nil), r.Native.Diagnostics...)
	if r.pluginKind == pluginObservationJoined {
		diagnostics = append(diagnostics, r.pluginOutcome.Diagnostics...)
	}
	return diagnostics, true
}

// JoinedPluginOutcome returns the structured result for joined plugin work.
func (r ObservationResult) JoinedPluginOutcome() (EslintPluginDispatchOutcome, bool) {
	if r.pluginKind != pluginObservationJoined {
		return EslintPluginDispatchOutcome{}, false
	}
	return clonePluginOutcome(r.pluginOutcome), true
}

type fixResultKind uint8

const (
	fixResultNone fixResultKind = iota
	fixResultApplied
)

// FixRoundResult records one atomically applied, non-empty in-memory change
// set. External persistence is represented only by AppliedFixResult's terminal
// commit fields.
type FixRoundResult struct {
	ChangedPaths       []string
	AppliedDiagnostics int
	RestoredInitial    bool
}

// AppliedFixResult describes the pipeline-owned in-memory autofix state. Final
// changes are the net initial-to-current delta, not the concatenation of round
// changes. Committed is true only when a configured terminal committer
// confirmed the complete final delta.
type AppliedFixResult struct {
	Initial      ObservationResult
	Last         ObservationResult
	Rounds       []FixRoundResult
	FinalChanges []FileChange
	// FinalSources contains the complete final text for every path to which at
	// least one fix was applied in a successful round. Unlike FinalChanges, it
	// retains paths whose fixes restored the initial text so in-memory consumers
	// can still report that fixes were applied.
	FinalSources []SourceFileSnapshot
	// Verified means Last observes the final in-memory source represented by
	// FinalSources (and its net delta in FinalChanges). It does not promise that
	// another fix round would produce no edits.
	Verified       bool
	Committed      bool
	CommittedPaths []string
}

// fixResult is a private tagged state: exactly one of none or applied is
// meaningful for the request kind that produced it.
type fixResult struct {
	kind           fixResultKind
	initial        ObservationResult
	rounds         []FixRoundResult
	finalChanges   []FileChange
	finalSources   []SourceFileSnapshot
	verified       bool
	committed      bool
	committedPaths []string
}

func (r fixResult) applied(last ObservationResult) (AppliedFixResult, bool) {
	if r.kind != fixResultApplied {
		return AppliedFixResult{}, false
	}
	return AppliedFixResult{
		Initial:        r.initial,
		Last:           last,
		Rounds:         cloneFixRounds(r.rounds),
		FinalChanges:   cloneFileChanges(r.finalChanges),
		FinalSources:   cloneSourceFileSnapshots(r.finalSources),
		Verified:       r.verified,
		Committed:      r.committed,
		CommittedPaths: append([]string(nil), r.committedPaths...),
	}, true
}

// PluginDispatchRecord carries recoverable protocol notices and a joined
// plugin transport failure for integration presentation. Historical
// diagnostics stay on observations instead of retaining every superseded
// generation through autofix rounds.
type PluginDispatchRecord struct {
	Observation   int
	Notices       []EslintPluginProtocolNotice
	DispatchError error
}

// PipelineResult exposes the last observed generation separately from the
// tagged fix state. With autofix and no final verification, Observation may be
// older than the in-memory snapshot; AppliedFixes().Verified is then false.
type PipelineResult struct {
	Observation ObservationResult
	fix         fixResult

	executedRules  map[string]struct{}
	pluginOutcomes []PluginDispatchRecord
}

// AppliedFixes returns the bounded autofix history with PipelineResult's own
// last observation. Keeping this accessor on the aggregate prevents callers
// from pairing fix history with an unrelated observation.
func (r PipelineResult) AppliedFixes() (AppliedFixResult, bool) {
	return r.fix.applied(r.Observation)
}

// ExecutedRules returns the union of native rules executed across observations.
func (r PipelineResult) ExecutedRules() map[string]struct{} {
	result := make(map[string]struct{}, len(r.executedRules))
	for name := range r.executedRules {
		result[name] = struct{}{}
	}
	return result
}

// PluginOutcomes returns joined plugin transport outcomes in observation order.
func (r PipelineResult) PluginOutcomes() []PluginDispatchRecord {
	result := make([]PluginDispatchRecord, len(r.pluginOutcomes))
	for index, record := range r.pluginOutcomes {
		result[index] = record
		result[index].Notices = append([]EslintPluginProtocolNotice(nil), record.Notices...)
	}
	return result
}

func clonePluginOutcome(outcome EslintPluginDispatchOutcome) EslintPluginDispatchOutcome {
	outcome.Diagnostics = append([]rule.RuleDiagnostic(nil), outcome.Diagnostics...)
	outcome.Notices = append([]EslintPluginProtocolNotice(nil), outcome.Notices...)
	return outcome
}

func cloneFileChanges(changes []FileChange) []FileChange {
	return append([]FileChange(nil), changes...)
}

func cloneSourceFileSnapshots(files []SourceFileSnapshot) []SourceFileSnapshot {
	return append([]SourceFileSnapshot(nil), files...)
}

func cloneFixRounds(rounds []FixRoundResult) []FixRoundResult {
	result := make([]FixRoundResult, len(rounds))
	for index, round := range rounds {
		result[index] = round
		result[index].ChangedPaths = append([]string(nil), round.ChangedPaths...)
	}
	return result
}
