import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-unicode-regexp', {
  valid: [
    '/foo/u',
    '/foo/gimuy',
    "RegExp('', 'u')",
    "RegExp('', `u`)",
    "new RegExp('', 'u')",
    "RegExp('', 'gimuy')",
    "RegExp('', `gimuy`)",
    'RegExp(...patternAndFlags)',
    "new RegExp('', 'gimuy')",
    "const flags = 'u'; new RegExp('', flags)",
    "const flags = 'g'; new RegExp('', flags + 'u')",
    "const flags = 'gimu'; new RegExp('foo', flags[3])",
    "new RegExp('', flags)",
    "function f(flags) { return new RegExp('', flags) }",
    "function f(RegExp) { return new RegExp('foo') }",
    'function f(patternAndFlags) { return new RegExp(...patternAndFlags) }',
    {
      code: "new globalThis.RegExp('foo')",
      languageOptions: { ecmaVersion: 6 } as any,
    },
    {
      code: "new globalThis.RegExp('foo')",
      languageOptions: { ecmaVersion: 2017 } as any,
    },
    {
      code: "new globalThis.RegExp('foo', 'u')",
      languageOptions: { ecmaVersion: 2020 } as any,
    },
    {
      code: "globalThis.RegExp('foo', 'u')",
      languageOptions: { ecmaVersion: 2020 } as any,
    },
    {
      code: "const flags = 'u'; new globalThis.RegExp('', flags)",
      languageOptions: { ecmaVersion: 2020 } as any,
    },
    {
      code: "const flags = 'g'; new globalThis.RegExp('', flags + 'u')",
      languageOptions: { ecmaVersion: 2020 } as any,
    },
    {
      code: "const flags = 'gimu'; new globalThis.RegExp('foo', flags[3])",
      languageOptions: { ecmaVersion: 2020 } as any,
    },
    {
      code: "class C { #RegExp; foo() { new globalThis.#RegExp('foo') } }",
      languageOptions: { ecmaVersion: 2022 } as any,
    },
    { code: '/foo/u', options: { requireFlag: 'u' } },
    { code: "new RegExp('foo', 'u')", options: { requireFlag: 'u' } },

    // require the `v` flag.
    { code: '/foo/v', languageOptions: { ecmaVersion: 2024 } as any },
    { code: '/foo/gimvy', languageOptions: { ecmaVersion: 2024 } as any },
    {
      code: "RegExp('', 'v')",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "RegExp('', `v`)",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "new RegExp('', 'v')",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "RegExp('', 'gimvy')",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "RegExp('', `gimvy`)",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "new RegExp('', 'gimvy')",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "const flags = 'v'; new RegExp('', flags)",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "const flags = 'g'; new RegExp('', flags + 'v')",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "const flags = 'gimv'; new RegExp('foo', flags[3])",
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: '/foo/v',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
    },
    {
      code: "new RegExp('foo', 'v')",
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
    },
  ],
  invalid: [
    {
      code: '/\\a/',
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: '/foo/',
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: '/foo/gimy',
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: 'RegExp()',
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "RegExp('foo')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "RegExp('\\\\a')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "RegExp('foo', '')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "RegExp('foo', 'gimy')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "RegExp('foo', `gimy`)",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo',)",
      languageOptions: { ecmaVersion: 2017 } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', false)",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', 1)",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', '')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', 'gimy')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp(('foo'))",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp(('unrelated', 'foo'))",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "const flags = 'gi'; new RegExp('foo', flags)",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 21 }],
    },
    {
      code: "const flags = 'gi'; new RegExp('foo', ('unrelated', flags))",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 21 }],
    },
    {
      code: "let flags; new RegExp('foo', flags = 'g')",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 12 }],
    },
    {
      code: 'const flags = `gi`; new RegExp(`foo`, (`unrelated`, flags))',
      errors: [{ messageId: 'requireUFlag', line: 1, column: 21 }],
    },
    {
      code: "const flags = 'gimu'; new RegExp('foo', flags[0])",
      errors: [{ messageId: 'requireUFlag', line: 1, column: 23 }],
    },
    {
      code: "new window.RegExp('foo')",
      languageOptions: { globals: { window: 'readonly' } } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new global.RegExp('foo')",
      languageOptions: { globals: { global: 'writable' } } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new globalThis.RegExp('foo')",
      languageOptions: { ecmaVersion: 2020 } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: '/foo/',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: '/foo/u',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: '/foo/u',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 6 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: '/[[a]/u',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', 'u')",
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('[[a]', 'u')",
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp("foo", "\\u0067")',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp("foo", `\\u0067`)',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp("foo", "\\u0075")',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp("foo", `\\u0075`)',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: 'const regularFlags = "sm"; new RegExp("foo", `${regularFlags}g`)',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 28 }],
    },
    {
      code: 'const regularFlags = "smu"; new RegExp("foo", `${regularFlags}g`)',
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 29 }],
    },
    {
      code: '/foo/v',
      options: { requireFlag: 'u' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo')",
      options: { requireFlag: 'v' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireVFlag', line: 1, column: 1 }],
    },
    {
      code: "new RegExp('foo', 'v')",
      options: { requireFlag: 'u' },
      languageOptions: { ecmaVersion: 2024 } as any,
      errors: [{ messageId: 'requireUFlag', line: 1, column: 1 }],
    },
  ],
});
