package unicornutil

import "testing"

func TestCallExpressionParenthesesRange(t *testing.T) {
	for _, code := range []string{
		`array.flat /* outside */ (2)`,
		`fn?.()`,
		`fn<(value: number) => number>()`,
		`fn(/\)/,)`,
		"fn(`)`,)",
	} {
		t.Run(code, func(t *testing.T) {
			sourceFile := parseTestSource(code)
			call := findTestCall(t, sourceFile, code)
			opening, closing, ok := CallExpressionParenthesesRange(sourceFile, call)
			if !ok {
				t.Fatal("CallExpressionParenthesesRange returned false")
			}
			if got := sourceFile.Text()[opening.Pos():opening.End()]; got != "(" {
				t.Fatalf("opening token = %q, want (", got)
			}
			if got := sourceFile.Text()[closing.Pos():closing.End()]; got != ")" {
				t.Fatalf("closing token = %q, want )", got)
			}
			if opening.Pos() <= call.Expression().End() && code == `array.flat /* outside */ (2)` {
				t.Fatal("opening token must start after the callee-side comment")
			}
		})
	}
}
