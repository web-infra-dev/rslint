package filename_case

import "testing"

// TestNodeExtnameParity locks in Node.js `path.extname` behaviour for the
// edge cases the upstream rule and Node both handle but a naive `LastIndex`
// implementation would get wrong (notably all-dots basenames).
func TestNodeExtnameParity(t *testing.T) {
	cases := []struct{ in, want string }{
		// no dot
		{"foo", ""},
		// leading-only dot (hidden-style filename, no real extension)
		{".foo", ""},
		{".test_utils", ""},
		// regular extension
		{"foo.js", ".js"},
		{".foo.js", ".js"},
		{".test_utils.js", ".js"},
		// trailing dot — Node returns `.`
		{"foo.", "."},
		// all-dots basename — Node returns ``
		{"..", ""},
		{"...", ""},
		{"....", ""},
		// double-dot prefix followed by real extension
		{"..js", ".js"},
		{"...js", ".js"},
		// only the last dot defines the extension
		{"foo..js", ".js"},
		{"foo.bar.baz", ".baz"},
		// empty basename
		{"", ""},
	}
	for _, c := range cases {
		got := nodeExtname(c.in)
		if got != c.want {
			t.Errorf("nodeExtname(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
