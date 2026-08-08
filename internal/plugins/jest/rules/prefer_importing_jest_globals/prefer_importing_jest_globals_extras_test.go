package prefer_importing_jest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_importing_jest_globals"
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
			// ---- Branch lock-in: types=["jest"] without sourceType / import → require autofix ----
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
				Options: []interface{}{map[string]interface{}{`types`: []interface{}{`jest`}}},
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
				SourceType: `module`,
				Options:    []interface{}{map[string]interface{}{`types`: []interface{}{`jest`}}},
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
				SourceType: `script`,
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
				Options: []interface{}{map[string]interface{}{`types`: []interface{}{`hook`}}},
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
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
		},
	)
}
