// Additional raw-text, AST adaptation and edit-demand coverage.
package no_console_spaces_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_console_spaces"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConsoleSpacesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Optional, computed, private and non-console calls
		{Code: "console.log?.(\"a \", \" b\");", FileName: "case.js"},
		{Code: "console?.log(\"a \", \" b\");", FileName: "case.js"},
		{Code: "console?.log?.(\"a \", \" b\");", FileName: "case.js"},
		{Code: "(console?.log)(\"a \", \" b\");", FileName: "case.js"},
		{Code: "(console.log)?.(\"a \", \" b\");", FileName: "case.js"},
		{Code: "globalThis.console.log(\"a \", \" b\");", FileName: "case.js"},
		{Code: "class C { #log() {} run(console) { console.#log(\"a \", \" b\"); } }", FileName: "case.js"},
		{Code: "console[`log`](\"a \", \" b\");", FileName: "case.js"},
		// Raw text and non-literal arguments
		{Code: "console.log(\"a\", \"\", \"b\");", FileName: "case.js"},
		{Code: "console.log(\"a\", \"  middle  \", \"b\");", FileName: "case.js"},
		{Code: "console.log(\"a \" + value, value + \" b\");", FileName: "case.js"},
		{Code: "const message = \"x \"; console.log(message, message);", FileName: "case.js"},
		{Code: "console.log(tag`a `, tag` b`);", FileName: "case.js"},
		{Code: "console.log(\"a\\u0020\", \"\\x20b\", \"a\\u{20}\", \"\\u0020b\");", FileName: "case.js"},
		{Code: "console.log(\"a\\t\", \"\\tb\", \"a\\n\", \"\\nb\");", FileName: "case.js"},
		{Code: "console.log(\"a\u00a0\", \"\u00a0b\", \"a\u2003\", \"\u2003b\");", FileName: "case.js"},
		{Code: "console.log(...[\"a \", \" b\"]);", FileName: "case.js"},
		// TypeScript wrappers and JSX
		{Code: "(console as Console).log(\"a \", \" b\");", FileName: "case.tsx"},
		{Code: "console!.log(\"a \", \" b\");", FileName: "case.tsx"},
		{Code: "(console.log as Function)(\"a \", \" b\");", FileName: "case.tsx"},
		{Code: "console.log!(\"a \", \" b\");", FileName: "case.tsx"},
		{Code: "console.log(\"a \" as const, \" b\" satisfies string);", FileName: "case.tsx"},
		{Code: "console.log(\"a \"!, \" b\"!);", FileName: "case.tsx"},
		{Code: "const element = <console.log title=\" a \" />;", FileName: "case.tsx"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Raw text and non-literal arguments
		{Code: "console.log(\"a \", ...values, \" b\");", FileName: "case.js", Output: []string{"console.log(\"a\", ...values, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 15), spaceError("log", "leading", 1, 31)}},
		{Code: "console.log(0, \" both \", null, \" end\");", FileName: "case.js", Output: []string{"console.log(0, \"both\", null, \"end\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 17), spaceError("log", "trailing", 1, 22), spaceError("log", "leading", 1, 33)}},
		{Code: "console.log(\"a\", \" leading  \", \"  trailing \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\", \"leading  \", \"  trailing\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 43)}},
		{Code: "console.log(\"a\", ` ${value} `, \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\", `${value}`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 28)}},
		{Code: "console.log(`a\n `, ` b`);", FileName: "case.js", Output: []string{"console.log(`a\n`, `b`);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 2, 1), spaceError("log", "leading", 2, 6)}},
		{Code: "console.log(`a\r\n `, \"b\");", FileName: "case.js", Output: []string{"console.log(`a\r\n`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 2, 1)}},
		{Code: "console.log(`\n`, \" b\");", FileName: "case.js", Output: []string{"console.log(`\n`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 2, 5)}},
		{Code: "console.log(\"\\u0020 \", \" \\x20\", \"b\");", FileName: "case.js", Output: []string{"console.log(\"\\u0020\", \"\\x20\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 20), spaceError("log", "leading", 1, 25)}},
		{Code: "console.log(\"😀 \", \" 中文 \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"😀\", \"中文\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 16), spaceError("log", "leading", 1, 21), spaceError("log", "trailing", 1, 24)}},
		{Code: "\"😀\";\r\nconsole.log(\r\n\"a \",\r\n\" b\"\r\n);", FileName: "case.js", Output: []string{"\"😀\";\r\nconsole.log(\r\n\"a\",\r\n\"b\"\r\n);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 3, 3), spaceError("log", "leading", 4, 2)}},
		// Parentheses, comments, JSDoc and local bindings
		{Code: "(console).log((\"a \"), (\" b\"));", FileName: "case.js", Output: []string{"(console).log((\"a\"), (\"b\"));"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 18), spaceError("log", "leading", 1, 25)}},
		{Code: "(console.log)(/* first */ (\"a \"), /* second */ (\" b\"));", FileName: "case.js", Output: []string{"(console.log)(/* first */ (\"a\"), /* second */ (\"b\"));"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 30), spaceError("log", "leading", 1, 50)}},
		{Code: "/** @type {Console} */ (console).log(\"a \", \"b\");", FileName: "case.js", Output: []string{"/** @type {Console} */ (console).log(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 40)}},
		{Code: "(/** @type {Function} */ (console.log))(\"a \", \"b\");", FileName: "case.js", Output: []string{"(/** @type {Function} */ (console.log))(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 43)}},
		{Code: "console.log(/** @type {string} */ (\"a \"), \"b\");", FileName: "case.js", Output: []string{"console.log(/** @type {string} */ (\"a\"), \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 38)}},
		{Code: "function log(console) { console.warn(\"a \", \" b\"); }", FileName: "case.js", Output: []string{"function log(console) { console.warn(\"a\", \"b\"); }"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("warn", "trailing", 1, 40), spaceError("warn", "leading", 1, 45)}},
		{Code: "con\\u0073ole.l\\u006fg(\"a \", \"b\");", FileName: "case.js", Output: []string{"con\\u0073ole.l\\u006fg(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 25)}},
		// TypeScript wrappers and JSX
		{Code: "console.log<string>(\"a \", \" b\");", FileName: "case.tsx", Output: []string{"console.log<string>(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 23), spaceError("log", "leading", 1, 28)}},
		{Code: "const element = <div>{console.log(\"a \", \" b\")}</div>;", FileName: "case.tsx", Output: []string{"const element = <div>{console.log(\"a\", \"b\")}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 37), spaceError("log", "leading", 1, 42)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}

func TestNoConsoleSpacesEditDemand(t *testing.T) {
	for _, testCase := range []struct {
		name, source, output string
		count                int
	}{
		{"multiple fixes", `console.log("a ", " both ", "b");`, `console.log("a", "both", "b");`, 3},
		{"escaped space", `console.log("a\ ", "b");`, `console.log("a", "b");`, 1},
		{"next line disabled", "// eslint-disable-next-line unicorn/no-console-spaces\nconsole.log(\"a \", \" b\");", "// eslint-disable-next-line unicorn/no-console-spaces\nconsole.log(\"a \", \" b\");", 0},
		{"same line disabled", `console.log("a ", " b"); // eslint-disable-line unicorn/no-console-spaces`, `console.log("a ", " b"); // eslint-disable-line unicorn/no-console-spaces`, 0},
		{"block enable", "/* eslint-disable unicorn/no-console-spaces */\nconsole.log(\"a \", \" b\");\n/* eslint-enable unicorn/no-console-spaces */\nconsole.log(\"c \", \" d\");", "/* eslint-disable unicorn/no-console-spaces */\nconsole.log(\"a \", \" b\");\n/* eslint-enable unicorn/no-console-spaces */\nconsole.log(\"c\", \"d\");", 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			r := no_console_spaces.NoConsoleSpacesRule
			var baseline []rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, d)
					}},
				})
				if len(diagnostics) != testCase.count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), testCase.count)
				}
				if demand == rule.EditDemandAll {
					baseline = diagnostics
				}
				wantFix := demand&rule.EditDemandAutofix != 0 && testCase.count > 0
				for index, diagnostic := range diagnostics {
					all := baseline[index]
					if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) || diagnostic.Severity != all.Severity {
						t.Errorf("demand %d changed diagnostic identity", demand)
					}
					if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
						t.Errorf("demand %d: unexpected edit artifacts", demand)
					}
					if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
						t.Errorf("demand %d changed autofix artifacts", demand)
					}
				}
				wantOutput := testCase.source
				if wantFix {
					wantOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics)
				if output != wantOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
			}
		})
	}
}

// Nested calls, line boundaries and authored TypeScript wrappers.
// Expected diagnostics follow native listener order: outer calls before inner calls.
func TestNoConsoleSpacesSyntaxBoundaries(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "console.log((/* comment */ \" both \"));", FileName: "case.js"},
		{Code: "(console.log<string>)(\"a \", \" b\");", FileName: "case.ts"},
		{Code: "console.log(<string>\"a \", <string>\" b\");", FileName: "case.ts"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "console.log(` ${console.log(\"a \", \"b\")} `);", FileName: "case.js", Output: []string{"console.log(` ${console.log(\"a\", \"b\")} `);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 31)}},
		{Code: "console.log(` ${console.log(\"a \", \"b\")} `, \"c\");", FileName: "case.js", Output: []string{"console.log(` ${console.log(\"a\", \"b\")}`, \"c\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 40), spaceError("log", "trailing", 1, 31)}},
		{Code: "console.log(\"a\", ` ${console.log(\"b \", \"c\")} `, \"d\");", FileName: "case.js", Output: []string{"console.log(\"a\", `${console.log(\"b\", \"c\")}`, \"d\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 45), spaceError("log", "trailing", 1, 36)}},
		{Code: "console.log(/* before */\n \"a \", /* after */\n \" b\");", FileName: "case.js", Output: []string{"console.log(/* before */\n \"a\", /* after */\n \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 2, 4), spaceError("log", "leading", 3, 3)}},
		{Code: "\ufeff" + "console.log(\"a \", \" b\");", FileName: "case.js", Output: []string{"\ufeff" + "console.log(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 15), spaceError("log", "leading", 1, 20)}},
		{Code: "console.log(`a\u2028 `, ` b\u2029`);", FileName: "case.js", Output: []string{"console.log(`a\u2028`, `b\u2029`);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 2, 1), spaceError("log", "leading", 2, 6)}},
		{Code: "console.log('it\\'s ', \"b\");", FileName: "case.js", Output: []string{"console.log('it\\'s', \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 19)}},
		{Code: "class C extends console.log(\"a \", \"b\") {}", FileName: "case.ts", Output: []string{"class C extends console.log(\"a\", \"b\") {}"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 31)}},
		{Code: "const x = <A {...console.log(\"a \", \" b\")} />;", FileName: "case.tsx", Output: []string{"const x = <A {...console.log(\"a\", \"b\")} />;"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 32), spaceError("log", "leading", 1, 37)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}

// Unlike Unicorn v76.0.0, fixes remove an escaped space together with its backslash so the closing quote remains valid.
func TestNoConsoleSpacesEscapedSpaces(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "console.log(\"a\\ \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 16)}},
		{Code: "console.log('abc\\ ', 'def');", FileName: "case.js", Output: []string{"console.log('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 18)}},
		{Code: "console.log(\"a\\\\ \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\\\\\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17)}},
		{Code: "console.log(\"a\\\\\\ \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\\\\\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 18)}},
		{Code: "console.log(\"a\\\\\\\\ \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\\\\\\\\\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 19)}},
		{Code: "console.log(`a\\ `, \"b\");", FileName: "case.js", Output: []string{"console.log(`a`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 16)}},
		{Code: "console.log(`a${value}\\ `, \"b\");", FileName: "case.js", Output: []string{"console.log(`a${value}`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 24)}},
		{Code: "console.log(`a${value}\\\\ `, \"b\");", FileName: "case.js", Output: []string{"console.log(`a${value}\\\\`, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 25)}},
		{Code: "console.log(\"a\", \" \\ \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\", \"\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 21)}},
		{Code: "console.log(\"a\", ` \\ `, \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\", ``, \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 21)}},
		{Code: "console.log(\"a\\\n \", \"b\");", FileName: "case.js", Output: []string{"console.log(\"a\\\n\", \"b\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 2, 1)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}
