import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.ts';

ruleTester.run('prefer-jest-mocked', {} as never, {
  valid: [
    { code: 'foo();', filename },
    { code: 'jest.mocked(foo).mockReturnValue(1);', filename },
    { code: 'bar.mockReturnValue(1);', filename },
    { code: 'sinon.stub(foo).returns(1);', filename },
    { code: 'foo.mockImplementation(() => 1);', filename },
    { code: 'obj.foo();', filename },
    { code: 'mockFn.mockReturnValue(1);', filename },
    { code: `arr[0]();`, filename },
    { code: 'obj.foo.mockReturnValue(1);', filename },
    { code: `jest.spyOn(obj, 'foo').mockReturnValue(1);`, filename },
    { code: '(foo as Mock.jest).mockReturnValue(1);', filename },
    {
      filename,
      code: `
      type MockType = jest.Mock;
      const mockFn = jest.fn();
      (mockFn as MockType).mockReturnValue(1);
    `,
    },
  ],
  invalid: [
    {
      filename,
      code: `(foo as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 18,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as unknown as string as unknown as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 50,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as unknown as jest.Mock as unknown as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 53,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(<jest.Mock>foo).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 16,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as jest.Mock).mockImplementation(1);`,
      output: `(jest.mocked(foo)).mockImplementation(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 18,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as unknown as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 29,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(<jest.Mock>foo as unknown).mockReturnValue(1);`,
      output: `(jest.mocked(foo) as unknown).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 16,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(Obj.foo as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked(Obj.foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 22,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `([].foo as jest.Mock).mockReturnValue(1);`,
      output: `(jest.mocked([].foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 21,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 28,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as jest.MockedFunction).mockImplementation(1);`,
      output: `(jest.mocked(foo)).mockImplementation(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 28,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(foo as unknown as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 39,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(Obj.foo as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(Obj.foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 32,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(new Array(0).fill(null).foo as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(new Array(0).fill(null).foo)).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 52,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(jest.fn(() => foo) as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(jest.fn(() => foo))).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 43,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `const mockedUseFocused = useFocused as jest.MockedFunction<typeof useFocused>;`,
      output: `const mockedUseFocused = jest.mocked(useFocused);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 26,
          line: 1,
          endColumn: 78,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `const filter = (MessageService.getMessage as jest.Mock).mock.calls[0][0];`,
      output: `const filter = (jest.mocked(MessageService.getMessage)).mock.calls[0][0];`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 17,
          line: 1,
          endColumn: 55,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `
        class A {}
        (foo as jest.MockedClass<A>)
      `,
      output: `
        class A {}
        (jest.mocked(foo))
      `,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 2,
          endColumn: 28,
          endLine: 2,
        },
      ],
    },
    {
      filename,
      code: `(foo as jest.MockedObject<{method: () => void}>)`,
      output: `(jest.mocked(foo))`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 48,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `(Obj['foo'] as jest.MockedFunction).mockReturnValue(1);`,
      output: `(jest.mocked(Obj['foo'])).mockReturnValue(1);`,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 2,
          line: 1,
          endColumn: 35,
          endLine: 1,
        },
      ],
    },
    {
      filename,
      code: `
        (
          new Array(100)
            .fill(undefined)
            .map(x => x.value)
            .filter(v => !!v).myProperty as jest.MockedFunction<{
            method: () => void;
          }>
        ).mockReturnValue(1);
      `,
      output: `
        (
          jest.mocked(new Array(100)
            .fill(undefined)
            .map(x => x.value)
            .filter(v => !!v).myProperty)
        ).mockReturnValue(1);
      `,
      options: [],
      errors: [
        {
          messageId: 'useJestMocked',
          column: 3,
          line: 2,
          endColumn: 5,
          endLine: 7,
        },
      ],
    },
  ],
});
