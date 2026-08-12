package minimatch_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/utils/minimatch"
)

// TestMatch pins this port against minimatch 3.1.5 itself: every expectation
// below is what `minimatch(path, pattern)` answers on that version.
func TestMatch(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// ---- extended glob: zero or one ----
		{"/src/?(server)/*", "/src/server/a.ts", true},
		{"/src/?(server)/*", "/src/a.ts", false},
		{"/src/?(server)/*", "/src/x(server)/a.ts", false},
		{"/src/?(server)/*", "/src/?(server)/a.ts", false},

		// ---- extended glob: exactly one of ----
		{"/src/@(server|shared)/**/*", "/src/server/a/b.ts", true},
		{"/src/@(server|shared)/**/*", "/src/shared/a.ts", true},
		{"/src/@(server|shared)/**/*", "/src/client/a.ts", false},
		{"/src/@(server|shared)/**/*", "/src/@(server|shared)/a.ts", false},

		// ---- extended glob: one or more, zero or more ----
		{"/src/+(server)/*", "/src/server/a.ts", true},
		{"/src/+(server)/*", "/src/serverserver/a.ts", true},
		{"/src/+(server)/*", "/src/a.ts", false},
		{"/src/*(server)/*", "/src/server/a.ts", true},
		{"/src/*(server)/*", "/src/serverserver/a.ts", true},
		{"/src/*(server)/*", "/src/other/a.ts", false},

		// ---- extended glob: negation ----
		{"/src/!(server)/**/*", "/src/client/a.ts", true},
		{"/src/!(server)/**/*", "/src/server/a.ts", false},
		{"/src/!(server)/**/*", "/src/client/deep/a.ts", true},
		{"/src/!(server|shared)/*", "/src/client/a.ts", true},
		{"/src/!(server|shared)/*", "/src/server/a.ts", false},
		{"/src/!(server|shared)/*", "/src/shared/a.ts", false},
		{"*.!(js)", "a.ts", true},
		{"*.!(js)", "a.js", false},
		{"*.!(js)", "a.json", true},

		// The lookahead of a leading negated list has to reach the end of the
		// pattern, or `a.xyz.yz` would slip past `*.!(x).!(y|z)`.
		{"*.!(x).!(y|z)", "a.xyz.yz", true},
		{"*.!(x).!(y|z)", "a.b.c", true},
		{"*.!(x).!(y|z)", "a.x.y", false},

		// A negated list nested in a list that never closes still compiles.
		{"/src/!(a!(a|b)", "/src/x", false},
		{"/src/!(a!(a|b)", "/src/!(ax", true},
		{"/src/!(a*(*.js|!(*.json))", "/src/x", false},

		// ---- extended glob: combined and nested lists ----
		{"/src/@(a|b)/+(c|d)/*", "/src/a/c/x.ts", true},
		{"/src/@(a|b)/+(c|d)/*", "/src/b/dd/x.ts", true},
		{"/src/@(a|b)/+(c|d)/*", "/src/c/c/x.ts", false},
		{"/src/*(*.js|!(*.json))", "/src/a.js", true},
		{"/src/*(*.js|!(*.json))", "/src/a.json", true},
		{"/src/*(*.js|!(*.json))", "/src/a.ts", true},

		// A list left open at the end of a part matches itself.
		{"/src/+(a|b", "/src/+(a|b", true},
		{"/src/+(a|b", "/src/a", false},

		// ---- wildcards and dot names ----
		{"/src/**/*", "/src/a.ts", true},
		{"/src/**/*", "/src/a/b.ts", true},
		{"/src/**/*", "/src/.hidden.ts", false},
		{"/src/**/*", "/src/.d/b.ts", false},
		{"/src/*", "/src/a.ts", true},
		{"/src/*", "/src/.hidden.ts", false},
		{"/src/*", "/src/a/b.ts", false},
		{"/src/a?c.ts", "/src/abc.ts", true},
		{"/src/a?c.ts", "/src/ac.ts", false},
		{"/src/**", "/src/a.ts", true},
		{"/src/**", "/src/a/b/c.ts", true},
		{"/src/**", "/src", false},
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/b", true},
		{"a/**/b", "a/x/y/b", true},

		// A trailing separator leaves an empty final part, which `*` covers but
		// only as the very last one.
		{"a/*", "a/b/", true},
		{"a/*", "a/b", true},
		{"a/*", "a/", false},

		// ---- brace expansion ----
		{"/src/{a,b}/*.ts", "/src/a/x.ts", true},
		{"/src/{a,b}/*.ts", "/src/b/x.ts", true},
		{"/src/{a,b}/*.ts", "/src/c/x.ts", false},
		{"/src/{0..3}.ts", "/src/0.ts", true},
		{"/src/{0..3}.ts", "/src/2.ts", true},
		{"/src/{0..3}.ts", "/src/4.ts", false},
		{"/src/{a,b/c}/d", "/src/a/d", true},
		{"/src/{a,b/c}/d", "/src/b/c/d", true},
		{"/src/{a,b/c}/d", "/src/b/d", false},

		// ---- character classes ----
		{"/src/[abc].ts", "/src/a.ts", true},
		{"/src/[abc].ts", "/src/d.ts", false},
		{"/src/[!abc].ts", "/src/a.ts", false},
		{"/src/[!abc].ts", "/src/d.ts", true},

		// An invalid or unterminated class matches itself.
		{"/src/[z-a]/x", "/src/[z-a]/x", true},
		{"/src/[z-a]/x", "/src/a/x", false},
		{"[!-\n]", "[!-\n]", true},
		{"[!-\n]", "a", false},
		{"[]^?]", "[]^?]", true},
		{"[]^?]", "]", false},
		{"/src/[abc/x", "/src/[abc/x", true},
		{"/src/[abc/x", "/src/a/x", false},

		// A class naming nothing matches any one character.
		{"/src/[!]", "/src/a", true},
		{"/src/[!]", "/src/]", true},
		{"/src/[!]", "/src/.", false},
		{"/src/[!]", "/src/ab", false},
		{"/src/[^]", "/src/a", true},
		{"/src/[^]", "/src/.", false},
		{"/src/x[!]y", "/src/x.y", true},
		{"/src/[!]]", "/src/a]", true},
		{"/src/[!]]", "/src/]", false},

		// A class naming a character a regexp reads as syntax of its own is
		// still a class, and names that character.
		{"/src/[[]", "/src/[", true},
		{"/src/[[]", "/src/a", false},
		{`/src/[\z]`, "/src/z", true},
		{`/src/[\z]`, "/src/a", false},
		{"/src/[!-[]", "/src/a", true},
		{"/src/[!-[]", "/src/[", false},

		// ---- line terminators ----
		// The guard that keeps a wildcard from matching an empty name is a
		// JavaScript `.`, which no line terminator matches, and the end of a
		// name is the end of it rather than the newline that ends it.
		{"?", "\r", false},
		{"?", "\n", false},
		{"?", "\u2028", false},
		{"?", "\u2029", false},
		{"*", "a\n", true},
		{"a*", "a\r", true},
		{"a", "a\n", false},
		{"@(a)", "a\n", false},
		{"*$", "a$", true},
		{"*$", "a$\n", false},
		{"@($|a)", "$", true},

		// The scan that decides a pattern is worth expanding stops at a line
		// terminator too, so this brace set is never expanded at all.
		{"{a\rb,c}", "c", false},
		{"{a\rb,c}", "{a\rb,c}", true},
		{"{a,c}", "c", true},

		// ---- surrounding whitespace ----
		// A pattern is trimmed the way JavaScript trims a string, which reads
		// U+FEFF as blank and U+0085 as a character of its own.
		{"\ufeffa", "a", true},
		{"a\ufeff", "a", true},
		{"\u00a0a", "a", true},
		{"\va", "a", true},
		{"\u0085a", "a", false},
		{"\u0085a", "\u0085a", true},

		// ---- characters outside ASCII ----
		{"/src/caf*", "/src/café", true},
		{"/src/*é", "/src/café", true},
		{"/src/?é", "/src/aé", true},
		{"/src/日本*", "/src/日本語", true},
		{"/src/*(é|a)", "/src/é", true},
		{"/src/[é]", "/src/é", true},
		{"/src/[é]", "/src/a", false},

		// ---- escaping and negated patterns ----
		// A `\` escapes on every platform: it is never a path separator, so
		// these answer the same wherever the test runs.
		{`/src/\*.ts`, "/src/*.ts", true},
		{`/src/\*.ts`, "/src/a.ts", false},
		{`a\*b/*`, "a*b/ooo", true},
		{`[\\]`, `\`, true},
		{"!/src/**", "/src/a.ts", false},
		{"!/src/**", "/other/a.ts", true},

		// ---- a Windows drive path is nothing special to the matcher ----
		{"C:/repo/server/**/*", "C:/repo/server/a.ts", true},
		{"C:/repo/server/**/*", "C:/repo/client/a.ts", false},
	}

	for _, test := range tests {
		t.Run(test.pattern+" vs "+test.path, func(t *testing.T) {
			got := minimatch.Match(test.pattern, test.path, minimatch.Options{})
			if got != test.want {
				t.Errorf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.want)
			}
		})
	}
}

// TestMatchOptions covers the options this port carries over from minimatch.
func TestMatchOptions(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		options minimatch.Options
		want    bool
	}{
		{
			name:    "Dot lets a wildcard reach a dot name",
			pattern: "/src/*",
			path:    "/src/.hidden.ts",
			options: minimatch.Options{Dot: true},
			want:    true,
		},
		{
			name:    "NoExt leaves a list to match literally",
			pattern: "/src/?(server)/*",
			path:    "/src/server/a.ts",
			options: minimatch.Options{NoExt: true},
			want:    false,
		},
		{
			name:    "NoBrace leaves a brace set to match literally",
			pattern: "/src/{a,b}/x.ts",
			path:    "/src/a/x.ts",
			options: minimatch.Options{NoBrace: true},
			want:    false,
		},
		{
			name:    "NoCase ignores case",
			pattern: "/src/Server/*.ts",
			path:    "/src/server/a.ts",
			options: minimatch.Options{NoCase: true},
			want:    true,
		},
		{
			name:    "NoGlobStar keeps `**` inside one part",
			pattern: "/src/**/a.ts",
			path:    "/src/one/two/a.ts",
			options: minimatch.Options{NoGlobStar: true},
			want:    false,
		},
		{
			name:    "NoNegate keeps a leading `!` literal",
			pattern: "!/src/a.ts",
			path:    "/src/a.ts",
			options: minimatch.Options{NoNegate: true},
			want:    false,
		},
		{
			name:    "MatchBase matches a slashless pattern against the basename",
			pattern: "a?c.ts",
			path:    "/src/deep/abc.ts",
			options: minimatch.Options{MatchBase: true},
			want:    true,
		},
		{
			name:    "FlipNegate reports the hit of a negated pattern",
			pattern: "!/src/**",
			path:    "/src/a.ts",
			options: minimatch.Options{FlipNegate: true},
			want:    true,
		},
		{
			name:    "a comment matches nothing",
			pattern: "#/src/a.ts",
			path:    "#/src/a.ts",
			options: minimatch.Options{},
			want:    false,
		},
		{
			name:    "NoComment keeps a leading `#` literal",
			pattern: "#/src/a.ts",
			path:    "#/src/a.ts",
			options: minimatch.Options{NoComment: true},
			want:    true,
		},
		{
			name:    "an empty pattern matches only the empty path",
			pattern: "",
			path:    "",
			options: minimatch.Options{},
			want:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := minimatch.Match(test.pattern, test.path, test.options)
			if got != test.want {
				t.Errorf("Match(%q, %q, %+v) = %v, want %v", test.pattern, test.path, test.options, got, test.want)
			}
		})
	}
}

// TestMatchNoCase pins NoCase against the `i` flag of a JavaScript regexp
// written without the `u` flag, which minimatch hands its compiled pattern.
// JavaScript compares two characters by their uppercase, with the one rule
// that a character outside ASCII never uppercases into it: U+03A3 Σ, U+03C3 σ
// and U+03C2 ς are one character to it, while U+212A K and `k` are two, and
// neither answer is the one Unicode case folding gives.
func TestMatchNoCase(t *testing.T) {
	type testCase struct {
		pattern string
		path    string
		want    bool
	}
	tests := []testCase{
		{"/src/Server/*.ts", "/src/server/a.ts", true},
		{"caf\u00e9", "CAF\u00c9", true},
		{"@(a|b)", "B", true},

		// a capital sigma, a small sigma and a final sigma are one character
		{"\u03a3", "\u03c2", true},
		{"\u03a3", "\u03c3", true},
		{"\u03c2", "\u03a3", true},
		{"[\u03a3]", "\u03c2", true},

		// a kelvin sign is not a `k`, and a long s is not an `s`
		{"\u212a", "k", false},
		{"k", "\u212a", false},
		{"K", "k", true},
		{"[a-z]", "\u212a", false},
		{"[a-z]", "\u017f", false},
		{"[a-z]", "K", true},
		{"[a-z]", "s", true},

		// a capital eszett does not uppercase onto the small one
		{"\u00df", "\u1e9e", false},

		// a regexp without `u` does not fold a supplementary-plane letter
		{"\U00010428", "\U00010400", false},
		{"\U00010400", "\U00010428", false},

		// a negated class turns down whatever the widened class covers
		{"[!a]", "A", false},
		{"[!a-z]", "K", false},

		// a `-` that ends a class stands for itself, widened class or not
		{"[a-]", "a", true},
		{"[a-]", "A", true},
		{"[a-]", "-", true},
	}

	// These mappings were added in Unicode 16 and 17 after the Unicode 15
	// tables in Go 1.26. Node 26 uses Unicode 17 for regexp canonicalization.
	unicode17Pairs := [][2]rune{
		{0x019B, 0xA7DC},
		{0x0264, 0xA7CB},
		{0x1C8A, 0x1C89},
		{0xA7CD, 0xA7CC},
		{0xA7CF, 0xA7CE},
		{0xA7D3, 0xA7D2},
		{0xA7D5, 0xA7D4},
		{0xA7DB, 0xA7DA},
	}
	for _, pair := range unicode17Pairs {
		tests = append(tests,
			testCase{pattern: string(pair[0]), path: string(pair[1]), want: true},
			testCase{pattern: string(pair[1]), path: string(pair[0]), want: true},
			testCase{pattern: "[" + string(pair[0]) + "]", path: string(pair[1]), want: true},
			testCase{pattern: "[" + string(pair[1]) + "]", path: string(pair[0]), want: true},
		)
	}

	// Full uppercase expands each of these into multiple UTF-16 code units,
	// so ECMAScript keeps it distinct from Go's simple uppercase character.
	multiUnitUppercase := [][2]rune{{0x1FB3, 0x1FBC}, {0x1FC3, 0x1FCC}, {0x1FF3, 0x1FFC}}
	for _, bounds := range [][2]rune{{0x1F80, 0x1F87}, {0x1F90, 0x1F97}, {0x1FA0, 0x1FA7}} {
		for r := bounds[0]; r <= bounds[1]; r++ {
			multiUnitUppercase = append(multiUnitUppercase, [2]rune{r, r + 8})
		}
	}
	for _, pair := range multiUnitUppercase {
		tests = append(tests,
			testCase{pattern: string(pair[0]), path: string(pair[1]), want: false},
			testCase{pattern: string(pair[1]), path: string(pair[0]), want: false},
			testCase{pattern: "[" + string(pair[0]) + "]", path: string(pair[1]), want: false},
			testCase{pattern: "[" + string(pair[1]) + "]", path: string(pair[0]), want: false},
		)
	}

	for _, test := range tests {
		t.Run(test.pattern+" vs "+test.path, func(t *testing.T) {
			got := minimatch.Match(test.pattern, test.path, minimatch.Options{NoCase: true})
			if got != test.want {
				t.Errorf("Match(%q, %q, NoCase) = %v, want %v", test.pattern, test.path, got, test.want)
			}
		})
	}
}

// TestMatchOverLongPattern covers the length minimatch refuses a pattern past,
// which it measures before it reads anything else about the pattern. Nothing
// the pattern says is honored after that, a leading `!` included, so a pattern
// this long matches nothing rather than everything.
//
// The length is the one String.prototype.length reports, in UTF-16 code units,
// so a pattern written in characters that take three bytes to spell fits three
// times what its size in memory would allow.
func TestMatchOverLongPattern(t *testing.T) {
	longest := strings.Repeat("a/", 32767) + "a"
	tooLong := strings.Repeat("a/", 32768) + "a"
	wideLongest := strings.Repeat("界", 65536)
	wideTooLong := strings.Repeat("界", 65537)

	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{name: "the longest pattern still compiles", pattern: longest, path: longest, want: true},
		{name: "negating the longest pattern still compiles", pattern: "!" + longest, path: "x", want: true},
		{name: "one code unit past matches nothing", pattern: tooLong, path: tooLong, want: false},
		{name: "negating one code unit past matches nothing either", pattern: "!" + tooLong, path: "x", want: false},
		{name: "the longest pattern of wide characters still compiles", pattern: wideLongest, path: wideLongest, want: true},
		{name: "one wide character past matches nothing", pattern: wideTooLong, path: wideTooLong, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := minimatch.Match(test.pattern, test.path, minimatch.Options{})
			if got != test.want {
				t.Errorf("Match(<%d bytes>, <%d bytes>) = %v, want %v", len(test.pattern), len(test.path), got, test.want)
			}
		})
	}
}

// TestMatchManyGlobstars covers a pattern carrying more `**` than any rule
// would write. Dividing the path between them is a search, and one that misses
// explores every division, so the answer has to come from a walk that
// remembers where it has already failed rather than from raw backtracking.
func TestMatchManyGlobstars(t *testing.T) {
	parts := make([]string, 0, 32)
	for i := range 32 {
		parts = append(parts, string(rune('a'+i%26))+strconv.Itoa(i))
	}
	path := strings.Join(parts, "/")

	globstars := strings.TrimSuffix(strings.Repeat("**/", 16), "/")

	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "a miss on the last part", pattern: globstars + "/zz", want: false},
		{name: "a hit on the last part", pattern: globstars + "/" + parts[len(parts)-1], want: true},
		{name: "a miss in the middle", pattern: globstars + "/zz/**", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := time.Now()
			got := minimatch.Match(test.pattern, path, minimatch.Options{})
			if got != test.want {
				t.Errorf("Match(%q, %q) = %v, want %v", test.pattern, path, got, test.want)
			}
			if elapsed := time.Since(start); elapsed > 5*time.Second {
				t.Errorf("Match(%q, %q) took %s, want a walk that does not backtrack exponentially", test.pattern, path, elapsed)
			}
		})
	}
}

// TestBraceExpand pins the brace expansion against brace-expansion 1.1.16,
// the version minimatch 3.1.5 resolves its `^1.1.7` dependency to.
func TestBraceExpand(t *testing.T) {
	tests := []struct {
		pattern string
		want    []string
	}{
		{"a{b,c}d", []string{"abd", "acd"}},
		{"a{b,}c", []string{"abc", "ac"}},
		{"a{0..3}d", []string{"a0d", "a1d", "a2d", "a3d"}},
		{"a{b,c{d,e}f}g", []string{"abg", "acdfg", "acefg"}},
		{"a{b,c}d{e,f}g", []string{"abdeg", "abdfg", "acdeg", "acdfg"}},
		{"a{01..3}d", []string{"a01d", "a02d", "a03d"}},
		{"a{3..1}d", []string{"a3d", "a2d", "a1d"}},
		{"a{0..6..2}d", []string{"a0d", "a2d", "a4d", "a6d"}},
		{"a{a..c}d", []string{"aad", "abd", "acd"}},
		{"a{-3..-1}d", []string{"a-3d", "a-2d", "a-1d"}},
		{"a{-03..-1}d", []string{"a-03d", "a-02d", "a-01d"}},
		// A step of zero counts as one.
		{"a{1..3..0}d", []string{"a1d", "a2d", "a3d"}},
		// An endpoint is the number JavaScript reads it as, which stops
		// counting in whole numbers past 2^53 but never wraps around: a step
		// that would carry the sequence past its end point ends it instead.
		{"a{2147483647..2147483648}d", []string{"a2147483647d", "a2147483648d"}},
		{"a{-2147483648..-2147483649}d", []string{"a-2147483648d", "a-2147483649d"}},
		{"{9223372036854770000..9223372036854775807..4096}", []string{"9223372036854770000", "9223372036854774000"}},
		{"{-9223372036854770000..-9223372036854775807..4096}", []string{"-9223372036854770000", "-9223372036854774000"}},
		// The `{a},b}` rewrite starts over at the top level, where an
		// expansion that reduces to nothing drops out.
		{"{{}},}", []string{"{}}"}},
		{"{{}},,a}", []string{"{}}", "a"}},
		// The `}` that rewrite looks for has to be on the comma's own line.
		{"a{b}c,\n}", []string{"a{b}c,\n}"}},
		// Invalid sets are left alone.
		{"a{2..}b", []string{"a{2..}b"}},
		{"a{b}c", []string{"a{b}c"}},
		{"a{b,c}", []string{"ab", "ac"}},
		{`a\{b,c}d`, []string{"a{b,c}d"}},
		{"x{{a,b}}y", []string{"x{a}y", "x{b}y"}},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			got := minimatch.BraceExpand(test.pattern)
			if len(got) != len(test.want) {
				t.Fatalf("BraceExpand(%q) = %q, want %q", test.pattern, got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("BraceExpand(%q) = %q, want %q", test.pattern, got, test.want)
				}
			}
		})
	}
}
