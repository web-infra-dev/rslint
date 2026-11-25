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

func TestStaticPropertyShapeMutations(t *testing.T) {
	for _, test := range []struct {
		name, operation string
		changed         bool
	}{
		{"copy", "", false},
		{"named write", "copy.value = 2;", false},
		{"numeric write", "(copy as any)[0] = 2;", false},
		{"alias write", "const alias = copy; alias.value = 2;", false},
		{"descriptor", "Object.defineProperty(target, 'value', { get: () => undefined });", true},
		{"descriptor through result", "Object.defineProperty(copy, 'value', { get: () => undefined });", true},
		{"descriptor after alias write", "const alias = copy; alias.value = 2; Object.defineProperty(alias, 'value', { get: () => undefined });", true},
		{"descriptor method alias", "const define = Object.defineProperty; define(copy, 'value', { get: () => undefined });", true},
		{"prototype", "Object.setPrototypeOf(copy, {});", true},
		{"prototype write", "(copy as any).__proto__ = {};", true},
		{"computed write", "declare const key: string; (copy as any)[key] = {};", true},
		{"delete", "delete (copy as any).value;", true},
		{"unknown call", "unknown(copy);", true},
		{"stored reference", "const holder = { copy }; unknown(holder);", true},
		{"binding write", "target = () => 2;", true},
		{"unaccounted copy", "Object.assign(copy, {});", true},
		{"fresh rest", "const { ...other } = copy; Object.defineProperty(other, 'value', { get: () => undefined });", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			file := tspath.ResolvePath(root.Dir, "property-shape.ts")
			code := "declare function unknown(value: unknown): void; let target = () => 1; const copy = Object.assign(target, { value: 1 }); " + test.operation + " const result = target;"
			fs := NewOverlayVFS(root.FS, map[string]string{file: code})
			program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
			assert.NilError(t, err)
			source := program.GetSourceFile(file)
			assert.Assert(t, source != nil)
			typeChecker, done := program.GetTypeChecker(t.Context())
			defer done()
			result := findVariableInitializer(t, source, "result")
			accounted := map[*ast.Node]bool{findVariableInitializer(t, source, "copy"): true}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(typeChecker, source, nil)
			for range 2 {
				assert.Equal(t, evaluator.HasPropertyShapeMutation(result, accounted), test.changed)
			}
			ordinary := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
			assert.Assert(t, ordinary.HasPropertyShapeMutation(result, accounted))
		})
	}
}
