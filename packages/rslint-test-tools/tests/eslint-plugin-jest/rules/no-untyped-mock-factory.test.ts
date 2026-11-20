// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-untyped-mock-factory', {
  valid: [
    ...(
      [
        "jest.mock('random-number');",
        `jest.mock<typeof import('../moduleName')>('../moduleName', () => {
  return jest.fn(() => 42);
});`,
        `jest.mock<typeof import('./module')>('./module', () => ({
  ...jest.requireActual('./module'),
  foo: jest.fn()
}));`,
        `jest.mock<typeof import('foo')>('bar', () => ({
  ...jest.requireActual('bar'),
  foo: jest.fn()
}));`,
        `jest.doMock('./module', (): typeof import('./module') => ({
  ...jest.requireActual('./module'),
  foo: jest.fn()
}));`,
        `jest.mock('../moduleName', function (): typeof import('../moduleName') {
  return jest.fn(() => 42);
});`,
        `jest.mock<() => number>('random-num', () => {
  return jest.fn(() => 42);
});`,
        `jest['doMock']<() => number>('random-num', () => {
  return jest.fn(() => 42);
});`,
        `jest.mock<any>('random-num', () => {
  return jest.fn(() => 42);
});`,
        `jest.mock(
  '../moduleName',
  () => {
    return jest.fn(() => 42)
  },
  {virtual: true},
);`,
        `jest.mock('../moduleName', function (): (() => number) {
  return jest.fn(() => 42);
});`,
        `mockito<() => number>('foo', () => {
  return jest.fn(() => 42);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
  ],
  invalid: [
    ...(
      [
        [
          `jest.mock('../moduleName', () => {
  return jest.fn(() => 42);
});`,
          "Add a type parameter to the mock factory such as `typeof import('../moduleName')`",
          3,
          3,
          `jest.mock<typeof import('../moduleName')>('../moduleName', () => {
  return jest.fn(() => 42);
});`,
        ],
        [
          `jest.mock("./module", () => ({
  ...jest.requireActual('./module'),
  foo: jest.fn()
}));`,
          'Add a type parameter to the mock factory such as `typeof import("./module")`',
          4,
          4,
          `jest.mock<typeof import("./module")>("./module", () => ({
  ...jest.requireActual('./module'),
  foo: jest.fn()
}));`,
        ],
        [
          `jest.mock('random-num', () => {
  return jest.fn(() => 42);
});`,
          "Add a type parameter to the mock factory such as `typeof import('random-num')`",
          3,
          3,
          `jest.mock<typeof import('random-num')>('random-num', () => {
  return jest.fn(() => 42);
});`,
        ],
        [
          `jest.doMock('random-num', () => {
  return jest.fn(() => 42);
});`,
          "Add a type parameter to the mock factory such as `typeof import('random-num')`",
          3,
          3,
          `jest.doMock<typeof import('random-num')>('random-num', () => {
  return jest.fn(() => 42);
});`,
        ],
        [
          `jest['mock']('random-num', () => {
  return jest.fn(() => 42);
});`,
          "Add a type parameter to the mock factory such as `typeof import('random-num')`",
          3,
          3,
          `jest['mock']<typeof import('random-num')>('random-num', () => {
  return jest.fn(() => 42);
});`,
        ],
      ] as const
    ).map(([code, message, endLine, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message,
          line: 1,
          column: 1,
          endLine,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 5',
      code: `const moduleToMock = 'random-num';
jest.mock(moduleToMock, () => {
  return jest.fn(() => 42);
});`,
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message:
            'Add a type parameter to the mock factory such as `typeof import(./module-name)`',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
  ],
});
