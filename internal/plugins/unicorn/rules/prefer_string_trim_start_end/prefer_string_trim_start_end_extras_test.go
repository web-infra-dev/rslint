// TestPreferStringTrimStartEndExtras locks in branches and tsgo edge shapes
// that the upstream test suite does not exercise.
package prefer_string_trim_start_end_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_trim_start_end"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferStringTrimStartEndExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_string_trim_start_end.PreferStringTrimStartEndRule,
		[]rule_tester.ValidTestCase{
			// ---- Real-user: upstream PR #1768 optional-call direction ----
			// Optional calls are deliberately allowed.
			{Code: `foo.trimRight?.()`, FileName: "file.js"},

			// ---- Dimension 4: receiver and callee wrappers ----
			{Code: `(foo.trimLeft)?.()`, FileName: "file.js"},

			// ---- Dimension 4: access/key forms ----
			// Only non-computed identifier properties are matched.
			{Code: "foo[`trimLeft`]()", FileName: "file.js"},
			{Code: `foo[0]()`, FileName: "file.js"},
			{Code: `foo[Symbol.iterator]()`, FileName: "file.js"},
			{Code: `class C { #trimLeft() {} method() { this.#trimLeft(); } }`, FileName: "file.js"},

			// ---- Dimension 4: graceful degradation ----
			{Code: `const object: Record<string, unknown> = {}; const { ...rest } = object; rest.trimRight()`, FileName: "file.ts"},

			// Locks in the known-non-string union branch.
			{Code: `function f(foo: number[] | Set<number>) { foo.trimRight(); }`, FileName: "file.ts"},
			// Locks in the authored TypeScript assertion path.
			{Code: `(foo as number[]).trimLeft()`, FileName: "file.ts"},

			// N/A Dimension 4 declaration/container forms: the rule targets call
			// expressions, not function or class declarations.
			// N/A Dimension 4 body-absent forms: overloads, abstract members, and
			// declarations cannot themselves be trim method calls.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver and expression wrappers ----
			// Parenthesized callees are transparent in ESTree and must remain so in tsgo.
			trimInvalid(`(foo.trimLeft)()`, "file.js", "trimLeft", "trimStart", 0),
			trimInvalid(`((foo)).trimRight()`, "file.js", "trimRight", "trimEnd", 0),
			trimInvalid(`foo!.trimLeft()`, "file.ts", "trimLeft", "trimStart", 0),
			// A known string assertion stays a target; satisfies classifies its expression.
			trimInvalid(`(foo as string).trimLeft()`, "file.ts", "trimLeft", "trimStart", 0),
			trimInvalid(`(foo satisfies string).trimRight()`, "file.ts", "trimRight", "trimEnd", 0),
			// Mixed target/non-target unions are unknown, so the rule reports.
			trimInvalid(`function f(foo: string | number) { foo.trimRight(); }`, "file.ts", "trimRight", "trimEnd", 0),

			// ---- Dimension 4: nesting/traversal boundary ----
			// Locks in upstream isMethodCall() argument-count arm: the outer call
			// is ignored while the nested zero-argument call is still reported.
			trimInvalid(`foo.trimLeft(foo.trimRight())`, "file.js", "trimRight", "trimEnd", 0),

			// ---- Dimension 4: graceful degradation ----
			// An object spread is not statically classified in a source-only file,
			// but it must remain safe to inspect and fix.
			trimInvalid(`({ ...foo }).trimLeft()`, "file.js", "trimLeft", "trimStart", 0),

			// ---- Real-user: upstream PR #1768 optional-member direction ----
			// Optional access reports even though optional calls do not.
			trimInvalid(`foo?.trimRight()`, "file.js", "trimRight", "trimEnd", 0),

			// ---- Dimension 3: autofix comment boundary ----
			trimInvalid(`foo./* before */trimLeft/* after */()`, "file.js", "trimLeft", "trimStart", 0),

			// ---- Regression: symbol-less checker types remain unknown ----
			// Unicorn reports literal primitives and symbol-less object types
			// because the checker cannot prove a named non-string type for them.
			trimInvalid(`(1).trimLeft()`, "file.ts", "trimLeft", "trimStart", 0),
			trimInvalid(`const value = 1; value.trimRight()`, "file.ts", "trimRight", "trimEnd", 0),
			trimInvalid(`([1, 2] as const).trimLeft()`, "file.ts", "trimLeft", "trimStart", 0),
		},
	)
}

func TestPreferStringTrimStartEndEditDemand(t *testing.T) {
	t.Parallel()

	const source = `foo.trimLeft()`
	const fixedSource = `foo.trimStart()`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.js",
		Path:     "/edit-demand.js",
	}, source, core.ScriptKindJS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()

		comments := rule.NewCommentStore(sourceFile)
		diagnostics := make([]rule.RuleDiagnostic, 0, 1)
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(
			prefer_string_trim_start_end.PreferStringTrimStartEndRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)

		listeners := prefer_string_trim_start_end.PreferStringTrimStartEndRule.Run(ctx, nil)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if listener := listeners[node.Kind]; listener != nil {
				listener(node)
			}
			return node.ForEachChild(visit)
		}
		sourceFile.AsNode().ForEachChild(visit)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		withoutEdits := diagnostic
		withoutEdits.FixesPtr = nil
		withoutEdits.Suggestions = nil
		want := base
		want.FixesPtr = nil
		want.Suggestions = nil
		if !reflect.DeepEqual(withoutEdits, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, withoutEdits, want)
		}
		if diagnostic.Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions", demand)
		}
	}

	if diagnostics[rule.EditDemandNone].FixesPtr != nil {
		t.Errorf("none demand unexpectedly materialized fixes")
	}
	if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
		t.Errorf("suggestion-only demand unexpectedly materialized fixes")
	}
	autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
	allFixes := diagnostics[rule.EditDemandAll].FixesPtr
	if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
		t.Fatalf("autofix artifacts differ between autofix-only and all demand")
	}
	output, unapplied, fixed := linter.ApplyRuleFixes(source, []rule.RuleDiagnostic{diagnostics[rule.EditDemandAll]})
	if !fixed || len(unapplied) != 0 || output != fixedSource {
		t.Fatalf("autofix result = %q, unapplied = %d, fixed = %v; want %q", output, len(unapplied), fixed, fixedSource)
	}
}

func TestPreferStringTrimStartEndSourceOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want int
	}{
		{
			name: "parameter array annotation",
			code: `function f(foo: number[]) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "authored non-string assertion",
			code: `(foo as number[]).trimRight();`,
			want: 0,
		},
		{
			name: "local type alias",
			code: `type Numbers = number[]; function f(foo: Numbers) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "local interface",
			code: `interface Collection {} function f(foo: Collection) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "local interface heritage",
			code: `interface Base {} interface Collection extends Base {} function f(foo: Collection) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "local class",
			code: `class Collection {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "local class heritage",
			code: `class Base {} class Collection extends Base {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "const alias in class heritage",
			code: `class Base {} const Alias = Base; class Collection extends Alias {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "const class expression in heritage",
			code: `const Alias = class {}; class Collection extends Alias {} function f(foo: Collection) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "direct class expression heritage remains unknown",
			code: `class Collection extends class {} {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 1,
		},
		{
			name: "using binding is not a const alias",
			code: `using Alias = class {}; class Collection extends Alias {} function f(foo: Collection) { foo.trimRight(); }`,
			want: 1,
		},
		{
			name: "await using binding is not a const alias",
			code: `await using Alias = class {}; class Collection extends Alias {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 1,
		},
		{
			name: "constrained type parameter",
			code: `function f<T extends number[]>(foo: T) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "readonly annotation",
			code: `function f(foo: readonly (number | boolean)[]) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "parenthesized annotation",
			code: `function f(foo: (number[])) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "non-string literal annotation",
			code: `function f(foo: 1) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "all non-string union",
			code: `function f(foo: number | boolean) { foo.trimLeft(); }`,
			want: 0,
		},
		{
			name: "all non-string intersection",
			code: `function f(foo: number[] & {brand: true}) { foo.trimRight(); }`,
			want: 0,
		},
		{
			name: "const initializer assertion",
			code: `const foo = value as number[]; foo.trimLeft();`,
			want: 0,
		},
		{
			name: "string annotation remains target",
			code: `function f(foo: string) { foo.trimLeft(); }`,
			want: 1,
		},
		{
			name: "string literal annotation remains target",
			code: `function f(foo: "value") { foo.trimRight(); }`,
			want: 1,
		},
		{
			name: "mixed union remains unknown",
			code: `function f(foo: string | number) { foo.trimRight(); }`,
			want: 1,
		},
		{
			name: "unresolved type reference remains unknown",
			code: `function f(foo: ExternalCollection) { foo.trimLeft(); }`,
			want: 1,
		},
		{
			name: "unresolved interface heritage remains unknown",
			code: `interface Collection extends ExternalCollection {} function f(foo: Collection) { foo.trimRight(); }`,
			want: 1,
		},
		{
			name: "unresolved class heritage remains unknown",
			code: `class Collection extends ExternalCollection {} function f(foo: Collection) { foo.trimLeft(); }`,
			want: 1,
		},
		{
			name: "any assertion falls through",
			code: `(foo as any).trimLeft();`,
			want: 1,
		},
		{
			name: "satisfies annotation is ignored",
			code: `(foo satisfies number[]).trimRight();`,
			want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			diagnostics := lintPreferStringTrimStartEndSourceOnly(t, test.code)
			if len(diagnostics) != test.want {
				t.Fatalf("diagnostics = %d, want %d: %+v", len(diagnostics), test.want, diagnostics)
			}
		})
	}
}

func lintPreferStringTrimStartEndSourceOnly(t *testing.T, code string) []rule.RuleDiagnostic {
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
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     prefer_string_trim_start_end.PreferStringTrimStartEndRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("source-only fixture unexpectedly received a TypeChecker")
					}
					return prefer_string_trim_start_end.PreferStringTrimStartEndRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("prepare source-only lint plan: %v", err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	if err != nil {
		t.Fatalf("lint source-only program: %v", err)
	}
	return diagnostics
}
