package unicornutil

import "testing"

func TestHasOptionalChainElement(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "plain property", expression: "lib.mod.CustomError"},
		{name: "direct optional call", expression: "Error?.()", want: true},
		{name: "optional property", expression: "lib?.mod.CustomError", want: true},
		{name: "optional element", expression: "lib?.[key].CustomError", want: true},
		{name: "parentheses", expression: "((lib?.mod)).CustomError", want: true},
		{name: "as expression", expression: "(lib?.mod as any).CustomError", want: true},
		{name: "type assertion", expression: "(<any>lib?.mod).CustomError", want: true},
		{name: "satisfies expression", expression: "(lib?.mod satisfies any).CustomError", want: true},
		{name: "non-null expression", expression: "(lib?.mod!).CustomError", want: true},
		{
			name:       "nested wrappers",
			expression: "((((lib?.mod as any)!) satisfies any) as any).CustomError",
			want:       true,
		},
		{name: "optional call receiver", expression: "(factory?.() as any).CustomError", want: true},
		{
			name:       "non-optional call reaches optional callee",
			expression: "((lib?.mod as any)()).CustomError",
			want:       true,
		},
		{name: "non-optional call plain callee", expression: "((lib.mod as any)()).CustomError"},
		{name: "call argument boundary", expression: "maker(value?.member).CustomError"},
		{name: "element key boundary", expression: "lib[value?.member].CustomError"},
		{name: "instantiation boundary", expression: "(lib?.mod<T>).CustomError"},
	}

	if HasOptionalChainElement(nil) {
		t.Fatal("nil must not contain an optional-chain element")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := "consume(" + test.expression + ")"
			sourceFile := parseTestSource(code)
			call := findTestCall(t, sourceFile, code).AsCallExpression()
			if call == nil || len(call.Arguments.Nodes) != 1 {
				t.Fatalf("%q is not a one-argument call", code)
			}
			if got := HasOptionalChainElement(call.Arguments.Nodes[0]); got != test.want {
				t.Fatalf("HasOptionalChainElement(%q) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}
