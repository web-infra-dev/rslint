package prefer_importing_jest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_importing_jest_globals"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferImportingJestGlobalsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_importing_jest_globals.PreferImportingJestGlobalsRule,
		[]rule_tester.ValidTestCase{
			{Code: `
        // with import
        import { test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `},
			{Code: `
        // with import
        import { 'test' as test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `},
			{Code: `
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `, Options: []interface{}{map[string]interface{}{`types`: []interface{}{`jest`}}}},
			{Code: `
        const { it } = require('@jest/globals');
        it('should pass', () => {
            expect(true).toBeDefined();
        });
      `, Options: []interface{}{map[string]interface{}{`types`: []interface{}{`test`}}}},
			{Code: `
        // with require
        const { test, expect } = require('@jest/globals');
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `},
			{Code: "\n        const { test, expect } = require(`@jest/globals`);\n        test('should pass', () => {\n            expect(true).toBeDefined();\n        });\n      "},
			{Code: `
        import { it as itChecks } from '@jest/globals';
        itChecks("foo");
      `},
			{Code: `
        import { 'it' as itChecks } from '@jest/globals';
        itChecks("foo");
      `},
			{Code: `
        const { test } = require('@jest/globals');
        test("foo");
      `},
			{Code: `
        const { test } = require('my-test-library');
        test("foo");
      `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        import describe from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import { describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        import { describe as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import { describe as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        import { describe as context } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
				Output: []string{`
        import { describe, describe as context, expect, test } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        import { 'describe' as describe } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import { 'describe' as describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        import { 'describe' as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import { 'describe' as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
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
			{
				Code: `
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { test } from '@jest/globals';
        import { xit } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { describe, expect, test } from '@jest/globals';
        import { xit } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 7, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        console.log('hello');
        import * as fs from 'fs';
        const { test, 'describe': describe } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        console.log('hello');
        import * as fs from 'fs';
        import { describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 7, Column: 11, EndColumn: 17},
				},
			},
			{
				Code: `
        console.log('hello');
        import jestGlobals from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        console.log('hello');
        import { describe, expect, jestGlobals, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
				Output: []string{`
        import { describe, test } from '@jest/globals';
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        const {describe} = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        const {describe: context} = require('@jest/globals');
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { describe: context, expect, test } = require('@jest/globals');
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        const {describe: context} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
				Output: []string{`
        const { describe, describe: context, expect, test } = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        const {describe: []} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: "\n        const {describe} = require(`@jest/globals`);\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      ",
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 11, EndColumn: 15},
				},
			},
			{
				Code:            "\n        const source = 'globals';\n        const {describe} = require(`@jest/${source}`);\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      ",
				Output:          []string{"\n        const { expect, test } = require('@jest/globals');\n        const source = 'globals';\n        const {describe} = require(`@jest/${source}`);\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 5, Column: 11, EndColumn: 15},
				},
			},
			{
				Code: `
        const { [() => {}]: it } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        console.log('hello');
        const fs = require('fs');
        const { test, 'describe': describe } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        console.log('hello');
        const fs = require('fs');
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 7, Column: 11, EndColumn: 17},
				},
			},
			{
				Code: `
        console.log('hello');
        const jestGlobals = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        console.log('hello');
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 4, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
				Output: []string{`
        const { describe, test } = require('@jest/globals');
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 17},
				},
			},
			{
				// Shebang must be at byte 0 — avoid a leading newline from a raw string.
				Code:            "#!/usr/bin/env node\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n",
				Output:          []string{"#!/usr/bin/env node\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n"},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 2, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        // with comment above
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        // with comment above
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        'use strict';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
				Output: []string{`
        'use strict';
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code:            "\n        `use strict`;\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      ",
				Output:          []string{"\n        `use strict`;\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 9, EndColumn: 17},
				},
			},
			{
				Code: `
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `,
				Output: []string{`
        const { describe, expect, jest, test } = require('@jest/globals');
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `commonjs`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 25, EndColumn: 29},
				},
			},
			{
				Code: `
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `,
				Output: []string{`
        import { describe, expect, jest, test } from '@jest/globals';
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `},
				LanguageOptions: rule.LanguageOptions{SourceType: `module`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `preferImportingJestGlobal`, Line: 3, Column: 25, EndColumn: 29},
				},
			},
		},
	)
}
