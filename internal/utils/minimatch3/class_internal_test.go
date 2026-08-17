package minimatch3

import "testing"

func TestIsValidJSClass(t *testing.T) {
	tests := []struct {
		name  string
		class string
		want  bool
	}{
		{name: "ordinary", class: "[a-z]", want: true},
		{name: "nested opening bracket", class: "[[]", want: true},
		{name: "descending range", class: "[z-a]", want: false},
		{name: "ambiguous closing bracket", class: "[]^?]", want: false},
		{name: "descending range to line feed", class: "[!-\n]", want: false},
		{name: "line feed", class: "[a\n]", want: true},
		{name: "escaped line feed starts range", class: "[\\\n-a]", want: true},
		{name: "two backslashes before line feed", class: "[\\\\\n-a]", want: true},
		{name: "descending carriage return range", class: "[\r-\n]", want: false},
		{name: "escaped line separator starts descending range", class: "[\\\u2028-a]", want: false},
		{name: "slash", class: "[/]", want: true},
		{name: "escaped slash", class: "[\\/]", want: true},
		{name: "slash after two backslashes", class: "[\\\\/]", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidJSClass(test.class); got != test.want {
				t.Errorf("isValidJSClass(%q) = %v, want %v", test.class, got, test.want)
			}
		})
	}
}
