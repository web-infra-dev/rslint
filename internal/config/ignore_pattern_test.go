package config

import (
	"testing"

	"gotest.tools/v3/assert"
)

// Unit tests for ignore_pattern.go: the structured IgnorePattern pipeline and
// the unified directory-prune predicate (canPruneDir) plus its negation-reach
// helpers. The walk integration that consumes them lives in
// file_discovery_dir_prune_test.go.

// --- Unit: ParseIgnorePattern classification ---
// ParseIgnorePattern is the SINGLE source of truth for (raw → Glob, Negated,
// Kind); every predicate downstream reads these fields. Indirect coverage via
// the predicates can mask a misclassification (e.g. target/**/*.ext lands in
// dirAbsoluteBlock but canPruneDir still returns false because the matchGlob
// probe misses), so pin every switch arm and boundary directly here.

func TestParseIgnorePattern_Classification(t *testing.T) {
	cases := []struct {
		raw     string
		glob    string
		negated bool
		kind    dirKind
	}{
		{"build", "build", false, dirAbsoluteBlock},                                  // bare name → default arm
		{"*.log", "*.log", false, dirAbsoluteBlock},                                  // pure file pattern → default arm
		{"**/build", "**/build", false, dirAbsoluteBlock},                            // **/ prefix bare
		{"dir/*", "dir/*", false, dirNone},                                           // single level → none
		{"dir/**", "dir/**", false, dirAbsoluteBlock},                                // dir-level absolute
		{"dir/**/*", "dir/**/*", false, dirFileLevelCover},                           // file-level cover
		{"dir/**/*.ext", "dir/**/*.ext", false, dirAbsoluteBlock},                    // ext filter → not /**/* suffix → default
		{"", "", false, dirNone},                                                     // empty → none arm
		{"!", "", true, dirNone},                                                     // negated-then-empty → none
		{"!dir/**", "dir/**", true, dirAbsoluteBlock},                                // negation does not change Kind
		{"!dir/**/*", "dir/**/*", true, dirFileLevelCover},                           // negated file-level
		{"{a,b}", "{a,b}", false, dirAbsoluteBlock},                                  // brace bare
		{"**/{target,dist}/**/*", "**/{target,dist}/**/*", false, dirFileLevelCover}, // brace name + /**/*
		{"target/**/*.{ts,tsx}", "target/**/*.{ts,tsx}", false, dirAbsoluteBlock},    // brace ext filter
		{"./dir/**", "dir/**", false, dirAbsoluteBlock},                              // ./ normalized away
		{"dir//**", "dir/**", false, dirAbsoluteBlock},                               // // collapsed
		{"node_modules/**", "node_modules/**", false, dirAbsoluteBlock},              // default-ignore glob actual class
	}
	for _, c := range cases {
		p := ParseIgnorePattern(c.raw)
		if p.Glob != c.glob || p.Negated != c.negated || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Glob:%q Negated:%v Kind:%d}, want {Glob:%q Negated:%v Kind:%d}",
				c.raw, p.Glob, p.Negated, p.Kind, c.glob, c.negated, c.kind)
		}
	}
}

// The directory role is classified on the RAW suffix (after `!`-strip), NOT the
// normalized Glob — required for byte-equivalence with the linter's pre-refactor
// isFileLevelPattern. The `./*` case is the regression that drove this: it
// normalizes to bare `*`; classifying the Glob would make it dirAbsoluteBlock
// and silently drop nested files (isDirAbsolutelyBlocked("src") true, but
// isFileIgnored("src/sub/f") false). On the raw `./*` it ends `/*` → dirNone, so
// nothing is over-blocked. Glob remains the normalized matcher. Values verified
// by running ParseIgnorePattern directly.
func TestParseIgnorePattern_RawSuffixClassification(t *testing.T) {
	cases := []struct {
		raw  string
		glob string
		kind dirKind
	}{
		{"./*", "*", dirNone},                               // REGRESSION: raw ends "/*" → none (NOT absolute on bare "*")
		{".//*", "*", dirNone},                              // same, double slash
		{"./{a,b}/../*", "*", dirNone},                      // normalizes to bare "*" but raw ends "/*" → none
		{"weird/*/.", "weird/*", dirAbsoluteBlock},          // raw ends "/." (not "/*" nor "/**/*") → absolute
		{"target/../dist", "dist", dirAbsoluteBlock},        // raw has no dir suffix → absolute
		{"x/./y/**/*", "x/y/**/*", dirFileLevelCover},       // raw ends "/**/*" → file-level cover
		{"a//b", "a/b", dirAbsoluteBlock},                   // raw has no dir suffix → absolute
		{"./target/**/*", "target/**/*", dirFileLevelCover}, // raw ends "/**/*" → still prunable (rspack path safe)
		{"foo/..", "", dirAbsoluteBlock},                    // normalizes to empty but raw non-empty → absolute (empty Glob; matches main's empty-block quirk on absolute dirs)
		{"", "", dirNone},                                   // raw empty → none (matches main's pattern=="" skip)
		{"!", "", dirNone},                                  // !-only → empty body → none
	}
	for _, c := range cases {
		p := ParseIgnorePattern(c.raw)
		if p.Glob != c.glob || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Glob:%q Kind:%d}, want {Glob:%q Kind:%d}",
				c.raw, p.Glob, p.Kind, c.glob, c.kind)
		}
	}
}

// Regression for the silent-skip bug: ignores ["./*"] must NOT cause a nested
// file's dir to be treated as absolutely blocked. The prune decision must agree
// with the file-level decision (sound), and GetConfigForFile must still lint the
// nested file.
func TestParseIgnorePattern_DotStarDoesNotOverBlock(t *testing.T) {
	pats := ParseIgnorePatterns([]string{"./*"})
	if isDirAbsolutelyBlocked("src", pats) {
		t.Error(`"./*" must not absolutely-block dir "src"`)
	}
	if canPruneDir("src", pats, buildNegReach(pats)) {
		t.Error(`"./*" must not prune dir "src"`)
	}
	cfg := RslintConfig{
		{Ignores: []string{"./*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},
	}
	if cfg.GetConfigForFile("/repo/src/app/main.ts", "/repo") == nil {
		t.Error(`nested file src/app/main.ts must still be linted under ignores ["./*"]`)
	}
}

// --- Unit: literalSegmentPrefix ---

func TestLiteralSegmentPrefix(t *testing.T) {
	cases := []struct{ in, want string }{
		{"tests/e2e/**/*", "tests/e2e"},
		{"target/keep/**/*", "target/keep"},
		{"sub/path/target/**/*", "sub/path/target"},
		{"target/important.ts", "target/important.ts"}, // no metachar
		{"target/", "target"},                          // trailing slash trimmed
		{"a/b*c/d", "a"},                               // '*' metachar in 2nd segment
		{"a/b?c/d", "a"},                               // '?' metachar in 2nd segment
		{"tests/{e2e,unit}/**/*", "tests"},             // brace in 2nd segment
		{"*.log", ""},                                  // leading wildcard, no concrete segment
		{"**/keep/**/*", ""},                           // leading **/ (handled as unrooted upstream)
		{"[abc]/x", ""},                                // leading bracket
		{"{a,b}/x", ""},                                // leading brace
		{"a?b/x", ""},                                  // '?' in first segment
	}
	for _, c := range cases {
		if got := literalSegmentPrefix(c.in); got != c.want {
			t.Errorf("literalSegmentPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- Unit: buildNegReach ---

func TestBuildNegReach(t *testing.T) {
	got := buildNegReach(ParseIgnorePatterns([]string{
		"**/target/**/*",                    // positive, ignored
		"!sub/path/target/**/*",             // rooted negation
		"!**/keep/**/*",                     // unrooted negation
		"!*.log",                            // leading wildcard → unrooted (conservative)
		"!tests/rspack-test/*/node_modules", // metachar in middle → literal up to it
		"!./dotslash/dir/**/*",              // normalized → literal "dotslash/dir"
		"!src/{a,b}/keep/**/*",              // brace metachar → literal up to it
	})).prefixes
	want := []negPrefix{
		{literal: "sub/path/target"},
		{unrooted: true},
		{unrooted: true},
		{literal: "tests/rspack-test"},
		{literal: "dotslash/dir"},
		{literal: "src"},
	}
	assert.Equal(t, len(got), len(want), "got %+v", got)
	for i := range want {
		assert.Equal(t, got[i], want[i], "entry %d", i)
	}
}

// --- Unit: canPruneDir predicate table ---
// Mirrors the adversarial review matrix. Each row: ignore set + dir → prune?
// canPruneDir is the UNIFIED predicate: it prunes both absolute blocks (dir/**)
// and negation-aware file-level covers (dir/**/*), so a bare dir/** returns true
// here (the pre-refactor walk pruned absolute blocks via isDirPathBlocked only).

func TestCanPruneDir(t *testing.T) {
	cases := []struct {
		name    string
		ignores []string
		dir     string
		want    bool
	}{
		{"file-level target, no negation", []string{"**/target/**/*"}, "target", true},
		{"file-level target, nested dir", []string{"**/target/**/*"}, "target/build", true},
		{"negation re-includes child of excluded → no prune", []string{"target/**/*", "!target/keep/**/*"}, "target", false},
		{"negation re-includes child → keep dir itself not pruned", []string{"target/**/*", "!target/keep/**/*"}, "target/keep", false},
		{"negation re-includes child → sibling still pruned", []string{"target/**/*", "!target/keep/**/*"}, "target/other", true},
		{"negation full path → top-level target pruned", []string{"**/target/**/*", "!sub/path/target/**/*"}, "target", true},
		{"negation full path → the re-included target not pruned", []string{"**/target/**/*", "!sub/path/target/**/*"}, "sub/path/target", false},
		{"negation full path → ancestor of re-included not pruned", []string{"**/tests/**/*", "!tests/e2e/**/*"}, "tests", false},
		{"unrooted negation → conservative, never prune", []string{"**/build/**/*", "!**/keep/**/*"}, "build", false},
		{"sibling-name negation does not protect", []string{"**/target/**/*", "!targetX/keep/**/*"}, "target", true},
		{"unrelated negation does not protect", []string{"target/**/*", "!dist/keep/**/*"}, "target", true},
		{"directory-level pattern dir/** is an absolute block → prune", []string{"target/**"}, "target", true},
		{"no matching pattern", []string{"**/dist/**/*"}, "target", false},
		{"plain file pattern is not a directory cover", []string{"**/*.log"}, "target", false},
		{"single-star X/* covers one level only → no prune", []string{"build/*"}, "build", false},
		{"extension-filtered X/**/*.ext does not cover subtree", []string{"target/**/*.log"}, "target", false},
		{"dotslash negation protects after normalize", []string{"target/**/*", "!./target/keep/**/*"}, "target", false},
		{"double-slash negation protects after normalize", []string{"target/**/*", "!target//keep/**/*"}, "target", false},
		{"uppercase pattern does not prune lowercase dir", []string{"**/Target/**/*"}, "target", false},
		{"case-exact pattern prunes", []string{"**/Target/**/*"}, "Target", true},
		{"brace-extension filter does not cover subtree", []string{"target/**/*.{ts,tsx}"}, "target", false},
		{"brace-name dir cover prunes", []string{"**/{target,dist}/**/*"}, "target", true},
	}
	for _, c := range cases {
		parsed := ParseIgnorePatterns(c.ignores)
		if got := canPruneDir(c.dir, parsed, buildNegReach(parsed)); got != c.want {
			t.Errorf("%s: canPruneDir(%q, %v) = %v, want %v", c.name, c.dir, c.ignores, got, c.want)
		}
	}
}

func TestGlobalIgnoreMatcherReopensMatchingDirectoryNode(t *testing.T) {
	for _, test := range []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "bare directory node", pattern: "!ignored", want: true},
		{name: "rooted directory does not match relative node", pattern: "!/ignored", want: false},
		{name: "directory-only node", pattern: "!ignored/", want: true},
		{name: "directory and contents", pattern: "!ignored/**", want: true},
		{name: "directory files", pattern: "!ignored/**/*", want: false},
		{name: "one file", pattern: "!ignored/keep.ts", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			matcher := NewGlobalIgnoreMatcher(
				RslintConfig{{Ignores: []string{test.pattern}}},
				"/repo",
				nil,
			)
			assert.Equal(t, matcher.ReopensDirectoryNode("/repo/ignored", ""), test.want)
		})
	}

	for _, test := range []struct {
		name     string
		patterns []string
		want     bool
	}{
		{name: "single-level descendant", patterns: []string{"!dir/*"}, want: true},
		{name: "recursive descendant", patterns: []string{"!dir/**/*"}, want: true},
		{name: "exact descendant", patterns: []string{"!dir/child"}, want: true},
		{name: "later positive wins", patterns: []string{"!dir/child/", "dir/*"}, want: false},
		{name: "later negation wins", patterns: []string{"dir/*", "!dir/child/"}, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			matcher := NewGlobalIgnoreMatcher(
				RslintConfig{{Ignores: test.patterns}},
				"/repo",
				nil,
			)
			assert.Equal(t, matcher.ReopensDirectoryNode("/repo/dir/child", ""), test.want)
		})
	}
}

// --- Unit: ParseIgnorePattern remaining switch-arm inputs ---
// TestParseIgnorePattern_Classification already pins braces/`**`/`./`/`//`. This
// fills the glob-metachar and rooted/trailing forms the task flags: `?`,
// `[abc]`, rooted `/foo`, trailing `foo/`, bare `**`. All land in
// dirAbsoluteBlock (none ends `/**/*` nor `/*`-not-`/**`), so a future switch
// change that special-cased any of them would flip a Kind here. Values verified
// by running ParseIgnorePattern directly.
func TestParseIgnorePattern_GlobMetacharAndRootedArms(t *testing.T) {
	cases := []struct {
		raw     string
		glob    string
		negated bool
		kind    dirKind
	}{
		{"?", "?", false, dirAbsoluteBlock},                  // single-char wildcard, no dir suffix
		{"[abc]", "[abc]", false, dirAbsoluteBlock},          // bracket class, no dir suffix
		{"**", "**", false, dirAbsoluteBlock},                // bare doublestar
		{"/foo", "/foo", false, dirAbsoluteBlock},            // rooted: leading "/" survives NormalizePath
		{"foo/", "foo/", false, dirAbsoluteBlock},            // trailing slash: not "/*" nor "/**/*"
		{"src/?.ts", "src/?.ts", false, dirAbsoluteBlock},    // `?` in name, file-ish, still default arm
		{"a[bc]d", "a[bc]d", false, dirAbsoluteBlock},        // bracket mid-segment
		{"!?", "?", true, dirAbsoluteBlock},                  // negated single-char wildcard
		{"![abc]/x", "[abc]/x", true, dirAbsoluteBlock},      // negated bracket dir (no /**/* suffix)
		{"!/foo/**/*", "/foo/**/*", true, dirFileLevelCover}, // rooted negated file-level cover
	}
	for _, c := range cases {
		p := ParseIgnorePattern(c.raw)
		if p.Glob != c.glob || p.Negated != c.negated || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Glob:%q Negated:%v Kind:%d}, want {Glob:%q Negated:%v Kind:%d}",
				c.raw, p.Glob, p.Negated, p.Kind, c.glob, c.negated, c.kind)
		}
	}
}

// --- Unit: isDirAbsolutelyBlocked ancestor-segment scan ---
// The predicate matches dirPath itself AND every ancestor prefix (probing both
// `partial` and `partial+"/x"`). This is the "directory blocking is absolute"
// machinery shared by GetConfigForFile and canPruneDir. The existing coverage
// reaches it only indirectly (single-segment dirs in the canPruneDir table, or
// via the walk); pin the multi-segment ancestor scan directly so a regression
// in the segment loop (e.g. off-by-one on `j`, or dropping the `/x` probe)
// can't hide behind file-level agreement. Values verified by running
// isDirAbsolutelyBlocked directly.
func TestIsDirAbsolutelyBlocked_AncestorScan(t *testing.T) {
	cases := []struct {
		name    string
		dir     string
		ignores []string
		want    bool
	}{
		// pattern matches a MIDDLE ancestor segment of a/b/c.
		{"**/b blocks via ancestor a/b", "a/b/c", []string{"**/b"}, true},
		{"**/b/** blocks via ancestor a/b", "a/b/c", []string{"**/b/**"}, true},
		{"a/** blocks via ancestor a", "a/b/c", []string{"a/**"}, true},
		{"bare ancestor name blocks descendant", "foo/bar/baz", []string{"foo"}, true},
		{"bare name blocks self", "foo", []string{"foo"}, true},
		// negative: no ancestor segment matches.
		{"**/b does not block x/y/z", "x/y/z", []string{"**/b"}, false},
		{"sibling-name ancestor not matched", "target/sub", []string{"targetx"}, false},
		// the `dir+"/x"` probe is what lets a bare-dir pattern match a leaf dir:
		// `build` matches "build" directly; `**/dist` matches "a/dist" leaf.
		{"**/dist blocks leaf dist via /x probe", "a/dist", []string{"**/dist"}, true},
		// file-level and negation patterns are excluded by Kind — never block.
		{"file-level cover never dir-blocks", "target/sub", []string{"target/**/*"}, false},
		{"single-level X/* never dir-blocks", "dist/sub", []string{"dist/*"}, false},
		{"negated absolute never dir-blocks", "build/sub", []string{"!build/**"}, false},
	}
	for _, c := range cases {
		parsed := ParseIgnorePatterns(c.ignores)
		if got := isDirAbsolutelyBlocked(c.dir, parsed); got != c.want {
			t.Errorf("%s: isDirAbsolutelyBlocked(%q, %v) = %v, want %v", c.name, c.dir, c.ignores, got, c.want)
		}
	}
}

// --- Unit: segPrefixEither / negReach.overlaps segment boundary ---
// overlaps drives the negation-aware half of canPruneDir. Its correctness hinges
// on segPrefixEither being SEGMENT-anchored (bidirectional): a negation target
// inside the dir, OR the dir inside the negation's covered range, but never a
// mere string-prefix collision (`foo` vs `foobar`). The canPruneDir table hits
// the happy paths; pin the boundary directly — especially the false cases that a
// naive strings.HasPrefix would wrongly report true.
func TestSegPrefixEither_Boundary(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"target", "target", true},        // equal
		{"target", "target/keep", true},   // a is segment-prefix of b
		{"tests/e2e", "tests", true},      // b is segment-prefix of a (dir inside negation range)
		{"tests", "tests/e2e/spec", true}, // a is deep segment-prefix of b
		{"target", "targetX", false},      // string-prefix but NOT segment-prefix
		{"foo", "foobar", false},          // classic non-overlap
		{"foobar", "foo", false},          // reversed non-overlap
		{"a/b", "a/bc", false},            // segment boundary inside second segment
		{"a/b", "a/b/c", true},            // nested
		{"", "anything", false},           // empty is NOT a segment-prefix of a non-empty path (no leading "/")
		{"", "", true},                    // empty equals empty
	}
	for _, c := range cases {
		if got := segPrefixEither(c.a, c.b); got != c.want {
			t.Errorf("segPrefixEither(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// overlaps composes segPrefixEither over a built negReach. Pin the boundary at
// the negReach level too: a sibling-name negation (`!targetX/...`) must NOT
// protect `target`, and an unrooted negation always overlaps. This is the exact
// soundness lever canPruneDir relies on.
func TestNegReach_OverlapsBoundary(t *testing.T) {
	cases := []struct {
		name     string
		negation string
		dir      string
		want     bool
	}{
		{"exact target protects target", "!target/keep/**/*", "target", true},
		{"sibling targetX does not protect target", "!targetX/keep/**/*", "target", false},
		{"foo negation does not overlap foobar", "!foo/**/*", "foobar", false},
		{"deep negation protects ancestor dir", "!tests/e2e/**/*", "tests", true},
		{"deep negation protects exact dir", "!tests/e2e/**/*", "tests/e2e", true},
		{"unrooted negation overlaps anything", "!**/keep/**/*", "whatever/deep", true},
		{"leading-wildcard negation is unrooted", "!*.log", "anything", true},
	}
	for _, c := range cases {
		parsed := ParseIgnorePatterns([]string{c.negation})
		neg := buildNegReach(parsed)
		if got := neg.overlaps(c.dir); got != c.want {
			t.Errorf("%s: overlaps(%q) under %q = %v, want %v", c.name, c.dir, c.negation, got, c.want)
		}
	}
}

// A case-insensitive filesystem changes matching, not whether an anchored
// negation has an anchored prefix. Preserve that prefix and fold only the
// overlap comparison; otherwise one unrelated negation disables every
// file-level directory prune on macOS/Windows.
func TestNegReach_CaseInsensitiveAnchoredPrefix(t *testing.T) {
	parsed := ParseIgnorePatterns([]string{
		"**/TARGET/**/*",
		"!Scripts/**/Debug",
	})
	for i := range parsed {
		parsed[i].CaseInsensitive = true
	}

	neg := buildNegReach(parsed)
	assert.Equal(t, len(neg.prefixes), 1)
	assert.Equal(t, neg.prefixes[0].unrooted, false)
	assert.Equal(t, neg.prefixes[0].literal, "Scripts")
	assert.Equal(t, neg.prefixes[0].caseInsensitive, true)
	assert.Assert(t, neg.overlaps("scripts"), "case-folded exact prefix must overlap")
	assert.Assert(t, neg.overlaps("SCRIPTS/pkg"), "case-folded descendant must overlap")
	assert.Assert(t, !neg.overlaps("target"), "unrelated target must remain prunable")
	assert.Assert(t, canPruneDir("target", parsed, neg), "unrelated ignored target must be pruned")
}

// --- Unit: buildNegReach with glob-metachar / double-slash negations ---
// TestBuildNegReach covers brace, dotslash, and mid-pattern metachars. Fill the
// `?`, `[abc]`, and double-slash forms: literalSegmentPrefix must stop at the
// first metachar (so `?`/`[` in the FIRST segment → unrooted; in a LATER segment
// → literal up to the segment before it), and normalize must collapse `//`
// before the literal is extracted. Values verified by running buildNegReach
// directly.
func TestBuildNegReach_MetacharAndDoubleSlash(t *testing.T) {
	got := buildNegReach(ParseIgnorePatterns([]string{
		"!src/a?b/keep/**/*",   // `?` in 2nd segment → literal "src"
		"!src/[abc]/keep/**/*", // bracket in 2nd segment → literal "src"
		"!a?b/keep/**/*",       // `?` in 1st segment → no concrete prefix → unrooted
		"![abc]/keep/**/*",     // bracket in 1st segment → unrooted
		"!a/b//c/**/*",         // double-slash collapses → literal "a/b/c"
		"!deep/path/no/meta",   // no metachar at all → whole path is the literal
	})).prefixes
	want := []negPrefix{
		{literal: "src"},
		{literal: "src"},
		{unrooted: true},
		{unrooted: true},
		{literal: "a/b/c"},
		{literal: "deep/path/no/meta"},
	}
	assert.Equal(t, len(got), len(want), "got %+v", got)
	for i := range want {
		assert.Equal(t, got[i], want[i], "entry %d", i)
	}
}
