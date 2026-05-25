import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-param-reassign', {
  valid: [
    // No reassignment
    'function foo(a) { var b = a; }',
    'function foo(a) { for (b in a); }',
    'function foo(a) { for (b of a); }',

    // Property modifications (props: false, default)
    "function foo(a) { a.prop = 'value'; }",
    'function foo(a) { for (a.prop in obj); }',
    'function foo(a) { for (a.prop of arr); }',
    'function foo(a) { a.b = 0; }',
    'function foo(a) { delete a.b; }',
    'function foo(a) { ++a.b; }',

    // Shadowing and globals
    'function foo(a) { (function() { var a = 12; a++; })(); }',
    'function foo() { someGlobal = 13; }',

    // With props: true - does not flag non-property reads
    {
      code: 'function foo(a) { bar(a.b).c = 0; }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { data[a.b] = 0; }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { +a.b; }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { (a ? [] : [])[0] = 1; }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { (a.b ? [] : [])[0] = 1; }',
      options: { props: true },
    },

    // ignorePropertyModificationsFor
    {
      code: 'function foo(a) { a.b = 0; }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },
    {
      code: 'function foo(a) { ++a.b; }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },
    {
      code: 'function foo(a) { delete a.b; }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },
    {
      code: 'function foo(a) { for (a.b in obj); }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },
    {
      code: 'function foo(a) { for (a.b of arr); }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },
    {
      code: 'function foo(a) { a.b.c = 0; }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
    },

    // ignorePropertyModificationsForRegex
    {
      code: 'function foo(aFoo) { aFoo.b = 0; }',
      options: { props: true, ignorePropertyModificationsForRegex: ['^a.*$'] },
    },
    {
      code: 'function foo(aFoo) { ++aFoo.b; }',
      options: { props: true, ignorePropertyModificationsForRegex: ['^a.*$'] },
    },
    {
      code: 'function foo(aFoo) { delete aFoo.b; }',
      options: { props: true, ignorePropertyModificationsForRegex: ['^a.*$'] },
    },
    {
      code: 'function foo(aFoo) { aFoo.b.c = 0; }',
      options: { props: true, ignorePropertyModificationsForRegex: ['^a.*$'] },
    },

    // Destructuring / loop patterns that do not reassign params
    {
      code: 'function foo(a) { ({ [a]: variable } = value) }',
      options: { props: true },
    },
    'function foo(a) { ([...a.b] = obj); }',
    'function foo(a) { ({...a.b} = obj); }',
    {
      code: 'function foo(a) { for (obj[a.b] in obj); }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { for (obj[a.b] of arr); }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { for (bar in a.b); }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { for (bar of a.b); }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { for (bar in baz) a.b; }',
      options: { props: true },
    },
    {
      code: 'function foo(a) { for (bar of baz) a.b; }',
      options: { props: true },
    },
  ],
  invalid: [
    // Direct reassignment
    {
      code: 'function foo(bar) { bar = 13; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { bar += 13; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { (function() { bar = 13; })(); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { ++bar; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { bar++; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { --bar; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { bar--; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },

    // Destructured parameters
    {
      code: 'function foo({bar}) { bar = 13; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo([, {bar}]) { bar = 13; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { ({bar} = {}); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { ({x: [, bar = 0]} = {}); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },

    // Loop assignment
    {
      code: 'function foo(bar) { for (bar in baz); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(bar) { for (bar of baz); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },

    // Property modification with props: true
    {
      code: 'function foo(bar) { bar.a = 0; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { bar.get(0).a = 0; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { delete bar.a; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { ++bar.a; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { for (bar.a in {}); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { for (bar.a of []); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { (bar ? bar : [])[0] = 1; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { [bar.a] = []; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },

    // Parameter reassignment in destructuring
    {
      code: 'function foo(a) { ({a} = obj); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(a) { ([...a] = obj); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(a) { ({...a} = obj); }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },

    // Spread/rest with property access
    {
      code: 'function foo(a) { ([...a.b] = obj); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(a) { ({...a.b} = obj); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(a) { for ([a.b] of []); }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },

    // Logical assignment operators
    {
      code: 'function foo(a) { a &&= b; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(a) { a ||= b; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(a) { a ??= b; }',
      errors: [{ messageId: 'assignmentToFunctionParam' }],
    },
    {
      code: 'function foo(a) { a.b &&= c; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(a) { a.b.c ||= d; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(a) { a[b] ??= c; }',
      options: { props: true },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },

    // ignorePropertyModificationsFor bypass
    {
      code: 'function foo(bar) { [bar.a] = []; }',
      options: { props: true, ignorePropertyModificationsFor: ['a'] },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
    {
      code: 'function foo(bar) { [bar.a] = []; }',
      options: { props: true, ignorePropertyModificationsForRegex: ['^B.*$'] },
      errors: [{ messageId: 'assignmentToFunctionParamProp' }],
    },
  ],
});
