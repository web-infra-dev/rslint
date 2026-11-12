// Upstream: ESLint v10.8.1 tests/lib/rules/constructor-super.js.
// Inputs and existing expectations are preserved; full messages and ranges
// are enriched from the pinned ESLint version. This rule has no edits.
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { expect, test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({ languageOptions: { ecmaVersion: 2021 } });

const cases = {
  valid: [
    // non derived classes.
    'class A { }',
    'class A { constructor() { } }',

    /*
     * inherit from non constructors.
     * those are valid if we don't define the constructor.
     */
    'class A extends null { }',

    // derived classes.
    'class A extends B { }',
    'class A extends B { constructor() { super(); } }',
    'class A extends B { constructor() { if (true) { super(); } else { super(); } } }',
    'class A extends (class B {}) { constructor() { super(); } }',
    'class A extends (B = C) { constructor() { super(); } }',
    'class A extends (B &&= C) { constructor() { super(); } }',
    'class A extends (B ||= C) { constructor() { super(); } }',
    'class A extends (B ??= C) { constructor() { super(); } }',
    'class A extends (B ||= 5) { constructor() { super(); } }',
    'class A extends (B ??= 5) { constructor() { super(); } }',
    'class A extends (B || C) { constructor() { super(); } }',
    'class A extends (5 && B) { constructor() { super(); } }',

    // A future improvement could detect the left side as statically falsy, making this invalid.
    'class A extends (false && B) { constructor() { super(); } }',
    'class A extends (B || 5) { constructor() { super(); } }',
    'class A extends (B ?? 5) { constructor() { super(); } }',

    'class A extends (a ? B : C) { constructor() { super(); } }',
    'class A extends (B, C) { constructor() { super(); } }',

    // nested.
    'class A { constructor() { class B extends C { constructor() { super(); } } } }',
    'class A extends B { constructor() { super(); class C extends D { constructor() { super(); } } } }',
    'class A extends B { constructor() { super(); class C { constructor() { } } } }',

    // multi code path.
    'class A extends B { constructor() { a ? super() : super(); } }',
    'class A extends B { constructor() { if (a) super(); else super(); } }',
    'class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); } } }',
    'class A extends B { constructor() { try {} finally { super(); } } }',
    'class A extends B { constructor() { if (a) throw Error(); super(); } }',

    // returning value is a substitute of 'super()'.
    'class A extends B { constructor() { if (true) return a; super(); } }',
    'class A extends null { constructor() { return a; } }',
    'class A { constructor() { return a; } }',

    // https://github.com/eslint/eslint/issues/5261
    'class A extends B { constructor(a) { super(); for (const b of a) { this.a(); } } }',
    'class A extends B { constructor(a) { super(); for (b in a) ( foo(b) ); } }',

    // https://github.com/eslint/eslint/issues/5319
    'class Foo extends Object { constructor(method) { super(); this.method = method || function() {}; } }',

    // https://github.com/eslint/eslint/issues/5394
    [
      'class A extends Object {',
      '    constructor() {',
      '        super();',
      '        for (let i = 0; i < 0; i++);',
      '    }',
      '}',
    ].join('\n'),
    [
      'class A extends Object {',
      '    constructor() {',
      '        super();',
      '        for (; i < 0; i++);',
      '    }',
      '}',
    ].join('\n'),
    [
      'class A extends Object {',
      '    constructor() {',
      '        super();',
      '        for (let i = 0;; i++) {',
      '            if (foo) break;',
      '        }',
      '    }',
      '}',
    ].join('\n'),
    [
      'class A extends Object {',
      '    constructor() {',
      '        super();',
      '        for (let i = 0; i < 0;);',
      '    }',
      '}',
    ].join('\n'),
    [
      'class A extends Object {',
      '    constructor() {',
      '        super();',
      '        for (let i = 0;;) {',
      '            if (foo) break;',
      '        }',
      '    }',
      '}',
    ].join('\n'),

    // https://github.com/eslint/eslint/issues/8848
    `
            class A extends B {
                constructor(props) {
                    super(props);

                    try {
                        let arr = [];
                        for (let a of arr) {
                        }
                    } catch (err) {
                    }
                }
            }
        `,

    // Optional chaining
    'class A extends obj?.prop { constructor() { super(); } }',

    `
            class A extends Base {
                constructor(list) {
                    for (const a of list) {
                        if (a.foo) {
                            super(a);
                            return;
                        }
                    }
                    super();
                }
            }
        `,
  ],
  invalid: [
    // inherit from non constructors.
    {
      code: 'class A extends null { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'class A extends null { constructor() { } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'class A extends 100 { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: "class A extends 'test' { constructor() { super(); } }",
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'class A extends (B = 5) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'class A extends (B && 5) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      // `B &&= 5` evaluates either to a falsy value of `B` (which, then, cannot be a constructor), or to '5'
      code: 'class A extends (B &&= 5) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'class A extends (B += C) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'class A extends (B -= C) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'class A extends (B **= C) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'class A extends (B |= C) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'class A extends (B &= C) { constructor() { super(); } }',
      errors: [
        {
          messageId: 'badSuper',
          message: "Unexpected 'super()' because 'super' is not a constructor.",
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // derived classes.
    {
      code: 'class A extends B { constructor() { } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { for (var a of b) super.foo(); } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 68,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { for (var i = 1; i < 10; i++) super.foo(); } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 80,
        },
      ],
    },

    // nested execution scope.
    {
      code: 'class A extends B { constructor() { var c = class extends D { constructor() { super(); } } } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 93,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { var c = () => super(); } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { class C extends D { constructor() { super(); } } } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 87,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { var C = class extends D { constructor() { super(); } } } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 93,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { super(); class C extends D { constructor() { } } } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 83,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { super(); var C = class extends D { constructor() { } } } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 72,
          endLine: 1,
          endColumn: 89,
        },
      ],
    },

    // lacked in some code path.
    {
      code: 'class A extends B { constructor() { if (a) super(); } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { if (a); else super(); } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { a && super(); } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { switch (a) { case 0: super(); } } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { switch (a) { case 0: break; default: super(); } } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 86,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { try { super(); } catch (err) {} } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { try { a; } catch (err) { super(); } } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 74,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { if (a) return; super(); } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },

    // duplicate.
    {
      code: 'class A extends B { constructor() { super(); super(); } }',
      errors: [
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { super() || super(); } }',
      errors: [
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { if (a) super(); super(); } }',
      errors: [
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { switch (a) { case 0: super(); default: super(); } } }',
      errors: [
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 1,
          column: 76,
          endLine: 1,
          endColumn: 83,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(a) { while (a) super(); } }',
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 58,
        },
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // ignores `super()` on unreachable paths.
    {
      code: 'class A extends B { constructor() { return; super(); } }',
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // https://github.com/eslint/eslint/issues/8248
    {
      code: `class Foo extends Bar {
                constructor() {
                    for (a in b) for (c in d);
                }
            }`,
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 2,
          column: 17,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },

    {
      code: `class C extends D {

                constructor() {
                    do {
                        something();
                    } while (foo);
                }

            }`,
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 3,
          column: 17,
          endLine: 7,
          endColumn: 18,
        },
      ],
    },
    {
      code: `class C extends D {

                constructor() {
                    for (let i = 1;;i++) {
                        if (bar) {
                            break;
                        }
                    }
                }

            }`,
      errors: [
        {
          messageId: 'missingAll',
          message: "Expected to call 'super()'.",
          line: 3,
          column: 17,
          endLine: 9,
          endColumn: 18,
        },
      ],
    },
    {
      code: `class C extends D {

                constructor() {
                    do {
                        super();
                    } while (foo);
                }

            }`,
      errors: [
        {
          messageId: 'duplicate',
          message: "Unexpected duplicate 'super()'.",
          line: 5,
          column: 25,
          endLine: 5,
          endColumn: 32,
        },
      ],
    },
    {
      code: `class C extends D {

                constructor() {
                    while (foo) {
                        if (bar) {
                            super();
                            break;
                        }
                    }
                }

            }`,
      errors: [
        {
          messageId: 'missingSome',
          message: "Lacked a call of 'super()' in some code paths.",
          line: 3,
          column: 17,
          endLine: 10,
          endColumn: 18,
        },
      ],
    },
  ],
};

ruleTester.run('constructor-super', cases);

// Exercise the actual fix pipeline as well as the tester's no-edit assertions.
test('autofix leaves every upstream constructor-super case unchanged', async () => {
  const root = await fs.realpath(
    await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-constructor-super-')),
  );
  try {
    const codes = [...cases.valid, ...cases.invalid.map((item) => item.code)];
    const result = await lint({
      workingDirectory: root,
      configDirectory: root,
      config: [
        {
          languageOptions: {
            ecmaVersion: 2021,
            parserOptions: { projectService: false },
          },
          rules: { 'constructor-super': 'error' },
        },
      ],
      fileContents: Object.fromEntries(
        codes.map((code, i) => [path.join(root, `case-${i}.js`), code]),
      ),
      fix: true,
    });
    expect(result.fileCount).toBe(codes.length);
    expect(result.diagnostics).toHaveLength(41);
    expect(result.output).toBeUndefined();
    for (const diagnostic of result.diagnostics) {
      expect(diagnostic.fixes).toBeUndefined();
      expect(diagnostic.suggestions).toBeUndefined();
    }
  } finally {
    await fs.rm(root, { recursive: true, force: true });
  }
});
