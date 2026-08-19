package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/output"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/term"
)

// lintArgs is the parsed CLI flag set, decoupled from the global flag
// package so each entry point can build it: parseLintFlags from argv, and
// runCLI additionally from the IPC init-handshake payload.
type lintArgs struct {
	Init           bool
	Config         string
	Fix            bool
	TypeCheck      bool
	TypeCheckOnly  bool
	TraceOut       string
	CpuprofOut     string
	SingleThreaded bool
	Format         string
	NoColor        bool
	ForceColor     bool
	// StdoutIsTTY is the TTY fact for the real output destination, reported
	// by the Node host via the IPC init payload (the Go process's own stdout
	// is an IPC pipe and says nothing about the user's terminal). False when
	// unavailable (for example in the wasm build).
	StdoutIsTTY bool
	Quiet       bool
	Timing      bool
	TimingLimit int
	MaxWarnings int
	StartTimeMs int64
	RuleFlags   []string
	// Positional args resolved into existing-dir vs file paths.
	AllowFiles []string
	AllowDirs  []string
	// FS is an optional run-scoped filesystem shared with Go config discovery.
	// Native CLI supplies one cached instance so the later target/program phases
	// reuse directory entries already read by the staged frontier.
	FS vfs.FS
	// DeferTimingTable, when non-nil, receives the rendered --timing table
	// instead of it being printed to stderr by the pipeline. The IPC entry
	// uses it to emit the table only after the async stdout forwarding has
	// drained, so the table cannot interleave with the lint report.
	DeferTimingTable func(table string)
	// ConfigCatalog is the immutable result of Go-owned JS/TS config discovery.
	// A nil or empty automatic catalog selects the native JSON/JSONC loader;
	// explicit catalogs remain authoritative even when their config is empty.
	ConfigCatalog *discovery.ConfigCatalog
}

// repeatedFlag collects multiple values for the same flag (e.g. --rule used multiple times).
type repeatedFlag []string

func (f *repeatedFlag) String() string     { return strings.Join(*f, ", ") }
func (f *repeatedFlag) Set(v string) error { *f = append(*f, v); return nil }

const usage = `🚀 Rslint - Rocket Speed Linter

Usage:
  rslint [OPTIONS] [files...]

Options:
  --init                Initialize a default config in the current directory.
  -c, --config PATH     Which rslint config file to use.
  --format FORMAT       Output format: default | jsonline | github | gitlab
  --fix                 Automatically fix problems
  --type-check          Enable TypeScript type checking
  --type-check-only     Run only TypeScript type checking (skip all lint rules)
  --no-color            Disable colored output
  --force-color         Force colored output
  --quiet               Report errors only
  --timing [all|N]      Print a per-rule timing table (all rules, or top N)
  --max-warnings Int    Number of warnings to trigger nonzero exit code
  --rule RULE           Rule override, e.g. 'no-console: error' (repeatable)
  -h, --help            Show help
`

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
func remapDiagnosticTargetPaths(diags []rule.RuleDiagnostic, targetPathBySourcePath map[string]string, filesystems ...vfs.FS) {
	if len(targetPathBySourcePath) == 0 {
		return
	}
	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}
	for i := range diags {
		targetPath := targetPathBySourcePath[diags[i].FilePath]
		if targetPath == "" {
			targetPath = targetPathBySourcePath[canonicalFilesystemPathID(diags[i].FilePath, fsys)]
		}
		if targetPath != "" {
			diags[i].FilePath = targetPath
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
		candidateIsCaller := exactFilesystemPathID(candidate.FilePath) == exactFilesystemPathID(callerTarget)
		currentIsCaller := exactFilesystemPathID(current.FilePath) == exactFilesystemPathID(callerTarget)
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

// applyFixPass applies auto-fixes for all files in diagnosticsByFile,
// writes fixed content to disk, and returns the number of issues fixed. Write
// failures are returned after all independent files have been attempted.
func applyFixPass(diagnosticsByFile map[string][]rule.RuleDiagnostic, fsys vfs.FS) (int, error) {
	fixed := 0
	fileNames := make([]string, 0, len(diagnosticsByFile))
	for fileName := range diagnosticsByFile {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	var writeErrors []error
	for _, fileName := range fileNames {
		fileDiagnostics := diagnosticsByFile[fileName]
		var diagnosticsWithFixes []rule.RuleDiagnostic
		for _, d := range fileDiagnostics {
			if len(d.Fixes()) > 0 {
				diagnosticsWithFixes = append(diagnosticsWithFixes, d)
			}
		}
		if len(diagnosticsWithFixes) == 0 {
			continue
		}

		// The parsed text has no byte order mark — decoding the file consumed
		// it — so put it back before fixing. ApplyRuleFixes carries it through
		// to the bytes written here, and drops it only for a fix that asks.
		originalContent := diagnosticsWithFixes[0].SourceFile.Text()
		if utils.SourceHasBOM(fsys, fileName) {
			originalContent = utils.BOM + originalContent
		}
		fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(originalContent, diagnosticsWithFixes)

		if wasFixed {
			err := os.WriteFile(fileName, []byte(fixedContent), 0644)
			if err != nil {
				writeErrors = append(writeErrors, fmt.Errorf("write fixed file %q: %w", fileName, err))
			} else {
				fixed += len(diagnosticsWithFixes) - len(unapplied)
			}
		}
	}
	return fixed, errors.Join(writeErrors...)
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

// collectAllowFileWarnings explains, for each CLI-specified file in
// allowFiles, why it won't be visited by Phase 1 (lint). Program membership
// is deliberately not consulted: lint targets are resolved before type-info
// binding, so a file outside every tsconfig can still be linted.
// Returns nil for empty allowFiles.
//
// This is a Phase-1 concern only. In --type-check-only mode the lint phase
// is skipped, so callers MUST gate emission on `!typeCheckOnly` — otherwise
// users see misleading messages like "is ignored because of a matching
// ignore pattern" while Phase 2 happily reports type errors for that same
// file. See website/docs/en/guide/type-checking.md ("type-check is not
// constrained by `files`/`ignores`").
func collectAllowFileWarnings(
	allowFiles []string,
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	useCaseSensitive bool,
	filesystems ...vfs.FS,
) []allowFileWarning {
	if len(allowFiles) == 0 {
		return nil
	}

	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}
	var configOwnerResolver *rslintconfig.ConfigOwnerResolver
	if configMap != nil {
		configOwnerResolver = rslintconfig.NewConfigOwnerResolver(configMap, fsys)
	}

	var out []allowFileWarning
	for _, f := range allowFiles {
		var cfgDir string
		var cfg rslintconfig.RslintConfig
		if configMap != nil {
			// Canonical fallback is file-specific: two symlinked files in the
			// same lexical directory can resolve to different config roots.
			cfgDir, cfg = configOwnerResolver.Resolve(f)
		} else {
			cfgDir = currentDirectory
			cfg = rslintConfig
		}
		matchFile := f
		matchDir := cfgDir
		if fsys != nil && cfgDir != "" {
			canonicalPath := authoritativeFilesystemPath(f, fsys)
			matchFile = (rslintconfig.DiscoveredLintTarget{
				Path:            tspath.NormalizePath(f),
				CanonicalPath:   canonicalPath,
				ConfigDirectory: cfgDir,
			}).MatchPath(fsys)
			matchDir = authoritativeFilesystemPath(cfgDir, fsys)
		}
		if rslintconfig.IsDefaultExcludedPath(matchFile, matchDir, useCaseSensitive) ||
			(cfg != nil && cfg.IsFileIgnored(matchFile, matchDir)) {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileIgnored})
			continue
		}

		if _, err := os.Stat(f); err != nil {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileNotFound})
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

// validateTypeCheckOnlyFlags rejects --type-check-only combined with flags
// whose semantics depend on the lint phase that this mode disables. Returns
// (0, "") when the combination is valid (or --type-check-only isn't set);
// otherwise returns (exitCode > 0, stderr message). Pulled out as a pure
// function so the policy can be exercised in unit tests.
func validateTypeCheckOnlyFlags(typeCheckOnly, fix bool, ruleFlags []string) (int, string) {
	if !typeCheckOnly {
		return 0, ""
	}
	if fix {
		return 2, "error: --fix cannot be combined with --type-check-only (no lint rules run, nothing to fix)"
	}
	if len(ruleFlags) > 0 {
		return 2, "error: --rule cannot be combined with --type-check-only (no lint rules run)"
	}
	return 0, ""
}

func isBroadProjectLoadScope(
	allowFiles []string,
	allowDirs []string,
	currentDirectory string,
	useCaseSensitive bool,
) bool {
	if len(allowDirs) == 0 {
		return len(allowFiles) == 0
	}
	options := tspath.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	for _, directory := range allowDirs {
		if tspath.ContainsPath(directory, currentDirectory, options) {
			return true
		}
	}
	return false
}

func cloneConfigMap(configMap map[string]rslintconfig.RslintConfig) map[string]rslintconfig.RslintConfig {
	if configMap == nil {
		return nil
	}
	cloned := make(map[string]rslintconfig.RslintConfig, len(configMap))
	for dir, cfg := range configMap {
		cloned[dir] = slices.Clone(cfg)
	}
	return cloned
}

// validateResolvedRuleOptions runs rule-options schema validation over the
// resolved configuration: every configMap config in multi-config mode (each
// message suffixed with the owning config directory, since the same rule can
// be misconfigured differently per config), or the single rslintConfig
// otherwise. It returns normalized copies plus one formatted message per
// failure, leaving the input configs untouched.
func validateResolvedRuleOptions(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
) (map[string]rslintconfig.RslintConfig, rslintconfig.RslintConfig, []string) {
	var messages []string
	if configMap == nil {
		normalized, optionsErrors := rslintconfig.ValidateRuleOptions(rslintConfig, rslintconfig.GlobalRuleRegistry)
		for _, optionsError := range optionsErrors {
			messages = append(messages, optionsError.Error())
		}
		return nil, normalized, messages
	}
	configDirs := make([]string, 0, len(configMap))
	for dir := range configMap {
		configDirs = append(configDirs, dir)
	}
	slices.Sort(configDirs)
	normalizedMap := make(map[string]rslintconfig.RslintConfig, len(configMap))
	for _, dir := range configDirs {
		normalized, optionsErrors := rslintconfig.ValidateRuleOptions(configMap[dir], rslintconfig.GlobalRuleRegistry)
		normalizedMap[dir] = normalized
		for _, optionsError := range optionsErrors {
			messages = append(messages, fmt.Sprintf("%s (config at %s)", optionsError.Error(), dir))
		}
	}
	return normalizedMap, rslintConfig, messages
}

// resolveStartTime returns the start time for timing output.
// If startTimeMs (epoch millis from the Node.js entry point) is positive,
// it is used so the reported duration covers end-to-end execution.
// Otherwise falls back to time.Now().
func resolveStartTime(startTimeMs int64) time.Time {
	if startTimeMs > 0 {
		return time.UnixMilli(startTimeMs)
	}
	return time.Now()
}

// loadGitignoreAndProjects overlaps two independent CLI preparation steps for
// the single-config JSON path. Program construction itself belongs to the
// loader; this small scheduling decision remains command orchestration.
func loadGitignoreAndProjects(
	config rslintconfig.RslintConfig,
	configDirectory string,
	targetFiles []string,
	singleThreaded bool,
	session *loader.Session,
) (rslintconfig.RslintConfig, loader.ProjectSet, error) {
	var (
		configWithIgnores rslintconfig.RslintConfig
		projects          loader.ProjectSet
		programErr        error
	)
	work := core.NewWorkGroup(singleThreaded)
	work.Queue(func() {
		configWithIgnores = rslintconfig.ConfigWithGitignore(
			config,
			configDirectory,
			session.FS(),
			targetFiles,
		)
	})
	work.Queue(func() {
		projects, programErr = session.BuildProject(configDirectory, config, singleThreaded)
	})
	work.RunAndWait()
	if programErr != nil {
		return config, loader.ProjectSet{}, programErr
	}
	return configWithIgnores, projects, nil
}

// parseLintFlags parses the lint CLI flags out of argv into a lintArgs.
// It uses a fresh FlagSet (not the global flag.CommandLine) so it is
// callable more than once per process, and ContinueOnError so a bad flag
// returns a fatal exit code instead of os.Exit-ing past caller cleanup.
// A non-zero fatalExitCode means the caller should return it immediately
// (the diagnostic was already printed to stderr).
func parseLintFlags(argv []string) (args lintArgs, help bool, fatalExitCode int) {
	fs := flag.NewFlagSet("rslint", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var ruleFlags repeatedFlag

	fs.StringVar(&args.Format, "format", "default", "output format")
	fs.StringVar(&args.Config, "config", "", "which rslint config to use")
	fs.StringVar(&args.Config, "c", "", "which rslint config to use")
	fs.BoolVar(&args.Init, "init", false, "initialize a default config in the current directory")
	fs.BoolVar(&args.Fix, "fix", false, "automatically fix problems")
	fs.BoolVar(&args.TypeCheck, "type-check", false, "enable TypeScript type checking")
	fs.BoolVar(&args.TypeCheckOnly, "type-check-only", false, "run only TypeScript type checking (skip all lint rules)")
	fs.BoolVar(&help, "help", false, "show help")
	fs.BoolVar(&help, "h", false, "show help")
	fs.BoolVar(&args.NoColor, "no-color", false, "disable colored output")
	fs.BoolVar(&args.ForceColor, "force-color", false, "force colored output")
	fs.BoolVar(&args.Quiet, "quiet", false, "report errors only")
	var timingValue string
	fs.StringVar(&timingValue, "timing", "", "print a per-rule timing table: 'all' or a top rule count")
	fs.IntVar(&args.MaxWarnings, "max-warnings", -1, "Number of warnings to trigger nonzero exit code")

	fs.StringVar(&args.TraceOut, "trace", "", "file to put trace to")
	fs.StringVar(&args.CpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	fs.BoolVar(&args.SingleThreaded, "singleThreaded", false, "run in single threaded mode")
	fs.Int64Var(&args.StartTimeMs, "start-time", 0, "internal: epoch milliseconds from Node.js entry point")
	fs.Var(&ruleFlags, "rule", "rule override, e.g. 'no-console: error' (repeatable)")

	if err := fs.Parse(argv); err != nil {
		// ContinueOnError: fs already printed the diagnostic to stderr.
		return args, help, 2
	}
	args.RuleFlags = []string(ruleFlags)

	// The Node.js entry point fills in the default "all" when the user
	// passes a bare --timing, so the value is always present here.
	switch {
	case timingValue == "":
	case strings.EqualFold(timingValue, "all"):
		args.Timing = true
	default:
		n, err := strconv.Atoi(timingValue)
		if err != nil || n <= 0 {
			fmt.Fprintf(os.Stderr, "invalid value %q for flag -timing: expected \"all\" or a positive rule count\n", timingValue)
			return args, help, 2
		}
		args.Timing = true
		args.TimingLimit = n
	}

	// --type-check-only implies --type-check and skips all lint rules.
	// Reject incompatible flag combinations before doing any work.
	if code, msg := validateTypeCheckOnlyFlags(args.TypeCheckOnly, args.Fix, args.RuleFlags); code != 0 {
		fmt.Fprintln(os.Stderr, msg)
		return args, help, code
	}
	if args.TypeCheckOnly {
		args.TypeCheck = true
	}

	// Collect file/directory arguments for targeted linting (e.g. rslint file1.ts src/)
	if positional := fs.Args(); len(positional) > 0 {
		for _, arg := range positional {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path %s: %v\n", arg, err)
				return lintArgs{}, help, 1
			}
			// NOTE: we intentionally do NOT call filepath.EvalSymlinks here.
			// EvalSymlinks resolves symlinks (macOS /tmp → /private/tmp) and
			// Windows 8.3 short names to long names, but the rest
			// of the pipeline (os.Getwd, TypeScript file names, configDir)
			// uses unresolved CWD-based paths. Resolving only file args would
			// create a format mismatch causing failures in lint-target detection,
			// config matching, dir scoping, and gitignore checks.
			// Edge cases (e.g. user passes a symlink-resolved absolute path)
			// are handled by isFileAllowed's os.SameFile fallback in linter.go
			// and program file index realpath aliases during target binding.
			normalized := tspath.NormalizePath(absPath)
			info, statErr := os.Stat(absPath)
			if statErr == nil && info.IsDir() {
				args.AllowDirs = append(args.AllowDirs, normalized)
			} else {
				args.AllowFiles = append(args.AllowFiles, normalized)
			}
		}
	}

	return args, help, 0
}

// executeLintPipeline runs the full lint flow (config load → program build →
// lint target planning/Program loading → lint → optional --fix loop → report) and
// returns the process exit code. Shared by the IPC entry (runCLI) and the wasm
// native fallback.
func executeLintPipeline(args lintArgs, ctx context.Context, dispatch linter.EslintPluginDispatcher) int {
	// Unpack into locals so the pipeline body below stays verbatim — only the
	// flag-parse front matter lives in parseLintFlags.
	init := args.Init
	config := args.Config
	configCatalog := args.ConfigCatalog
	usesJSConfig := configCatalog != nil && (configCatalog.Explicit || len(configCatalog.Configs) > 0)
	fix := args.Fix
	typeCheck := args.TypeCheck
	typeCheckOnly := args.TypeCheckOnly
	traceOut := args.TraceOut
	cpuprofOut := args.CpuprofOut
	singleThreaded := args.SingleThreaded
	quiet := args.Quiet
	maxWarnings := args.MaxWarnings
	startTimeMs := args.StartTimeMs
	ruleFlags := args.RuleFlags
	allowFiles := args.AllowFiles
	allowDirs := args.AllowDirs
	// --timing enables per-rule timing. One collector is shared across the
	// initial lint and every --fix re-lint pass, so the table reflects total
	// rule cost for the whole run.
	var timingCollector *linter.TimingCollector
	timingLimit := args.TimingLimit
	if args.Timing && !typeCheckOnly {
		timingCollector = linter.NewTimingCollector()
	}
	format := output.FormatDefault
	if !init {
		var formatErr error
		format, formatErr = output.ParseFormat(args.Format)
		if formatErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", formatErr)
			return 2
		}
	}

	// Resolve color against the real output destination. In native CLI runs
	// the Go process writes to an IPC pipe, so the Node host supplies the TTY
	// fact for the stdout that ultimately receives forwarded output frames.
	forceColorEnv, forceColorEnvSet := os.LookupEnv("FORCE_COLOR")
	_, noColorEnvSet := os.LookupEnv("NO_COLOR")
	colorEnabled := term.ResolveColorEnabled(term.ColorInputs{
		NoColorFlag:      args.NoColor,
		ForceColorFlag:   args.ForceColor,
		ForceColorEnv:    forceColorEnv,
		ForceColorEnvSet: forceColorEnvSet,
		NoColorEnvSet:    noColorEnvSet,
		GithubActionsEnv: os.Getenv("GITHUB_ACTIONS"),
		Term:             os.Getenv("TERM"),
		StdoutIsTTY:      args.StdoutIsTTY,
	})

	enableVirtualTerminalProcessing()
	timeBefore := resolveStartTime(startTimeMs)

	if traceOut != "" {
		f, err := os.Create(traceOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating trace file: %v\n", err)
			return 1
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if cpuprofOut != "" {
		f, err := os.Create(cpuprofOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating cpuprof file: %v\n", err)
			return 1
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error starting cpu profiling: %v\n", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)
	workingDirectory := currentDirectory

	if init {
		if err := rslintconfig.InitDefaultConfig(currentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing config: %v\n", err)
			return 1
		}
		return 0
	}

	fs := args.FS
	if fs == nil {
		fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	}

	// The run-scoped loader owns source snapshots, compiler metadata, Program
	// construction, target binding, and fix-generation rebuilds. Integrations
	// consume only its unified Program results.
	programSession := loader.NewSession(fs)
	fs = programSession.FS()

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllRules()
	// Register placeholder rules for mounted ESLint plugins so their rule
	// names resolve (and route to the Node worker) instead of being dropped.
	var eslintPlugins []rslintconfig.EslintPluginEntry
	if usesJSConfig {
		eslintPlugins = configCatalog.EslintPlugins
	}
	rslintconfig.RegisterEslintPluginRules(eslintPlugins)
	var rslintConfig rslintconfig.RslintConfig

	// configMap holds per-directory configs for automatically discovered JS/TS
	// configs. Explicit JS/TS and JSON/JSONC configs use rslintConfig instead.
	var configMap map[string]rslintconfig.RslintConfig

	var configTargetScopes map[string]rslintconfig.LintDiscoveryScope

	// Program-wide type checking builds every configured project. Plain linting
	// waits for target discovery and builds only the projects owned by configs
	// that govern at least one selected target.
	var projectSet loader.ProjectSet
	buildAllPrograms := typeCheck || typeCheckOnly

	if usesJSConfig {
		configDirectories := configCatalog.ConfigDirectories()
		if configCatalog.Explicit {
			if len(configDirectories) != 1 {
				fmt.Fprintf(os.Stderr, "error: explicit config catalog contains %d configs, want exactly one\n", len(configDirectories))
				return 1
			}
			currentDirectory = configDirectories[0]
			rslintConfig = slices.Clone(configCatalog.Configs[currentDirectory])
			if buildAllPrograms {
				projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		} else {
			configMap = make(map[string]rslintconfig.RslintConfig, len(configCatalog.Configs))
			configTargetScopes = make(map[string]rslintconfig.LintDiscoveryScope, len(configCatalog.Scopes))
			for _, configDir := range configDirectories {
				configMap[configDir] = slices.Clone(configCatalog.Configs[configDir])
				if scope, ok := configCatalog.Scopes[configDir]; ok {
					scope.Files = slices.Clone(scope.Files)
					configTargetScopes[configDir] = scope
				}
			}

			if buildAllPrograms {
				projectSet, err = programSession.BuildProjects(configMap, singleThreaded)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					return 1
				}
			}
		}
	} else {
		// Load configuration from file (JSON config path, isJSConfig stays false)
		loader := rslintconfig.NewConfigLoader(fs, currentDirectory)
		rslintConfig, currentDirectory, err = loader.LoadRslintConfiguration(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if config != "" {
			// Explicit --config follows flat-config invocation semantics: file,
			// ignore, project, and implicit no-args target scope are rooted at
			// the cwd where rslint was invoked, not the config file's directory.
			currentDirectory = workingDirectory
		}

		var exactTargetFiles []string
		if len(allowFiles) > 0 && len(allowDirs) == 0 {
			exactTargetFiles = allowFiles
		}
		if typeCheckOnly {
			projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
		} else if buildAllPrograms {
			rslintConfig, projectSet, err = loadGitignoreAndProjects(
				rslintConfig, currentDirectory, exactTargetFiles, singleThreaded, programSession,
			)
		} else {
			rslintConfig = rslintconfig.ConfigWithGitignore(rslintConfig, currentDirectory, fs, exactTargetFiles)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}

	targetConfigMap := cloneConfigMap(configMap)
	targetRslintConfig := slices.Clone(rslintConfig)

	// Apply --rule CLI overrides by appending a synthetic ConfigEntry. Target
	// discovery below intentionally uses the pre-override config snapshots
	// above, so --rule overlays already-selected lint targets without widening
	// discovery by itself.
	if len(ruleFlags) > 0 {
		cliEntry, err := rslintconfig.BuildCLIRuleEntry(ruleFlags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if cliEntry != nil {
			if configMap != nil {
				for dir, cfg := range configMap {
					configMap[dir] = append(cfg, *cliEntry)
				}
			} else {
				rslintConfig = append(rslintConfig, *cliEntry)
			}
		}
	}

	// Validate every configured rule's options against its declared schema —
	// a separate step after configuration is fully resolved (including --rule
	// overrides) and before any linting starts, so a bad config fails fast
	// with every failure reported at once instead of surfacing mid-lint.
	var optionsMessages []string
	configMap, rslintConfig, optionsMessages = validateResolvedRuleOptions(configMap, rslintConfig)
	if len(optionsMessages) > 0 {
		for _, message := range optionsMessages {
			fmt.Fprintf(os.Stderr, "error: %s\n", message)
		}
		return 1
	}

	// Use CWD for display paths (not any config directory).
	// In multi-config mode, currentDirectory was never reassigned from os.Getwd(),
	// so it already holds the normalized CWD.
	cwd := workingDirectory

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: true,
	}
	broadProjectLoad := isBroadProjectLoadScope(
		allowFiles,
		allowDirs,
		cwd,
		fs.UseCaseSensitiveFileNames(),
	)

	// No args → implicit CWD scoping (same as `rslint .`), matching ESLint.
	// This keeps an explicit --config outside the current directory from
	// widening the scanned root to the config file's directory.
	if len(allowFiles) == 0 && len(allowDirs) == 0 {
		allowDirs = []string{cwd}
	}

	// --- Lint target discovery and Program loading ---
	programs := projectSet.Programs()
	programConfigMap := configMap
	buildSingleConfigPrograms := buildAllPrograms
	var (
		targetPlan                 rslintconfig.LintTargetPlan
		loadedPrograms             loader.LoadResult
		targetsByProgram           [][]string
		targetPathBySourcePath     map[string]string
		configPathBySourcePath     map[string]string
		ownerConfigDirBySourcePath map[string]string
	)
	// --type-check-only is program-wide and pays no lint-target discovery,
	// target binding/parsing, config-resolution, or Program-loading cost.
	if !typeCheckOnly {
		targetPlan, err = rslintconfig.ResolveLintTargetPlan(
			targetConfigMap,
			targetRslintConfig,
			currentDirectory,
			configTargetScopes,
			fs,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if !buildAllPrograms {
			if configMap != nil {
				programConfigMap = targetPlan.ActiveConfigs(configMap)
				if broadProjectLoad {
					projectSet, err = programSession.BuildProjects(programConfigMap, singleThreaded)
				} else {
					projectSet, err = programSession.BuildTargetProjects(programConfigMap, targetPlan, singleThreaded)
				}
			} else if len(targetPlan.Targets) > 0 {
				buildSingleConfigPrograms = true
				if broadProjectLoad {
					projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
				} else {
					projectSet, err = programSession.BuildTargetProject(currentDirectory, rslintConfig, targetPlan, singleThreaded)
				}
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		}
		loadedPrograms, err = programSession.LoadCLI(projectSet, targetPlan, currentDirectory, singleThreaded)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		programs = loadedPrograms.Programs
		targetsByProgram = loadedPrograms.TargetsByProgram
		targetPathBySourcePath = loadedPrograms.TargetPathBySourcePath
		configPathBySourcePath = loadedPrograms.ConfigPathBySourcePath
		ownerConfigDirBySourcePath = loadedPrograms.OwnerConfigDirBySourcePath
	}

	// Rebuild ts-go Programs and bind the original stable target plan again on
	// every fix pass. A target can move between rslint Programs when fixes
	// change the import graph.
	createPrograms := func() (loader.LoadResult, error) {
		var rebuilt loader.ProjectSet
		var err error
		if configMap != nil {
			if buildAllPrograms || broadProjectLoad {
				rebuilt, err = programSession.BuildProjects(programConfigMap, singleThreaded)
			} else {
				rebuilt, err = programSession.BuildTargetProjects(programConfigMap, targetPlan, singleThreaded)
			}
		} else if buildSingleConfigPrograms {
			if buildAllPrograms || broadProjectLoad {
				rebuilt, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
			} else {
				rebuilt, err = programSession.BuildTargetProject(currentDirectory, rslintConfig, targetPlan, singleThreaded)
			}
		}
		if err != nil {
			return loader.LoadResult{}, err
		}
		return programSession.LoadCLI(rebuilt, targetPlan, currentDirectory, singleThreaded)
	}

	// Phase 1: Collect all diagnostics (no printing yet).
	// Like ESLint, diagnostics are collected first, then printed at the end.
	// This ensures --fix only shows remaining unfixed issues.
	var allDiags []rule.RuleDiagnostic
	var diagsMu sync.Mutex
	fixedCount := 0

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range diagnosticsChan {
			allDiags = append(allDiags, d)
		}
	}()

	enforcePlugins := usesJSConfig
	fileConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap:                  configMap,
		Config:                     rslintConfig,
		CurrentDirectory:           currentDirectory,
		EnforcePlugins:             enforcePlugins,
		ConfigPathBySourcePath:     configPathBySourcePath,
		OwnerConfigDirBySourcePath: ownerConfigDirBySourcePath,
		SourceMappingsCanonical:    true,
		FS:                         fs,
	})
	getRulesForFile := func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
		return fileConfigResolver.EnabledRulesForFile(sourceFile.FileName())
	}

	// Target discovery already excluded default paths, global ignores, and
	// .gitignore entries. Target ownership and deduplication were already
	// resolved in targetsByProgram.
	// Source-only Programs expose no checker or
	// program-wide diagnostic capability; the linter consumes that distinction
	// through the common Program contract rather than a parallel skip mask.
	syntaxDiagnostics := collectTargetSyntacticDiagnostics(
		programs,
		targetsByProgram,
		typeCheck,
		typeCheckOnly,
	)
	for _, diagnostic := range syntaxDiagnostics {
		diagnosticsChan <- diagnostic
	}

	// In --type-check-only mode, skip the lint phase entirely by passing
	// nil for GetRulesForFile. RunLinter's Phase 1 is gated on this being
	// non-nil; Phase 2 (type-check) runs independently.
	var rulesForFile linter.RuleHandler
	if !typeCheckOnly {
		rulesForFile = getRulesForFile
	}

	nativeEditDemand := rule.EditDemandNone
	if fix {
		nativeEditDemand = rule.EditDemandAutofix
	}
	runOpts := linter.RunLinterOptions{
		Programs:        programs,
		SingleThreaded:  singleThreaded,
		Cwd:             cwd,
		Scope:           linter.FileScope{Files: allowFiles, Dirs: allowDirs},
		TargetFiles:     targetsByProgram,
		GetRulesForFile: rulesForFile,
		TypeCheck:       typeCheck,
		Timing:          timingCollector,
		Consumer: rule.DiagnosticConsumer{
			Demand: nativeEditDemand,
			Report: func(d rule.RuleDiagnostic) {
				diagnosticsChan <- d
			},
		},
	}
	preparedPlan, planErr := linter.PrepareLintPlan(runOpts)
	if planErr != nil {
		close(diagnosticsChan)
		wg.Wait()
		fmt.Fprintf(os.Stderr, "error preparing lint plan: %v\n", planErr)
		return 1
	}
	runOpts.PreparedPlan = preparedPlan
	// Dispatch eslint-plugin rules to the Node worker in parallel with the
	// native lint pass; results are awaited + merged before output / --fix.
	// ONLY when plugins are actually configured — otherwise the whole reverse-
	// dispatch is skipped so the native-only path pays nothing for the feature.
	// Both paths consume the same prepared file/rule plan.
	hasEslintPlugins := len(eslintPlugins) > 0
	pluginResolver := pluginConfigResolver{
		lintResolver: fileConfigResolver,
	}
	var pluginCh <-chan []rule.RuleDiagnostic
	if hasEslintPlugins {
		pluginInputs := buildPluginFileInputs(runOpts.PreparedPlan, pluginResolver)
		pluginCh = dispatchPluginLintAsync(ctx, dispatch, pluginInputs, fix, pluginSuggestionsMode(fix), timingCollector)
	}

	lintResult, err := linter.RunLinter(runOpts)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	lintedfileCount := lintResult.LintedFileCount

	wg.Wait()
	// Merge eslint-plugin diagnostics (dispatched in parallel) now that the
	// native diagnostics goroutine has drained.
	if pluginCh != nil {
		allDiags = append(allDiags, (<-pluginCh)...)
	}
	remapDiagnosticTargetPaths(allDiags, targetPathBySourcePath, fs)

	// Emit per-file warnings for CLI-specified files that won't be linted.
	// Distinguishes "not found in project" vs "ignored by pattern", aligned
	// with ESLint v10's warning behavior. Skipped in --type-check-only mode:
	// these are lint-phase concepts and would mislead users about Phase 2
	// (which runs program-wide regardless of CLI scope and rslint ignores).
	if !typeCheckOnly {
		warnings := collectAllowFileWarnings(allowFiles, configMap, rslintConfig, currentDirectory, fs.UseCaseSensitiveFileNames(), fs)
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, formatAllowFileWarning(w, comparePathOptions))
		}
	}
	scopeRestricted := len(allowFiles) > 0 || len(allowDirs) > 0
	if shouldShortCircuitOutput(typeCheckOnly, typeCheck, scopeRestricted, lintedfileCount) {
		return 0
	}

	// Phase 2: Apply fixes if --fix flag is enabled.
	// Uses multi-pass fixing: after applying fixes, rebuild programs and re-lint
	// to catch cascading issues (e.g. no-wrapper-object-types fix triggers no-inferrable-types).
	// After fixing, allDiags is replaced with remaining (unfixed) diagnostics.
	const maxFixPasses = 10
	if fix && len(allDiags) > 0 {
		diagnosticsByFile := groupDiagsByFile(allDiags)
		passFixed, fixErr := applyFixPass(diagnosticsByFile, fs)
		// Replace the entire source generation after every write attempt and
		// before any Program rebuild. os.WriteFile may truncate or partially
		// mutate a file even when it ultimately returns an error, and whole-
		// generation invalidation also covers caller/source/symlink aliases.
		programSession.InvalidateSourceSnapshots()
		if fixErr != nil {
			fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
			return 1
		}
		fixedCount += passFixed

		// Re-lint → fix → re-lint → fix → ... until stable or maxFixPasses.
		// Skip if nothing was fixed in the first pass (no need to re-lint).
		for pass := 1; pass <= maxFixPasses && fixedCount > 0; pass++ {
			newBinding, err := createPrograms()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error rebuilding Programs after fixes: %v\n", err)
				return 1
			}
			if len(newBinding.Programs) == 0 {
				fmt.Fprintln(os.Stderr, "error rebuilding Programs after fixes: no Program returned")
				return 1
			}

			// Re-lint using the fresh binding derived from the stable target plan.
			fixTargetsByProgram := newBinding.TargetsByProgram
			fixTargetPathBySourcePath := newBinding.TargetPathBySourcePath
			fixConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
				ConfigMap:                  configMap,
				Config:                     rslintConfig,
				CurrentDirectory:           currentDirectory,
				EnforcePlugins:             enforcePlugins,
				ConfigPathBySourcePath:     newBinding.ConfigPathBySourcePath,
				OwnerConfigDirBySourcePath: newBinding.OwnerConfigDirBySourcePath,
				SourceMappingsCanonical:    true,
				FS:                         fs,
			})
			fixGetRulesForFile := func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
				return fixConfigResolver.EnabledRulesForFile(sourceFile.FileName())
			}
			var fixRulesForFile linter.RuleHandler
			if !typeCheckOnly {
				fixRulesForFile = fixGetRulesForFile
			}
			passEditDemand := rule.EditDemandAutofix
			if pass == maxFixPasses {
				// This pass only verifies the bytes written by the final
				// allowed write pass; no further fixes can be applied.
				passEditDemand = rule.EditDemandNone
			}
			var passDiags []rule.RuleDiagnostic
			fixSyntaxDiagnostics := collectTargetSyntacticDiagnostics(
				newBinding.Programs,
				fixTargetsByProgram,
				typeCheck,
				typeCheckOnly,
			)
			passDiags = append(passDiags, fixSyntaxDiagnostics...)
			fixRunOpts := linter.RunLinterOptions{
				Programs:        newBinding.Programs,
				SingleThreaded:  singleThreaded,
				Cwd:             cwd,
				Scope:           linter.FileScope{Files: allowFiles, Dirs: allowDirs},
				TargetFiles:     fixTargetsByProgram,
				GetRulesForFile: fixRulesForFile,
				TypeCheck:       typeCheck,
				Timing:          timingCollector,
				Consumer: rule.DiagnosticConsumer{
					Demand: passEditDemand,
					Report: func(d rule.RuleDiagnostic) {
						diagsMu.Lock()
						passDiags = append(passDiags, d)
						diagsMu.Unlock()
					},
				},
			}
			fixPreparedPlan, planErr := linter.PrepareLintPlan(fixRunOpts)
			if planErr != nil {
				fmt.Fprintf(os.Stderr, "error preparing lint plan after fixes: %v\n", planErr)
				return 1
			}
			fixRunOpts.PreparedPlan = fixPreparedPlan
			// Re-dispatch plugin rules each pass (only when configured): the
			// worker re-reads the post-fix file content, and merging here keeps
			// plugin diagnostics from being lost when allDiags is replaced.
			// Each pass prepares a fresh plan for the rebuilt target binding.
			var fixPluginCh <-chan []rule.RuleDiagnostic
			if hasEslintPlugins {
				fixPluginInputs := buildPluginFileInputs(fixRunOpts.PreparedPlan, pluginConfigResolver{
					lintResolver: fixConfigResolver,
				})
				fixPluginCh = dispatchPluginLintAsync(ctx, dispatch, fixPluginInputs, fix, pluginSuggestionsMode(fix), timingCollector)
			}
			passResult, passErr := linter.RunLinter(fixRunOpts)
			var fixPluginDiags []rule.RuleDiagnostic
			if fixPluginCh != nil {
				fixPluginDiags = <-fixPluginCh
			}
			if passErr != nil {
				fmt.Fprintf(os.Stderr, "error running linter after fixes: %v\n", passErr)
				return 1
			}
			if passResult != nil {
				for name := range passResult.ExecutedRules {
					lintResult.ExecutedRules[name] = struct{}{}
				}
			}
			// Merge this pass's plugin diagnostics before applying fixes so
			// plugin fixes participate and plugin diagnostics survive.
			passDiags = append(passDiags, fixPluginDiags...)
			remapDiagnosticTargetPaths(passDiags, fixTargetPathBySourcePath, fs)

			// Replace allDiags with latest post-fix diagnostics.
			allDiags = passDiags
			if pass == maxFixPasses {
				// The maximum number of write passes has already run (the initial
				// pass plus maxFixPasses-1 loop passes). This extra pass is the
				// required final verification of the bytes written by pass 10.
				break
			}

			passFixed, fixErr := applyFixPass(groupDiagsByFile(passDiags), fs)
			// See the first fix pass above: invalidate before inspecting the
			// result so a partially successful write can never feed a rebuild.
			programSession.InvalidateSourceSnapshots()
			if fixErr != nil {
				fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
				return 1
			}
			if passFixed == 0 {
				break // Stable — allDiags reflect final state
			}
			fixedCount += passFixed
		}
	}

	allDiags = deduplicateTypeScriptDiagnostics(allDiags, fs, targetPlan.PreferredCallerPaths())

	// Diagnostics arrive in completion order — programs and, within a
	// program, file shards run in parallel — so impose a deterministic
	// order before printing. The key is (file, start position) only,
	// deliberately with NO end/rule tie-break: ESLint orders same-start
	// diagnostics by emission order (parent reported before nested child),
	// and a file's diagnostics are all emitted by a single worker, so under
	// a STABLE sort this key is already fully deterministic. Keep this
	// comparator in sync with the --api one in api.go (same policy over
	// api.Diagnostic).
	slices.SortStableFunc(allDiags, func(a, b rule.RuleDiagnostic) int {
		if c := strings.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}
		return cmp.Compare(a.Range.Pos(), b.Range.Pos())
	})

	// Phase 3: Build one report from the final post-fix diagnostics, then let
	// the CLI output subsystem own format dispatch, colors, and summary text.
	mode := output.ModeLint
	if typeCheckOnly {
		mode = output.ModeTypeCheckOnly
	} else if typeCheck {
		mode = output.ModeLintAndTypeCheck
	}

	typeCheckedFileCount := 0
	if typeCheck {
		// Count compiler-capable Program root files (tsconfig include/files),
		// not transitive declarations, for every summary that includes type-check.
		seen := make(map[string]struct{})
		for _, prog := range programs {
			if !prog.CanProvideProgramDiagnostics() {
				continue
			}
			for _, fileName := range prog.RootFileNames() {
				seen[fileName] = struct{}{}
			}
		}
		typeCheckedFileCount = len(seen)
	}

	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}

	report := output.NewReport(allDiags, output.Metadata{
		Mode:             mode,
		LintedFiles:      int(lintedfileCount),
		TypeCheckedFiles: typeCheckedFileCount,
		Rules:            len(lintResult.ExecutedRules),
		Threads:          threadsCount,
		FixedIssues:      fixedCount,
		StartedAt:        timeBefore,
	})
	if err := output.Render(os.Stdout, report, output.Options{
		Format:       format,
		ComparePaths: comparePathOptions,
		Quiet:        quiet,
		ColorEnabled: colorEnabled,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error writing lint report: %v\n", err)
		return 1
	}
	// The timing table goes to stderr so machine-readable stdout formats
	// (jsonline/github/gitlab) stay parseable with --timing enabled.
	if timingCollector != nil {
		table := output.FormatRuleTimingTable(timingCollector.Timings(), timingLimit)
		if args.DeferTimingTable != nil {
			args.DeferTimingTable(table)
		} else {
			fmt.Fprint(os.Stderr, table)
		}
	}
	counts := report.Counts()

	tooManyWarnings := maxWarnings >= 0 && counts.Warnings > maxWarnings

	if counts.Errors == 0 && tooManyWarnings {
		fmt.Fprintf(os.Stderr, "Rslint found too many warnings (maximum: %d).\n", maxWarnings)
	}

	// Exit with non-zero status code if errors were found
	if counts.Errors > 0 || tooManyWarnings {
		return 1
	}
	return 0
}
