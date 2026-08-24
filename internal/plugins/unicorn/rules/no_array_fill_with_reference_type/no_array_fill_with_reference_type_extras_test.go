// TestNoArrayFillWithReferenceTypeExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package no_array_fill_with_reference_type_test

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_array_fill_with_reference_type "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_fill_with_reference_type"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func extraError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestNoArrayFillWithReferenceTypeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional-chain receiver / call boundaries ----
			{Code: `((array))?.fill({})`, FileName: "file.js"},
			{Code: `array.fill?.({})`, FileName: "file.js"},
			{Code: `(array?.fill)({})`, FileName: "file.js"},

			// ---- Dimension 4: access / key forms ----
			{Code: `array["fill"]({})`, FileName: "file.js"},
			{Code: "array[`fill`]({})", FileName: "file.js"},
			{Code: `array[0]({})`, FileName: "file.js"},
			{Code: `array[Symbol.iterator]({})`, FileName: "file.js"},
			{Code: `class Collection { #fill(value) {} use() { this.#fill({}); } }`, FileName: "file.js"},

			// ---- Dimension 4: unmatched declaration / container forms ----
			{Code: `const {value} = {value: {}}; array.fill(value)`, FileName: "file.js"},
			{Code: `const [value] = [{}]; array.fill(value)`, FileName: "file.js"},
			{Code: `const value = {}; const alias = value; array.fill(alias)`, FileName: "file.js"},
			{Code: `var value = {}; var value = {}; array.fill(value)`, FileName: "file.js"},
			{Code: `const value = {}; { const value = 0; array.fill(value); }`, FileName: "file.js"},

			// ---- Dimension 4: nesting / traversal boundaries ----
			{Code: `const value = 0; function outer() { const value = {}; function inner(value) { array.fill(value); } }`, FileName: "file.js"},

			// ---- Dimension 4: graceful degradation ----
			{Code: `array.fill()`, FileName: "file.js"},
			{Code: `array.fill(...values)`, FileName: "file.js"},
			{Code: `const value = {...source}; array.fill(source)`, FileName: "file.js"},
			{Code: `const {value, ...rest} = source; array.fill(value)`, FileName: "file.js"},
			{Code: `declare const value: object; array.fill(value)`, FileName: "file.ts", Tsx: false},

			// N/A: the rule does not target functions/classes as containers and does
			// not walk their bodies specially; async/generator/overload forms cannot
			// affect a local fill-call match.
			// N/A: the rule has no autofix or suggestion, so edit boundaries and
			// edit-demand invariance do not apply.

			// Locks in upstream isReferenceExpression() false arms: functions and
			// regular expressions are deliberately not classified as reference fill
			// values by this rule.
			{Code: `array.fill(async () => ({}))`, FileName: "file.js"},
			{Code: `array.fill(function* () {})`, FileName: "file.js"},
			{Code: `array.fill(/reference/)`, FileName: "file.js"},

			// Locks in upstream isGlobalIdentifier(): only the global RegExp
			// constructor is exempt, including through harmless parentheses.
			{Code: `array.fill(new (RegExp)("x"))`, FileName: "file.js"},
			{Code: `interface RegExp { marker: true } array.fill(new RegExp("x"))`, FileName: "file.ts", Tsx: false},

			// Locks in upstream getConstVariableInitializer() guards: non-const,
			// destructured, parameter, and recursively aliased values are not chased.
			{Code: `let value = {}; array.fill(value)`, FileName: "file.js"},
			{Code: `var value = {}; array.fill(value)`, FileName: "file.js"},
			{Code: `function f(value = {}) { array.fill(value); }`, FileName: "file.js"},

			// Locks in upstream isKnownNonArray(): every typed-array constructor
			// form is excluded, and a class subclass of Array stays a known
			// non-array because checkClassHeritage is false.
			{Code: `new Uint8Array(3).fill({})`, FileName: "file.ts", Tsx: false},
			{Code: `const collection = new Set(); collection.fill({})`, FileName: "file.js"},
			{Code: `({fill() {}}).fill({})`, FileName: "file.js"},
			{Code: `(() => {}).fill({})`, FileName: "file.js"},
			{Code: `(class {}).fill({})`, FileName: "file.js"},
			{Code: "`value`.fill({})", FileName: "file.js"},
			{Code: `class Items extends Array<object> {} const items = new Items(); items.fill({});`, FileName: "file.ts", Tsx: false},
			{Code: `function f(foo: Set<object> | Map<object, object>) { foo.fill({}); }`, FileName: "file.ts", Tsx: false},
			// Interface heritage is followed only through an explicit binding
			// annotation. Type information for other receiver shapes must not
			// widen through the interface's base types.
			{Code: `interface Items extends Array<object> {} declare function get(): Items; get().fill({})`, FileName: "file.ts", Tsx: false},
			{Code: `interface Items extends Array<object> {} declare const o: {items: Items}; o.items.fill({})`, FileName: "file.ts", Tsx: false},
			{Code: `namespace N { export interface Items extends Array<object> {} } declare const i: N.Items; i.fill({})`, FileName: "file.ts", Tsx: false},
			// A function declaration's return type is not a type annotation on
			// the function binding itself.
			{Code: `function items(): object[] { return [] } items.fill({})`, FileName: "file.ts", Tsx: false},

			// Uninformative assertions fall through to the wrapped receiver, while
			// local type definitions named like built-ins take precedence over the
			// well-known Array / ReadonlyArray names.
			{Code: `function f(foo: Uint8Array) { (foo as any).fill({}); (foo as unknown).fill({}); }`, FileName: "file.ts", Tsx: false},
			{Code: `function f(foo: Set<object>) { (foo as any).fill({}); (foo as unknown).fill({}); }`, FileName: "file.ts", Tsx: false},
			{Code: `interface Array { fill(value: object): void } function f(foo: Array) { foo.fill({}); }`, FileName: "file.ts", Tsx: false},
			{Code: `interface ReadonlyArray { fill(value: object): void } function f(foo: ReadonlyArray) { foo.fill({}); }`, FileName: "file.ts", Tsx: false},
			{Code: `new ReadonlyArray().fill({})`, FileName: "file.ts", Tsx: false},

			// ---- Real-user: #2657 — dynamic class/function references ----
			// The proposal includes Class/function values. Upstream intentionally
			// limits the rule to direct class expressions, so identifier references
			// to declarations remain unreported.
			{Code: `class Item {} new Array(3).fill(Item)`, FileName: "file.js"},
			{Code: `function createItem() {} new Array(3).fill(createItem)`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver / expression wrappers ----
			{Code: `((array)).fill((({})))`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 18, 1, 20)}},
			{Code: `(array!).fill({})`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 15, 1, 17)}},
			{Code: `(array as object[]).fill({})`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 26, 1, 28)}},
			{Code: `(array satisfies object[]).fill({})`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 33, 1, 35)}},
			{Code: `array.fill(((new Map())))`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 14, 1, 23)}},

			// ---- Dimension 4: nesting / traversal boundaries ----
			{Code: `const value = {}; function outer() { array.fill(value); function inner() { const value = 0; array.fill(value); } }`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 49, 1, 54)}},

			// ---- Dimension 4: graceful degradation with a sibling spread ----
			{Code: `array.fill({}, ...values)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 14)}},

			// Locks in upstream isReferenceExpression() true arms.
			{Code: `array.fill({method() {}, ...source})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 36)}},
			{Code: `array.fill([, ...values])`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 25)}},
			{Code: `array.fill(class Named {})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 26)}},
			{Code: `array.fill(new (class {})())`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 28)}},

			// Locks in upstream isGlobalIdentifier() shadowing branch.
			{Code: `function f(RegExp) { array.fill(new RegExp()) }`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 33, 1, 45)}},
			{Code: `/* global RegExp:off */ array.fill(new RegExp())`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 36, 1, 48)}},
			// A TS assertion around the constructor callee is not transparent to
			// upstream isGlobalIdentifier(), so it remains a reference value.
			{Code: `array.fill(new (RegExp as any)("x"))`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 12, 1, 36)}},

			// Locks in upstream getConstVariableInitializer() accepted path through
			// TS wrappers around both the reference and its initializer.
			{Code: `const value = (({} as Foo)); array.fill((value!))`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 42, 1, 48)}},

			// Nullish normalizes to a non-target and participates in the union
			// vote. This rule reports both target and unknown receivers, as it does
			// for mixed unions.
			{Code: `function f(foo: object[] | undefined) { foo!.fill({}); }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 51, 1, 53)}},
			{Code: `function f(foo: object[] | Set<object>) { foo.fill({}); }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 52, 1, 54)}},
			// The type-information path classifies by symbol name, even when a
			// module-local declaration shadows the standard Array interface.
			{Code: `export {}; interface Array<T> { fill(v: object): void } declare function get(): Array<object>; get().fill({})`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 107, 1, 109)}},
			{Code: `function f(foo: Uint8Array) { (foo as object[]).fill({}); }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{extraError(1, 54, 1, 56)}},
			{Code: `Array.of(1, 2, 3).fill({})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 24, 1, 26)}},
			{Code: `const array = []; array.fill({})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 30, 1, 32)}},

			// ---- Real-user: #2657 — proposed Map/Set and custom class shapes ----
			{Code: `new Array(3).fill(new String("value"))`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 19, 1, 38)}},
			{Code: `class Item {} const item = new Item(); new Array(3).fill(item)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{extraError(1, 58, 1, 62)}},
		},
	)
}

func TestNoArrayFillWithReferenceTypeSourceOnlyReceiverClassification(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		diagnostics int
	}{
		{name: "static primitive", code: `"x".fill({})`},
		{name: "typed array annotation", code: `function f(foo: Uint8Array) { foo.fill({}); }`},
		{name: "keyed collection annotation", code: `function f(foo: Set<object>) { foo.fill({}); }`},
		{name: "type alias annotation", code: `type Bytes = Uint8Array; function f(foo: Bytes) { foo.fill({}); }`},
		{name: "imported type alias annotation", code: `import type { Uint8Array as Bytes } from "./types"; function f(foo: Bytes) { foo.fill({}); }`},
		{name: "recursive const aliases", code: `const a = b; const b = new Uint8Array(3); a.fill({});`},
		{name: "recursive primitive aliases", code: `const a = b; const b = "x"; a.fill({});`},
		{name: "sequence expression", code: `(0, new Uint8Array()).fill({})`},
		{name: "non-array conditional", code: `(flag ? new Set() : new Uint8Array()).fill({})`},
		{name: "array annotation", code: `function f(foo: object[]) { foo.fill({}); }`, diagnostics: 1},
		{name: "recursive array aliases", code: `const a = b; const b = []; a.fill({});`, diagnostics: 1},
		{name: "array conditional", code: `(flag ? [] : new Uint8Array()).fill({})`, diagnostics: 1},
		{name: "unknown receiver", code: `function f(foo) { foo.fill({}); }`, diagnostics: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := lintSourceOnly(t, test.code); len(got) != test.diagnostics {
				t.Fatalf("diagnostic count = %d, want %d: %+v", len(got), test.diagnostics, got)
			}
		})
	}
}

func TestNoArrayFillWithReferenceTypeDoesNotResolveConstAcrossFiles(t *testing.T) {
	t.Parallel()

	dir := tspath.NormalizePath(t.TempDir())
	declarationFile := tspath.NormalizePath(filepath.Join(dir, "declaration.ts"))
	usageFile := tspath.NormalizePath(filepath.Join(dir, "usage.ts"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		declarationFile: `const value = {}`,
		usageFile:       `array.fill(value)`,
	})
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		Target: core.ScriptTargetESNext,
	}, []string{declarationFile, usageFile}, host)
	if err != nil {
		t.Fatalf("create typed program: %v", err)
	}

	diagnostics := make([]rule.RuleDiagnostic, 0, 1)
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(program),
		File:        usageFile,
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule.Run(ctx, nil)
				},
			}}
		},
		ExcludePaths: []string{},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostic count = %d, want 0: %+v", len(diagnostics), diagnostics)
	}
}

func lintSourceOnly(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()

	dir := tspath.NormalizePath(t.TempDir())
	fileName := tspath.NormalizePath(filepath.Join(dir, "file.ts"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(dir, fs),
		CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("create source-only program: %v", err)
	}

	diagnostics := make([]rule.RuleDiagnostic, 0, 1)
	_, err = linter.RunLinter(linter.RunLinterOptions{
		Programs:       []*lintprogram.Program{sourceProgram},
		SingleThreaded: true,
		Scope:          linter.FileScope{Files: []string{fileName}},
		ExcludePaths:   []string{},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("source-only fixture unexpectedly received a TypeChecker")
					}
					return no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	if err != nil {
		t.Fatalf("lint source-only program: %v", err)
	}
	return diagnostics
}
