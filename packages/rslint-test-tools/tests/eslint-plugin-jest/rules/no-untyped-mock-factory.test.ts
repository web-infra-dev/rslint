import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
ruleTester.run('no-untyped-mock-factory', {} as never, {
  valid: [
    { code: "jest.mock('random-number');" },
    {
      code: "jest.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "jest.mock<typeof import('./module')>('./module', () => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));",
    },
    {
      code: "jest.mock<typeof import('foo')>('bar', () => ({\n  ...jest.requireActual('bar'),\n  foo: jest.fn()\n}));",
    },
    {
      code: "jest.doMock('./module', (): typeof import('./module') => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));",
    },
    {
      code: "jest.mock('../moduleName', function (): typeof import('../moduleName') {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "jest.mock<() => number>('random-num', () => {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "jest['doMock']<() => number>('random-num', () => {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "jest.mock<any>('random-num', () => {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "jest.mock(\n  '../moduleName',\n  () => {\n    return jest.fn(() => 42)\n  },\n  {virtual: true},\n);",
    },
    {
      code: "jest.mock('../moduleName', function (): (() => number) {\n  return jest.fn(() => 42);\n});",
    },
    {
      code: "mockito<() => number>('foo', () => {\n  return jest.fn(() => 42);\n});",
    },
  ],
  invalid: [
    {
      code: "jest.mock('../moduleName', () => {\n  return jest.fn(() => 42);\n});",
      output:
        "jest.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return jest.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: 'jest.mock("./module", () => ({\n  ...jest.requireActual(\'./module\'),\n  foo: jest.fn()\n}));',
      output:
        'jest.mock<typeof import("./module")>("./module", () => ({\n  ...jest.requireActual(\'./module\'),\n  foo: jest.fn()\n}));',
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "jest.mock('random-num', () => {\n  return jest.fn(() => 42);\n});",
      output:
        "jest.mock<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "jest.doMock('random-num', () => {\n  return jest.fn(() => 42);\n});",
      output:
        "jest.doMock<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "jest['mock']('random-num', () => {\n  return jest.fn(() => 42);\n});",
      output:
        "jest['mock']<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});",
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
    {
      code: "const moduleToMock = 'random-num';\njest.mock(moduleToMock, () => {\n  return jest.fn(() => 42);\n});",
      output: null,
      errors: [{ messageId: 'addTypeParameterToModuleMock' }],
    },
  ],
});
