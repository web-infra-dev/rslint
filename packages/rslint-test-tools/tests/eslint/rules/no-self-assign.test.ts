import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-self-assign', {
  valid: [
    'var a = a',
    'a = b',
    'a += a',
    'a = +a',
    'a = [a]',
    'a &= a',
    'a |= a',
    'let a = a',
    'const a = a',
    'a = a.b',
    'a = -a',

    // Array destructuring
    '[a] = a',
    '[a = 1] = [a]',
    '[a, b] = [b, a]',
    '[a,, b] = [, b, a]',
    '[x, a] = [...x, a]',
    '[...a] = [...a, 1]',
    '[a, ...b] = [0, ...b, 1]',

    // Object destructuring
    '({a} = a)',
    '({a = 1} = {a})',
    '({a: b} = {a})',
    '({a} = {a: b})',
    '({a} = {a() {}})',
    '({a} = {[a]: a})',
    '({[a]: b} = {[a]: b})',
    '({a, ...b} = {a, ...b})',
    '({a: b} = {a: c})',

    // Member expressions with props:true (default)
    { code: 'a.b = a.c', options: { props: true } },
    { code: 'a.b = c.b', options: { props: true } },
    { code: 'a.b = a[b]', options: { props: true } },
    { code: 'a[b] = a.b', options: { props: true } },
    { code: 'a.b().c = a.b().c', options: { props: true } },
    { code: 'b().c = b().c', options: { props: true } },
    { code: 'a[b + 1] = a[b + 1]', options: { props: true } },
    { code: 'a.null = a[/(?<zero>0)/]', options: { props: true } },
    { code: 'this.x = this.y', options: { props: true } },
    'a[0] = a[1]',

    // Member expressions with props:false
    { code: 'a.b = a.b', options: { props: false } },
    { code: 'a.b.c = a.b.c', options: { props: false } },
    { code: 'a[b] = a[b]', options: { props: false } },
    { code: "a['b'] = a['b']", options: { props: false } },
    { code: 'this.x = this.x', options: { props: false } },
    { code: 'a[0] = a[0]', options: { props: false } },

    // Spread copy
    'a = {...a}',
  ],
  invalid: [
    // Basic identifiers
    {
      code: 'a = a',
      errors: [{ messageId: 'selfAssignment' }],
    },

    // Array destructuring
    {
      code: '[a] = [a]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '[a, b] = [a, b]',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },
    {
      code: '[a, b] = [a, c]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '[a, b] = [, b]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '[a, ...b] = [a, ...b]',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },
    {
      code: '[[a], {b}] = [[a], {b}]',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },

    // Object destructuring
    {
      code: '({a} = {a})',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '({a: b} = {a: b})',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({'a': b} = {'a': b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({a: b} = {'a': b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({'a': b} = {a: b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '({1: b} = {1: b})',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({1: b} = {'1': b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({'1': b} = {1: b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({['a']: b} = {a: b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "({'a': b} = {[`a`]: b})",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '({1: b} = {[1]: b})',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '({a, b} = {a, b})',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },
    {
      code: '({a, b} = {b, a})',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },
    {
      code: '({a, b} = {c, a})',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: '({a: {b}, c: [d]} = {a: {b}, c: [d]})',
      errors: [
        { messageId: 'selfAssignment' },
        { messageId: 'selfAssignment' },
      ],
    },
    {
      code: '({a, b} = {a, ...x, b})',
      errors: [{ messageId: 'selfAssignment' }],
    },

    // Member expressions (props:true default)
    {
      code: 'a.b = a.b',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b.c = a.b.c',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a[b] = a[b]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "a['b'] = a['b']",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b = a.b',
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b.c = a.b.c',
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a[b] = a[b]',
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "a['b'] = a['b']",
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'this.x = this.x',
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a[1] = a[1]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a[0] = a[0]',
      errors: [{ messageId: 'selfAssignment' }],
    },
    // Cross-type member expression: a.b = a['b']
    {
      code: "a.b = a['b']",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "a['b'] = a.b",
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: "a[0] = a['0']",
      errors: [{ messageId: 'selfAssignment' }],
    },
    // Multiline element access
    {
      code: "a[\n    'b'\n] = a[\n    'b'\n]",
      errors: [{ messageId: 'selfAssignment' }],
    },
    // Regex literal vs string literal
    {
      code: "a['/(?<zero>0)/'] = a[/(?<zero>0)/]",
      options: { props: true },
      errors: [{ messageId: 'selfAssignment' }],
    },

    // Optional chaining - still self-assignment
    {
      code: '(a?.b).c = (a?.b).c',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b = a?.b',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a[0] = a?.[0]',
      errors: [{ messageId: 'selfAssignment' }],
    },

    // Logical assignment operators
    {
      code: 'a &&= a',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a ||= a',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a ??= a',
      errors: [{ messageId: 'selfAssignment' }],
    },
  ],
});
