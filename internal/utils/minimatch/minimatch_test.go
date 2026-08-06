package minimatch_test

import (
	"testing"

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
		{"/src/[abc/x", "/src/[abc/x", true},
		{"/src/[abc/x", "/src/a/x", false},

		// A class naming a character a regexp reads as syntax of its own is
		// still a class, and names that character.
		{"/src/[[]", "/src/[", true},
		{"/src/[[]", "/src/a", false},
		{`/src/[\z]`, "/src/z", true},
		{`/src/[\z]`, "/src/a", false},
		{"/src/[!-[]", "/src/a", true},
		{"/src/[!-[]", "/src/[", false},

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
