import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('yoda', {
  valid: [
    // "never" mode (default)
    'if (value === "red") {}',
    'if (value === value) {}',
    'if (value != 5) {}',
    'if (5 & foo) {}',
    'if (5 === 4) {}',
    { code: 'if (value === "red") {}', options: ['never'] as any },

    // "always" mode
    { code: 'if ("blue" === value) {}', options: ['always'] as any },
    { code: 'if (4 != value) {}', options: ['always'] as any },
    { code: 'if (foo & 4) {}', options: ['always'] as any },

    // Range exception
    {
      code: 'if (0 <= x && x < 1) {}',
      options: ['never', { exceptRange: true }] as any,
    },
    {
      code: 'if (x < -1 || 1 < x) {}',
      options: ['never', { exceptRange: true }] as any,
    },
    {
      code: 'if (0 <= index && index < list.length) {}',
      options: ['never', { exceptRange: true }] as any,
    },

    // onlyEquality
    {
      code: 'if (0 < x && x <= 1) {}',
      options: ['never', { onlyEquality: true }] as any,
    },
    {
      code: "if (x !== 'foo' && 'foo' !== x) {}",
      options: ['never', { onlyEquality: true }] as any,
    },
  ],
  invalid: [
    {
      code: 'if ("red" == value) {}',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (true === value) {}',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (-1 < str.indexOf(substr)) {}',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (value == "red") {}',
      options: ['always'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (value === true) {}',
      options: ['always'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (a < 0 && 0 <= b && b < 1) {}',
      options: ['never', { exceptRange: true }] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 14,
        },
      ],
    },
    {
      code: 'if (3 == a) {}',
      options: ['never', { onlyEquality: true }] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if (0 <= x && x < 1) {}',
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'if ( /* a */ 0 /* b */ < /* c */ foo /* d */ ) {}',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 14,
        },
      ],
    },
    {
      code: 'x=1 < a',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'expected',
          line: 1,
          column: 3,
        },
      ],
    },
  ],
});
