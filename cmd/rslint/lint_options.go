package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/suppressions"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	StdoutIsTTY    bool
	Quiet          bool
	Timing         bool
	TimingLimit    int
	MaxWarnings    int
	StartTimeMs    int64
	RuleFlags      []string
	Suppressions   suppressions.Options
	PassOnUnpruned bool
	// AllowFiles and AllowDirs are prepared target paths. Positional retains raw
	// arguments until the run-scoped filesystem is available.
	AllowFiles      []string
	AllowDirs       []string
	Positional      []string
	GlobRoots       []rslintconfig.DirectoryIdentity
	ScanDirectories map[string][]string
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
	// CompleteConfigActivation validates the CLI's provisional plugin metadata
	// after read-only target/Program preparation and before any lint or fixes.
	// The command also joins it on early failure. Nil means already activated.
	CompleteConfigActivation func() error
}

// repeatedFlag collects multiple values for the same flag (e.g. --rule used multiple times).
type repeatedFlag []string

func (f *repeatedFlag) String() string     { return strings.Join(*f, ", ") }
func (f *repeatedFlag) Set(v string) error { *f = append(*f, v); return nil }

func suppressionBoolFlag(fs *flag.FlagSet, target *bool, name, usage string) {
	for _, spelling := range []string{name, "no-" + name} {
		fs.BoolFunc(spelling, usage, func(value string) error {
			switch ecmascript.StringTrim(value) {
			case "true":
				*target = spelling == name
			case "false":
				*target = spelling != name
			default:
				return errors.New("expected true or false")
			}
			return nil
		})
	}
}

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
  --suppress-all        Record all current lint errors
  --suppress-rule RULE  Record errors for selected rules (repeatable, a,b or "[a,b]")
  --suppressions-location PATH  Use a different suppressions file or directory
  --prune-suppressions  Remove unused suppressions
  --pass-on-unpruned-suppressions  Allow unused suppressions
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
	var suppressRuleFlags repeatedFlag
	var suppressAll, prune bool

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
	suppressionBoolFlag(fs, &suppressAll, "suppress-all", "record all lint errors")
	suppressionBoolFlag(fs, &prune, "prune-suppressions", "remove unused suppressions")
	suppressionBoolFlag(fs, &args.PassOnUnpruned, "pass-on-unpruned-suppressions", "allow unused suppressions")
	fs.StringVar(&args.Suppressions.Location, "suppressions-location", "", "suppressions file or directory")
	fs.Var(&suppressRuleFlags, "suppress-rule", "rule IDs to record (repeatable, a,b or \"[a,b]\")")

	if err := fs.Parse(argv); err != nil {
		// ContinueOnError: fs already printed the diagnostic to stderr.
		return args, help, 2
	}
	args.RuleFlags = []string(ruleFlags)
	if !help && len(suppressRuleFlags) > 0 {
		rules := []string{}
		for _, raw := range suppressRuleFlags {
			parsed, err := parseSuppressRuleList(raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return args, help, 2
			}
			rules = append(rules, parsed...)
		}
		args.Suppressions.Rules = rules
	}
	suppressRuleSet := len(suppressRuleFlags) > 0
	switch {
	case help:
		// The flag values were parsed above. Help does not run a suppression mode.
	case suppressAll && suppressRuleSet || prune && (suppressAll || suppressRuleSet):
		fmt.Fprintln(os.Stderr, "error: choose only one of --suppress-all, --suppress-rule and --prune-suppressions")
		return args, help, 2
	case suppressAll || suppressRuleSet:
		args.Suppressions.Mode = suppressions.Record
	case prune:
		args.Suppressions.Mode = suppressions.Prune
	}
	if !help && args.TypeCheckOnly && args.Suppressions.Mode != suppressions.Check {
		fmt.Fprintln(os.Stderr, "error: --type-check-only cannot record or prune lint suppressions")
		return args, help, 2
	}

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

	args.Positional = fs.Args()

	return args, help, 0
}

// parseSuppressRuleList parses one --suppress-rule occurrence. It accepts a
// single rule ID, a comma-separated list, or an ESLint-style bracketed list
// such as "[no-var, no-debugger]", including quoted entries and an optional
// trailing comma.
func parseSuppressRuleList(raw string) ([]string, error) {
	s := ecmascript.StringTrim(raw)
	if s == "" {
		return []string{}, nil
	}
	if strings.HasPrefix(s, "[") {
		if !strings.HasSuffix(s, "]") {
			return nil, fmt.Errorf("invalid --suppress-rule value %q: missing closing ']'", raw)
		}
		s = ecmascript.StringTrim(s[1 : len(s)-1])
		if s == "" {
			return []string{}, nil
		}
	} else if strings.HasSuffix(s, "]") {
		return nil, fmt.Errorf("invalid --suppress-rule value %q: unexpected ']'", raw)
	}

	rules := []string{}
	for rest := s; ; {
		rest = ecmascript.StringTrim(rest)
		if rest == "" {
			if len(rules) > 0 {
				break // Allow a trailing comma after at least one rule ID.
			}
			return nil, fmt.Errorf("invalid --suppress-rule value %q: empty rule ID", raw)
		}
		var item, after string
		if rest[0] == '"' || rest[0] == '\'' {
			var next int
			var err error
			item, next, err = parseQuotedRuleID(rest, 0, raw)
			if err != nil {
				return nil, err
			}
			after = ecmascript.StringTrim(rest[next:])
			if after != "" && after[0] != ',' {
				return nil, fmt.Errorf("invalid --suppress-rule value %q: unexpected %q after quoted rule ID", raw, after[0])
			}
		} else {
			head := rest
			if comma := strings.IndexByte(rest, ','); comma >= 0 {
				head, after = rest[:comma], rest[comma:]
			}
			if idx := strings.IndexAny(head, "[]"); idx >= 0 {
				return nil, fmt.Errorf("invalid --suppress-rule value %q: unexpected %q", raw, head[idx])
			}
			item = ecmascript.StringTrim(head)
		}
		if item == "" {
			return nil, fmt.Errorf("invalid --suppress-rule value %q: empty rule ID", raw)
		}
		rules = append(rules, item)
		if after == "" {
			break
		}
		rest = after[1:] // skip ','
	}
	return rules, nil
}

func parseQuotedRuleID(s string, start int, raw string) (string, int, error) {
	quote := s[start]
	var out strings.Builder
	for i := start + 1; i < len(s); i++ {
		ch := s[i]
		if ch == quote {
			return out.String(), i + 1, nil
		}
		if ch != '\\' {
			out.WriteByte(ch)
			continue
		}
		i++
		if i >= len(s) {
			return "", 0, fmt.Errorf("invalid --suppress-rule value %q: unterminated escape", raw)
		}
		switch s[i] {
		case '\\', '\'', '"':
			out.WriteByte(s[i])
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case 'u':
			if i+4 >= len(s) {
				return "", 0, fmt.Errorf("invalid --suppress-rule value %q: incomplete unicode escape", raw)
			}
			code, err := strconv.ParseUint(s[i+1:i+5], 16, 32)
			if err != nil {
				return "", 0, fmt.Errorf("invalid --suppress-rule value %q: invalid unicode escape", raw)
			}
			r := rune(code)
			if !utf8.ValidRune(r) {
				return "", 0, fmt.Errorf("invalid --suppress-rule value %q: invalid unicode code point", raw)
			}
			out.WriteRune(r)
			i += 4
		default:
			out.WriteByte(s[i])
		}
	}
	return "", 0, fmt.Errorf("invalid --suppress-rule value %q: unterminated string", raw)
}
