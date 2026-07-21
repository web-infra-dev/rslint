package utils

import (
	"math"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestExtractRegexPatternAndFlags(t *testing.T) {
	tests := []struct {
		input   string
		pattern string
		flags   string
	}{
		{`/abc/`, `abc`, ``},
		{`/abc/gi`, `abc`, `gi`},
		{`/abc/v`, `abc`, `v`},
		{`/a\/b/`, `a\/b`, ``},
		{`//`, ``, ``},
		{"/abc/gi" + "ms", "abc", "gi" + "ms"},
		{``, ``, ``},
		{`x`, ``, ``},
	}
	for _, tt := range tests {
		p, f := ExtractRegexPatternAndFlags(tt.input)
		if p != tt.pattern || f != tt.flags {
			t.Errorf("ExtractRegexPatternAndFlags(%q) = (%q, %q), want (%q, %q)", tt.input, p, f, tt.pattern, tt.flags)
		}
	}
}

func TestIsValidRegexLiteral(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		want    bool
	}{
		{name: "basic", literal: `/abc/g`, want: true},
		{name: "unicode sets", literal: `/[[A--B]]/v`, want: true},
		{name: "inline modifier", literal: `/(?i:foo)bar/`, want: true},
		{name: "annex b decimal escape", literal: `/\78\126\5934/`, want: true},
		{name: "invalid unicode property", literal: `/\p{NotAProperty}/u`, want: false},
		{name: "invalid v set", literal: `/[[A&&&]]/v`, want: false},
		{name: "invalid flag", literal: `/a/-`, want: false},
		{name: "conflicting unicode flags", literal: `/a/uv`, want: false},
		{name: "unterminated class", literal: `/[a/`, want: false},
		{name: "not a literal", literal: `abc`, want: false},
	}
	for _, tt := range tests {
		if got := IsValidRegexLiteral(tt.literal); got != tt.want {
			t.Errorf("%s: IsValidRegexLiteral(%q) = %v, want %v", tt.name, tt.literal, got, tt.want)
		}
	}
}

func TestHasCommentInsideNode(t *testing.T) {
	source := "const a = \"https://example.com/*x*/\";\n" +
		"const b = /\\/\\//;\n" +
		"const c = `// raw`;\n" +
		"const d = 1 /* keep */ + 2;\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "string containing comment-like text", text: "\"https://example.com/*x*/\"", want: false},
		{name: "regex containing slash text", text: "/\\/\\//", want: false},
		{name: "template containing line-comment text", text: "`// raw`", want: false},
		{name: "actual block comment", text: "1 /* keep */ + 2", want: true},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.text)
		if got := HasCommentInsideNode(sourceFile, node); got != tt.want {
			t.Errorf("%s: HasCommentInsideNode() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBracedNodeInnerRange(t *testing.T) {
	source := "class Foo { static {} }\n" +
		"class Bar {\n" +
		"  static {\n" +
		"    \n" +
		"  }\n" +
		"}\n" +
		"function work() { doWork(); }\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name      string
		blockText string
		want      string
	}{
		{name: "empty one-line block", blockText: "{}", want: ""},
		{name: "whitespace-only multiline block", blockText: "{\n    \n  }", want: "\n    \n  "},
		{name: "non-empty function block", blockText: "{ doWork(); }", want: " doWork(); "},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.blockText)
		gotRange := BracedNodeInnerRange(sourceFile, node)
		if got := sourceFile.Text()[gotRange.Pos():gotRange.End()]; got != tt.want {
			t.Errorf("%s: BracedNodeInnerRange() text = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGetStaticStringLiteralValue(t *testing.T) {
	source := "const empty = \"\";\n" +
		"const template = `raw`;\n" +
		"const number = 0;\n" +
		"const parenthesized = (\"x\");\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name     string
		text     string
		want     string
		wantOkay bool
	}{
		{name: "empty string", text: `""`, want: "", wantOkay: true},
		{name: "no substitution template", text: "`raw`", want: "raw", wantOkay: true},
		{name: "numeric literal", text: "0", want: "", wantOkay: false},
		{name: "parentheses are caller controlled", text: `("x")`, want: "", wantOkay: false},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.text)
		got, ok := GetStaticStringLiteralValue(node)
		if got != tt.want || ok != tt.wantOkay {
			t.Errorf("%s: GetStaticStringLiteralValue() = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.want, tt.wantOkay)
		}
	}
}

func TestAccessExpressionStaticName(t *testing.T) {
	source := "object.property;\n" +
		"object[\"property\"];\n" +
		"object[(\"property\")];\n" +
		"object[\"property\" as const];\n" +
		"object[dynamic];\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name     string
		text     string
		want     string
		wantOkay bool
	}{
		{name: "dot property", text: `object.property`, want: "property", wantOkay: true},
		{name: "static string element", text: `object["property"]`, want: "property", wantOkay: true},
		{name: "parenthesized element key", text: `object[("property")]`, want: "property", wantOkay: true},
		{name: "asserted element key", text: `object["property" as const]`, want: "property", wantOkay: true},
		{name: "dynamic element key", text: `object[dynamic]`, wantOkay: false},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.text)
		got, ok := AccessExpressionStaticName(node)
		if got != tt.want || ok != tt.wantOkay {
			t.Errorf("%s: AccessExpressionStaticName() = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.want, tt.wantOkay)
		}
	}
}

func TestIsIntegerElementAccess(t *testing.T) {
	source := "array[0];\n" +
		"array[(1e2)];\n" +
		"array[0x10];\n" +
		"array[1_000];\n" +
		"array[1e-400];\n" +
		"array?.[0];\n" +
		"array[1.5];\n" +
		"array[5e-324];\n" +
		"array[1e309];\n" +
		"array[-1];\n" +
		"array[1n];\n" +
		"array[\"0\"];\n" +
		"array[0 as number];\n" +
		"array.value;\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		text string
		want bool
	}{
		{text: "array[0]", want: true},
		{text: "array[(1e2)]", want: true},
		{text: "array[0x10]", want: true},
		{text: "array[1_000]", want: true},
		// tsgo normalizes this tiny literal to zero.
		{text: "array[1e-400]", want: true},
		// Optionality is a separate concern for callers; the key is still an integer.
		{text: "array?.[0]", want: true},
		{text: "array[1.5]", want: false},
		{text: "array[5e-324]", want: false},
		{text: "array[1e309]", want: false},
		{text: "array[-1]", want: false},
		{text: "array[1n]", want: false},
		{text: "array[\"0\"]", want: false},
		{text: "array[0 as number]", want: false},
		{text: "array.value", want: false},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.text)
		if got := IsIntegerElementAccess(node); got != tt.want {
			t.Errorf("IsIntegerElementAccess(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
	if IsIntegerElementAccess(nil) {
		t.Fatal("IsIntegerElementAccess(nil) = true, want false")
	}
}

func TestEslintLikePrecedence(t *testing.T) {
	source := "(a, b);\n" +
		"(a = b);\n" +
		"(() => value);\n" +
		"(condition ? yes : no);\n" +
		"call();\n" +
		"object.property;\n" +
		"object?.property;\n" +
		"(object?.property!);\n" +
		"(object!);\n" +
		"(object as any);\n" +
		"(<div />);\n" +
		"(<section>text</section>);\n" +
		"(<></>);\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		Path:     "/test.tsx",
	}, source, core.ScriptKindTSX)

	tests := []struct {
		text string
		want int
	}{
		{text: "a, b", want: 0},
		{text: "a = b", want: 1},
		{text: "() => value", want: 1},
		{text: "condition ? yes : no", want: 3},
		{text: "call()", want: 18},
		{text: "object.property", want: 20},
		{text: "object?.property", want: 18},
		// @typescript-eslint/parser exposes this complete optional chain as a
		// ChainExpression even though tsgo's outer node is NonNullExpression.
		{text: "object?.property!", want: 18},
		{text: "object!", want: -1},
		{text: "object as any", want: -1},
	}
	for _, tt := range tests {
		node := findNodeWithText(t, sourceFile, tt.text)
		if got := EslintLikePrecedence(node); got != tt.want {
			t.Errorf("EslintLikePrecedence(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
	for _, kind := range []ast.Kind{
		ast.KindJsxSelfClosingElement,
		ast.KindJsxElement,
		ast.KindJsxFragment,
	} {
		node := findFirstNodeOfKind(t, sourceFile, kind)
		if got := EslintLikePrecedence(node); got != 20 {
			t.Errorf("EslintLikePrecedence(%v) = %d, want 20", kind, got)
		}
	}
	if got := EslintLikePrecedence(nil); got != -1 {
		t.Fatalf("EslintLikePrecedence(nil) = %d, want -1", got)
	}
}

func TestResolveLegacyMaxOption(t *testing.T) {
	tests := []struct {
		name       string
		options    any
		defaultMax int
		want       int
	}{
		{name: "nil uses default", defaultMax: 3, want: 3},
		{name: "empty array uses default", options: []interface{}{}, defaultMax: 3, want: 3},
		{name: "bare number", options: 4, defaultMax: 3, want: 4},
		{name: "array number", options: []interface{}{float64(5)}, defaultMax: 3, want: 5},
		{name: "bare max object", options: map[string]interface{}{"max": 6}, defaultMax: 3, want: 6},
		{name: "array maximum object", options: []interface{}{map[string]interface{}{"maximum": 7, "max": 1}}, defaultMax: 3, want: 7},
		{name: "zero maximum falls through to max", options: []interface{}{map[string]interface{}{"maximum": 0, "max": 8}}, defaultMax: 3, want: 8},
		{name: "zero maximum without fallback disables", options: []interface{}{map[string]interface{}{"maximum": 0}}, defaultMax: 3, want: math.MaxInt},
		{name: "nonnumeric max disables", options: map[string]interface{}{"max": "wide"}, defaultMax: 3, want: math.MaxInt},
		{name: "object without max keys uses default", options: map[string]interface{}{"foo": 1}, defaultMax: 3, want: 3},
	}

	for _, tt := range tests {
		if got := ResolveLegacyMaxOption(tt.options, tt.defaultMax); got != tt.want {
			t.Errorf("%s: ResolveLegacyMaxOption() = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestGetPropertyDisplayName(t *testing.T) {
	source := "const obj = {\n" +
		"  id() {},\n" +
		"  \"quoted\"() {},\n" +
		"  0() {},\n" +
		"  [`computed`]() {},\n" +
		"  [dynamic]() {},\n" +
		"};\n" +
		"class C { #private() {} }\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name       string
		memberText string
		want       string
	}{
		{name: "identifier", memberText: "id() {}", want: "id"},
		{name: "string literal", memberText: `"quoted"() {}`, want: "quoted"},
		{name: "numeric literal", memberText: "0() {}", want: "0"},
		{name: "static computed template", memberText: "[`computed`]() {}", want: "computed"},
		{name: "dynamic computed", memberText: "[dynamic]() {}", want: ""},
		{name: "private identifier", memberText: "#private() {}", want: "#private"},
	}
	for _, tt := range tests {
		member := findNodeWithText(t, sourceFile, tt.memberText)
		if got := GetPropertyDisplayName(member.Name()); got != tt.want {
			t.Errorf("%s: GetPropertyDisplayName() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestIsThisVoidParameter(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "function f(this: void, value: void) {}\nfunction g(this: Foo) {}\n", core.ScriptKindTS)

	if IsThisVoidParameter(nil) {
		t.Fatal("IsThisVoidParameter(nil) = true, want false")
	}
	if got := IsThisVoidParameter(findNodeWithText(t, sourceFile, "this: void")); !got {
		t.Fatal("IsThisVoidParameter(this: void) = false, want true")
	}
	if got := IsThisVoidParameter(findNodeWithText(t, sourceFile, "value: void")); got {
		t.Fatal("IsThisVoidParameter(value: void) = true, want false")
	}
	if got := IsThisVoidParameter(findNodeWithText(t, sourceFile, "this: Foo")); got {
		t.Fatal("IsThisVoidParameter(this: Foo) = true, want false")
	}
}

func findNodeWithText(t *testing.T, sourceFile *ast.SourceFile, text string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found != nil || node == nil {
			return
		}
		if TrimmedNodeText(sourceFile, node) == text {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found != nil
		})
	}
	visit(&sourceFile.Node)
	if found == nil {
		t.Fatalf("missing node with text %q", text)
	}
	return found
}

func findFirstNodeOfKind(t *testing.T, sourceFile *ast.SourceFile, kind ast.Kind) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found != nil || node == nil {
			return
		}
		if node.Kind == kind {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found != nil
		})
	}
	visit(&sourceFile.Node)
	if found == nil {
		t.Fatalf("missing node of kind %v", kind)
	}
	return found
}

func TestDefaultIgnoreDirGlobs(t *testing.T) {
	globs := DefaultIgnoreDirGlobs()

	if len(globs) != len(DefaultExcludeDirNames) {
		t.Fatalf("Expected %d globs, got %d", len(DefaultExcludeDirNames), len(globs))
	}

	for i, name := range DefaultExcludeDirNames {
		expected := name + "/**"
		if globs[i] != expected {
			t.Errorf("Expected glob %q for dir %q, got %q", expected, name, globs[i])
		}
	}
}

func TestDefaultExcludeDirNames_ContainsExpected(t *testing.T) {
	expected := map[string]bool{"node_modules": false, ".git": false}

	for _, name := range DefaultExcludeDirNames {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected %q in DefaultExcludeDirNames", name)
		}
	}
}

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		// basic
		{"a", "b", -1},
		{"b", "a", 1},
		{"a", "a", 0},
		// numeric segments
		{"a2", "a10", -1},
		{"a10", "a2", 1},
		{"a1", "a1", 0},
		// leading zeros
		{"a01", "a1", 0},
		{"a02", "a1", 1},
		// length difference
		{"a", "ab", -1},
		{"ab", "a", 1},
		// multi-byte UTF-8 characters
		{"α1", "α2", -1},
		{"α2", "α10", -1},
		{"中1", "中2", -1},
		{"中10", "中2", 1},
		// empty
		{"", "", 0},
		{"a", "", 1},
		{"", "a", -1},
	}
	for _, tt := range tests {
		got := NaturalCompare(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("NaturalCompare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsConstructorName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// ── ASCII constructor forms ──
		{"Foo", true},
		{"FooBar", true},
		{"_Foo", true},
		{"$Foo", true},
		{"__Foo", true},
		{"_0Foo", true},
		{"$_Foo", true},
		{"____Foo", true},

		// ── ASCII non-constructor forms ──
		{"foo", false},
		{"fooBar", false},
		{"_foo", false},
		{"$foo", false},
		{"_0foo", false},

		// ── All-prefix → not a constructor ──
		{"", false},
		{"_", false},
		{"$", false},
		{"$$", false},
		{"_8", false},
		{"_0$_", false},

		// ── Unicode uppercase identifiers → constructor ──
		// Greek capital Pi; verifies rune-aware iteration.
		{"Πfoo", true},
		{"_Πfoo", true},
		// Cyrillic capital "Д".
		{"Дelta", true},
		// Latin Extended capital "Ǆ".
		{"ǄName", true},

		// ── Unicode lowercase identifiers → not constructor ──
		{"πfoo", false},
		{"_πfoo", false},
		{"дelta", false},

		// ── Non-ASCII digits are NOT stripped as prefix (matches ESLint's
		// `[0-9]` which only accepts ASCII 0–9). An Arabic-Indic digit at
		// the start is the first non-prefix rune and `unicode.IsUpper`
		// returns false for it → not a constructor.
		{"٠Foo", false},
	}
	for _, tt := range tests {
		got := IsConstructorName(tt.name)
		if got != tt.want {
			t.Errorf("IsConstructorName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
