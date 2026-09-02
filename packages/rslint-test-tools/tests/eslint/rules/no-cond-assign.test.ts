import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-cond-assign', {
  valid: [
    // Upstream ESLint v10.9.1 suite.
    `var x = 0; if (x == 0) { var b = 1; }`,
    {
      code: `var x = 0; if (x == 0) { var b = 1; }`,
      options: ['always'] as any,
    },
    `var x = 5; while (x < 5) { x = x + 1; }`,
    `if ((someNode = someNode.parentNode) !== null) { }`,
    {
      code: `if ((someNode = someNode.parentNode) !== null) { }`,
      options: ['except-parens'] as any,
    },
    `if ((a = b));`,
    `while ((a = b));`,
    `do {} while ((a = b));`,
    `for (;(a = b););`,
    `for (;;) {}`,
    `if (someNode || (someNode = parentNode)) { }`,
    `while (someNode || (someNode = parentNode)) { }`,
    `do { } while (someNode || (someNode = parentNode));`,
    `for (;someNode || (someNode = parentNode););`,
    {
      code: `if ((function(node) { return node = parentNode; })(someNode)) { }`,
      options: ['except-parens'] as any,
    },
    {
      code: `if ((function(node) { return node = parentNode; })(someNode)) { }`,
      options: ['always'] as any,
    },
    {
      code: `if ((node => node = parentNode)(someNode)) { }`,
      options: ['except-parens'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `if ((node => node = parentNode)(someNode)) { }`,
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `if (function(node) { return node = parentNode; }) { }`,
      options: ['except-parens'] as any,
    },
    {
      code: `if (function(node) { return node = parentNode; }) { }`,
      options: ['always'] as any,
    },
    { code: `x = 0;`, options: ['always'] as any },
    `var x; var b = (x === 0) ? 1 : 0;`,
    {
      code: `switch (foo) { case a = b: bar(); }`,
      options: ['except-parens'] as any,
    },
    {
      code: `switch (foo) { case a = b: bar(); }`,
      options: ['always'] as any,
    },
    {
      code: `switch (foo) { case baz + (a = b): bar(); }`,
      options: ['always'] as any,
    },

    // A method's FunctionExpression still shields its own executable slots.
    {
      code: `if (class { method() { value = next() } }) {}`,
      options: ['always'] as any,
    },
    {
      code: `if (class { method(parameter = value = next()) {} }) {}`,
      options: ['always'] as any,
    },
    {
      code: `if (class { [() => value = next()]() {} }) {}`,
      options: ['always'] as any,
    },
    {
      code: `if (class { @dec(() => value = next()) method() {} }) {}`,
      options: ['always'] as any,
    },
  ],
  invalid: [
    // Upstream ESLint v10.9.1 suite.
    {
      code: `var x; if (x = 0) { var b = 1; }`,
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `var x; while (x = 0) { var b = 1; }`,
      errors: [{ messageId: 'missing' }],
    },
    {
      code: `var x = 0, y; do { y = x; } while (x = x + 1);`,
      errors: [{ messageId: 'missing' }],
    },
    {
      code: `var x; for(; x+=1 ;){};`,
      errors: [{ messageId: 'missing' }],
    },
    {
      code: `var x; if ((x) = (0));`,
      errors: [{ messageId: 'missing' }],
    },
    {
      code: `if (someNode || (someNode = parentNode)) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', column: 18, endColumn: 39 }],
    },
    {
      code: `while (someNode || (someNode = parentNode)) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `do { } while (someNode || (someNode = parentNode));`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `for (; (typeof l === 'undefined' ? (l = 0) : l); i++) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `if (x = 0) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `while (x = 0) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `do { } while (x = x + 1);`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `for(; x = y; ) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `if ((x = 0)) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `while ((x = 0)) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `do { } while ((x = x + 1));`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `for(; (x = y); ) { }`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `var x; var b = (x = 0) ? 1 : 0;`,
      errors: [{ messageId: 'missing' }],
    },
    {
      code: `var x; var b = x && (y = 0) ? 1 : 0;`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `(((3496.29)).bkufyydt = 2e308) ? foo : bar;`,
      errors: [{ messageId: 'missing' }],
    },

    // ts-go combines MethodDefinition and FunctionExpression, but these
    // member-level slots remain outside the function boundary in ESTree.
    {
      code: `if (class { [value = next()]() {} }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 14 }],
    },
    {
      code: `if ({ [value = next()]() {} }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 8 }],
    },
    {
      code: `if (class { get [value = next()]() { return 1 } }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 18 }],
    },
    {
      code: `if (class { set [value = next()](nextValue) {} }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 18 }],
    },
    {
      code: `if (class { static async *[value = next()]() {} }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 28 }],
    },
    {
      code: `if (class { @dec(value = next()) method() {} }) {}`,
      options: ['always'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 18 }],
    },
  ],
});
