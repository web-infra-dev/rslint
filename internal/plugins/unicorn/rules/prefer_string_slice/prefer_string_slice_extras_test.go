package prefer_string_slice_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_slice"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferStringSliceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_string_slice.PreferStringSliceRule,
		[]rule_tester.ValidTestCase{
			// Call and property shapes.
			{Code: "new foo.substr(1);", FileName: "file.js"},
			{Code: "foo[\"substring\"](1);", FileName: "file.js"},
			{Code: "foo[substring](1);", FileName: "file.js"},
			{Code: "class C { #substr() {} method() { this.#substr(1); } }", FileName: "file.js"},
			{Code: "(foo?.substring)(1);", FileName: "file.js"},
			{Code: "(foo.substring as Function)(1);", FileName: "file.ts"},
			// Source annotations and unknown receivers.
			{Code: "function f(value: number[]) { value.substring(1); }", FileName: "file.ts"},
			// Single-character reference comparisons.
			{Code: "foo.substring(object[\"in\" + \"dex\"], object.index + 1);", FileName: "file.js"},
			{Code: "foo.substring(object[index], object[index] + 1);", FileName: "file.js"},
			{Code: "class C { #index = 0; method() { foo.substring(this.#index, this.#index + 1); } }", FileName: "file.js"},
			{Code: "foo.substring(`abc`.length, \"abc\".length + 1);", FileName: "file.js"},
			{Code: "foo.substring((index as number), index + 1);", FileName: "file.ts"},
			{Code: "foo.substring(index<number>, index + 1);", FileName: "file.ts"},
			// Additional TypeScript and comment boundaries.
			{Code: "text.substring(obj[\"index\" as string], obj.index + 1);", FileName: "file.ts"},
			{Code: "text.substring(obj[index as number], obj[index] + 1);", FileName: "file.ts"},
			{Code: "text.substring(obj[\"index\" satisfies string], obj.index + 1);", FileName: "file.ts"},
			{Code: "text.substring((fn<number>).index, fn.index + 1);", FileName: "file.ts"},
			{Code: "class Collection { substr(start: number) { return String(start); } }\nconst read = (): Collection => new Collection();\nread().substr(1);", FileName: "file.ts"},
			// Available type information excludes inferred non-string receivers.
			{Code: "const value = [1, 2]; value.substr(1);", FileName: "file.js"},
			{Code: "const value = {substring() {}}; value.substring(1);", FileName: "file.js"},
			{Code: "class Collection { substr(start: number) { return String(start); } }\nconst read = () => new Collection();\nread().substr(1);", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			// Call and property shapes.
			{Code: "((foo).substring)((-2));", FileName: "file.js", Output: []string{"((foo).slice)(0);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "(foo.substr)(start);", FileName: "file.js", Output: []string{"(foo.slice)(start);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			{Code: "(/** @type {string} */ (foo)).substring(3);", FileName: "file.js", Output: []string{"(/** @type {string} */ (foo)).slice(3);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
			}},
			{Code: "const element = <div>{text.substring(3)}</div>;", FileName: "file.tsx", Output: []string{"const element = <div>{text.slice(3)}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 40},
			}},
			// Source annotations and unknown receivers.
			{Code: "function f(value: string | number[]) { value.substring(1); }", FileName: "file.ts", Output: []string{"function f(value: string | number[]) { value.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 58},
			}},
			{Code: "(value as string).substring(3);", FileName: "file.ts", Output: []string{"(value as string).slice(3);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "value!.substr(1);", FileName: "file.ts", Output: []string{"value!.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
			}},
			{Code: "(value satisfies string).substring(1);", FileName: "file.ts", Output: []string{"(value satisfies string).slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			}},

			// Single-character reference comparisons.
			{Code: "foo.substring(2 ** 53, 2 ** 53 + 1);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
			}},
			{Code: "foo.substring(-1, 0);", FileName: "file.js", Output: []string{"foo.slice(0, 0);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo.substring(index - 2, index);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "foo.substring(object.left, object.right + 1);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			}},
			{Code: "foo.substring(object[index], object[other] + 1);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
			}},
			{Code: "foo.substring(object.index, other.index + 1);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			}},
			{Code: "foo.substring(getIndex(), getIndex() + 1);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
			}},
			// Literal values and length properties.
			{Code: "const start = 0; foo.substr(start, length);", FileName: "file.js", Output: []string{"const start = 0; foo.slice(start, Math.max(0, length === undefined ? Infinity : length));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 43},
			}},
			{Code: "let start = 0; start = 2; foo.substr(start, length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 52},
			}},
			{Code: "foo.substr(-1, 2);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			}},
			{Code: "foo.substr(1_000, 0x10);", FileName: "file.js", Output: []string{"foo.slice(1_000, 1016);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substr(0, foo.length - 0x2);", FileName: "file.js", Output: []string{"foo.slice(0, -0x2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "foo.substr(0, (foo.length) - 2);", FileName: "file.js", Output: []string{"foo.slice(0, -2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "foo.substr(0, (foo).length - 2);", FileName: "file.js", Output: []string{"foo.slice(0, -2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "foo.substr(0, foo.length - 1_000);", FileName: "file.js", Output: []string{"foo.slice(0, -1_000);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - 9007199254740992);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "(foo as string).substr(0, foo.length - 2);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
			}},
			{Code: "\"ab\\u{63}\".substr(0, `abc`.length - 2);", FileName: "file.js", Output: []string{"\"ab\\u{63}\".slice(0, -2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
			}},
			{Code: "foo.substring(foo?.length);", FileName: "file.js", Output: []string{"foo.slice(Math.max(0, foo?.length));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo?.length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "foo.substring((--index));", FileName: "file.js", Output: []string{"foo.slice(Math.max(0, (--index)));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.substring(0, 1e400);", FileName: "file.js", Output: []string{"foo.slice(0, 1e400);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			// Comments and duplicated evaluation.
			{Code: "foo.substring((/* keep */ start));", FileName: "file.js", Output: []string{"foo.slice(Math.max(0, (/* keep */ start)));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "foo.substring(0, (/* keep */ end));", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, /* keep */ (foo.length - 2));", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
			}},
			{Code: "foo.substring(0, \"/* text */\");", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, \"/* text */\" === undefined ? Infinity : \"/* text */\"));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "foo.substring(0, String.raw`end`);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "foo.substring(0, ({...ends}).end);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "foo.substring(0, () => tag`end`);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			}},
			{Code: "foo.substr(foo.substring(3));", FileName: "file.js", Output: []string{"foo.slice(foo.slice(3));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 28},
			}},
			{Code: "const café = \"😀\"; café.\n  substring(3);", FileName: "file.js", Output: []string{"const café = \"😀\"; café.\n  slice(3);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 20, EndLine: 2, EndColumn: 15},
			}},
			{Code: "const text = \"😀\"; text.substr(0, 3);", FileName: "file.js", Output: []string{"const text = \"😀\"; text.slice(0, 3);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 37},
			}},
			// Static evaluation uses the file's configured globals.
			{Code: "foo.substr(Math.PI - Math.PI, length);", FileName: "file.js", Output: []string{"foo.slice(Math.PI - Math.PI, Math.max(0, length === undefined ? Infinity : length));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			}},
			{Code: "foo.substr(Math.PI - Math.PI, length);", FileName: "file.js", Globals: map[string]any{"Math": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			}},
			// Additional TypeScript and comment boundaries.
			{Code: "text.substring(0, <T>value);", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, <T>value === undefined ? Infinity : <T>value));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			{Code: "text.substring(0, (<T>value));", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, (<T>value) === undefined ? Infinity : (<T>value)));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "text.substring(0, value!);", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, value! === undefined ? Infinity : value!));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "text.substring(0, fn<number>);", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, fn<number> === undefined ? Infinity : fn<number>));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "text.substring(0, value as number);", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, (value as number) === undefined ? Infinity : (value as number)));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			}},
			{Code: "text.substring(0, value satisfies number);", FileName: "file.ts", Output: []string{"text.slice(0, Math.max(0, (value satisfies number) === undefined ? Infinity : (value satisfies number)));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "text.substring(0, a << b);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "text.substring(5 /* keep */, 2);", FileName: "file.js", Output: []string{"text.slice(2 /* keep */, 5);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "text.substring((5 /* keep */), 2);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
		})
}

func TestPreferStringSliceImportedReceivers(t *testing.T) {
	directory := txtarfs.MustParseFile(t, "testdata/receivers.txtar").Materialize(t, "typed")
	root := rule_tester.Root{Dir: tspath.NormalizePath(directory), FS: bundled.WrapFS(osvfs.FS())}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_string_slice.PreferStringSliceRule,
		[]rule_tester.ValidTestCase{
			{Code: "import {read} from './dep';\nread().substr();", FileName: "caller.ts"},
			{Code: "import {read} from './dep';\nread?.().substr();", FileName: "caller.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "import {readString as read} from './dep';\nread().substr();", FileName: "caller.ts", Output: []string{"import {readString as read} from './dep';\nread().slice();"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
			}},
		})
}

func lintPreferStringSliceSourceOnly(t *testing.T, code, name string) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "review/"+name)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs: []*lintprogram.Program{program}, TargetsByProgram: [][]string{{fileName}}, SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: prefer_string_slice.PreferStringSliceRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("source-only test received type information")
					}
					return prefer_string_slice.PreferStringSliceRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{LintPlan: plan, SingleThreaded: true,
		Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
	})
	if err != nil {
		t.Fatal(err)
	}
	return diagnostics
}

func TestPreferStringSliceArtifactsFollowDemand(t *testing.T) {
	for _, tc := range []struct{ source, output string }{
		{"text.substr(2, 3);", "text.slice(2, 5);"},
		{"text.substring(0, end);", "text.slice(0, Math.max(0, end === undefined ? Infinity : end));"},
		{"text.substring(0, getEnd());", "text.substring(0, getEnd());"},
		{"text.substring(0, options.end);", "text.substring(0, options.end);"},
		{"function f(Math, start) { return text.substring(start); }", "function f(Math, start) { return text.substring(start); }"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(tc.source, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: prefer_string_slice.PreferStringSliceRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return prefer_string_slice.PreferStringSliceRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity || diagnostic.RuleName != baseline.RuleName {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := tc.source != tc.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d produced suggestions", demand)
				}
				wantOutput := tc.source
				if wantFix {
					wantOutput = tc.output
				}
				output, _, fixed := linter.ApplyRuleFixes(tc.source, []rule.RuleDiagnostic{diagnostic})
				if output != wantOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
			}
		})
	}
}

func TestPreferStringSliceReviewSourceOnly(t *testing.T) {
	tests := []struct {
		Code, FileName, Output string
		Errors                 []rule_tester.InvalidTestCaseError
	}{
		{Code: "/** @type {{substring: function(): string}} */\nconst o = {substring() {return \"custom\";}};\no.substring();", FileName: "file.js", Output: "/** @type {{substring: function(): string}} */\nconst o = {substring() {return \"custom\";}};\no.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {string} */\nconst o = \"abc\";\no.substring();", FileName: "file.js", Output: "/** @type {string} */\nconst o = \"abc\";\no.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14}}},
		{Code: "/** @param {{substring: function(): string}} o */\nfunction f(o) { o.substring(); }", FileName: "file.js", Output: "/** @param {{substring: function(): string}} o */\nfunction f(o) { o.substring(); }", Errors: []rule_tester.InvalidTestCaseError{}},
		// An explicit array annotation excludes string receivers even without a project.
		{Code: "/** @type {string[]} */\nconst o = [];\no.substring();", FileName: "file.js", Output: "/** @type {string[]} */\nconst o = [];\no.substring();", Errors: nil},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.().substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.().substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "const read = (): string => \"abc\";\nread?.().substr();", FileName: "file.ts", Output: "const read = (): string => \"abc\";\nread?.().slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 18}}},
		{Code: "class Collection { substr() {return \"custom\";} }\nconst read = (): Collection => new Collection();\nread?.().substr();", FileName: "file.ts", Output: "class Collection { substr() {return \"custom\";} }\nconst read = (): Collection => new Collection();\nread?.().substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @returns {{substr: function(): string}} */\nfunction read() {return {substr() {return \"custom\";}};}\nread?.().substr();", FileName: "file.js", Output: "/** @returns {{substr: function(): string}} */\nfunction read() {return {substr() {return \"custom\";}};}\nread?.().substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @returns {string} */\nfunction read() {return \"abc\";}\nread?.().substr();", FileName: "file.js", Output: "/** @returns {string} */\nfunction read() {return \"abc\";}\nread?.().slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 18}}},
		{Code: "/** @type {{substring(): string}} */\nconst value = item as string;\nvalue.substring();", FileName: "file.ts", Output: "/** @type {{substring(): string}} */\nconst value = item as string;\nvalue.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 18}}},
		{Code: "let n = 0; const o = {get end() {return ++n;}}; [\"abc\".substring(0, o.end), n];", FileName: "file.js", Output: "let n = 0; const o = {get end() {return ++n;}}; [\"abc\".substring(0, o.end), n];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 50, EndLine: 1, EndColumn: 75}}},
		{Code: "let n = 0; const o = {get end() {return ++n;}}; [\"abc\".substr(0, o.end), n];", FileName: "file.js", Output: "let n = 0; const o = {get end() {return ++n;}}; [\"abc\".substr(0, o.end), n];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 50, EndLine: 1, EndColumn: 72}}},
		{Code: "let n = 0; const value = {valueOf() {return ++n;}}; [\"abc\".substring(0, +value), n];", FileName: "file.js", Output: "let n = 0; const value = {valueOf() {return ++n;}}; [\"abc\".substring(0, +value), n];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 80}}},
		{Code: "let n = 0; const value = {valueOf() {return ++n;}}; [\"abc\".substr(0, value + 0), n];", FileName: "file.js", Output: "let n = 0; const value = {valueOf() {return ++n;}}; [\"abc\".substr(0, value + 0), n];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 80}}},
		{Code: "function f(Math, n) {return \"abc\".substring(n);} f(null, 1);", FileName: "file.js", Output: "function f(Math, n) {return \"abc\".substring(n);} f(null, 1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 47}}},
		{Code: "function f(undefined, n) {return \"abc\".substring(0, n);} f(2, 2);", FileName: "file.js", Output: "function f(undefined, n) {return \"abc\".substring(0, n);} f(2, 2);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 55}}},
		{Code: "function f(Infinity, n) {return \"abc\".substr(0, n);} f(0);", FileName: "file.js", Output: "function f(Infinity, n) {return \"abc\".substr(0, n);} f(0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 51}}},
		{Code: "function f(Infinity) {return text.substr(1, 1e400);}", FileName: "file.js", Output: "function f(Infinity) {return text.substr(1, 1e400);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 51}}},
		{Code: "function f(Infinity) {return text.substring(1e400);}", FileName: "file.js", Output: "function f(Infinity) {return text.substring(1e400);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 51}}},
		{Code: "function f(Infinity) {return text.substring(1, 1e400);}", FileName: "file.js", Output: "function f(Infinity) {return text.slice(1, 1e400);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 54}}},
		{Code: "function f(Math, undefined, Infinity) {return text.substr(1, 2);}", FileName: "file.js", Output: "function f(Math, undefined, Infinity) {return text.slice(1, 3);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 64}}},
		{Code: "function f(Math, undefined, Infinity) {return text.substring();}", FileName: "file.js", Output: "function f(Math, undefined, Infinity) {return text.slice();}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 63}}},
		{Code: "function f(Math, undefined, Infinity) {return text.substr(0, text.length - 2);}", FileName: "file.js", Output: "function f(Math, undefined, Infinity) {return text.slice(0, -2);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 78}}},
		{Code: "function f(Math, n) {return text.substring(n, 0);}", FileName: "file.js", Output: "function f(Math, n) {return text.substring(n, 0);}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 49}}},
		{Code: "function f(n) {return text.substring(n, 0);}", FileName: "file.js", Output: "function f(n) {return text.slice(0, Math.max(0, n));}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 43}}},
		{Code: "function f(n) {return text.substring(0, n);}", FileName: "file.js", Output: "function f(n) {return text.slice(0, Math.max(0, n === undefined ? Infinity : n));}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 43}}},
		{Code: "text.substr(1, 1e400);", FileName: "file.js", Output: "text.slice(1, Infinity);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}},
		{Code: "text.substring(0, (Math) => Math);", FileName: "file.js", Output: "text.slice(0, Math.max(0, ((Math) => Math) === undefined ? Infinity : ((Math) => Math)));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34}}},
		// Retained JSDoc types and transparent non-null wrappers.
		{Code: "/** @type {{substring: function(): string} & {extra: number}} */\nconst o = {substring() {return \"custom\";}, extra: 1};\no.substring();", FileName: "file.js", Output: "/** @type {{substring: function(): string} & {extra: number}} */\nconst o = {substring() {return \"custom\";}, extra: 1};\no.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {function(): void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", FileName: "file.js", Output: "/** @type {function(): void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {{substring: function(): string} | {substring: function(): string, extra: number}} */\nconst o = {substring() {return \"custom\";}};\no.substring();", FileName: "file.js", Output: "/** @type {{substring: function(): string} | {substring: function(): string, extra: number}} */\nconst o = {substring() {return \"custom\";}};\no.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {() => void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", FileName: "file.js", Output: "/** @type {() => void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {string & {extra: number}} */\nconst o = \"abc\";\no.substring();", FileName: "file.js", Output: "/** @type {string & {extra: number}} */\nconst o = \"abc\";\no.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14}}},
		{Code: "/** @type {string | {substring: function(): string}} */\nconst o = \"abc\";\no.substring();", FileName: "file.js", Output: "/** @type {string | {substring: function(): string}} */\nconst o = \"abc\";\no.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14}}},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.()!.substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.()!.substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.()!!.substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\nread?.()!!.substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "const read = (): string => \"abc\";\nread?.()!.substr();", FileName: "file.ts", Output: "const read = (): string => \"abc\";\nread?.()!.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 19}}},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.()! as string).substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.()! as string).slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 31}}},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.() as string)!.substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.() as string)!.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 31}}},
		{Code: "const read = (): string => \"abc\";\n(read?.() as {substr(): string})!.substr();", FileName: "file.ts", Output: "const read = (): string => \"abc\";\n(read?.() as {substr(): string})!.substr();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.()! as string satisfies string)!.substr();", FileName: "file.ts", Output: "const read = (): {substr(): string} => ({substr() {return \"custom\";}});\n(read?.()! as string satisfies string)!.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 49}}},
		{Code: "/** @type {function (): void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", FileName: "file.js", Output: "/** @type {function (): void} */\nconst f = () => {};\nf.substring = () => \"custom\";\nf.substring();", Errors: []rule_tester.InvalidTestCaseError{}},
		{Code: "/** @type {function} */\nconst o = \"abc\";\no.substring();", FileName: "file.js", Output: "/** @type {function} */\nconst o = \"abc\";\no.slice();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14}}},
	}
	for _, test := range tests {
		t.Run(test.Code, func(t *testing.T) {
			diagnostics := lintPreferStringSliceSourceOnly(t, test.Code, test.FileName)
			if len(diagnostics) != len(test.Errors) {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(diagnostics), len(test.Errors), diagnostics)
			}
			for i, diagnostic := range diagnostics {
				expected := test.Errors[i]
				line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
				endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.End())
				if diagnostic.Message.Id != expected.MessageId || diagnostic.Message.Description != expected.Message ||
					line+1 != expected.Line || int(column)+1 != expected.Column || endLine+1 != expected.EndLine || int(endColumn)+1 != expected.EndColumn {
					t.Fatalf("unexpected diagnostic: %+v at %d:%d-%d:%d", diagnostic.Message, line+1, column+1, endLine+1, endColumn+1)
				}
			}
			output, _, fixed := linter.ApplyRuleFixes(test.Code, diagnostics)
			if output != test.Output || fixed != (test.Output != test.Code) {
				t.Fatalf("fixed=%v, output=%q, want %q", fixed, output, test.Output)
			}
		})
	}
}
