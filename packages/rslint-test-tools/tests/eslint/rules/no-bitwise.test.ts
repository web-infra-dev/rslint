import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-bitwise', {
  valid: [
    // Non-bitwise operators
    'a + b',
    '!a',
    'a && b',
    'a || b',
    'a += b',
    'a &&= b',
    'a ||= b',
    'a ??= b',
    // allow
    { code: '~[1, 2, 3].indexOf(1)', options: [{ allow: ['~'] }] as any },
    { code: '~1<<2 === -8', options: [{ allow: ['~', '<<'] }] as any },
    // int32Hint
    { code: 'a|0', options: [{ int32Hint: true }] as any },
    { code: 'a | (0)', options: [{ int32Hint: true }] as any },
    { code: 'a | 0.0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0x0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0b0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0o0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0e0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0E0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0e+0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0e-0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0.', options: [{ int32Hint: true }] as any },
    { code: 'a | .0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0.0e0', options: [{ int32Hint: true }] as any },
    { code: 'a | 0X0', options: [{ int32Hint: true }] as any },
    { code: 'a|0', options: [{ allow: ['|'], int32Hint: false }] as any },
    // Type-level unions/intersections must not be flagged
    'type T = A | B',
    'type T = A & B',
    'let x: string | number = 1',
    'function f<T extends A | B>(x: T): T { return x; }',
  ],
  invalid: [
    { code: 'a ^ b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a | b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a & b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a << b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a >> b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a >>> b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a|0', errors: [{ messageId: 'unexpected' }] },
    { code: '~a', errors: [{ messageId: 'unexpected' }] },
    { code: 'a ^= b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a |= b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a &= b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a <<= b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a >>= b', errors: [{ messageId: 'unexpected' }] },
    { code: 'a >>>= b', errors: [{ messageId: 'unexpected' }] },
    // Nested / chained
    {
      code: 'a | b | c',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: '~~a',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: '~(a | b)',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    { code: '(a | b)', errors: [{ messageId: 'unexpected' }] },
    // int32Hint boundaries
    {
      code: 'a | 1',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a | 0x10',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a | 1.0',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '0 | a',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a & 0',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a | -0',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a | 0n',
      options: [{ int32Hint: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    // TS value-level contexts
    {
      code: 'enum E { A = 1 << 0, B = 1 << 1 }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    { code: 'const o = { [a | b]: 1 }', errors: [{ messageId: 'unexpected' }] },
    {
      code: 'class C { m() { return a | b; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // allow only covers listed operators
    {
      code: '~a',
      options: [{ allow: ['|'] }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    // Mixed precedence — (a + b) & c — only `&` reports
    { code: 'a + b & c', errors: [{ messageId: 'unexpected' }] },
    // Right-associative compound bitwise assignment
    {
      code: 'a |= b |= c',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // Computed element access
    { code: 'obj[a | b]', errors: [{ messageId: 'unexpected' }] },
    // Template literal span
    { code: '`${a | b}`', errors: [{ messageId: 'unexpected' }] },
  ],
});
