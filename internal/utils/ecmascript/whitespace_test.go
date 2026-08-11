package ecmascript

import "testing"

// Named so a test case reads as text rather than as an escape sequence
// running into the letter beside it.
const (
	bom  = "\ufeff" // U+FEFF byte order mark
	nel  = "\u0085" // U+0085 next line
	nbsp = "\u00a0" // U+00A0 no-break space
	ls   = "\u2028" // U+2028 line separator
)

// Every expectation here is what JavaScript answers: `s.trim()`, and
// `s.trim() === ""` for the blank cases.
func TestWhiteSpace(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "space", r: ' ', want: true},
		{name: "tab", r: '\t', want: true},
		{name: "vertical tab", r: '\v', want: true},
		{name: "form feed", r: '\f', want: true},
		{name: "newline", r: '\n', want: true},
		{name: "carriage return", r: '\r', want: true},
		{name: "line separator", r: 0x2028, want: true},
		{name: "paragraph separator", r: 0x2029, want: true},
		{name: "no-break space", r: 0x00A0, want: true},
		{name: "ideographic space", r: 0x3000, want: true},
		// U+FEFF is whitespace to JavaScript but not to Unicode, and U+0085
		// is the other way around. Go's own predicates get both backwards.
		{name: "byte order mark", r: 0xFEFF, want: true},
		{name: "next line", r: 0x0085, want: false},
		{name: "zero width space", r: 0x200B, want: false},
		{name: "letter", r: 'a', want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsWhiteSpace(test.r); got != test.want {
				t.Errorf("IsWhiteSpace(%U) = %v, want %v", test.r, got, test.want)
			}
		})
	}
}

func TestIsLineTerminator(t *testing.T) {
	for _, r := range []rune{'\n', '\r', 0x2028, 0x2029} {
		if !IsLineTerminator(r) {
			t.Errorf("IsLineTerminator(%U) = false, want true", r)
		}
		if !IsWhiteSpace(r) {
			t.Errorf("IsWhiteSpace(%U) = false, want true: every terminator is whitespace", r)
		}
	}
	// A form feed and a vertical tab are whitespace without ending a line.
	for _, r := range []rune{'\t', '\v', '\f', ' ', 0xFEFF} {
		if IsLineTerminator(r) {
			t.Errorf("IsLineTerminator(%U) = true, want false", r)
		}
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "nothing to trim", in: "a", want: "a"},
		{name: "spaces", in: "  a  ", want: "a"},
		{name: "byte order mark", in: bom + "a" + bom, want: "a"},
		{name: "no-break space", in: nbsp + "a" + nbsp, want: "a"},
		{name: "vertical tab", in: "\va\v", want: "a"},
		{name: "form feed", in: "\fa\f", want: "a"},
		{name: "line separator", in: ls + "a" + ls, want: "a"},
		// strings.TrimSpace would take this one off.
		{name: "next line stays", in: nel + "a", want: nel + "a"},
		{name: "all whitespace", in: " \t\n", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Trim(test.in); got != test.want {
				t.Errorf("Trim(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty", in: "", want: true},
		{name: "spaces", in: " \t\n", want: true},
		{name: "byte order mark", in: bom, want: true},
		{name: "no-break spaces", in: nbsp + nbsp, want: true},
		{name: "next line", in: nel, want: false},
		{name: "a letter", in: " a ", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsBlank(test.in); got != test.want {
				t.Errorf("IsBlank(%q) = %v, want %v", test.in, got, test.want)
			}
		})
	}
}
