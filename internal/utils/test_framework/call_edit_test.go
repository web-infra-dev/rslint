package test_framework

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestFollowTypeAssertionChain(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: `consume(((value as unknown) as string))`, want: `value`},
		{code: `consume(<number>(value))`, want: `value`},
		{code: `consume(value!)`, want: `value!`},
		{code: `consume(value satisfies number)`, want: `value satisfies number`},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			sourceFile, call := parseFirstCall(t, test.code)
			arguments := call.Arguments()
			if len(arguments) != 1 {
				t.Fatalf("arguments = %d, want 1", len(arguments))
			}
			got := FollowTypeAssertionChain(arguments[0])
			if text := utils.TrimmedNodeText(sourceFile, got); text != test.want {
				t.Fatalf("unwrapped expression = %q, want %q", text, test.want)
			}
		})
	}
}

func TestInvokedAccessorCall(t *testing.T) {
	sourceFile, outerCall := parseFirstCall(t, `expect(value).toBe(true)()`)
	entries := GetMemberEntries(outerCall)
	if len(entries) == 0 {
		t.Fatal("expected a member entry")
	}

	matcherCall := InvokedAccessorCall(&entries[len(entries)-1])
	if matcherCall == nil {
		t.Fatal("expected the matcher call")
	}
	if got := utils.TrimmedNodeText(sourceFile, matcherCall); got != `expect(value).toBe(true)` {
		t.Fatalf("matcher call = %q", got)
	}
}

func TestCallArgumentListRange(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "arguments", code: `expect(value).toBe(1, 2,)`, want: `1, 2,`},
		{name: "comments", code: `expect(value).toBe(/* keep */ 1)`, want: `/* keep */ 1`},
		{name: "type arguments", code: `expect(value).toBe<(x: string) => void>(fn)`, want: `fn`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, call := parseFirstCall(t, test.code)
			textRange, ok := CallArgumentListRange(sourceFile, call)
			if !ok {
				t.Fatalf("CallArgumentListRange returned false for %q", test.code)
			}
			if got := test.code[textRange.Pos():textRange.End()]; got != test.want {
				t.Fatalf("argument list = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCallTypeArgumentListRange(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "single", code: `expect(value).toEqual<null>(null)`, want: `<null>`},
		{name: "nested", code: `expect(value).toEqual<Array<null>>(null)`, want: `<Array<null>>`},
		{name: "spaced", code: `expect(value).toEqual  < null > (null)`, want: `< null >`},
		{name: "optional call", code: `expect(value).toEqual?.<null>(null)`, want: `<null>`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, call := parseFirstCall(t, test.code)
			textRange, ok := CallTypeArgumentListRange(sourceFile, call)
			if !ok {
				t.Fatalf("CallTypeArgumentListRange returned false for %q", test.code)
			}
			if got := test.code[textRange.Pos():textRange.End()]; got != test.want {
				t.Fatalf("type argument list = %q, want %q", got, test.want)
			}
		})
	}

	t.Run("no type arguments", func(t *testing.T) {
		sourceFile, call := parseFirstCall(t, `expect(value).toEqual(null)`)
		if _, ok := CallTypeArgumentListRange(sourceFile, call); ok {
			t.Fatal("CallTypeArgumentListRange returned true without type arguments")
		}
	})
}
