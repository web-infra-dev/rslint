import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unneeded-ternary', {
  valid: [
    // ---- Upstream ESLint suite ----
    `config.newIsCap = config.newIsCap !== false`,
    `var a = x === 2 ? 'Yes' : 'No';`,
    `var a = x === 2 ? true : 'No';`,
    `var a = x === 2 ? 'Yes' : false;`,
    `var a = x === 2 ? 'true' : 'false';`,
    `var a = foo ? foo : bar;`,
    `var value = 'a';var canSet = true;var result = value || (canSet ? 'unset' : 'can not set')`,
    `var a = foo ? bar : foo;`,
    `foo ? bar : foo;`,
    `var a = f(x ? x : 1)`,
    `f(x ? x : 1);`,
    `foo ? foo : bar;`,
    `var a = foo ? 'Yes' : foo;`,
    {
      code: `var a = foo ? 'Yes' : foo;`,
      options: { defaultAssignment: false },
    },
    {
      code: `var a = foo ? bar : foo;`,
      options: { defaultAssignment: false },
    },
    {
      code: `foo ? bar : foo;`,
      options: { defaultAssignment: false },
    },
  ],
  invalid: [
    // ---- Boolean-literal ternary ----
    {
      code: `var a = x === 2 ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x >= 2 ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x === 1 ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x != 1 ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = foo() ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = !foo() ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = foo + bar ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x instanceof foo ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = foo ? false : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = foo() ? false : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x instanceof foo ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = !foo ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- defaultAssignment: false ----
    {
      code: `var result = value ? value : canSet ? 'unset' : 'can not set'`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 14 }],
    },
    {
      code: `foo ? foo : (bar ? baz : qux)`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },
    {
      code: `function* fn() { foo ? foo : yield bar }`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 18 }],
    },
    {
      code: `var a = foo ? foo : 'No';`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = ((foo)) ? (((((foo))))) : ((((((((((((((bar))))))))))))));`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : c => c;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : c = 0;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : (c => c);`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : (c = 0);`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : (c) => (c);`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : c, d;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = b ? b : (c, d);`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `f(x ? x : 1);`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 3 }],
    },
    {
      code: `x ? x : 1;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },
    {
      code: `var a = foo ? foo : bar;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },
    {
      code: `var a = foo ? foo : a ?? b;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 9 }],
    },

    // ---- TypeScript-specific (AsExpression) ----
    {
      code: `foo as any ? false : true`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 1 }],
    },
    {
      code: `foo ? foo : bar as any`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },

    // ---- Unary-keyword tests (typeof / delete / void / await) ----
    {
      code: `var a = typeof x === 'string' ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = typeof x ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = typeof x === 'string' ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = delete obj.x ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = void x ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `async function f() { return await x ? true : false; }`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 29 }],
    },

    // ---- Update expressions ----
    {
      code: `var a = ++x ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x++ ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Paren-wrapped tests ----
    {
      code: `var a = (foo + bar) ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = (typeof x === 'string') ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Compact-spacing operator inversion ----
    {
      code: `var a = x===1?false:true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = x  ===   1 ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Multi-byte CJK identifiers ----
    {
      code: `var a = 中 === 国 ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- TypeScript: SatisfiesExpression ----
    {
      code: `var a = x satisfies number ? false : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `foo ? foo : bar satisfies number`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },

    // ---- Nested default-assignment: rule reports both inner and outer ----
    {
      code: `foo ? foo : (foo ? foo : c)`,
      options: { defaultAssignment: false },
      errors: [
        { messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 },
        { messageId: 'unnecessaryConditionalAssignment', line: 1, column: 14 },
      ],
    },

    // ---- Boolean() call as test: not a boolean expression in our rule
    // (matches ESLint), so falls through to `!!Boolean(x)`. ----
    {
      code: `var a = Boolean(x) ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Both arms equal but test has side effects: report, no fix ----
    {
      code: `var a = x === 1 ? true : true;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Right-associative conditional as alternate ----
    {
      code: `b ? b : c ? d : e;`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },

    // ---- Function/class expressions as alternate ----
    {
      code: `b ? b : function() { return 1; };`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },
    {
      code: `b ? b : class { static foo() {} };`,
      options: { defaultAssignment: false },
      errors: [{ messageId: 'unnecessaryConditionalAssignment', line: 1, column: 1 }],
    },

    // ---- new / template literal / property access tests ----
    {
      code: `var a = new Foo() ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: 'var a = `abc` ? true : false;',
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = obj["foo"] ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
    {
      code: `var a = obj.foo ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },

    // ---- Optional chain as test ----
    {
      code: `var a = foo?.bar ? true : false;`,
      errors: [{ messageId: 'unnecessaryConditionalExpression', line: 1, column: 9 }],
    },
  ],
});
