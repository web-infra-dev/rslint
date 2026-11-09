package prefer_math_trunc_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_trunc"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferMathTruncExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `0 | value; value >>> 0; value &= 0; value | 0n; value | "0";`, FileName: "input.js"},
		{Code: `value | 5e-324; parseInt(String(value), 10.000000000000002);`, FileName: "input.js"},
		{Code: `parseInt(String?.(value), 10); (Number?.parseInt)(String(value), 10); parseInt(String(value), +10);`, FileName: "input.js"},
		{Code: `~(~value as number); value | (0 as number); parseInt(String(value) as string, 10); (parseInt as typeof parseInt)(String(value), 10);`, FileName: "input.ts"},
		{Code: `parseInt((String<() => number>)(value), 1_0);`, FileName: "input.ts"},
		{Code: `parseInt(String(value), 10); Number.parseInt(String(value), 10);`, FileName: "input.js", Globals: map[string]any{"String": "off"}},
		{Code: `parseInt(String(value), 10); Number.parseInt(String(value), 10);`, FileName: "input.js", Globals: map[string]any{"parseInt": "off", "Number": "off"}},
	}
	invalid := []rule_tester.InvalidTestCase{
		// The message retains the literal spelling even when its value rounds to zero.
		truncationCase("input.js", "const prefix = \"😀\"; value | (\n  0e-999\n);",
			truncationError("bitwise", "| 0e-999", 1, 22, 3, 2, `const prefix = "😀"; Math.trunc(value);`)),
		truncationCase("input.js", `(values[index])<<=(0o0);`,
			truncationError("bitwise", "<<= 0o0", 1, 1, 1, 24, `(values[index])=(Math.trunc(values[index]));`)),
		truncationCase("input.js", `class Box { #value = 1.25; update() { this.#value ^= 0; } }`,
			truncationError("bitwise", "^= 0", 1, 39, 1, 55, `class Box { #value = 1.25; update() { this.#value = Math.trunc(this.#value); } }`)),
		truncationCase("input.tsx", `const component = <Component value={~~number} />;`,
			truncationError("bitwise-not", "~~", 1, 37, 1, 45, `const component = <Component value={Math.trunc(number)} />;`)),
		truncationCase("input.js", `(Number).p\u0061rseInt(((String))( (value), ), (0xA));`,
			truncationError("parse-int", "Number.parseInt(String(...), 10)", 1, 1, 1, 54, `Math.trunc( (value), );`)),
		truncationCase("input.js", `parseInt(String(value /* ) */, /* ) , */), 10);`,
			truncationError("parse-int", "parseInt(String(...), 10)", 1, 1, 1, 47, `Math.trunc(value /* ) */, /* ) , */);`)),
		truncationCase("input.ts", `parseInt(String((value satisfies number) /* ) */), 0x0_A);`,
			truncationError("parse-int", "parseInt(String(...), 10)", 1, 1, 1, 58, `Math.trunc((value satisfies number) /* ) */);`)),
		truncationCase("input.js", `(/** @type {number} */ (value)) | 0;`,
			truncationError("bitwise", "| 0", 1, 1, 1, 36, `Math.trunc((/** @type {number} */ (value)));`)),
		truncationCase("input.js", `~(/** @type {number} */ (~value));`,
			truncationError("bitwise-not", "~~", 1, 1, 1, 34, "")),
		truncationCase("input.js", `get().x |= 0; values[++index] >>= 0; values[(index = 0)] <<= 0;`,
			truncationError("bitwise", "|= 0", 1, 1, 1, 13, ""),
			truncationError("bitwise", ">>= 0", 1, 15, 1, 36, ""),
			truncationError("bitwise", "<<= 0", 1, 38, 1, 63, "")),
		// Upstream would duplicate the tagged template call in its suggestion.
		truncationCase("input.js", "values[key`index`] |= 0;",
			truncationError("bitwise", "|= 0", 1, 1, 1, 24, "")),
		truncationCase("input.js", `function f(Math) { return ~~value; } { const Math = {}; value |= 0; parseInt(String(value), 10); }
~~outer;`,
			truncationError("bitwise-not", "~~", 1, 27, 1, 34, ""),
			truncationError("bitwise", "|= 0", 1, 57, 1, 67, ""),
			truncationError("parse-int", "parseInt(String(...), 10)", 1, 69, 1, 96, ""),
			truncationError("bitwise-not", "~~", 2, 1, 2, 8, `function f(Math) { return ~~value; } { const Math = {}; value |= 0; parseInt(String(value), 10); }
Math.trunc(outer);`)),
		{
			FileName: "input.js",
			Code:     `parseInt(String(value), 10); ~~value;`,
			Globals:  map[string]any{"Math": "off"},
			Errors: []rule_tester.InvalidTestCaseError{
				truncationError("parse-int", "parseInt(String(...), 10)", 1, 1, 1, 28, ""),
				truncationError("bitwise-not", "~~", 1, 30, 1, 37, ""),
			},
		},
		truncationCase("input.ts", `const n = (~~value)as number;`,
			truncationError("bitwise-not", "~~", 1, 12, 1, 19, `const n = (Math.trunc(value)) as number;`)),
		truncationCase("input.js", `function* f() { return ~~(yield* read()); } for (const item of(~~input)) read(item);`,
			truncationError("bitwise-not", "~~", 1, 24, 1, 41, `function* f() { return Math.trunc(yield* read()); } for (const item of(~~input)) read(item);`),
			truncationError("bitwise-not", "~~", 1, 64, 1, 71, `function* f() { return ~~(yield* read()); } for (const item of (Math.trunc(input))) read(item);`)),
		truncationCase("input.js", `function cast(value) { return (/** @type {number} */ (~~value)); }`,
			truncationError("bitwise-not", "~~", 1, 55, 1, 62, `function cast(value) { return (/** @type {number} */ (Math.trunc(value))); }`)),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule, valid, invalid)
}

func TestPreferMathTruncEditDemand(t *testing.T) {
	const source = `value | 0;
~~value;
Number.parseInt(String(value), 10);
items[0] ^= 0;
items[index++] |= 0;`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: source})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	lintProgram := program.NewFromCompiler(compilerProgram)
	results := make(map[rule.EditDemand][]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintProgram, File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: prefer_math_trunc.PreferMathTruncRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_math_trunc.PreferMathTruncRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				results[demand] = append(results[demand], d)
			}},
		})
		if len(results[demand]) != 5 {
			t.Fatalf("demand %d: got %d diagnostics, want 5", demand, len(results[demand]))
		}
	}
	for demand, diagnostics := range results {
		for i, diagnostic := range diagnostics {
			metadata := diagnostic
			metadata.FixesPtr, metadata.Suggestions = nil, nil
			if !reflect.DeepEqual(metadata, results[rule.EditDemandNone][i]) {
				t.Errorf("demand %d changed diagnostic %d metadata", demand, i)
			}
			wantSuggestion := demand&rule.EditDemandSuggestion != 0 && i < 4
			if diagnostic.FixesPtr != nil || (diagnostic.Suggestions != nil) != wantSuggestion {
				t.Fatalf("demand %d: diagnostic %d has unexpected edits", demand, i)
			}
			if wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, results[rule.EditDemandAll][i].Suggestions) {
				t.Errorf("demand %d changed diagnostic %d suggestions", demand, i)
			}
		}
	}
}
