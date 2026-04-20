import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-throw-literal', {
  valid: [
    // ---- ESLint upstream valid cases ----
    'throw new Error();',
    "throw new Error('error');",
    "throw Error('error');",
    'var e = new Error(); throw e;',
    'try {throw new Error();} catch (e) {throw e;};',
    'throw a;', // Identifier
    'throw foo();', // CallExpression
    'throw new foo();', // NewExpression
    'throw foo.bar;', // PropertyAccessExpression
    'throw foo[bar];', // ElementAccessExpression
    'class C { #field: any; foo() { throw foo.#field; } }', // private member
    'throw foo = new Error();', // AssignmentExpression `=`
    "throw foo.bar ||= 'literal'", // logical-assign `||=`
    "throw foo[bar] ??= 'literal'", // logical-assign `??=`
    'throw 1, 2, new Error();', // SequenceExpression
    "throw 'literal' && new Error();", // LogicalExpression `&&`
    "throw new Error() || 'literal';", // LogicalExpression `||`
    "throw foo ? new Error() : 'literal';", // ConditionalExpression
    "throw foo ? 'literal' : new Error();", // ConditionalExpression
    'throw tag `${foo}`;', // TaggedTemplateExpression
    'function* foo() { var index = 0; throw yield index++; }', // YieldExpression
    'async function foo() { throw await bar; }', // AwaitExpression
    'throw obj?.foo', // optional chain (PropertyAccess)
    'throw obj?.foo()', // optional chain (CallExpression)

    // ---- TS / paren wrappers ----
    'throw (foo);',
    'throw ((new Error()));',
    'throw cond ? new Error() : foo();',
    'throw (1, new Error());',
    'throw (1, undefined);',
    'throw foo &&= new Error();',
    'throw a || b || new Error();',
    'throw a ?? new Error();',
    "throw (cond ? new Error() : foo()) || 'literal';",
    'throw obj?.[key];',
    'throw (((foo)));',
    'async function* foo() { throw await (yield 1); }',
    // OUTER access shape drives classification — `foo!.bar` is a top-level
    // PropertyAccessExpression so it could be Error.
    'throw foo!.bar;',
    'throw (foo as Error).bar;',
    // ---- JSX coverage lives in Go tests (no_throw_literal_test.go Tsx:true) ----
  ],
  invalid: [
    // ---- ESLint upstream invalid cases ----
    {
      code: "throw 'error';",
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw 0;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw false;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw null;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw {};',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw undefined;',
      errors: [{ messageId: 'undef' }],
    },

    // ---- String concatenation ----
    {
      code: "throw 'a' + 'b';",
      errors: [{ messageId: 'object' }],
    },
    {
      code: "var b = new Error(); throw 'a' + b;",
      errors: [{ messageId: 'object' }],
    },

    // ---- AssignmentExpression ----
    {
      code: "throw foo = 'error';",
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw foo += new Error();',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw foo &= new Error();',
      errors: [{ messageId: 'object' }],
    },
    {
      code: "throw foo &&= 'literal'",
      errors: [{ messageId: 'object' }],
    },

    // ---- SequenceExpression ----
    {
      code: 'throw new Error(), 1, 2, 3;',
      errors: [{ messageId: 'object' }],
    },

    // ---- LogicalExpression ----
    {
      code: "throw 'literal' && 'not an Error';",
      errors: [{ messageId: 'object' }],
    },
    {
      code: "throw foo && 'literal'",
      errors: [{ messageId: 'object' }],
    },

    // ---- ConditionalExpression ----
    {
      code: "throw foo ? 'not an Error' : 'literal';",
      errors: [{ messageId: 'object' }],
    },

    // ---- TemplateLiteral ----
    {
      code: 'throw `${err}`;',
      errors: [{ messageId: 'object' }],
    },

    // ---- Extra TS edge cases ----
    {
      code: "throw ('error');",
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw (undefined);',
      errors: [{ messageId: 'undef' }],
    },
    {
      code: "throw (foo(), 'error');",
      errors: [{ messageId: 'object' }],
    },
    {
      code: "throw cond ? 'a' : 'b';",
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw 1n;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw /foo/;',
      errors: [{ messageId: 'object' }],
    },
    // `void 0` is a UnaryExpression, not Identifier `undefined` → "object".
    {
      code: 'throw void 0;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw class {};',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw function() {};',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw () => {};',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw [];',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'class C { foo() { throw this; } }',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'function foo() { throw new.target; }',
      errors: [{ messageId: 'object' }],
    },
    {
      code: "throw a && b && 'literal';",
      errors: [{ messageId: 'object' }],
    },
    // ---- TS assertion wrappers — NOT transparent in upstream ESLint ----
    {
      code: 'throw foo as Error;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw foo!;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw foo satisfies unknown;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw <Error>foo;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw obj?.foo!;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: "throw 'foo' as Error;",
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw (5 as any);',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw (((undefined)));',
      errors: [{ messageId: 'undef' }],
    },
    {
      code: 'throw undefined as any;',
      errors: [{ messageId: 'object' }],
    },
    {
      code: 'throw undefined!;',
      errors: [{ messageId: 'object' }],
    },
    // Multi-line throw expression — diagnostic range must span both lines.
    {
      code: "throw 'a' +\n  'b';",
      errors: [{ messageId: 'object' }],
    },
  ],
});
