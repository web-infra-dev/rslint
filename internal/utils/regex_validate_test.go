package utils

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func TestRegexPatternLiteral(t *testing.T) {
	high := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	tests := []struct {
		name    string
		pattern string
		flags   RegexFlags
		want    string
		ok      bool
	}{
		{name: "empty u", flags: RegexFlags{Unicode: true}, want: `/(?:)/u`, ok: true},
		{name: "empty v", flags: RegexFlags{UnicodeSets: true}, want: `/(?:)/v`, ok: true},
		{name: "unescaped slash", pattern: `/`, flags: RegexFlags{Unicode: true}, want: `/\//u`, ok: true},
		{name: "escaped slash", pattern: `\/`, flags: RegexFlags{Unicode: true}, want: `/\//u`, ok: true},
		{name: "two backslashes and slash", pattern: `\\/`, flags: RegexFlags{Unicode: true}, want: `/\\\//u`, ok: true},
		{name: "three backslashes and slash", pattern: `\\\/`, flags: RegexFlags{Unicode: true}, want: `/\\\//u`, ok: true},
		{name: "line feed", pattern: "\n", flags: RegexFlags{Unicode: true}, want: `/\n/u`, ok: true},
		{name: "carriage return", pattern: "\r", flags: RegexFlags{Unicode: true}, want: `/\r/u`, ok: true},
		{name: "line separator", pattern: "\u2028", flags: RegexFlags{Unicode: true}, want: `/\u2028/u`, ok: true},
		{name: "paragraph separator", pattern: "\u2029", flags: RegexFlags{Unicode: true}, want: `/\u2029/u`, ok: true},
		{name: "two backslashes and line feed", pattern: "\\\\\n", flags: RegexFlags{Unicode: true}, want: `/\\\n/u`, ok: true},
		{name: "escaped line feed", pattern: "\\\n", flags: RegexFlags{Unicode: true}},
		{name: "lone surrogate", pattern: high, flags: RegexFlags{Unicode: true}, want: `/\uD800/u`, ok: true},
		{name: "surrogate pair", pattern: "😀", flags: RegexFlags{Unicode: true}, want: `/\uD83D\uDE00/u`, ok: true},
		{name: "escaped surrogate", pattern: `\` + high, flags: RegexFlags{Unicode: true}},
		{name: "trailing backslash", pattern: `\`, flags: RegexFlags{Unicode: true}},
		{name: "invalid utf8", pattern: string([]byte{0xFF}), flags: RegexFlags{Unicode: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := regexPatternLiteral(test.pattern, test.flags)
			if got != test.want || ok != test.ok {
				t.Fatalf("regexPatternLiteral(%q) = (%q, %v), want (%q, %v)", test.pattern, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestIsValidRegexPatternUnicode(t *testing.T) {
	high := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	low := ecmascript.StringFromCodeUnits([]uint16{0xDC00})
	u := RegexFlags{Unicode: true}
	v := RegexFlags{UnicodeSets: true}
	tests := []struct {
		name    string
		pattern string
		flags   RegexFlags
		want    bool
	}{
		{name: "simple", pattern: `abc`, flags: u, want: true},
		{name: "empty", flags: u, want: true},
		{name: "slash", pattern: `/`, flags: u, want: true},
		{name: "escaped slash", pattern: `\/`, flags: u, want: true},
		{name: "two backslashes and slash", pattern: `\\/`, flags: u, want: true},
		{name: "line feed", pattern: "\n", flags: u, want: true},
		{name: "escaped line feed", pattern: "\\\n", flags: u},
		{name: "two backslashes and line feed", pattern: "\\\\\n", flags: u, want: true},
		{name: "lone high surrogate", pattern: high, flags: u, want: true},
		{name: "lone low surrogate", pattern: low, flags: u, want: true},
		{name: "escaped lone surrogate", pattern: `\` + high, flags: u},
		{name: "escaped astral", pattern: `\` + "😀", flags: u},
		{name: "invalid identity escape", pattern: `\q`, flags: u},
		{name: "class identity escape", pattern: `[\B]`, flags: u},
		{name: "class named-looking escape", pattern: `[\k<a>]`, flags: u},
		{name: "missing numeric backreference", pattern: `\1`, flags: u},
		{name: "second numeric backreference missing", pattern: `(a)\2`, flags: u},
		{name: "legacy octal", pattern: `\01`, flags: u},
		{name: "escaped declaration name", pattern: `(?<\u0061>x)\k<a>`, flags: u, want: true},
		{name: "capture name beyond bundled Unicode data", pattern: `(?<𲎰>x)`, flags: u},
		{name: "escaped reference name", pattern: `(?<a>x)\k<\u0061>`, flags: u, want: true},
		{name: "dollar name", pattern: `(?<\u0024>x)\k<$>`, flags: u, want: true},
		{name: "astral escaped name", pattern: `(?<\uD835\uDC9C>x)\k<\u{1D49C}>`, flags: u, want: true},
		{name: "astral raw surrogate-pair name", pattern: "(?<" + high + low + ">x)", flags: u, want: true},
		{name: "braced leading zeros", pattern: `(?<\u{00000061}>x)\k<a>`, flags: u, want: true},
		{name: "forward named reference", pattern: `\k<a>(?<a>x)`, flags: u, want: true},
		{name: "unresolved named reference", pattern: `\k<a>`, flags: u},
		{name: "sequential duplicate name", pattern: `(?<a>x)(?<a>y)`, flags: u},
		{name: "braced surrogate halves", pattern: `(?<\u{D835}\u{DC9C}>x)`, flags: u},
		{name: "mixed surrogate halves", pattern: `(?<\uD835\u{DC9C}>x)`, flags: u},
		{name: "lone high surrogate name", pattern: `(?<\uD835>x)`, flags: u},
		{name: "lone low surrogate name", pattern: `(?<\uDC9C>x)`, flags: u},
		{name: "group text in class", pattern: `[(?<a>x)]\k<a>`, flags: u},
		{name: "unicode property long alias", pattern: `\p{Letter}`, flags: u, want: true},
		{name: "unicode property script", pattern: `\p{Script=Greek}`, flags: u, want: true},
		{name: "unicode set subtraction", pattern: `[\d--[3]]`, flags: v, want: true},
		{name: "unicode set string", pattern: `[\q{abc}]`, flags: v, want: true},
		{name: "raw slash in u class", pattern: `[/]`, flags: u, want: true},
		{name: "raw slash in v class", pattern: `[/]`, flags: v},
		{name: "escaped slash in v class", pattern: `[\/]`, flags: v, want: true},
		{name: "raw slash in v string", pattern: `[\q{a/b}]`, flags: v},
		{name: "escaped slash in v string", pattern: `[\q{a\/b}]`, flags: v, want: true},
		{name: "raw slash in nested v class", pattern: `[[/]]`, flags: v},
		{name: "trailing hyphen in u class", pattern: `[a-]`, flags: u, want: true},
		{name: "trailing hyphen in v class", pattern: `[a-]`, flags: v},
		{name: "trailing hyphen in negated v class", pattern: `[^a-]`, flags: v},
		{name: "trailing hyphen in nested v class", pattern: `[[a-]]`, flags: v},
		{name: "escaped trailing hyphen in v class", pattern: `[a\-]`, flags: v, want: true},
		{name: "subtraction hyphens in v class", pattern: `[a--b]`, flags: v, want: true},
		{name: "escaped hyphen in v string", pattern: `[\q{a\-}]`, flags: v, want: true},
		{name: "leading hyphen in v class", pattern: `[-a]`, flags: v},
		{name: "escaped leading hyphen in v class", pattern: `[\-a]`, flags: v, want: true},
		{name: "doubled caret in v class", pattern: `[a^^b]`, flags: v},
		{name: "doubled dollar in nested v class", pattern: `[[a$$b]]`, flags: v},
		{name: "negation caret is not a doubled punctuator", pattern: `[^^]`, flags: v, want: true},
		{name: "escaped doubled caret in v class", pattern: `[\^\^]`, flags: v, want: true},
		{name: "doubled dollar in v string", pattern: `[\q{a$$b}]`, flags: v},
		{name: "doubled caret in nested v string", pattern: `[[\q{a^^b}]]`, flags: v},
		{name: "escaped doubled dollar in v string", pattern: `[\q{\$\$}]`, flags: v, want: true},
		{name: "string in positive v class", pattern: `[a\q{bc}]`, flags: v, want: true},
		{name: "property in positive v class", pattern: `[a\p{Letter}]`, flags: v, want: true},
		{name: "both unicode modes", pattern: `a`, flags: RegexFlags{Unicode: true, UnicodeSets: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexPattern(test.pattern, test.flags); got != test.want {
				t.Fatalf("IsValidRegexPattern(%q, %+v) = %v, want %v", test.pattern, test.flags, got, test.want)
			}
		})
	}
}

func TestIsValidRegexPatternUnicodeSuggestionSafety(t *testing.T) {
	high := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	low := ecmascript.StringFromCodeUnits([]uint16{0xDFFF})
	v := RegexFlags{UnicodeSets: true}
	for _, test := range []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "single-code-point string in negated class", pattern: `[^a\q{b}]`, want: true},
		{name: "single-code-point string alternatives", pattern: `[^a\q{b|c}]`, want: true},
		{name: "raw astral string", pattern: `[^a\q{😀}]`, want: true},
		{name: "raw lone high surrogate string", pattern: `[^a\q{` + high + `}]`, want: true},
		{name: "raw lone low surrogate string", pattern: `[^a\q{` + low + `}]`, want: true},
		{name: "braced astral escape", pattern: `[^a\q{\u{1F600}}]`, want: true},
		{name: "fixed surrogate pair", pattern: `[^a\q{\uD83D\uDE00}]`, want: true},
		{name: "later string in negated class", pattern: `[^a\q{bc}]`},
		{name: "empty string in negated class", pattern: `[^a\q{}]`},
		{name: "empty alternative in negated class", pattern: `[^a\q{|b}]`},
		{name: "ordinary property in negated class", pattern: `[^a\p{Letter}]`, want: true},
		{name: "script property in negated class", pattern: `[^a\p{Script=Greek}]`, want: true},
		{name: "complement property in negated class", pattern: `[^a\P{Letter}]`, want: true},
		{name: "string property in negated class", pattern: `[^a\p{Basic_Emoji}]`},
		{name: "string after range in negated class", pattern: `[^a-z\q{bc}]`},
		{name: "string property after range in negated class", pattern: `[^a-z\p{Basic_Emoji}]`},
		{name: "intersection narrows strings to code points", pattern: `[^\q{ab}&&a]`, want: true},
		{name: "intersection retains strings", pattern: `[^\q{ab}&&\q{cd}]`},
		{name: "subtraction ignores string right operand", pattern: `[^a--\q{ab}]`, want: true},
		{name: "subtraction retains string left operand", pattern: `[^\q{ab}--a]`},
		{name: "nested intersection narrows strings", pattern: `[^a[\q{bc}&&a]]`, want: true},
		{name: "nested union retains strings", pattern: `[^a[\q{bc}]]`},
		{name: "string in positive class", pattern: `[a\q{bc}]`, want: true},
		{name: "property in positive class", pattern: `[a\p{Letter}]`, want: true},
		// The bundled TypeScript scanner is the grammar authority and can trail
		// the Unicode data used by the JavaScript runtime and regexpp.
		{name: "property from newer Unicode data", pattern: `[^a\p{Script=Gara}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexPatternForECMAVersion(test.pattern, v, 2024); got != test.want {
				t.Fatalf("IsValidRegexPatternForECMAVersion(%q) = %v, want %v", test.pattern, got, test.want)
			}
		})
	}
}

func TestIsValidRegexPatternFailsClosedOnScannerGaps(t *testing.T) {
	for _, test := range []struct {
		name    string
		pattern string
		flags   RegexFlags
	}{
		{
			name:    "duplicate name control flow",
			pattern: `(?=(?<a>x))(?<a>y)`,
			flags:   RegexFlags{Unicode: true},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if IsValidRegexPattern(test.pattern, test.flags) {
				t.Fatalf("IsValidRegexPattern(%q) = true, want fail-closed false", test.pattern)
			}
		})
	}
}

func TestIsValidRegexPatternECMAVersion(t *testing.T) {
	u := RegexFlags{Unicode: true}
	for _, test := range []struct {
		name        string
		pattern     string
		ecmaVersion int
		want        bool
	}{
		{name: "named capture before introduction", pattern: `(?<a>x)`, ecmaVersion: 2017},
		{name: "named capture after introduction", pattern: `(?<a>x)`, ecmaVersion: 2018, want: true},
		{name: "lookbehind before introduction", pattern: `(?<=a)b`, ecmaVersion: 2017},
		{name: "lookbehind after introduction", pattern: `(?<=a)b`, ecmaVersion: 2018, want: true},
		{name: "duplicate alternatives before relaxation", pattern: `(?<a>x)|(?<a>y)`, ecmaVersion: 2024},
		{name: "duplicate alternatives after relaxation", pattern: `(?<a>x)|(?<a>y)`, ecmaVersion: 2025, want: true},
		{name: "escaped duplicate alternatives before relaxation", pattern: `(?<\u0061>x)|(?<a>y)`, ecmaVersion: 2024},
		{name: "escaped duplicate alternatives after relaxation", pattern: `(?<\u0061>x)|(?<a>y)`, ecmaVersion: 2025, want: true},
		{name: "nested mutually exclusive duplicates", pattern: `(?:(?<a>x)|(?:y|(?<a>z)))`, ecmaVersion: 2025, want: true},
		{name: "three mutually exclusive duplicates", pattern: `(?<a>x)|(?<a>y)|(?<a>z)`, ecmaVersion: 2025, want: true},
		{name: "nested three-way mutually exclusive duplicates", pattern: `(?<a>x)|(?:(?<a>y)|(?<a>z))`, ecmaVersion: 2025, want: true},
		{name: "backreference to mutually exclusive duplicates", pattern: `(?<a>x)|(?<a>y)\k<a>`, ecmaVersion: 2025, want: true},
		{name: "unsafe nested alternative duplicate", pattern: `(?:x|(?<a>y))(?<a>z)`, ecmaVersion: 2025},
		{name: "unsafe lookahead duplicate", pattern: `(?=(?<a>x))(?<a>y)`, ecmaVersion: 2025},
		{name: "unsafe duplicate after exclusive alternatives", pattern: `(?:(?<a>x)|y)(?<a>z)`, ecmaVersion: 2025},
		{name: "third duplicate compatible with its predecessor", pattern: `(?<a>x)|(?<a>y)(?<a>z)`, ecmaVersion: 2025},
		{name: "quantifier does not make duplicates exclusive", pattern: `(?:(?<a>x))?(?<a>y)`, ecmaVersion: 2025},
		{name: "sequential duplicate remains invalid", pattern: `(?<a>x)(?<a>y)`, ecmaVersion: 2025},
		{name: "inline modifiers before introduction", pattern: `(?i:a)`, ecmaVersion: 2024},
		{name: "inline modifiers after introduction", pattern: `(?i:a)`, ecmaVersion: 2025, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexPatternForECMAVersion(test.pattern, u, test.ecmaVersion); got != test.want {
				t.Fatalf("IsValidRegexPatternForECMAVersion(%q, %d) = %v, want %v", test.pattern, test.ecmaVersion, got, test.want)
			}
		})
	}
}

func TestIsValidRegexPatternFlagECMAVersion(t *testing.T) {
	for _, test := range []struct {
		name        string
		pattern     string
		flags       RegexFlags
		ecmaVersion int
		want        bool
	}{
		{name: "u flag before introduction", pattern: `a`, flags: RegexFlags{Unicode: true}, ecmaVersion: 2014},
		{name: "u flag at introduction", pattern: `a`, flags: RegexFlags{Unicode: true}, ecmaVersion: 2015, want: true},
		{name: "v flag before introduction", pattern: `a`, flags: RegexFlags{UnicodeSets: true}, ecmaVersion: 2023},
		{name: "v flag at introduction", pattern: `a`, flags: RegexFlags{UnicodeSets: true}, ecmaVersion: 2024, want: true},
		{name: "legacy p identity escape", pattern: `\p{L}`, ecmaVersion: 2017, want: true},
		{name: "legacy k identity escape", pattern: `\k<a>`, ecmaVersion: 2017, want: true},
		{name: "escaped backslash before p in u class", pattern: `[\\p]`, flags: RegexFlags{Unicode: true}, ecmaVersion: 2017, want: true},
		{name: "u property before introduction", pattern: `\p{Letter}`, flags: RegexFlags{Unicode: true}, ecmaVersion: 2017},
		{name: "u property at introduction", pattern: `\p{Letter}`, flags: RegexFlags{Unicode: true}, ecmaVersion: 2018, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexPatternForECMAVersion(test.pattern, test.flags, test.ecmaVersion); got != test.want {
				t.Fatalf("IsValidRegexPatternForECMAVersion(%q, %+v, %d) = %v, want %v", test.pattern, test.flags, test.ecmaVersion, got, test.want)
			}
		})
	}
}

func TestIsValidRegexPatternLegacyES2025DuplicateNames(t *testing.T) {
	for _, test := range []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "top-level alternatives", pattern: `(?<a>x)|(?<a>y)`, want: true},
		{name: "escaped and raw names", pattern: `(?<\u0061>x)|(?<a>y)`, want: true},
		{name: "forward backreference", pattern: `\k<a>(?<a>x)|(?<a>y)`, want: true},
		{name: "sequential duplicate", pattern: `(?<a>x)(?<a>y)`},
		{name: "lookahead duplicate", pattern: `(?=(?<a>x))(?<a>y)`},
		{name: "group-looking text keeps legacy identity escape", pattern: `[(?<a>x)]\k<a>`, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexPattern(test.pattern, RegexFlags{}); got != test.want {
				t.Fatalf("IsValidRegexPattern(%q) = %v, want %v", test.pattern, got, test.want)
			}
		})
	}

	deepPattern := strings.Repeat("(?:", 128) + `(?<a>x)|(?<a>y)` + strings.Repeat(")", 128)
	if !IsValidRegexPattern(deepPattern, RegexFlags{}) {
		t.Fatal("deep mutually exclusive duplicate pattern should be valid")
	}

	var nestedAlternatives strings.Builder
	for range 127 {
		nestedAlternatives.WriteString(`(?<a>x)|(?:`)
	}
	nestedAlternatives.WriteString(`(?<a>x)`)
	nestedAlternatives.WriteString(strings.Repeat(")", 127))
	if !IsValidRegexPattern(nestedAlternatives.String(), RegexFlags{}) {
		t.Fatal("nested chain of mutually exclusive duplicate patterns should be valid")
	}
}

func TestIsValidRegexPatternLegacyPath(t *testing.T) {
	// Annex B accepts escapes that u/v reject. This locks in that the tsgo
	// literal adapter is confined to the Unicode modes.
	for _, pattern := range []string{"\\\n", `\q`, `\01`} {
		if !IsValidRegexPattern(pattern, RegexFlags{}) {
			t.Errorf("legacy IsValidRegexPattern(%q) = false, want true", pattern)
		}
	}
}
