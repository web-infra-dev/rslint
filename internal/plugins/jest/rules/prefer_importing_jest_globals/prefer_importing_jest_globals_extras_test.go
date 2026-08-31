package prefer_importing_jest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_importing_jest_globals"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferImportingJestGlobalsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_importing_jest_globals.PreferImportingJestGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized call expression ----
			{Code: `
        import { test, expect } from '@jest/globals';
        (test)('should pass', () => {
          (expect)(true).toBeDefined();
        });
      `},
			// ---- Dimension 4: already-imported via element-access style local usage ----
			{Code: `
        import { describe } from '@jest/globals';
        describe['skip']("suite", () => {});
      `},
			// N/A: optional-chain on Jest globals (`describe?.()`) is not a valid Jest call chain upstream.
			// N/A: type assertion wrappers on the callee are stripped by ParseJestFnCall for member resolution.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Branch lock-in: types=["jest"] with sourceType=commonjs → require autofix ----
			{
				Code: `
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { jest } = require('@jest/globals');
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Options:         []interface{}{map[string]interface{}{`types`: []interface{}{`jest`}}},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 13},
				},
			},
			// ---- Branch lock-in: languageOptions.sourceType=module → import autofix ----
			{
				Code: `
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import { jest } from '@jest/globals';
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `module`},
				Options:         []interface{}{map[string]interface{}{`types`: []interface{}{`jest`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 13},
				},
			},
			// ---- Branch lock-in: languageOptions.sourceType=script forces require even with import ----
			{
				Code: `
        import fs from 'fs';
        describe("suite", () => {
          test("foo");
        })
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        import fs from 'fs';
        describe("suite", () => {
          test("foo");
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `script`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Dimension 4: parenthesized call still reports the head identifier ----
			{
				Code: `
        (describe)("suite", () => {
          (test)("foo");
          (expect)(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        (describe)("suite", () => {
          (test)("foo");
          (expect)(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 10, EndColumn: 18},
				},
			},
			// ---- Dimension 4: member access / computed property on global ----
			{
				Code: `
        describe["skip"]("suite", () => {
          test("foo");
        })
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        describe["skip"]("suite", () => {
          test("foo");
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 17},
				},
			},
			// ---- Branch lock-in: hook type filtering ----
			{
				Code: `
        beforeEach(() => {});
        describe("suite", () => {
          test("foo");
        })
      `,
				Output: []string{`
        const { beforeEach } = require('@jest/globals');
        beforeEach(() => {});
        describe("suite", () => {
          test("foo");
        })
      `},
				Options:         []interface{}{map[string]interface{}{`types`: []interface{}{`hook`}}},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 19},
				},
			},
			// ---- Branch lock-in: file with existing import ⇒ import-form autofix ----
			{
				Code: `
        import fs from 'fs';
        describe("suite", () => {
          test("foo");
        })
      `,
				Output: []string{`
        import { describe, test } from '@jest/globals';
        import fs from 'fs';
        describe("suite", () => {
          test("foo");
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Real-user shape: only expect used as global ----
			{
				Code: `
        import { describe, test } from '@jest/globals';
        describe("suite", () => {
          test("foo", () => {
            expect(1).toBe(1);
          });
        });
      `,
				Output: []string{`
        import { describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo", () => {
            expect(1).toBe(1);
          });
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 5, Column: 13, EndColumn: 19},
				},
			},
			// ---- Branch lock-in: merge ObjectPattern from the matching declarator ----
			// Replacing the VariableStatement drops sibling declarators (upstream parity).
			{
				Code: `
        const x = 1, { describe: context } = require('@jest/globals');
        describe("suite", () => {
          context("inner", () => {
            test("foo");
          });
        })
      `,
				Output: []string{`
        const { describe, describe: context, test } = require('@jest/globals');
        describe("suite", () => {
          context("inner", () => {
            test("foo");
          });
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- AST parity: computed string require key preserves its local alias ----
			{
				Code: `
        const { ["describe"]: context } = require('@jest/globals');
        describe("suite", () => context("inner", () => test("foo")));
      `,
				Output: []string{`
        const { describe, describe: context, test } = require('@jest/globals');
        describe("suite", () => context("inner", () => test("foo")));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- AST parity: computed template require key preserves its local alias ----
			{
				Code:            "\n        const { [`describe`]: context } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      ",
				Output:          []string{"\n        const { describe, describe: context, test } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- AST parity: parentheses around require arguments are transparent ----
			{
				Code: `
        const { ["describe"]: context } = require(('@jest/globals'));
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, describe: context, test } = require('@jest/globals');
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- ESTree parity: parentheses around the require callee are transparent ----
			{
				Code: `
        const { ["describe"]: context } = ((require))('@jest/globals');
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, describe: context, test } = require('@jest/globals');
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Upstream parity: require matching only inspects the first argument ----
			{
				Code: `
        const { ["describe"]: context } = require('@jest/globals', extra);
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, describe: context, test } = require('@jest/globals');
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Upstream parity: optional require calls are not merge targets ----
			{
				Code: `
        const { ["describe"]: context } = require?.('@jest/globals');
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        const { ["describe"]: context } = require?.('@jest/globals');
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// Keep parenthesized arguments from bypassing the optional-call guard.
			{
				Code: `
        const { ["describe"]: context } = require?.(('@jest/globals'));
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        const { ["describe"]: context } = require?.(('@jest/globals'));
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// Upstream compares a template module specifier's raw text, not its cooked value.
			{
				Code:            "\n        const { [\"describe\"]: context } = require(`@jest/\\u0067lobals`);\n        describe(\"suite\", () => test(\"foo\"));\n      ",
				Output:          []string{"\n        const { describe, test } = require('@jest/globals');\n        const { [\"describe\"]: context } = require(`@jest/\\u0067lobals`);\n        describe(\"suite\", () => test(\"foo\"));\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// String literals use their cooked value, unlike template module specifiers.
			{
				Code:            "\n        const { [\"describe\"]: context } = require('@jest/\\u0067lobals');\n        describe(\"suite\", () => test(\"foo\"));\n      ",
				Output:          []string{"\n        const { describe, describe: context, test } = require('@jest/globals');\n        describe(\"suite\", () => test(\"foo\"));\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Upstream parity: computed numeric require keys are unsupported ----
			{
				Code: `
        const { [1]: context, [0x2]: another } = require('@jest/globals');
        describe("suite", () => context("inner", () => test("foo")));
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        describe("suite", () => context("inner", () => test("foo")));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Upstream parity: defaulted aliases are unsupported accessor values ----
			{
				Code:            "\n        const { [\"describe\"]: context = fallback, [`test`]: spec = fallback } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      ",
				Output:          []string{"\n        const { describe, test } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- Upstream parity: template require keys preserve their raw value ----
			{
				Code:            "\n        const { [`descr\\u0069be`]: context } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      ",
				Output:          []string{"\n        const { descr\\u0069be: context, describe, test } = require('@jest/globals');\n        describe(\"suite\", () => context(\"inner\", () => test(\"foo\")));\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			// ---- JavaScript parity: generated names sort by UTF-16 code units ----
			{
				Code: `
        const { ["𐐀"]: astral, ["Ａ"]: fullwidth } = require('@jest/globals');
        describe("suite", () => test("foo"));
      `,
				Output: []string{`
        const { describe, test, 𐐀: astral, Ａ: fullwidth } = require('@jest/globals');
        describe("suite", () => test("foo"));
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
		},
	)
}
