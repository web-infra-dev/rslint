// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-jest-mocked', {
  valid: [
    ...(
      [
        'foo();',
        'jest.mocked(foo).mockReturnValue(1);',
        'bar.mockReturnValue(1);',
        'sinon.stub(foo).returns(1);',
        'foo.mockImplementation(() => 1);',
        'obj.foo();',
        'mockFn.mockReturnValue(1);',
        'arr[0]();',
        'obj.foo.mockReturnValue(1);',
        "jest.spyOn(obj, 'foo').mockReturnValue(1);",
        '(foo as Mock.jest).mockReturnValue(1);',
        `
      type MockType = jest.Mock;
      const mockFn = jest.fn();
      (mockFn as MockType).mockReturnValue(1);
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
      filename: 'src/virtual.ts',
    })),
    ...(
      [
        'foo();',
        'jest.mocked(foo).mockReturnValue(1);',
        'bar.mockReturnValue(1);',
        'sinon.stub(foo).returns(1);',
        'foo.mockImplementation(() => 1);',
        'obj.foo();',
        'mockFn.mockReturnValue(1);',
        'arr[0]();',
        'obj.foo.mockReturnValue(1);',
        "jest.spyOn(obj, 'foo').mockReturnValue(1);",
        '(foo as Mock.jest).mockReturnValue(1);',
        `type MockType = jest.Mock;
const mockFn = jest.fn();
(mockFn as MockType).mockReturnValue(1);`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-jest-mocked / valid ${index}`,
      code,
    })),
  ],
  invalid: [
    ...(
      [
        [
          '(foo as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          18,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as unknown as string as unknown as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          50,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as unknown as jest.Mock as unknown as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          53,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(<jest.Mock>foo).mockReturnValue(1);',
          1,
          2,
          1,
          16,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.Mock).mockImplementation(1);',
          1,
          2,
          1,
          18,
          '(jest.mocked(foo)).mockImplementation(1);',
        ],
        [
          '(foo as unknown as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          29,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(<jest.Mock>foo as unknown).mockReturnValue(1);',
          1,
          2,
          1,
          16,
          '(jest.mocked(foo) as unknown).mockReturnValue(1);',
        ],
        [
          '(Obj.foo as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          22,
          '(jest.mocked(Obj.foo)).mockReturnValue(1);',
        ],
        [
          '([].foo as jest.Mock).mockReturnValue(1);',
          1,
          2,
          1,
          21,
          '(jest.mocked([].foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.MockedFunction).mockReturnValue(1);',
          1,
          2,
          1,
          28,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.MockedFunction).mockImplementation(1);',
          1,
          2,
          1,
          28,
          '(jest.mocked(foo)).mockImplementation(1);',
        ],
        [
          '(foo as unknown as jest.MockedFunction).mockReturnValue(1);',
          1,
          2,
          1,
          39,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(Obj.foo as jest.MockedFunction).mockReturnValue(1);',
          1,
          2,
          1,
          32,
          '(jest.mocked(Obj.foo)).mockReturnValue(1);',
        ],
        [
          '(new Array(0).fill(null).foo as jest.MockedFunction).mockReturnValue(1);',
          1,
          2,
          1,
          52,
          '(jest.mocked(new Array(0).fill(null).foo)).mockReturnValue(1);',
        ],
        [
          '(jest.fn(() => foo) as jest.MockedFunction).mockReturnValue(1);',
          1,
          2,
          1,
          43,
          '(jest.mocked(jest.fn(() => foo))).mockReturnValue(1);',
        ],
        [
          'const mockedUseFocused = useFocused as jest.MockedFunction<typeof useFocused>;',
          1,
          26,
          1,
          78,
          'const mockedUseFocused = jest.mocked(useFocused);',
        ],
        [
          'const filter = (MessageService.getMessage as jest.Mock).mock.calls[0][0];',
          1,
          17,
          1,
          55,
          'const filter = (jest.mocked(MessageService.getMessage)).mock.calls[0][0];',
        ],
        [
          `
        class A {}
        (foo as jest.MockedClass<A>)
     \u0020`,
          3,
          10,
          3,
          36,
          `
        class A {}
        (jest.mocked(foo))
     \u0020`,
        ],
        [
          '(foo as jest.MockedObject<{method: () => void}>)',
          1,
          2,
          1,
          48,
          '(jest.mocked(foo))',
        ],
        [
          "(Obj['foo'] as jest.MockedFunction).mockReturnValue(1);",
          1,
          2,
          1,
          35,
          "(jest.mocked(Obj['foo'])).mockReturnValue(1);",
        ],
        [
          `
        (
          new Array(100)
            .fill(undefined)
            .map(x => x.value)
            .filter(v => !!v).myProperty as jest.MockedFunction<{
            method: () => void;
          }>
        ).mockReturnValue(1);
     \u0020`,
          3,
          11,
          8,
          13,
          `
        (
          jest.mocked(new Array(100)
            .fill(undefined)
            .map(x => x.value)
            .filter(v => !!v).myProperty)
        ).mockReturnValue(1);
     \u0020`,
        ],
      ] as const
    ).map(
      ([code, line, column, endLine, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        filename: 'src/virtual.ts',
        options: [],
        errors: [
          {
            messageId: 'useJestMocked',
            message: 'Prefer `jest.mocked()` over type assertions',
            line,
            column,
            endLine,
            endColumn,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          '(foo as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 0',
          1,
          2,
          1,
          18,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as unknown as string as unknown as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 1',
          1,
          2,
          1,
          50,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as unknown as jest.Mock as unknown as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 2',
          1,
          2,
          1,
          53,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.Mock).mockImplementation(1);',
          'upstream prefer-jest-mocked / invalid 4',
          1,
          2,
          1,
          18,
          '(jest.mocked(foo)).mockImplementation(1);',
        ],
        [
          '(foo as unknown as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 5',
          1,
          2,
          1,
          29,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(Obj.foo as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 7',
          1,
          2,
          1,
          22,
          '(jest.mocked(Obj.foo)).mockReturnValue(1);',
        ],
        [
          '([].foo as jest.Mock).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 8',
          1,
          2,
          1,
          21,
          '(jest.mocked([].foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.MockedFunction).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 9',
          1,
          2,
          1,
          28,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(foo as jest.MockedFunction).mockImplementation(1);',
          'upstream prefer-jest-mocked / invalid 10',
          1,
          2,
          1,
          28,
          '(jest.mocked(foo)).mockImplementation(1);',
        ],
        [
          '(foo as unknown as jest.MockedFunction).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 11',
          1,
          2,
          1,
          39,
          '(jest.mocked(foo)).mockReturnValue(1);',
        ],
        [
          '(Obj.foo as jest.MockedFunction).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 12',
          1,
          2,
          1,
          32,
          '(jest.mocked(Obj.foo)).mockReturnValue(1);',
        ],
        [
          '(new Array(0).fill(null).foo as jest.MockedFunction).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 13',
          1,
          2,
          1,
          52,
          '(jest.mocked(new Array(0).fill(null).foo)).mockReturnValue(1);',
        ],
        [
          '(jest.fn(() => foo) as jest.MockedFunction).mockReturnValue(1);',
          'upstream prefer-jest-mocked / invalid 14',
          1,
          2,
          1,
          43,
          '(jest.mocked(jest.fn(() => foo))).mockReturnValue(1);',
        ],
        [
          'const mockedUseFocused = useFocused as jest.MockedFunction<typeof useFocused>;',
          'upstream prefer-jest-mocked / invalid 15',
          1,
          26,
          1,
          78,
          'const mockedUseFocused = jest.mocked(useFocused);',
        ],
        [
          'const filter = (MessageService.getMessage as jest.Mock).mock.calls[0][0];',
          'upstream prefer-jest-mocked / invalid 16',
          1,
          17,
          1,
          55,
          'const filter = (jest.mocked(MessageService.getMessage)).mock.calls[0][0];',
        ],
        [
          `class A {}
(foo as jest.MockedClass<A>)`,
          'upstream prefer-jest-mocked / invalid 17',
          2,
          2,
          2,
          28,
          `class A {}
(jest.mocked(foo))`,
        ],
        [
          '(foo as jest.MockedObject<{method: () => void}>)',
          'upstream prefer-jest-mocked / invalid 18',
          1,
          2,
          1,
          48,
          '(jest.mocked(foo))',
        ],
        [
          "(Obj['foo'] as jest.MockedFunction).mockReturnValue(1);",
          'upstream prefer-jest-mocked / invalid 19',
          1,
          2,
          1,
          35,
          "(jest.mocked(Obj['foo'])).mockReturnValue(1);",
        ],
        [
          `(
  new Array(100)
    .fill(undefined)
    .map(x => x.value)
    .filter(v => !!v).myProperty as jest.MockedFunction<{
    method: () => void;
  }>
).mockReturnValue(1);`,
          'upstream prefer-jest-mocked / invalid 20',
          2,
          3,
          7,
          5,
          `(
  jest.mocked(new Array(100)
    .fill(undefined)
    .map(x => x.value)
    .filter(v => !!v).myProperty)
).mockReturnValue(1);`,
        ],
      ] as const
    ).map(
      ([code, name, line, column, endLine, endColumn, output]): TestCase => ({
        name,
        code,
        options: [],
        errors: [
          {
            messageId: 'useJestMocked',
            message: 'Prefer `jest.mocked()` over type assertions',
            line,
            column,
            endLine,
            endColumn,
          },
        ],
        output,
      }),
    ),
  ],
});
