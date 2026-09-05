package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

// lintArgs is the parsed CLI flag set, decoupled from the global flag
// package so each entry point can build it: parseLintFlags from argv, and
// runCLI additionally from the IPC init-handshake payload.
type lintArgs struct {
	Init           bool
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
	// StartWriter, when non-nil, is the integration-owned destination for the
	// default format's start line. The IPC CLI supplies an acknowledged writer
	// so post-start work cannot write inherited stderr before the real stdout
	// destination has completed the start write.
	StartWriter io.Writer
	// ConfigCatalog is the immutable result of Go-owned config discovery.
	// Every lint invocation requires a non-empty automatic catalog or one
	// explicit config entry.
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
  -c, --config PATH     Which JS/TS module config file to use.
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
			// Lexical and canonical aliases are frozen during target discovery;
			// the Program loader then resolves those aliases while binding the
			// exact target projection.
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
