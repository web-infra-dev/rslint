// cspell:ignore globrex
package nodeutil

import "testing"

// Reference: globrex 0.1.2 with {globstar: true}, as used by eslint-plugin-n.
func TestCompileGlob(t *testing.T) {
	for _, tc := range []struct {
		pattern, name string
		want          bool
	}{
		{"*", "", true},
		{"*", "foo", true},
		{"*", "a/b", false},
		{"*", ".", true},
		{"**", "../foo", true},
		{"**", "a/../b", true},
		{"pkg/*", "pkg/", true},
		{"pkg/*", "pkg/a/b", false},
		{"pkg/**", "pkg/a/b", true},
		{"pkg/**", "pkg/", true},
		{"pkg/**", "pkg", false},
		{"pkg/**/file", "pkg/file", true},
		{"pkg/**/file", "pkg/a/file", true},
		{"pkg/**/file", "pkg/.hidden/file", true},
		{"pkg/**/file", "pkg/../file", true},
		{"pkg/***", "pkg/a/b", true},
		{"pkg/a**b", "pkg/a/x/b", false},
		{"pkg/a**b", "pkg/a-long-b", true},
		{"a//b", "a/b", true},
		{"a//b", "a//b", true},
		{"a/b", "a//b", false},
		{"file?", "files", false},
		{"file?", "file?", true},
		{"[ab]", "a", false},
		{"[ab]", "[ab]", true},
		{"{a,b}", "{a,b}", true},
		{"{a,b}", "a", false},
		{"@(a|b)", "@(a|b)", true},
		{"@(a|b)", "a", false},
		{"!(a)", "!(a)", true},
		{"a\\*", "a\\file", true},
		{"a\\*", "a*", false},
		{"#name", "#name", true},
		{" name ", " name ", true},
		{" name ", "name", false},
		{"模块/*", "模块/💡", true},
		{"\xed\xa0\xbd*", "💡", true},
		{"*\xed\xb2\xa1", "💡", true},
		{"*\xed\xa0\xbd", "\xed\xa0\xbd", true},
		{"\xed\xa0\xbd*", "�", false},
		{"*\xed\xa0\xbd", "\xed\xb0\x80", false},
		{"💡*", "\xed\xa0\xbd\xed\xb2\xa1", true},
		{"*", "\n", true},
		{"name", "name\n", false},
		{"$+^|(){}[]?\\", "$+^|(){}[]?\\", true},
		{"a*\nb", "a\nb", true},
		{"a/", "a/", true},
		{"a/", "a", false},
		{"", "", true},
		{"", "\n", false},
	} {
		if got := CompileGlob(tc.pattern).Match(tc.name); got != tc.want {
			t.Errorf("CompileGlob(%q).Match(%q) = %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}
