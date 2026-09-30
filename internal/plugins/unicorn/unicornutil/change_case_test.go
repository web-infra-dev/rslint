package unicornutil

import "testing"

func TestPascalLikeTransform(t *testing.T) {
	cases := []struct {
		word  string
		index int
		want  string
	}{
		// First-word digit start: NO leading `_`.
		{"123", 0, "123"},
		{"1foo", 0, "1foo"},
		{"5", 0, "5"},
		// Non-first-word digit start: leading `_`.
		{"123", 1, "_123"},
		{"1foo", 1, "_1foo"},
		{"5", 2, "_5"},
		// Non-first-word letter start: regular pascal-style capitalize.
		{"foo", 1, "Foo"},
		{"bar", 2, "Bar"},
		// First-word letter start: regular pascal-style capitalize.
		{"foo", 0, "Foo"},
		// JavaScript indexes one UTF-16 code unit at the word's start.
		{"𐐀", 1, "𐐀"},
		{"𐐨", 1, "𐐨"},
		// Empty word stays empty regardless of index.
		{"", 0, ""},
		{"", 1, ""},
	}
	for _, c := range cases {
		got := pascalLikeTransform(c.word, c.index)
		if got != c.want {
			t.Errorf("pascalLikeTransform(%q, %d) = %q, want %q", c.word, c.index, got, c.want)
		}
	}
}
