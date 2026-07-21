package unicornutil

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func parseTestSource(code string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)
}

func findTestNode(t *testing.T, sourceFile *ast.SourceFile, text string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found != nil || node == nil {
			return
		}
		if utils.TrimmedNodeText(sourceFile, node) == text {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found != nil
		})
	}
	visit(sourceFile.AsNode())
	if found == nil {
		t.Fatalf("missing node with text %q", text)
	}
	return found
}

func findTestCall(t *testing.T, sourceFile *ast.SourceFile, text string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found != nil || node == nil {
			return
		}
		if ast.IsCallExpression(node) && utils.TrimmedNodeText(sourceFile, node) == text {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found != nil
		})
	}
	visit(sourceFile.AsNode())
	if found == nil {
		t.Fatalf("missing call with text %q", text)
	}
	return found
}

func TestPlainParameterIdentifierAndIsSameIdentifier(t *testing.T) {
	sourceFile := parseTestSource(`
array.flatMap((value?: unknown) => ((value)));
array.flatMap((value = []) => value);
array.flatMap((...values) => values);
array.flatMap(([value]) => value);
array.flatMap((asserted: unknown) => asserted!);
`)

	typedArrow := findTestNode(t, sourceFile, "(value?: unknown) => ((value))")
	typedParameter := typedArrow.Parameters()[0]
	identifier := PlainParameterIdentifier(typedParameter)
	if identifier == nil || identifier.Text() != "value" {
		t.Fatal("typed optional parameter should expose its identifier")
	}
	if !IsSameIdentifier(identifier, typedArrow.Body()) {
		t.Fatal("parenthesized body identifier should match the parameter")
	}

	for _, parameterText := range []string{"value = []", "...values", "[value]"} {
		parameter := findTestNode(t, sourceFile, parameterText)
		if got := PlainParameterIdentifier(parameter); got != nil {
			t.Fatalf("PlainParameterIdentifier(%q) = %q, want nil",
				parameterText, got.Text())
		}
	}

	assertedArrow := findTestNode(t, sourceFile, "(asserted: unknown) => asserted!")
	if IsSameIdentifier(
		PlainParameterIdentifier(assertedArrow.Parameters()[0]),
		assertedArrow.Body(),
	) {
		t.Fatal("TypeScript assertion wrappers must not be transparent")
	}
}

func TestNodeMatchesPathNestedBoundaries(t *testing.T) {
	sourceFile := parseTestSource(`
utils.deep.flat(value);
(utils.deep).flat(value);
utils?.deep.flat(value);
utils.deep?.flat(value);
utils["deep"].flat(value);
`)
	tests := []struct {
		text string
		want bool
	}{
		{text: "utils.deep.flat", want: true},
		{text: "(utils.deep).flat", want: true},
		{text: "utils?.deep.flat"},
		{text: "utils.deep?.flat"},
		{text: `utils["deep"].flat`},
	}
	for _, test := range tests {
		node := findTestNode(t, sourceFile, test.text)
		if got := NodeMatchesPath(node, "utils.deep.flat"); got != test.want {
			t.Fatalf("NodeMatchesPath(%q) = %v, want %v", test.text, got, test.want)
		}
	}
}

func TestShouldAddParenthesesToMemberExpressionObject(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "identifier", code: "consume(value)", want: false},
		{name: "call", code: "consume(getValue())", want: false},
		{name: "dynamic import", code: "consume(import('value'))", want: true},
		{name: "function expression", code: "consume(function(){})", want: false},
		{name: "new with arguments", code: "consume(new Value())", want: false},
		{name: "new without arguments", code: "consume(new Value)", want: true},
		{name: "decimal integer", code: "consume(1)", want: true},
		{name: "decimal fraction", code: "consume(1.5)", want: false},
		{name: "hex integer", code: "consume(0x1)", want: false},
		{name: "object", code: "consume({value: 1})", want: true},
		{name: "class", code: "consume(class {})", want: true},
		{name: "arrow", code: "consume(value => value)", want: true},
		{name: "conditional", code: "consume(flag ? left : right)", want: true},
		{name: "update", code: "consume(++value)", want: true},
		{name: "string", code: `consume("value")`, want: false},
		{name: "regular expression", code: "consume(/value/)", want: false},
		{name: "bigint", code: "consume(1n)", want: false},
		{name: "template", code: "consume(`value`)", want: false},
		{name: "boolean", code: "consume(true)", want: false},
		{name: "null", code: "consume(null)", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseTestSource(test.code)
			call := findTestCall(t, sourceFile, test.code)
			if !ast.IsCallExpression(call) || len(call.Arguments()) != 1 {
				t.Fatalf("%q is not a one-argument call", test.code)
			}
			if got := ShouldAddParenthesesToMemberExpressionObject(
				sourceFile,
				call.Arguments()[0],
			); got != test.want {
				t.Fatalf("ShouldAddParenthesesToMemberExpressionObject(%q) = %v, want %v",
					test.code, got, test.want)
			}
		})
	}
}

func TestSpaceAroundKeywordFixesUsesESTreeTokenClasses(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		target    string
		wantFixes int
	}{
		{
			name:   "TypeScript as stays identifier-like",
			code:   "const result = consume(value)as unknown;",
			target: "consume(value)",
		},
		{
			name:   "TypeScript satisfies stays identifier-like",
			code:   "const result = consume(value)satisfies unknown;",
			target: "consume(value)",
		},
		{
			name:      "ECMAScript keyword after",
			code:      "const result = consume(value)instanceof Array;",
			target:    "consume(value)",
			wantFixes: 1,
		},
		{
			name:      "contextual of before",
			code:      "for (const item of[].concat(value)) {}",
			target:    "[].concat(value)",
			wantFixes: 1,
		},
		{
			name:      "contextual await before",
			code:      "async function consume() { await[].concat(value); }",
			target:    "[].concat(value)",
			wantFixes: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseTestSource(test.code)
			node := findTestCall(t, sourceFile, test.target)
			fixes := SpaceAroundKeywordFixes(sourceFile, node)
			if len(fixes) != test.wantFixes {
				t.Fatalf("SpaceAroundKeywordFixes(%q) returned %d fixes, want %d",
					test.code, len(fixes), test.wantFixes)
			}
			for _, fix := range fixes {
				if fix.Text != " " {
					t.Fatalf("SpaceAroundKeywordFixes(%q) inserted %q, want one space",
						test.code, fix.Text)
				}
			}
		})
	}
}
