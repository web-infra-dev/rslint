package ecmascript

import (
	"testing"
)

// Every expectation here is what JavaScript itself answers about the literal,
// checked against Node: a literal it accepts is true, one it rejects at parse
// time is false.
func TestIsValidRegexLiteral(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		want    bool
	}{
		{name: "basic", literal: `/abc/g`, want: true},
		{name: "slash equals token", literal: `/=/`, want: true},
		{name: "unicode sets", literal: `/[[A--B]]/v`, want: true},
		{name: "single v ampersand", literal: `/[&]/v`, want: true},
		{name: "escaped and single v ampersands", literal: `/[\&&]/v`, want: true},
		{name: "inline modifier", literal: `/(?i:foo)bar/`, want: true},
		{name: "annex b decimal escape", literal: `/\78\126\5934/`, want: true},
		{name: "invalid unicode property", literal: `/\p{NotAProperty}/u`, want: false},
		{name: "invalid v set", literal: `/[[A&&&]]/v`, want: false},
		{name: "invalid flag", literal: `/a/-`, want: false},
		{name: "duplicate flag", literal: `/a/gg`, want: false},
		{name: "conflicting unicode flags", literal: `/a/uv`, want: false},
		{name: "unterminated class", literal: `/[a/`, want: false},
		{name: "not a literal", literal: `abc`, want: false},
		{name: "trailing source", literal: `/a/g;`, want: false},
		// An empty pattern has to be written `/(?:)/`, because `//` opens a
		// comment.
		{name: "empty pattern", literal: `/(?:)/`, want: true},
		{name: "two slashes are a comment", literal: `//`, want: false},

		// A backreference to a group that does not exist is a legacy escape
		// where Annex B applies, and an error where it does not. This is the
		// one place the answer turns on the flags rather than the pattern.
		{name: "dangling backreference", literal: `/\1/`, want: true},
		{name: "dangling backreference under i", literal: `/\2/i`, want: true},
		{name: "dangling backreference under u", literal: `/\1/u`, want: false},
		{name: "dangling backreference under v", literal: `/\1/v`, want: false},
		{name: "dangling backreference under iu", literal: `/\2/iu`, want: false},
		{name: "annex b decimal escape under u", literal: `/\78/u`, want: false},
		// A backreference that does resolve stays valid under every flag.
		{name: "resolved backreference", literal: `/(a)\1/`, want: true},
		{name: "resolved backreference under u", literal: `/(a)\1/u`, want: true},
		{name: "missing named backreference under u", literal: `/\k<missing>/u`, want: false},
		// Unicode 16.0 and 17.0 Script / Script_Extensions values (long names and ISO 15924 short codes).
		// cspell:ignore Tulu Tigalari Beria Gukh Krai Onao Sunu Todr Tutg Berf Sidt Tayo Tols
		{name: "unicode 16 script Garay under v", literal: `/\p{Script=Garay}/v`, want: true},
		{name: "unicode 16 script short Gara under u", literal: `/\p{sc=Gara}/u`, want: true},
		{name: "unicode 16 script short Gukh under u", literal: `/\p{sc=Gukh}/u`, want: true},
		{name: "unicode 16 script short Krai under u", literal: `/\p{sc=Krai}/u`, want: true},
		{name: "unicode 16 script short Onao under u", literal: `/\p{sc=Onao}/u`, want: true},
		{name: "unicode 16 script short Sunu under u", literal: `/\p{sc=Sunu}/u`, want: true},
		{name: "unicode 16 script short Todr under u", literal: `/\p{sc=Todr}/u`, want: true},
		{name: "unicode 16 script short Tutg under u", literal: `/\p{sc=Tutg}/u`, want: true},
		{name: "unicode 16 script extensions Tulu_Tigalari under v", literal: `/\P{Script_Extensions=Tulu_Tigalari}/v`, want: true},
		{name: "unicode 17 script Beria_Erfe under v", literal: `/\p{Script=Beria_Erfe}/v`, want: true},
		{name: "unicode 17 script short Berf under u", literal: `/\p{sc=Berf}/u`, want: true},
		{name: "unicode 17 script short Sidt under u", literal: `/\p{sc=Sidt}/u`, want: true},
		{name: "unicode 17 script short Tayo under u", literal: `/\p{sc=Tayo}/u`, want: true},
		{name: "unicode 17 script short Tols under u", literal: `/\p{scx=Tols}/u`, want: true},
		{name: "unicode 16 script in negated v class", literal: `/[^a\p{Script=Gara}]/v`, want: true},
		{name: "unicode 16 script in v set subtraction", literal: `/[\p{Script=Garay}--a]/v`, want: true},
		{name: "unicode 16 script in v set intersection", literal: `/[\p{Script=Garay}&&\p{ASCII}]/v`, want: true},
		{name: "unicode 16 script nested v intersection", literal: `/[[\p{sc=Gara}]&&[a-z]]/v`, want: true},
		{name: "unicode 16 script after escaped backslash", literal: `/\\\p{Script=Garay}/v`, want: true},
		{name: "unicode 16 script with escaped p", literal: `/\\p{Script=Garay}/v`, want: false},
		{name: "unicode 16 script case-sensitive property name", literal: `/\p{script=Garay}/u`, want: false},
		{name: "unicode 16 script case-sensitive lowercase value", literal: `/\p{Script=garay}/u`, want: false},
		{name: "unicode 16 script case-sensitive uppercase short", literal: `/\p{Script=GARA}/u`, want: false},
		{name: "unicode 16 script rejected as General_Category", literal: `/\p{General_Category=Garay}/v`, want: false},
		{name: "unicode 16 script rejected as gc", literal: `/\p{gc=Gara}/u`, want: false},
		{name: "unicode 16 script rejected without Script= prefix", literal: `/\p{Garay}/v`, want: false},
		{name: "unicode 16 script followed by invalid script", literal: `/\p{Script=Garay}\p{Script=NotAScript}/v`, want: false},
		{name: "unicode 16 script trailing punctuation", literal: `/\p{Script=Garay!}/v`, want: false},
		{name: "unicode 16 script empty value", literal: `/\p{Script=}/v`, want: false},
		{name: "unicode 16 script missing closing brace", literal: `/\p{Script=Garay/v`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexLiteral(test.literal); got != test.want {
				t.Errorf("IsValidRegexLiteral(%q) = %v, want %v", test.literal, got, test.want)
			}
		})
	}
}

func TestIsValidRegexLiteralForVersion(t *testing.T) {
	tests := []struct {
		literal string
		version int
		want    bool
	}{
		{literal: `/abc/u`, version: 3, want: false},
		{literal: `/abc/u`, version: 2015, want: true},
		{literal: `/a/d`, version: 2021, want: false},
		{literal: `/a/d`, version: 2022, want: true},
		{literal: `/[[A--B]]/v`, version: 2023, want: false},
		{literal: `/[[A--B]]/v`, version: 2024, want: true},
		{literal: `/(?i:foo)bar/`, version: 2024, want: false},
		{literal: `/(?i:foo)bar/`, version: 2025, want: true},
		{literal: `/[(?i:)]/`, version: 2024, want: true},
		{literal: `/\(\?i:/`, version: 2024, want: true},
		{literal: ``, version: 2024, want: false},
	}

	for _, test := range tests {
		if got := IsValidRegexLiteralForVersion(test.literal, test.version); got != test.want {
			t.Errorf("IsValidRegexLiteralForVersion(%q, %d) = %v, want %v", test.literal, test.version, got, test.want)
		}
	}
}

func TestRegexLiteralCharacterEventCutoff(t *testing.T) {
	tests := []struct {
		literal      string
		position     int
		unterminated bool
		hasError     bool
	}{
		{literal: `/\x1f/`},
		{literal: `/=/`},
		{literal: `/[.&\x1f]/v`},
		{literal: `/\u{NOT_HEX}\x1f/`},
		{literal: `/\u{NOT_HEX}\x1f/u`, position: 4, hasError: true},
		{literal: `/\x1f\u{NOT_HEX}/u`, position: 8, hasError: true},
		{literal: `/\x1f[/`, position: 0, unterminated: true, hasError: true},
		{literal: `/\x1f/uv`, position: 7, hasError: true},
		// regexpp resolves named backreferences after character callbacks, so
		// this semantic error must not truncate a rule's collected events.
		{literal: `/\k<missing>\x1f/u`},
		{literal: `/(?<a>a)\k<missing>\x1f/u`},
		{literal: `/\p{Script=Garay}\x1f/v`},
		{literal: `/\p{sc=Gara}\x1f/v`},
		{literal: `/\p{Script=Garay}\p{Script=Bad}\x1f/v`, position: 27, hasError: true},
	}
	for _, test := range tests {
		t.Run(test.literal, func(t *testing.T) {
			position, unterminated, hasError := RegexLiteralCharacterEventCutoff(test.literal)
			if position != test.position || unterminated != test.unterminated || hasError != test.hasError {
				t.Fatalf(
					"RegexLiteralCharacterEventCutoff(%q) = (%d, %v, %v), want (%d, %v, %v)",
					test.literal,
					position,
					unterminated,
					hasError,
					test.position,
					test.unterminated,
					test.hasError,
				)
			}
		})
	}
}

func TestCanonicalizeRegExpLiteral(t *testing.T) {
	for _, test := range []struct {
		literal string
		want    string
	}{
		{literal: `/a/mi`, want: `/a/im`},
		{literal: `/a/dgimsuy`, want: `/a/dgimsuy`},
		{literal: `/a/-`, want: `/a/-`},
	} {
		if got := CanonicalizeRegExpLiteral(test.literal); got != test.want {
			t.Errorf("CanonicalizeRegExpLiteral(%q) = %q, want %q", test.literal, got, test.want)
		}
	}
}
