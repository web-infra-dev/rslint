package test_framework

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
)

// lastAccessor parses code, takes the first call expression's member chain and
// returns the source file plus the node of its final member entry.
func lastAccessor(t *testing.T, code string) (*ast.SourceFile, *ast.Node) {
	t.Helper()
	sourceFile, call := parseFirstCall(t, code)
	entries := GetMemberEntries(call)
	if len(entries) == 0 {
		t.Fatalf("no member entries in %q", code)
	}
	return sourceFile, entries[len(entries)-1].Node
}

func TestAccessorRangeSkipsLeadingTrivia(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantToken string
		wantValue string
	}{
		{
			name:      "identifier",
			code:      "a.b()",
			wantToken: "b",
			wantValue: "b",
		},
		{
			name:      "newline after dot",
			code:      "a.\n  b()",
			wantToken: "b",
			wantValue: "b",
		},
		{
			name:      "comment after dot",
			code:      "a. /* c */ b()",
			wantToken: "b",
			wantValue: "b",
		},
		{
			name:      "line comment after dot",
			code:      "a. // c\n  b()",
			wantToken: "b",
			wantValue: "b",
		},
		{
			// The trivia belongs to the dot rather than the name here, so this
			// case passed even before the helper existed. Kept as a control.
			name:      "newline before dot",
			code:      "a\n  .b()",
			wantToken: "b",
			wantValue: "b",
		},
		{
			name:      "string literal",
			code:      "a['b']()",
			wantToken: "'b'",
			wantValue: "b",
		},
		{
			name:      "double quoted string literal",
			code:      `a["b"]()`,
			wantToken: `"b"`,
			wantValue: "b",
		},
		{
			name:      "comment before string literal",
			code:      "a[/* c */ 'b']()",
			wantToken: "'b'",
			wantValue: "b",
		},
		{
			name:      "newline before string literal",
			code:      "a[\n  'b'\n]()",
			wantToken: "'b'",
			wantValue: "b",
		},
		{
			name:      "template literal",
			code:      "a[`b`]()",
			wantToken: "`b`",
			wantValue: "b",
		},
		{
			// The value range must span the raw escape sequence, not the
			// decoded text that MemberEntry.Name carries.
			name:      "escaped string literal",
			code:      `a['to\x42e']()`,
			wantToken: `'to\x42e'`,
			wantValue: `to\x42e`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, node := lastAccessor(t, test.code)

			tokenRange, ok := AccessorRange(sourceFile, node)
			if !ok {
				t.Fatalf("AccessorRange returned false for %q", test.code)
			}
			if got := test.code[tokenRange.Pos():tokenRange.End()]; got != test.wantToken {
				t.Errorf("AccessorRange = %q, want %q", got, test.wantToken)
			}

			valueRange, ok := AccessorValueRange(sourceFile, node)
			if !ok {
				t.Fatalf("AccessorValueRange returned false for %q", test.code)
			}
			if got := test.code[valueRange.Pos():valueRange.End()]; got != test.wantValue {
				t.Errorf("AccessorValueRange = %q, want %q", got, test.wantValue)
			}
		})
	}
}

func TestAccessorReplacement(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "identifier",
			code: "a.b()",
			want: "a.x()",
		},
		{
			name: "newline after dot is preserved",
			code: "a.\n  b()",
			want: "a.\n  x()",
		},
		{
			name: "comment after dot is preserved",
			code: "a. /* c */ b()",
			want: "a. /* c */ x()",
		},
		{
			name: "single quotes are preserved",
			code: "a['b']()",
			want: "a['x']()",
		},
		{
			name: "double quotes are preserved",
			code: `a["b"]()`,
			want: `a["x"]()`,
		},
		{
			// Before this helper existed the hand-written `Pos()+1 / End()-1`
			// offsets landed inside the comment here and produced source that
			// no longer parses.
			name: "comment before string literal is preserved",
			code: "a[/* c */ 'b']()",
			want: "a[/* c */ 'x']()",
		},
		{
			name: "backticks are preserved",
			code: "a[`b`]()",
			want: "a[`x`]()",
		},
		{
			name: "escape sequence is replaced wholesale",
			code: `a['to\x42e']()`,
			want: `a['x']()`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, node := lastAccessor(t, test.code)

			replaceRange, text, ok := AccessorReplacement(sourceFile, node, "x")
			if !ok {
				t.Fatalf("AccessorReplacement returned false for %q", test.code)
			}

			got := test.code[:replaceRange.Pos()] + text + test.code[replaceRange.End():]
			if got != test.want {
				t.Errorf("applied fix = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAccessorRejectsUnsupportedNodes(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		if _, ok := AccessorRange(nil, nil); ok {
			t.Error("AccessorRange(nil, nil) = true, want false")
		}
		if _, ok := AccessorValueRange(nil, nil); ok {
			t.Error("AccessorValueRange(nil, nil) = true, want false")
		}
		if _, _, ok := AccessorReplacement(nil, nil, "x"); ok {
			t.Error("AccessorReplacement(nil, nil) = true, want false")
		}
		if IsAccessorNode(nil) {
			t.Error("IsAccessorNode(nil) = true, want false")
		}
	})

	t.Run("nil source file", func(t *testing.T) {
		_, node := lastAccessor(t, "a.b()")
		if _, ok := AccessorRange(nil, node); ok {
			t.Error("AccessorRange with nil source file = true, want false")
		}
	})

	t.Run("computed identifier key is read as a member name", func(t *testing.T) {
		// `a[b]` reads the variable b, yet the chain reports a member named
		// "b". This matches eslint-plugin-jest, whose getNodeChain
		// (src/rules/utils/parseJestFnCall.ts:44) also ignores
		// MemberExpression.computed, so a rule sees `expect(x)[toBeCalled]()`
		// as the toBeCalled matcher. Locked here as parity, not endorsed:
		// changing it would change which code rules report on, which is well
		// outside the range-handling this file is responsible for.
		_, call := parseFirstCall(t, "a[b]()")
		entries := GetMemberEntries(call)
		if got := JoinMemberEntries(entries); got != "a.b" {
			t.Errorf("GetMemberEntries(a[b]()) = %q, want %q", got, "a.b")
		}
	})

	t.Run("non-statically-nameable key yields no entries", func(t *testing.T) {
		for _, code := range []string{"a[b.c]()", "a[getName()]()", "a[0]()"} {
			_, call := parseFirstCall(t, code)
			if entries := GetMemberEntries(call); len(entries) != 0 {
				t.Errorf("GetMemberEntries(%s) = %d entries, want 0", code, len(entries))
			}
		}
	})

	t.Run("private identifier cannot be renamed", func(t *testing.T) {
		sourceFile, node := lastAccessor(t, "class C { #b() {} m() { this.#b() } }")
		if node.Kind != ast.KindPrivateIdentifier {
			t.Fatalf("expected a private identifier, got %v", node.Kind)
		}
		if _, ok := AccessorRange(sourceFile, node); !ok {
			t.Error("AccessorRange on a private identifier = false, want true")
		}
		if _, _, ok := AccessorReplacement(sourceFile, node, "x"); ok {
			t.Error("AccessorReplacement on a private identifier = true, want false")
		}
	})

	t.Run("non-accessor node", func(t *testing.T) {
		sourceFile, call := parseFirstCall(t, "a.b()")
		if IsAccessorNode(call) {
			t.Error("IsAccessorNode(CallExpression) = true, want false")
		}
		if _, ok := AccessorRange(sourceFile, call); ok {
			t.Error("AccessorRange(CallExpression) = true, want false")
		}
	})
}
