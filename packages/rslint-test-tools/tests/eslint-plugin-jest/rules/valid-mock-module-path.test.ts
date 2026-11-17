// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(
  fixtureConfig(
    'internal/plugins/jest/rules/valid_mock_module_path/testdata/fixtures.txtar',
  ),
).run('valid-mock-module-path', {
  valid: [
    ...(
      [
        'jest.mock("./fixtures/module")',
        'jest.mock("./fixtures/module", () => {})',
        'jest.mock()',
        'jest.doMock("./fixtures/module", () => {})',
        'describe("foo", () => {});',
        'jest.doMock("./fixtures/module")',
        'jest.mock("./fixtures/module/foo.ts")',
        'jest.doMock("./fixtures/module/foo.ts")',
        'jest.mock("./fixtures/module/foo.js")',
        'jest.doMock("./fixtures/module/foo.js")',
        'jest.mock("eslint")',
        'jest.doMock("eslint")',
        'jest.mock("child_process")',
        'jest.mock(() => {})',
        `const a = "../module/does/not/exist";
jest.mock(a);`,
        'jest.mock("./fixtures/module/jsx/foo")',
        'jest.mock("./fixtures/module/tsx/foo")',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
    })),
    {
      name: 'run 0 valid 17',
      code: 'jest.mock("./fixtures/module/tsx/foo")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.jsx'],
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'jest.mock("./fixtures/module/bar")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.json'],
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'jest.mock("./fixtures/module/bar")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.css'],
        },
      ],
    },
    ...(
      [
        'jest.mock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { ...{} })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: [] })',
        'jest.mock("../module/does/not/exist", undefined, { virtual: true })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: true })',
        'jest.doMock("../module/does/not/exist", undefined, { "virtual": true })',
        'jest.doMock("../module/does/not/exist", undefined, { ["virtual"]: true })',
        'jest.doMock("../module/does/not/exist", undefined, { [`virtual`]: true })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: 1 })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: "yes" })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 20}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
    })),
    ...(
      [
        'jest.mock("./fixtures/module")',
        'jest.mock("./fixtures/module", () => {})',
        'jest.mock()',
        'jest.doMock("./fixtures/module", () => {})',
        'describe("foo", () => {});',
        'jest.doMock("./fixtures/module")',
        'jest.mock("./fixtures/module/foo.ts")',
        'jest.doMock("./fixtures/module/foo.ts")',
        'jest.mock("./fixtures/module/foo.js")',
        'jest.doMock("./fixtures/module/foo.js")',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-mock-module-path / valid ${index}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      languageOptions: {
        ecmaVersion: 2018,
      },
    })),
    ...(
      [
        'jest.mock("eslint")',
        'jest.doMock("eslint")',
        'jest.mock("child_process")',
        'jest.mock(() => {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-mock-module-path / valid ${index + 10}`,
      code,
      languageOptions: {
        ecmaVersion: 2018,
      },
    })),
    ...(
      [
        `const a = "../module/does/not/exist";
jest.mock(a);`,
        'jest.mock("./fixtures/module/jsx/foo")',
        'jest.mock("./fixtures/module/tsx/foo")',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-mock-module-path / valid ${index + 14}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      languageOptions: {
        ecmaVersion: 2018,
      },
    })),
    {
      name: 'upstream valid-mock-module-path / valid 17',
      code: 'jest.mock("./fixtures/module/tsx/foo")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.jsx'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2018,
      },
    },
    {
      name: 'upstream valid-mock-module-path / valid 18',
      code: 'jest.mock("./fixtures/module/bar")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.json'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2018,
      },
    },
    {
      name: 'upstream valid-mock-module-path / valid 19',
      code: 'jest.mock("./fixtures/module/bar")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.css'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2018,
      },
    },
    ...(
      [
        'jest.mock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { ...{} })',
        'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: [] })',
        'jest.mock("../module/does/not/exist", undefined, { virtual: true })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: true })',
        'jest.doMock("../module/does/not/exist", undefined, { "virtual": true })',
        'jest.doMock("../module/does/not/exist", undefined, { ["virtual"]: true })',
        'jest.doMock("../module/does/not/exist", undefined, { [`virtual`]: true })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: 1 })',
        'jest.doMock("../module/does/not/exist", undefined, { virtual: "yes" })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-mock-module-path / valid ${index + 20}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      languageOptions: {
        ecmaVersion: 2018,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          "jest.mock('../module/does/not/exist')",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          38,
        ],
        [
          'jest.mock("../file/does/not/exist.ts")',
          'Module path "../file/does/not/exist.ts" does not exist or is not exported',
          39,
        ],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 2',
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.tsx'],
        },
      ],
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message:
            'Module path "./fixtures/module/foo.jsx" does not exist or is not exported',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: undefined,
        },
      ],
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message:
            'Module path "./fixtures/module/foo.jsx" does not exist or is not exported',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    ...(
      [
        [
          'jest.mock("@doesnotexist/module")',
          'Module path "@doesnotexist/module" does not exist or is not exported',
          1,
          34,
        ],
        [
          'jest.mock("jest-util/build/isInteractive")',
          'Module path "jest-util/build/isInteractive" does not exist or is not exported',
          1,
          43,
        ],
        [
          'jest.mock("jackspeak/dist/commonjs/parse-args.js")',
          'Module path "jackspeak/dist/commonjs/parse-args.js" does not exist or is not exported',
          1,
          51,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, {})",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          53,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, { assumeExists: true })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          73,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, { virtual: false })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          69,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { virtual: false })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          71,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { virtual: 0 })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          67,
        ],
        [
          `const virtual = false;

jest.doMock('../module/does/not/exist', undefined, { virtual })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          64,
        ],
        [
          `const virtual = true;

jest.doMock('../module/does/not/exist', undefined, { virtual })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          64,
        ],
        [
          `const prop = 'virtual';

jest.doMock('../module/does/not/exist', undefined, { [prop]: true })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          69,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { ...{ virtual: true } })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          77,
        ],
      ] as const
    ).map(([code, message, line, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message,
          line,
          column: 1,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          "jest.mock('../module/does/not/exist')",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          38,
        ],
        [
          'jest.mock("../file/does/not/exist.ts")',
          'Module path "../file/does/not/exist.ts" does not exist or is not exported',
          39,
        ],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `upstream valid-mock-module-path / invalid ${index}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      languageOptions: {
        ecmaVersion: 2018,
      },
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-mock-module-path / invalid 2',
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: ['.tsx'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2018,
      },
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message:
            'Module path "./fixtures/module/foo.jsx" does not exist or is not exported',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream valid-mock-module-path / invalid 3',
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      options: [
        {
          moduleFileExtensions: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2018,
      },
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message:
            'Module path "./fixtures/module/foo.jsx" does not exist or is not exported',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    ...(
      [
        [
          'jest.mock("@doesnotexist/module")',
          'Module path "@doesnotexist/module" does not exist or is not exported',
          1,
          34,
        ],
        [
          'jest.mock("jest-util/build/isInteractive")',
          'Module path "jest-util/build/isInteractive" does not exist or is not exported',
          1,
          43,
        ],
        [
          'jest.mock("jackspeak/dist/commonjs/parse-args.js")',
          'Module path "jackspeak/dist/commonjs/parse-args.js" does not exist or is not exported',
          1,
          51,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, {})",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          53,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, { assumeExists: true })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          73,
        ],
        [
          "jest.mock('../module/does/not/exist', undefined, { virtual: false })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          69,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { virtual: false })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          71,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { virtual: 0 })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          67,
        ],
        [
          `const virtual = false;

jest.doMock('../module/does/not/exist', undefined, { virtual })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          64,
        ],
        [
          `const virtual = true;

jest.doMock('../module/does/not/exist', undefined, { virtual })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          64,
        ],
        [
          `const prop = 'virtual';

jest.doMock('../module/does/not/exist', undefined, { [prop]: true })`,
          "Module path '../module/does/not/exist' does not exist or is not exported",
          3,
          69,
        ],
        [
          "jest.doMock('../module/does/not/exist', undefined, { ...{ virtual: true } })",
          "Module path '../module/does/not/exist' does not exist or is not exported",
          1,
          77,
        ],
      ] as const
    ).map(([code, message, line, endColumn], index): TestCase => ({
      name: `upstream valid-mock-module-path / invalid ${index + 4}`,
      code,
      filename: 'src/rules/__tests__/valid-mock-module-path.test.ts',
      languageOptions: {
        ecmaVersion: 2018,
      },
      errors: [
        {
          messageId: 'invalidMockModulePath',
          message,
          line,
          column: 1,
          endLine: line,
          endColumn,
        },
      ],
    })),
  ],
});
