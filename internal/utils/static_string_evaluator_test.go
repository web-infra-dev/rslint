package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestStaticStringEvaluator(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	code := "\n" +
		"const direct = \"then\";\n" +
		"const left = \"th\";\n" +
		"const concat = left + \"en\";\n" +
		"const template = `${left}en`;\n" +
		"const asserted = (\"then\" as string);\n" +
		"const satisfiesValue = \"then\" satisfies string;\n" +
		"const nested = `${concat}`;\n" +
		"const templateBoolean = `${false}`;\n" +
		"const conditional = true ? \"then\" : \"no\";\n" +
		"const conditionalFromConst = direct ? direct : \"no\";\n" +
		"const unresolvedConditional = flag ? \"then\" : \"no\";\n" +
		"const logicalOr = \"\" || \"then\";\n" +
		"const logicalAnd = \"then\" && \"then\";\n" +
		"const nullish = null ?? \"then\";\n" +
		"const undefinedNullish = undefined ?? \"then\";\n" +
		"const stringCall = String(\"then\");\n" +
		"const stringNumberCall = String(1 + 2);\n" +
		"const stringNoArgumentCall = String();\n" +
		"const stringRaw = String.raw`then`;\n" +
		"const stringRawSubstitution = String.raw`th${\"e\"}n`;\n" +
		"const RawString = String;\n" +
		"const stringRawAlias = RawString.raw`then`;\n" +
		"let MutableRawString = String;\n" +
		"MutableRawString = {raw: value => \"then\"} as any;\n" +
		"const stringRawMutableAlias = MutableRawString.raw`then`;\n" +
		"const typedRawStringAlias: StringConstructor = fake;\n" +
		"const stringRawTypedAlias = typedRawStringAlias.raw`then`;\n" +
		"{ const String = value => \"then\"; const shadowedStringCall = String(\"then\"); }\n" +
		"{ const String = { raw: value => \"then\" }; const shadowedStringRaw = String.raw`then`; }\n" +
		"let letStatic = \"then\";\n" +
		"var varStatic = \"then\";\n" +
		"const letStaticUse = letStatic;\n" +
		"const varStaticUse = varStatic;\n" +
		"let letWritten = \"then\";\n" +
		"letWritten = \"other\";\n" +
		"const letWrittenUse = letWritten;\n" +
		"var varWritten = \"then\";\n" +
		"varWritten++;\n" +
		"const varWrittenUse = varWritten;\n" +
		"let destructuredWritten = \"then\";\n" +
		"({destructuredWritten} = other);\n" +
		"const destructuredWrittenUse = destructuredWritten;\n" +
		"const notThen = \"not-then\";\n" +
		"const cycle = cycle;\n" +
		"let letValue = \"then\";\n" +
		"const letUse = letValue;\n" +
		"const numeric = 1 + 2;\n" +
		"const unknownUse = unknownValue;\n"

	fs := NewOverlayVFSForFile(filePath, code)
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")

	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)

	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{name: "direct", want: "then", ok: true},
		{name: "concat", want: "then", ok: true},
		{name: "template", want: "then", ok: true},
		{name: "asserted", want: "then", ok: true},
		{name: "satisfiesValue", want: "then", ok: true},
		{name: "nested", want: "then", ok: true},
		{name: "templateBoolean", want: "false", ok: true},
		{name: "conditional", want: "then", ok: true},
		{name: "conditionalFromConst", want: "then", ok: true},
		{name: "unresolvedConditional"},
		{name: "logicalOr", want: "then", ok: true},
		{name: "logicalAnd", want: "then", ok: true},
		{name: "nullish", want: "then", ok: true},
		{name: "undefinedNullish", want: "then", ok: true},
		{name: "stringCall", want: "then", ok: true},
		{name: "stringNumberCall", want: "3", ok: true},
		{name: "stringNoArgumentCall", want: "", ok: true},
		{name: "stringRaw", want: "then", ok: true},
		{name: "stringRawSubstitution", want: "then", ok: true},
		{name: "stringRawAlias", want: "then", ok: true},
		{name: "stringRawMutableAlias"},
		{name: "stringRawTypedAlias"},
		{name: "shadowedStringCall"},
		{name: "shadowedStringRaw"},
		{name: "letStatic", want: "then", ok: true},
		{name: "varStatic", want: "then", ok: true},
		{name: "letStaticUse", want: "then", ok: true},
		{name: "varStaticUse", want: "then", ok: true},
		{name: "letWritten", want: "then", ok: true},
		{name: "letWrittenUse"},
		{name: "varWritten", want: "then", ok: true},
		{name: "varWrittenUse"},
		{name: "destructuredWritten", want: "then", ok: true},
		{name: "destructuredWrittenUse"},
		{name: "notThen", want: "not-then", ok: true},
		{name: "cycle"},
		{name: "letUse", want: "then", ok: true},
		{name: "numeric"},
		{name: "unknownUse"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, tt.name))
			if got != tt.want || ok != tt.ok {
				t.Fatalf("Eval(%s) = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func findVariableInitializer(t *testing.T, sourceFile *ast.SourceFile, bindingName string) *ast.Node {
	t.Helper()

	var initializer *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			if declaration != nil && declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Name().AsIdentifier().Text == bindingName {
				initializer = declaration.Initializer
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	if initializer == nil {
		t.Fatalf("missing initializer for %q", bindingName)
	}
	return initializer
}
