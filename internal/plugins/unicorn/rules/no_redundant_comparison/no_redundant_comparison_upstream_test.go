// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-redundant-comparison.js
// All diagnostic expectations are preserved. Unsafe edits are withheld as documented.
package no_redundant_comparison_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_redundant_comparison"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoRedundantComparisonUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// explicit side effects invalidate equality, strict inequality, and bound implications
		{Code: "let first = 1; let second = 1; first === second && first === 1 && (second = 2) === 2 && second === 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; first === second && first === 1 && second++ === 1 && second === 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; function change() { second = 2; return true; } first === second && first === 1 && change() && second === 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; function change() { second = 2; return true; } first === second && first === 1 && change`value` && second === 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 2; first === 1 && (first = 2) === 2 && second === 1 && first === second;", FileName: "case.js"},
		{Code: "let first = 1; let second = 2; first === 1 && first !== second && (second = 1) === 1 && second !== 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; first === second && first > 0 && (second = 0) === 0 && second > 0;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; let third = 1; first === second && second === third && first === 1 && (third = 2) === 2 && third === 1;", FileName: "case.js"},
		// Explicit side effects and wrappers
		{Code: "first === second && first === 1 && /* keep */ change() && second === 1;", FileName: "case.js"},
		{Code: "((first === second) && (first === 1) && ((second = 2) === 2) && (second === 1));", FileName: "case.js"},
		{Code: "first === second && first === 1 && (condition ? change() : true) && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && (second = third, true) && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && delete object.value && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && new SomeClass() && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && (change`value` === true) && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && (condition ? change`value` : true) && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && (change`value`, true) && second === 1;", FileName: "case.js"},
		{Code: "first === second && first === 1 && /* keep */ ((change`value`)) && second === 1;", FileName: "case.js"},
		{Code: "let first = 1; let second = 1; (first as number) === second && first === 1 && ((second = 2) as number) === 2 && second === 1;", FileName: "case.ts"},
		{Code: "first! === second && first === 1 && (change() satisfies boolean) && second === 1;", FileName: "case.ts"},
		{Code: "first! === second && first === 1 && (change`value` satisfies boolean) && second === 1;", FileName: "case.ts"},
		// Snapshot cases
		{Code: "a > 0 || b > 0 || a === b", FileName: "case.js"},
		{Code: "a > 0 ?? b > 0 ?? a === b", FileName: "case.js"},
		{Code: "a > 0", FileName: "case.js"},
		{Code: "a === b", FileName: "case.js"},
		{Code: "a > 0 && a === b", FileName: "case.js"},
		{Code: "a > 0 && b > 0", FileName: "case.js"},
		{Code: "a !== b", FileName: "case.js"},
		{Code: "a > 0 && b > 0 && a !== b", FileName: "case.js"},
		{Code: "a > 0 && b < 0 && a === b", FileName: "case.js"},
		{Code: "a === \"foo\" && b !== \"bar\" && a !== b", FileName: "case.js"},
		{Code: "a !== \"foo\" && b !== \"bar\" && a === b", FileName: "case.js"},
		{Code: "foo() && bar() && a === b", FileName: "case.js"},
		{Code: "a && b && a === b", FileName: "case.js"},
		{Code: "a === b && b === c && a === c", FileName: "case.js"},
		{Code: "a?.x > 0 && b?.x > 0 && a?.x === b?.x", FileName: "case.js"},
		{Code: "a > 0 && a > 0 && a === a", FileName: "case.js"},
		{Code: "a > 0 && c > 0 && a === b", FileName: "case.js"},
		{Code: "a == 0 && b == 0 && a === b", FileName: "case.js"},
		{Code: "a > foo() && b > foo() && a === b", FileName: "case.js"},
		{Code: "if (a === b && a !== 'x' && isEnabled() && b !== 'x') {}", FileName: "case.js"},
		{Code: "a > 0 && b < 10 && a === b", FileName: "case.js"},
		{Code: "a[f()] > 0 && b[f()] > 0 && a[f()] === b[f()]", FileName: "case.js"},
		{Code: "const modes = new Set([1]); modes.clear(); value === other && value > (modes.size ? 3 : 2) && other > 3;", FileName: "case.js"},
		{Code: "const modes = new Set([1]); modes.clear(); value === other && value > -(modes.size ? 2 : 3) && other > -2;", FileName: "case.js"},
		{Code: "const modes = new Set([1]); modes.clear(); value === other && value > ((modes.size && 3) || lowerBound) && other > 3;", FileName: "case.js"},
		{Code: "const object = {value: true}; Object.defineProperty(object, \"value\", {get() { return false; }}); value === other && value > (object.value ? 3 : lowerBound) && other > 3;", FileName: "case.js"},
		{Code: "const modes = new Set([1]); modes.clear(); value === other && value > ((modes.size ? 3 : 2) as number) && other > 3;", FileName: "case.ts"},
		// Documentation examples
		{Code: "if (a > 0 && a === b) {}", FileName: "case.js"},
		{Code: "if (b > 2 && a === b) {}", FileName: "case.js"},
		{Code: "if (a !== 'test' && a === b) {}", FileName: "case.js"},
		{Code: "if (a === 'test' && a !== b) {}", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Semicolon insertion
		{
			Code: "foo()\nb !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const previous = value\nb !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "foo() // keep\nb !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "foo();\nb !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (ready)\nb !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "foo()\nb !== 1 && (a as number) === 1 && a !== b", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "foo()\na > 0 && (a > 1) && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// pure strict comparisons retain autofixes
		{
			Code: "const first = 1; const second = 1; first === second && first === 1 && second === 1;", FileName: "case.js",
			Output: []string{"const first = 1; const second = 1; first === second && first === 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 71, EndLine: 1, EndColumn: 83,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const first = 1; const second = 2; first === 1 && first !== second && second !== 1;", FileName: "case.js",
			Output: []string{"const first = 1; const second = 2; first === 1 && first !== second;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 71, EndLine: 1, EndColumn: 83,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const first = 1; const second = 1; const third = 1; first === second && second === third && first === 1 && third === 1;", FileName: "case.js",
			Output: []string{"const first = 1; const second = 1; const third = 1; first === second && second === third && first === 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 108, EndLine: 1, EndColumn: 119,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// implicit ordering coercion retains the existing suggestion
		{
			Code: "const first = 1; const second = 1; first === second && first > 0 && second > 0;", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 69, EndLine: 1, EndColumn: 79,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-redundant-comparison/suggestion", Output: "const first = 1; const second = 1; first === second && first > 0;"},
					},
				},
			},
		},
		// Explicit side effects and wrappers
		{
			Code: "first === second && first === 1 && /* keep */ second === 1;", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 47, EndLine: 1, EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(first === second) && (first === 1) && (second === 1);", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 41, EndLine: 1, EndColumn: 53,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "first === second && first === \"value\" && `value` && second === \"value\";", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 53, EndLine: 1, EndColumn: 71,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "((first as number) === second) && (first === 1) && (second! === 1);", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 53, EndLine: 1, EndColumn: 66,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// Snapshot cases
		{
			Code: "a > 0 && b > 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && b > 2 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < c && b < c && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a !== \"test\" && b !== \"test\" && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === \"test\" && b === \"test\" && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === \"test\" && b !== \"test\" && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a !== \"test\" && b === \"test\" && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "b > 0 && a > 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a >= 5 && b > 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && b >= 5 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && b >= 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > -2 && b > 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < 10 && b < 100 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "obj.x > 0 && obj.y > 0 && obj.x === obj.y", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a[key] > 0 && b[key] > 0 && a[key] === b[key]", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && b > 0 && a === b && c > 0", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (a === b && a !== 'x' && b !== 'x' && isEnabled()) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (isEnabled() && a === b && a !== 'x' && b !== 'x') {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 44, EndLine: 1, EndColumn: 53,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (a === b && a !== 'x' && b !== 'x' && list.includes(a)) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && c > 0 && a === b && b === c", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "0 < a && 0 < b && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > b && c > b && a === c", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < \"foo\" && b < \"foo\" && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "this.x > 0 && this.y > 0 && this.x === this.y", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a > 0 && b > 0 && a === b) || c", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (\n\ta > 0 &&\n\tb > 0 &&\n\ta === b\n) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 3, Column: 2, EndLine: 3, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > 0 && /* keep */ b > 0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 21, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a as number) > 0 && (b as number) > 0 && a === b", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 22, EndLine: 1, EndColumn: 39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a! > 0 && b! > 0 && a === b", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a satisfies number) > 0 && (b satisfies number) > 0 && a === b", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 29, EndLine: 1, EndColumn: 53,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a !== b && 1 === a && b !== 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && c === d && b === c && a > 5 && d > 3", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 43, EndLine: 1, EndColumn: 48,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && b === c && a === c && a > 5 && c > 3", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 43, EndLine: 1, EndColumn: 48,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && c === a && c > 5 && b > 3", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < 5 && b < 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a <= 3 && b < 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// Documentation examples
		{
			Code: "if (a > 0 && b > 0 && a === b) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (a > 0 && b > 2 && a === b) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 5, EndLine: 1, EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (a !== 'test' && b !== 'test' && a === b) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (a === 'test' && b !== 'test' && a !== b) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, valid, invalid)
}
