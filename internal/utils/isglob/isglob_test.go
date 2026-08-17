package isglob

import "testing"

// Every expectation here is what is-glob 4.0.3 answers for the same string, read
// off the package itself rather than off its README.
func TestIs(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		// a path with no glob character in it
		{in: "", want: false},
		{in: "a", want: false},
		{in: "abc", want: false},
		{in: "a/b/c", want: false},
		{in: "src/index.ts", want: false},
		{in: "./a", want: false},
		{in: "/abs/path", want: false},
		{in: "a.b.c", want: false},
		// a star, and the question mark that only counts next to one of `].+)`
		{in: "*", want: true},
		{in: "*.js", want: true},
		{in: "**/*.ts", want: true},
		{in: "a*b", want: true},
		{in: "?", want: false},
		{in: "a?", want: false},
		{in: "a?b", want: false},
		{in: "?.js", want: false},
		// a leading bang negates, anywhere else it is a character
		{in: "!foo", want: true},
		{in: "!", want: true},
		{in: "!*.js", want: true},
		{in: "a!b", want: false},
		// a character class, which has to close to count
		{in: "[abc]", want: true},
		{in: "[abc", want: false},
		{in: "abc]", want: false},
		{in: "[]", want: false},
		{in: "[a-z]/x", want: true},
		{in: "a[b]c", want: true},
		{in: "[!a]", want: true},
		// a brace list, which has to close to count
		{in: "{a,b}", want: true},
		{in: "{a", want: false},
		{in: "a}", want: false},
		{in: "{}", want: false},
		{in: "a{b,c}d", want: true},
		{in: "{a,b}/c", want: true},
		// a parenthesised list, and the regexp-shaped groups is-glob also takes
		{in: "(a|b)", want: true},
		{in: "(a)", want: false},
		{in: "(?:a)", want: true},
		{in: "(?!a)", want: true},
		{in: "(?=a)", want: true},
		{in: "(?:)", want: false},
		{in: "a(b|c)d", want: true},
		// an extended glob list
		{in: "!(a)", want: true},
		{in: "@(a)", want: true},
		{in: "?(a)", want: true},
		{in: "+(a)", want: true},
		{in: "*(a)", want: true},
		{in: "@(a|b)", want: true},
		{in: "foo@(bar)", want: true},
		{in: "@(a", want: false},
		{in: "a@()b", want: true},
		// a backslash escapes the character after it
		{in: `\*`, want: false},
		{in: `\*.js`, want: false},
		{in: `\!foo`, want: false},
		{in: `\[abc]`, want: false},
		{in: `\{a,b}`, want: false},
		{in: `\@(a)`, want: false},
		{in: `\(a|b)`, want: false},
		{in: `a\*b`, want: false},
		{in: `\\*`, want: true},
		{in: `\`, want: false},
		{in: `a\`, want: false},
		{in: `\!`, want: false},
		{in: `\a`, want: false},
		// shapes that sit on the edge of one of the scans above
		{in: "].?", want: true},
		{in: "+.?", want: true},
		{in: ").?", want: true},
		{in: "..?", want: true},
		{in: "a].?b", want: true},
		{in: "foo/bar", want: false},
		{in: "foo/*.js", want: true},
		{in: `foo\*/bar`, want: false},
	}

	for _, test := range tests {
		if got := Is(test.in); got != test.want {
			t.Errorf("Is(%q) = %v, want %v", test.in, got, test.want)
		}
	}
}
