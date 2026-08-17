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
		{name: "unicode sets", literal: `/[[A--B]]/v`, want: true},
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidRegexLiteral(test.literal); got != test.want {
				t.Errorf("IsValidRegexLiteral(%q) = %v, want %v", test.literal, got, test.want)
			}
		})
	}
}
