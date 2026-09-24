import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
ruleTester.run('no-untyped-mock-factory', {} as never, {
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
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: 'rs.mock("./module", () => ({\n  ...rs.requireActual(\'./module\'),\n  foo: rs.fn()\n}));',
      output:
        'rs.mock<typeof import("./module")>("./module", () => ({\n  ...rs.requireActual(\'./module\'),\n  foo: rs.fn()\n}));',
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "rs.mock('random-num', () => {\n  return rs.fn(() => 42);\n});",
      output:
        "rs.mock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "rs.doMock('random-num', () => {\n  return rs.fn(() => 42);\n});",
      output:
        "rs.doMock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "const moduleToMock = 'random-num';\nrs.mock(moduleToMock, () => {\n  return rs.fn(() => 42);\n});",
      output: null,
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
  ],
});
