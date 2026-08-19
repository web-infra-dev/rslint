package isglob

import "testing"

// Every expectation here is what is-extglob 2.1.1 answers for the same string, read
// off the package itself rather than off its README.
func TestIsExtglob(t *testing.T) {
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
		{in: "*", want: false},
		{in: "*.js", want: false},
		{in: "**/*.ts", want: false},
		{in: "a*b", want: false},
		{in: "?", want: false},
		{in: "a?", want: false},
		{in: "a?b", want: false},
		{in: "?.js", want: false},
		// a leading bang negates, anywhere else it is a character
		{in: "!foo", want: false},
		{in: "!", want: false},
		{in: "!*.js", want: false},
		{in: "a!b", want: false},
		// a character class, which has to close to count
		{in: "[abc]", want: false},
		{in: "[abc", want: false},
		{in: "abc]", want: false},
		{in: "[]", want: false},
		{in: "[a-z]/x", want: false},
		{in: "a[b]c", want: false},
		{in: "[!a]", want: false},
		// a brace list, which has to close to count
		{in: "{a,b}", want: false},
		{in: "{a", want: false},
		{in: "a}", want: false},
		{in: "{}", want: false},
		{in: "a{b,c}d", want: false},
		{in: "{a,b}/c", want: false},
		// a parenthesised list, and the regexp-shaped groups is-glob also takes
		{in: "(a|b)", want: false},
		{in: "(a)", want: false},
		{in: "(?:a)", want: false},
		{in: "(?!a)", want: false},
		{in: "(?=a)", want: false},
		{in: "(?:)", want: false},
		{in: "a(b|c)d", want: false},
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
		{in: `\\*`, want: false},
		{in: `\`, want: false},
		{in: `a\`, want: false},
		{in: `\!`, want: false},
		{in: `\a`, want: false},
		// shapes that sit on the edge of one of the scans above
		{in: "].?", want: false},
		{in: "+.?", want: false},
		{in: ").?", want: false},
		{in: "..?", want: false},
		{in: "a].?b", want: false},
		{in: "foo/bar", want: false},
		{in: "foo/*.js", want: false},
		{in: `foo\*/bar`, want: false},
	}

	for _, test := range tests {
		if got := IsExtglob(test.in); got != test.want {
			t.Errorf("IsExtglob(%q) = %v, want %v", test.in, got, test.want)
		}
	}
}
