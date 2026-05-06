package reactutil

import "testing"

// Each row corresponds to a probe taken against the upstream
// `eslint-plugin-react/lib/rules/jsx-handler-names.js` rule (master) running
// the bundled minimatch via `RuleTester`. The expected outcome is what
// upstream produced; if a row breaks here, our `MatchGlob` has drifted from
// upstream's minimatch semantics.
func TestMatchGlob_UpstreamMinimatchAlignment(t *testing.T) {
	type c struct {
		name    string
		text    string
		pattern string
		want    bool
	}
	cases := []c{
		// ---- Plain wildcards (already supported pre-upgrade) ----
		{"star prefix", "Foobar", "Foo*", true},
		{"star middle", "Foobar", "F*r", true},
		{"star nomatch", "Bar", "Foo*", false},
		{"question one char", "Fooa", "Foo?", true},
		{"question zero char", "Foo", "Foo?", false},
		{"question two char", "Fooab", "Foo?", false},
		{"empty text", "", "*", false},
		{"empty pattern", "Foo", "", false},

		// ---- Brace expansion ----
		{"brace match a", "Foo1", "Foo{1,2}", true},
		{"brace match b", "Foo2", "Foo{1,2}", true},
		{"brace nomatch", "Foo3", "Foo{1,2}", false},
		{"brace 4-way match", "Foo3", "Foo{1,2,3,4}", true},
		{"brace single (no comma) literal", "Foo{x}", "Foo{x}", true},
		{"brace single (no comma) nomatch", "Foox", "Foo{x}", false},
		// Nested brace
		{"nested brace match outer", "X.a", "X.{a,{b,c}}", true},
		{"nested brace match inner", "X.c", "X.{a,{b,c}}", true},
		{"nested brace nomatch", "X.d", "X.{a,{b,c}}", false},

		// ---- Character class ----
		{"class match A", "TestA", "Test[ABC]", true},
		{"class match C", "TestC", "Test[ABC]", true},
		{"class nomatch D", "TestD", "Test[ABC]", false},
		// Negated class via `!`
		{"neg-bang match B", "TestB", "Test[!A]", true},
		{"neg-bang nomatch A", "TestA", "Test[!A]", false},
		// Negated class via `^`
		{"neg-caret match B", "TestB", "Test[^A]", true},
		{"neg-caret nomatch A", "TestA", "Test[^A]", false},
		// Range class
		{"range A-Z match A", "FooA", "Foo[A-Z]", true},
		{"range A-Z nomatch a", "Fooa", "Foo[A-Z]", false},

		// ---- Leading `!` whole-pattern negation ----
		{"!pattern match other", "Bar", "!Foo", true},
		{"!pattern no self", "Foo", "!Foo", false},
		// Double `!` cancels (per minimatch).
		{"!!pattern self", "Foo", "!!Foo", true},

		// ---- Extglob ----
		{"?(a) match base", "Test", "Test?(a)", true},
		{"?(a) match base+a", "Testa", "Test?(a)", true},
		{"?(a) nomatch base+aa", "Testaa", "Test?(a)", false},
		{"*(ab) match base", "Test", "Test*(ab)", true},
		{"*(ab) match base+ab", "Testab", "Test*(ab)", true},
		{"*(ab) match base+abab", "Testabab", "Test*(ab)", true},
		{"+(a|b) nomatch base", "Test", "Test+(a|b)", false},
		{"+(a|b) match base+a", "Testa", "Test+(a|b)", true},
		{"+(a|b) match base+ab", "Testab", "Test+(a|b)", true},
		{"@(a|b) match a", "Testa", "Test@(a|b)", true},
		{"@(a|b) nomatch ab", "Testab", "Test@(a|b)", false},

		// ---- Mixed brace + extglob + wildcard ----
		{"brace then star", "Foo.aBar", "Foo.{a,b}*", true},
		{"brace then star nomatch", "Foo.cBar", "Foo.{a,b}*", false},

		// ---- Dot literal ----
		// minimatch's `.` is just a literal char (not regex meta). Our impl
		// must escape it.
		{"dot literal match", "Foo.Bar", "Foo.Bar", true},
		{"dot literal nomatch", "FooXBar", "Foo.Bar", false},
		// `*` matches dotted paths in component-name context (no `/`).
		{"star matches dotted", "Foo.X.Y", "Foo.*", true},

		// ---- Backslash escape ----
		{"escape brace", "{x}", `\{x\}`, true},
		{"escape star", "a*b", `a\*b`, true},

		// ---- Multibyte / unicode ----
		{"unicode literal", "Año", "Año", true},
		{"unicode star", "Año_Foo", "Año_*", true},
		{"emoji literal", "🚀X", "🚀*", true},

		// ---- Empty / malformed ----
		// A pattern that produces a malformed regex falls back to "no match"
		// (we don't crash). `[` with no closing `]` is treated as literal `[`.
		{"unbalanced [ as literal", "[abc", "[abc", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchGlob(tc.text, tc.pattern)
			if got != tc.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tc.text, tc.pattern, got, tc.want)
			}
		})
	}
}

// Locks in caching behavior — repeated calls with the same pattern must
// return the same compiled regex (a sync.Map hit, not a recompilation).
func TestGlobToRegex_Caching(t *testing.T) {
	pat := "Foo{1,2,3}"
	a := GlobToRegex(pat)
	b := GlobToRegex(pat)
	if a == nil || b == nil {
		t.Fatalf("expected non-nil regex for %q", pat)
	}
	if a != b {
		t.Errorf("expected cached identity for repeated GlobToRegex(%q)", pat)
	}
}
