// TestNoUnusedVarsExtras locks in branches and edge shapes that the upstream
// suite does not exercise. Each case points at its Dimension 4 row or the
// real-user report it protects so future refactors cannot silently weaken the
// port.
//
// Dimension 4 notes:
//   - Expression wrappers are usage sites, so nested parentheses and TS
//     assertions must preserve the underlying reference.
//   - Access/key forms distinguish a plain property name from computed and
//     shorthand references.
//   - Declaration/container forms cover nested functions, shadowing, and the
//     string-form "local" option.
//   - Same-kind nesting and source-order reporting are covered by multiple
//     declarations in one file.
//   - Graceful degradation without a checker is tested separately.
//   - N/A autofix boundary: ESLint core no-unused-vars offers suggestions, not
//     automatic fixes.
package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: nested expression wrappers preserve the usage ----
			{Code: `const wrapped = 1; ((wrapped as number)!);`},

			// ---- Dimension 4: computed and shorthand keys are value references ----
			{Code: `const computed = "key"; ({ [computed]: true });`},
			{Code: `const shorthand = 1; ({ shorthand });`},

			// ---- Dimension 4: closure containers preserve captured references ----
			{Code: `let captured; const read = () => captured; read();`},

			// ---- Dimension 4: string-form options skip only the global scope ----
			{Code: `const globalOnly = 1;`, Options: []interface{}{"local"}},

			// ---- Dimension 4: JavaScript-compatible regexp syntax uses regexp2 ----
			{Code: `const _unused = 1;`, Options: map[string]interface{}{"varsIgnorePattern": `(?<=_)unused`}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: a dot-property name is not a reference to the local ----
			extraUnusedCase(
				`const local = 1; object.local;`,
				"local",
				true,
				1,
				7,
				12,
				` object.local;`,
			),

			// ---- Dimension 4: an inner parameter does not use a shadowed outer binding ----
			extraUnusedCase(
				`const item = 1; function use(item) { return item; } use(1);`,
				"item",
				true,
				1,
				7,
				11,
				` function use(item) { return item; } use(1);`,
			),

			// ---- Dimension 4: string-form "local" still checks nested scopes ----
			{
				Code:    "function outer() { const local = 1; }\nouter();",
				Options: []interface{}{"local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("local", true, 1, 26, 31, "function outer() {  }\nouter();"),
				},
			},

			// ---- Dimension 4: diagnostics remain in source order across declarations ----
			{
				Code: "const later = 1;\nconst earlier = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("later", true, 1, 7, 12, "\nconst earlier = 2;"),
					extraUnusedError("earlier", true, 2, 7, 14, "const later = 1;\n"),
				},
			},

			// ---- Real-user: eslint/eslint#14324 reports the last write, not the last read ----
			extraUnusedCase(
				"let x = [];\nx = x.concat(x);",
				"x",
				true,
				2,
				1,
				2,
				"",
			),

			// ---- Real-user: eslint/eslint#14325 keeps comma-expression updates unused ----
			extraUnusedCase(
				"let x = 0;\nx++, x = 0;\nx = 3;",
				"x",
				true,
				3,
				1,
				2,
				"",
			),
		},
	)
}

func TestNoUnusedVarsWithoutTypeChecker(t *testing.T) {
	t.Parallel()
	listeners := NoUnusedVarsRule.Run(rule.RuleContext{}, nil)
	if len(listeners) != 0 {
		t.Fatalf("expected graceful degradation without a type checker, got %d listeners", len(listeners))
	}
}

func extraUnusedCase(code string, name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			extraUnusedError(name, assigned, line, column, endColumn, suggestionOutput),
		},
	}
}

func extraUnusedError(name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	action := "defined"
	if assigned {
		action = "assigned a value"
	}
	result := rule_tester.InvalidTestCaseError{
		MessageId: "unusedVar",
		Message:   "'" + name + "' is " + action + " but never used.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
	if suggestionOutput != "" {
		result.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: "removeVar",
				Output:    suggestionOutput,
			},
		}
	}
	return result
}

func extraUnusedErrorWithSuggestion(name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	result := extraUnusedError(name, assigned, line, column, endColumn, "")
	result.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
		{
			MessageId: "removeVar",
			Output:    suggestionOutput,
		},
	}
	return result
}
