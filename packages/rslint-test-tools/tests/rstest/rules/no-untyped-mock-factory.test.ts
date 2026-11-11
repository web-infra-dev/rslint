import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);
ruleTester.run('no-untyped-mock-factory', {
  valid: [
    { code: "rs.mock('random-number');" },
    {
      code: "rs.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs.mock<typeof import('./module')>('./module', () => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));",
    },
    {
      code: "rs.mock<typeof import('foo')>('bar', () => ({\n  ...rs.requireActual('bar'),\n  foo: rs.fn()\n}));",
    },
    {
      code: "rs.doMock('./module', (): typeof import('./module') => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));",
    },
    {
      code: "rs.mock('../moduleName', function (): typeof import('../moduleName') {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs.mock<() => number>('random-num', () => {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs['doMock']<() => number>('random-num', () => {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs.mock<any>('random-num', () => {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs.mock(\n  '../moduleName',\n  () => {\n    return rs.fn(() => 42)\n  },\n  {virtual: true},\n);",
    },
    {
      code: "rs.mock('../moduleName', function (): (() => number) {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "mockito<() => number>('foo', () => {\n  return rs.fn(() => 42);\n});",
    },
    {
      code: "rs['mock']('random-num', () => {\n  return rs.fn(() => 42);\n});",
    },
  ],
  invalid: [
    {
      code: "rs.mock('../moduleName', () => {\n  return rs.fn(() => 42);\n});",
      output:
        "rs.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return rs.fn(() => 42);\n});",
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message:
            "Add a type parameter to the mock factory such as `typeof import('../moduleName')`",
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'rs.mock("./module", () => ({\n  ...rs.requireActual(\'./module\'),\n  foo: rs.fn()\n}));',
      output:
        'rs.mock<typeof import("./module")>("./module", () => ({\n  ...rs.requireActual(\'./module\'),\n  foo: rs.fn()\n}));',
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message:
            'Add a type parameter to the mock factory such as `typeof import("./module")`',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 4,
        },
      ],
    },
    {
      code: "rs.mock('random-num', () => {\n  return rs.fn(() => 42);\n});",
      output:
        "rs.mock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});",
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message:
            "Add a type parameter to the mock factory such as `typeof import('random-num')`",
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: "rs.doMock('random-num', () => {\n  return rs.fn(() => 42);\n});",
      output:
        "rs.doMock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});",
      errors: [
        {
          messageId: 'addTypeParameterToModuleMock',
          message:
            "Add a type parameter to the mock factory such as `typeof import('random-num')`",
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: "const moduleToMock = 'random-num';\nrs.mock(moduleToMock, () => {\n  return rs.fn(() => 42);\n});",
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
