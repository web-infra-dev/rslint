package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
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
	Text            *string        `json:"text,omitempty"`
	ConfigKey       string         `json:"configKey"`
	LanguageOptions map[string]any `json:"languageOptions,omitempty"`
	Settings        map[string]any `json:"settings,omitempty"`
}

type EslintPluginLintRequest struct {
	Files           []EslintPluginLintFile            `json:"files"`
	Rules           map[string]EslintPluginRuleConfig `json:"rules"`
	Fix             bool                              `json:"fix"`
	SuggestionsMode string                            `json:"suggestionsMode"`
}

// SuggestionsMode values for EslintPluginLintRequest.SuggestionsMode — the wire
// contract the Node worker interprets. "eager" materializes each suggestion's
// fix (the CLI/LSP --fix path applies them like fixes); "off" records only the
// suggestion descriptors without running their fixers.
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
}

type EslintPluginLintResult struct {
	Results []EslintPluginFileResult `json:"results"`
}

// EslintPluginDispatcher sends one batch reverse-request to the Node host
// and returns its result. The CLI implements it over the generic IPC
// channel; the LSP server over an `rslint/pluginLint` request.
type EslintPluginDispatcher func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error)

// EslintPluginFileInput is one file's plugin-lint input as the caller
// (CLI/LSP) assembles it before dispatch.
type EslintPluginFileInput struct {
	Path string
	// Text is the overlay source SENT TO THE WORKER on the wire. The LSP sets
	// it to the unsaved-buffer content; the CLI leaves it nil so the worker
	// reads disk itself (avoids the ~60MB structuredClone of shipping every
	// file's text across the worker_threads boundary).
	Text *string
	// SourceFile is the frame Go REBUILDS diagnostics against (Go-local; never
	// sent on the wire). The CLI sets it to the ts-go *ast.SourceFile the native
	// pass already loaded (decoded + BOM-stripped), so Go reuses that frame
	// instead of re-reading/re-decoding the file — and plugin diagnostics share
	// the exact frame as native ones. nil for the LSP, which rebuilds against
	// the overlay Text (the worker linted that same string).
	SourceFile      ast.SourceFileLike
	ConfigKey       string
	LanguageOptions map[string]any
	Settings        map[string]any
	// Rules are the plugin rules (IsEslintPluginRule) configured for this
	// file, carrying Name / Options / Severity.
	Rules []ConfiguredRule
}

// BuildEslintPluginFileInput assembles one file's plugin-lint input from its
// enabled rules + per-file languageOptions/settings. Shared by the CLI and LSP
// hosts (F1): both filter the IsEslintPluginRule subset and assemble the input;
// it returns ok=false when the file has no plugin rules (caller skips dispatch).
// The caller supplies the frame: sourceFile (CLI — the native ts-go SourceFile)
// or text (LSP — the overlay the worker lints); see EslintPluginFileInput.
func BuildEslintPluginFileInput(filePath, configKey string, rules []ConfiguredRule, languageOptions, settings map[string]any, text *string, sourceFile ast.SourceFileLike) (EslintPluginFileInput, bool) {
	var pluginRules []ConfiguredRule
	for _, r := range rules {
		if r.IsEslintPluginRule {
			pluginRules = append(pluginRules, r)
		}
	}
	if len(pluginRules) == 0 {
		return EslintPluginFileInput{}, false
	}
	return EslintPluginFileInput{
		Path:            filePath,
		Text:            text,
		SourceFile:      sourceFile,
		ConfigKey:       configKey,
		LanguageOptions: languageOptions,
		Settings:        settings,
		Rules:           pluginRules,
	}, true
}

// eslintPluginShutdownSentinel is the ONLY benign parseError the worker emits:
// a graceful pool teardown (pool.shutdown()) racing an in-flight task. Every
// other parseError means the file did NOT finish linting — the native parser
// (oxc) recovers from syntax errors and returns a best-effort AST instead of a
// parseError, so a parseError is never an ordinary syntax error but always an
// abnormal worker failure (fs read, size guard, panic, normalize, crash,
// configKey miss) that must surface.
const eslintPluginShutdownSentinel = "shutdown"

// DispatchEslintPluginRules groups files by their rule-config signature,
// sends each homogeneous batch to the Node host, and feeds the rebuilt
// diagnostics through onDiagnostic — the SAME sink as native rules, so they
// merge into the unified output / --fix pipeline. Intended to be called
// from a goroutine running in parallel with native linting; a panic is
// trapped and returned as an error so it can't crash the process.
func DispatchEslintPluginRules(
	ctx context.Context,
	dispatch EslintPluginDispatcher,
	files []EslintPluginFileInput,
	fix bool,
	suggestionsMode string,
	onDiagnostic DiagnosticHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("eslint-plugin dispatch panicked: %v", r)
		}
	}()

	if len(files) == 0 || dispatch == nil {
		return nil
	}

	for _, batch := range groupEslintPluginFiles(files) {
		req := buildEslintPluginRequest(batch, fix, suggestionsMode)
		res, dispatchErr := dispatch(ctx, req)
		if dispatchErr != nil {
			return dispatchErr
		}
		if res == nil {
			continue
		}
		if applyErr := applyEslintPluginResults(batch, res, onDiagnostic); applyErr != nil {
			return applyErr
		}
	}
	return nil
}

// groupEslintPluginFiles buckets files by (configKey + rule-config
// signature). A batch shares a single `rules` map on the wire, so every
// file in it must agree on the configured rules — including each rule's
// options AND severity (severity is reattached from the matching rule, so
// two files configuring the same rule at different severities must not
// share a batch).
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
		Name     string `json:"name"`
		Options  any    `json:"options"`
		Severity int    `json:"severity"`
	}
	sig := make([]sigRule, 0, len(f.Rules))
	for _, r := range f.Rules {
		sig = append(sig, sigRule{Name: r.Name, Options: r.Options, Severity: r.Severity.Int()})
	}
	sort.Slice(sig, func(i, j int) bool { return sig[i].Name < sig[j].Name })
	b, err := json.Marshal(struct {
		ConfigKey string    `json:"configKey"`
		Rules     []sigRule `json:"rules"`
	}{f.ConfigKey, sig})
	if err != nil {
		// A rule's Options is `any`; a value that can't be marshaled would break
		// the batch key. Fall back to a per-file key so the file lints in its own
		// batch rather than silently mis-grouping with others.
		return f.Path
	}
	return string(b)
}

func buildEslintPluginRequest(batch []EslintPluginFileInput, fix bool, suggestionsMode string) EslintPluginLintRequest {
	rules := map[string]EslintPluginRuleConfig{}
	for _, r := range batch[0].Rules {
		rules[r.Name] = EslintPluginRuleConfig{Options: optionsToArray(r.Options)}
	}
	wireFiles := make([]EslintPluginLintFile, 0, len(batch))
	for _, f := range batch {
		wireFiles = append(wireFiles, EslintPluginLintFile{
			Path:            f.Path,
			Text:            f.Text,
			ConfigKey:       f.ConfigKey,
			LanguageOptions: f.LanguageOptions,
			Settings:        f.Settings,
		})
	}
	return EslintPluginLintRequest{
		Files:           wireFiles,
		Rules:           rules,
		Fix:             fix,
		SuggestionsMode: suggestionsMode,
	}
}

// optionsToArray normalizes rslint's RuleConfig.Options (nil / single value
// / array) into ESLint's options array (context.options).
func optionsToArray(opts any) []any {
	switch v := opts.(type) {
	case nil:
		return []any{}
	case []any:
		return v
	default:
		return []any{opts}
	}
}

func applyEslintPluginResults(batch []EslintPluginFileInput, res *EslintPluginLintResult, onDiagnostic DiagnosticHandler) error {
	byPath := map[string]*EslintPluginFileResult{}
	for i := range res.Results {
		byPath[res.Results[i].FilePath] = &res.Results[i]
	}
	for _, f := range batch {
		fr, ok := byPath[f.Path]
		if !ok {
			fmt.Fprintf(os.Stderr, "rslint: plugin-lint returned no result for %q\n", f.Path)
			continue
		}
		if fr.Cancelled {
			return context.Canceled
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
		// The worker computes every offset against the BOM-STRIPPED source (it
		// slices a leading BOM before parsing). When the rebuild frame KEEPS a
		// leading BOM — the LSP overlay, whose frame must match the BOM-inclusive
		// editor document + native diagnostics + the with-BOM content source.fix-
		// All splices into — shift every worker offset past the BOM into this
		// frame. The CLI frame (ts-go SourceFile) is already BOM-stripped, so the
		// shift is 0 there.
		bomShift := 0
		if strings.HasPrefix(text, utf8BOM) {
			bomShift = len(utf8BOM)
		}
		clamp := func(p int) int {
			p += bomShift
			if p < 0 {
				return 0
			}
			if p > len(text) {
				return len(text)
			}
			return p
		}

		// Surface every worker failure. oxc recovers from syntax errors and
		// returns a best-effort AST, so a parseError is never an ordinary
		// syntax error — it is always an abnormal failure (fs read, size guard,
		// panic, normalize, worker crash, configKey miss) that must affect the
		// exit code and be visible. The sole benign value is the "shutdown"
		// sentinel (graceful pool teardown racing a task).
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
				fmt.Fprintf(os.Stderr, "rslint: plugin diagnostic for unconfigured rule %q in %q\n", d.RuleName, f.Path)
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

// eslintPluginSourceFile returns the frame Go rebuilds plugin diagnostics
// against, whose byte offsets must match the worker's wire offsets.
//
//   - CLI: f.SourceFile is the ts-go *ast.SourceFile the native pass already
//     loaded — decoded and BOM-stripped by ts-go's vfs. For well-formed UTF-8
//     its byte frame is identical to the worker's readFileSync('utf8') frame, so
//     Go reuses it directly: no disk re-read, no re-decode, and plugin
//     diagnostics share the exact frame as native ones. The two frames diverge
//     only where ts-go and the worker decode bytes differently: UTF-16-encoded
//     files (ts-go transcodes UTF-16→UTF-8; the worker reads utf8) and
//     byte-malformed UTF-8 (ts-go keeps the raw bytes; the worker substitutes
//     U+FFFD). On those rare CLI-disk inputs a plugin offset after the
//     divergence can be byte-shifted (the clamp keeps it in-bounds — never a
//     panic). The LSP path is immune: it ships req.text and rebuilds against
//     that identical string.
//   - LSP: f.Text is the overlay string the worker linted, kept VERBATIM
//     (including any leading BOM). The LSP editor document and the native
//     diagnostics are BOM-inclusive (ts-go's scanner treats a leading BOM as
//     whitespace but keeps it in the text with inclusive positions), and
//     source.fixAll splices into the with-BOM overlay \u2014 so the plugin frame must
//     keep the BOM too. The worker reports BOM-stripped offsets (it slices the
//     BOM before parsing); applyEslintPluginResults shifts them back past the
//     BOM into this frame (see bomShift).
//
// ok=false only if the caller supplied neither (defensive).
func eslintPluginSourceFile(f EslintPluginFileInput) (ast.SourceFileLike, bool) {
	if f.SourceFile != nil {
		return f.SourceFile, true
	}
	if f.Text != nil {
		return newTextSourceFile(*f.Text), true
	}
	return nil, false
}

// utf8BOM is the UTF-8 encoding of U+FEFF (bytes EF BB BF).
const utf8BOM = "\ufeff"

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
