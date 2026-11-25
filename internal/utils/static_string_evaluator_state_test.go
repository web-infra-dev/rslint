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
				assert.Equal(t, evaluator.HasPropertyShapeMutation(result, nil, nil, accounted), test.changed)
			}
			ordinary := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
			assert.Assert(t, ordinary.HasPropertyShapeMutation(result, nil, nil, accounted))
		})
	}
}

func TestStaticPropertyReadShapeMutations(t *testing.T) {
	for _, test := range []struct {
		name, before, after, read string
		changed                   bool
	}{
		{"plain", "", "", "copy.value", false},
		{"keys", "Object.keys(copy);", "", "copy.value", false},
		{"keys alias", "const keys = Object.keys; keys(copy);", "", "copy.value", false},
		{"repeated keys alias", "const keys = Object.keys; keys(copy); keys(copy);", "", "copy.value", false},
		{"own property check", "Object.hasOwn(copy, 'value');", "", "copy.value", false},
		{"own property check alias", "const own = Object.hasOwn; own(copy, 'value');", "", "copy.value", false},
		{"own property descriptor", "Object.getOwnPropertyDescriptor(copy, 'value');", "", "copy.value", false},
		{"primitive property key", "const key: string = String(Math.random()); Object.hasOwn(copy, key);", "", "copy.value", false},
		{"property key coercion", "const key = {toString() { unknown(copy); return 'value'; }}; Object.hasOwn(copy, key as unknown as string);", "", "copy.value", true},
		{"empty copy", "Object.assign(copy, {});", "", "copy.value", false},
		{"other descriptor", "Object.defineProperty(copy, 'tag', {value: 2});", "", "copy.value", false},
		{"descriptor getter", "Object.defineProperty(copy, 'tag', {get value() { unknown(copy); return 2; }});", "", "copy.value", true},
		{"earlier descriptor", "Object.defineProperty(copy, 'value', {value: undefined});", "", "copy.value", true},
		{"later descriptor", "", "Object.defineProperty(copy, 'value', {value: undefined});", "copy.value", false},
		{"earlier escape", "unknown(copy);", "", "copy.value", true},
		{"later escape", "", "unknown(copy);", "copy.value", false},
		{"closure", "function mutate() { unknown(copy); } mutate();", "", "copy.value", true},
		{"loop", "for (let i = 0; i < 2; i++) {", "unknown(copy); }", "copy.value", true},
		{"array push", "copy.push(1);", "", "copy.value", false},
		{"array method write", "target.push = () => 0; copy.push(1);", "", "copy.value", true},
		{"array method descriptor", "Object.defineProperty(target, 'push', {value: () => 0}); copy.push(1);", "", "copy.value", true},
		{"changed builtin", "Object.keys = () => []; Object.keys(copy);", "", "copy.value", true},
		{"builtin descriptor", "Object.defineProperty(Object, 'keys', {value: () => []}); Object.keys(copy);", "", "copy.value", true},
		{"object prototype", "Object.defineProperty(Object.prototype, 'value', {get: () => undefined});", "", "copy.value", true},
		{"array prototype", "Object.defineProperty(Array.prototype, 'value', {get: () => undefined});", "", "copy.value", true},
		{"constructor tag", "(function tag(...values: unknown[]) {})`value ${Object}`;", "", "copy.value", true},
		{"length write", "copy.length = 0;", "", "copy[0]", true},
		{"length alias", "const alias = copy; alias.length = 0;", "", "copy[0]", true},
		{"length reassigned alias", "let alias: number[] = []; alias = copy; alias.length = 0;", "", "copy[0]", true},
		{"length stored alias", "const holder = {items: [] as number[]}; holder.items = copy; holder.items.length = 0;", "", "copy[0]", true},
		{"length copied alias", "const holder = {items: [] as number[]}; Object.assign(holder, {items: copy}); holder.items.length = 0;", "", "copy[0]", true},
		{"length getter alias", "const holder = {get items() { return copy; }}; holder.items.length = 0;", "", "copy[0]", true},
		{"length descriptor alias", "const holder = {items: [] as number[]}; Object.defineProperty(holder, 'items', {get() { return copy; }}); holder.items.length = 0;", "", "copy[0]", true},
		{"later length write", "", "copy.length = 0;", "copy[0]", false},
		{"length write keeps named properties", "copy.length = 0;", "", "copy.value", false},
		{"length write keeps length", "copy.length = 0;", "", "copy.length", false},
		{"length write keeps non-index", "copy.length = 0;", "", "copy['4294967295']", false},
		{"length write removes large index", "copy.length = 0;", "", "copy['4294967294']", true},
		{"length write in earlier static block", "class C {static {copy.length = 0;} static {", "}}", "copy[0]", true},
		{"length write in later static block", "class C {static {", "} static {copy.length = 0;}}", "copy[0]", false},
		{"restored index", "copy.length = 0; copy[0] = 2;", "", "copy[0]", false},
		{"restored before truncation", "copy[0] = 2; copy.length = 0;", "", "copy[0]", true},
		{"conditionally restored index", "copy.length = 0; if (Math.random()) { copy[0] = 2; }", "", "copy[0]", true},
		{"restored other index", "copy.length = 0; copy[1] = 2;", "", "copy[0]", true},
		{"computed alias before static block", "let alias: number[] = []; class C { static { Object.defineProperty(alias, 'value', {value: undefined}); } [(alias = copy, 'method')]() {} }", "", "copy.value", true},
		{"computed store before static block", "const holder = {items: [] as number[]}; class C { static { Object.defineProperty(holder.items, 'value', {value: undefined}); } [(Object.assign(holder, {items: copy}), 'method')]() {} }", "", "copy.value", true},
		{"closure restores binding", "let alias: object = copy; function restore() { alias = copy; } alias = {}; restore(); Object.defineProperty(alias, 'value', {value: undefined});", "", "copy.value", true},
		{"captured mutable alias", "let alias: object = copy; const saved = alias; alias = {}; Object.defineProperty(saved, 'value', {value: undefined});", "", "copy.value", true},
		{"multiple alias reads in one call", "let alias: object = copy; const saved = alias; alias = {}; unknown([alias, saved]);", "", "copy.value", true},
		{"default assignment alias", "let alias: object = {}; ({alias = copy} = {} as {alias?: object}); Object.defineProperty(alias, 'value', {value: undefined});", "", "copy.value", true},
		{"array default assignment alias", "let alias: object = {}; [alias = copy] = []; Object.defineProperty(alias, 'value', {value: undefined});", "", "copy.value", true},
		{"iteration binding alias", "for (const alias of [copy]) { Object.defineProperty(alias, 'value', {value: undefined}); }", "", "copy.value", true},
		{"delete through assigned alias", "let alias: {value?: number} = {}; alias = copy; delete alias.value;", "", "copy.value", true},
		{"computed write through assigned alias", "let alias: Record<string, unknown> = {}; alias = copy; const key: string = 'value'; alias[key] = undefined;", "", "copy.value", true},
		{"captured length alias", "let alias: number[] = copy; const saved = alias; alias = []; saved.length = 0;", "", "copy[0]", true},
		{"restored between truncations", "copy.length = 0; copy[0] = 2;", "copy.length = 0;", "copy[0]", false},
		{"deferred truncation after restoration", "copy.length = 0; copy[0] = 2; const holder = {get value() { copy.length = 0; return 0; }}; holder.value;", "", "copy[0]", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			file := tspath.ResolvePath(root.Dir, "property-read.ts")
			code := "declare function unknown(value: unknown): void; function read() { const target: number[] = []; const copy = Object.assign(target, {0: 1, value: 1}); " + test.before + " const result = " + test.read + "; " + test.after + " }"
			fs := NewOverlayVFS(root.FS, map[string]string{file: code})
			program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
			assert.NilError(t, err)
			source := program.GetSourceFile(file)
			assert.Assert(t, source != nil)
			typeChecker, done := program.GetTypeChecker(t.Context())
			defer done()
			access := findVariableInitializer(t, source, "result")
			copyCall := findVariableInitializer(t, source, "copy")
			origin := findVariableInitializer(t, source, "target")
			accounted := map[*ast.Node]bool{copyCall: true}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(typeChecker, source, nil)
			for range 2 {
				assert.Equal(t, evaluator.HasPropertyShapeMutation(copyCall.Arguments()[0], origin, access, accounted), test.changed)
			}
		})
	}
}
