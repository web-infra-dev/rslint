package linter

// compat_runner.go — the single dispatch entry point for eslint-plugin
// (compat) lint. Independent of any specific ingest path:
//
//   - RunLinter (native + compat mixed) collects CompatFileEntry from
//     its per-program file loop, then hands off to DispatchCompat.
//   - cmd/rslint's compat-only fast path skips ts-go Program creation
//     entirely, builds CompatFileEntry directly from a fs glob walk,
//     and calls DispatchCompat.
//
// Both ingest paths share the same bucketing / batching / diagnostic
// reconstruction code — there's no duplicate compat dispatch
// implementation living next to runLintRulesInProgram any more.
//
// The `SourceFile` field on CompatFileEntry is `ast.SourceFileLike`
// (an interface), not `*ast.SourceFile`. RunLinter passes the real
// ts-go SourceFile (already in memory from Program parse); the
// compat-only path leaves it nil and DispatchCompat substitutes a
// lightSourceFile (text + lazy ECMA line map). Downstream formatters
// only consume the interface (Text + ECMALineMap), so they're agnostic
// to which side built the entry.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// CompatFileEntry is one file's input to the compat dispatch pipeline.
// Callers fill Path/Text/Rules/Severity/LanguageOptions/Settings/ConfigKey
// and either populate SourceFile (when a real ts-go SourceFile is
// already available — e.g. from RunLinter's Program walk) or leave it
// nil (the compat-only fast path; DispatchCompat builds a
// lightSourceFile from Text).
type CompatFileEntry struct {
	Path string
	Text string
	// SourceFile, if non-nil, is used as RuleDiagnostic.SourceFile for
	// every diagnostic produced from this file. nil means DispatchCompat
	// will materialize a lightSourceFile from `Text` for the diagnostic
	// (the compat-only path doesn't build a ts-go SourceFile).
	SourceFile ast.SourceFileLike

	Rules           map[string]CompatRuleConfig
	Severity        map[string]rule.DiagnosticSeverity
	LanguageOptions map[string]any
	Settings        map[string]interface{}
	ConfigKey       string
}

// DispatchCompatOptions configures DispatchCompat.
type DispatchCompatOptions struct {
	Files           []CompatFileEntry
	Dispatcher      CompatBatchHandler
	OnDiagnostic    DiagnosticHandler
	CollectFixes    bool
	SuggestionsMode string
	Ctx             context.Context
	// IncludeFileText ships each entry's Text in the batch
	// (CompatLintFile.Text) so the worker lints that exact content instead
	// of re-reading disk — set for unsaved LSP/--api buffers (#3).
	IncludeFileText bool
}

// DispatchCompat runs the compat (eslint-plugin) lint path on the
// supplied files. Files are bucketed by rule-config signature, then
// each bucket is sent as a single homogeneous batch through Dispatcher.
//
// Returns a LintResult shaped like RunLinter's so the two paths can be
// blended by callers that mix native + compat (RunLinter itself uses
// this same function under the hood).
func DispatchCompat(opts DispatchCompatOptions) (*LintResult, error) {
	res := &LintResult{LintedFileCount: int32(len(opts.Files))}
	if len(opts.Files) == 0 || opts.Dispatcher == nil {
		return res, nil
	}
	sugMode := opts.SuggestionsMode
	if sugMode == "" {
		sugMode = "off"
	}
	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	executedRules := make(map[string]struct{})
	var failedCount int32

	buckets := groupCompatByRuleSig(opts.Files)
dispatchLoop:
	for _, bucket := range buckets {
		// Ctrl-C / LSP supersession cancels ctx. Stop dispatching the
		// remaining buckets rather than firing each one only to fail with
		// context.Canceled. Worker/child teardown is owned by the signal
		// path (engine.ts safeKillGo + pool.shutdown), not this loop.
		select {
		case <-ctx.Done():
			break dispatchLoop
		default:
		}
		for ruleName := range bucket[0].Rules {
			executedRules[ruleName] = struct{}{}
		}
		if dispatchCompatBucket(
			ctx,
			opts.Dispatcher,
			bucket,
			opts.OnDiagnostic,
			sugMode,
			opts.CollectFixes,
			opts.IncludeFileText,
		) {
			failedCount++
		}
	}

	res.ExecutedRules = executedRules
	res.CompatDispatchErrors = failedCount
	return res, nil
}

// groupCompatByRuleSig buckets entries whose rule-config signatures
// match — common case for a single config without overrides yields one
// bucket covering every file. Bucket iteration is sorted by signature
// so test snapshots remain deterministic across runs.
func groupCompatByRuleSig(entries []CompatFileEntry) [][]CompatFileEntry {
	if len(entries) == 0 {
		return nil
	}
	bySig := make(map[string][]CompatFileEntry, 1)
	for _, e := range entries {
		sig := compatRuleSignature(e.Rules, e.Severity)
		bySig[sig] = append(bySig[sig], e)
	}
	keys := make([]string, 0, len(bySig))
	for k := range bySig {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([][]CompatFileEntry, 0, len(bySig))
	for _, k := range keys {
		out = append(out, bySig[k])
	}
	return out
}

// compatRuleSignature returns a deterministic string fingerprint of a
// per-file rule configuration. Two files with identical (rule options,
// severity) maps produce equal signatures; any divergence on any rule
// produces different signatures.
func compatRuleSignature(
	rules map[string]CompatRuleConfig,
	severities map[string]rule.DiagnosticSeverity,
) string {
	names := make([]string, 0, len(rules))
	for k := range rules {
		names = append(names, k)
	}
	sort.Strings(names)
	var sb strings.Builder
	for _, n := range names {
		sb.WriteString(n)
		sb.WriteByte('|')
		// Options shape — JSON is canonical enough since the inputs are
		// produced by NormalizeRuleOptions (consistent shape per config).
		b, err := json.Marshal(rules[n].Options)
		if err != nil {
			// Options are JSON-safe by construction; a marshal error isn't
			// expected, but fall back to %v so the signature stays defined
			// rather than silently dropping the rule's options.
			b = fmt.Appendf(nil, "%v", rules[n].Options)
		}
		sb.Write(b)
		sb.WriteByte('|')
		// Severity (numeric).
		fmt.Fprintf(&sb, "%d", severities[n])
		sb.WriteByte(';')
	}
	return sb.String()
}

// CompatRuleMaps projects the IsEslintPluginRule entries of `rules` into the
// per-file data a CompatFileEntry carries: the rule→options and rule→severity
// maps, plus the file-wide LanguageOptions / Settings / ConfigKey (all compat
// rules in a file share the merged config, so the first compat rule's view is
// authoritative). ok is false when `rules` has no compat rule — callers skip
// the file. Shared by the CLI ingest (cmd.buildCompatFileInputs) and the
// in-Program path (runLintRulesInProgram) so the projection lives in one place.
func CompatRuleMaps(rules []ConfiguredRule) (
	ruleMap map[string]CompatRuleConfig,
	sevMap map[string]rule.DiagnosticSeverity,
	langOpts map[string]any,
	settings map[string]any,
	configKey string,
	ok bool,
) {
	ruleMap = make(map[string]CompatRuleConfig)
	sevMap = make(map[string]rule.DiagnosticSeverity)
	for _, r := range rules {
		if !r.IsEslintPluginRule {
			continue
		}
		ruleMap[r.Name] = CompatRuleConfig{Options: NormalizeRuleOptions(r.Options)}
		sevMap[r.Name] = r.Severity
		if !ok {
			langOpts = r.LanguageOptions
			settings = r.Settings
			configKey = r.ConfigKey
			ok = true
		}
	}
	return
}

// clampRange clamps a worker-reported [start, end] byte range into
// [0, textLen] AND enforces start <= end. Worker positions come from its
// own readFileSync view, which can exceed — or, for a malformed range,
// invert — the Go-side text; either would panic the formatter's
// text[start:end] slice downstream. Independent upper-bound clamping alone
// can still leave start > end, so re-order defensively.
func clampRange(start, end, textLen int) (int, int) {
	start = max(0, min(start, textLen))
	end = max(0, min(end, textLen))
	if end < start {
		end = start
	}
	return start, end
}

// ValidateAndNormalizeCompatResults enforces the compat dispatcher
// contract on a batch response. It is shared by the CLI runner
// (dispatchCompatBucket) and the LSP dispatcher (lintcompat_dispatcher)
// so the guard lives in exactly one place (#17):
//
//   - cardinality: exactly one result per input file;
//   - presence:    every result maps to some input file — byte-equal
//     first, then a tspath.NormalizePath fallback that rewrites the
//     result's FilePath IN PLACE back to the input string so downstream
//     path lookups (keyed by the input bytes) still hit;
//   - uniqueness:  no input file is returned twice (which would silently
//     drop another file's diagnostics).
//
// Ordering is intentionally NOT checked — both callers look up bucket data
// by path, not index. Returns nil on success; a descriptive (prefix-free)
// error otherwise, so each caller can add its own context.
func ValidateAndNormalizeCompatResults(results []CompatFileResult, files []CompatLintFile) error {
	if len(results) != len(files) {
		return fmt.Errorf("returned %d results for a %d-file batch (expected one result per input file)", len(results), len(files))
	}
	knownPaths := make(map[string]bool, len(files))
	normToOriginal := make(map[string]string, len(files))
	for _, f := range files {
		knownPaths[f.Path] = true
		normToOriginal[tspath.NormalizePath(f.Path)] = f.Path
	}
	seen := make(map[string]bool, len(results))
	for i, r := range results {
		matchedPath := r.FilePath
		if !knownPaths[matchedPath] {
			if orig, ok := normToOriginal[tspath.NormalizePath(r.FilePath)]; ok {
				results[i].FilePath = orig
				matchedPath = orig
			} else {
				return fmt.Errorf("returned unknown path %q (no input file matched even after normalize)", r.FilePath)
			}
		}
		if seen[matchedPath] {
			return fmt.Errorf("returned duplicate result for path %q", matchedPath)
		}
		seen[matchedPath] = true
	}
	return nil
}

// dispatchCompatBucket builds and sends a homogeneous batch for one
// bucket, streams the per-file diagnostics back through onDiagnostic,
// and returns true on dispatcher failure (caller aggregates).
func dispatchCompatBucket(
	ctx context.Context,
	dispatcher CompatBatchHandler,
	bucket []CompatFileEntry,
	onDiagnostic DiagnosticHandler,
	sugMode string,
	collectFixes bool,
	includeText bool,
) (failed bool) {
	if len(bucket) == 0 {
		return false
	}

	batchFiles := make([]CompatLintFile, 0, len(bucket))
	for _, e := range bucket {
		f := CompatLintFile{
			Path:            e.Path,
			LanguageOptions: e.LanguageOptions,
			Settings:        e.Settings,
			ConfigKey:       e.ConfigKey,
		}
		if includeText {
			// LSP/--api: carry the (possibly unsaved) overlay so the
			// worker lints what the editor shows, not the stale disk file.
			// Copy into a local before taking its address — `e` is the
			// loop variable and would otherwise alias across iterations.
			text := e.Text
			f.Text = &text
		}
		batchFiles = append(batchFiles, f)
	}

	// Every entry in the bucket has identical rule/severity maps (by
	// construction in groupCompatByRuleSig). Take the first.
	unionRules := bucket[0].Rules
	sevByRule := bucket[0].Severity

	batch := CompatBatch{
		Files:           batchFiles,
		Rules:           unionRules,
		CollectFixes:    collectFixes,
		SuggestionsMode: sugMode,
	}

	results, dispErr := dispatcher(ctx, batch)
	if dispErr != nil {
		// A cancelled ctx (user Ctrl-C, LSP debounce supersession) is a
		// deliberate stop, not a dispatcher failure: don't log it and don't
		// count it toward CompatDispatchErrors. Exit code is already forced
		// to 130 on signal, and worker/child cleanup runs via the signal path.
		if errors.Is(dispErr, context.Canceled) {
			return false
		}
		fmt.Fprintf(os.Stderr, "[rslint] compat dispatcher error: %v\n", dispErr)
		return true
	}

	// Validate the batch response (cardinality + path presence/uniqueness)
	// and normalize result paths in place — shared with the LSP dispatcher.
	// See ValidateAndNormalizeCompatResults (#17).
	if err := ValidateAndNormalizeCompatResults(results, batch.Files); err != nil {
		fmt.Fprintf(os.Stderr, "[rslint] compat dispatcher %v; marking compat-failed\n", err)
		return true
	}

	// Per-bucket lookups for diagnostic shaping.
	srcByPath := make(map[string]ast.SourceFileLike, len(bucket))
	textByPath := make(map[string]string, len(bucket))
	for _, e := range bucket {
		srcByPath[e.Path] = e.SourceFile
		textByPath[e.Path] = e.Text
	}

	for _, fileResult := range results {
		// Surface per-rule failures (plugin create() throw, missing rule
		// in loaded plugins, etc.) regardless of whether ParseError is
		// set. They're independent: a file can parse fine but still
		// have a buggy rule.
		for _, re := range fileResult.RuleErrors {
			fmt.Fprintf(os.Stderr, "[rslint] %s: rule %s failed: %s\n",
				fileResult.FilePath, re.Rule, re.Message)
		}
		if fileResult.ParseError != "" {
			fmt.Fprintf(os.Stderr, "[rslint] %s: compat layer parse error: %s\n",
				fileResult.FilePath, fileResult.ParseError)
			continue
		}
		// Cancelled files: the worker bails mid-walk so `Diagnostics`
		// is partial / potentially stale (LSP supersession of a newer
		// keystroke). Streaming would flicker stale-against-fresh.
		// Drop and continue.
		if fileResult.Cancelled {
			continue
		}

		// Resolve the SourceFile used for this diagnostic:
		//   1. Bucket-supplied SourceFile (real ts-go AST from RunLinter)
		//   2. lightSourceFile shim built from Text (compat-only path)
		//   3. Skip if neither is available (shouldn't happen given
		//      order/cardinality guards, but defensive).
		sourceFile := srcByPath[fileResult.FilePath]
		if sourceFile == nil {
			text, ok := textByPath[fileResult.FilePath]
			if !ok {
				continue
			}
			sourceFile = newLightSourceFile(text)
		}

		// Worker positions come from its own readFileSync view, which can
		// exceed the Go-side text length (unsaved LSP buffer, BOM/CRLF or
		// encoding skew). Clamp every worker-supplied position to
		// [0, len(text)] so a stale/skewed position can't panic the
		// formatter's slice in scanner.GetECMALineAndUTF16CharacterOfPosition.
		textLen := len(sourceFile.Text())
		for _, d := range fileResult.Diagnostics {
			// sevByRule is built from bucket[0].Severity — the exact rule set
			// sent to this worker batch. A missing key means the worker
			// reported a diagnostic for a rule not configured here (a
			// contract violation). Skip it: indexing the map directly would
			// return the zero value, and SeverityError is iota 0, so an
			// unconfigured rule would be silently surfaced as a build-breaking
			// error (and a configured 'warn' likewise promoted). Drop + warn
			// rather than guess a severity from the zero value.
			sev, ok := sevByRule[d.RuleName]
			if !ok {
				fmt.Fprintf(os.Stderr,
					"[rslint] %s: worker reported diagnostic for unconfigured rule %q; skipping\n",
					fileResult.FilePath, d.RuleName)
				continue
			}
			diag := rule.RuleDiagnostic{
				RuleName:   d.RuleName,
				Range:      core.NewTextRange(clampRange(d.StartPos, d.EndPos, textLen)),
				Message:    rule.RuleMessage{Id: d.MessageId, Description: d.Message},
				FilePath:   fileResult.FilePath,
				SourceFile: sourceFile,
				Severity:   sev,
			}
			if len(d.Fixes) > 0 {
				fixes := make([]rule.RuleFix, 0, len(d.Fixes))
				for _, f := range d.Fixes {
					fixes = append(fixes, rule.RuleFix{
						Text:  f.Text,
						Range: core.NewTextRange(clampRange(f.Range[0], f.Range[1], textLen)),
					})
				}
				diag.FixesPtr = &fixes
			}
			if len(d.Suggestions) > 0 {
				sugs := make([]rule.RuleSuggestion, 0, len(d.Suggestions))
				for _, s := range d.Suggestions {
					sugFixes := make([]rule.RuleFix, 0, len(s.Fixes))
					for _, f := range s.Fixes {
						sugFixes = append(sugFixes, rule.RuleFix{
							Text:  f.Text,
							Range: core.NewTextRange(clampRange(f.Range[0], f.Range[1], textLen)),
						})
					}
					sugs = append(sugs, rule.RuleSuggestion{
						Message:  rule.RuleMessage{Id: s.MessageId, Description: s.Desc},
						FixesArr: sugFixes,
					})
				}
				diag.Suggestions = &sugs
			}
			onDiagnostic(diag)
		}
	}
	return false
}

// lightSourceFile is the per-file SourceFileLike shim used by the
// compat-only path. Holds just the text and a lazily-built ECMA
// line-start map — no AST, no scope, no type info. The downstream
// formatter calls `scanner.GetECMALineAndUTF16CharacterOfPosition` and
// `scanner.GetECMALineStarts`, both of which take SourceFileLike and
// consume only Text() + ECMALineMap(), so the shim is sufficient.
type lightSourceFile struct {
	text    string
	lineMap core.ECMALineStarts
}

func newLightSourceFile(text string) *lightSourceFile {
	return &lightSourceFile{text: text}
}

// Text satisfies ast.SourceFileLike.
func (l *lightSourceFile) Text() string { return l.text }

// ECMALineMap satisfies ast.SourceFileLike. Lazy on first call —
// callers that never display line/column for a file (no diagnostic
// reported) skip the scan entirely.
func (l *lightSourceFile) ECMALineMap() []core.TextPos {
	if l.lineMap == nil {
		l.lineMap = core.ComputeECMALineStarts(l.text)
	}
	return l.lineMap
}
