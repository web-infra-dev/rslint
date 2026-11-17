// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-hooks-on-top', {
  valid: [
    ...(
      [
        `
      describe('foo', () => {
        beforeEach(() => {});
        someSetupFn();
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
   \u0020`,
        `
      describe('foo', () => {
        someSetupFn();
        beforeEach(() => {});
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
   \u0020`,
        `
      describe.skip('foo', () => {
        beforeEach(() => {});
        beforeAll(() => {});

        test('bar', () => {
          someFn();
        });
      });

      describe('foo', () => {
        beforeEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
   \u0020`,
        `
      describe('foo', () => {
        beforeEach(() => {});
        test('bar', () => {
          someFn();
        });

        describe('inner_foo', () => {
          beforeEach(() => {});
          test('inner bar', () => {
            someFn();
          });
        });
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        `describe('foo', () => {
  beforeEach(() => {});
  someSetupFn();
  afterEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
        `describe('foo', () => {
  someSetupFn();
  beforeEach(() => {});
  afterEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream basic describe block / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream multiple describe blocks / valid 0',
      code: `describe.skip('foo', () => {
  beforeEach(() => {});
  beforeAll(() => {});

  test('bar', () => {
    someFn();
  });
});

describe('foo', () => {
  beforeEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream nested describe blocks / valid 0',
      code: `describe('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  describe('inner_foo', () => {
    beforeEach(() => {});
    test('inner bar', () => {
      someFn();
    });
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
  ],
  invalid: [
    ...(
      [
        `
        describe('foo', () => {
          beforeEach(() => {});
          test('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });
        });
     \u0020`,
        `
        describe('foo', () => {
          beforeEach(() => {});
          test.each\`\`('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test.only('bar', () => {
            someFn();
          });
        });
     \u0020`,
        `
        describe('foo', () => {
          beforeEach(() => {});
          test.only.each\`\`('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test.only('bar', () => {
            someFn();
          });
        });
     \u0020`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 30,
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
      code: `
        describe.skip('foo', () => {
          beforeEach(() => {});
          test('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });
        });
        describe('foo', () => {
          beforeEach(() => {});
          beforeEach(() => {});
          beforeAll(() => {});

          test('bar', () => {
            someFn();
          });
        });

        describe('foo', () => {
          test('bar', () => {
            someFn();
          });

          beforeEach(() => {});
          beforeEach(() => {});
          beforeAll(() => {});
        });
     \u0020`,
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 30,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 28,
          column: 11,
          endLine: 28,
          endColumn: 31,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 29,
          column: 11,
          endLine: 29,
          endColumn: 31,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 30,
          column: 11,
          endLine: 30,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: `
        describe('foo', () => {
          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });

          describe('inner_foo', () => {
            beforeEach(() => {});
            test('inner bar', () => {
              someFn();
            });

            test('inner bar', () => {
              someFn();
            });

            beforeAll(() => {});
            afterAll(() => {});
            test('inner bar', () => {
              someFn();
            });
          });
        });
     \u0020`,
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 18,
          column: 13,
          endLine: 18,
          endColumn: 32,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 19,
          column: 13,
          endLine: 19,
          endColumn: 31,
        },
      ],
    },
    ...(
      [
        `describe('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });
});`,
        `describe('foo', () => {
  beforeEach(() => {});
  test.each\`\`('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test.only('bar', () => {
    someFn();
  });
});`,
        `describe('foo', () => {
  beforeEach(() => {});
  test.only.each\`\`('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test.only('bar', () => {
    someFn();
  });
});`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream basic describe block / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 22,
        },
      ],
    })),
    {
      name: 'upstream multiple describe blocks / invalid 0',
      code: `describe.skip('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });
});
describe('foo', () => {
  beforeEach(() => {});
  beforeEach(() => {});
  beforeAll(() => {});

  test('bar', () => {
    someFn();
  });
});

describe('foo', () => {
  test('bar', () => {
    someFn();
  });

  beforeEach(() => {});
  beforeEach(() => {});
  beforeAll(() => {});
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 22,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 27,
          column: 3,
          endLine: 27,
          endColumn: 23,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 28,
          column: 3,
          endLine: 28,
          endColumn: 23,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 29,
          column: 3,
          endLine: 29,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream nested describe blocks / invalid 0',
      code: `describe('foo', () => {
  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });

  describe('inner_foo', () => {
    beforeEach(() => {});
    test('inner bar', () => {
      someFn();
    });

    test('inner bar', () => {
      someFn();
    });

    beforeAll(() => {});
    afterAll(() => {});
    test('inner bar', () => {
      someFn();
    });
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 17,
          column: 5,
          endLine: 17,
          endColumn: 24,
        },
        {
          messageId: 'noHookOnTop',
          message: 'Hooks should come before test cases',
          line: 18,
          column: 5,
          endLine: 18,
          endColumn: 23,
        },
      ],
    },
  ],
});
