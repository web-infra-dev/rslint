package typescriptutil

import "testing"

func TestJSXFactoryRoot(t *testing.T) {
	for _, testCase := range []struct {
		input string
		want  string
	}{
		{input: "React.createElement", want: "React"},
		{input: "h", want: "h"},
		{input: "  Preact.h  ", want: "Preact"},
	} {
		if got := JSXFactoryRoot(testCase.input); got != testCase.want {
			t.Fatalf("JSXFactoryRoot(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
}
