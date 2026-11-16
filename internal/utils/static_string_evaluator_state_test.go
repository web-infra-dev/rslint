package utils

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestStaticCallEscapeScanCost(t *testing.T) {
	for _, length := range []int{8, 16, 32, 64, 128} {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			code := "unknown" + strings.Repeat(".slice()", length) + ";"
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/chain.ts", Path: "/chain.ts"}, code, core.ScriptKindTS)
			resolver := &countingStaticReferenceResolver{}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(nil, source, resolver)
			evaluator.ensureReferenceFlags()
			if resolver.calls > 64*(length+1) {
				t.Fatalf("%d resolutions exceed linear bound for %d calls", resolver.calls, length)
			}
			if evaluator.evaluation != nil {
				t.Fatal("reference-scan values must not escape into ordinary evaluation")
			}
		})
	}
}

func TestStaticReferenceScanCacheIsolation(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
	code := `
		const array = ["fill"];
		const other = ["p"];
		function unused() { array[method]("unused"); }
		const fill = "fill";
		array[fill]("x");
		const method = array[0];
		const value = other[0] + method;
	`
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	source := program.GetSourceFile(filePath)
	assert.Assert(t, source != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	evaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
	if _, known := evaluator.Eval(findVariableInitializer(t, source, "value")); known {
		t.Fatal("mutation-scan values must not escape into ordinary evaluation")
	}
}
