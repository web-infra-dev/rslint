// TestPreferCodePointExtras locks in branches and AST shapes that the upstream
// suite does not exercise: optional-call rejection, computed/optional
// `fromCharCode` receivers including JSDoc-transparent ones, every numeric
// operator and compound assignment, and numeric-context wrappers around
// optional chains.
package prefer_code_point_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_code_point"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferCodePointExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_code_point.PreferCodePointRule,
		[]rule_tester.ValidTestCase{
			// TypeScript wrappers on the receiver do not change matching.
			validTS(`(str as string).charCodeAt?.(0)`),

			// Optional calls are not `charCodeAt` calls in the upstream matcher.
			valid(`string.charCodeAt?.(index)`),
			valid(`string?.charCodeAt?.(index)`),
			valid(`foo?.charCodeAt?.(0)`),

			// Non-call uses of the method reference.
			valid(`foo.charCodeAt.call(foo, 0)`),
			valid(`const method = foo.charCodeAt`),

			// `fromCharCode` needs a non-optional dot access on the `String` identifier.
			valid(`String?.fromCharCode(code)`),
			valid(`String["fromCharCode"](code)`),
			valid(`String[fromCharCode](code)`),
			valid(`window.String.fromCharCode(code)`),
			valid(`NotString.fromCharCode(code)`),
			// Authored TypeScript assertions are not transparent in ESTree.
			validTS(`(String as any).fromCharCode(code)`),
		},
		[]rule_tester.InvalidTestCase{
			// No argument-count constraint is applied.
			invalid(`string.charCodeAt()`, "charCodeAt", "codePointAt", true),

			// Parenthesized callees are transparent.
			invalid(`(string.charCodeAt)(index)`, "charCodeAt", "codePointAt", true),

			// The `fromCharCode` member itself is reported, even when called optionally
			// or used as a value.
			invalid(`String.fromCharCode?.(code)`, "fromCharCode", "fromCodePoint", true),
			invalid(`String.fromCharCode.length`, "fromCharCode", "fromCodePoint", true),
			invalid(`function f(String) { return String.fromCharCode(1); }`, "fromCharCode", "fromCodePoint", true),

			// JSDoc casts are comments in ESTree, so the receiver remains the
			// `String` identifier for a JavaScript file.
			invalid(`/** @type {any} */ (String).fromCharCode(65)`, "fromCharCode", "fromCodePoint", true),
			invalid(`(/** @type {any} */ (String)).fromCharCode(65)`, "fromCharCode", "fromCodePoint", true),
			invalid(`/** @type {any} */ (String).fromCharCode?.(65)`, "fromCharCode", "fromCodePoint", true),

			// ---- Numeric binary operators keep the diagnostic but drop the suggestion ----
			invalid(`string.charCodeAt(0) / 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) % 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) ** 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) & mask`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) | mask`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) << 1`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) >> 1`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) >>> 1`, "charCodeAt", "codePointAt", false),
			invalid(`1 + string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),

			// ---- Numeric compound assignments require the call on the right ----
			invalid(`sum -= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum *= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum /= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum %= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum **= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum &= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum |= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum ^= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum <<= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum >>= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum >>>= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),

			// Logical assignments are not numeric contexts.
			invalid(`sum &&= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`sum ||= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`sum ??= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// Plain assignment is not a numeric context either.
			invalid(`x = string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// ---- Unary operators ----
			invalid(`+string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`~string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`typeof string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`void string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// ---- Wrappers around optional chains ----
			invalid(`(string?.charCodeAt(0)) + 1`, "charCodeAt", "codePointAt", false),
			invalid(`sum += string.charCodeAt(0) && 1`, "charCodeAt", "codePointAt", false),
			invalidTS(`string?.charCodeAt(0)! + 1`, "charCodeAt", "codePointAt", false),
			invalidTS(`string?.charCodeAt(0)!`, "charCodeAt", "codePointAt", true),
		},
	)
}

func TestPreferCodePointEditDemand(t *testing.T) {
	const code = "string.charCodeAt(index);\n" +
		"string.charCodeAt(index) + 1;\n" +
		"String.fromCharCode(65);"
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	prog, err := program.NewFromRoots(program.RootOptions{
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		RootFileNames:   []string{file},
		CompilerOptions: program.SourceOnlyCompilerOptions(),
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	target := prefer_code_point.PreferCodePointRule
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: prog, File: file,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: target.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return target.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d: unexpected autofix for diagnostic %d", demand, i)
			}
			// Only the non-numeric diagnostics offer a rename suggestion.
			wantSuggestion := (i == 0 || i == 2) && demand&rule.EditDemandSuggestion != 0
			if (diagnostic.Suggestions != nil) != wantSuggestion {
				t.Errorf("demand %d: unexpected suggestions for diagnostic %d", demand, i)
			}
			if wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all[i].Suggestions) {
				t.Errorf("demand %d: suggestions differ for diagnostic %d", demand, i)
			}
		}
	}
}
