import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors the upstream logical-assignment-operators valid/invalid semantic set
// (Layer 1). It verifies rule registration + wire protocol + ESLint-compatible
// diagnostic shape; the exhaustive edge-shape / branch lock-in coverage lives
// in the Go suite (logical_assignment_operators_extras_test.go).
ruleTester.run('logical-assignment-operators', {
  valid: [
    'a || b',
    'a && b',
    'a ?? b',
    'a || a || b',
    'var a = a || b',
    'a === undefined ? a : b',
    'while (a) a = b',
    'a ||= b',
    'a &&= b',
    'a ??= b',
    'a += a || b',
    'a *= a || b',
    'a ||= a || b',
    'a &&= a || b',
    'a = a',
    'a = b',
    'a = a === b',
    'a = a + b',
    'a = a / b',
    'a = fn(a) || b',
    'a = false || c',
    'a = f() || g()',
    'a = b || c',
    'a = b || a',
    'object.a = object.b || c',
    '[a] = a || b',
    '({ a } = a || b)',
    '(a = b) || a',
    'a + (a = b)',
    'a || (b ||= c)',
    'a || (b &&= c)',
    'a || b === 0',
    'a || fn()',
    'a || (b && c)',
    'a || (b ?? c)',
    'a || (b = c)',
    'a || (a ||= b)',
    'fn() || (a = b)',
    'a.b || (a = b)',
    'a?.b || (a.b = b)',
    {
      code: 'class Class { #prop; constructor() { this.#prop || (this.prop = value) } }',
    },
    {
      code: 'class Class { #prop; constructor() { this.prop || (this.#prop = value) } }',
    },
    'if (a) a = b',
    {
      code: 'if (a) a = b',
      options: ['always', { enforceForIfStatements: false }] as any,
    },
    {
      code: 'if (a) { a = b } else {}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) { a = b } else if (a) {}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (unrelated) {} else if (a) a = b; else {}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (unrelated) {} else if (a) a = b; else if (unrelated) {}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) {}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) { before; a = b }',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) { a = b; after }',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) throw new Error()',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) a',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) a ||= b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) b = a',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) { a() }',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) { a += a || b }',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (true) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (predicate(a)) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a?.b) a.b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (!a?.b) a.b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === b) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a != null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null && a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === 0 || a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === 1) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a == null || a == undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === !0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === +0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === undefined || a === void 0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === void void 0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: "if (a === null || a === void 'string') a = b",
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === void fn()) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a == a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a == b) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null == null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (undefined == undefined) undefined = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null == x) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null == fn()) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null === a || a === 0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (0 === a || null === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (1 === a || a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (undefined === a || 1 === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === b) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (b === undefined || a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null === a || b === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null === null || undefined === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null === null || a === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (undefined === undefined || a === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (null === undefined || a === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '{\n   const undefined = 0;\n   if (a == undefined) a = b\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '(() => {\n   const undefined = 0;\n   if (condition) {\n       if (a == undefined) a = b\n   }\n})()',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '{\n   if (a == undefined) a = b\n}\nvar undefined = 0;',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '{\n   const undefined = 0;\n   if (undefined == null) undefined = b\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '{\n   const undefined = 0;\n   if (a === undefined || a === null) a = b\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: '{\n   const undefined = 0;\n   if (undefined === a || null === a) a = b\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (!a) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (!!a) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a == null) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || a === undefined) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || b === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (a === null || b === undefined) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'if (Boolean(a)) b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    {
      code: 'function fn(Boolean) {\n   if (Boolean(a)) a = b\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
    },
    { code: 'a = a || b', options: ['never'] as any },
    { code: 'a = a && b', options: ['never'] as any },
    { code: 'a = a ?? b', options: ['never'] as any },
    { code: 'a = b', options: ['never'] as any },
    { code: 'a += b', options: ['never'] as any },
    { code: 'a -= b', options: ['never'] as any },
    { code: 'a.b = a.b || c', options: ['never'] as any },
    { code: 'a = a && b || c', options: ['always'] as any },
    { code: 'a = a && b && c || d', options: ['always'] as any },
    { code: 'a = (a || b) || c', options: ['always'] as any },
    { code: 'a = (a && b) && c', options: ['always'] as any },
    { code: 'a = (a ?? b) ?? c', options: ['always'] as any },
  ],
  invalid: [
    { code: 'a = a || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a && b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a ?? b', errors: [{ messageId: 'assignment' }] },
    { code: 'foo = foo || bar', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || fn()', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || b && c', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || (b || c)', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || (b ? c : d)', errors: [{ messageId: 'assignment' }] },
    { code: '/* before */ a = a || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || b // after', errors: [{ messageId: 'assignment' }] },
    { code: 'a /* between */ = a || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = /** @type */ a || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || /* between */ b', errors: [{ messageId: 'assignment' }] },
    { code: '(a) = a || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = (a) || b', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || (b)', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || ((b))', errors: [{ messageId: 'assignment' }] },
    { code: '(a = a || b)', errors: [{ messageId: 'assignment' }] },
    { code: 'a = a || (f(), b)', errors: [{ messageId: 'assignment' }] },
    { code: 'a.b = a.b ?? c', errors: [{ messageId: 'assignment' }] },
    { code: 'a.b.c = a.b.c ?? d', errors: [{ messageId: 'assignment' }] },
    { code: 'a[b] = a[b] ?? c', errors: [{ messageId: 'assignment' }] },
    { code: "a['b'] = a['b'] ?? c", errors: [{ messageId: 'assignment' }] },
    { code: "a.b = a['b'] ?? c", errors: [{ messageId: 'assignment' }] },
    { code: "a['b'] = a.b ?? c", errors: [{ messageId: 'assignment' }] },
    {
      code: 'this.prop = this.prop ?? {}',
      errors: [{ messageId: 'assignment' }],
    },
    { code: 'with (object) a = a || b', errors: [{ messageId: 'assignment' }] },
    {
      code: 'with (object) { a = a || b }',
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'with (object) { if (condition) a = a || b }',
      errors: [{ messageId: 'assignment' }],
    },
    { code: 'with (a = a || b) {}', errors: [{ messageId: 'assignment' }] },
    {
      code: 'with (object) {} a = a || b',
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || b; with (object) {}',
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'if (condition) a = a || b',
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'with (object) {\n  "use strict";\n   a = a || b\n}',
      errors: [{ messageId: 'assignment' }],
    },
    { code: 'fn(a = a || b)', errors: [{ messageId: 'assignment' }] },
    { code: 'fn((a = a || b))', errors: [{ messageId: 'assignment' }] },
    { code: '(a = a || b) ? c : d', errors: [{ messageId: 'assignment' }] },
    { code: 'a = b = b || c', errors: [{ messageId: 'assignment' }] },
    { code: 'a || (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a && (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a ?? (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'foo ?? (foo = bar)', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = 0)', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = fn())', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = (b || c))', errors: [{ messageId: 'logical' }] },
    { code: '(a) || (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a || ((a) = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = (b))', errors: [{ messageId: 'logical' }] },
    { code: 'a || ((a = b))', errors: [{ messageId: 'logical' }] },
    { code: 'a || (((a = b)))', errors: [{ messageId: 'logical' }] },
    { code: 'a || ( ( a = b ) )', errors: [{ messageId: 'logical' }] },
    { code: '/* before */ a || (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = b) // after', errors: [{ messageId: 'logical' }] },
    { code: 'a /* between */ || (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a || /* between */ (a = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a.b || (a.b = c)', errors: [{ messageId: 'logical' }] },
    {
      code: 'class Class { #prop; constructor() { this.#prop || (this.#prop = value) } }',
      errors: [{ messageId: 'logical' }],
    },
    { code: "a['b'] || (a['b'] = c)", errors: [{ messageId: 'logical' }] },
    { code: 'a[0] || (a[0] = b)', errors: [{ messageId: 'logical' }] },
    { code: 'a[this] || (a[this] = b)', errors: [{ messageId: 'logical' }] },
    { code: 'foo.bar || (foo.bar = baz)', errors: [{ messageId: 'logical' }] },
    { code: 'a.b.c || (a.b.c = d)', errors: [{ messageId: 'logical' }] },
    { code: 'a[b.c] || (a[b.c] = d)', errors: [{ messageId: 'logical' }] },
    { code: 'a[b?.c] || (a[b?.c] = d)', errors: [{ messageId: 'logical' }] },
    {
      code: 'with (object) a.b || (a.b = c)',
      errors: [{ messageId: 'logical' }],
    },
    { code: 'a = a.b || (a.b = {})', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = 0) || b', errors: [{ messageId: 'logical' }] },
    { code: '(a || (a = 0)) || b', errors: [{ messageId: 'logical' }] },
    { code: 'a || (b || (b = 0))', errors: [{ messageId: 'logical' }] },
    { code: 'a = b || (b = c)', errors: [{ messageId: 'logical' }] },
    { code: 'a || (a = 0) ? b : c', errors: [{ messageId: 'logical' }] },
    { code: 'fn(a || (a = 0))', errors: [{ messageId: 'logical' }] },
    {
      code: 'if (a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (Boolean(a)) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (!!a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (!a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (!Boolean(a)) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a == undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a == null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === null || a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === undefined || a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === null || a === void 0) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === void 0 || a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b; }',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: '{ const undefined = 0; }\nif (a == undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a == undefined) a = b\n{ const undefined = 0; }',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (null == a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (undefined == a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (undefined === a || a === null) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === undefined || null === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (undefined === a || null === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (null === a || a === undefined) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a === null || undefined === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (null === a || undefined === a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if ((a)) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) a = (b)',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) (a = b)',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: ';if (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: '{ if (a) (a) = b }',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'fn();if (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'fn()\nif (a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'id\nif (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'object.prop\nif (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'object[computed]\nif (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'fn()\nif (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) a = b; fn();',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b }',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b; }\nfn();',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b }\nfn();',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b } fn();',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) { a = b\n} fn();',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) a  =  b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a)\n a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) {\n a = b; \n}',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: '/* before */ if (a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) a = b /* after */',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) /* between */ a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) a = /* between */ b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a.b) a.b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a[b]) a[b] = c',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: "if (a['b']) a['b'] = c",
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (this.prop) this.prop = value',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: '(class extends SuperClass { method() { if (super.prop) super.prop = value } })',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'with (object) if (a) a = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a.b === undefined || a.b === null) a.b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a.b.c) a.b.c = d',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a.b.c.d) a.b.c.d = e',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a[b].c) a[b].c = d',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'with (object) if (a.b) a.b = c',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} else if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (a) {} else if (b) {} else if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} else\nif (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {\n}\nelse if (a) {\na = b;\n}',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) statement; else if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) id\nelse if (a) (a) = b',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} else if (a) a = b; else if (c) c = d',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) { /* body */ } else if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} /* before else */ else if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} else // Line\nif (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (unrelated) {} else /* Block */ if (a) a = b;',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'if (array) array = array.filter(predicate)',
      options: ['always', { enforceForIfStatements: true }] as any,
      errors: [{ messageId: 'if' }],
    },
    {
      code: 'a ||= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a &&= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo ||= bar',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a.b ||= c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a[b] ||= c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "a['b'] ||= c",
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'this.prop ||= 0',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'with (object) a ||= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a) ||= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= (b)',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a ||= b)',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '/* before */ a ||= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= b // after',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a /* before */ ||= b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= /* after */ b',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= b && c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a &&= b || c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= b || c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a &&= b && c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= b || c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= b && c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= b ?? c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= (b || c)',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ??= b + c',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a ||= b as number;',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a.b.c || (a.b.c = d as number)',
      errors: [{ messageId: 'logical' }],
    },
    {
      code: 'a.b.c || (a.b.c = (d as number))',
      errors: [{ messageId: 'logical' }],
    },
    {
      code: '(a.b.c || (a.b.c = d)) as number',
      errors: [{ messageId: 'logical' }],
    },
    {
      code: 'a = a || b || c',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a && b && c',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a ?? b ?? c',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || b && c',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || b || c || d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a && b && c && d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a ?? b ?? c ?? d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || b || c && d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || b && c || d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = (a) || b || c',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = a || (b || c) || d',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = (a || b || c)',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
    {
      code: 'a = ((a) || (b || c) || d)',
      options: ['always'] as any,
      errors: [{ messageId: 'assignment' }],
    },
  ],
});
