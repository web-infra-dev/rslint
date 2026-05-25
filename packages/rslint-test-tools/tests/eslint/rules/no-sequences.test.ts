import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-sequences', {
  valid: [
    // ---- Upstream ESLint suite ----
    `var arr = [1, 2];`,
    `var obj = {a: 1, b: 2};`,
    `var a = 1, b = 2;`,
    `var foo = (1, 2);`,
    `(0,eval)("foo()");`,
    `for (i = 1, j = 2;; i++, j++);`,
    `foo(a, (b, c), d);`,
    `do {} while ((doSomething(), !!test));`,
    `for ((doSomething(), somethingElse()); (doSomething(), !!test); );`,
    `if ((doSomething(), !!test));`,
    `switch ((doSomething(), val)) {}`,
    `while ((doSomething(), !!test));`,
    `with ((doSomething(), val)) {}`,
    `a => ((doSomething(), a))`,

    // Options object without the "allowInParentheses" property
    { code: `var foo = (1, 2);`, options: [{}] as any },

    // Explicitly set option "allowInParentheses" to default value
    {
      code: `var foo = (1, 2);`,
      options: [{ allowInParentheses: true }] as any,
    },

    // allowInParentheses: false — for-init / for-update are always allowed
    {
      code: `for ((i = 0, j = 0); test; );`,
      options: [{ allowInParentheses: false }] as any,
    },
    {
      code: `for (; test; (i++, j++));`,
      options: [{ allowInParentheses: false }] as any,
    },

    // https://github.com/eslint/eslint/issues/14572 — return of a parenthesised sequence
    `const foo = () => { return ((bar = 123), 10) }`,
    `const foo = () => (((bar = 123), 10));`,

    // ---- Extra containers: single pair of parens is enough to exempt.
    `for (x in (a, b)) {}`,
    `for (x of (a, b)) {}`,
    `function f() { throw (a, b); }`,
    `function f() { return (a, b); }`,

    // TypeScript: `as` / `satisfies` bind tighter than comma, so sequence
    // must be wrapped for the AssignmentExpression position.
    `const x = (a, b) as number;`,
    `const x = (a, b) satisfies number;`,
    `const x = ((a, b)) as number;`,

    // Template literal / tagged template substitution
    'const x = `${(a, b)}`;',
    'const x = tag`${(a, b)}`;',

    // Function / arrow parameter default value
    `function f(x = (a, b)) {}`,
    `const f = (x = (a, b)) => x;`,

    // Optional-chain / element access — parenthesised sequence is fine
    `foo?.((a, b));`,
    `foo?.[(a, b)];`,
    `foo[(a, b)];`,

    // Class field initializer
    `class C { x = (a, b); }`,
    `class C { static x = (a, b); }`,

    // Object computed key
    `const obj = { [(a, b)]: 1 };`,

    // Conditional-expression slots
    `const x = (a, b) ? c : d;`,
    `const x = a ? (b, c) : d;`,
    `const x = a ? b : (c, d);`,
  ],
  invalid: [
    // ---- Upstream ESLint suite ----
    {
      code: `1, 2;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 2 }],
    },
    {
      code: `a = 1, 2`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 6 }],
    },
    {
      code: `do {} while (doSomething(), !!test);`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 27 }],
    },
    {
      code: `for (; doSomething(), !!test; );`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 21 }],
    },
    {
      code: `if (doSomething(), !!test);`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 18 }],
    },
    {
      code: `switch (doSomething(), val) {}`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 22 }],
    },
    {
      code: `while (doSomething(), !!test);`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 21 }],
    },
    {
      code: `with (doSomething(), val) {}`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 20 }],
    },
    {
      code: `a => (doSomething(), a)`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 20 }],
    },
    {
      code: `(1), 2`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 4 }],
    },
    {
      code: `((1)) , (2)`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 7 }],
    },
    {
      code: `while((1) , 2);`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 11 }],
    },

    // ---- allowInParentheses: false — sequences flagged even inside parens ----
    {
      code: `var foo = (1, 2);`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 13 }],
    },
    {
      code: `(0,eval)("foo()");`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 3 }],
    },
    {
      code: `foo(a, (b, c), d);`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 10 }],
    },
    {
      code: `do {} while ((doSomething(), !!test));`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 28 }],
    },
    {
      code: `for (; (doSomething(), !!test); );`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 22 }],
    },
    {
      code: `if ((doSomething(), !!test));`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 19 }],
    },
    {
      code: `switch ((doSomething(), val)) {}`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 23 }],
    },
    {
      code: `while ((doSomething(), !!test));`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 22 }],
    },
    {
      code: `with ((doSomething(), val)) {}`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 21 }],
    },
    {
      code: `a => ((doSomething(), a))`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 21 }],
    },

    // ---- for-in / for-of RHS ----
    {
      code: `for (x in a, b) {}`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 12 }],
    },
    {
      code: `for (x in (a, b)) {}`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 13 }],
    },
    {
      code: `for (x of (a, b)) {}`,
      options: [{ allowInParentheses: false }] as any,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 13 }],
    },

    // ---- throw / return ----
    {
      code: `function f() { throw a, b; }`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 23 }],
    },
    {
      code: `function f() { return a, b; }`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 24 }],
    },
    {
      code: `() => { throw a, b; }`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 16 }],
    },

    // ---- TypeScript `as` / `satisfies` bind tighter than comma ----
    {
      code: `a, b as number;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 2 }],
    },
    {
      code: `a, b satisfies number;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 2 }],
    },

    // ---- Template literal / tagged template substitution ----
    {
      code: '`${a, b}`;',
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 5 }],
    },
    {
      code: 'tag`${a, b}`;',
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 8 }],
    },

    // ---- Element access with comma inside brackets ----
    {
      code: `foo[a, b];`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 6 }],
    },
    {
      code: `foo?.[a, b];`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 8 }],
    },

    // ---- Conditional-expression boundary (`a ? b : c, d`) ----
    {
      code: `a ? b : c, d;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 10 }],
    },

    // ---- Multi-byte position assertions (UTF-16 code units) ----
    // Surrogate pair in a string literal — emoji counts as 2 UTF-16 units
    {
      code: `'🍎', 1;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 5 }],
    },
    // BMP CJK identifier — each is 1 UTF-16 unit but 3 UTF-8 bytes
    {
      code: `中, 文;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 2 }],
    },
    {
      code: 'function f() {\n  return "🍎", 1;\n}',
      errors: [{ messageId: 'unexpectedCommaExpression', line: 2, column: 14 }],
    },

    // ---- Longer chain (4 elements) — still reports once, leftmost comma
    {
      code: `a, b, c, d;`,
      errors: [{ messageId: 'unexpectedCommaExpression', line: 1, column: 2 }],
    },
  ],
});
