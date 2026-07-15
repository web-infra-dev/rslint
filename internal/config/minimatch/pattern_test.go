package minimatch

// cspell:ignore dotdot xabcd xacd xyawq xyza xyzwq

import (
	"errors"
	"strings"
	"testing"
	"unicode"
)

func TestSearchPatternMinimatchSurface(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		match   bool
	}{
		{name: "doublestar exact", pattern: "src/**/*.js", path: "src/deep/a.js", match: true},
		{name: "extension mismatch", pattern: "src/**/*.js", path: "src/deep/a.ts", match: false},
		{name: "brace", pattern: "src/{a,b}.js", path: "src/b.js", match: true},
		{name: "extglob", pattern: "src/@(a|b).js", path: "src/a.js", match: true},
		{name: "posix class", pattern: "file[[:digit:]].js", path: "file7.js", match: true},
		{name: "repeated slash", pattern: "src//utils/*.js", path: "src/utils/a.js", match: true},
		{name: "level one dotdot", pattern: "src/../lib/*.js", path: "lib/a.js", match: true},
		{name: "internal dot retained", pattern: "a/./b.js", path: "a/b.js", match: false},
		{name: "one leading dot normalized at search boundary", pattern: "./src/**", path: "src/a.js", match: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pattern, err := CompileSearchPattern(test.pattern, ".")
			if err != nil {
				t.Fatal(err)
			}
			if got := pattern.Match(test.path); got != test.match {
				t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
			}
		})
	}
}

func TestSearchPatternNegationAndPartialMatch(t *testing.T) {
	negated, err := CompileSearchPattern("!src/**", ".")
	if err != nil {
		t.Fatal(err)
	}
	if negated.Match("src/a.js") || !negated.Match("lib/a.js") {
		t.Fatal("leading negation did not invert the exact matcher")
	}

	pattern, err := CompileSearchPattern("src/**/*.js", ".")
	if err != nil {
		t.Fatal(err)
	}
	if !pattern.PartialMatch("src/deep") || pattern.PartialMatch("lib") {
		t.Fatal("partial matching did not preserve viable directory prefixes")
	}
}

func TestCompileRelativePatternPreservesLeadingDotSegment(t *testing.T) {
	pattern, err := CompileRelativePattern("./src/**")
	if err != nil {
		t.Fatal(err)
	}
	if pattern.Match("src/a.js") {
		t.Fatal("CompileRelativePattern must not remove ConfigArray-observable ./")
	}
}

func TestBraceExpansionUsesMinimatchLimit(t *testing.T) {
	pattern, err := CompileRelativePattern("file{1..1025}.js")
	if err != nil {
		t.Fatalf("1025 alternatives are below Minimatch's expansion limit: %v", err)
	}
	if !pattern.Match("file1025.js") {
		t.Fatal("numeric brace range did not include its upper endpoint")
	}
}

func TestBraceExpansionMatchesBraceExpansionFiveEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		match   bool
	}{
		{name: "dollar brace is literal", pattern: "${a,b}.js", path: "$a.js", match: false},
		{name: "dollar brace literal spelling", pattern: "${a,b}.js", path: "${a,b}.js", match: true},
		{name: "prefixed dollar brace is literal", pattern: "x${a,b}.js", path: "x$a.js", match: false},
		{name: "double dollar still protects brace", pattern: "$${a,b}.js", path: "$${a,b}.js", match: true},
		{name: "postfix brace still expands", pattern: "${a,b}{c,d}.js", path: "${a,b}d.js", match: true},
		{name: "numeric zero step defaults to one", pattern: "{1..3..0}.js", path: "2.js", match: true},
		{name: "descending numeric zero step", pattern: "{3..1..0}.js", path: "2.js", match: true},
		{name: "alpha zero step defaults to one", pattern: "{a..c..0}.js", path: "b.js", match: true},
		{name: "descending alpha zero step", pattern: "{c..a..0}.js", path: "b.js", match: true},
		{name: "unicode range stays literal", pattern: "{α..γ}.js", path: "β.js", match: false},
		{name: "unicode range literal spelling", pattern: "{α..γ}.js", path: "{α..γ}.js", match: true},
		{name: "punctuation range stays literal", pattern: "{!..~}.js", path: "A.js", match: false},
		{name: "alpha range backslash becomes empty", pattern: "{A..z}*.js", path: "1.js", match: true},
		{name: "braces expand before character classes", pattern: "[{a,b}].js", path: "a.js", match: true},
		{name: "expanded class excludes brace punctuation", pattern: "[{a,b}].js", path: "{.js", match: false},
		{name: "malformed close bash recovery first", pattern: "a{},b}c", path: "a}c", match: true},
		{name: "malformed close bash recovery second", pattern: "a{},b}c", path: "abc", match: true},
		{name: "malformed close does not stay literal", pattern: "a{},b}c", path: "a{},b}c", match: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pattern, err := CompileRelativePattern(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := pattern.Match(test.path); got != test.match {
				t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
			}
		})
	}

	expanded, err := expandSearchBraces("{1..100001}.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(expanded) != maxBraceExpansions || expanded[len(expanded)-1] != "100000.js" {
		t.Fatalf("large range = len %d, last %q; want truncated at 100000.js", len(expanded), expanded[len(expanded)-1])
	}
}

func TestEmptyExtglobMatchesMinimatchTen(t *testing.T) {
	for _, operator := range []string{"@", "?", "+", "*"} {
		t.Run("standalone "+operator, func(t *testing.T) {
			pattern, err := CompileRelativePattern(operator + "()")
			if err != nil {
				t.Fatal(err)
			}
			if !pattern.Match(operator+"()") || pattern.Match("") || pattern.Match("a") {
				t.Fatalf("standalone %s() did not become its Minimatch literal", operator)
			}
		})
		t.Run("embedded "+operator, func(t *testing.T) {
			pattern, err := CompileRelativePattern(operator + "().js")
			if err != nil {
				t.Fatal(err)
			}
			if !pattern.Match(".js") || pattern.Match("a.js") || pattern.Match(operator+"().js") {
				t.Fatalf("embedded %s() did not consume exactly the empty string", operator)
			}
		})
	}

	tests := []struct {
		pattern string
		matches []string
		misses  []string
	}{
		{pattern: "x@()", matches: []string{"x"}, misses: []string{"x@()"}},
		{pattern: "@()x", matches: []string{"x"}, misses: []string{"@()x"}},
		{pattern: "@(|).js", matches: []string{".js"}, misses: []string{"a.js"}},
		{pattern: "+(a|).js", matches: []string{".js", "a.js", "aa.js"}},
		{pattern: "@(@())", matches: []string{"@()"}, misses: []string{"", "@(@())"}},
		{pattern: "x@(@())y", matches: []string{"xy"}, misses: []string{"x@()y"}},
		{pattern: "a/!()", matches: []string{"a/x"}, misses: []string{"a/"}},
		{pattern: "a/!().js", matches: []string{"a/a.js"}, misses: []string{"a/.js"}},
		{pattern: "x!()y", matches: []string{"xay"}, misses: []string{"xy"}},
		// At the beginning of the whole pattern, the first ! is Minimatch's
		// pattern-negation marker rather than an extglob operator.
		{pattern: "!().js", matches: []string{".js", "a.js"}, misses: []string{"().js"}},
	}
	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			pattern, err := CompileRelativePattern(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range test.matches {
				if !pattern.Match(path) {
					t.Fatalf("Match(%q, %q) = false, want true", test.pattern, path)
				}
			}
			for _, path := range test.misses {
				if pattern.Match(path) {
					t.Fatalf("Match(%q, %q) = true, want false", test.pattern, path)
				}
			}
		})
	}
}

func TestExtglobDefaultRecursionLimitMatchesMinimatchTen(t *testing.T) {
	pattern, err := CompileRelativePattern(`@(x!(y!(z!(a))))`)
	if err != nil {
		t.Fatal(err)
	}
	if pattern.Match("xyza") {
		t.Fatal("extglob parsing exceeded Minimatch's default recursion limit")
	}
	if !pattern.Match("xyza)") {
		t.Fatal("over-depth extglob tail did not degrade to Minimatch's literal close")
	}
}

func TestNegativeExtglobSuffixAndFlatteningMatchMinimatchTen(t *testing.T) {
	tests := []struct {
		pattern string
		matches []string
		misses  []string
	}{
		{pattern: `x!(a)*`, matches: []string{"xba"}, misses: []string{"xaa"}},
		{pattern: `x!(a)?`, matches: []string{"xbb"}, misses: []string{"xaa"}},
		{pattern: `x!(a!(b)c)d`, matches: []string{"xabcd"}, misses: []string{"xacd"}},
		{pattern: `x!(y!(z)w)q`, matches: []string{"xyzwq"}, misses: []string{"xyawq"}},
		{pattern: `@(x!(y!(z)))q`, matches: []string{"xyq"}, misses: []string{"xq"}},
		// Whole-tree flattening usurps !(!(...)) to @ before the syntactic
		// empty-ext flag is interpreted.
		{pattern: `@(!(!(a)))`, matches: []string{"a"}, misses: []string{"b", "aa"}},
	}
	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			pattern, err := CompileRelativePattern(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range test.matches {
				if !pattern.Match(value) {
					t.Fatalf("Match(%q, %q) = false, want true", test.pattern, value)
				}
			}
			for _, value := range test.misses {
				if pattern.Match(value) {
					t.Fatalf("Match(%q, %q) = true, want false", test.pattern, value)
				}
			}
		})
	}
}

func TestGlobParent(t *testing.T) {
	tests := map[string]string{
		"packages/*/src/**/*.js": "packages",
		"/repo/src/*.js":         "/repo/src",
		"literal/path.js":        "literal",
	}
	for pattern, want := range tests {
		if got := GlobParent(pattern); got != want {
			t.Fatalf("GlobParent(%q) = %q, want %q", pattern, got, want)
		}
	}
}

func TestSearchPatternLengthUsesJavaScriptUTF16Units(t *testing.T) {
	for _, pattern := range []string{
		strings.Repeat("a", maxSearchPatternUTF16Units),
		strings.Repeat("😀", maxSearchPatternUTF16Units/2),
	} {
		if _, err := CompileRelativePattern(pattern); err != nil {
			t.Fatalf("boundary pattern failed: %v", err)
		}
	}
	for _, pattern := range []string{
		strings.Repeat("a", maxSearchPatternUTF16Units+1),
		strings.Repeat("😀", maxSearchPatternUTF16Units/2) + "a",
	} {
		if _, err := CompileRelativePattern(pattern); !errors.Is(err, ErrInvalidSearchPattern) {
			t.Fatalf("overlong pattern error = %v, want ErrInvalidSearchPattern", err)
		}
	}
}

func TestSearchPatternUsesMinimatchSegmentEncoding(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		match   bool
	}{
		{pattern: "?", path: "😀", match: false},
		{pattern: "??", path: "😀", match: true},
		{pattern: "[😀]", path: "😀", match: false},
		{pattern: "[!a]", path: "😀", match: false},
		{pattern: "?[[:alpha:]]", path: "😀a", match: true},
		{pattern: "[😀][[:alpha:]]", path: "😀a", match: true},
		{pattern: "@(x|[[:alpha:]])?", path: "x😀", match: true},
		{pattern: "?/[[:alpha:]]", path: "😀/a", match: false},
		{pattern: "{?,[[:alpha:]]}", path: "😀", match: false},
		{pattern: "[[:ascii:]]?", path: "a😀", match: false},
		{pattern: "[[:ascii:]]??", path: "a😀", match: true},
		{pattern: "[[:alpha:]]", path: "Ⅻ", match: true},
		{pattern: "[[:alnum:]]", path: "Ⅻ", match: true},
		{pattern: "[[:word:]]", path: "Ⅻ", match: true},
	}
	for _, test := range tests {
		t.Run(test.pattern+" against "+test.path, func(t *testing.T) {
			pattern, err := CompileRelativePattern(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := pattern.Match(test.path); got != test.match {
				t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
			}
		})
	}

	literal, err := CompileRelativePattern("😀/file.js")
	if err != nil {
		t.Fatal(err)
	}
	if prefixes := literal.LiteralPrefixes(); len(prefixes) != 1 || prefixes[0] != "😀/file.js" {
		t.Fatalf("LiteralPrefixes() = %q, want astral literal preserved", prefixes)
	}
}

func TestPOSIXClassesUseNodeUnicode16Baseline(t *testing.T) {
	if unicode.Version != searchUnicodeBaseVersion {
		t.Fatalf("Go Unicode version = %s, update the Unicode %s Minimatch delta from base %s", unicode.Version, searchMinimatchUnicodeVersion, searchUnicodeBaseVersion)
	}
	tests := []struct {
		class     string
		character rune
		match     bool
	}{
		{class: "alpha", character: 0x105c0, match: true},
		{class: "alnum", character: 0x10d40, match: true},
		{class: "digit", character: 0x10d40, match: true},
		{class: "lower", character: 0x1c8a, match: true},
		{class: "upper", character: 0x1c89, match: true},
		{class: "punct", character: 0x1b4e, match: true},
		{class: "graph", character: 0x1fae9, match: true},
		// Minimatch 10.2.x maps [:print:] to category C itself. A code point
		// newly assigned in Unicode 16 therefore stops matching this class.
		{class: "print", character: 0x1fae9, match: false},
	}
	for _, test := range tests {
		pattern, err := CompileRelativePattern("[[:" + test.class + ":]]")
		if err != nil {
			t.Fatal(err)
		}
		if got := pattern.Match(string(test.character)); got != test.match {
			t.Fatalf("Unicode 16 %s class for U+%04X = %v, want %v", test.class, test.character, got, test.match)
		}
	}
}

func TestGitignorePatternUsesByteSemantics(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		match   bool
	}{
		{pattern: "?", path: "😀", match: false},
		{pattern: "??", path: "😀", match: false},
		{pattern: "????", path: "😀", match: true},
		{pattern: "😀", path: "😀", match: true},
		{pattern: "[[:alpha:]]*", path: "é", match: false},
		{pattern: "[[:alpha:]]*", path: "aé", match: true},
	}
	for _, test := range tests {
		t.Run(test.pattern+" against "+test.path, func(t *testing.T) {
			pattern, err := CompileGitignorePattern(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := pattern.Match(test.path); got != test.match {
				t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
			}
		})
	}
}

func TestGlobstarRecursionLimitMatchesMinimatchTen(t *testing.T) {
	makePattern := func(sections int) string {
		return strings.Repeat("**/a/", sections) + "**/z"
	}
	makePath := func(sections int) string {
		return strings.Repeat("a/", sections) + "z"
	}
	for _, test := range []struct {
		sections int
		match    bool
	}{
		{sections: maxSearchGlobstarRecursion, match: true},
		{sections: maxSearchGlobstarRecursion + 1, match: false},
	} {
		pattern, err := CompileRelativePattern(makePattern(test.sections))
		if err != nil {
			t.Fatal(err)
		}
		path := makePath(test.sections)
		if got := pattern.Match(path); got != test.match {
			t.Fatalf("Match with %d body sections = %v, want %v", test.sections, got, test.match)
		}
		if !pattern.PartialMatch(path) {
			t.Fatalf("PartialMatch with %d body sections must conservatively remain true", test.sections)
		}
	}
}

func TestGlobstarHeadTailBehaviorMatchesMinimatchTen(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		match   bool
	}{
		{pattern: "a/**", path: "a", match: false},
		{pattern: "a/**", path: "a/", match: true},
		{pattern: "a/**/b", path: "a/b", match: true},
		{pattern: "a/**/b", path: "a/b/", match: true},
		{pattern: "**", path: "", match: true},
	}
	for _, test := range tests {
		pattern, err := CompileRelativePattern(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		if got := pattern.Match(test.path); got != test.match {
			t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
		}
	}
}

func TestGlobstarUnequalBodySectionBoundsMatchMinimatchTen(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		match   bool
	}{
		{pattern: "**/a/b/**/c/**/z", path: "a/b/c/z", match: false},
		{pattern: "**/a/b/**/c/**/z", path: "x/a/b/c/z", match: false},
		{pattern: "**/a/**/b/c/**/z", path: "a/b/c/z", match: true},
		{pattern: "**/a/b/c/**/d/**/z", path: "a/b/c/d/z", match: false},
	}
	for _, test := range tests {
		pattern, err := CompileRelativePattern(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		if got := pattern.Match(test.path); got != test.match {
			t.Fatalf("Match(%q, %q) = %v, want %v", test.pattern, test.path, got, test.match)
		}
		if !pattern.PartialMatch(test.path) {
			t.Fatalf("PartialMatch(%q, %q) must remain conservatively true", test.pattern, test.path)
		}
	}
}
