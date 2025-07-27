import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: ['./tsconfig.json'],
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-non-null-asserted-optional-chain', {
  valid: [
    'foo.bar!;',
    'foo.bar!.baz;',
    'foo.bar!.baz();',
    'foo.bar()!;',
    'foo.bar()!();',
    'foo.bar()!.baz;',
    'foo?.bar;',
    'foo?.bar();',
    '(foo?.bar).baz!;',
    '(foo?.bar()).baz!;',
    'foo?.bar!.baz;',
    'foo?.bar!();',
    "foo?.['bar']!.baz;",
  ],
  invalid: [
    {
      code: 'foo?.bar!;',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: 'foo?.bar;',
            },
          ],
        },
      ],
    },
    {
      code: "foo?.['bar']!;",
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: "foo?.['bar'];",
            },
          ],
        },
      ],
    },
    {
      code: 'foo?.bar()!;',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: 'foo?.bar();',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.bar?.()!;',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: 'foo.bar?.();',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar)!.baz',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar).baz',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar)!().baz',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar)().baz',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar)!',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar)',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar)!()',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar)()',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar!)',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar)',
            },
          ],
        },
      ],
    },
    {
      code: '(foo?.bar!)()',
      errors: [
        {
          messageId: 'noNonNullOptionalChain',
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              output: '(foo?.bar)()',
            },
          ],
        },
      ],
    },
  ],
});