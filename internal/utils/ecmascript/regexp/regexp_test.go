package regexp

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// Every expectation below is what `new RegExp(source, flags).test(subject)`
// answers in JavaScript.
func TestTest(t *testing.T) {
	const (
		LS = "\u2028"
		PS = "\u2029"
	)

	tests := []struct {
		name    string
		source  string
		flags   string
		subject string
		want    bool
	}{
		// ---- `.` stops at every line terminator, not just `\n` ----
		{name: "dot vs LF", source: "^.$", subject: "\n"},
		{name: "dot vs CR", source: "^.$", subject: "\r"},
		{name: "dot vs LS", source: "^.$", subject: LS},
		{name: "dot vs PS", source: "^.$", subject: PS},
		{name: "dot vs letter", source: "^.$", subject: "a", want: true},
		{name: "dotAll vs LS", source: "^.$", flags: "s", subject: LS, want: true},
		{name: "dotAll vs LF", source: "^.$", flags: "s", subject: "\n", want: true},

		// ---- anchors without `m` mean the very ends ----
		{name: "end before trailing LF", source: "^a$", subject: "a\n"},
		{name: "start after leading LF", source: "^a$", subject: "\na"},
		{name: "exact", source: "^a$", subject: "a", want: true},

		// ---- anchors with `m` break on all four terminators ----
		{name: "m start after LF", source: "^b", flags: "m", subject: "a\nb", want: true},
		{name: "m start after CR", source: "^b", flags: "m", subject: "a\rb", want: true},
		{name: "m start after LS", source: "^b", flags: "m", subject: "a" + LS + "b", want: true},
		{name: "m start after PS", source: "^b", flags: "m", subject: "a" + PS + "b", want: true},
		{name: "m start mid-word", source: "^b", flags: "m", subject: "ab"},
		{name: "m end before CR", source: "a$", flags: "m", subject: "a\rb", want: true},
		{name: "m end before LS", source: "a$", flags: "m", subject: "a" + LS + "b", want: true},
		{name: "m end mid-word", source: "a$", flags: "m", subject: "ab"},

		// ---- `i` compares the way JavaScript compares ----
		{name: "i ascii", source: "^a$", flags: "i", subject: "A", want: true},
		{name: "i sigma final", source: "^\u03a3$", flags: "i", subject: "\u03c2", want: true},
		{name: "i sigma small", source: "^\u03a3$", flags: "i", subject: "\u03c3", want: true},
		// A character outside ASCII never folds into ASCII, so these do not
		// match — which is where Unicode folding gives the other answer.
		{name: "i kelvin sign", source: "^k$", flags: "i", subject: "\u212a"},
		{name: "i long s", source: "^s$", flags: "i", subject: "\u017f"},
		{name: "i eszett", source: "^\u00df$", flags: "i", subject: "\u1e9e"},
		{name: "i range covers case", source: "^[a-z]$", flags: "i", subject: "K", want: true},
		{name: "i range excludes kelvin", source: "^[a-z]$", flags: "i", subject: "\u212a"},
		{name: "i escaped range", source: `^[\u0041-\u005A]$`, flags: "i", subject: "k", want: true},
		{name: "i negated class", source: "^[^a]$", flags: "i", subject: "A"},
		{name: "i trailing dash class", source: "^[a-]$", flags: "i", subject: "A", want: true},
		{name: "i trailing dash literal", source: "^[a-]$", flags: "i", subject: "-", want: true},
		{name: "i unicode escape", source: `^\u0041$`, flags: "i", subject: "a", want: true},
		// A group name is not text, so widening must leave it alone.
		{name: "i named group", source: "^(?<n>a)$", flags: "i", subject: "A", want: true},
		{name: "i named backreference", source: `^(?<w>a)\k<w>$`, flags: "i", subject: "aA", want: true},
		{name: "i numbered backreference", source: `^(a)\1$`, flags: "i", subject: "Aa", want: true},

		// ---- `u` switches `i` to simple case folding, which has no rule
		// about ASCII, so every pair the plain `i` cases above keep apart
		// joins back up ----
		{name: "iu kelvin sign", source: "^k$", flags: "iu", subject: "\u212a", want: true},
		{name: "iu long s", source: "^s$", flags: "iu", subject: "\u017f", want: true},
		{name: "iu angstrom sign", source: "^\u00e5$", flags: "iu", subject: "\u212b", want: true},
		{name: "iu ohm sign", source: "^\u03c9$", flags: "iu", subject: "\u2126", want: true},
		{name: "iu eszett", source: "^\u00df$", flags: "iu", subject: "\u1e9e", want: true},
		// What both readings agree on stays put.
		{name: "iu ascii", source: "^k$", flags: "iu", subject: "K", want: true},
		{name: "iu sigma final", source: "^\u03a3$", flags: "iu", subject: "\u03c2", want: true},
		{name: "iu unicode escape", source: `^\u0041$`, flags: "iu", subject: "a", want: true},
		// A class is widened by the same reading as a literal.
		{name: "iu class covers kelvin", source: "^[k]$", flags: "iu", subject: "\u212a", want: true},
		{name: "iu range covers kelvin", source: "^[a-z]$", flags: "iu", subject: "\u212a", want: true},
		{name: "iu range covers long s", source: "^[a-z]$", flags: "iu", subject: "\u017f", want: true},
		// A negated class goes on negating whatever the widened class covers.
		{name: "iu negated class", source: "^[^a]$", flags: "iu", subject: "A"},
		{name: "iu negated class kelvin", source: "^[^k]$", flags: "iu", subject: "\u212a"},

		// ---- a property escape names a set the pattern cannot spell back, so
		// under `i` it is compared by regexp2 rather than widened. Its name is
		// part of the escape and must not be widened as text ----
		{name: "u property escape", source: `^\p{Ll}$`, flags: "u", subject: "a", want: true},
		{name: "u property escape is exact", source: `^\p{Ll}$`, flags: "u", subject: "A"},
		{name: "iu property escape", source: `^\p{Ll}$`, flags: "iu", subject: "A", want: true},
		{name: "iu property escape in class", source: `^[\p{Ll}]$`, flags: "iu", subject: "A", want: true},
		{name: "iu property escape beside a literal", source: `^[\p{Ll}x]$`, flags: "iu", subject: "X", want: true},
		{name: "iu property escape the other way", source: `^\p{Lu}$`, flags: "iu", subject: "a", want: true},

		// ---- `(?i-m:…)` turns a flag on or off over one group ----
		{name: "modifier group turns i on", source: "^(?i:a)$", subject: "A", want: true},
		{name: "modifier group ends at its close", source: "^(?i:a)b$", subject: "AB"},
		{name: "modifier group restores at its close", source: "^(?i:a)b$", subject: "Ab", want: true},
		{name: "modifier group turns i off", source: "^(?-i:b)$", flags: "i", subject: "B"},
		{name: "modifier group turns i off and matches", source: "^(?-i:b)$", flags: "i", subject: "b", want: true},
		{name: "modifier group nests", source: "^(?i:a(?-i:b)c)$", subject: "AbC", want: true},
		{name: "modifier group nests and holds", source: "^(?i:a(?-i:b)c)$", subject: "ABC"},
		{name: "modifier group turns s on", source: "^(?s:.)$", subject: "\n", want: true},
		{name: "modifier group turns s off", source: "^(?-s:.)$", flags: "s", subject: "\n"},
		{name: "modifier group turns m on", source: "(?m:^b)", subject: "a\nb", want: true},
		{name: "modifier group turns m off", source: "(?-m:^b)", flags: "m", subject: "a\nb"},
		{name: "modifier group takes both sides", source: "^(?i-m:a)$", flags: "m", subject: "A", want: true},
		{name: "modifier group takes an empty second side", source: "^(?i-:a)$", subject: "A", want: true},
		{name: "modifier group reaches a word boundary", source: `(?i:\b)`, flags: "u", subject: "\u017f", want: true},

		// ---- an escape that resolves to a character is written as that
		// character; passed through, .NET would read its own meaning ----
		{name: "identity A anchor", source: `\A`, flags: "", subject: "B"},
		{name: "identity A anchor matches A", source: `\A`, flags: "", subject: "A", want: true},
		{name: "identity alarm", source: `\a`, flags: "", subject: "a", want: true},
		{name: "identity escape char", source: `\e`, flags: "", subject: "e", want: true},
		{name: "identity in class", source: `[\a]`, flags: "", subject: "a", want: true},
		{name: "property escape without u", source: `[\p]`, flags: "", subject: "p", want: true},
		// Annex B: a `\c` no control letter follows is a backslash and a `c`.
		{name: "bare control escape in class", source: `[\c]`, flags: "", subject: "c", want: true},
		{name: "bare control escape in class takes backslash", source: `[\c]`, flags: "", subject: "\\", want: true},
		{name: "control escape still works", source: `\cA`, flags: "", subject: "\x01", want: true},
		// ---- a word boundary is ASCII, where .NET reads a Unicode word set ----
		{name: "boundary between two greek letters", source: `\b`, flags: "", subject: "\u03a3\u03c3"},
		{name: "boundary after a non-ascii letter", source: `\bfoo`, flags: "", subject: "\u00e9foo", want: true},
		{name: "non-boundary between greek letters", source: `\B`, flags: "", subject: "\u03a3\u03c3", want: true},
		{name: "boundary inside a word", source: `a\bb`, flags: "", subject: "ab"},
		{name: "underscore is a word character", source: `^_\b`, flags: "", subject: "_", want: true},
		// Under `u` and `i` the two characters that fold into ASCII join the set.
		{name: "long s is a word character under iu", source: `\b`, flags: "iu", subject: "\u017f", want: true},
		{name: "long s is not one otherwise", source: `\b`, flags: "", subject: "\u017f"},
		// ---- a sticky pattern anchors at lastIndex, which is 0 here ----
		{name: "sticky anchors at the start", source: `b`, flags: "y", subject: "ab"},
		{name: "sticky matches at the start", source: `b`, flags: "y", subject: "ba", want: true},
		{name: "sticky anchors every branch", source: `a|b`, flags: "y", subject: "ba", want: true},

		// ---- classes only JavaScript spells ----
		{name: "empty class never matches", source: "[]", subject: "a"},
		{name: "negated empty class", source: "^[^]$", subject: "\n", want: true},

		// ---- syntax RE2 cannot take at all ----
		{name: "lookahead", source: "^(?=.*b)a", subject: "ab", want: true},
		{name: "lookahead negative", source: "^(?!.*x)a", subject: "ab", want: true},
		{name: "lookbehind", source: "(?<=a)b", subject: "ab", want: true},
		{name: "backreference", source: `^(a)\1$`, subject: "aa", want: true},

		// ---- `\d` and `\w` stay ASCII, as they do without `u` ----
		{name: "digit vs arabic-indic", source: `^\d$`, subject: "\u0660"},
		{name: "word vs e-acute", source: `^\w$`, subject: "\u00e9"},
		// Under `u` and `i` together `\w` gains the two characters that fold
		// into it, the same two a word boundary gains.
		{name: "iu word vs kelvin sign", source: `^\w$`, flags: "iu", subject: "\u212a", want: true},
		{name: "iu word vs long s", source: `^\w$`, flags: "iu", subject: "\u017f", want: true},
		{name: "iu non-word vs kelvin sign", source: `^\W$`, flags: "iu", subject: "\u212a"},
		{name: "iu word class vs kelvin sign", source: `^[\w]$`, flags: "iu", subject: "\u212a", want: true},
		{name: "iu non-word class vs long s", source: `^[\W]$`, flags: "iu", subject: "\u017f"},
		{name: "iu non-word class vs a space", source: `^[\W]$`, flags: "iu", subject: " ", want: true},
		{name: "iu word class beside a literal", source: `^[\wx]$`, flags: "iu", subject: "\u212a", want: true},
		{name: "i word vs kelvin sign", source: `^\w$`, flags: "i", subject: "\u212a"},
		{name: "u word vs kelvin sign", source: `^\w$`, flags: "u", subject: "\u212a"},

		// ---- an escape is read whole, and only then widened ----
		{name: "i control escape", source: `^\cA$`, flags: "i", subject: "\x01", want: true},
		{name: "i control escape is not its letter", source: `^\cA$`, flags: "i", subject: "a"},
		{name: "i control escape in class", source: `^[\cA]$`, flags: "i", subject: "\x01", want: true},
		{name: "i control escape on a lowercase letter", source: `^\ca$`, flags: "i", subject: "\x01", want: true},
		// A class is the one place Annex B lets a digit or an underscore name
		// the control character.
		{name: "control escape on a digit in class", source: `^[\c1]$`, subject: "\x11", want: true},

		// ---- Annex B's legacy octal escapes, which `u` has no room for ----
		{name: "legacy octal of two digits", source: `^\07$`, subject: "\a", want: true},
		{name: "legacy octal of three", source: `^\077$`, subject: "?", want: true},
		{name: "legacy octal stops at three", source: `^\0777$`, subject: "?7", want: true},
		{name: "legacy octal past the third", source: `^\47$`, subject: "'", want: true},
		{name: "legacy octal in class", source: `^[\07]$`, subject: "\a", want: true},
		{name: "nul then a digit", source: `^\08$`, subject: "\x008", want: true},
		// A number no group answers to is an octal escape rather than a
		// backreference.
		{name: "octal where no group answers", source: `^\1$`, subject: "\x01", want: true},
		{name: "octal in class where a group does", source: `^(a)[\1]$`, subject: "a\x01", want: true},

		// ---- without `u`, an escape no hex digits complete is its letter ----
		{name: "i incomplete hex escape", source: `^\x$`, flags: "i", subject: "X", want: true},
		{name: "i incomplete unicode escape", source: `^\u$`, flags: "i", subject: "U", want: true},
		{name: "i incomplete hex escape in class", source: `^[\x]$`, flags: "i", subject: "X", want: true},
		{name: "i hex escape on no hex digits", source: `^\xZZ$`, flags: "i", subject: "XZZ", want: true},

		// ---- inside a class `\b` is a backspace and `\B` is a letter ----
		{name: "i negated boundary in class", source: `^[\B]$`, flags: "i", subject: "b", want: true},
		{name: "backspace in class", source: `^[\b]$`, subject: "\b", want: true},
		{name: "backspace bounds a range", source: `^[\b-\r]$`, subject: "\n", want: true},

		// ---- a set escape bounds no range, so the `-` beside it is a `-` ----
		{name: "i set then dash", source: `^[\d-A]$`, flags: "i", subject: "a", want: true},
		{name: "i set then dash takes the dash", source: `^[\d-A]$`, flags: "i", subject: "-", want: true},
		{name: "i set then dash takes the set", source: `^[\d-A]$`, flags: "i", subject: "5", want: true},
		{name: "i dash then set", source: `^[A-\d]$`, flags: "i", subject: "a", want: true},
		{name: "i word set then dash", source: `^[\w-A]$`, flags: "i", subject: "a", want: true},
		{name: "i range is still a range", source: `^[A-C]$`, flags: "i", subject: "b", want: true},

		// ---- what a quantifier may follow ----
		{name: "quantified lookahead without u", source: `^(?=a)*a$`, subject: "a", want: true},
		{name: "brace that no bound closes", source: "^{", subject: "{", want: true},
		{name: "quantified escaped anchor", source: `^\^?a$`, subject: "a", want: true},
		{name: "anchor in a class", source: "^[$^]$", subject: "^", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			re, err := Compile(test.source, test.flags)
			if err != nil {
				t.Fatalf("Compile(%q, %q) = %v", test.source, test.flags, err)
			}
			if got := re.Test(test.subject); got != test.want {
				t.Errorf("/%s/%s.test(%q) = %v, want %v", test.source, test.flags, test.subject, got, test.want)
			}
		})
	}
}

func TestCompileRejects(t *testing.T) {
	tests := []struct {
		name   string
		source string
		flags  string
		is     error
	}{
		{name: "unknown flag", source: "a", flags: "q", is: ErrUnsupportedFlag},
		{name: "repeated flag", source: "a", flags: "ii", is: ErrUnsupportedFlag},
		{name: "v flag", source: "a", flags: "v", is: ErrUnsupportedFlag},
		// `(?i)` sets a flag in .NET and is a syntax error in JavaScript.
		// Taking it would give the pattern a meaning its author never wrote.
		{name: "inline flags", source: "(?i)a", is: ErrUnsupportedSyntax},
		{name: "atomic group", source: "(?>a)", is: ErrUnsupportedSyntax},
		// A modifier group names each of `i`, `m` and `s` once at most, over
		// both of its sides, and names at least one of them.
		{name: "modifier repeated", source: "(?ii:a)", is: ErrUnsupportedSyntax},
		{name: "modifier on both sides", source: "(?i-i:a)", is: ErrUnsupportedSyntax},
		{name: "modifier group naming none", source: "(?-:a)", is: ErrUnsupportedSyntax},
		{name: "modifier group naming a flag it cannot turn", source: "(?u:a)", is: ErrUnsupportedSyntax},

		// A quantifier has to have something to repeat, which has to be judged
		// before an assertion is lowered into syntax regexp2 reads otherwise.
		{name: "quantified start anchor", source: "^*", is: ErrUnsupportedSyntax},
		{name: "quantified end anchor", source: "$+", is: ErrUnsupportedSyntax},
		{name: "quantified boundary", source: `\b{1}`, is: ErrUnsupportedSyntax},
		{name: "quantified negated boundary", source: `\B?`, is: ErrUnsupportedSyntax},
		{name: "quantified anchor after an identity escape", source: `\p^?`, is: ErrUnsupportedSyntax},
		{name: "quantified lookbehind", source: `(?<=a)*`, is: ErrUnsupportedSyntax},
		{name: "quantified lookahead under u", source: `(?=a)*`, flags: "u", is: ErrUnsupportedSyntax},

		// `u` takes back what Annex B allows.
		{name: "u incomplete hex escape", source: `\x`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u incomplete unicode escape", source: `\u12`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u unicode escape naming no character", source: `\u{110000}`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u legacy octal", source: `\07`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u backreference no group answers", source: `\1`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u control escape naming nothing", source: `\c`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u identity escape", source: `\a`, flags: "u", is: ErrUnsupportedSyntax},
		{name: "u set at the end of a range", source: `[\d-A]`, flags: "u", is: ErrUnsupportedSyntax},

		{name: "range running backwards", source: "[b-a]", is: ErrUnsupportedSyntax},
		{name: "class that no bracket closes", source: "[abc", is: ErrUnsupportedSyntax},
		{name: "backslash at the end", source: `a\`, is: ErrUnsupportedSyntax},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Compile(test.source, test.flags)
			if !errors.Is(err, test.is) {
				t.Errorf("Compile(%q, %q) error = %v, want %v", test.source, test.flags, err, test.is)
			}
		})
	}
}

func TestAccessorsKeepTheWrittenForm(t *testing.T) {
	re := MustCompile("^a.b$", "im")
	if re.Source() != "^a.b$" {
		t.Errorf("Source() = %q, want the pattern as written", re.Source())
	}
	if re.Flags() != "im" {
		t.Errorf("Flags() = %q, want %q", re.Flags(), "im")
	}
	if re.String() != "/^a.b$/im" {
		t.Errorf("String() = %q, want %q", re.String(), "/^a.b$/im")
	}
}

// TestMatchingIsBounded covers the bound on how long one subject may spend in
// one pattern. A pattern a user wrote can backtrack exponentially, so this
// reads the bound off the compiled pattern rather than handing it one that
// needs it — running the pathological case to watch it stop costs a core for
// as long as the bound allows, every time the suite runs.
func TestMatchingIsBounded(t *testing.T) {
	if MatchTimeout <= 0 {
		t.Fatalf("MatchTimeout = %s, want a finite bound", MatchTimeout)
	}
	for _, source := range []string{"(a|aa)+$", "^a$", "[a-z]+", `(?<n>x)\k<n>`} {
		re := MustCompile(source, "")
		if got := re.Unwrap().MatchTimeout; got != MatchTimeout {
			t.Errorf("%q compiled with MatchTimeout %s, want %s", source, got, MatchTimeout)
		}
	}
}

// TestOrTimeoutFailsOpen covers the answer a caller wants where a no is what
// produces a report. The bound is pulled in rather than waited out, so the
// pathological pattern stops at the first check rather than costing a core for
// a second.
func TestTestOrTimeoutFailsOpen(t *testing.T) {
	re := MustCompile(`^(?:(a+)+b|a+)$`, "")
	re.Unwrap().MatchTimeout = time.Nanosecond
	subject := strings.Repeat("a", 200)

	if re.Test(subject) {
		t.Fatal("the match settled inside the bound, so there is nothing to fail open on")
	}
	if !re.TestOrTimeout(subject) {
		t.Error("TestOrTimeout read a bound match as no match, which is what turns a skipped report into a false positive")
	}
}

func TestNilIsSafe(t *testing.T) {
	var re *RegExp
	if re.Test("anything") {
		t.Error("a nil RegExp matched")
	}
	if re.Unwrap() != nil {
		t.Error("a nil RegExp unwrapped to something")
	}
}
