// cspell:ignore subtest nukta fname
package no_misleading_character_class

// Combinatorial / property-style tests. Each test iterates a statically
// enumerated table of (input, expectation) pairs that exercises a specific
// invariant of the rule:
//
//   - detectors fire for every base-character × mark-character pair they cover
//   - breakers split a would-be pair for every (breaker × pair) combination
//   - misleading sequences outside a character class never fire (any flags)
//   - classes built from known-safe characters never fire (any flags)
//
// Everything is deterministic; failure messages identify the exact pair so
// you can reproduce by running `-run Name/subtest`.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// ---------------------------------------------------------------------------
// Fixed input tables
// ---------------------------------------------------------------------------

// Characters that are BMP, non-combining, non-surrogate, non-astral, and
// non-regex-meta — safe to place anywhere in a class.
var safeRunesTable = []rune{
	'a', 'b', 'c', 'X', 'Y', 'Z', '0', '1', '9',
	' ', '=', ':', ';', ',', '_', '~',
	'\u00A9', // © copyright sign
	'\u0410', // Cyrillic А
	'\u4E2D', // CJK 中
}

// Emoji bases that combine with skin-tone modifiers into emojiModifier pairs.
var modifierBasesTable = []rune{
	'\U0001F476', // 👶 BABY
	'\U0001F44D', // 👍 THUMBS UP
	'\U0001F483', // 💃 DANCER
	'\U0001F469', // 👩 WOMAN
}

// Emoji skin-tone modifiers (U+1F3FB..U+1F3FF).
var emojiModifiersTable = []rune{
	'\U0001F3FB', '\U0001F3FC', '\U0001F3FD', '\U0001F3FE', '\U0001F3FF',
}

// Regional indicator symbols (form flags when adjacent).
var regionalIndicatorsTable = []rune{
	'\U0001F1E6', '\U0001F1E8', '\U0001F1EF', '\U0001F1F5', '\U0001F1FA', '\U0001F1F8',
}

// Combining marks (Unicode category M).
var combiningMarksTable = []rune{
	'\u0300', '\u0301', '\u0302', '\u0303', '\u0305', '\u036F',
	'\uFE0F', // variation selector
	'\u0951', // Devanagari stress sign
	'\u093C', // Devanagari nukta
}

// Astrals that aren't modifiers/RIS (so they produce a surrogate pair
// without tripping a more-specific detector under u mode).
var genericAstralsTable = []rune{
	'\U0001F600', // 😀
	'\U0001F680', // 🚀
	'\U0001F4A9', // 💩
	'\U0001F44D', // 👍
	'\U0001F476', // 👶
}

// CharacterSet escapes that break sequences in a class.
var breakerEscapesTable = []string{`\d`, `\D`, `\w`, `\W`, `\s`, `\S`}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func hasKind(matches []foundMatch, kind string) bool {
	for _, m := range matches {
		if m.kind == kind {
			return true
		}
	}
	return false
}

func runeName(r rune) string {
	return fmt.Sprintf("U+%04X", r)
}

// ---------------------------------------------------------------------------
// Safe classes — must never fire under any flag
// ---------------------------------------------------------------------------

func TestRule_Combo_SafeClassesNeverFire(t *testing.T) {
	// Enumerate a set of concrete "safe" classes: each safe rune alone, and
	// a handful of multi-char mixtures.
	classes := []string{}
	for _, r := range safeRunesTable {
		classes = append(classes, fmt.Sprintf("[%c]", r))
	}
	// Pairs / triples using the first few safe runes — fixed, not random.
	// Use a fixed prefix size to keep the test deterministic and compact.
	prefixLen := 6
	if len(safeRunesTable) < prefixLen {
		prefixLen = len(safeRunesTable)
	}
	safePrefixes := safeRunesTable[:prefixLen]
	for i := range safePrefixes {
		for j := i + 1; j < len(safePrefixes); j++ {
			classes = append(classes, fmt.Sprintf("[%c%c]", safePrefixes[i], safePrefixes[j]))
		}
	}
	for i := range len(safePrefixes) - 2 {
		classes = append(classes,
			fmt.Sprintf("[%c%c%c]", safePrefixes[i], safePrefixes[i+1], safePrefixes[i+2]))
	}

	flagsTable := []struct {
		name  string
		flags utils.RegexFlags
	}{
		{"non-u", utils.RegexFlags{}},
		{"u", utils.RegexFlags{Unicode: true}},
		{"v", utils.RegexFlags{UnicodeSets: true}},
	}

	for _, pattern := range classes {
		for _, f := range flagsTable {
			name := fmt.Sprintf("%s/%s", f.name, pattern)
			t.Run(name, func(t *testing.T) {
				matches := scanPatternForMatches(pattern, f.flags, ruleOptions{}, 0)
				if len(matches) != 0 {
					t.Errorf("expected no matches for safe class %q; got %+v", pattern, matches)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Emoji modifier — full (base × modifier) cross product must fire
// ---------------------------------------------------------------------------

func TestRule_Combo_EmojiModifierEveryPair(t *testing.T) {
	for _, base := range modifierBasesTable {
		for _, mod := range emojiModifiersTable {
			name := fmt.Sprintf("%s+%s", runeName(base), runeName(mod))
			t.Run(name, func(t *testing.T) {
				pattern := fmt.Sprintf("[%c%c]", base, mod)
				matches := scanPatternForMatches(pattern, utils.RegexFlags{Unicode: true}, ruleOptions{}, 0)
				if !hasKind(matches, "emojiModifier") {
					t.Errorf("expected emojiModifier for %q; got %+v", pattern, matches)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Regional indicator — every pair of two RIS must fire
// ---------------------------------------------------------------------------

func TestRule_Combo_RegionalIndicatorEveryPair(t *testing.T) {
	for _, a := range regionalIndicatorsTable {
		for _, b := range regionalIndicatorsTable {
			name := fmt.Sprintf("%s+%s", runeName(a), runeName(b))
			t.Run(name, func(t *testing.T) {
				pattern := fmt.Sprintf("[%c%c]", a, b)
				matches := scanPatternForMatches(pattern, utils.RegexFlags{Unicode: true}, ruleOptions{}, 0)
				if !hasKind(matches, "regionalIndicatorSymbol") {
					t.Errorf("expected regionalIndicatorSymbol for %q; got %+v", pattern, matches)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Combining — every (ASCII-letter × combining mark) must fire
// ---------------------------------------------------------------------------

func TestRule_Combo_CombiningEveryPair(t *testing.T) {
	bases := []rune{'A', 'Z', 'a', 'z'} // representative boundaries
	for _, base := range bases {
		for _, mark := range combiningMarksTable {
			name := fmt.Sprintf("%c+%s", base, runeName(mark))
			t.Run(name, func(t *testing.T) {
				pattern := fmt.Sprintf("[%c%c]", base, mark)
				matches := scanPatternForMatches(pattern, utils.RegexFlags{Unicode: true}, ruleOptions{}, 0)
				if !hasKind(matches, "combiningClass") {
					t.Errorf("expected combiningClass for %q; got %+v", pattern, matches)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// surrogatePairWithoutUFlag — every astral under non-u mode must fire
// ---------------------------------------------------------------------------

func TestRule_Combo_SurrogatePairWithoutUFlagEveryAstral(t *testing.T) {
	astrals := append([]rune{}, genericAstralsTable...)
	astrals = append(astrals, modifierBasesTable...)
	astrals = append(astrals, regionalIndicatorsTable...)
	for _, a := range astrals {
		t.Run(runeName(a), func(t *testing.T) {
			pattern := fmt.Sprintf("[%c]", a)
			matches := scanPatternForMatches(pattern, utils.RegexFlags{}, ruleOptions{}, 0)
			if !hasKind(matches, "surrogatePairWithoutUFlag") {
				t.Errorf("expected surrogatePairWithoutUFlag for %q; got %+v", pattern, matches)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Breakers split every (breaker × base-modifier) combination
// ---------------------------------------------------------------------------

func TestRule_Combo_BreakerSplitsEveryPair(t *testing.T) {
	for _, base := range modifierBasesTable {
		for _, mod := range emojiModifiersTable {
			for _, br := range breakerEscapesTable {
				name := fmt.Sprintf("%s_%s_%s", runeName(base), br, runeName(mod))
				name = strings.ReplaceAll(name, `\`, "")
				t.Run(name, func(t *testing.T) {
					pattern := fmt.Sprintf("[%c%s%c]", base, br, mod)
					matches := scanPatternForMatches(pattern, utils.RegexFlags{Unicode: true}, ruleOptions{}, 0)
					if hasKind(matches, "emojiModifier") {
						t.Errorf("breaker %q did not split pair in %q; got %+v",
							br, pattern, matches)
					}
				})
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Ranges split pairs — `[base-mod]` must not fire
// ---------------------------------------------------------------------------

func TestRule_Combo_RangeSplitsEveryPair(t *testing.T) {
	for _, base := range modifierBasesTable {
		for _, mod := range emojiModifiersTable {
			// base must be < mod numerically for a valid range.
			if base >= mod {
				continue
			}
			name := fmt.Sprintf("%s-%s", runeName(base), runeName(mod))
			t.Run(name, func(t *testing.T) {
				pattern := fmt.Sprintf("[%c-%c]", base, mod)
				matches := scanPatternForMatches(pattern, utils.RegexFlags{Unicode: true}, ruleOptions{}, 0)
				if hasKind(matches, "emojiModifier") {
					t.Errorf("range %q falsely fired emojiModifier; got %+v", pattern, matches)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Outside-class misleading sequences never fire
// ---------------------------------------------------------------------------

func TestRule_Combo_OutsideClassNeverFires(t *testing.T) {
	cases := []string{}
	// Single misleading sequence anywhere in the pattern, outside any class.
	for _, base := range modifierBasesTable {
		for _, mod := range emojiModifiersTable {
			cases = append(cases, fmt.Sprintf("%c%c", base, mod))
			cases = append(cases, fmt.Sprintf("^%c%c$", base, mod))
			cases = append(cases, fmt.Sprintf("(%c%c)", base, mod))
		}
	}
	for _, a := range regionalIndicatorsTable {
		for _, b := range regionalIndicatorsTable {
			cases = append(cases, fmt.Sprintf("%c%c", a, b))
		}
	}
	flagsTable := []utils.RegexFlags{{}, {Unicode: true}}
	for _, pattern := range cases {
		for _, flags := range flagsTable {
			fname := "non-u"
			if flags.Unicode {
				fname = "u"
			}
			t.Run(fname+"/"+pattern, func(t *testing.T) {
				matches := scanPatternForMatches(pattern, flags, ruleOptions{}, 0)
				if len(matches) != 0 {
					t.Errorf("outside-class pattern %q (flags=%s) fired: %+v",
						pattern, fname, matches)
				}
			})
		}
	}
}

