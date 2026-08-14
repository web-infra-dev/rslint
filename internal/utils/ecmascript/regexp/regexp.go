// Package regexp compiles a JavaScript regexp the way JavaScript reads it.
// Import it as esregexp, so the standard library's regexp keeps its own name.
//
// A rule ported from ESLint is handed patterns its users wrote as JavaScript
// regexps, in rule options and in `new RegExp` calls in the source under lint.
// Neither Go engine answers about those the way JavaScript does:
//
//   - The standard library's regexp is RE2, which has no lookahead, no
//     lookbehind and no backreferences. A pattern using any of them does not
//     compile at all, and a rule that shrugs off the compile error silently
//     drops whatever option it came from.
//   - regexp2 backtracks and so takes the syntax, and its ECMAScript mode gets
//     most of the semantics right. What it still reads differently:
//     a `.` matches U+2028 and U+2029; `^` and `$` under `m` break on `\n`
//     alone where JavaScript breaks on all four line terminators; ignoring
//     case folds by Unicode's rules rather than JavaScript's; `\b` is drawn
//     over .NET's Unicode word set where ECMAScript keeps it to ASCII; an
//     escape .NET reads for itself — `\A`, `\a`, `\e` — is a plain character
//     to JavaScript, and one .NET does not know at all makes it refuse a
//     pattern JavaScript accepts; and `y` says where a match may start, which
//     regexp2 has no flag for.
//
// This package takes the second one and closes every one of those gaps by
// rewriting the pattern before handing it over. It also bounds how long a
// match may run: a backtracking engine on a pattern a user wrote is a way to
// hang the linter, and a match that overruns is reported as no match.
//
// Three things are deliberately not covered:
//
//   - The `v` flag's set syntax, which Compile refuses outright.
//   - Comparing a backreference, or a `\p{…}` property escape, without regard
//     to case. JavaScript compares a backreference by the same canonicalization
//     it compares a literal by, and draws a property escape from Unicode's own
//     tables; a rewritten pattern can spell neither, so a pattern using one
//     under `i` falls back to regexp2's own case-insensitivity, which is close.
//     `\P{…}` under `i` is where it is furthest off: JavaScript asks whether
//     any character comparing equal to this one is outside the property, and
//     regexp2 asks only about the one at hand. The fallback belongs to the
//     whole pattern, so a `(?i:…)` group in a pattern that does not itself
//     carry `i` has none to make, and compares both exactly.
//   - Exhaustive syntax validation under `u`. The `u` flag makes JavaScript
//     reject constructs it otherwise tolerates — a bare `]` outside a class is
//     the usual one — and those still compile here. Every pattern JavaScript
//     accepts is read correctly; some it rejects are read rather than refused.
package regexp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

// MatchTimeout bounds how long one subject may spend in one pattern. A pattern
// written by a user can backtrack exponentially — `(a|aa)+$` against a long
// run of `a` is the shape of it — and a linter that hangs on a config typo is
// worse than one that answers no. A pattern that matches settles in
// microseconds, so a match still running after this has found nothing.
const MatchTimeout = time.Second

// ErrUnsupportedFlag is returned for a flag this package does not implement.
var ErrUnsupportedFlag = errors.New("unsupported regexp flag")

// ErrUnsupportedSyntax is returned for syntax JavaScript itself rejects.
var ErrUnsupportedSyntax = errors.New("unsupported regexp syntax")

// RegExp is a compiled JavaScript regexp.
type RegExp struct {
	source string
	flags  string
	re     *regexp2.Regexp
	// exact is false when the pattern fell back to regexp2's own
	// case-insensitivity, which is close to JavaScript's but not identical.
	exact bool
}

type flagSet struct {
	ignoreCase bool
	multiline  bool
	dotAll     bool
	unicode    bool
	sticky     bool
}

// parseFlags reads a JavaScript flag string. `g` and `d` say how a match is
// driven rather than what matches, so they are accepted and ignored. `y` does
// change what matches and is kept.
func parseFlags(flags string) (flagSet, error) {
	var set flagSet
	seen := map[rune]bool{}
	for _, flag := range flags {
		if seen[flag] {
			return set, fmt.Errorf("%w: %q given twice", ErrUnsupportedFlag, flag)
		}
		seen[flag] = true
		switch flag {
		case 'i':
			set.ignoreCase = true
		case 'm':
			set.multiline = true
		case 's':
			set.dotAll = true
		case 'u':
			set.unicode = true
		case 'y':
			set.sticky = true
		case 'g', 'd':
		case 'v':
			return set, fmt.Errorf("%w: %q", ErrUnsupportedFlag, flag)
		default:
			return set, fmt.Errorf("%w: %q", ErrUnsupportedFlag, flag)
		}
	}
	return set, nil
}

// Compile compiles source under the given JavaScript flags, as
// `new RegExp(source, flags)` would. Pass "" for no flags.
func Compile(source string, flags string) (*RegExp, error) {
	set, err := parseFlags(flags)
	if err != nil {
		return nil, err
	}

	rewritten, exact, err := rewrite(source, rewriteOptions{
		multiline:  set.multiline,
		dotAll:     set.dotAll,
		ignoreCase: set.ignoreCase,
		unicode:    set.unicode,
	})
	if err != nil {
		return nil, err
	}
	if set.sticky {
		// A sticky regexp matches only at `lastIndex`, which is 0 for one just
		// compiled and never moves here, since nothing this package exposes
		// advances it. So it anchors at the start — around the whole pattern,
		// or a top-level `|` would anchor only its first branch.
		rewritten = inputStart + `(?:` + rewritten + `)`
	}

	// Multiline, Singleline and IgnoreCase are all handled by the rewrite, so
	// regexp2 is never told about them — being told twice would undo the work.
	options := regexp2.RegexOptions(regexp2.ECMAScript)
	if set.unicode {
		options |= regexp2.Unicode
	}
	if set.ignoreCase && !exact {
		options |= regexp2.IgnoreCase
	}

	re, err := regexp2.Compile(rewritten, options)
	if err != nil {
		// regexp2 quotes back the pattern it was handed, which is the
		// rewritten one. A message naming `[Oo][Kk]` where its author wrote
		// `ok` is no help, so the source goes back in as written.
		return nil, errors.New(strings.ReplaceAll(err.Error(), rewritten, source))
	}
	re.MatchTimeout = MatchTimeout

	return &RegExp{source: source, flags: flags, re: re, exact: exact}, nil
}

// MustCompile is Compile for a pattern written in this repository rather than
// read from a user, and panics rather than returning an error.
func MustCompile(source string, flags string) *RegExp {
	re, err := Compile(source, flags)
	if err != nil {
		panic(fmt.Sprintf("esregexp: cannot compile /%s/%s: %v", source, flags, err))
	}
	return re
}

// Test reports whether the regexp matches anywhere in s, as
// `RegExp.prototype.test` does.
//
// A match that overruns MatchTimeout counts as no match: the alternative is
// making every caller handle an error that only ever means "this pattern is
// pathological", and a lint rule has nothing better to do with that than skip.
func (r *RegExp) Test(s string) bool {
	if r == nil || r.re == nil {
		return false
	}
	matched, err := r.re.MatchString(s)
	return err == nil && matched
}

// TestOrError is Test for a caller that has to tell "no match" apart from "the
// match ran into MatchTimeout". The distinction matters where a report is
// guarded by an allow-pattern: reading a bound match as no-match would turn a
// skipped report into a false positive, so such a caller wants to fail open
// rather than take the bare answer.
func (r *RegExp) TestOrError(s string) (bool, error) {
	if r == nil || r.re == nil {
		return false, nil
	}
	return r.re.MatchString(s)
}

// Source returns the pattern as it was written, the way `RegExp.prototype
// .source` does — not the rewritten form handed to regexp2.
func (r *RegExp) Source() string { return r.source }

// Flags returns the flags as they were written.
func (r *RegExp) Flags() string { return r.flags }

// String renders the regexp the way JavaScript writes a regexp literal.
func (r *RegExp) String() string { return "/" + r.source + "/" + r.flags }

// Unwrap returns the compiled regexp2 pattern, for a caller that needs
// capture groups or replacement. The pattern it holds is the rewritten one, so
// its own String does not match Source.
func (r *RegExp) Unwrap() *regexp2.Regexp {
	if r == nil {
		return nil
	}
	return r.re
}
