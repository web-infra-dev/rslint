// Upstream reference: eslint-plugin-jest@29.16.1; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-mock-return-shorthand', {
  valid: [
    ...(
      [
        'jest.fn().mockReturnValue(42)',
        'jest.fn(() => 42)',
        'aVariable.mockImplementation()',
        'jest.fn().mockImplementation(async () => 1);',
        'aVariable.mockImplementation(() => value++)',
        "jest.spyOn(Thingy, 'method').mockImplementation(param => param * 2);",
        `let value = 1;

aVariable.mockImplementation(() => ({ value }));`,
        'jest.fn().mockImplementation(() => Promise.reject(13))',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'describe()',
        'it()',
        'describe.skip()',
        'it.skip()',
        'test()',
        'test.skip()',
        'var appliedOnly = describe.only; appliedOnly.apply(describe)',
        'var calledOnly = it.only; calledOnly.call(it)',
        'it.each()()',
        'it.each`table`()',
        'test.each()()',
        'test.each`table`()',
        'test.concurrent()',
        'jest.fn().mockReturnValue(42)',
        'jest.fn(() => Promise.resolve(42))',
        'jest.fn(() => 42)',
        'jest.fn(() => ({}))',
        'aVariable.mockImplementation',
        'aVariable.mockImplementation()',
        'jest.fn().mockImplementation(async () => 1);',
        'jest.fn().mockImplementation(async function () {});',
        `jest.fn().mockImplementation(async function () {
  return 42;
});`,
        `aVariable.mockImplementation(() => {
  if (true) {
    return 1;
  }

  return 2;
});`,
        'aVariable.mockImplementation(() => value++)',
        'aVariable.mockImplementationOnce(() => --value)',
        `const aValue = 0;
aVariable.mockImplementation(() => {
  return aValue++;
});`,
        `aVariable.mockImplementation(() => {
  aValue += 1;

  return aValue;
});`,
        `aVariable.mockImplementation(() => {
  aValue++;

  return aValue;
});`,
        'aVariable.mockReturnValue()',
        'aVariable.mockReturnValue(1)',
        'aVariable.mockReturnValue("hello world")',
        "jest.spyOn(Thingy, 'method').mockImplementation(param => param * 2);",
        "jest.spyOn(Thingy, 'method').mockImplementation(param => true ? param : 0);",
        `aVariable.mockImplementation(() => {
  const value = new Date();

  return Promise.resolve(value);
});`,
        `aVariable.mockImplementation(() => {
  throw new Error('oh noes!');
});`,
        'aVariable.mockImplementation(() => { /* do something */ });',
        `aVariable.mockImplementation(() => {
  const x = 1;

  console.log(x + 2);
});`,
        'aVariable.mockReturnValue(Promise.all([1, 2, 3]));',
        `let currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);

// stuff happens

currentX++;

// more stuff happens`,
        `let currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
        `let currentX = 0;
currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
        `var currentX = 0;
currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
        `var currentX = 0;
var currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
        `let doSomething = () => {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`,
        `let currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => {
  currentX += 1;

  return currentX;
});`,
        `const currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => {
  console.log('returning', currentX);

  return currentX;
});`,
        `let value = 1;

jest.fn().mockImplementation(() => ({ value }));`,
        `let value = 1;

aVariable.mockImplementation(() => [value]);`,
        `var value = 1;

aVariable.mockImplementation(() => [0, value, 2]);`,
        `let value = 1;

aVariable.mockImplementation(() => value + 1);`,
        `let value = 1;

aVariable.mockImplementation(() => 1 - value);`,
        `var value = 1;

aVariable.mockImplementation(() => {
  return { value: value + 1 };
});`,
        `var value = 1;

aVariable.mockImplementation(() => value * value + 1);
aVariable.mockImplementation(() => 1 + value / 2);
aVariable.mockImplementation(() => (1 + value) / 2);
aVariable.mockImplementation(() => {
  return { value: value + 1 };
});`,
        `let value = 1;

aVariable.mockImplementation(function () {
  return { items: [value] };
});`,
        `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: { value },
  }
});`,
        `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [{
    type: 'object',
    with: [1, 2, value],
  }]
});`,
        `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ];
});`,
        `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 3]},
    {type: 'object', with: [1, value]}
  ];
});`,
        `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  }
});`,
        `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }
});`,
        `let value = 1;

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]
});`,
        `let value = 1;

aVariable.mockImplementation(() => value & 1);
aVariable.mockImplementation(() => value | 1);
aVariable.mockImplementation(() => 1 & value);
aVariable.mockImplementation(() => 1 | value);`,
        `let value = 1;

aVariable.mockImplementation(() => !value);
aVariable.mockImplementation(() => ~value);
aVariable.mockImplementation(() => typeof value);`,
        `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx & my);
aVariable.mockImplementation(() => my | mx);`,
        `let value = 1;

aVariable.mockImplementation(() => value || 0);
aVariable.mockImplementation(() => 1 && value);
aVariable.mockImplementation(() => 1 ?? value);
aVariable.mockImplementation(() => 1 ?? (value && 0));`,
        `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => my ?? mx);
aVariable.mockImplementation(() => mx ?? (7 && my));`,
        `let value = [1];

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, ...value],
      },
    },
  }]
});`,
        `let value = 1;

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, ...[value]],
      },
    },
  }]
});`,
        `let obj = {};

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    ...obj,
  }
});`,
        `let value = 1;

aVariable.mockImplementation(function () {
  function mx() {
    return value;
  }
  return mx();
});`,
        `let value = 1;

jest.fn().mockImplementation(() => new Mx(value));
jest.fn().mockImplementation(() => new Mx(() => value));
jest.fn().mockImplementation(() => new Mx(() => { return value }));`,
        `let value = 1;

jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx?.(value));
jest.fn().mockImplementation(() => mx(value).my());
jest.fn().mockImplementation(() => mx(value).my);
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx?.my(value));
jest.fn().mockImplementation(() => mx?.my?.(value));
jest.fn().mockImplementation(() => mx.my?.(value));
jest.fn().mockImplementation(() => mx().my(value));
jest.fn().mockImplementation(() => mx()?.my(value));
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx(value).my(value));
jest.fn().mockImplementation(() => mx?.(value)?.my?.(value));
jest.fn().mockImplementation(() => new Mx().add(value));
jest.fn().mockImplementation(() => {
  return mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }])
});`,
        `let propName = 'world';

aVariable.mockImplementation(() => mx[propName]());
aVariable.mockImplementation(() => mx[propName]);
aVariable.mockImplementation(() => ({ [propName]: 1 }));`,
        `const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockImplementation(() => x ? value : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : value ? true : false);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockImplementation(() => true ? true : true ? false : value);

aVariable.mockImplementation(function() {
  if (x) {
    return value;
  } else {
    return 0;
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-mock-return-shorthand / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    // Upstream 29.16.1 classifies this as invalid; current native behavior is valid.
    ...(
      [
        // Upstream 29.16.1 classifies this as invalid; current native behavior is valid.
        'jest.fn().mockImplementation(() => Promise.reject(13))',
        // Upstream 29.16.1 classifies this as invalid; current native behavior is valid.
        `jest.fn().mockImplementation(() => {
  return Promise.reject(13);
})`,
        // Upstream 29.16.1 classifies this as invalid; current native behavior is valid.
        'aVariable.mockImplementation(() => Promise.reject(13))',
        // Upstream 29.16.1 classifies this as invalid; current native behavior is valid.
        `aVariable.mockImplementation(() => {
  return Promise.reject(13);
})`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-mock-return-shorthand / invalid ${index + 13}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'jest.fn().mockImplementation(() => "hello sunshine")',
          'Prefer mockReturnValue',
          1,
          11,
          29,
          'jest.fn().mockReturnValue("hello sunshine")',
        ],
        [
          'jest.fn().mockImplementationOnce(() => "hello world")',
          'Prefer mockReturnValueOnce',
          1,
          11,
          33,
          'jest.fn().mockReturnValueOnce("hello world")',
        ],
        [
          `aVariable.mockImplementation(() => {
  return "hello world";
})`,
          'Prefer mockReturnValue',
          1,
          11,
          29,
          'aVariable.mockReturnValue("hello world")',
        ],
        [
          `const currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
          'Prefer mockReturnValue',
          2,
          25,
          43,
          `const currentX = 0;
jest.spyOn(X, getCount).mockReturnValue(currentX);`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'useMockShorthand',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          'jest.fn().mockImplementation(() => "hello sunshine")',
          'upstream prefer-mock-return-shorthand / invalid 0',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue("hello sunshine")',
        ],
        [
          'jest.fn().mockImplementation(() => null)',
          'upstream prefer-mock-return-shorthand / invalid 1',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(null)',
        ],
        [
          `jest.fn().mockImplementation(() => {
  return null;
})`,
          'upstream prefer-mock-return-shorthand / invalid 2',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(null)',
        ],
        [
          'aVariable.mockImplementation(() => null)',
          'upstream prefer-mock-return-shorthand / invalid 3',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(null)',
        ],
        [
          `aVariable.mockImplementation(() => {
  return null;
})`,
          'upstream prefer-mock-return-shorthand / invalid 4',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(null)',
        ],
        [
          'jest.fn().mockImplementation(() => 0)',
          'upstream prefer-mock-return-shorthand / invalid 5',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(0)',
        ],
        [
          `jest.fn().mockImplementation(() => {
  return 0;
})`,
          'upstream prefer-mock-return-shorthand / invalid 6',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(0)',
        ],
        [
          'aVariable.mockImplementation(() => 0)',
          'upstream prefer-mock-return-shorthand / invalid 7',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(0)',
        ],
        [
          `aVariable.mockImplementation(() => {
  return 0;
})`,
          'upstream prefer-mock-return-shorthand / invalid 8',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(0)',
        ],
        [
          'jest.fn().mockImplementation(() => Promise.resolve(42))',
          'upstream prefer-mock-return-shorthand / invalid 9',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(Promise.resolve(42))',
        ],
        [
          `jest.fn().mockImplementation(() => {
  return Promise.resolve(42);
})`,
          'upstream prefer-mock-return-shorthand / invalid 10',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(Promise.resolve(42))',
        ],
        [
          'aVariable.mockImplementation(() => Promise.resolve(42))',
          'upstream prefer-mock-return-shorthand / invalid 11',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(Promise.resolve(42))',
        ],
        [
          `aVariable.mockImplementation(() => {
  return Promise.resolve(42);
})`,
          'upstream prefer-mock-return-shorthand / invalid 12',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue(Promise.resolve(42))',
        ],
        [
          'jest.fn().mockImplementation(() => [])',
          'upstream prefer-mock-return-shorthand / invalid 17',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue([])',
        ],
        [
          `jest.fn().mockImplementation(() => {
  return [];
})`,
          'upstream prefer-mock-return-shorthand / invalid 18',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue([])',
        ],
        [
          'aVariable.mockImplementation(() => [])',
          'upstream prefer-mock-return-shorthand / invalid 19',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue([])',
        ],
        [
          `aVariable.mockImplementation(() => {
  return [];
})`,
          'upstream prefer-mock-return-shorthand / invalid 20',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue([])',
        ],
        [
          'jest.fn().mockImplementation(() => ({}))',
          'upstream prefer-mock-return-shorthand / invalid 21',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue({})',
        ],
        [
          'jest.fn().mockImplementation(() => x)',
          'upstream prefer-mock-return-shorthand / invalid 22',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(x)',
        ],
        [
          'jest.fn().mockImplementation(() => true ? x : y)',
          'upstream prefer-mock-return-shorthand / invalid 23',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue(true ? x : y)',
        ],
        [
          `jest.fn().mockImplementation(function () {
  return "hello world";
})`,
          'upstream prefer-mock-return-shorthand / invalid 24',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue("hello world")',
        ],
        [
          'jest.fn().mockImplementation(() => "hello world")',
          'upstream prefer-mock-return-shorthand / invalid 25',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue("hello world")',
        ],
        [
          `jest.fn().mockImplementation(() => {
  return "hello world";
})`,
          'upstream prefer-mock-return-shorthand / invalid 26',
          'Prefer mockReturnValue',
          29,
          'jest.fn().mockReturnValue("hello world")',
        ],
        [
          'aVariable.mockImplementation(() => "hello world")',
          'upstream prefer-mock-return-shorthand / invalid 27',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue("hello world")',
        ],
        [
          `aVariable.mockImplementation(() => {
  return "hello world";
})`,
          'upstream prefer-mock-return-shorthand / invalid 28',
          'Prefer mockReturnValue',
          29,
          'aVariable.mockReturnValue("hello world")',
        ],
        [
          'jest.fn().mockImplementationOnce(() => "hello world")',
          'upstream prefer-mock-return-shorthand / invalid 29',
          'Prefer mockReturnValueOnce',
          33,
          'jest.fn().mockReturnValueOnce("hello world")',
        ],
        [
          'aVariable.mockImplementationOnce(() => "hello world")',
          'upstream prefer-mock-return-shorthand / invalid 30',
          'Prefer mockReturnValueOnce',
          33,
          'aVariable.mockReturnValueOnce("hello world")',
        ],
        [
          `aVariable.mockImplementation(() => ({
  target: 'world',
  message: 'hello'
}))`,
          'upstream prefer-mock-return-shorthand / invalid 31',
          'Prefer mockReturnValue',
          29,
          `aVariable.mockReturnValue({
  target: 'world',
  message: 'hello'
})`,
        ],
      ] as const
    ).map(([code, name, message, endColumn, output]): TestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 32',
      code: `aVariable
  .mockImplementation(() => 42)
  .mockImplementation(async () => 42)
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValue("hello world")`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 4,
          endLine: 4,
          endColumn: 22,
        },
      ],
      output: `aVariable
  .mockReturnValue(42)
  .mockImplementation(async () => 42)
  .mockReturnValue(Promise.resolve(42))
  .mockReturnValue("hello world")`,
    },
    ...(
      [
        [
          `aVariable
  .mockImplementationOnce(() => Promise.reject(42))
  .mockImplementation(() => "hello sunshine")
  .mockReturnValueOnce(Promise.reject(42))`,
          3,
          4,
          `aVariable
  .mockImplementationOnce(() => Promise.reject(42))
  .mockReturnValue("hello sunshine")
  .mockReturnValueOnce(Promise.reject(42))`,
        ],
        [
          'jest.fn().mockImplementation(() => [], xyz)',
          1,
          11,
          'jest.fn().mockReturnValue([], xyz)',
        ],
        [
          'jest.spyOn(fs, "readFile").mockImplementation(() => new Error("oh noes!"))',
          1,
          28,
          'jest.spyOn(fs, "readFile").mockReturnValue(new Error("oh noes!"))',
        ],
        [
          `aVariable.mockImplementation(() => {
  return Promise.resolve(value)
    .then(value => value + 1);
});`,
          1,
          11,
          `aVariable.mockReturnValue(Promise.resolve(value)
    .then(value => value + 1));`,
        ],
        [
          `aVariable.mockImplementation(() => {
  return Promise.all([1, 2, 3]);
});`,
          1,
          11,
          'aVariable.mockReturnValue(Promise.all([1, 2, 3]));',
        ],
        [
          `const currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
          2,
          25,
          `const currentX = 0;
jest.spyOn(X, getCount).mockReturnValue(currentX);`,
        ],
      ] as const
    ).map(([code, line, column, output], index): TestCase => ({
      name: `upstream prefer-mock-return-shorthand / invalid ${index + 33}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line,
          column,
          endLine: line,
          endColumn: column + 18,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 39',
      code: `import { currentX } from './elsewhere';
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 43,
        },
      ],
      output: `import { currentX } from './elsewhere';
jest.spyOn(X, getCount).mockReturnValue(currentX);`,
    },
    ...(
      [
        [
          `const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    jest.spyOn(X, getCount).mockImplementation(() => currentX);
  });
});`,
          5,
          29,
          `const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    jest.spyOn(X, getCount).mockReturnValue(currentX);
  });
});`,
        ],
        [
          `function doSomething() {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`,
          3,
          25,
          `function doSomething() {};

jest.spyOn(X, getCount).mockReturnValue(doSomething);`,
        ],
        [
          `const doSomething = () => {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`,
          3,
          25,
          `const doSomething = () => {};

jest.spyOn(X, getCount).mockReturnValue(doSomething);`,
        ],
        [
          `const value = 1;

aVariable.mockImplementation(() => [value]);`,
          3,
          11,
          `const value = 1;

aVariable.mockReturnValue([value]);`,
        ],
        [
          `const value = 1;

aVariable.mockImplementation(() => [0, value, 2]);`,
          3,
          11,
          `const value = 1;

aVariable.mockReturnValue([0, value, 2]);`,
        ],
        [
          `const value = 1;

aVariable.mockImplementation(() => [0,, value, 2]);`,
          3,
          11,
          `const value = 1;

aVariable.mockReturnValue([0,, value, 2]);`,
        ],
        [
          `const value = 1;

jest.fn().mockImplementation(() => ({ value }));`,
          3,
          11,
          `const value = 1;

jest.fn().mockReturnValue({ value });`,
        ],
        [
          `const value = 1;

aVariable.mockImplementation(() => ({ items: [value] }));`,
          3,
          11,
          `const value = 1;

aVariable.mockReturnValue({ items: [value] });`,
        ],
        [
          `const value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: { value },
  }
});`,
          3,
          11,
          `const value = 1;

aVariable.mockReturnValue({
    type: 'object',
    with: { value },
  });`,
        ],
        [
          `const vX = 1;
let vY = 1;

getPoint.mockImplementation(() => vX + vY);
getPoint.mockImplementation(() => {
  return { x: vX, y: 1 }
});`,
          5,
          10,
          `const vX = 1;
let vY = 1;

getPoint.mockImplementation(() => vX + vY);
getPoint.mockReturnValue({ x: vX, y: 1 });`,
        ],
      ] as const
    ).map(([code, line, column, output], index): TestCase => ({
      name: `upstream prefer-mock-return-shorthand / invalid ${index + 40}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line,
          column,
          endLine: line,
          endColumn: column + 18,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 50',
      code: `const value = 1;

aVariable.mockImplementation(() => value & 0);
aVariable.mockImplementation(() => 0 & value);
aVariable.mockImplementation(() => value | 1);
aVariable.mockImplementation(() => 1 | value);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 29,
        },
      ],
      output: `const value = 1;

aVariable.mockReturnValue(value & 0);
aVariable.mockReturnValue(0 & value);
aVariable.mockReturnValue(value | 1);
aVariable.mockReturnValue(1 | value);`,
    },
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 51',
      code: `const value = 1;

aVariable.mockImplementation(() => ~value);
aVariable.mockImplementation(() => !value);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 29,
        },
      ],
      output: `const value = 1;

aVariable.mockReturnValue(~value);
aVariable.mockReturnValue(!value);`,
    },
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 52',
      code: `const value = 1;

aVariable.mockImplementation(() => value + 1);
aVariable.mockImplementation(() => 1 + value);
aVariable.mockImplementation(() => value * value + 1);
aVariable.mockImplementation(() => 1 + value / 2);
aVariable.mockImplementation(() => (1 + value) / 2);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 29,
        },
      ],
      output: `const value = 1;

aVariable.mockReturnValue(value + 1);
aVariable.mockReturnValue(1 + value);
aVariable.mockReturnValue(value * value + 1);
aVariable.mockReturnValue(1 + value / 2);
aVariable.mockReturnValue((1 + value) / 2);`,
    },
    ...(
      [
        [
          `const value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: [1, 2, value],
  }
});`,
          'Prefer mockReturnValue',
          29,
          `const value = 1;

aVariable.mockReturnValue({
    type: 'object',
    with: [1, 2, value],
  });`,
        ],
        [
          `const obj = {};

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    ...obj,
  }
});`,
          'Prefer mockReturnValue',
          29,
          `const obj = {};

aVariable.mockReturnValue({
    type: 'object',
    ...obj,
  });`,
        ],
        [
          `const value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ];
});`,
          'Prefer mockReturnValueOnce',
          33,
          `const value = 1;

jest.fn().mockReturnValueOnce([
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ]);`,
        ],
        [
          `const value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, 0 + value]}
  ];
});`,
          'Prefer mockReturnValueOnce',
          33,
          `const value = 1;

jest.fn().mockReturnValueOnce([
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, 0 + value]}
  ]);`,
        ],
        [
          `const value = 1;

aVariable.mockImplementationOnce(() => {
  return {
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  }
});`,
          'Prefer mockReturnValueOnce',
          33,
          `const value = 1;

aVariable.mockReturnValueOnce({
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  });`,
        ],
        [
          `const value = 1;

aVariable.mockImplementationOnce(() => {
  return {
    type: 'object',
    with: {
      inner: {
        ...{ value },
      },
    },
  }
});`,
          'Prefer mockReturnValueOnce',
          33,
          `const value = 1;

aVariable.mockReturnValueOnce({
    type: 'object',
    with: {
      inner: {
        ...{ value },
      },
    },
  });`,
        ],
        [
          `const value = 1;

jest.fn().mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }
});`,
          'Prefer mockReturnValue',
          29,
          `const value = 1;

jest.fn().mockReturnValue({
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  });`,
        ],
        [
          `const value = 1;

jest.fn().mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]
});`,
          'Prefer mockReturnValue',
          29,
          `const value = 1;

jest.fn().mockReturnValue([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]);`,
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `upstream prefer-mock-return-shorthand / invalid ${index + 53}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 61',
      code: `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockImplementation(() => mx || 0);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => mx ?? (7 && my));
aVariable.mockImplementation(() => mx ?? (7 && 0));`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 29,
        },
      ],
      output: `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockReturnValue(mx || 0);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => mx ?? (7 && my));
aVariable.mockReturnValue(mx ?? (7 && 0));`,
    },
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 62',
      code: `const value = 1;

jest.fn().mockImplementation(() => new Mx(value));
jest.fn().mockImplementation(() => new Mx(() => value));
jest.fn().mockImplementation(() => new Mx(() => { return value }));`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 29,
        },
      ],
      output: `const value = 1;

jest.fn().mockReturnValue(new Mx(value));
jest.fn().mockReturnValue(new Mx(() => value));
jest.fn().mockReturnValue(new Mx(() => { return value }));`,
    },
    {
      name: 'upstream prefer-mock-return-shorthand / invalid 63',
      code: `const value = 1;

jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx?.(value));
jest.fn().mockImplementation(() => mx().my());
jest.fn().mockImplementation(() => mx().my);
jest.fn().mockImplementation(() => mx.my());
jest.fn().mockImplementation(() => mx?.my());
jest.fn().mockImplementation(() => mx.my);
jest.fn().mockImplementation(() => mx(value).my());
jest.fn().mockImplementation(() => mx(value)?.my());
jest.fn().mockImplementation(() => mx(value).my);
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx().my(value));
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx.my?.(value));
jest.fn().mockImplementation(() => mx(value).my(value));
jest.fn().mockImplementation(() => mx?.(value)?.my?.(value));
jest.fn().mockImplementation(() => new Mx().add(value));
jest.fn().mockImplementation(() => {
  return mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }])
});`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 16,
          column: 11,
          endLine: 16,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 17,
          column: 11,
          endLine: 17,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 18,
          column: 11,
          endLine: 18,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 19,
          column: 11,
          endLine: 19,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 20,
          column: 11,
          endLine: 20,
          endColumn: 29,
        },
      ],
      output: `const value = 1;

jest.fn().mockReturnValue(mx(value));
jest.fn().mockReturnValue(mx?.(value));
jest.fn().mockReturnValue(mx().my());
jest.fn().mockReturnValue(mx().my);
jest.fn().mockReturnValue(mx.my());
jest.fn().mockReturnValue(mx?.my());
jest.fn().mockReturnValue(mx.my);
jest.fn().mockReturnValue(mx(value).my());
jest.fn().mockReturnValue(mx(value)?.my());
jest.fn().mockReturnValue(mx(value).my);
jest.fn().mockReturnValue(mx.my(value));
jest.fn().mockReturnValue(mx().my(value));
jest.fn().mockReturnValue(mx.my(value));
jest.fn().mockReturnValue(mx.my?.(value));
jest.fn().mockReturnValue(mx(value).my(value));
jest.fn().mockReturnValue(mx?.(value)?.my?.(value));
jest.fn().mockReturnValue(new Mx().add(value));
jest.fn().mockReturnValue(mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]));`,
    },
    ...(
      [
        [
          `const propName = 'world';

aVariable.mockImplementation(() => mx[propName]());
aVariable.mockImplementation(() => mx[propName]);
aVariable.mockImplementation(() => ({ [propName]: 1 }));`,
          3,
          4,
          5,
          `const propName = 'world';

aVariable.mockReturnValue(mx[propName]());
aVariable.mockReturnValue(mx[propName]);
aVariable.mockReturnValue({ [propName]: 1 });`,
        ],
        [
          `const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockImplementation(() => x ? true : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockImplementation(() => true ? true : true ? x : false);
aVariable.mockImplementation(() => true ? true : true ? true : false);`,
          5,
          9,
          10,
          `const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockReturnValue(x ? true : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockReturnValue(true ? true : true ? x : false);
aVariable.mockReturnValue(true ? true : true ? true : false);`,
        ],
      ] as const
    ).map(([code, line1, line2, line3, output], index): TestCase => ({
      name: `upstream prefer-mock-return-shorthand / invalid ${index + 64}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: line1,
          column: 11,
          endLine: line1,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: line2,
          column: 11,
          endLine: line2,
          endColumn: 29,
        },
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: line3,
          column: 11,
          endLine: line3,
          endColumn: 29,
        },
      ],
      output,
    })),
  ],
});
