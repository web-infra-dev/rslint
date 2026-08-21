package reactutil

import "testing"

// Every expectation here is what upstream's own predicate answers in Node 26,
// which is the point of the three helpers: they are read against
// eslint-plugin-react's output, not against Go's case tables. The cases past
// ASCII are the ones where the two part company — a character whose uppercase
// runs to several characters, a character Node knows a case for and Go's
// tables do not, and a character outside the BMP, which upstream reads as a
// lone surrogate with no case at all.
func TestFirstLetterCase(t *testing.T) {
	tests := []struct {
		name        string
		capitalized bool
		lower       bool
		casedLower  bool
	}{
		{name: "", capitalized: false, lower: false, casedLower: false},
		{name: "Foo", capitalized: true, lower: false, casedLower: false},
		{name: "foo", capitalized: false, lower: true, casedLower: true},
		{name: "_Foo", capitalized: true, lower: true, casedLower: false},
		{name: "中文", capitalized: true, lower: true, casedLower: false},
		// `ß` uppercases to `SS`, so it is not its own uppercase.
		{name: "ßar", capitalized: false, lower: true, casedLower: true},
		// Unicode 16 gave `ƛ` the capital `Ƛ`.
		{name: "ƛar", capitalized: false, lower: true, casedLower: true},
		{name: "Ƛar", capitalized: true, lower: false, casedLower: false},
		// A Garay letter: outside the BMP, so upstream reads a lone surrogate.
		{name: "\U00010D70ar", capitalized: true, lower: true, casedLower: false},
		// An Arabic-Indic digit has no case.
		{name: "٠Foo", capitalized: true, lower: true, casedLower: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsFirstLetterCapitalized(test.name); got != test.capitalized {
				t.Errorf("IsFirstLetterCapitalized(%q) = %v, want %v", test.name, got, test.capitalized)
			}
			if got := IsLowercaseFirstLetter(test.name); got != test.lower {
				t.Errorf("IsLowercaseFirstLetter(%q) = %v, want %v", test.name, got, test.lower)
			}
			if got := IsCasedLowercaseFirstLetter(test.name); got != test.casedLower {
				t.Errorf("IsCasedLowercaseFirstLetter(%q) = %v, want %v", test.name, got, test.casedLower)
			}
		})
	}
}
