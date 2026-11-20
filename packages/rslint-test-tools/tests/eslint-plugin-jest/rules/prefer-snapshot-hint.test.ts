// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const assertionSetup = `const myReusableTestBody = (value, snapshotHint) => {
  const innerFn = anotherValue => {
    expect(anotherValue).toMatchSnapshot();

    expect(value).toBe(1);
`;

new RuleTester(fixtureConfig()).run('prefer-snapshot-hint', {
  valid: [
    ...(
      [
        'expect(something).toStrictEqual(somethingElse);',
        "a().toEqual('b')",
        'expect(a);',
        'expect(1).toMatchSnapshot({}, "my snapshot");',
        'expect(1).toMatchSnapshot("my snapshot");',
        'expect(1).toMatchSnapshot(`my snapshot`);',
        `const x = {};
expect(1).toMatchSnapshot(x, "my snapshot");`,
        'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
        'expect(1).toMatchInlineSnapshot();',
        'expect(1).toThrowErrorMatchingInlineSnapshot();',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
      options: ['always'],
    })),
    ...(
      [
        'expect(something).toStrictEqual(somethingElse);',
        "a().toEqual('b')",
        'expect(a);',
        'expect(1).toMatchSnapshot({}, "my snapshot");',
        'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
        'expect(1).toMatchSnapshot({});',
        'expect(1).toThrowErrorMatchingSnapshot();',
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot(undefined, 'my first snapshot');
});`,
        `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot('this is a hint, all by itself');
  });

  it('is false', () => {
    expect(2).toMatchSnapshot('this is a hint');
    expect(2).toMatchSnapshot('and so is this');
  });
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(2).toMatchSnapshot('this is a hint');
  expect(2).toMatchSnapshot('and so is this');
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(2).toThrowErrorMatchingSnapshot();
});`,
        `it('is true', () => {
  expect(1).toStrictEqual(1);
  expect(1).toStrictEqual(2);
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(1).toStrictEqual(1);
  expect(1).toStrictEqual(2);
  expect(2).toThrowErrorMatchingSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchInlineSnapshot();
});

it('is false', () => {
  expect(1).toMatchInlineSnapshot();
  expect(1).toMatchInlineSnapshot();
  expect(1).toThrowErrorMatchingInlineSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(1).toMatchSnapshot();
});`,
        `import { it as itIs } from '@jest/globals';

it('is true', () => {
  expect(1).toMatchSnapshot();
});

itIs('false', () => {
  expect(1).toMatchSnapshot();
});`,
        assertionSetup +
          `  };

  expect(value).toBe(1);
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
        `const myReusableTestBody = (value, snapshotHint) => {
  const innerFn = anotherValue => {
    expect(value).toBe(1);
  };

  expect(value).toBe(1);
  expect(anotherValue).toMatchSnapshot();
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
        assertionSetup +
          `  };

  expect(value).toBe(1);
};

expect(1).toMatchSnapshot();`,
        `const register = () => {
  const first = () => expect('first').toMatchSnapshot();
  test('first', first);
  const second = () => expect('second').toMatchSnapshot();
  test('second', second);
};
describe('suite', register);`,
        `test('case', callback);
function callback(): void;
function callback() { expect('test').toMatchSnapshot(); }
function helper() { expect('helper').toMatchSnapshot(); }`,
        `describe('suite', () => {
  const a = () => expect(1).toMatchSnapshot();
  const b = () => expect(2).toMatchSnapshot();
  test('a', a as () => void);
  test('b', b satisfies () => void);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 1 valid ${index}`,
      code,
      options: ['multi'],
    })),
    ...(
      [
        "await expect(Promise.resolve(1)).resolves.toMatchSnapshot('named').then(() => {});",
        "await expect(Promise.resolve(1)).resolves.toMatchSnapshot('named').catch(() => {});",
        "await expect(Promise.resolve(1)).resolves.toMatchSnapshot('named').finally(() => {});",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 1 valid ${index + 22}`,
      code,
      options: ['always'],
    })),
    {
      name: 'run 1 valid 25',
      code: 'await expect(Promise.resolve(1)).resolves.toMatchSnapshot().then(() => {});',
      options: ['multi'],
    },
    ...(
      [
        'expect(something).toStrictEqual(somethingElse);',
        "a().toEqual('b')",
        'expect(a);',
        'expect(1).toMatchSnapshot({}, "my snapshot");',
        'expect(1).toMatchSnapshot("my snapshot");',
        'expect(1).toMatchSnapshot(`my snapshot`);',
        `const x = {};
expect(1).toMatchSnapshot(x, "my snapshot");`,
        'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
        'expect(1).toMatchInlineSnapshot();',
        'expect(1).toThrowErrorMatchingInlineSnapshot();',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-snapshot-hint (always) / valid ${index}`,
      code,
      options: ['always'],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'expect(something).toStrictEqual(somethingElse);',
        "a().toEqual('b')",
        'expect(a);',
        'expect(1).toMatchSnapshot({}, "my snapshot");',
        'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
        'expect(1).toMatchSnapshot({});',
        'expect(1).toThrowErrorMatchingSnapshot();',
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot(undefined, 'my first snapshot');
});`,
        `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot('this is a hint, all by itself');
  });

  it('is false', () => {
    expect(2).toMatchSnapshot('this is a hint');
    expect(2).toMatchSnapshot('and so is this');
  });
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(2).toMatchSnapshot('this is a hint');
  expect(2).toMatchSnapshot('and so is this');
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(2).toThrowErrorMatchingSnapshot();
});`,
        `it('is true', () => {
  expect(1).toStrictEqual(1);
  expect(1).toStrictEqual(2);
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(1).toStrictEqual(1);
  expect(1).toStrictEqual(2);
  expect(2).toThrowErrorMatchingSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchInlineSnapshot();
});

it('is false', () => {
  expect(1).toMatchInlineSnapshot();
  expect(1).toMatchInlineSnapshot();
  expect(1).toThrowErrorMatchingInlineSnapshot();
});`,
        `it('is true', () => {
  expect(1).toMatchSnapshot();
});

it('is false', () => {
  expect(1).toMatchSnapshot();
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-snapshot-hint (multi) / valid ${index}`,
      code,
      options: ['multi'],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream prefer-snapshot-hint (multi) / valid 15',
      code: `import { it as itIs } from '@jest/globals';

it('is true', () => {
  expect(1).toMatchSnapshot();
});

itIs('false', () => {
  expect(1).toMatchSnapshot();
});`,
      options: ['multi'],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        assertionSetup +
          `  };

  expect(value).toBe(1);
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
        `const myReusableTestBody = (value, snapshotHint) => {
  const innerFn = anotherValue => {
    expect(value).toBe(1);
  };

  expect(value).toBe(1);
  expect(anotherValue).toMatchSnapshot();
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
        assertionSetup +
          `  };

  expect(value).toBe(1);
};

expect(1).toMatchSnapshot();`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-snapshot-hint (multi) / valid ${index + 16}`,
      code,
      options: ['multi'],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['expect(1).toMatchSnapshot();', 1, 11, 26],
        ['expect(1).toMatchSnapshot({});', 1, 11, 26],
        [
          `const x = "we can't know if this is a string or not";
expect(1).toMatchSnapshot(x);`,
          2,
          11,
          26,
        ],
        ['expect(1).toThrowErrorMatchingSnapshot();', 1, 11, 39],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
});`,
          2,
          13,
          28,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 5',
      code: `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot();
});`,
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toThrowErrorMatchingSnapshot("my error");
});`,
          13,
          28,
        ],
        [
          `const expectSnapshot = value => {
  expect(value).toMatchSnapshot();
};`,
          17,
          32,
        ],
        [
          `const expectSnapshot = value => {
  expect(value).toThrowErrorMatchingSnapshot();
};`,
          17,
          45,
        ],
        [
          `it('is true', () => {
  { expect(1).toMatchSnapshot(); }
});`,
          15,
          30,
        ],
        [
          `const x = "snapshot";
expect(1).toMatchSnapshot(\`my \${x}\`);`,
          11,
          26,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 6}`,
      code,
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column,
          endLine: 2,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 1 invalid 0',
      code: `await expect(Promise.resolve(1))
  .resolves.toMatchSnapshot()
  .then(() => {});`,
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 1 invalid 1',
      code: `const outer = () => {
  expect('before').toMatchSnapshot();
  let callback = () => { expect('stale').toMatchSnapshot(); };
  callback = () => {};
  test('case', callback);
  expect('after').toMatchSnapshot();
};`,
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 42,
          endLine: 3,
          endColumn: 57,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 6,
          column: 19,
          endLine: 6,
          endColumn: 34,
        },
      ],
    },
    ...(
      [
        [
          `const outer = () => {
  expect('before').toMatchSnapshot();
  { const callback = () => expect('one').toMatchSnapshot(); test('one', callback); }
  { const callback = () => expect('two').toMatchSnapshot(); test('two', callback); }
  expect('after').toMatchSnapshot();
};`,
          20,
          35,
          5,
          19,
          34,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot();
});`,
          13,
          28,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toThrowErrorMatchingSnapshot();
});`,
          13,
          28,
          3,
          13,
          41,
        ],
        [
          `it('is true', () => {
  expect(1).toThrowErrorMatchingSnapshot();
  expect(2).toMatchSnapshot();
});`,
          13,
          41,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  expect(2).toMatchSnapshot({});
});`,
          13,
          28,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  {
    expect(2).toMatchSnapshot({});
  }
});`,
          13,
          28,
          4,
          15,
          30,
        ],
        [
          `it('is true', () => {
  { expect(1).toMatchSnapshot(); }
  { expect(2).toMatchSnapshot(); }
});`,
          15,
          30,
          3,
          15,
          30,
        ],
      ] as const
    ).map(
      (
        [code, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 1 invalid ${index + 2}`,
        code,
        options: ['multi'],
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: 2,
            column: column1,
            endLine: 2,
            endColumn: endColumn1,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot(undefined, 'my second snapshot');
});`,
          2,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  expect(2).toMatchSnapshot(undefined, 'my second snapshot');
});`,
          2,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({}, 'my first snapshot');
  expect(2).toMatchSnapshot(undefined);
});`,
          3,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `run 1 invalid ${index + 9}`,
      code,
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column: 13,
          endLine: line,
          endColumn: 28,
        },
      ],
    })),
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({}, 'my first snapshot');
  expect(2).toMatchSnapshot(undefined);
  expect(2).toMatchSnapshot();
});`,
          3,
          13,
          4,
          13,
        ],
        [
          `it('is true', () => {
  expect(2).toMatchSnapshot();
  expect(1).toMatchSnapshot({}, 'my second snapshot');
  expect(2).toMatchSnapshot();
});`,
          2,
          13,
          4,
          13,
        ],
        [
          `it('is true', () => {
  expect(2).toMatchSnapshot(undefined);
  expect(2).toMatchSnapshot();
  expect(1).toMatchSnapshot(null, 'my third snapshot');
});`,
          2,
          13,
          3,
          13,
        ],
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot();
  });
});`,
          7,
          15,
          8,
          15,
        ],
      ] as const
    ).map(([code, line1, column1, line2, column2], index): TestCase => ({
      name: `run 1 invalid ${index + 12}`,
      code,
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: line1,
          column: column1,
          endLine: line1,
          endColumn: column1 + 15,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: line2,
          column: column2,
          endLine: line2,
          endColumn: column2 + 15,
        },
      ],
    })),
    ...(
      [
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot('hello world');
  });
});`,
          7,
          15,
        ],
        [
          `describe('my tests', () => {
  describe('more tests', () => {
    it('is true', () => {
      expect(1).toMatchSnapshot();
    });
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot('hello world');
  });
});`,
          9,
          15,
        ],
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  describe('more tests', () => {
    it('is false', () => {
      expect(2).toMatchSnapshot();
      expect(2).toMatchSnapshot('hello world');
    });
  });
});`,
          8,
          17,
        ],
        [
          `import { describe as context, it as itIs } from '@jest/globals';

describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  context('more tests', () => {
    itIs('false', () => {
      expect(2).toMatchSnapshot();
      expect(2).toMatchSnapshot('hello world');
    });
  });
});`,
          10,
          17,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 1 invalid ${index + 16}`,
      code,
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column,
          endLine: line,
          endColumn: column + 15,
        },
      ],
    })),
    ...(
      [
        [
          `const myReusableTestBody = (value, snapshotHint) => {
  expect(value).toMatchSnapshot();

  const innerFn = anotherValue => {
    expect(anotherValue).toMatchSnapshot();
  };

  expect(value).toBe(1);
  expect(value + 1).toMatchSnapshot(null);
  expect(value + 2).toThrowErrorMatchingSnapshot(snapshotHint);
};`,
          2,
          17,
          5,
          26,
          9,
          21,
          36,
        ],
        [
          `const myReusableTestBody = (value, snapshotHint) => {
  expect(value).toMatchSnapshot();

  const innerFn = anotherValue => {
    expect(anotherValue).toMatchSnapshot();

    expect(value).toBe(1);
    expect(value + 1).toMatchSnapshot(null);
    expect(value + 2).toMatchSnapshot(null, snapshotHint);
  };
};`,
          2,
          17,
          5,
          26,
          8,
          23,
          38,
        ],
        [
          assertionSetup +
            `    expect(value + 1).toMatchSnapshot(null);
    expect(value + 2).toMatchSnapshot(null, snapshotHint);
  };

  expect(value).toThrowErrorMatchingSnapshot();
};`,
          3,
          26,
          6,
          23,
          10,
          17,
          45,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, line2, column2, line3, column3, endColumn3],
        index,
      ): TestCase => ({
        name: `run 1 invalid ${index + 20}`,
        code,
        options: ['multi'],
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: column1 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: column2 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line3,
            column: column3,
            endLine: line3,
            endColumn: endColumn3,
          },
        ],
      }),
    ),
    ...(
      [
        [
          assertionSetup +
            `  };

  expect(value).toMatchSnapshot();
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
          3,
          26,
          8,
          17,
          32,
        ],
        [
          `const myReusableTestBody = value => {
  expect(value).toMatchSnapshot();
};

expect(1).toMatchSnapshot();
expect(1).toThrowErrorMatchingSnapshot();`,
          5,
          11,
          6,
          11,
          39,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 1 invalid ${index + 23}`,
        code,
        options: ['multi'],
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: column1 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        ['expect(1).toMatchSnapshot();', 1, 11, 26],
        ['expect(1).toMatchSnapshot({});', 1, 11, 26],
        [
          `const x = "we can't know if this is a string or not";
expect(1).toMatchSnapshot(x);`,
          2,
          11,
          26,
        ],
        ['expect(1).toThrowErrorMatchingSnapshot();', 1, 11, 39],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
});`,
          2,
          13,
          28,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream prefer-snapshot-hint (always) / invalid ${index}`,
      code,
      options: ['always'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream prefer-snapshot-hint (always) / invalid 5',
      code: `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot();
});`,
      options: ['always'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toThrowErrorMatchingSnapshot("my error");
});`,
          13,
          28,
        ],
        [
          `const expectSnapshot = value => {
  expect(value).toMatchSnapshot();
};`,
          17,
          32,
        ],
        [
          `const expectSnapshot = value => {
  expect(value).toThrowErrorMatchingSnapshot();
};`,
          17,
          45,
        ],
        [
          `it('is true', () => {
  { expect(1).toMatchSnapshot(); }
});`,
          15,
          30,
        ],
        [
          `const x = "snapshot";
expect(1).toMatchSnapshot(\`my \${x}\`);`,
          11,
          26,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream prefer-snapshot-hint (always) / invalid ${index + 6}`,
      code,
      options: ['always'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column,
          endLine: 2,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot();
});`,
          13,
          28,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toThrowErrorMatchingSnapshot();
});`,
          13,
          28,
          3,
          13,
          41,
        ],
        [
          `it('is true', () => {
  expect(1).toThrowErrorMatchingSnapshot();
  expect(2).toMatchSnapshot();
});`,
          13,
          41,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  expect(2).toMatchSnapshot({});
});`,
          13,
          28,
          3,
          13,
          28,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  {
    expect(2).toMatchSnapshot({});
  }
});`,
          13,
          28,
          4,
          15,
          30,
        ],
        [
          `it('is true', () => {
  { expect(1).toMatchSnapshot(); }
  { expect(2).toMatchSnapshot(); }
});`,
          15,
          30,
          3,
          15,
          30,
        ],
      ] as const
    ).map(
      (
        [code, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream prefer-snapshot-hint (multi) / invalid ${index}`,
        code,
        options: ['multi'],
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: 2,
            column: column1,
            endLine: 2,
            endColumn: endColumn1,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot();
  expect(2).toMatchSnapshot(undefined, 'my second snapshot');
});`,
          2,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({});
  expect(2).toMatchSnapshot(undefined, 'my second snapshot');
});`,
          2,
        ],
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({}, 'my first snapshot');
  expect(2).toMatchSnapshot(undefined);
});`,
          3,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `upstream prefer-snapshot-hint (multi) / invalid ${index + 6}`,
      code,
      options: ['multi'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column: 13,
          endLine: line,
          endColumn: 28,
        },
      ],
    })),
    ...(
      [
        [
          `it('is true', () => {
  expect(1).toMatchSnapshot({}, 'my first snapshot');
  expect(2).toMatchSnapshot(undefined);
  expect(2).toMatchSnapshot();
});`,
          3,
          13,
          4,
          13,
        ],
        [
          `it('is true', () => {
  expect(2).toMatchSnapshot();
  expect(1).toMatchSnapshot({}, 'my second snapshot');
  expect(2).toMatchSnapshot();
});`,
          2,
          13,
          4,
          13,
        ],
        [
          `it('is true', () => {
  expect(2).toMatchSnapshot(undefined);
  expect(2).toMatchSnapshot();
  expect(1).toMatchSnapshot(null, 'my third snapshot');
});`,
          2,
          13,
          3,
          13,
        ],
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot();
  });
});`,
          7,
          15,
          8,
          15,
        ],
      ] as const
    ).map(([code, line1, column1, line2, column2], index): TestCase => ({
      name: `upstream prefer-snapshot-hint (multi) / invalid ${index + 9}`,
      code,
      options: ['multi'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: line1,
          column: column1,
          endLine: line1,
          endColumn: column1 + 15,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: line2,
          column: column2,
          endLine: line2,
          endColumn: column2 + 15,
        },
      ],
    })),
    ...(
      [
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot('hello world');
  });
});`,
          7,
          15,
        ],
        [
          `describe('my tests', () => {
  describe('more tests', () => {
    it('is true', () => {
      expect(1).toMatchSnapshot();
    });
  });

  it('is false', () => {
    expect(2).toMatchSnapshot();
    expect(2).toMatchSnapshot('hello world');
  });
});`,
          9,
          15,
        ],
        [
          `describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  describe('more tests', () => {
    it('is false', () => {
      expect(2).toMatchSnapshot();
      expect(2).toMatchSnapshot('hello world');
    });
  });
});`,
          8,
          17,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream prefer-snapshot-hint (multi) / invalid ${index + 13}`,
      code,
      options: ['multi'],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line,
          column,
          endLine: line,
          endColumn: column + 15,
        },
      ],
    })),
    {
      name: 'upstream prefer-snapshot-hint (multi) / invalid 16',
      code: `import { describe as context, it as itIs } from '@jest/globals';

describe('my tests', () => {
  it('is true', () => {
    expect(1).toMatchSnapshot();
  });

  context('more tests', () => {
    itIs('false', () => {
      expect(2).toMatchSnapshot();
      expect(2).toMatchSnapshot('hello world');
    });
  });
});`,
      options: ['multi'],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 10,
          column: 17,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    ...(
      [
        [
          `const myReusableTestBody = (value, snapshotHint) => {
  expect(value).toMatchSnapshot();

  const innerFn = anotherValue => {
    expect(anotherValue).toMatchSnapshot();
  };

  expect(value).toBe(1);
  expect(value + 1).toMatchSnapshot(null);
  expect(value + 2).toThrowErrorMatchingSnapshot(snapshotHint);
};`,
          2,
          17,
          5,
          26,
          9,
          21,
          36,
        ],
        [
          `const myReusableTestBody = (value, snapshotHint) => {
  expect(value).toMatchSnapshot();

  const innerFn = anotherValue => {
    expect(anotherValue).toMatchSnapshot();

    expect(value).toBe(1);
    expect(value + 1).toMatchSnapshot(null);
    expect(value + 2).toMatchSnapshot(null, snapshotHint);
  };
};`,
          2,
          17,
          5,
          26,
          8,
          23,
          38,
        ],
        [
          assertionSetup +
            `    expect(value + 1).toMatchSnapshot(null);
    expect(value + 2).toMatchSnapshot(null, snapshotHint);
  };

  expect(value).toThrowErrorMatchingSnapshot();
};`,
          3,
          26,
          6,
          23,
          10,
          17,
          45,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, line2, column2, line3, column3, endColumn3],
        index,
      ): TestCase => ({
        name: `upstream prefer-snapshot-hint (multi) / invalid ${index + 17}`,
        code,
        options: ['multi'],
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: column1 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: column2 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line3,
            column: column3,
            endLine: line3,
            endColumn: endColumn3,
          },
        ],
      }),
    ),
    ...(
      [
        [
          assertionSetup +
            `  };

  expect(value).toMatchSnapshot();
};

it('my test', () => {
  expect(1).toMatchSnapshot();
});`,
          3,
          26,
          8,
          17,
          32,
        ],
        [
          `const myReusableTestBody = value => {
  expect(value).toMatchSnapshot();
};

expect(1).toMatchSnapshot();
expect(1).toThrowErrorMatchingSnapshot();`,
          5,
          11,
          6,
          11,
          39,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream prefer-snapshot-hint (multi) / invalid ${index + 20}`,
        code,
        options: ['multi'],
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: column1 + 15,
          },
          {
            messageId: 'missingHint',
            message: 'You should provide a hint for this snapshot',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
  ],
});
