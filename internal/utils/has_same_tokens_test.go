package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

// parseBinaryOperands parses code expected to contain a top-level BinaryExpression
// statement and returns its source file plus (left, right) operand nodes.
// Used to exercise HasSameTokens on both sides of a comparison.
func parseBinaryOperands(t *testing.T, code string) (*ast.SourceFile, *ast.Node, *ast.Node) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	fs := NewOverlayVFSForFile(filePath, code)
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program for code: "+code)
	sourceFile := program.GetSourceFile(filePath)

	var bin *ast.Node
	var find func(node *ast.Node) bool
	find = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindBinaryExpression {
			bin = node
			return true
		}
		return node.ForEachChild(find)
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if find(stmt) {
			break
		}
	}
	if bin == nil {
		t.Fatalf("no BinaryExpression found in code: %s", code)
	}
	b := bin.AsBinaryExpression()
	return sourceFile, b.Left, b.Right
}

func TestHasSameTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		// ---- Basic equality / inequality ----
		{"identifier equal", `x === x`, true},
		{"identifier differ", `x === y`, false},
		{"numeric equal", `1 === 1`, true},
		{"numeric differ", `1 === 2`, false},
		{"string equal", `'a' === 'a'`, true},
		{"string differ", `'a' === 'b'`, false},

		// ---- Whitespace / comments inside operand are skipped as trivia ----
		{"whitespace in member chain", `foo.bar().baz.qux === foo.bar ().baz .qux`, true},
		{"comment between tokens", `/* c */ a /* c */ === /* c */ a /* c */`, true},
		{"newlines inside operand", "a\n.\nb\n===\na.b", true},

		// ---- Member / call expressions ----
		{"property chain equal", `a.b.c === a.b.c`, true},
		{"property chain differ", `a.b.c === a.b.d`, false},
		{"element access equal", `a[0] === a[0]`, true},
		{"element access differ", `a[0] === a[1]`, false},
		{"call equal", `foo() === foo()`, true},
		{"call differ args", `foo(1) === foo(2)`, false},

		// ---- Optional chain preserved in tokens (?. vs .) ----
		{"optional vs non-optional", `a?.b === a.b`, false},
		{"optional equal", `a?.b === a?.b`, true},

		// ---- Private identifier vs bracket string access ----
		{"private vs bracket string", `this.#f === this['#f']`, false},

		// ---- ESLint-distinctive cases that AreNodesStructurallyEqual would collapse ----
		// (These are the reason HasSameTokens exists alongside AreNodesStructurallyEqual.)
		{"hex vs decimal literal", `0x1 === 1`, false},
		{"bigint hex vs decimal", `0x1n === 1n`, false},
		{"scientific vs decimal", `1e2 === 100`, false},
		{"trailing-zero decimal", `1.0 === 1`, false},
		{"single vs double quote", `'a' === "a"`, false},

		// ---- Template literals ----
		{"template equal", "`a${x}b` === `a${x}b`", true},
		{"template head differ", "`a${x}b` === `c${x}d`", false},
		{"template expr differ", "`a${x}b` === `a${y}b`", false},

		// ---- Parentheses on the operand are transparent ----
		// ESLint: ESTree doesn't model parens as nodes, so getTokens(node.left)
		// on `(x)` returns tokens inside the Identifier `x`'s own range — just
		// `[x]`. rslint: SkipParentheses in HasSameTokens drops the parens
		// before comparison. Both flag `(x) === x`.
		{"paren left, bare right", `(x) === x`, true},
		{"paren both sides", `(x) === (x)`, true},
		{"nested parens", `((x)) === x`, true},

		// ---- Unary / type-only syntax ----
		{"unary vs bare", `+x === x`, false},
		{"typeof equal", `typeof x === typeof x`, true},
		{"as-expression equal", `(x as number) === (x as number)`, true},
		{"as-expression differ type", `(x as number) === (x as string)`, false},
		{"non-null equal", `(x!) === (x!)`, true},

		// ---- Operator fields stored outside ForEachChild ----
		// tsgo's PrefixUnaryExpression / PostfixUnaryExpression store their
		// operator as a Kind enum, not as a *Node child, so a naive
		// ForEachChild-based compare would incorrectly collapse different
		// operators. Lock in operator-sensitivity here.
		{"prefix unary same op", `+x === +x`, true},
		{"prefix unary differ op", `+x === -x`, false},
		{"prefix unary update same", `++x === ++x`, true},
		{"prefix unary update differ", `++x === --x`, false},
		{"prefix typeof vs plus", `typeof x === +x`, false}, // different Kind entirely
		{"postfix unary same op", `x++ === x++`, true},
		{"postfix unary differ op", `x++ === x--`, false},
		{"prefix vs postfix update", `++x === x++`, false}, // different Kind entirely
		{"prefix logical-not same", `!x === !x`, true},     // same op → equal
		{"prefix bitwise-not same", `~x === ~x`, true},
		{"prefix bitwise vs logical not", `~x === !x`, false},

		// ---- MetaProperty (new.target / import.meta) ----
		// KeywordToken is a Kind field, not a child — name already
		// distinguishes them in practice, but the keyword check is principled.
		{"new.target equal", `new.target === new.target`, true},

		// ---- Regex ----
		{"regex equal", `/a/ === /a/`, true},
		{"regex differ", `/a/ === /b/`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sf, left, right := parseBinaryOperands(t, tt.code)
			got := HasSameTokens(sf, left, right)
			if got != tt.want {
				t.Errorf("HasSameTokens(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestHasSameTokensNilHandling(t *testing.T) {
	t.Parallel()

	sf, left, _ := parseBinaryOperands(t, `x === x`)

	if !HasSameTokens(sf, nil, nil) {
		t.Error("HasSameTokens(nil, nil) = false, want true")
	}
	if HasSameTokens(sf, nil, left) {
		t.Error("HasSameTokens(nil, left) = true, want false")
	}
	if HasSameTokens(sf, left, nil) {
		t.Error("HasSameTokens(left, nil) = true, want false")
	}
}

// TestAreNodesStructurallyEqual exercises the structural equality helper.
// Documents how AreNodesStructurallyEqual intentionally diverges from
// HasSameTokens — it normalizes literal source form (so `0x1` == `1`) — and
// locks in the operator-sensitivity fix for Prefix/Postfix unary expressions
// where ForEachChild does not visit the Operator field.
func TestAreNodesStructurallyEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		// ---- Agreement with HasSameTokens on the common cases ----
		{"identifier equal", `x === x`, true},
		{"identifier differ", `x === y`, false},
		{"property chain equal", `a.b.c === a.b.c`, true},
		{"property chain differ", `a.b.c === a.b.d`, false},
		{"call equal", `foo() === foo()`, true},
		{"optional vs non-optional", `a?.b === a.b`, false},
		{"private vs bracket string", `this.#f === this['#f']`, false},
		{"parens transparent", `(x) === x`, true},

		// ---- Normalization: intentionally collapses literal source forms ----
		// (This is the defining difference from HasSameTokens.)
		{"hex vs decimal collapse", `0x1 === 1`, true},
		{"bigint hex vs decimal collapse", `0x1n === 1n`, true},
		{"scientific vs decimal collapse", `1e2 === 100`, true},
		{"trailing-zero decimal collapse", `1.0 === 1`, true},
		{"single vs double quote collapse", `'a' === "a"`, true},
		{"different strings still differ", `'a' === 'b'`, false},
		{"different numbers still differ", `1 === 2`, false},

		// ---- Operator-sensitivity fix (Prefix/Postfix unary) ----
		{"prefix unary same op", `+x === +x`, true},
		{"prefix unary differ op", `+x === -x`, false},
		{"prefix update same", `++x === ++x`, true},
		{"prefix update differ", `++x === --x`, false},
		{"postfix unary same op", `x++ === x++`, true},
		{"postfix unary differ op", `x++ === x--`, false},
		{"bitwise vs logical not", `~x === !x`, false},

		// ---- MetaProperty ----
		{"new.target equal", `new.target === new.target`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, left, right := parseBinaryOperands(t, tt.code)
			got := AreNodesStructurallyEqual(left, right)
			if got != tt.want {
				t.Errorf("AreNodesStructurallyEqual(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestAreNodesStructurallyEqualNilHandling(t *testing.T) {
	t.Parallel()

	_, left, _ := parseBinaryOperands(t, `x === x`)

	if !AreNodesStructurallyEqual(nil, nil) {
		t.Error("AreNodesStructurallyEqual(nil, nil) = false, want true")
	}
	if AreNodesStructurallyEqual(nil, left) {
		t.Error("AreNodesStructurallyEqual(nil, left) = true, want false")
	}
	if AreNodesStructurallyEqual(left, nil) {
		t.Error("AreNodesStructurallyEqual(left, nil) = true, want false")
	}
}
