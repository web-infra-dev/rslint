import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('use-isnan', {
  valid: [
    // ── Non-comparison usage ──
    'var x = NaN;',
    'isNaN(NaN) === true;',
    'Number.isNaN(NaN) === true;',
    'isNaN(123);',
    'Number.isNaN(123);',

    // ── Arithmetic operators (not comparison) ──
    'NaN + 1;',
    '1 + NaN;',
    'NaN - 1;',
    'NaN * 2;',
    '2 / NaN;',
    'Number.NaN + 1;',

    // ── Assignment (not comparison) ──
    'var q; if (q = NaN) {}',

    // ── Lookalike identifiers that are NOT NaN ──
    'x === Nan;',
    'x === nan;',
    'x === NAN;',
    'x === Number.Nan;',
    'x === window.NaN;',
    'x === globalThis.NaN;',
    'x === Math.NaN;',
    'x === Number[NaN];',
    'x === Number.NaN.toString();',

    // ── Sequence expression: NaN NOT last ──
    'x === (NaN, 1);',
    'x === (NaN, a);',
    'x === (a, NaN, 1);',
    'x === (Number.NaN, 1);',

    // ── Nested comma: only ONE level resolved (matches ESLint) ──
    'x === (a, (b, NaN));',
    'x === (a, (b, Number.NaN));',

    // ── switch: enforceForSwitchCase: false ──
    {
      code: 'switch(NaN) { case foo: break; }',
      options: { enforceForSwitchCase: false },
    },
    {
      code: 'switch(foo) { case NaN: break; }',
      options: { enforceForSwitchCase: false },
    },

    // ── switch: valid discriminant ──
    'switch(foo) { case bar: break; }',
    'switch(true) { case true: break; }',
    'switch(Nan) {}',
    "switch('NaN') {}",
    'switch(foo(NaN)) {}',
    'switch(foo.NaN) {}',
    'switch((NaN, 1)) {}',
    'switch((Number.NaN, 1)) {}',

    // ── switch: valid case clause ──
    'switch(foo) { case Nan: break; }',
    "switch(foo) { case 'NaN': break; }",
    'switch(foo) { case foo(NaN): break; }',
    'switch(foo) { case foo.NaN: break; }',
    'switch(foo) { case bar: NaN; }',
    'switch(foo) { default: NaN; }',
    'switch(foo) { case (NaN, 1): break; }',

    // ── indexOf: default enforceForIndexOf=false ──
    'foo.indexOf(NaN)',
    'foo.lastIndexOf(NaN)',

    // ── indexOf: enforceForIndexOf=true, valid ──
    {
      code: 'foo.indexOf(bar)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.lastIndexOf(bar)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf(NaN, 0, extra)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.lastIndexOf(NaN, 0, extra)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf(a, NaN)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf()',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf(Nan)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf((NaN, 1))',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.indexOf((Number.NaN, 1))',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: nested comma is NOT resolved (matches ESLint) ──
    {
      code: 'foo.indexOf((a, (b, NaN)))',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: not a method call ──
    {
      code: 'indexOf(NaN)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'lastIndexOf(NaN)',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: case-sensitive / wrong method ──
    {
      code: 'foo.IndexOf(NaN)',
      options: { enforceForIndexOf: true },
    },
    {
      code: 'foo.bar(NaN)',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: computed with identifier (not string literal) ──
    {
      code: 'foo[indexOf](NaN)',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: new expression ──
    {
      code: 'new foo.indexOf(NaN)',
      options: { enforceForIndexOf: true },
    },

    // ── indexOf: indirect call ──
    {
      code: 'foo.indexOf.call(arr, NaN)',
      options: { enforceForIndexOf: true },
    },
  ],
  invalid: [
    // ═══ DIMENSION 1: All comparison operators × NaN ═══
    {
      code: '123 == NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '123 === NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN === "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN == "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '123 != NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '123 !== NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN < "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '"abc" < NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN > "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '"abc" >= NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN <= "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },

    // ═══ DIMENSION 2: All NaN representations ═══
    {
      code: '123 === Number.NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'Number.NaN === "abc";',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: "x === Number['NaN'];",
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: "x !== Number['NaN'];",
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === Number?.NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: "x === Number?.['NaN'];",
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === Number[`NaN`];',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },

    // ═══ DIMENSION 3: Sequence expressions (comma) ═══
    {
      code: 'x = (foo, NaN) === 1;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x = 1 === (foo, NaN);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x = (a, b, NaN) === 1;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x = (a, Number.NaN) === 1;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: "x = (a, Number['NaN']) === 1;",
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x = (1, 2) === NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === (doStuff(), NaN);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    // parens inside comma: NaN wrapped in parens as last comma element
    {
      code: 'x === (a, (NaN));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === (a, ((NaN)));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === (a, (Number.NaN));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },

    // ═══ DIMENSION 4: Parenthesization depth ═══
    {
      code: 'x === (NaN);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === ((NaN));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '((x)) === ((NaN));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'x === ((a, NaN));',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },

    // ═══ DIMENSION 5: NaN both sides ═══
    {
      code: 'NaN === NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN == Number.NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },

    // ═══ DIMENSION 6: Switch discriminant variations ═══
    {
      code: 'switch(NaN) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch(Number.NaN) { case 1: break; }',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: "switch(Number['NaN']) {}",
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch(Number?.NaN) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch((a, NaN)) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch((a, b, NaN)) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch((NaN)) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch(((NaN))) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch((doStuff(), NaN)) {}',
      errors: [{ messageId: 'switchNaN' }],
    },
    {
      code: 'switch((doStuff(), Number.NaN)) {}',
      errors: [{ messageId: 'switchNaN' }],
    },

    // ═══ DIMENSION 7: Case clause variations ═══
    {
      code: 'switch(foo) { case NaN: break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case Number.NaN: break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: "switch(foo) { case Number['NaN']: break; }",
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case Number?.NaN: break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case (NaN): break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case ((NaN)): break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case (a, NaN): break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    {
      code: 'switch(foo) { case (a, Number.NaN): break; }',
      errors: [{ messageId: 'caseNaN' }],
    },
    // multiple NaN cases
    {
      code: 'switch(foo) { case NaN: case Number.NaN: break; }',
      errors: [{ messageId: 'caseNaN' }, { messageId: 'caseNaN' }],
    },
    // switch(NaN) + case NaN = 2 errors
    {
      code: 'switch(NaN) { case NaN: break; }',
      errors: [{ messageId: 'switchNaN' }, { messageId: 'caseNaN' }],
    },
    {
      code: 'switch(Number.NaN) { case Number.NaN: break; }',
      errors: [{ messageId: 'switchNaN' }, { messageId: 'caseNaN' }],
    },

    // ═══ DIMENSION 8: indexOf callee variations ═══
    {
      code: 'foo.indexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.lastIndexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: "foo['indexOf'](NaN)",
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: "foo['lastIndexOf'](NaN)",
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo[`indexOf`](NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo?.indexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo?.lastIndexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf?.(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo?.indexOf?.(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: '(foo?.indexOf)(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: '((foo.indexOf))(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo().indexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.bar.indexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.bar.baz.lastIndexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: '(a || b).indexOf(NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },

    // ═══ DIMENSION 9: indexOf arg variations ═══
    {
      code: 'foo.indexOf(Number.NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: "foo.indexOf(Number['NaN'])",
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf(Number?.NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf((a, NaN))',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf((a, Number.NaN))',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf((NaN))',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf(NaN, 1)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf(NaN, b)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.lastIndexOf(NaN, NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.indexOf(Number.NaN, 1)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },
    {
      code: 'foo.lastIndexOf(Number.NaN)',
      options: { enforceForIndexOf: true },
      errors: [{ messageId: 'indexOfNaN' }],
    },

    // ═══ DIMENSION 10: NaN in nested expression contexts ═══
    {
      code: 'if (NaN === x) {}',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'while (x !== NaN) {}',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'for (; NaN < x;) {}',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: "var t = x === NaN ? 'yes' : 'no';",
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'var u = [NaN === 1];',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'var v = { key: NaN === 1 };',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'function f() { return NaN === x; }',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'var g = () => NaN === x;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'var h = function() { return x === NaN; };',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'console.log(NaN === 1);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'void (NaN === x);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '!(NaN === x);',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
  ],
});
