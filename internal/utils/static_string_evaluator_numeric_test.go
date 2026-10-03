package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestStaticMathBuiltinComposition(t *testing.T) {
	cases := []struct{ method, arguments, want string }{
		{"abs", "-1", "1"},
		{"acos", "1", "0"},
		{"acosh", "1", "0"},
		{"asin", "0", "0"},
		{"asinh", "0", "0"},
		{"atan", "0", "0"},
		{"atanh", "0", "0"},
		{"atan2", "0, 1", "0"},
		{"cbrt", "8", "2"},
		{"ceil", "1.1", "2"},
		{"clz32", "1", "31"},
		{"cos", "0", "1"},
		{"cosh", "0", "1"},
		{"exp", "0", "1"},
		{"expm1", "0", "0"},
		{"f16round", "1.00048828125", "1"},
		{"floor", "1.9", "1"},
		{"fround", "1.5", "1.5"},
		{"hypot", "3, 4", "5"},
		{"imul", "4294967295, 5", "-5"},
		{"log", "1", "0"},
		{"log10", "10", "1"},
		{"log1p", "0", "0"},
		{"log2", "8", "3"},
		{"max", "1, 2", "2"},
		{"min", "1, 2", "1"},
		{"pow", "2, 3", "8"},
		{"round", "1.5", "2"},
		{"sign", "-2", "-1"},
		{"sin", "0", "0"},
		{"sinh", "0", "0"},
		{"sqrt", "9", "3"},
		{"tan", "0", "0"},
		{"tanh", "0", "0"},
		{"trunc", "1.9", "1"},
	}
	var code strings.Builder
	code.WriteString("const mathAlias = Math;\n")
	for index, test := range cases {
		fmt.Fprintf(&code, "const method%d = mathAlias.%s;\n", index, test.method)
		fmt.Fprintf(&code, "const direct%d = Math.%s(%s);\n", index, test.method, test.arguments)
		fmt.Fprintf(&code, "const alias%d = method%d(%s);\n", index, index, test.arguments)
		fmt.Fprintf(&code, "const unary%d = +Math.%s(%s);\n", index, test.method, test.arguments)
		fmt.Fprintf(&code, "const string%d = String(alias%d);\n", index, index)
	}
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code.String()})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	source := program.GetSourceFile(filePath)
	assert.Assert(t, source != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	evaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
	for index, test := range cases {
		t.Run(test.method, func(t *testing.T) {
			for _, form := range []string{"direct", "alias", "unary", "string"} {
				name := fmt.Sprintf("%s%d", form, index)
				got, known := evaluator.EvalToString(findVariableInitializer(t, source, name))
				if !known || got != test.want {
					t.Fatalf("%s = (%q, %v), want %q", name, got, known, test.want)
				}
			}
		})
	}
}
