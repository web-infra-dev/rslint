package sort_keys

import "testing"

// TestNaturalCompare cross-checks naturalCompare against known outputs of the
// `natural-compare` npm package (v1.4.0) that upstream sort-keys uses for its
// `natural` option, including its non-obvious character-weight ordering
// (verified directly against the real package while porting: '_' < 'A' < 'a',
// and a digit run only starts on a nonzero digit, so "a01" < "a1").
func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"a", "b", -1},
		{"b", "a", 1},
		{"a", "a", 0},
		{"", "a", -1},
		{"a", "", 1},
		{"", "", 0},
		{"a2", "a10", -1},
		{"a10", "a2", 1},
		{"a01", "a1", -1},
		{"a02", "a1", -1},
		{"9", "10", -1},
		{"10", "9", 1},
		{"2", "10", -1},
		{"10", "2", 1},
		{"$", "_", -1},
		{"_", "$", 1},
		{"_", "A", -1},
		{"A", "_", 1},
		{"_", "a", -1},
		{"a", "_", 1},
		{"A", "a", -1},
		{"a", "A", 1},
		{"Z", "a", -1},
		{"a", "Z", 1},
		{"$:1", "A:3", -1},
		{"A:3", "$:1", 1},
		{"中1", "中2", -1},
		{"中2", "中10", -1},
		{"中10", "中2", 1},
		{"α1", "α2", -1},
		{"α2", "α10", -1},
		{"a0", "a1", -1},
		{"a1", "a0", 1},
		{"a00", "a01", -1},
		{"a0b", "a0c", -1},
		{"10a", "10b", -1},
		{"9a", "9b", -1},
		{"AB", "ab", -1},
		{"ab", "Ab", 1},
		{"Ab", "aB", -1},
		{"-", ".", 1},
		{".", "0", -1},
		{"0", "9", -1},
		{"9", ":", 1},
		{":", "@", -1},
		{"@", "A", -1},
		{"Z", "[", 1},
		{"[", "`", -1},
		{"`", "a", -1},
		{"z", "{", 1},
		{"{", "~", -1},
	}
	for _, tt := range tests {
		if got := naturalCompare(tt.a, tt.b); got != tt.want {
			t.Errorf("naturalCompare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
