// Package term owns the CLI's color-output decision.
//
// The decision is made exactly once, at lint-pipeline entry, by
// ResolveColorEnabled. Nothing else in the process may mutate the resulting
// color state afterwards; in particular, fatih/color's package-init guess
// (which keys off the Go process's own stdout — a pipe in IPC mode, never the
// user's terminal) is unconditionally overwritten by this decision.
package term

// ColorInputs carries every signal the color decision consumes. The caller
// collects flags and raw environment values once; this package owns their
// semantics. Env values are passed as (value, set) pairs where "set to empty
// string" and "unset" differ in meaning.
type ColorInputs struct {
	// NoColorFlag and ForceColorFlag mirror the --no-color / --force-color
	// CLI flags.
	NoColorFlag    bool
	ForceColorFlag bool

	// ForceColorEnv / ForceColorEnvSet mirror os.LookupEnv("FORCE_COLOR").
	ForceColorEnv    string
	ForceColorEnvSet bool

	// NoColorEnvSet is true when NO_COLOR is present in the environment,
	// including set-but-empty (Node/ESLint v10 semantics; deliberately
	// stricter than no-color.org's "non-empty" wording).
	NoColorEnvSet bool

	// GithubActionsEnv mirrors os.Getenv("GITHUB_ACTIONS").
	GithubActionsEnv string

	// Term mirrors os.Getenv("TERM").
	Term string

	// StdoutIsTTY is the fact reported by the owner of the real output fd.
	// In IPC mode that owner is the Node host (its process.stdout receives
	// the forwarded lint output), so the value travels in the init payload;
	// when unknown (old peer, wasm build) it stays false.
	StdoutIsTTY bool
}

// ResolveColorEnabled reports whether CLI output should be colorized.
//
// Precedence, strongest first — each tier is consulted only when every
// stronger tier is absent, and deciding tiers terminate the walk:
//
//  1. --no-color flag        → off
//  2. --force-color flag     → on
//  3. FORCE_COLOR env set    → by value: "", "1", "true", "2", "3" → on;
//     "0", "false", unrecognized → off
//  4. NO_COLOR env set       → off (any value, including empty)
//  5. TERM == "dumb"         → off
//  6. GITHUB_ACTIONS env non-empty → on (any non-empty value, even "false")
//  7. stdout-is-TTY fact     → on/off
//
// References: explicit flags beating env follows ESLint v10 (whose stylish
// formatter colors via node:util styleText; an explicit --color there ignores
// NO_COLOR). FORCE_COLOR value semantics and FORCE_COLOR-over-NO_COLOR follow
// Node (which warns that NO_COLOR is ignored when FORCE_COLOR is set);
// supports-color/chalk agree on the FORCE_COLOR values.
//
// Documented deviations from ESLint v10:
//   - Tier 6: ESLint emits no color on GitHub Actions (piped stdout); rslint
//     deliberately keeps its existing force-on so workflow logs render ANSI.
//     TERM=dumb still wins over it, matching both references' relative order.
//   - Tier 7: a bare TTY check, with no TERM capability lookup — fine for the
//     16-color output rslint emits; exotic TERM values on a real TTY get
//     color where ESLint might not.
func ResolveColorEnabled(in ColorInputs) bool {
	switch {
	case in.NoColorFlag:
		return false
	case in.ForceColorFlag:
		return true
	case in.ForceColorEnvSet:
		return forceColorValueEnables(in.ForceColorEnv)
	case in.NoColorEnvSet:
		return false
	case in.Term == "dumb":
		return false
	case in.GithubActionsEnv != "":
		return true
	default:
		return in.StdoutIsTTY
	}
}

// forceColorValueEnables maps a FORCE_COLOR value to on/off following Node's
// util.styleText / supports-color semantics: empty (set-but-empty), "1",
// "true" and the higher color levels "2"/"3" enable; "0", "false" and any
// unrecognized value disable.
func forceColorValueEnables(v string) bool {
	switch v {
	case "", "1", "true", "2", "3":
		return true
	default:
		return false
	}
}
