package prefer_date_now_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_date_now"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferDateNowExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `new Date()?.getTime()`},
		{Code: `new Date().valueOf?.()`},
		{Code: `(new Date()?.getTime)()`},
		{Code: `Number?.(new Date())`},
		{Code: `BigInt?.(new Date())`},
		{Code: `Number(...[new Date()])`},
		{Code: `new Date(0).getTime(); Number(new Date(0)); foo -= new Date(0); new Date(0) * 2;`},
		{Code: `!new Date(); ~new Date(); void new Date(); new Date() === 0; foo = new Date();`},
		{Code: `class Clock { #getTime() {} read() { return new Date().#getTime(); } }`},
		// Authored TypeScript wrappers remain visible to ESTree selectors.
		{Code: `+(new Date() as Date)`, FileName: "file.ts"},
		{Code: `-(new Date() satisfies Date)`, FileName: "file.ts"},
		{Code: `new Date()!.getTime()`, FileName: "file.ts"},
		{Code: `(Number as Function)(new Date())`, FileName: "file.ts"},
		{Code: `new (Date as DateConstructor)().getTime()`, FileName: "file.ts"},
		{Code: `Number(new Date() as Date)`, FileName: "file.ts"},
		{Code: `(new Date() as Date) * 2; foo -= new Date() as Date;`, FileName: "file.ts"},
	}
	invalid := []rule_tester.InvalidTestCase{
		invalidCase(`new (Date)().getTime()`, `Date.now()`, methodMessageID, "getTime"),
		invalidCase(`((new Date().getTime))()`, `Date.now()`, methodMessageID, "getTime"),
		invalidCase(`(Number)((new Date()))`, `Date.now()`, numberMessageID, "(Number)((new Date()))"),
		invalidCase(`BigInt((new Date()))`, `BigInt((Date.now()))`, dateMessageID, "new Date()"),
		invalidCase(`function clock(Date, Number) { return Number(new Date()); }`, `function clock(Date, Number) { return Date.now(); }`, numberMessageID, "Number(new Date())"),
		invalidCase(`async function clock() { return await+new Date; }`, `async function clock() { return await Date.now(); }`, dateMessageID, "+new Date"),
		// Upstream inserts keyword spacing outside the parenthesized range too.
		invalidCase(`function clock() { return(+new Date); }`, `function clock() { return (Date.now()); }`, dateMessageID, "+new Date"),
		invalidCase(`function clock() { throw+new Date; }`, `function clock() { throw Date.now(); }`, dateMessageID, "+new Date"),
		invalidCase(`const timestamp = +/* keep outside? */(new Date);`, `const timestamp = Date.now();`, dateMessageID, "+/* keep outside? */(new Date)"),
		invalidCase(`const 时钟 = "😀"; new Date().getTime();`, `const 时钟 = "😀"; Date.now();`, methodMessageID, "getTime"),
		invalidCase("const timestamp = Number(\n  new Date()\n);", "const timestamp = Date.now();", numberMessageID, "Number(\n  new Date()\n)"),
		invalidCase("const timestamp = new Date()\n  .getTime();", "const timestamp = Date.now();", methodMessageID, "getTime"),
		// JavaScript JSDoc casts are transparent, unlike authored TS assertions.
		invalidCase(`+(/** @type {Date} */ (new Date()))`, `Date.now()`, dateMessageID, "+(/** @type {Date} */ (new Date()))"),
		invalidCase(`new (/** @type {DateConstructor} */ (Date))().getTime()`, `Date.now()`, methodMessageID, "getTime"),
		// The full parenthesized range includes synthetic JSDoc cast wrappers.
		invalidCase(`function clock(){return(/** @type {number} */ (+new Date));}`, `function clock(){return (/** @type {number} */ (Date.now()));}`, dateMessageID, "+new Date"),
		invalidCase(`function clock(){return(/** @type {number} */ (+new Date))in {}}`, `function clock(){return (/** @type {number} */ (Date.now())) in {}}`, dateMessageID, "+new Date"),
		invalidCase(`function clock(){return(/** @satisfies {number} */ (+new Date));}`, `function clock(){return (/** @satisfies {number} */ (Date.now()));}`, dateMessageID, "+new Date"),
	}
	for _, operator := range []string{"-", "*", "/", "%", "**"} {
		invalid = append(invalid, invalidCase(
			"const elapsed = new Date() "+operator+" new Date();",
			"const elapsed = Date.now() "+operator+" Date.now();",
			dateMessageID, "new Date()", "new Date()"))
	}
	// Type arguments and JSX containers do not change the runtime call shape.
	tsCase := invalidCase(`new Date<string>().getTime<number>()`, `Date.now()`, methodMessageID, "getTime")
	tsCase.FileName = "file.ts"
	invalid = append(invalid, tsCase)
	jsxCase := invalidCase(`const Clock = () => <time>{new Date().getTime()}</time>;`, `const Clock = () => <time>{Date.now()}</time>;`, methodMessageID, "getTime")
	jsxCase.FileName = "file.tsx"
	jsxCase.Tsx = true
	invalid = append(invalid, jsxCase)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_date_now.PreferDateNowRule, valid, invalid)
}

func TestPreferDateNowEditDemand(t *testing.T) {
	t.Run("keyword spacing", func(t *testing.T) {
		testPreferDateNowEditDemand(t,
			`function clock(){return(/** @type {number} */ (+new Date))in {}}`,
			`function clock(){return (/** @type {number} */ (Date.now())) in {}}`,
			rule.RuleMessage{Id: dateMessageID, Description: "Prefer `Date.now()` over `new Date()`."},
		)
	})
	for _, method := range []string{"getTime", "valueOf"} {
		t.Run(method, func(t *testing.T) {
			testPreferDateNowEditDemand(t, "new Date()."+method+"()", "Date.now()", rule.RuleMessage{
				Id:          methodMessageID,
				Description: "Prefer `Date.now()` over `Date#" + method + "()`.",
				Data:        map[string]string{"method": method},
			})
		})
	}
}

func testPreferDateNowEditDemand(t *testing.T, source, fixedSource string, message rule.RuleMessage) {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: source})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	if compilerProgram.GetSourceFile(fileName) == nil {
		t.Fatal("missing edit-demand source file")
	}
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program.NewFromCompiler(compilerProgram), File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: prefer_date_now.PreferDateNowRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_date_now.PreferDateNowRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				got = append(got, diagnostic)
			}},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}
	base := diagnostics[rule.EditDemandNone]
	if !reflect.DeepEqual(base.Message, message) {
		t.Fatalf("message = %#v, want %#v", base.Message, message)
	}
	for demand, diagnostic := range diagnostics {
		metadata := diagnostic
		metadata.FixesPtr = nil
		metadata.Suggestions = nil
		if !reflect.DeepEqual(metadata, base) {
			t.Errorf("demand %d changed diagnostic metadata", demand)
		}
		if diagnostic.Suggestions != nil {
			t.Errorf("demand %d unexpectedly produced suggestions", demand)
		}
		if demand == rule.EditDemandNone || demand == rule.EditDemandSuggestion {
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d unexpectedly produced fixes", demand)
			}
			continue
		}
		if diagnostic.FixesPtr == nil || !reflect.DeepEqual(diagnostic.FixesPtr, diagnostics[rule.EditDemandAll].FixesPtr) {
			t.Fatalf("demand %d has missing or inconsistent fixes", demand)
		}
	}
	// Applying fixes sorts their slices in place. Compare all demands before
	// doing so, regardless of the map's iteration order.
	for _, demand := range []rule.EditDemand{rule.EditDemandAutofix, rule.EditDemandAll} {
		diagnostic := diagnostics[demand]
		output, unapplied, fixed := linter.ApplyRuleFixes(source, []rule.RuleDiagnostic{diagnostic})
		if !fixed || len(unapplied) != 0 || output != fixedSource {
			t.Errorf("demand %d: fixed = %v, unapplied = %d, output = %q", demand, fixed, len(unapplied), output)
		}
	}
}
