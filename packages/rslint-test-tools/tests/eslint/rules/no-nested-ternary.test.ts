import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-nested-ternary', {
  valid: [
    'foo ? doBar() : doBaz();',
    'var foo = bar === baz ? qux : quxx;',
    'a ? b : c; d ? e : f;',
    // Test position is a ternary — ESLint only checks consequent/alternate.
    'var x = (a ? b : c) ? d : e;',
    // Ternary inside an arrow / object / array / call is not a direct branch.
    'var x = a ? () => b : () => (c ? d : e);',
    'var x = a ? { k: b ? c : d } : e;',
    'var x = a ? [b ? c : d] : e;',
    'var x = a ? foo(b ? c : d) : e;',
    // TypeScript outer expressions are not stripped by ESTree.
    'var x = a ? (b ? c : d) as any : e;',
    'var x = a ? (b ? c : d)! : e;',
    'var x = a ? (b ? c : d) satisfies unknown : e;',
    'var x = a ? <any>(b ? c : d) : e;',
    // TypeScript conditional TYPE — different AST kind.
    'type T = A extends B ? C : D extends E ? F : G;',
    // Composite TS wrapper around a nested ternary branch.
    'var x = a ? ((b ? c : d) as any)! : e;',
  ],
  invalid: [
    {
      code: 'foo ? bar : baz === qux ? quxx : foobar;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 1 }],
    },
    {
      code: 'foo ? baz === qux ? quxx : foobar : bar;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 1 }],
    },
    {
      code: 'var a = foo ? (bar ? baz : qux) : quux;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'var a = foo ? bar : (baz ? qux : quux);',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'var a = foo ? ((bar ? baz : qux)) : quux;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'var a = a ? b ? c : d : e ? f : g;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'var a = a ? b : c ? d : e ? f : g;',
      errors: [
        { messageId: 'noNestedTernary', line: 1, column: 9 },
        { messageId: 'noNestedTernary', line: 1, column: 17 },
      ],
    },
    {
      code: 'var a = foo ? /* x */ (bar ? baz : qux) /* y */ : quux;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'foo(a ? b ? c : d : e);',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 5 }],
    },
    {
      code: 'var s = `${a ? b : c ? d : e}`;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 12 }],
    },
    {
      code: 'function f() { return a ? b : c ? d : e; }',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 23 }],
    },
    {
      code: 'const f = () => a ? b : c ? d : e;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 17 }],
    },
    {
      code: 'var a = foo\n  ? bar\n  : baz ? qux : quux;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    // Opaque wrappers don't hide a nested ternary one level below.
    {
      code: '!(a ? b : c ? d : e);',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 3 }],
    },
    {
      code: 'if (a ? b : c ? d : e) {}',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 5 }],
    },
    {
      code: 'var o = { [a ? b : c ? d : e]: 1 };',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 12 }],
    },
    {
      code: 'throw a ? b : c ? d : e;',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 7 }],
    },
    {
      code: 'foo(...(a ? b : c ? d : e));',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 9 }],
    },
    {
      code: 'obj[a ? b : c ? d : e];',
      errors: [{ messageId: 'noNestedTernary', line: 1, column: 5 }],
    },
  ],
});
