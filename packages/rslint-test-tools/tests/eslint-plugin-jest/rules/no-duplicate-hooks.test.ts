// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-duplicate-hooks', {
  valid: [
    ...(
      [
        `
      describe("foo", () => {
        beforeEach(() => {})
        test("bar", () => {
          someFn();
        })
      })
   \u0020`,
        `
      beforeEach(() => {})
      test("bar", () => {
        someFn();
      })
   \u0020`,
        `
      describe("foo", () => {
        beforeAll(() => {}),
        beforeEach(() => {})
        afterEach(() => {})
        afterAll(() => {})

        test("bar", () => {
          someFn();
        })
      })
   \u0020`,
        `
      describe.skip("foo", () => {
        beforeEach(() => {}),
        beforeAll(() => {}),
        test("bar", () => {
          someFn();
        })
      })
      describe("foo", () => {
        beforeEach(() => {}),
        beforeAll(() => {}),
        test("bar", () => {
          someFn();
        })
      })
   \u0020`,
        `
      describe("foo", () => {
        beforeEach(() => {}),
        test("bar", () => {
          someFn();
        })
        describe("inner_foo", () => {
          beforeEach(() => {})
          test("inner bar", () => {
            someFn();
          })
        })
      })
   \u0020`,
        `
      describe.each(['hello'])('%s', () => {
        beforeEach(() => {});

        it('is fine', () => {});
      });
   \u0020`,
        `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });
      });
   \u0020`,
        `
      describe.each\`\`('%s', () => {
        beforeEach(() => {});

        it('is fine', () => {});
      });
   \u0020`,
        `
      describe('something', () => {
        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
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
        `describe("foo", () => {
  beforeEach(() => {})
  test("bar", () => {
    someFn();
  })
})`,
        `beforeEach(() => {})
test("bar", () => {
  someFn();
})`,
        `describe("foo", () => {
  beforeAll(() => {}),
  beforeEach(() => {})
  afterEach(() => {})
  afterAll(() => {})

  test("bar", () => {
    someFn();
  })
})`,
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
      code: `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})
describe("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream nested describe blocks / valid 0',
      code: `describe("foo", () => {
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
  describe("inner_foo", () => {
    beforeEach(() => {})
    test("inner bar", () => {
      someFn();
    })
  })
})`,
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        `describe.each(['hello'])('%s', () => {
  beforeEach(() => {});

  it('is fine', () => {});
});`,
        `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });
});`,
        `describe.each\`\`('%s', () => {
  beforeEach(() => {});

  it('is fine', () => {});
});`,
        `describe('something', () => {
  describe.each\`\`('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each\`\`('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream describe.each blocks / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate beforeEach in describe block',
          4,
          11,
          31,
        ],
        [
          `
        describe.skip("foo", () => {
          beforeEach(() => {}),
          beforeAll(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate beforeAll in describe block',
          5,
          11,
          30,
        ],
        [
          `
        describe.skip("foo", () => {
          afterEach(() => {}),
          afterEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate afterEach in describe block',
          4,
          11,
          30,
        ],
        [
          `
        import { afterEach } from '@jest/globals';

        describe.skip("foo", () => {
          afterEach(() => {}),
          afterEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate afterEach in describe block',
          6,
          11,
          30,
        ],
        [
          `
        import { afterEach, afterEach as somethingElse } from '@jest/globals';

        describe.skip("foo", () => {
          afterEach(() => {}),
          somethingElse(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate afterEach in describe block',
          6,
          11,
          34,
        ],
        [
          `
        describe.skip("foo", () => {
          afterAll(() => {}),
          afterAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate afterAll in describe block',
          4,
          11,
          29,
        ],
        [
          `
        afterAll(() => {}),
        afterAll(() => {}),
        test("bar", () => {
          someFn();
        })
     \u0020`,
          'Duplicate afterAll in describe block',
          3,
          9,
          27,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noDuplicateHook',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          beforeEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate beforeEach in describe block',
          31,
          'Duplicate beforeEach in describe block',
          5,
          31,
        ],
        [
          `
        describe.skip("foo", () => {
          afterAll(() => {}),
          afterAll(() => {}),
          beforeAll(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          'Duplicate afterAll in describe block',
          29,
          'Duplicate beforeAll in describe block',
          6,
          30,
        ],
      ] as const
    ).map(
      (
        [code, message1, endColumn1, message2, line2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 7}`,
        code,
        errors: [
          {
            messageId: 'noDuplicateHook',
            message: message1,
            line: 4,
            column: 11,
            endLine: 4,
            endColumn: endColumn1,
          },
          {
            messageId: 'noDuplicateHook',
            message: message2,
            line: line2,
            column: 11,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `
        describe.skip("foo", () => {
          beforeEach(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
     \u0020`,
          11,
          11,
        ],
        [
          `
        describe("foo", () => {
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
          describe("inner_foo", () => {
            beforeEach(() => {})
            beforeEach(() => {})
            test("inner bar", () => {
              someFn();
            })
          })
        })
     \u0020`,
          9,
          13,
        ],
        [
          `
      describe.each(['hello'])('%s', () => {
        beforeEach(() => {});
        beforeEach(() => {});

        it('is not fine', () => {});
      });
   \u0020`,
          4,
          9,
        ],
        [
          `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          beforeEach(() => {});
          beforeEach(() => {});

          it('is not fine', () => {});
        });
      });
   \u0020`,
          11,
          11,
        ],
        [
          `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          describe('some more', () => {
            beforeEach(() => {});
            beforeEach(() => {});

            it('is not fine', () => {});
          });
        });
      });
   \u0020`,
          12,
          13,
        ],
        [
          `
      describe.each\`\`('%s', () => {
        beforeEach(() => {});
        beforeEach(() => {});

        it('is fine', () => {});
      });
   \u0020`,
          4,
          9,
        ],
        [
          `
      describe('something', () => {
        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each\`\`('%s', () => {
          beforeEach(() => {});
          beforeEach(() => {});

          it('is not fine', () => {});
        });
      });
   \u0020`,
          11,
          11,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 9}`,
      code,
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeEach in describe block',
          line,
          column,
          endLine: line,
          endColumn: column + 20,
        },
      ],
    })),
    ...(
      [
        [
          `describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          'Duplicate beforeEach in describe block',
          3,
          23,
        ],
        [
          `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          'Duplicate beforeAll in describe block',
          4,
          22,
        ],
        [
          `describe.skip("foo", () => {
  afterEach(() => {}),
  afterEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          'Duplicate afterEach in describe block',
          3,
          22,
        ],
      ] as const
    ).map(([code, message, line, endColumn], index): TestCase => ({
      name: `upstream basic describe block / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message,
          line,
          column: 3,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `import { afterEach } from '@jest/globals';

describe.skip("foo", () => {
  afterEach(() => {}),
  afterEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          22,
        ],
        [
          `import { afterEach, afterEach as somethingElse } from '@jest/globals';

describe.skip("foo", () => {
  afterEach(() => {}),
  somethingElse(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          26,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream basic describe block / invalid ${index + 3}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate afterEach in describe block',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `describe.skip("foo", () => {
  afterAll(() => {}),
  afterAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          3,
          3,
        ],
        [
          `afterAll(() => {}),
afterAll(() => {}),
test("bar", () => {
  someFn();
})`,
          2,
          1,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream basic describe block / invalid ${index + 5}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate afterAll in describe block',
          line,
          column,
          endLine: line,
          endColumn: column + 18,
        },
      ],
    })),
    ...(
      [
        [
          `describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          'Duplicate beforeEach in describe block',
          23,
          'Duplicate beforeEach in describe block',
          4,
          23,
        ],
        [
          `describe.skip("foo", () => {
  afterAll(() => {}),
  afterAll(() => {}),
  beforeAll(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
          'Duplicate afterAll in describe block',
          21,
          'Duplicate beforeAll in describe block',
          5,
          22,
        ],
      ] as const
    ).map(
      (
        [code, message1, endColumn1, message2, line2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream basic describe block / invalid ${index + 7}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'noDuplicateHook',
            message: message1,
            line: 3,
            column: 3,
            endLine: 3,
            endColumn: endColumn1,
          },
          {
            messageId: 'noDuplicateHook',
            message: message2,
            line: line2,
            column: 3,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    {
      name: 'upstream multiple describe blocks / invalid 0',
      code: `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})
describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeEach in describe block',
          line: 10,
          column: 3,
          endLine: 10,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream nested describe blocks / invalid 0',
      code: `describe("foo", () => {
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
  describe("inner_foo", () => {
    beforeEach(() => {})
    beforeEach(() => {})
    test("inner bar", () => {
      someFn();
    })
  })
})`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeEach in describe block',
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 25,
        },
      ],
    },
    ...(
      [
        [
          `describe.each(['hello'])('%s', () => {
  beforeEach(() => {});
  beforeEach(() => {});

  it('is not fine', () => {});
});`,
          3,
          3,
        ],
        [
          `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    beforeEach(() => {});
    beforeEach(() => {});

    it('is not fine', () => {});
  });
});`,
          10,
          5,
        ],
        [
          `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    describe('some more', () => {
      beforeEach(() => {});
      beforeEach(() => {});

      it('is not fine', () => {});
    });
  });
});`,
          11,
          7,
        ],
        [
          `describe.each\`\`('%s', () => {
  beforeEach(() => {});
  beforeEach(() => {});

  it('is fine', () => {});
});`,
          3,
          3,
        ],
        [
          `describe('something', () => {
  describe.each\`\`('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each\`\`('%s', () => {
    beforeEach(() => {});
    beforeEach(() => {});

    it('is not fine', () => {});
  });
});`,
          10,
          5,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream describe.each blocks / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeEach in describe block',
          line,
          column,
          endLine: line,
          endColumn: column + 20,
        },
      ],
    })),
  ],
});
