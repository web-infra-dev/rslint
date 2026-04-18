import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-self-compare', {
  valid: [
    // ---- Upstream ESLint suite ----
    `if (x === y) { }`,
    `if (1 === 2) { }`,
    `y=x*x`,
    `foo.bar.baz === foo.bar.qux`,
    `class C { #field; foo() { this.#field === this['#field']; } }`,
    `class C { #field; foo() { this['#field'] === this.#field; } }`,

    // ---- Non-comparison binary operators ----
    `x + x`,
    `x - x`,
    `x * x`,
    `x && x`,
    `x || x`,
    `x ?? x`,
    `x, x`,
    `x = x`,

    // ---- Structurally different operands ----
    `foo() === bar()`,
    `a.b === a.c`,
    `a[0] === a[1]`,
    `a?.b === a.b`,

    // ---- Different literal kinds / values ----
    `1 === 1n`,
    `1 === 2`,
    `'a' === 'b'`,
  ],
  invalid: [
    // ---- Upstream ESLint suite ----
    {
      code: `if (x === x) { }`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 5 }],
    },
    {
      code: `if (x !== x) { }`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 5 }],
    },
    {
      code: `if (x > x) { }`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 5 }],
    },
    {
      code: `if ('x' > 'x') { }`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 5 }],
    },
    {
      code: `do {} while (x === x)`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 14 }],
    },
    {
      code: `x === x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x !== x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x == x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x != x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x > x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x < x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x >= x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `x <= x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `foo.bar().baz.qux >= foo.bar ().baz .qux`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `class C { #field; foo() { this.#field === this.#field; } }`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 27 }],
    },

    // ---- Extra coverage ----
    {
      code: `a.b.c === a.b.c`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `a[0] === a[0]`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `foo() === foo()`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: `(x) === x`,
      errors: [{ messageId: 'comparingToSelf', line: 1, column: 1 }],
    },
    {
      code: 'if (\n  x\n  ===\n  x\n) {}',
      errors: [{ messageId: 'comparingToSelf', line: 2, column: 3 }],
    },
  ],
});
