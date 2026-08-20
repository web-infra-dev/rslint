package utils

import "testing"

func TestIsESTreeStringLiteral(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "string", expression: `"value"`, want: true},
		{name: "parenthesized string", expression: `((("value")))`, want: true},
		{name: "template", expression: "`value`"},
		{name: "number", expression: `1`},
		{name: "identifier", expression: `value`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := IsESTreeStringLiteral(parseInitializer(t, test.expression)); got != test.want {
				t.Fatalf("IsESTreeStringLiteral(%s) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
	if IsESTreeStringLiteral(nil) {
		t.Fatal("nil is an ESTree string literal")
	}
}
